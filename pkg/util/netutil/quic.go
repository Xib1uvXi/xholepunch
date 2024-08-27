package netutil

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/quic-go/quic-go"
	"math/big"
	"net"
	"sync"
	"time"
)

const protos = "quic"

var (
	streamHandlerTimeout = 30 * time.Second
)

var (
	ErrStreamHandlerNotSet = errors.InternalServer("stream handler not set", "stream handler not set")
)

type ReliableUDPStreamHandler = func(conn quic.Connection, stream quic.Stream) error

var (
	sererTLSConfig  *tls.Config
	clientTLSConfig *tls.Config

	onceServer sync.Once
	onceClient sync.Once
)

func ServerTLSConfig() *tls.Config {
	onceServer.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			panic(err)
		}
		template := x509.Certificate{SerialNumber: big.NewInt(1)}
		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
		if err != nil {
			panic(err)
		}
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

		tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			panic(err)
		}
		sererTLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			NextProtos:   []string{protos},
		}
	})

	return sererTLSConfig
}

func ClientTLSConfig() *tls.Config {
	onceClient.Do(func() {
		clientTLSConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{protos},
		}
	})

	return clientTLSConfig
}

type ReliableUDPServer struct {
	ctx       context.Context
	cancelFuc context.CancelFunc

	UDPConn        *net.UDPConn
	quicListener   *quic.Listener
	streamHandler  ReliableUDPStreamHandler
	setHandlerOnce sync.Once
	stopC          chan struct{}
}

func NewReliableUDPServer(udpConn *net.UDPConn) (*ReliableUDPServer, error) {
	quicListener, err := quic.Listen(udpConn, ServerTLSConfig(), nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ReliableUDPServer{
		ctx:          ctx,
		cancelFuc:    cancel,
		UDPConn:      udpConn,
		quicListener: quicListener,
		stopC:        make(chan struct{}),
	}, nil
}

// SetStreamHandler sets the stream handler
func (s *ReliableUDPServer) SetStreamHandler(handler ReliableUDPStreamHandler) {
	s.setHandlerOnce.Do(func() {
		s.streamHandler = handler
	})
}

// Start starts the server
func (s *ReliableUDPServer) Start() error {
	if s.streamHandler == nil {
		return ErrStreamHandlerNotSet
	}

	go s.acceptLoop()

	return nil
}

// Close closes the server
func (s *ReliableUDPServer) Close() error {
	s.cancelFuc()
	close(s.stopC)

	if err := s.quicListener.Close(); err != nil {
		return err
	}

	if err := s.UDPConn.Close(); err != nil {
		return err
	}

	return nil
}

func (s *ReliableUDPServer) acceptLoop() {
	for {
		select {
		case <-s.stopC:
			return
		default:
		}

		conn, err := s.quicListener.Accept(s.ctx)
		if err != nil {
			log.Errorf("accept quic error: %s", err)
			continue
		}

		go s.handleConn(conn)
	}
}

// handle quic connection
func (s *ReliableUDPServer) handleConn(quicConn quic.Connection) {
	for {
		select {
		case <-s.stopC:
			return
		default:
		}

		stream, err := quicConn.AcceptStream(s.ctx)
		if err != nil {
			if quicErr, ok := err.(*quic.ApplicationError); ok && quicErr.ErrorCode == 11 {
				return
			}

			log.Warnf("accept stream error: %s", err)
			continue
		}

		go s.acceptSteamLoop(quicConn, stream)
	}

}

func (s *ReliableUDPServer) acceptSteamLoop(quicConn quic.Connection, stream quic.Stream) {
	defer stream.Close()
	_ = stream.SetDeadline(time.Now().Add(streamHandlerTimeout))

	if err := s.streamHandler(quicConn, stream); err != nil {
		log.Errorf("stream handler error: %s", err)
	}

	return
}

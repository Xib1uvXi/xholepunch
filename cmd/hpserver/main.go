package main

import (
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/go-kratos/kratos/v2/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := log.With(log.DefaultLogger,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	log.SetLogger(logger)
	log.DefaultLogger = logger

	serverAddrStr := ":4321"
	server, err := rendezvous.Builder(serverAddrStr)

	if err != nil {
		panic(err)
	}

	log.Infof("Server started at %s", serverAddrStr)

	go server.Serve()
	defer server.Close()

	signalChan := make(chan os.Signal, 1)
	go signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signalChan

	log.Infof("Received signal: %s", sig)
}

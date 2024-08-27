package tests

import (
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/Xib1uvXi/xholepunch/pkg/traversal"
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCase(t *testing.T) {
	serverAddrStr := "127.0.0.1:7777"
	server, err := rendezvous.Builder(serverAddrStr)
	require.NoError(t, err)
	go server.Serve()
	defer server.Close()

	cLocal, err := traversal.Builder(serverAddrStr, int8(types.PortRestrictedCone))
	require.NoError(t, err)
	defer cLocal.Close()

	cRemote, err := traversal.Builder(serverAddrStr, int8(types.Symmetric))
	require.NoError(t, err)
	defer cRemote.Close()

	token := "token-111"

	go func() {
		require.NoError(t, cLocal.HolePunching(token))
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, cRemote.HolePunching(token))
	}()

	time.Sleep(10 * time.Second)

}

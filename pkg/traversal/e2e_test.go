package traversal

import (
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func Test_newE2EHandler(t *testing.T) {
	laddr := "127.0.0.1:14000"
	raddr := "127.0.0.1:15000"
	localNAT := int8(types.FullConeNATType)
	remoteNAT := int8(types.FullConeNATType)

	cLocal, err := newE2EHandler(laddr, raddr, localNAT, remoteNAT)
	require.NoError(t, err)
	defer cLocal.Close()

	cRemote, err := newE2EHandler(raddr, laddr, remoteNAT, localNAT)
	require.NoError(t, err)
	defer cRemote.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		result, err := cLocal.HolePunch()
		require.NoError(t, err)
		require.NotNil(t, result)
		t.Logf("local result: %v", result)
	}()

	go func() {
		defer wg.Done()
		result, err := cRemote.HolePunch()
		require.NoError(t, err)
		require.NotNil(t, result)
		t.Logf("remote result: %v", result)
	}()

	wg.Wait()

}

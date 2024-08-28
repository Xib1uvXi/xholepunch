package predictor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewPortBitmap(t *testing.T) {
	pb := NewPortBitmap()
	require.NotNil(t, pb)

	checkPort := 1000
	use, err := pb.IsPortSet(checkPort)
	require.NoError(t, err)
	require.False(t, use)

	err = pb.SetPort(checkPort)
	require.NoError(t, err)

	use, err = pb.IsPortSet(checkPort)
	require.NoError(t, err)
	require.True(t, use)
}

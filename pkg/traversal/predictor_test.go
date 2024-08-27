package traversal

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLinearPortPredictor_NextPort(t *testing.T) {
	tp := NewLinearPortPredictor(0)

	require.Equal(t, 1, tp.NextPort())
}

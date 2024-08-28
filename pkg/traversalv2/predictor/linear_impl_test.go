package predictor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewLinearPortPredictor(t *testing.T) {
	tp := NewLinearPortPredictor(0)

	require.Equal(t, 1, tp.NextPort())
}

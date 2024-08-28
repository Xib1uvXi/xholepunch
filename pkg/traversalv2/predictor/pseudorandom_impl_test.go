package predictor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPseudorandomPortPredictor_initLinear(t *testing.T) {
	p := NewPseudorandomPortPredictor(100, 100)
	require.Len(t, p.ports, 100)
}

func TestPseudorandomPortPredictor_initLRandom(t *testing.T) {
	p := NewPseudorandomPortPredictor(100, 460)
	require.Len(t, p.ports, 460)

	t.Logf("ports: %v", p.ports)
}

func TestPseudorandomPortPredictor_NextPort(t *testing.T) {
	p := NewPseudorandomPortPredictor(100, 50)

	list := make([]int, 0, 0)

	for {
		p := p.NextPort()
		if p == 0 {
			break
		}

		list = append(list, p)
	}

	require.Len(t, list, 50)
	t.Logf("list: %v", list)

	require.Len(t, p.ports, 0)

}

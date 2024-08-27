package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNATType_String(t *testing.T) {
	case1 := "FullCone"
	case2 := "RestrictedCone"
	case3 := "PortRestrictedCone"
	case4 := "Symmetric"

	require.Equal(t, case1, NATType(FullConeNATType).String())
	require.Equal(t, case2, NATType(RestrictedCone).String())
	require.Equal(t, case3, NATType(PortRestrictedCone).String())
	require.Equal(t, case4, NATType(Symmetric).String())
	require.Equal(t, "UnKnown", NATType(0).String())
}

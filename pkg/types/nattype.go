package types

type NATType int8

func (t NATType) String() string {
	switch t {
	case FullConeNATType:
		return "FullCone"
	case RestrictedCone:
		return "RestrictedCone"
	case PortRestrictedCone:
		return "PortRestrictedCone"
	case Symmetric:
		return "Symmetric"
	default:
		return "UnKnown"
	}
}

const (
	FullConeNATType = iota + 1
	RestrictedCone
	PortRestrictedCone
	Symmetric
)

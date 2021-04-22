package regions

// Region is region's type
type Region int

func (r Region) String() string {
	switch r {
	case ChinaNorth:
		return "cn-n1"
	case ChinaEast:
		return "cn-e1"
	case USWest:
		return "us-w1"
	default:
		return "invalid"
	}
}

func (r Region) EnvString() string {
	switch r {
	case ChinaNorth, ChinaEast:
		return "cn"
	case USWest:
		return "us"
	default:
		return "invalid"
	}
}

// Description is region's readable description
func (r Region) Description() string {
	switch r {
	case ChinaNorth:
		return "China North"
	case USWest:
		return "United States"
	case ChinaEast:
		return "China East"
	default:
		return "invalid"
	}
}

// API server regions
const (
	Invalid Region = iota
	ChinaNorth
	USWest
	ChinaEast
)

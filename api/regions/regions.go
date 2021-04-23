package regions

// Region is region's type
type Region int

func Parse(region string) Region {
	switch region {
	case "cn", "CN", "cn-n1":
		return ChinaNorth
	case "tab", "TAB", "cn-e1":
		return ChinaEast
	case "us", "US", "us-w1":
		return USWest
	default:
		return Invalid
	}
}

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
		return "CN"
	case USWest:
		return "US"
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

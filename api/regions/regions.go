package regions

// Region is region's type
type Region int

func (r Region) String() string {
	switch r {
	case CN:
		return "cn"
	case US:
		return "us"
	case TAB:
		return "tab"
	default:
		return "invalid"
	}
}

// Description is region's readable description
func (r Region) Description() string {
	switch r {
	case CN:
		return "China North"
	case US:
		return "United States"
	case TAB:
		return "China East"
	default:
		return "invalid"
	}
}

// API server regions
const (
	Invalid Region = iota
	CN
	US
	TAB
)

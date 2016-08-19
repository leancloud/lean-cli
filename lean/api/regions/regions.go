package regions

// Region is region's type
type Region int

func (r Region) String() string {
	switch r {
	case CN:
		return "cn"
	case US:
		return "us"
	default:
		return "invalid"
	}
}

// API server regions
const (
	Invalid Region = iota
	CN
	US
)

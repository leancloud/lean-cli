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

// APIServerURL returns this region's API Server URL
func (r Region) APIServerURL() string {
	switch r {
	case CN:
		return "https://api.leancloud.cn"
	case US:
		return "https://us-api.leancloud.cn"
	case TAB:
		return "https://e1-api.leancloud.cn"
	default:
		return ""
	}
}

// API server regions
const (
	Invalid Region = iota
	CN
	US
	TAB
)

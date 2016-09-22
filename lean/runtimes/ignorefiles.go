package runtimes

// DefaultIgnorePatterns returns current runtime's default ignore patterns
func (runtime *Runtime) DefaultIgnorePatterns() []string {
	switch runtime.Name {
	case "node.js":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"node_modules/",
		}
	case "java":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"target/",
		}
	case "php":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"vendor/",
		}
	case "python":
		return []string{
			".git/",
			".avoscloud/",
			".leancloud/",
			"venv",
			"*.pyc",
			"__pycache__/",
		}
	default:
		panic("invalid runtime")
	}
}

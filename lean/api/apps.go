package api

// GetAppList returns the current user's all LeanCloud application
func GetAppList() ([]interface{}, error) {
	client, err := NewCookieAuthClient()
	if err != nil {
		return nil, err
	}

	result, err := client.get("/clients/self/apps", nil)
	if err != nil {
		return nil, err
	}
	return result.MustArray(), nil
}

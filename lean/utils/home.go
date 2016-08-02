package utils

import (
	"os"
	"os/user"
	"path/filepath"
)

// HomeDir returns the system current user home directory
func HomeDir() string {
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		return homeDir
	}

	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return currentUser.HomeDir
}

// ConfigDir returns the current user's XDG config dir
func ConfigDir() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir != "" {
		return configDir
	}

	return filepath.Join(HomeDir(), ".config")
}

package utils

import (
	"os"
	"os/user"
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

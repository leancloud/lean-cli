package utils

import "log"

// CheckError will check if the err is nil, and exit the whole program
func CheckError(err error) {
	if err == nil {
		return
	}
	log.Fatalln(err)
}

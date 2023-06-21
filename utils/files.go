package utils

import "os"

func Exists(path string) (bool, error) {
	// CC: https://stackoverflow.com/a/10510783
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

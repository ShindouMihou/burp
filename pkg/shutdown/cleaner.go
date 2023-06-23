package shutdown

import (
	"burp/pkg/fileutils"
	"errors"
	"os"
)

var CleanupDirectories = []string{
	fileutils.JoinHomePath(".burpy", ".build"),
	fileutils.JoinHomePath(".burpy", ".files", ".packaged"),
}

func Cleanup() error {
	for _, directory := range CleanupDirectories {
		if err := os.RemoveAll(directory); err != nil {
			return errors.Join(errors.New("failed to cleanup "+directory+" folder"), err)
		}
	}
	return nil
}

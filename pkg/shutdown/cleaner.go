package shutdown

import (
	"errors"
	"os"
)

func Cleanup() error {
	if err := os.RemoveAll(".burp/"); err != nil {
		return errors.Join(errors.New("failed to cleanup .burp/ folder"), err)
	}
	return nil
}

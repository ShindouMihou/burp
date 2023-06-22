package reader

import (
	"burp/burper"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

func Open(file string) (*os.File, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Close(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Err(err).Str("origin", "burp_toml").Msg("Failed to close body")
	}
}

func Read(file string) (*burper.Tree, error) {
	f, err := Open(file)
	if err != nil {
		return nil, err
	}
	defer Close(f)
	return burper.New(f)
}

func Save(file string, data []byte) error {
	if strings.Contains(file, "/") {
		err := os.MkdirAll(filepath.Dir(file), os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer Close(f)
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

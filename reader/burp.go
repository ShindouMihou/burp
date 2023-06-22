package reader

import (
	"burp/burper"
	"burp/utils"
	"crypto/sha256"
	"encoding/hex"
	"github.com/docker/docker/pkg/archive"
	"github.com/rs/zerolog/log"
	"io"
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

func Create(file string) (*os.File, error) {
	if strings.Contains(file, "/") {
		err := os.MkdirAll(filepath.Dir(file), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Save(file string, data []byte) error {
	f, err := Create(file)
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

func Copy(source, dest string) (*string, error) {
	f, err := Create(dest)
	if err != nil {
		return nil, err
	}
	defer Close(f)
	r, err := Open(source)
	if err != nil {
		return nil, err
	}
	defer Close(r)
	hash := sha256.New()
	r2 := io.TeeReader(r, hash)
	_, err = io.Copy(f, r2)
	if err != nil {
		return nil, err
	}
	return utils.Ptr(hex.EncodeToString(hash.Sum(nil))), nil
}

func Tar(src, dest string) error {
	r, err := archive.TarWithOptions(src, &archive.TarOptions{})
	if err != nil {
		return err
	}
	f, err := Create(dest)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	return nil
}

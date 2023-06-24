package fileutils

import (
	"burp/pkg/utils"
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
		log.Err(err).Str("file", f.Name()).Msg("Failed to close body")
	}
}

func MkdirParent(file string) error {
	if strings.Contains(file, "/") {
		if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func Create(file string) (*os.File, error) {
	if err := MkdirParent(file); err != nil {
		return nil, err
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

func Sanitize(key string) string {
	key = filepath.Clean(filepath.Base(key))
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

var homeDirectory = ""

func GetHomeDir() string {
	if homeDirectory == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Panic().Err(err).Msg("Failed to get home directory path")
		}
		homeDirectory = home
	}
	return homeDirectory
}

func JoinHomePath(paths ...string) string {
	return filepath.Join(GetHomeDir(), filepath.Join(paths...))
}

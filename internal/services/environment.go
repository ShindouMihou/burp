package services

import (
	"bufio"
	"burp/internal/burper"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"bytes"
	"io"
	"path/filepath"
	"strings"
)

func (env *Environment) Translate(dir string) (*string, error) {
	f, err := fileutils.Open(filepath.Join(dir, env.Baseline))
	if err != nil {
		return nil, err
	}
	defer fileutils.Close(f)
	buf := bufio.NewScanner(f)
	var b strings.Builder
	for buf.Scan() {
		line := make([]byte, len(buf.Bytes()))
		copy(line, buf.Bytes())

		if bytes.HasPrefix(line, burper.COMMENT_KEY) {
			b.Write(line)
			b.Write(burper.NEWLINE_KEY)
			continue
		}
		parts := bytes.SplitN(line, burper.EQUALS_KEY, 2)
		if len(parts) != 2 {
			b.Write(line)
			b.Write(burper.NEWLINE_KEY)
			continue
		}
		key, _ := parts[0], parts[1]
		if replacement, exists := env.Replacements[string(key)]; exists {
			b.WriteString(string(key) + "=" + replacement)
			b.Write(burper.NEWLINE_KEY)
			continue
		}
		b.Write(line)
		b.Write(burper.NEWLINE_KEY)
	}
	return utils.Ptr(b.String()), nil
}

func EnvironmentReadBuffer(reader io.Reader) []string {
	buf := bufio.NewScanner(reader)
	var e []string
	for buf.Scan() {
		line := buf.Bytes()
		if bytes.HasPrefix(line, burper.COMMENT_KEY) {
			continue
		}
		parts := bytes.SplitN(line, burper.EQUALS_KEY, 2)
		if len(parts) != 2 {
			continue
		}
		e = append(e, string(line))
	}
	return e
}

func (env *Environment) Read(dir string) ([]string, error) {
	d := filepath.Join(dir, env.Baseline)
	f, err := fileutils.Open(d)
	if err != nil {
		return nil, err
	}
	defer fileutils.Close(f)
	return EnvironmentReadBuffer(f), nil
}

func (env *Environment) Save(dir string) error {
	translation, err := env.Translate(dir)
	if err != nil {
		return err
	}
	d := filepath.Join(dir, env.Baseline)
	return fileutils.Save(d, []byte(*translation))
}

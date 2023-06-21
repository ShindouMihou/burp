package services

import (
	"bufio"
	"burp/burper"
	"burp/reader"
	"burp/utils"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func (env *Environment) Translate(dir string) (*string, error) {
	f, err := reader.Open(dir + env.File)
	if err != nil {
		return nil, err
	}
	defer reader.Close(f)
	buf := bufio.NewScanner(f)
	var b strings.Builder
	for buf.Scan() {
		line := buf.Bytes()
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

func (env *Environment) Read(dir string) ([]string, error) {
	d := filepath.Join(dir, env.Output)
	f, err := reader.Open(d)
	if err != nil {
		return nil, err
	}
	defer reader.Close(f)
	buf := bufio.NewScanner(f)
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
	return e, nil
}

func (env *Environment) Save(dir string) error {
	translation, err := env.Translate(dir)
	if err != nil {
		return err
	}
	d := filepath.Join(dir, env.Output)
	if strings.Contains(env.Output, "/") {
		err = os.MkdirAll(filepath.Dir(d), os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(d)
	if err != nil {
		return err
	}
	defer reader.Close(f)
	_, err = f.Write([]byte(*translation))
	if err != nil {
		return err
	}
	return nil
}

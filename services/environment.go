package services

import (
	"bufio"
	"burp/burper"
	"burp/reader"
	"burp/utils"
	"bytes"
	"os"
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

func (env *Environment) Save(dir string) error {
	translation, err := env.Translate(dir)
	if err != nil {
		return err
	}
	if strings.Contains(env.Output, "/") {
		err = os.MkdirAll(dir+env.Output, os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(dir + env.Output)
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

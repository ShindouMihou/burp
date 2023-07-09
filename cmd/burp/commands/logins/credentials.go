package logins

import (
	"burp/cmd/burp/api"
	"encoding/json"
	"fmt"
	"github.com/portainer/libcrypto"
	"github.com/ttacon/chalk"
	"os"
	"path/filepath"
)

var Folder string
var Servers []string

func Unlock(keys *api.Keys) (*api.Secrets, error) {
	keys.Sanitize()
	file := filepath.Join(Folder, keys.Name+".json")

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	bytes, err = libcrypto.Decrypt(bytes, []byte(keys.Encryption))
	if err != nil {
		return nil, err
	}
	var secrets api.Secrets
	if err = json.Unmarshal(bytes, &secrets); err != nil {
		return nil, err
	}
	secrets.Sanitize()
	return &secrets, nil
}

func MustUnlock(keys *api.Keys) (*api.Secrets, bool) {
	secrets, err := Unlock(keys)
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "You said the wrong magic word (encryption key), that's bad!")
		fmt.Println(chalk.Red, err.Error())
		return nil, false
	}
	return secrets, true
}

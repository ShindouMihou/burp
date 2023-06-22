package server

import (
	"burp/pkg/env"
	"burp/pkg/fileutils"
	"errors"
	"github.com/portainer/libcrypto"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"time"
)

var TemporarySslDirectory = ".certs"

func GetSsl() (cert string, key string, err error) {
	if env.SslCertificatePath.OrNull() != nil && env.SslKeyPath.OrNull() != nil {
		return env.SslCertificatePath.Get(), env.SslKeyPath.Get(), nil
	}

	log.Warn().Msg("SSL Certificates cannot be found.")
	log.Warn().Msg("Although Burp can auto-generate its own SSL certificates, it's more recommended to create your own.")
	log.Warn().Msg("You may receive error logs such as \"tls: bad certificate\" due to self-signed certificates.")

	certificatePath := filepath.Join(TemporarySslDirectory, "ssl.cert")
	keyPath := filepath.Join(TemporarySslDirectory, "key.pem")

	hasGeneratedBefore := true

	var checkError = func(err error) error {
		if err == nil {
			return nil
		}
		if errors.Is(err, os.ErrNotExist) {
			hasGeneratedBefore = false
			return nil
		}
		return err
	}

	certificateFile, err := fileutils.Open(certificatePath)
	if checkError(err) != nil {
		return "", "", err
	}
	keyFile, err := fileutils.Open(keyPath)
	if checkError(err) != nil {
		return "", "", err
	}

	if hasGeneratedBefore {
		fileutils.Close(certificateFile)
		fileutils.Close(keyFile)
		return certificatePath, keyPath, nil
	}

	if err := fileutils.MkdirParent(certificatePath); err != nil {
		return "", "", err
	}
	// Thanks Portainer for libcrypto! (CC: https://github.com/portainer/libcrypto)
	err = libcrypto.GenerateCertsForHost("localhost", "0.0.0.0", certificatePath, keyPath,
		time.Now().AddDate(5, 0, 0))
	if err != nil {
		return "", "", err
	}
	return certificatePath, keyPath, nil
}

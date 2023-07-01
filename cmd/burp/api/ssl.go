package api

import (
	"burp/cmd/burp-agent/server"
	"burp/pkg/fileutils"
	"net/url"
	"path/filepath"
)

func SaveCertificate(host string, name string) error {
	certificateRoute, _ := url.JoinPath(host, "ssl.cert")
	certificate, err := CreateInsecure().Get(certificateRoute)
	if err != nil {
		return err
	}
	sslCertificateLocation := filepath.Join(server.TemporarySslDirectory, name, "ssl.cert")
	if err := fileutils.Save(sslCertificateLocation, certificate.Body()); err != nil {
		return err
	}
	return nil
}

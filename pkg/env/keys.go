package env

import "os"

const (
	AgentMode          Key = "AGENT_MODE"
	BurpSignature      Key = "BURP_SIGNATURE"
	BurpSecret         Key = "BURP_SECRET"
	GitToml            Key = "GIT_TOML"
	DockerToml         Key = "DOCKER_TOML"
	SslCertificatePath Key = "SSL_CERTIFICATE_PATH"
	SslKeyPath         Key = "SSL_KEY_PATH"
	ServerIp           Key = "SERVER_IP"
	ServerHostname     Key = "SERVER_HOSTNAME"
)

type Key string

func (key Key) OrNull() *string {
	return GetOrNull(string(key))
}

func (key Key) Get() string {
	return os.Getenv(string(key))
}

func (key Key) Or(def string) string {
	return GetDefault(string(key), def)
}

func (key Key) String() string {
	return string(key)
}

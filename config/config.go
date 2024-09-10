package config

type Config struct {
	NatsServerURL string
	NatsHubDomain string
	Credentials   Credentials
}

type Credentials struct {
	Username string
	Password string
}

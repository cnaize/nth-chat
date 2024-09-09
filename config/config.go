package config

type Config struct {
	ServerURL string
	Creds     Credentials
}

type Credentials struct {
	Username string
	Password string
}

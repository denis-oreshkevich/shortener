package config

type Conf struct {
	ServerAddress string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL       string `json:"base_url" env:"BASE_URL"`
	BasePath      string
	FsPath        string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN   string `json:"database_dsn" env:"DATABASE_DSN"`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`
	ConfFilePath  string `json:"-" env:"CONFIG"`
}

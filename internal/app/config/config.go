package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net/url"
	"os"
	"strings"
)

var conf Conf

// Get Conf global variable that holds properties from command line and environment variables.
func Get() Conf {
	return conf
}

// Parse func parses command line and environment variables and init Config [config.Conf]
func Parse() error {
	flag.StringVar(&conf.ConfFilePath, "c", "./conf/config.json",
		"Path to config JSON file")

	file, err := os.ReadFile(conf.ConfFilePath)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	err = json.Unmarshal(file, &conf)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	flag.StringVar(&conf.ServerAddress, "a", "", "HTTP server address")

	flag.StringVar(&conf.BaseURL, "b", "", "HTTP server base URL")

	flag.StringVar(&conf.FsPath, "f", "", "Path to file storage")

	//host=localhost port=5433 user=postgres password=postgres dbname=courses sslmode=disable
	flag.StringVar(&conf.DatabaseDSN, "d", "", "Data Source Name (DSN)")

	flag.BoolVar(&conf.EnableHTTPS, "s", false, "Enables HTTPS")

	flag.Parse()

	err = env.Parse(&conf)
	if err != nil {
		return fmt.Errorf("env.Parse: %w", err)
	}
	parsed, err := url.Parse(conf.BaseURL)
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}

	conf.BasePath = strings.TrimSuffix(parsed.Path, "/")

	return nil
}

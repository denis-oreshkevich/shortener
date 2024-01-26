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

	a := flag.String("a", "", "HTTP server address")
	if len(*a) != 0 {
		conf.ServerAddress = *a
	}

	b := flag.String("b", "", "HTTP server base URL")
	if len(*b) != 0 {
		conf.BaseURL = *b
	}

	f := flag.String("f", "", "Path to file storage")
	if len(*f) != 0 {
		conf.FsPath = *f
	}

	//host=localhost port=5433 user=postgres password=postgres dbname=courses sslmode=disable
	d := flag.String("d", "", "Data Source Name (DSN)")
	if len(*d) != 0 {
		conf.DatabaseDSN = *d
	}

	s := flag.Bool("s", false, "Enables HTTPS")
	if !*s {
		conf.EnableHTTPS = *s
	}

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

package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
	"net"
	"net/url"
	"os"
	"strings"
)

const (
	ServerAddressEnvName = "SERVER_ADDRESS"

	BaseURLEnvName = "BASE_URL"

	FileStoragePath = "FILE_STORAGE_PATH"

	DatabaseDSN = "DATABASE_DSN"

	defaultHost = "localhost"

	defaultPort = "8080"

	defaultScheme = "http"
)

var conf Conf

func Get() Conf {
	return conf
}

type initStructure struct {
	envName  string
	argVal   string
	initFunc func(s string) error
}

func Parse() error {
	conf = Conf{}
	a := flag.String("a", fmt.Sprintf("%s:%s", defaultHost, defaultPort), "HTTP server address")
	b := flag.String("b", fmt.Sprintf("%s://%s:%s", defaultScheme, defaultHost, defaultPort), "HTTP server base URL")
	// /tmp/short-url-db.json
	f := flag.String("f", "", "Path to storage file")
	//host=localhost port=5433 user=postgres password=postgres dbname=courses sslmode=disable
	d := flag.String("d", "", "Database connection")
	flag.Parse()

	isa := initStructure{
		envName:  ServerAddressEnvName,
		argVal:   *a,
		initFunc: serverAddrFunc(),
	}

	err := initAppParam(isa)
	if err != nil {
		return err
	}

	ibu := initStructure{
		envName:  BaseURLEnvName,
		argVal:   *b,
		initFunc: baseURLFunc(),
	}

	err = initAppParam(ibu)
	if err != nil {
		return err
	}

	ifs := initStructure{
		envName: FileStoragePath,
		argVal:  *f,
		initFunc: func(s string) error {
			conf.fsPath = s
			return nil
		},
	}
	err = initAppParam(ifs)
	if err != nil {
		return err
	}

	idb := initStructure{
		envName: DatabaseDSN,
		argVal:  *d,
		initFunc: func(s string) error {
			conf.databaseDSN = s
			return nil
		},
	}
	err = initAppParam(idb)
	if err != nil {
		return err
	}

	logger.Log.Info(fmt.Sprintf("Result configuration: %+v\n", conf))
	return nil
}

func initAppParam(is initStructure) error {
	sa, ex := os.LookupEnv(is.envName)
	if !ex {
		sa = is.argVal
		logger.Log.Info(fmt.Sprintf("Env variable %s not found try to init from command line args\n", is.envName))
	}
	err := is.initFunc(sa)
	return err
}

func serverAddrFunc() func(s string) error {
	return func(hp string) error {
		if hp == "" {
			return errors.New("serverAddrFunc arg is empty")
		}
		host, port, err := net.SplitHostPort(hp)
		if err != nil {
			return fmt.Errorf("serverAddrFunc split host %w", err)
		}
		conf.host = host
		conf.port = port
		return nil
	}
}

func baseURLFunc() func(s string) error {
	return func(u string) error {
		if !validator.URL(u) {
			return errors.New("baseURLFunc validating URL")
		}
		parsed, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("baseURLFunc parse url %w", err)
		}

		host, port, err := net.SplitHostPort(parsed.Host)
		if err != nil {
			return fmt.Errorf("split host %w", err)
		}

		conf.scheme = parsed.Scheme
		conf.host = host
		conf.port = port
		conf.basePath = strings.TrimSuffix(parsed.Path, "/")

		conf.baseURL = strings.TrimSuffix(u, "/")
		return nil
	}
}

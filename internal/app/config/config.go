package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
)

// Constants for configuration.
const (
	// ServerAddressEnvName Server address environment variable name.
	ServerAddressEnvName = "SERVER_ADDRESS"

	// BaseURLEnvName Base URL environment variable name.
	BaseURLEnvName = "BASE_URL"

	// FileStoragePath File storage environment variable name.
	FileStoragePath = "FILE_STORAGE_PATH"

	databaseDSN = "DATABASE_DSN"

	enableHTTPS = "ENABLE_HTTPS"

	defaultHost = "localhost"

	defaultPort = "8080"

	defaultScheme = "http"
)

var conf Conf
var cfJSON confJSON

// Get Conf global variable that holds properties from command line and environment variables.
func Get() Conf {
	return conf
}

type initStructure struct {
	envName    string
	argVal     string
	defaultVal string
	initFunc   func(s string) error
}

// Parse func parses command line and environment variables and init Conf [config.Conf]
func Parse() error {
	conf = Conf{}
	a := flag.String("a", "", "HTTP server address")
	b := flag.String("b", "", "HTTP server base URL")
	// /tmp/short-url-db.json
	f := flag.String("f", "", "Path to storage file")
	//host=localhost port=5433 user=postgres password=postgres dbname=courses sslmode=disable
	d := flag.String("d", "", "Database connection")
	s := flag.String("s", "", "Enables HTTPS")
	c := flag.String("c", "./conf/config.json", "Path to configuration file")
	flag.Parse()

	file, err := os.ReadFile(*c)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	err = json.Unmarshal(file, &cfJSON)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	isa := initStructure{
		envName:    ServerAddressEnvName,
		argVal:     *a,
		defaultVal: cfJSON.ServerAddress,
		initFunc:   serverAddrFunc(),
	}

	err = initAppParam(isa)
	if err != nil {
		return err
	}

	ibu := initStructure{
		envName:    BaseURLEnvName,
		argVal:     *b,
		defaultVal: cfJSON.BaseURL,
		initFunc:   baseURLFunc(),
	}

	err = initAppParam(ibu)
	if err != nil {
		return err
	}

	ifs := initStructure{
		envName:    FileStoragePath,
		argVal:     *f,
		defaultVal: cfJSON.FsPath,
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
		envName:    databaseDSN,
		argVal:     *d,
		defaultVal: cfJSON.DatabaseDSN,
		initFunc: func(s string) error {
			conf.databaseDSN = s
			return nil
		},
	}
	err = initAppParam(idb)
	if err != nil {
		return err
	}

	ieh := initStructure{
		envName:    enableHTTPS,
		argVal:     *s,
		defaultVal: cfJSON.EnableHTTPS,
		initFunc: func(s string) error {
			if len(s) == 0 {
				conf.enableHTTPS = false
				return nil
			}
			boolValue, boolErr := strconv.ParseBool(s)
			if boolErr != nil {
				return fmt.Errorf("strconv.ParseBool: %w", err)
			}
			conf.enableHTTPS = boolValue
			return nil
		},
	}
	err = initAppParam(ieh)
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
	if len(sa) == 0 {
		sa = is.defaultVal
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

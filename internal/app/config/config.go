package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
	"net"
	"net/url"
	"os"
	"strings"
)

const (
	ServerAddressEnvName = "SERVER_ADDRESS"

	BaseURLEnvName = "BASE_URL"

	defaultHost = "localhost"

	defaultPort = "8080"

	defaultScheme = "http"
)

var srvConf ServerConf

func Get() ServerConf {
	return srvConf
}

type initStructure struct {
	envName  string
	argVal   string
	initFunc func(s string) error
}

func Parse() error {
	srvConf = ServerConf{}
	a := flag.String("a", fmt.Sprintf("%s:%s", defaultHost, defaultPort), "HTTP server address")
	b := flag.String("b", fmt.Sprintf("%s://%s:%s", defaultScheme, defaultHost, defaultPort), "HTTP server base URL")
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
	fmt.Printf("Result server configuration: %+v\n", srvConf)
	return nil
}

func initAppParam(is initStructure) error {
	sa, ex := os.LookupEnv(is.envName)
	if !ex {
		sa = is.argVal
		fmt.Printf("Env variable %s not found try to init from command line args\n", is.envName)
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
		srvConf.host = host
		srvConf.port = port
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

		srvConf.scheme = parsed.Scheme
		srvConf.host = host
		srvConf.port = port
		srvConf.basePath = strings.TrimSuffix(parsed.Path, "/")

		srvConf.baseURL = strings.TrimSuffix(u, "/")
		return nil
	}
}

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

	defaultProtocol = "http"
)

var srvConf ServerConf

type ServerConf struct {
	Scheme   string
	Host     string
	Port     string
	BaseURL  string
	BasePath string
}

type initStructure struct {
	envName  string
	argName  string
	argVal   string
	usage    string
	initFunc func(s string) error
}

func Get() ServerConf {
	return srvConf
}

func init() {
	srvConf = ServerConf{}
	//flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	srvAddr := initStructure{
		envName:  ServerAddressEnvName,
		argName:  "a",
		argVal:   "",
		usage:    "HTTP server address",
		initFunc: initServerAddrFunc(),
	}
	initAppArg(srvAddr)
	bu := initStructure{
		envName:  BaseURLEnvName,
		argName:  "b",
		argVal:   "",
		usage:    "HTTP server base URL",
		initFunc: initBaseURLFunc(),
	}
	initAppArg(bu)

	initDefault()
	fmt.Printf("Result server configuration: %+v\n", srvConf)
}

func initAppArg(is initStructure) {
	v, ex := os.LookupEnv(is.envName)
	inFunc := is.initFunc
	if ex {
		flag.String(is.argName, is.argVal, is.usage)
		inFunc(v)
	} else {
		fmt.Printf("Env variable %s not found try to init from command line args\n", is.envName)
		flag.Func(is.argName, is.usage, is.initFunc)
	}
}

func initServerAddrFunc() func(hp string) error {
	return func(hp string) error {
		if hp == "" {
			return errors.New("argument is empty")
		}
		host, port, er := net.SplitHostPort(hp)
		if er != nil {
			return er
		}
		srvConf.Host = host
		srvConf.Port = port
		return nil
	}
}

func initBaseURLFunc() func(u string) error {
	return func(u string) error {
		if !validator.URL(u) {
			return errors.New("error validating URL")
		}
		parsed, err := url.Parse(u)
		if err != nil {
			return err
		}

		host, port, er := net.SplitHostPort(parsed.Host)
		if er != nil {
			return er
		}

		srvConf.Scheme = parsed.Scheme
		srvConf.Host = host
		srvConf.Port = port
		srvConf.BasePath = strings.TrimSuffix(parsed.Path, "/")

		srvConf.BaseURL = strings.TrimSuffix(u, "/")
		return nil
	}
}

func initDefault() {
	if srvConf.Scheme == "" {
		srvConf.Scheme = defaultProtocol
	}

	if srvConf.Host == "" {
		srvConf.Host = defaultHost
	}

	if srvConf.Port == "" {
		srvConf.Port = defaultPort
	}

	if srvConf.BaseURL == "" {
		srvConf.BaseURL = fmt.Sprintf("%s://%s:%s%s", srvConf.Scheme, srvConf.Host, srvConf.Port, srvConf.BasePath)
	}
}

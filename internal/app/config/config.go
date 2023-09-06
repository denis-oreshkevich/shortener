package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/util/validator"
	"net"
	"net/url"
	"strings"
)

const (
	defaultHost = "localhost"

	defaultPort = "8080"

	defaultProtocol = "http"
)

var srvConf *ServerConf

type ServerConf struct {
	Protocol string
	Host     string
	Port     string
	BaseURL  string
	BasePath string
}

func Get() ServerConf {
	fmt.Printf("Server configuration: %v\n", *srvConf)
	return *srvConf
}

func init() {
	srvConf = &ServerConf{}
	flag.Func("b", "HTTP server base URL path", initB())
	flag.Func("a", "HTTP server address", initA())

	if srvConf.Protocol == "" {
		srvConf.Protocol = defaultProtocol
	}

	if srvConf.Host == "" {
		srvConf.Host = defaultHost
	}

	if srvConf.Port == "" {
		srvConf.Port = defaultPort
	}

	if srvConf.BaseURL == "" {
		srvConf.BaseURL = fmt.Sprintf("%s://%s:%s%s", srvConf.Protocol, srvConf.Host, srvConf.Port, srvConf.BasePath)
	}
}

func initB() func(s string) error {
	return func(s string) error {
		if !validator.URL(s) {
			return errors.New("error validating URL")
		}
		parsed, err := url.Parse(s)
		if err != nil {
			return err
		}

		host, port, er := net.SplitHostPort(parsed.Host)
		if er != nil {
			return er
		}

		srvConf.Protocol = parsed.Scheme
		srvConf.Host = host
		srvConf.Port = port
		srvConf.BasePath = strings.TrimSuffix(parsed.Path, "/")

		srvConf.BaseURL = strings.TrimSuffix(s, "/")

		return nil
	}
}

func initA() func(s string) error {
	return func(s string) error {
		hp := strings.Split(s, ":")
		if len(hp) != 2 {
			return errors.New("need address in a form host:port")
		}
		srvConf.Host = hp[0]
		srvConf.Port = hp[1]
		return nil
	}
}

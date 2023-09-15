package config

type ServerConf struct {
	scheme   string
	host     string
	port     string
	baseURL  string
	basePath string
}

func (s ServerConf) Scheme() string {
	return s.scheme
}

func (s ServerConf) Host() string {
	return s.host
}

func (s ServerConf) Port() string {
	return s.port
}

func (s ServerConf) BaseURL() string {
	return s.baseURL
}

func (s ServerConf) BasePath() string {
	return s.basePath
}

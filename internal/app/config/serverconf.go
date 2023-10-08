package config

type Conf struct {
	scheme   string
	host     string
	port     string
	baseURL  string
	basePath string
	fsPath   string
}

func (s Conf) Scheme() string {
	return s.scheme
}

func (s Conf) Host() string {
	return s.host
}

func (s Conf) Port() string {
	return s.port
}

func (s Conf) BaseURL() string {
	return s.baseURL
}

func (s Conf) BasePath() string {
	return s.basePath
}

func (s Conf) FsPath() string {
	return s.fsPath
}

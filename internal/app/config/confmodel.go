package config

// Conf model that represents a configuration from ENV or command line.
type Conf struct {
	scheme      string
	host        string
	port        string
	baseURL     string
	basePath    string
	fsPath      string
	databaseDSN string
	enableHTTPS bool
}

// Scheme getter for field scheme.
func (s Conf) Scheme() string {
	return s.scheme
}

// Host getter for field host.
func (s Conf) Host() string {
	return s.host
}

// Port getter for field port.
func (s Conf) Port() string {
	return s.port
}

// BaseURL getter for field baseURL.
func (s Conf) BaseURL() string {
	return s.baseURL
}

// BasePath getter for field basePath.
func (s Conf) BasePath() string {
	return s.basePath
}

// FsPath getter for field fsPath.
func (s Conf) FsPath() string {
	return s.fsPath
}

// DatabaseDSN getter for field databaseDSN.
func (s Conf) DatabaseDSN() string {
	return s.databaseDSN
}

// EnableHTTPS getter for field enableHTTPS.
func (s Conf) EnableHTTPS() bool {
	return s.enableHTTPS
}

package server

// URLModel model represents the URL in JSON format.
type URLModel struct {
	URL string `json:"url"`
}

// NewURL creates new [URLModel].
func NewURL(url string) URLModel {
	return URLModel{URL: url}
}

// ResultModel model represents the result URL in JSON format.
type ResultModel struct {
	Result string `json:"result"`
}

// NewResult creates new [ResultModel].
func NewResult(res string) ResultModel {
	return ResultModel{Result: res}
}

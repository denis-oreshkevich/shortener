package handler

type URLModel struct {
	URL string `json:"url"`
}

func NewURL(url string) URLModel {
	return URLModel{URL: url}
}

type ResultModel struct {
	Result string `json:"result"`
}

func NewResult(res string) ResultModel {
	return ResultModel{Result: res}
}

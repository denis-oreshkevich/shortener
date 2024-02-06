package model

type Stat struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

func NewStat(urls, users int) Stat {
	return Stat{
		URLs:  urls,
		Users: users,
	}
}

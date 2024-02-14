package model

// Stat model to return statistic.
type Stat struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// NewStat returns new Stat model.
func NewStat(urls, users int) Stat {
	return Stat{
		URLs:  urls,
		Users: users,
	}
}

package data

import "time"

type Challenge struct {
	Id        int       `json:"id_ch"`
	Name      string    `json:"name"`
	Desc      string    `json:"desc"`
	Rules     string    `json:"rules"`
	Status    string    `json:"status"`
	Points    string    `json:"points"`
	CreatedAt time.Time `json:"created_at"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Photo     string    `json:"photo"`
	File      string    `json:"file"`
	Accepted  bool      `json:"accepted"`
	Type      string    `json:"type"`
	Mail      string    `json:"mail"`
}

type LocalCh struct {
	Id     int       `json:"id_ch"`
	Name   string    `json:"name"`
	Desc   string    `json:"desc"`
	UserId int       `json:"id_i"`
	Status string    `json:"status"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Mail   string    `json:"mail"`
}

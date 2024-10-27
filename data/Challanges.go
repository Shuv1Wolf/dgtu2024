package data

import (
	"encoding/json"
	"time"
)

type Challenge struct {
	Id           int       `json:"id_ch"`
	Name         string    `json:"name"`
	Desc         string    `json:"desc"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	Accepted     bool      `json:"accepted"`
	Type         string    `json:"type"`
	Creator      string    `json:"creator"`
	Interest     string    `json:"interest"`
	Steps        int       `json:"steps"`
	Sleep_millis int       `json:"sleep_millis"`
	UserId       int       `json:"id_u"`
}

type LocalCh struct {
	Id          int       `json:"id_g"`
	Name        string    `json:"name"`
	Desc        string    `json:"desc"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Steps       int       `json:"steps"`
	SleepMillis int       `json:"sleep_millis"`
	UserId      int       `json:"id_u"`
}

const timeFormat = "2006-01-02 15:04:05.999999-07:00"

func (c *Challenge) UnmarshalJSON(data []byte) error {
	type Alias Challenge
	aux := &struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	start, err := time.Parse(timeFormat, aux.Start)
	if err != nil {
		return err
	}
	end, err := time.Parse(timeFormat, aux.End)
	if err != nil {
		return err
	}

	c.Start = start
	c.End = end
	return nil
}

func (c *LocalCh) UnmarshalJSON(data []byte) error {
	type Alias LocalCh
	aux := &struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	start, err := time.Parse(timeFormat, aux.Start)
	if err != nil {
		return err
	}
	end, err := time.Parse(timeFormat, aux.End)
	if err != nil {
		return err
	}

	c.Start = start
	c.End = end
	return nil
}

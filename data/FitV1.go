package data

import "golang.org/x/oauth2"

type FitV1 struct {
	Id    string        `json:"id"`
	Mail  string        `json:"mail"`
	Token *oauth2.Token `json:"token"`
}

func (k FitV1) Clone() FitV1 {
	return FitV1{
		Id:    k.Id,
		Mail:  k.Mail,
		Token: k.Token,
	}
}

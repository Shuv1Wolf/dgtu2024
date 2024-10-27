package client1

import (
	"context"
	"encoding/json"
	"hack/data"
	"io/ioutil"

	cdata "github.com/pip-services4/pip-services4-go/pip-services4-commons-go/data"
	"github.com/pip-services4/pip-services4-go/pip-services4-http-go/clients"
)

type ChallengeMap map[string]data.Challenge

type UserEmail struct {
	Email string `json:"email"`
}

type BackHttpClientV1 struct {
	clients.RestClient
}

func NewBackHttpClientV1() *BackHttpClientV1 {
	client := BackHttpClientV1{}
	client.RestClient = *clients.NewRestClient()

	return &client
}

func (c *BackHttpClientV1) GetChallengesByMails(ctx context.Context, traceId string, mails []string) (ch map[string][]data.Challenge, err error) {

	var emailObjects []map[string]string
	for _, email := range mails {
		emailObjects = append(emailObjects, map[string]string{"email": email})
	}

	calValue, calErr := c.Call(ctx, "post", "statuses/challenges_by_emails", cdata.NewEmptyStringValueMap(), emailObjects)
	if calErr != nil {
		return nil, calErr
	}

	defer calValue.Body.Close()

	body, err := ioutil.ReadAll(calValue.Body)
	if err != nil {
		return nil, err
	}

	var data map[string][]data.Challenge
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *BackHttpClientV1) GetGoalsByMails(ctx context.Context, traceId string, mails []string) (ch map[string][]data.LocalCh, err error) {

	var emailObjects []map[string]string
	for _, email := range mails {
		emailObjects = append(emailObjects, map[string]string{"email": email})
	}

	calValue, calErr := c.Call(ctx, "post", "statuses/goals_by_emails", cdata.NewEmptyStringValueMap(), emailObjects)
	if calErr != nil {
		return nil, calErr
	}

	defer calValue.Body.Close()

	body, err := ioutil.ReadAll(calValue.Body)
	if err != nil {
		return nil, err
	}

	var data map[string][]data.LocalCh
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type goalsStatus struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

func (c *BackHttpClientV1) PatchStatusGoals(ctx context.Context, traceId string, id int, status string) (err error) {

	req := goalsStatus{
		Id:     id,
		Status: status,
	}
	_, calErr := c.Call(ctx, "patch", "statuses/", cdata.NewEmptyStringValueMap(), req)
	if calErr != nil {
		return calErr
	}

	return nil
}

type achievement struct {
	UserId int `json:"id_u"`
	IdCh   int `json:"id_gach"`
}

func (c *BackHttpClientV1) AddAchievement(ctx context.Context, traceId string, userId int, IdCh int) (err error) {

	req := achievement{
		UserId: userId,
		IdCh:   IdCh,
	}
	_, calErr := c.Call(ctx, "post", "challenges/add_achievement", cdata.NewEmptyStringValueMap(), req)
	if calErr != nil {
		return calErr
	}

	return nil
}

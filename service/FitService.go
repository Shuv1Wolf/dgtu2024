package service

import (
	"context"
	"errors"
	"fmt"
	client1 "hack/client/version1"
	"hack/data"
	persist "hack/persistence"
	"net/http"
	"os"
	"time"

	cconf "github.com/pip-services4/pip-services4-go/pip-services4-components-go/config"
	exec "github.com/pip-services4/pip-services4-go/pip-services4-components-go/exec"
	cref "github.com/pip-services4/pip-services4-go/pip-services4-components-go/refer"
	"github.com/pip-services4/pip-services4-go/pip-services4-data-go/query"
	log "github.com/pip-services4/pip-services4-go/pip-services4-observability-go/log"
	ccmd "github.com/pip-services4/pip-services4-go/pip-services4-rpc-go/commands"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/fitness/v1"
)

const (
	NONSTARTED = "not started"
	PROGRESS   = "in progress"
	FAILED     = "failed"
	COMPLETED  = "completed"
)

type FitService struct {
	persistence persist.IFitPersistence
	client      client1.BackHttpClientV1
	commandSet  *FitCommandSet

	oauthConfig *oauth2.Config

	timer  exec.FixedRateTimer
	Logger *log.CompositeLogger

	challenges      map[string][]data.Challenge
	localChallenges map[string][]data.LocalCh

	exist map[string]interface{}
}

func NewFitService() *FitService {
	c := &FitService{}
	c.Logger = log.NewCompositeLogger()

	client := client1.NewBackHttpClientV1()

	httpConfig := cconf.NewConfigParamsFromTuples(
		"connection.protocol", "http",
		"connection.port", "8001",
		"connection.host", "89.46.131.17",
	)
	client.Configure(context.Background(), httpConfig)
	client.Open(context.Background())

	c.client = *client

	redirectPort := os.Getenv("HTTP_REDIRECT_PORT")

	c.timer = *exec.NewFixedRateTimerFromCallback(func(ctx context.Context) {
		users, err := c.persistence.GetPage(ctx)
		if err != nil || len(users.Data) == 0 {
			return
		}

		mails := make([]string, 0, len(users.Data))

		for _, user := range users.Data {
			mails = append(mails, user.Mail)
		}

		ch, err := c.client.GetChallengesByMails(context.Background(), "trace", mails)
		if err != nil || len(users.Data) == 0 {
			return
		}
		lCh, err := c.client.GetGoalsByMails(context.Background(), "trace", mails)
		if err != nil || len(users.Data) == 0 {
			return
		}

		c.challenges = ch
		c.localChallenges = lCh

		c.worker(ctx, users)

	}, 15000, 0, 1)

	go func() {
		http.HandleFunc("/v1/fit/callback", c.callbackHandler)
		fmt.Println("CallbackHandler :" + redirectPort)
		http.ListenAndServe(":"+redirectPort, nil)
	}()

	c.timer.Start(context.Background())

	return c
}

func (c *FitService) Configure(ctx context.Context, config *cconf.ConfigParams) {
	// Read configuration parameters here...
}

func (c *FitService) SetReferences(ctx context.Context, references cref.IReferences) {
	locator := cref.NewDescriptor("fit", "persistence", "*", "*", "1.0")
	p, err := references.GetOneRequired(locator)
	if p != nil && err == nil {
		if _pers, ok := p.(persist.IFitPersistence); ok {
			c.persistence = _pers
			return
		}
	}
	panic(cref.NewReferenceError(ctx, locator))
}

func (c *FitService) GetCommandSet() *ccmd.CommandSet {
	if c.commandSet == nil {
		c.commandSet = NewFitCommandSet(c)
	}
	return &c.commandSet.CommandSet
}

func (c *FitService) callbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		http.Error(w, "state or code not found", http.StatusBadRequest)
		c.Logger.Info(context.Background(), "state or code not found")
		return
	}
	mail := state

	token, err := c.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		c.Logger.Info(context.Background(), "failed to exchange token")
		return
	}

	fmt.Fprintf(w, "Authorization successful for %s\n", mail)

	data := data.FitV1{
		Mail:  mail,
		Token: token,
	}

	c.persistence.Create(context.Background(), data)
}

func (c *FitService) GoogleAuthorization(ctx context.Context, mail string) (string, error) {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURL := os.Getenv("REDIRECT_URL")

	c.oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{fitness.FitnessActivityReadScope,
			"https://www.googleapis.com/auth/fitness.sleep.read"},
		Endpoint: google.Endpoint,
	}

	state := mail
	url := c.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	if len(url) == 0 {
		return "", errors.New("google link is empty")
	}

	return url, nil
}

func (c *FitService) worker(ctx context.Context, users query.DataPage[data.FitV1]) {

	for _, user := range users.Data {
		ch := c.challenges[user.Mail]
		for _, i := range ch {
			if i.Type == "step" || i.Type == "steps" {

				if time.Now().UTC().Before(i.Start) {
					c.Logger.Debug(ctx, "не начат")
					continue
				}

				steps, err := c.fetchStepData(user.Token, i.Start, i.End)
				if err != nil {
					c.Logger.Error(ctx, err, "")
					c.persistence.DeleteById(ctx, user.Id)
				}

				if steps >= int64(i.Steps) {
					c.Logger.Debug(ctx, "выполнен")

					_, exist := c.exist[fmt.Sprintf("%d_%d", i.UserId, i.Id)]
					if exist {
						continue
					}

					c.client.AddAchievement(ctx, "trace", i.UserId, i.Id)
					c.exist[fmt.Sprintf("%d_%d", i.UserId, i.Id)] = nil

				} else if i.End.Before(time.Now().UTC()) {
					c.Logger.Debug(ctx, "провален")
				} else {
					c.Logger.Debug(ctx, "порогресс")
				}
			}

			if i.Type == "sleep" {

				if time.Now().UTC().Before(i.Start) {
					c.Logger.Debug(ctx, "НЕ НАЧАТ")
					continue
				}

				sleep, err := c.fetchSleepData(user.Token, i.Start, i.End)
				if err != nil {
					c.Logger.Error(ctx, err, "")
					c.persistence.DeleteById(ctx, user.Id)
				}

				if sleep >= int64(i.Sleep_millis) {
					c.Logger.Debug(ctx, "выполнен")

					_, exist := c.exist[fmt.Sprintf("%d_%d", i.UserId, i.Id)]
					if exist {
						continue
					}

					c.client.AddAchievement(ctx, "trace", i.UserId, i.Id)
					c.exist[fmt.Sprintf("%d_%d", i.UserId, i.Id)] = nil

				} else if i.End.Before(time.Now().UTC()) {
					c.Logger.Debug(ctx, "провален")
				} else {
					c.Logger.Debug(ctx, "порогресс")
				}
			}
		}

		lch := c.localChallenges[user.Mail]
		for _, i := range lch {
			if i.Type == "step" || i.Type == "steps" {

				if time.Now().UTC().Before(i.Start) {
					c.Logger.Debug(ctx, "НЕ НАЧАТ")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, NONSTARTED)
					continue
				}

				steps, err := c.fetchStepData(user.Token, i.Start, i.End)
				if err != nil {
					c.Logger.Error(ctx, err, "")
					c.persistence.DeleteById(ctx, user.Id)
				}

				if steps >= int64(i.Steps) {
					c.Logger.Debug(ctx, "выполнен")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, COMPLETED)
				} else if i.End.Before(time.Now().UTC()) {
					c.Logger.Debug(ctx, "провален")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, FAILED)
				} else {
					c.Logger.Debug(ctx, "порогресс")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, PROGRESS)
				}
			}

			if i.Type == "sleep" {

				if time.Now().UTC().Before(i.Start) {
					c.Logger.Debug(ctx, "НЕ НАЧАТ")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, NONSTARTED)
					continue
				}

				sleep, err := c.fetchSleepData(user.Token, i.Start, i.End)
				if err != nil {
					c.Logger.Error(ctx, err, "")
					c.persistence.DeleteById(ctx, user.Id)
				}

				if sleep >= int64(i.SleepMillis) {
					c.Logger.Debug(ctx, "выполнен")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, COMPLETED)
				} else if i.End.Before(time.Now().UTC()) {
					c.Logger.Debug(ctx, "провален")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, FAILED)
				} else {
					c.Logger.Debug(ctx, "порогресс")
					c.client.PatchStatusGoals(ctx, "trace", i.Id, PROGRESS)
				}
			}
		}
	}
}

func (c *FitService) fetchStepData(token *oauth2.Token, start time.Time, end time.Time) (int64, error) {
	client := c.oauthConfig.Client(context.Background(), token)
	fitnessService, err := fitness.New(client)
	if err != nil {
		return 0, err
	}

	startTimeMillis := start.UnixNano() / int64(time.Millisecond)
	endTimeMillis := end.UnixNano() / int64(time.Millisecond)

	request := &fitness.AggregateRequest{
		AggregateBy: []*fitness.AggregateBy{
			{DataTypeName: "com.google.step_count.delta"},
		},
		StartTimeMillis: startTimeMillis,
		EndTimeMillis:   endTimeMillis,
	}

	response, err := fitnessService.Users.Dataset.Aggregate("me", request).Do()
	if err != nil {
		return 0, err
	}

	totalSteps := int64(0)
	for _, bucket := range response.Bucket {
		for _, dataset := range bucket.Dataset {
			for _, point := range dataset.Point {
				if len(point.Value) > 0 {
					totalSteps += point.Value[0].IntVal
				}
			}
		}
	}
	return totalSteps, nil
}

func (c *FitService) fetchSleepData(token *oauth2.Token, start time.Time, end time.Time) (int64, error) {
	client := c.oauthConfig.Client(context.Background(), token)
	fitnessService, err := fitness.New(client)
	if err != nil {
		return 0, err
	}

	startTimeMillis := start.UnixNano() / int64(time.Millisecond)
	endTimeMillis := end.UnixNano() / int64(time.Millisecond)

	request := &fitness.AggregateRequest{
		AggregateBy: []*fitness.AggregateBy{
			{DataTypeName: "com.google.sleep.segment"},
		},
		StartTimeMillis: startTimeMillis,
		EndTimeMillis:   endTimeMillis,
	}

	response, err := fitnessService.Users.Dataset.Aggregate("me", request).Do()
	if err != nil {
		return 0, err
	}

	totalSleepMillis := int64(0)
	for _, bucket := range response.Bucket {
		for _, dataset := range bucket.Dataset {
			for _, point := range dataset.Point {
				if len(point.Value) > 0 && point.StartTimeNanos != 0 && point.EndTimeNanos != 0 {
					duration := (point.EndTimeNanos - point.StartTimeNanos) / int64(time.Millisecond)
					totalSleepMillis += duration
				}
			}
		}
	}

	return totalSleepMillis, nil
}

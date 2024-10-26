package service

import (
	"context"
	"errors"
	"fmt"
	"hack/data"
	persist "hack/persistence"
	"net/http"
	"os"
	"time"

	cconf "github.com/pip-services4/pip-services4-go/pip-services4-components-go/config"
	exec "github.com/pip-services4/pip-services4-go/pip-services4-components-go/exec"
	cref "github.com/pip-services4/pip-services4-go/pip-services4-components-go/refer"
	log "github.com/pip-services4/pip-services4-go/pip-services4-observability-go/log"
	ccmd "github.com/pip-services4/pip-services4-go/pip-services4-rpc-go/commands"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/fitness/v1"
)

type FitService struct {
	persistence persist.IFitPersistence
	commandSet  *FitCommandSet

	oauthConfig *oauth2.Config

	timer  exec.FixedRateTimer
	Logger *log.CompositeLogger

	challenges      map[string]data.Challenge
	localChallenges map[string]data.LocalCh
}

func NewFitService() *FitService {
	c := &FitService{}
	c.Logger = log.NewCompositeLogger()
	redirectPort := os.Getenv("HTTP_REDIRECT_PORT")

	c.timer = *exec.NewFixedRateTimerFromCallback(func(ctx context.Context) {
		c.updateChalenges(ctx)
		c.worker(ctx)

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
	fmt.Fprintf(w, "Access token: %s\n", token.AccessToken)

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

func (c *FitService) worker(ctx context.Context) {
	users, err := c.persistence.GetPage(context.Background())
	if err != nil || len(users.Data) == 0 {
		return
	}

	// TODO:
	for _, user := range users.Data {
		steps, err := c.fetchStepData(user.Token, time.Now().Add(-50*time.Hour), time.Now())
		if err != nil {
			c.Logger.Error(ctx, err, "")
			c.persistence.DeleteById(ctx, user.Id)
		}
		c.Logger.Info(ctx, fmt.Sprintf("%d", steps))

		sleep, err := c.fetchSleepData(user.Token, time.Now().Add(-50*time.Hour), time.Now())
		if err != nil {
			c.Logger.Error(ctx, err, "")
			c.persistence.DeleteById(ctx, user.Id)
		}
		c.Logger.Info(ctx, fmt.Sprintf("%d", sleep))
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

func (c *FitService) updateChalenges(context.Context) {
	// TODO:
	ch := make([]data.Challenge, 100, 100)
	chMap := make(map[string]data.Challenge, len(ch))

	for _, challenge := range ch {
		chMap[challenge.Mail] = challenge
	}
	c.challenges = chMap

	// TODO:
	locCh := make([]data.LocalCh, 100, 100)
	locChMap := make(map[string]data.LocalCh, len(locCh))

	for _, challenge := range locCh {
		locChMap[challenge.Mail] = challenge
	}
	c.challenges = chMap
}

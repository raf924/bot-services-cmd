package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/raf924/bot/pkg/bot/command"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type joke struct {
	Kind string `json:"kind"`
	Data struct {
		Selftext string `json:"selftext"`
		Title    string `json:"title"`
		Id       string `json:"id"`
		Url      string `json:"url"`
	} `json:"data"`
}

type jokesResponse struct {
	Data struct {
		Count    int    `json:"dist"`
		Children []joke `json:"children"`
	} `json:"data"`
}

type JokeCommand struct {
	command.NoOpInterceptor
	jokeSources []func() (string, error)
}

func (j *JokeCommand) Init(_ command.Executor) error {
	j.jokeSources = append(j.jokeSources, j.fetchDadJoke, j.fetchFromReddit)
	return nil
}

func (j *JokeCommand) Name() string {
	return "joke"
}

func (j *JokeCommand) Aliases() []string {
	return []string{"j"}
}

func (j *JokeCommand) fetchFromReddit() (string, error) {
	req, _ := http.NewRequest("GET", "https://www.reddit.com/r/jokes/.json", nil)
	req.Header.Set("TE", "Trailers")
	req.Header.Set("Accept", "text/json,application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req.Header.Set("Host", "www.reddit.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request status code: %d: %s", resp.StatusCode, resp.Status)
	}
	var jr jokesResponse
	err = json.NewDecoder(resp.Body).Decode(&jr)
	if err != nil {
		return "", err
	}
	jokeIndex := rand.Intn(jr.Data.Count)
	joke := jr.Data.Children[jokeIndex].Data
	return fmt.Sprintf("%s\n%s - %s", joke.Title, joke.Selftext, joke.Url), nil
}

func (j *JokeCommand) fetchDadJoke() (string, error) {
	req, _ := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("User-Agent", "TBotT (https://github.com/raf924/bot-services-cmd)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request status code: %d: %s", resp.StatusCode, resp.Status)

	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (j *JokeCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	var jokeSources = []func() (string, error){j.fetchDadJoke, j.fetchFromReddit}
	rand.Seed(time.Now().Unix())
	err := fmt.Errorf("no joke source chosen")
	var joke string
	for len(jokeSources) > 0 && err != nil {
		sourceChoice := rand.Intn(len(jokeSources))
		joke, err = jokeSources[sourceChoice]()
		if err != nil {
			jokeSources = append(jokeSources[:sourceChoice], jokeSources[sourceChoice+1:]...)
		}
	}
	if err != nil {
		return nil, err
	}

	return []*messages.BotPacket{
		{
			Timestamp: timestamppb.Now(),
			Message:   joke,
			Recipient: command.GetUser(),
			Private:   command.GetPrivate(),
		},
	}, nil
}

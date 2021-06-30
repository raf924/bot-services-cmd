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
	Count int `json:"dist"`
	Data  struct {
		Children []joke `json:"children"`
	} `json:"data"`
}

type JokeCommand struct {
	command.NoOpInterceptor
}

func (j *JokeCommand) Init(_ command.Executor) error {
	return nil
}

func (j *JokeCommand) Name() string {
	return "joke"
}

func (j *JokeCommand) Aliases() []string {
	return []string{"j"}
}

func (j *JokeCommand) fetchFromReddit() (string, error) {
	resp, err := http.DefaultClient.Get("https://www.reddit.com/r/jokes/.json")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("")

	}
	var jr jokesResponse
	err = json.NewDecoder(resp.Body).Decode(&jr)
	if err != nil {
		return "", err
	}
	jokeIndex := rand.Intn(jr.Count)
	joke := jr.Data.Children[jokeIndex].Data
	return fmt.Sprintf("%s\n%s - %s", joke.Title, joke.Selftext, joke.Url), nil
}

func (j *JokeCommand) fetchDadJoke() (string, error) {
	req, _ := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	req.Header.Set("User-Agent", "TBotT (https://github.com/raf924/bot-services-cmd)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("")

	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (j *JokeCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	rand.Seed(time.Now().Unix())
	sourceChoice := rand.Intn(2)
	var joke string
	var err error
	switch sourceChoice {
	case 0:
		joke, err = j.fetchFromReddit()
	case 1:
		joke, err = j.fetchDadJoke()
	default:
		err = fmt.Errorf("invalid choice")
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
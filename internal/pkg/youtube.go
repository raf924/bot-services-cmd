package pkg

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/raf924/bot/pkg/bot/command"
	"github.com/raf924/bot/pkg/domain"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
	"regexp"
)

var _ command.Command = (*YoutubeCommand)(nil)

var ytRegex = regexp.MustCompile(`(?i)(?:youtube\.com/\S*(?:(?:/e(?:mbed))?/|watch\?(?:\S*?&?v=))|youtu\.be/)([a-zA-Z0-9_-]{6,11})`)

type YoutubeCommand struct {
	command.NoOpCommand
	getter func(videoId string) (string, error)
	apiKey string
}

func (y *YoutubeCommand) getByApi(videoId string) (string, error) {
	yS, err := youtube.NewService(context.Background(), option.WithAPIKey(y.apiKey))
	if err != nil {
		log.Println(err)
		return y.getByScraping(videoId)
	}
	resp, err := yS.Videos.List([]string{"snippet"}).Id(videoId).Do()
	if err != nil {
		log.Println(err)
		return y.getByScraping(videoId)
	}
	return resp.Items[0].Snippet.Title, nil
}

func (y *YoutubeCommand) getByScraping(videoId string) (string, error) {
	titleChan := make(chan string)
	c := colly.NewCollector()
	c.OnHTML(".ytp-title-link", func(element *colly.HTMLElement) {
		titleChan <- element.Text
	})
	err := c.Visit(fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId))
	if err != nil {
		return "", err
	}
	return <-titleChan, nil
}

func (y *YoutubeCommand) Init(bot command.Executor) error {
	apiKey, ok := bot.ApiKeys()["youtube"]
	if !ok || len(apiKey) == 0 {
		y.getter = y.getByScraping
	} else {
		y.getter = y.getByApi
		y.apiKey = apiKey
	}
	return nil
}

func (y *YoutubeCommand) Name() string {
	return "youtube"
}

func (y *YoutubeCommand) Aliases() []string {
	return []string{"yt"}
}

func (y *YoutubeCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	return nil, nil
}

func (y *YoutubeCommand) OnChat(message *domain.ChatMessage) ([]*domain.ClientMessage, error) {
	if !ytRegex.MatchString(message.Message()) {
		return nil, nil
	}
	title, err := y.getter(ytRegex.FindAllStringSubmatch(message.Message(), -1)[0][1])
	if err != nil {
		return nil, err
	}
	var recipient *domain.User
	if message.Private() {
		recipient = message.Sender()
	}
	return []*domain.ClientMessage{
		domain.NewClientMessage(fmt.Sprintf("Video: %s", title), recipient, message.Private()),
	}, nil
}

func (y *YoutubeCommand) IgnoreSelf() bool {
	return true
}

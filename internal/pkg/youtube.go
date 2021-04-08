package pkg

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/raf924/bot/pkg/bot/command"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"regexp"
)

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

func (y *YoutubeCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	return nil, nil
}

func (y *YoutubeCommand) OnChat(message *messages.MessagePacket) ([]*messages.BotPacket, error) {
	if !ytRegex.MatchString(message.GetMessage()) {
		return nil, nil
	}
	title, err := y.getter(ytRegex.FindAllStringSubmatch(message.GetMessage(), -1)[0][1])
	if err != nil {
		return nil, err
	}
	var recipient *messages.User
	if message.GetPrivate() {
		recipient = message.GetUser()
	}
	return []*messages.BotPacket{&messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   fmt.Sprintf("Video: %s", title),
		Recipient: recipient,
		Private:   message.GetPrivate(),
	}}, nil
}

func (y *YoutubeCommand) IgnoreSelf() bool {
	return true
}

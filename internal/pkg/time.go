package pkg

import (
	"encoding/xml"
	"fmt"
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/bot"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	bot.HandleCommand(&TimeCommand{})
}

type timeResponse struct {
	XMLName xml.Name `xml:"weatherdata"`
	Weather []struct {
		Location string  `xml:"weatherlocationname,attr"`
		Timezone float64 `xml:"timezone,attr"`
	} `xml:"weather"`
}

type TimeCommand struct {
	WeatherCommand
}

func (t *TimeCommand) Name() string {
	return "time"
}

func (t *TimeCommand) Aliases() []string {
	return []string{"t"}
}

func (t *TimeCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	if len(command.Args) == 0 {
		return nil, fmt.Errorf("missing arguments")
	}
	wUrl, _ := url.Parse(t.weatherUrl.String())
	q := wUrl.Query()
	q.Set("weasearchstr", strings.Join(command.Args, " "))
	wUrl.RawQuery = q.Encode()
	resp, err := http.Get(wUrl.String())
	if err != nil {
		return nil, fmt.Errorf("get weather error: %s", err.Error())
	}
	var timeData timeResponse
	if err = xml.NewDecoder(resp.Body).Decode(&timeData); err != nil {
		return nil, fmt.Errorf("parse weather error: %s", err)
	}
	locTime := time.Now().UTC().In(time.FixedZone("", int(timeData.Weather[0].Timezone*60*60)))
	return []*messages.BotPacket{
		{
			Timestamp: timestamppb.Now(),
			Message:   fmt.Sprintf("%s - %s", locTime.Format("03:04:04 PM"), timeData.Weather[0].Location),
			Recipient: command.GetUser(),
			Private:   command.GetPrivate(),
		},
	}, nil
}

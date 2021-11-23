package pkg

import (
	"encoding/xml"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var _ command.Command = (*TimeCommand)(nil)

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

func (t *TimeCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	if len(command.Args()) == 0 {
		return nil, fmt.Errorf("missing arguments")
	}
	wUrl, _ := url.Parse(t.weatherUrl.String())
	q := wUrl.Query()
	q.Set("weasearchstr", strings.Join(command.Args(), " "))
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
	return []*domain.ClientMessage{
		domain.NewClientMessage(
			fmt.Sprintf("%s - %s", locTime.Format("03:04:04 PM"), timeData.Weather[0].Location),
			command.Sender(),
			command.Private(),
		),
	}, nil
}

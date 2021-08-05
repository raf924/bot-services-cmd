package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/raf924/bot/pkg/bot/command"
	messages "github.com/raf924/connector-api/pkg/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type urbanData struct {
	Definition string `json:"definition"`
	PermaLink  string `json:"permalink"`
}

type urbanResponse struct {
	List []urbanData `json:"list"`
}

type UrbanCommand struct {
	command.NoOpCommand
	urbanRequest *http.Request
}

func (u *UrbanCommand) Init(bot command.Executor) error {
	urbanURL, _ := url.Parse("http://api.urbandictionary.com/v0/define")
	urbanRequest, err := http.NewRequest("GET", urbanURL.String(), nil)
	if err != nil {
		return err
	}
	u.urbanRequest = urbanRequest
	return nil
}

func (u *UrbanCommand) Name() string {
	return "urban"
}

func (u *UrbanCommand) Aliases() []string {
	return []string{"u"}
}

func (u *UrbanCommand) Execute(command *messages.CommandPacket) ([]*messages.BotPacket, error) {
	urbanRequest := *u.urbanRequest
	urbanURL := *urbanRequest.URL
	urbanQuery := urbanURL.Query()
	urbanQuery.Set("term", strings.Join(command.GetArgs(), "+"))
	urbanURL.RawQuery = urbanQuery.Encode()
	urbanRequest.URL = &urbanURL
	netClient := http.Client{Timeout: time.Second * 30}
	response, err := netClient.Do(&urbanRequest)
	if err != nil {
		return nil, err
	}
	var urbanResponse urbanResponse
	err = json.NewDecoder(response.Body).Decode(&urbanResponse)
	if err != nil {
		return nil, err
	}
	if len(urbanResponse.List) == 0 {
		return nil, nil
	}
	definition := urbanResponse.List[0].Definition
	size := int(math.Min(500, float64(len(definition))))
	return []*messages.BotPacket{
		{
			Timestamp: timestamppb.Now(),
			Message:   fmt.Sprintf("%s - %s", urbanResponse.List[0].Definition[0:size], urbanResponse.List[0].PermaLink),
			Recipient: command.GetUser(),
			Private:   command.GetPrivate(),
		},
	}, nil
}

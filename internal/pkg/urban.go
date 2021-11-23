package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var _ command.Command = (*UrbanCommand)(nil)

type urbanData struct {
	Definition string `json:"definition"`
	PermaLink  string `json:"permalink"`
}

type urbanResponse struct {
	List []urbanData `json:"list"`
}

type UrbanCommand struct {
	command.NoOpInterceptor
	urbanRequest *http.Request
}

func (u *UrbanCommand) Init(command.Executor) error {
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

func (u *UrbanCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	urbanRequest := *u.urbanRequest
	urbanURL := *urbanRequest.URL
	urbanQuery := urbanURL.Query()
	urbanQuery.Set("term", strings.Join(command.Args(), "+"))
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
	return []*domain.ClientMessage{
		domain.NewClientMessage(
			fmt.Sprintf("%s - %s", urbanResponse.List[0].Definition[0:size], urbanResponse.List[0].PermaLink),
			command.Sender(),
			command.Private(),
		),
	}, nil
}

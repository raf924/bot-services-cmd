package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"net/http"
	"net/url"
	"strings"
)

var _ command.Command = (*SearchCommand)(nil)

type SearchResponse struct {
	Data struct {
		Result struct {
			Items []struct {
				Title       string `json:"title"`
				Url         string `json:"url"`
				Description string `json:"desc"`
			} `json:"items"`
		} `json:"result"`
	} `json:"data"`
}

type SearchCommand struct {
	command.NoOpCommand
}

func (s *SearchCommand) Init(command.Executor) error {
	return nil
}

func (s *SearchCommand) Name() string {
	return "search"
}

func (s *SearchCommand) Aliases() []string {
	return []string{"s", "g", "google"}
}

func (s *SearchCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	searchTerms := strings.TrimSpace(command.ArgString())
	searchUrl, _ := url.Parse("https://api.qwant.com/api/search/web?locale=en_us&count=3&t=web&uiv=4")
	searchQuery := searchUrl.Query()
	searchQuery.Set("q", searchTerms)
	searchUrl.RawQuery = searchQuery.Encode()
	req, err := http.NewRequest(http.MethodGet, searchUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("error")
	}
	var searchResponse SearchResponse
	err = json.NewDecoder(res.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}
	message := ""
	for _, item := range searchResponse.Data.Result.Items {
		message += fmt.Sprintf("%s %s\n%s\n\n", item.Title, item.Url, item.Description)
	}
	return []*domain.ClientMessage{
		domain.NewClientMessage(message, command.Sender(), command.Private()),
	}, nil
}

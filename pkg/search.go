package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"net/http"
	"net/url"
	"strings"
)

var _ command.Command = (*SearchCommand)(nil)

type MainlineItem struct {
	Title           string        `json:"title"`
	Favicon         string        `json:"favicon"`
	Url             string        `json:"url"`
	UrlPingSuffix   string        `json:"urlPingSuffix,omitempty"`
	Source          string        `json:"source"`
	Desc            string        `json:"desc"`
	Id              string        `json:"_id"`
	Type            string        `json:"_type,omitempty"`
	AdType          string        `json:"ad_type,omitempty"`
	Position        int           `json:"position,omitempty"`
	RawPosition     string        `json:"raw_position,omitempty"`
	ImpressionToken string        `json:"impressionToken,omitempty"`
	Images          []interface{} `json:"images,omitempty"`
	ActionItems     []interface{} `json:"actionItems,omitempty"`
	PriceItems      []interface{} `json:"priceItems,omitempty"`
	Links           []interface{} `json:"links,omitempty"`
}

type MainlineResult struct {
	Type  string         `json:"type"`
	Count int            `json:"count"`
	Items []MainlineItem `json:"items"`
}

type SearchResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result struct {
			Items struct {
				Mainline []MainlineResult `json:"mainline"`
			} `json:"items"`
		} `json:"result"`
	} `json:"data"`
}

type SearchCommand struct {
	command.NoOpInterceptor
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

func findWebResults(result []MainlineResult) ([]MainlineResult, error) {
	var webResults []MainlineResult
	for _, item := range result {
		if item.Type == "web" {
			webResults = append(webResults, item)
		}
	}
	if len(webResults) == 0 {
		return nil, fmt.Errorf("no web result found")
	}
	return webResults, nil
}

func (s *SearchCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	searchTerms := strings.TrimSpace(command.ArgString())
	searchUrl, _ := url.Parse("https://api.qwant.com/v3/search/web?origin=suggest&count=10&offset=0&safesearch=1&locale=en_US")
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
		return nil, fmt.Errorf("call is error %d, %s", res.StatusCode, res.Status)
	}
	var searchResponse SearchResponse
	err = json.NewDecoder(res.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}
	message := ""
	webResults, err := findWebResults(searchResponse.Data.Result.Items.Mainline)
	if err != nil {
		return nil, err
	}
	count := 0
	for _, webResult := range webResults {
		for _, item := range webResult.Items {
			message += fmt.Sprintf(">#### [%s](%s)\n%s\n\n", item.Title, item.Url, item.Desc)
			count += 1
			if count == 5 {
				break
			}
		}
		if count == 5 {
			break
		}
	}
	return []*domain.ClientMessage{
		domain.NewClientMessage(message, command.Sender(), command.Private()),
	}, nil
}

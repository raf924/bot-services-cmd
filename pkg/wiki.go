package pkg

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _ command.Command = (*WikiCommand)(nil)

type WikiSearchResponse struct {
	XMLName xml.Name `xml:"api"`
	Query   struct {
		Suggestion []struct {
			Title  string `xml:"title,attr"`
			PageId int    `xml:"pageid,attr"`
		} `xml:"search>p"`
	} `xml:"query"`
}

type WikiInfoResponse struct {
	XMLName xml.Name `xml:"api"`
	Query   struct {
		Pages []struct {
			Page struct {
				Url string `xml:"fullurl,attr"`
			} `xml:"page"`
		} `xml:"pages"`
	} `xml:"query"`
}

type WikiExtractResponse struct {
	XMLName xml.Name `xml:"api"`
	Query   struct {
		Pages []struct {
			Page struct {
				Extract string `xml:"extract"`
			} `xml:"page"`
		} `xml:"pages"`
	} `xml:"query"`
}

type WikiCommand struct {
	command.NoOpInterceptor
	wikiUrl *url.URL
}

func (w *WikiCommand) Init(command.Executor) error {
	wikiUrl, err := url.Parse("https://en.wikipedia.org/w/api.php?action=query&format=xml")
	if err != nil {
		return err
	}
	w.wikiUrl = wikiUrl
	return nil
}

func (w *WikiCommand) Name() string {
	return "wiki"
}

func (w *WikiCommand) Aliases() []string {
	return []string{}
}

func (w *WikiCommand) search(search string) (*WikiSearchResponse, error) {
	queryURL := *w.wikiUrl
	wikiQuery := queryURL.Query()
	wikiQuery.Set("srsearch", search)
	wikiQuery.Set("list", "search")
	queryURL.RawQuery = wikiQuery.Encode()
	netClient := http.Client{Timeout: time.Second * 30}
	log.Println(queryURL.String())
	wikiRequest, err := http.NewRequest(http.MethodGet, queryURL.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := netClient.Do(wikiRequest)
	if err != nil {
		return nil, err
	}
	var searchResponse WikiSearchResponse
	err = xml.NewDecoder(response.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (w *WikiCommand) info(pageId int) (*WikiInfoResponse, error) {
	var infoResponse WikiInfoResponse
	infoUrl := *w.wikiUrl
	wikiQuery := infoUrl.Query()
	wikiQuery.Set("prop", "info")
	wikiQuery.Set("inprop", "url")
	wikiQuery.Set("pageids", strconv.Itoa(pageId))
	infoUrl.RawQuery = wikiQuery.Encode()
	wikiRequest, err := http.NewRequest(http.MethodGet, infoUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	netClient := http.Client{Timeout: time.Second * 30}
	resp, err := netClient.Do(wikiRequest)
	if err != nil {
		return nil, err
	}
	err = xml.NewDecoder(resp.Body).Decode(&infoResponse)
	if err != nil {
		return nil, err
	}
	return &infoResponse, nil
}

func (w *WikiCommand) extract(pageId int) (*WikiExtractResponse, error) {
	var extractResponse WikiExtractResponse
	extractUrl := *w.wikiUrl
	wikiQuery := extractUrl.Query()
	wikiQuery.Set("prop", "extracts")
	wikiQuery.Set("explaintext", "")
	wikiQuery.Set("exintro", "")
	wikiQuery.Set("pageids", strconv.Itoa(pageId))
	extractUrl.RawQuery = wikiQuery.Encode()
	wikiRequest, err := http.NewRequest(http.MethodGet, extractUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	netClient := http.Client{Timeout: time.Second * 30}
	resp, err := netClient.Do(wikiRequest)
	if err != nil {
		return nil, err
	}
	err = xml.NewDecoder(resp.Body).Decode(&extractResponse)
	if err != nil {
		return nil, err
	}
	return &extractResponse, nil
}

func (w *WikiCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	searchResponse, err := w.search(strings.Join(command.Args(), " "))
	if err != nil {
		return nil, err
	}
	if len(searchResponse.Query.Suggestion) == 0 {
		return nil, errors.New("no result")
	}
	extractResponse, err := w.extract(searchResponse.Query.Suggestion[0].PageId)
	if err != nil {
		return nil, err
	}
	infoResponse, err := w.info(searchResponse.Query.Suggestion[0].PageId)
	if err != nil {
		return nil, err
	}
	title := searchResponse.Query.Suggestion[0].Title
	summary := extractResponse.Query.Pages[0].Page.Extract
	link := infoResponse.Query.Pages[0].Page.Url
	reply := fmt.Sprintf("[%s](%s)\n>%s", title, link, summary)
	return []*domain.ClientMessage{
		domain.NewClientMessage(reply, command.Sender(), command.Private()),
	}, nil
}

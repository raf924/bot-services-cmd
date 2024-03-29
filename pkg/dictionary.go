package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var _ command.Command = (*DictionaryCommand)(nil)

type DictionaryResponse []struct {
	Word      string `json:"word"`
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
	Meanings []struct {
		SpeechPart  string `json:"partOfSpeech"`
		Definitions []struct {
			Definition string `json:"definition"`
			Example    string `json:"example"`
		} `json:"definitions"`
	} `json:"meanings"`
}

type DictionaryCommand struct {
	command.NoOpInterceptor
	dictionaryUrl *url.URL
}

func (d *DictionaryCommand) Init(command.Executor) error {
	dU, err := url.Parse("https://api.dictionaryapi.dev")
	if err != nil {
		return err
	}
	d.dictionaryUrl = dU
	return nil
}

func (d *DictionaryCommand) Name() string {
	return "dictionary"
}

func (d *DictionaryCommand) Aliases() []string {
	return []string{"d", "define", "dict"}
}

func (d *DictionaryCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	dU := *d.dictionaryUrl
	dU.Path = path.Join("api", "v2", "entries", "en", url.PathEscape(strings.Join(command.Args(), " ")))
	client := http.Client{Timeout: time.Second * 5}
	log.Println(dU.String())
	resp, err := client.Get(dU.String())
	if err != nil {
		return nil, err
	}
	meaningText := ""
	var dictionaryResponse DictionaryResponse
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Println(string(b))
	err = json.Unmarshal(b, &dictionaryResponse)
	if err != nil {
		return nil, err
	}
	for _, meaning := range dictionaryResponse[0].Meanings {
		definitionText := ""
		for _, definition := range meaning.Definitions {
			definitionText += fmt.Sprintf(">- %s (ex: %s)\n", definition.Definition, definition.Example)
		}
		meaningText += fmt.Sprintf(">### %s\n%s", meaning.SpeechPart, definitionText)
	}

	reply := fmt.Sprintf("%s:\n[%s](%s)\n%s", dictionaryResponse[0].Word, dictionaryResponse[0].Phonetics[0].Text, dictionaryResponse[0].Phonetics[0].Audio, meaningText)
	return []*domain.ClientMessage{
		domain.NewClientMessage(reply, command.Sender(), command.Private()),
	}, nil
}

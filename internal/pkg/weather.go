package pkg

import (
	"encoding/xml"
	"fmt"
	"github.com/raf924/bot/api/messages"
	"github.com/raf924/bot/pkg/bot"
	"github.com/raf924/bot/pkg/bot/command"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"net/url"
	"strings"
)

const defaultDegreeType = "C"
const maxDays = 3

func init() {
	bot.HandleCommand(&WeatherCommand{})
}

type weatherResponse struct {
	XMLName xml.Name `xml:"weatherdata"`
	Weather []struct {
		XMLName  xml.Name `xml:"weather"`
		Location string   `xml:"weatherlocationname,attr"`
		Current  struct {
			Temperature float64 `xml:"temperature,attr"`
			Sky         string  `xml:"skytext,attr"`
			Date        string  `xml:"date,attr"`
		} `xml:"current"`
		Forecast []struct {
			Low  float64 `xml:"low,attr"`
			High float64 `xml:"high,attr"`
			Sky  string  `xml:"skytextday,attr"`
			Day  string  `xml:"day,attr"`
			Date string  `xml:"date,attr"`
		} `xml:"forecast"`
	} `xml:"weather"`
}

type WeatherCommand struct {
	command.NoOpCommand
	weatherUrl *url.URL
}

func (w *WeatherCommand) Init(bot command.Executor) error {
	var err error
	w.weatherUrl, err = url.Parse("http://weather.service.msn.com/find.aspx?src=outlook")
	return err
}

func (w *WeatherCommand) Name() string {
	return "weather"
}

func (w *WeatherCommand) Aliases() []string {
	return []string{"w"}
}

func (w *WeatherCommand) Execute(command *messages.CommandPacket) (*messages.BotPacket, error) {
	if len(command.Args) == 0 {
		return nil, nil
	}
	wUrl, _ := url.Parse(w.weatherUrl.String())
	q := wUrl.Query()
	lastArg := strings.ToLower(command.Args[len(command.Args)-1])
	var degreeType string
	switch strings.ToLower(command.Args[len(command.Args)-1]) {
	case "f":
	case "c":
		degreeType = strings.ToUpper(lastArg)
		command.Args = command.Args[:len(command.Args)-1]
	default:
		degreeType = defaultDegreeType
	}
	q.Set("weadegreetype", degreeType)
	q.Set("weasearchstr", strings.Join(command.Args, " "))
	wUrl.RawQuery = q.Encode()
	resp, err := http.Get(wUrl.String())
	if err != nil {
		return nil, fmt.Errorf("get weather error: %s", err.Error())
	}
	var weatherData weatherResponse
	if err = xml.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("parse weather error: %s", err)
	}
	current := weatherData.Weather[0].Current
	text := fmt.Sprintf("Showing weather for %s\nCurrent: %0.1f°%s - %s\n", weatherData.Weather[0].Location, current.Temperature, degreeType, current.Sky)
	for _, forecast := range weatherData.Weather[0].Forecast {
		if forecast.Date <= current.Date {
			continue
		}
		text += fmt.Sprintf("%s: %0.1f°%s to %0.1f°%s - %s\n", forecast.Day, forecast.Low, degreeType, forecast.High, degreeType, forecast.Sky)
	}
	return &messages.BotPacket{
		Timestamp: timestamppb.Now(),
		Message:   text,
		Recipient: command.User,
		Private:   command.GetPrivate(),
	}, nil
}

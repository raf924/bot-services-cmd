package pkg

import (
	"encoding/xml"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"net/http"
	"net/url"
	"strings"
)

var _ command.Command = (*WeatherCommand)(nil)

const defaultDegreeType = "C"
const maxDays = 3

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

func (w *WeatherCommand) Init(command.Executor) error {
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

func (w *WeatherCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	if len(command.Args()) == 0 {
		return nil, nil
	}
	wUrl, _ := url.Parse(w.weatherUrl.String())
	q := wUrl.Query()
	args := command.Args()
	lastArg := strings.ToLower(args[len(args)-1])
	var degreeType string
	switch strings.ToLower(lastArg) {
	case "f":
	case "c":
		degreeType = strings.ToUpper(lastArg)
		args = args[:len(args)-1]
	default:
		degreeType = defaultDegreeType
	}
	q.Set("weadegreetype", degreeType)
	q.Set("weasearchstr", strings.Join(args, " "))
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
	return []*domain.ClientMessage{
		domain.NewClientMessage(text, command.Sender(), command.Private()),
	}, nil
}

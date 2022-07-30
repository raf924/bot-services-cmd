package pkg

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _ command.Command = (*WeatherCommand)(nil)

const defaultDegreeType = Metrics

type timeResponse struct {
	TimeZone         string `json:"timeZone"`
	CurrentLocalTime string `json:"currentLocalTime"`
	CurrentUtcOffset struct {
		Seconds      int   `json:"seconds"`
		Milliseconds int   `json:"milliseconds"`
		Ticks        int64 `json:"ticks"`
		Nanoseconds  int64 `json:"nanoseconds"`
	} `json:"currentUtcOffset"`
	StandardUtcOffset struct {
		Seconds      int   `json:"seconds"`
		Milliseconds int   `json:"milliseconds"`
		Ticks        int64 `json:"ticks"`
		Nanoseconds  int64 `json:"nanoseconds"`
	} `json:"standardUtcOffset"`
	HasDayLightSaving      bool `json:"hasDayLightSaving"`
	IsDayLightSavingActive bool `json:"isDayLightSavingActive"`
	DstInterval            struct {
		DstName        string `json:"dstName"`
		DstOffsetToUtc struct {
			Seconds      int   `json:"seconds"`
			Milliseconds int   `json:"milliseconds"`
			Ticks        int64 `json:"ticks"`
			Nanoseconds  int64 `json:"nanoseconds"`
		} `json:"dstOffsetToUtc"`
		DstOffsetToStandardTime struct {
			Seconds      int   `json:"seconds"`
			Milliseconds int   `json:"milliseconds"`
			Ticks        int64 `json:"ticks"`
			Nanoseconds  int64 `json:"nanoseconds"`
		} `json:"dstOffsetToStandardTime"`
		DstStart    time.Time `json:"dstStart"`
		DstEnd      time.Time `json:"dstEnd"`
		DstDuration struct {
			Days                 int   `json:"days"`
			NanosecondOfDay      int   `json:"nanosecondOfDay"`
			Hours                int   `json:"hours"`
			Minutes              int   `json:"minutes"`
			Seconds              int   `json:"seconds"`
			Milliseconds         int   `json:"milliseconds"`
			SubsecondTicks       int   `json:"subsecondTicks"`
			SubsecondNanoseconds int   `json:"subsecondNanoseconds"`
			BclCompatibleTicks   int64 `json:"bclCompatibleTicks"`
			TotalDays            int   `json:"totalDays"`
			TotalHours           int   `json:"totalHours"`
			TotalMinutes         int   `json:"totalMinutes"`
			TotalSeconds         int   `json:"totalSeconds"`
			TotalMilliseconds    int64 `json:"totalMilliseconds"`
			TotalTicks           int64 `json:"totalTicks"`
			TotalNanoseconds     int64 `json:"totalNanoseconds"`
		} `json:"dstDuration"`
	} `json:"dstInterval"`
}

type geocodingResponse struct {
	PlaceId     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmId       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
}

type weatherResponse struct {
	Cnt  int `json:"cnt"`
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			TempMin   float64 `json:"temp_min"`
			TempMax   float64 `json:"temp_max"`
			Pressure  int     `json:"pressure"`
			SeaLevel  int     `json:"sea_level"`
			GrndLevel int     `json:"grnd_level"`
			Humidity  int     `json:"humidity"`
			TempKf    float64 `json:"temp_kf"`
		} `json:"main"`
		Weather []struct {
			Id          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Clouds struct {
			All int `json:"all"`
		} `json:"clouds"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   int     `json:"deg"`
			Gust  float64 `json:"gust"`
		} `json:"wind"`
		Visibility int     `json:"visibility"`
		Pop        float64 `json:"pop"`
		Sys        struct {
			Pod string `json:"pod"`
		} `json:"sys"`
		DtTxt string `json:"dt_txt"`
	} `json:"list"`
	City struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Coord struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
		Country    string `json:"country"`
		Population int    `json:"population"`
		Timezone   int    `json:"timezone"`
		Sunrise    int    `json:"sunrise"`
		Sunset     int    `json:"sunset"`
	} `json:"city"`
}

type degree string

const (
	Metrics degree = "C"

	Imperial degree = "F"
)

func (d degree) convert(temperatureKelvin float64) float64 {
	switch d {
	case Metrics:
		return temperatureKelvin - 273.15
	case Imperial:
		return temperatureKelvin*9/5 - 459.67
	}
	return temperatureKelvin
}

type location struct {
	name      string
	country   string
	latitude  float64
	longitude float64
}

type weatherForDay struct {
	temperature float64
	sky         string
	min         float64
	max         float64
	day         string
}

type forecast struct {
	current       weatherForDay
	followingDays []*weatherForDay
	location      string
}

type WeatherCommand struct {
	command.NoOpInterceptor
	baseUrl *url.URL
	apiKey  string
}

func (w *WeatherCommand) geocodingUrl(search string) *url.URL {
	baseUrl, _ := url.Parse("https://geocode.maps.co/search")
	query := baseUrl.Query()
	query.Set("q", search)
	baseUrl.RawQuery = query.Encode()
	return baseUrl
}

func (w *WeatherCommand) weatherUrl(latitude float64, longitude float64) *url.URL {
	baseUrl, _ := w.baseUrl.Parse("/data/2.5/forecast")
	query := baseUrl.Query()
	query.Set("lat", strconv.FormatFloat(latitude, 'f', 5, 64))
	query.Set("lon", strconv.FormatFloat(longitude, 'f', 5, 64))
	query.Set("appid", w.apiKey)
	baseUrl.RawQuery = query.Encode()
	return baseUrl
}

func (w *WeatherCommand) fetchLocationTime(latitude float64, longitude float64) (time.Time, error) {
	timeUrl, _ := url.Parse("https://www.timeapi.io/api/TimeZone/coordinate")
	query := timeUrl.Query()
	query.Set("latitude", strconv.FormatFloat(latitude, 'f', 15, 64))
	query.Set("longitude", strconv.FormatFloat(longitude, 'f', 15, 64))
	timeUrl.RawQuery = query.Encode()
	response, err := http.Get(timeUrl.String())
	if err != nil {
		return time.Time{}, err
	}
	var timeR timeResponse
	err = json.NewDecoder(response.Body).Decode(&timeR)
	if err != nil {
		return time.Time{}, err
	}
	location := time.FixedZone(timeR.TimeZone, timeR.CurrentUtcOffset.Seconds)
	return time.ParseInLocation("2006-01-02T15:04:05", timeR.CurrentLocalTime, location)
}

func (w *WeatherCommand) Init(executor command.Executor) error {
	var err error
	w.apiKey = executor.ApiKeys()["openweather"]
	w.baseUrl, err = url.Parse("https://api.openweathermap.org")
	return err
}

func (w *WeatherCommand) Name() string {
	return "weather"
}

func (w *WeatherCommand) Aliases() []string {
	return []string{"w"}
}

func (w *WeatherCommand) fetchLocation(search string) (*location, error) {
	geocodingUrl := w.geocodingUrl(search)
	response, err := http.Get(geocodingUrl.String())
	if err != nil {
		return nil, err
	}
	var geocodingResponses []geocodingResponse
	err = json.NewDecoder(response.Body).Decode(&geocodingResponses)
	if err != nil {
		return nil, err
	}
	displayName := geocodingResponses[0].DisplayName
	parts := strings.Split(displayName, ",")
	lat, err := strconv.ParseFloat(geocodingResponses[0].Lat, 64)
	if err != nil {
		return nil, err
	}
	lon, err := strconv.ParseFloat(geocodingResponses[0].Lon, 64)
	if err != nil {
		return nil, err
	}
	return &location{
		name:      parts[0],
		country:   strings.TrimSpace(parts[len(parts)-1]),
		latitude:  lat,
		longitude: lon,
	}, nil
}

func atMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (w *WeatherCommand) fetchWeather(currentTime time.Time, latitude, longitude float64) (*forecast, error) {
	weatherUrl := w.weatherUrl(latitude, longitude)
	response, err := http.Get(weatherUrl.String())
	if err != nil {
		return nil, err
	}
	var weatherResponse weatherResponse
	err = json.NewDecoder(response.Body).Decode(&weatherResponse)
	if err != nil {
		return nil, err
	}
	f := forecast{}
	currentWeather := weatherResponse.List[0]
	f.current = weatherForDay{temperature: currentWeather.Main.Temp, sky: currentWeather.Weather[0].Description}
	nextDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()+1, 0, 0, 0, 0, currentTime.Location())
	for _, ww := range weatherResponse.List {
		nextDayUnix := nextDay.Unix()
		if f.followingDays == nil && ww.Dt < nextDayUnix {
			continue
		}
		forecastDay := atMidnight(time.Unix(ww.Dt, 0).In(currentTime.Location()))
		if forecastDay.Unix() == nextDayUnix {
			f.followingDays = append(f.followingDays, &weatherForDay{
				temperature: 0,
				sky:         "",
				min:         math.MaxFloat64,
				max:         0,
				day:         nextDay.Weekday().String(),
			})
			nextDay = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day()+1, 0, 0, 0, 0, nextDay.Location())
		}
		followingDay := f.followingDays[len(f.followingDays)-1]
		followingDay.min = math.Min(followingDay.min, ww.Main.TempMin)
		followingDay.max = math.Max(followingDay.max, ww.Main.TempMax)
		if !strings.HasSuffix(followingDay.sky, ww.Weather[0].Description) {
			if followingDay.sky != "" {
				followingDay.sky += " - "
			}
			followingDay.sky += ww.Weather[0].Description
		}
	}
	f.followingDays = f.followingDays[:len(f.followingDays)-1]
	return &f, nil
}

func (w *WeatherCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	if len(command.Args()) == 0 {
		return nil, nil
	}
	args := command.Args()
	lastArg := strings.ToLower(args[len(args)-1])
	var units degree
	switch strings.ToLower(lastArg) {
	case "f":
		units = Imperial
		args = args[:len(args)-1]
	case "c":
		units = Metrics
		args = args[:len(args)-1]
	default:
		units = defaultDegreeType
	}
	geoSearch := strings.Join(args, " ")
	loc, err := w.fetchLocation(geoSearch)
	if err != nil {
		return nil, fmt.Errorf("get weather error: %s", err.Error())
	}

	locationTime, err := w.fetchLocationTime(loc.latitude, loc.longitude)
	if err != nil {
		return nil, err
	}

	weather, err := w.fetchWeather(locationTime, loc.latitude, loc.longitude)
	if err != nil {
		return nil, err
	}

	text := fmt.Sprintf("Showing weather for %s, %s\nCurrent: %0.1f°%s - %s\n", loc.name, loc.country, units.convert(weather.current.temperature), units, weather.current.sky)
	for _, followingDay := range weather.followingDays {
		text += fmt.Sprintf("%s: %0.1f°%s to %0.1f°%s -- %s\n", followingDay.day, units.convert(followingDay.min), units, units.convert(followingDay.max), units, followingDay.sky)
	}
	return []*domain.ClientMessage{
		domain.NewClientMessage(text, command.Sender(), command.Private()),
	}, nil
}

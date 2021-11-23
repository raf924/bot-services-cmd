package bot_services_cmd

import (
	"github.com/raf924/bot-services-cmd/internal/pkg"
	"github.com/raf924/connector-sdk/command"
)

func init() {
	command.HandleCommand(&pkg.SearchCommand{})
	command.HandleCommand(&pkg.DictionaryCommand{})
	command.HandleCommand(&pkg.TimeCommand{})
	command.HandleCommand(&pkg.WeatherCommand{})
	command.HandleCommand(&pkg.YoutubeCommand{})
	command.HandleCommand(&pkg.UrbanCommand{})
	command.HandleCommand(&pkg.WikiCommand{})
	command.HandleCommand(&pkg.JokeCommand{})
}

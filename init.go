package bot_services_cmd

import (
	"github.com/raf924/bot-services-cmd/internal/pkg"
	_ "github.com/raf924/bot-services-cmd/internal/pkg"
	"github.com/raf924/bot/pkg/bot"
)

//Side effects

func init() {
	bot.HandleCommand(&pkg.SearchCommand{})
	bot.HandleCommand(&pkg.DictionaryCommand{})
	bot.HandleCommand(&pkg.TimeCommand{})
	bot.HandleCommand(&pkg.WeatherCommand{})
	bot.HandleCommand(&pkg.YoutubeCommand{})
	bot.HandleCommand(&pkg.UrbanCommand{})
	bot.HandleCommand(&pkg.WikiCommand{})
}

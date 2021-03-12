package bot_weather_cmd

import (
	"github.com/raf924/bot-services-cmd/internal/pkg"
	_ "github.com/raf924/bot-services-cmd/internal/pkg"
	"github.com/raf924/bot/pkg/bot"
)

//Side effects

func init() {
	bot.HandleCommand(&pkg.SearchCommand{})
}

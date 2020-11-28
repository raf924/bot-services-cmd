package main

import (
	"github.com/raf924/bot-weather-cmd/internal/pkg"
	"github.com/raf924/bot/api/messages"
	"os"
	"os/user"
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		panic("wat")
	}
	wcmd := pkg.WeatherCommand{}
	err = wcmd.Init(nil)
	if err != nil {
		panic(err)
	}
	packet, err := wcmd.Execute(&messages.CommandPacket{
		Timestamp: nil,
		Command:   "",
		Args:      os.Args[1:],
		User: &messages.User{
			Nick:  currentUser.Name,
			Id:    currentUser.Uid,
			Mod:   false,
			Admin: false,
		},
		Private: false,
	})
	if err != nil {
		panic(err)
	}
	print(packet.Message)
}

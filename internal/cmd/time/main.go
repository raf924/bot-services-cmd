package main

import (
	"github.com/raf924/bot-services-cmd/internal/pkg"
	messages "github.com/raf924/connector-api/pkg/gen"
	"os"
	"os/user"
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		panic("wat")
	}
	wcmd := pkg.TimeCommand{}
	err = wcmd.Init(nil)
	if err != nil {
		panic(err)
	}
	_, err = wcmd.Execute(&messages.CommandPacket{
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
}

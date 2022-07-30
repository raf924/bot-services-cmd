package pkg

import (
	"fmt"
	"github.com/raf924/connector-sdk/command"
	"github.com/raf924/connector-sdk/domain"
)

var _ command.Command = (*TimeCommand)(nil)

type TimeCommand struct {
	WeatherCommand
}

func (t *TimeCommand) Name() string {
	return "time"
}

func (t *TimeCommand) Aliases() []string {
	return []string{"t"}
}

func (t *TimeCommand) Execute(command *domain.CommandMessage) ([]*domain.ClientMessage, error) {
	if len(command.Args()) == 0 {
		return nil, fmt.Errorf("missing arguments")
	}
	location, err := t.fetchLocation(command.ArgString())
	if err != nil {
		return nil, err
	}
	locationTime, err := t.fetchLocationTime(location.latitude, location.longitude)
	if err != nil {
		return nil, err
	}
	return []*domain.ClientMessage{
		domain.NewClientMessage(
			fmt.Sprintf("%s - %s, %s", locationTime.Format("03:04:04 PM"), location.name, location.country),
			command.Sender(),
			command.Private(),
		),
	}, nil
}

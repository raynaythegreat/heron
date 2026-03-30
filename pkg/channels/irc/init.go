package irc

import (
	"github.com/raynaythegreat/heron/pkg/bus"
	"github.com/raynaythegreat/heron/pkg/channels"
	"github.com/raynaythegreat/heron/pkg/config"
)

func init() {
	channels.RegisterFactory("irc", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		if !cfg.Channels.IRC.Enabled {
			return nil, nil
		}
		return NewIRCChannel(cfg.Channels.IRC, b)
	})
}

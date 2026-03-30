package telegram

import (
	"github.com/raynaythegreat/heron/pkg/bus"
	"github.com/raynaythegreat/heron/pkg/channels"
	"github.com/raynaythegreat/heron/pkg/config"
)

func init() {
	channels.RegisterFactory("telegram", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewTelegramChannel(cfg, b)
	})
}

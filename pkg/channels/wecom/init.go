package wecom

import (
	"github.com/raynaythegreat/heron/pkg/bus"
	"github.com/raynaythegreat/heron/pkg/channels"
	"github.com/raynaythegreat/heron/pkg/config"
)

func init() {
	channels.RegisterFactory("wecom", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewChannel(cfg.Channels.WeCom, b)
	})
}

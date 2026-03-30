package maixcam

import (
	"github.com/raynaythegreat/heron/pkg/bus"
	"github.com/raynaythegreat/heron/pkg/channels"
	"github.com/raynaythegreat/heron/pkg/config"
)

func init() {
	channels.RegisterFactory("maixcam", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewMaixCamChannel(cfg.Channels.MaixCam, b)
	})
}

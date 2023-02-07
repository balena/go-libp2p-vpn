package libp2pvpn

import (
	"fmt"

	"github.com/songgao/water"
)

// Defines a VPN interface.
type Interface struct {
	*water.Interface
}

// Config defines parameters required to create a TUN/TAP interface. It's only
// used when the device is initialized. A zero-value Config is a valid
// configuration.
type Config struct {
	water.Config
	LinkOptions
}

// Defines a VPN device modifier option.
type Option func(cfg *Config)

func DeviceType(deviceType water.DeviceType) Option {
	return func(cfg *Config) {
		cfg.Config.DeviceType = deviceType
	}
}

// Creates a new VPN device.
func NewDevice(opts ...Option) (*Interface, error) {
	cfg := Config{}

	for _, opt := range opts {
		opt(&cfg)
	}

	iface, err := water.New(cfg.Config)
	if err != nil {
		return nil, err
	}

	err = setupLink(iface, cfg.LinkOptions)
	if err != nil {
		iface.Close()
		return nil, fmt.Errorf("error while setting %s: %s", iface.Name(), err)
	}

	return &Interface{iface}, err
}

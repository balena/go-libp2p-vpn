//go:build darwin
// +build darwin

package libp2pvpn

import (
	"fmt"
	"os/exec"

	"github.com/songgao/water"
)

// Name is the name for the interface to be used.
//
// For TunTapOSXDriver, it should be something like "tap0".
// For SystemDriver, the name should match `utun[0-9]+`, e.g. utun233
func Name(name string) Option {
	return func(cfg *Config) {
		cfg.Name = name
	}
}

// Driver should be set if an alternative driver is desired
// e.g. TunTapOSXDriver
func Driver(driver water.MacOSDriverProvider) Option {
	return func(cfg *Config) {
		cfg.Driver = driver
	}
}

// LinkOptions defines parameters in Config that are specific to
// macOS. A zero-value of such type is valid, yielding an interface
// with default values.
type LinkOptions struct {
	LocalAddress  string
	RemoteAddress string
	MTU           int
}

// When setting a TUN interface on macOS, set up point-to-point addresses.
func TunnelIP(local, remote string) Option {
	return func(cfg *Config) {
		cfg.LocalAddress = local
		cfg.RemoteAddress = remote
	}
}

// SetMTU sets the Maximum Tansmission Unit Size for a
// Packet on the interface.
func MTU(mtu int) Option {
	return func(cfg *Config) {
		cfg.MTU = mtu
	}
}

func setupLink(iface *water.Interface, linkOpts LinkOptions) error {
	if linkOpts.MTU != 0 {
		err := ifconfig(iface.Name(), "mtu", fmt.Sprintf("%d", linkOpts.MTU))
		if err != nil {
			return fmt.Errorf("error setting mtu: %s", err)
		}
	}

	err := ifconfig(iface.Name(), "inet", linkOpts.LocalAddress, linkOpts.RemoteAddress, "up")
	if err != nil {
		return fmt.Errorf("error bringing interface up: %s", err)
	}

	return nil
}

func ifconfig(args ...string) error {
	cmd := exec.Command("ifconfig", args...)
	return cmd.Run()
}

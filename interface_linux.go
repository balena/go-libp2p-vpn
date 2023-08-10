//go:build linux
// +build linux

package libp2pvpn

import (
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

const (
	defaultMTU = 1500
)

// Name is the name to be set for the interface to be created. This overrides
// the default name assigned by OS such as tap0 or tun0. A zero-value of this
// field, i.e. an empty string, indicates that the default name should be used.
func Name(name string) Option {
	return func(cfg *Config) {
		cfg.Name = name
	}
}

// Persist specifies whether persistence mode for the interface device should
// be enabled or disabled.
func Persist(persist bool) Option {
	return func(cfg *Config) {
		cfg.Persist = persist
	}
}

// Permissions, if non-nil, specifies the owner and group owner for the
// interface.  A zero-value of this field, i.e. nil, indicates that no
// changes to owner or group will be made.
func DevicePermissions(permissions *water.DevicePermissions) Option {
	return func(cfg *Config) {
		cfg.Permissions = permissions
	}
}

// MultiQueue specifies whether the multiqueue flag should be set on the
// interface.  From version 3.8, Linux supports multiqueue tuntap which can
// uses multiple file descriptors (queues) to parallelize packets sending or
// receiving.
func MultiQueue(multiQueue bool) Option {
	return func(cfg *Config) {
		cfg.MultiQueue = multiQueue
	}
}

// LinkOptions defines parameters in Config that are specific to
// Linux. A zero-value of such type is valid, yielding an interface
// with default values.
type LinkOptions struct {
	LocalAddress string
	RemoteAddress string
	MTU          int
}

// Sets the interface's local address and subnet.
func LocalAddress(address string) Option {
	return func(cfg *Config) {
		cfg.LocalAddress = address
	}
}

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
	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return err
	}

	if linkOpts.MTU != 0 {
		err = netlink.LinkSetMTU(link, linkOpts.MTU)
		if err != nil {
			return err
		}
	}

	if linkOpts.LocalAddress != "" {
		addr, err := netlink.ParseAddr(linkOpts.LocalAddress)
		if err != nil {
			return err
		}

		err = netlink.AddrAdd(link, addr)
		if err != nil {
			return err
		}
	}

	return netlink.LinkSetUp(link)
}

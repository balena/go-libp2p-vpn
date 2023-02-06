package libp2pvpn

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	pool "github.com/libp2p/go-buffer-pool"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	logging "github.com/ipfs/go-log/v2"
)

const (
	maxMTU              = 1500
	protocolTemplate    = "/vpn/%s/1.0.0"
	ethernetMinimumSize = 64
	serviceName         = "libp2p.vpn"

	vpnSize = maxMTU + ethernetMinimumSize + 2
)

var log = logging.Logger("vpn")

type VPNService struct {
	*Interface
	peer   peer.ID
	stream *network.Stream
}

func New(peer peer.ID, opts ...Option) (*VPNService, error) {
	iface, err := NewDevice(opts...)
	if err != nil {
		return nil, err
	}

	vpn := VPNService{
		Interface: iface,
		peer:      peer,
	}

	return &vpn, nil
}

func (vpn *VPNService) Protocol() protocol.ID {
	if vpn.IsTAP() {
		return protocol.ID(fmt.Sprintf(protocolTemplate, "tap"))
	} else {
		return protocol.ID(fmt.Sprintf(protocolTemplate, "tun"))
	}
}

func (vpn *VPNService) Handler() network.StreamHandler {
	return func(s network.Stream) {
		peer := s.Conn().RemotePeer()
		if peer != vpn.peer {
			log.Errorf("rejecting stream from: %s", peer.Pretty())
			s.Reset()
			return
		}

		if err := s.Scope().SetService(serviceName); err != nil {
			log.Debugf("error attaching stream to vpn service: %s", err)
			s.Reset()
			return
		}

		if err := s.Scope().ReserveMemory(vpnSize, network.ReservationPriorityAlways); err != nil {
			log.Debugf("error reserving memory for vpn stream: %s", err)
			s.Reset()
			return
		}
		defer s.Scope().ReleaseMemory(vpnSize)

		packetSize := pool.Get(2)
		defer pool.Put(packetSize)

		packet := pool.Get(maxMTU + ethernetMinimumSize)
		defer pool.Put(packet)

		defer s.Close()

		for {
			// Read the incoming packet's size as a binary value.
			n, err := io.ReadFull(s, packetSize)
			if err != nil {
				log.Errorf("p2p->if: error reading frame size: %w", err)
				return
			}
			if n != 2 {
				log.Errorf("p2p->if: unexpected frame size %d", n)
				return
			}

			// Decode the incoming packet's size from binary.
			size := int(binary.LittleEndian.Uint16(packetSize))

			n, err = io.ReadFull(s, packet[:size])
			if err != nil {
				log.Errorf("p2p->if: error reading frame: %w", err)
				return
			}
			if n == 0 || n != size {
				log.Errorf("p2p->if: expected frame size %d, got %d", size, n)
				return
			}

			if _, err := vpn.Write(packet[:size]); err != nil {
				log.Errorf("p2p->if: error writing frame: %w", err)
				return
			}
		}
	}
}

func (vpn *VPNService) Serve(ctx context.Context, host host.Host) {
	packet := make([]byte, maxMTU+ethernetMinimumSize)

	for {
		packetSize, err := vpn.Read(packet)
		if err != nil {
			log.Errorf("if->p2p: error reading frame: %s", err)
			continue
		}

		if vpn.stream == nil {
			s, err := host.NewStream(ctx, vpn.peer, vpn.Protocol())
			if err != nil {
				log.Errorf("if->p2p: error creating stream: %s", err)
				continue
			}

			if err := s.Scope().SetService(serviceName); err != nil {
				log.Debugf("error attaching device to stream: %s", err)
				s.Reset()
				return
			}

			vpn.stream = &s
		}

		err = binary.Write(*vpn.stream, binary.LittleEndian, uint16(packetSize))
		if err != nil {
			log.Errorf("if->p2p: error writing frame size: %s", err)
			(*vpn.stream).Close()
			vpn.stream = nil
			continue
		}

		_, err = (*vpn.stream).Write(packet[:packetSize])
		if err != nil {
			log.Errorf("if->p2p: error writing frame: %s", err)
			(*vpn.stream).Close()
			vpn.stream = nil
			continue
		}
	}
}

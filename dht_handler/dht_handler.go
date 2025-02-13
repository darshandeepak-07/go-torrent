package dht

import (
	"context"
	"encoding/binary"

	"net"
	"time"

	"github.com/anacrolix/dht/v2"
)

type DHTClient struct {
    server *dht.Server
}

func NewDHTClient() (*DHTClient, error) {
    udpConn, err := net.ListenPacket("udp4", ":0")
    if err != nil {
        return nil, fmt.Errorf("failed to listen: %w", err)
    }

    cfg := dht.NewDefaultServerConfig()
    cfg.ConnectionTracking = true
    cfg.Conn = udpConn
    cfg.StartingNodes = dht.GlobalBootstrapAddrs("udp4")

    server, err := dht.NewServer(cfg)
    if err != nil {
        udpConn.Close()
        return nil, fmt.Errorf("failed to create DHT server: %w", err)
    }

    return &DHTClient{
        server: server,
    }, nil
}

func (c *DHTClient) Close() error {
    if c.server != nil {
        return c.server.Close()
    }
    return nil
}

func (c *DHTClient) FetchPeers(infoHash [20]byte, timeout time.Duration) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // Bootstrap the DHT
    if _, err := c.server.Bootstrap(); err != nil {
        return nil, fmt.Errorf("bootstrap failed: %w", err)
    }

    // Get peers for the infohash
    peers := make(chan dht.PeersValues, 100)
    c.server.GetPeers(ctx, string(infoHash[:]), peers)

    var allPeers []net.TCPAddr
    timer := time.NewTimer(timeout)
    defer timer.Stop()

collectPeers:
    for {
        select {
        case <-ctx.Done():
            break collectPeers
        case <-timer.C:
            break collectPeers
        case peersValues, ok := <-peers:
            if !ok {
                break collectPeers
            }
            for _, p := range peersValues.Peers {
                if ip := p.IP.To4(); ip != nil {
                    allPeers = append(allPeers, net.TCPAddr{
                        IP:   ip,
                        Port: p.Port,
                    })
                }
            }
        }
    }

    if len(allPeers) == 0 {
        return nil, nil
    }

    // Convert peers to compact format
    peerBytes := make([]byte, 0, len(allPeers)*6)
    for _, peer := range allPeers {
        ip := peer.IP.To4()
        if ip == nil {
            continue
        }
        
        peerBytes = append(peerBytes, ip...)
        portBytes := make([]byte, 2)
        binary.BigEndian.PutUint16(portBytes, uint16(peer.Port))
        peerBytes = append(peerBytes, portBytes...)
    }

    return peerBytes, nil
}
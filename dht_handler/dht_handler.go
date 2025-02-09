package dht

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/anacrolix/dht/v2"
)

func initDHTNode() (*dht.Server, error) {
	udpConn, err := net.ListenPacket("udp", ":6881")
	if err != nil {
		return nil, err
	}

	bootstrapNodes, err := dht.GlobalBootstrapAddrs("udp")
	if err != nil {
		log.Println("Failed to fetch bootstrap nodes:", err)
		return nil, err
	}

	cfg := dht.ServerConfig{
		Conn:          udpConn,
		StartingNodes: dht.StartingNodesGetter(func() ([]dht.Addr, error) { return bootstrapNodes, nil }),
		NoSecurity:    true,
	}

	server, err := dht.NewServer(&cfg)
	if err != nil {
		log.Println("Failed to create DHT server:", err)
		return nil, err
	}

	log.Println("DHT Node started on port 6881")
	return server, nil
}

func FetchPeers(infoHash [20]byte) ([]byte, error) {
	server, err := initDHTNode()
	if err != nil {
		log.Println("Failed to initialize DHT node:", err)
		return nil, err
	}

	if server == nil {
		log.Println("DHT server is nil")
		return nil, errors.New("DHT server failed to initialize")
	}
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stats, err := server.Bootstrap()
		if err != nil {
			log.Println("DHT Bootstrapping failed:", err)
			return
		}
		log.Println("DHT Bootstrapped:", stats)
	}()
	wg.Wait()
	time.Sleep(time.Second * 5)
	server.AnnounceTraversal(infoHash)
	if server.PeerStore() == nil {
		log.Println("DHT PeerStore is nil")
		return nil, errors.New("DHT PeerStore is unavailable")
	}

	peers := server.PeerStore().GetPeers(infoHash)
	if len(peers) == 0 {
		log.Println("No peers found for the given info hash")
		return nil, nil
	}

	var peerBytes []byte
	for _, peer := range peers {
		ip := peer.IP.To4()
		if ip == nil {
			log.Println("Skipping non-IPv4 peer:", peer)
			continue
		}
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(peer.Port))

		peerBytes = append(peerBytes, ip...)
		peerBytes = append(peerBytes, portBytes...)
		log.Println("Found Peer:", peer)
	}

	return peerBytes, nil
}

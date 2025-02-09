package peer

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/darshandeepak-07/go-torrent/model"
)

func getPeers(peersBinary []byte) []byte {
	peersLen := len(peersBinary)
	lenNeeded := peersLen - (peersLen % 6)
	return peersBinary[0:lenNeeded]
}

func ParsePeers(peersBinary []byte) ([]model.Peer, error) {
	if len(peersBinary)%6 != 0 {
		peersBinary = getPeers(peersBinary)
	}

	var peers []model.Peer

	for i := 0; i < len(peersBinary); i += 6 {
		ip := net.IP(peersBinary[i : i+4])
		port := binary.BigEndian.Uint16(peersBinary[i+4 : i+6])
		peers = append(peers, model.Peer{IP: ip, Port: port})
	}

	return peers, nil
}

func GeneratePeerID() [20]byte {
	var peerID [20]byte
	copy(peerID[:8], []byte("-DD0001-"))
	_, err := rand.Read(peerID[8:])
	if err != nil {
		panic(err)
	}
	return peerID
}

func ConnectToPeer(peer model.Peer) (net.Conn, error) {
	address := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %v", address, err)
	}
	fmt.Println("Connected to peer:", address)
	return conn, nil
}

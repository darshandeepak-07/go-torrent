package main

import (
	"fmt"
	"log"

	"github.com/darshandeepak-07/go-torrent/parser"
	"github.com/darshandeepak-07/go-torrent/peer"
	"github.com/darshandeepak-07/go-torrent/tracker"
)

func main() {
	torrent, err := parser.ParseTorrent("inter.torrent")
	if err != nil {
		log.Println("Failed to read torrent")
	}
	torrent.Print()

	infoHash, err := parser.ComputeInfoHash(torrent.Info)

	if err != nil {
		log.Println("Error computing info hash")
	}

	log.Println("Info hash : ", infoHash)
	peerId := peer.GeneratePeerID()
	peers, error := tracker.ContactTracker(torrent.Announce, infoHash, peerId, 1000)
	if error != nil {
		log.Fatal("Error contacting tracker", error)
	}
	listOfPeer, parseError := peer.ParsePeers(peers)
	if parseError != nil {
		log.Fatal("Error retriving peers : ", parseError)
	}
	for _, peerNode := range listOfPeer {
		conn, err := peer.ConnectToPeer(peerNode)
		if err != nil {
			fmt.Println("Connection Failed : ", err)
			continue
		}
		defer conn.Close()
	}
}

package main

import (
	"log"

	"github.com/darshandeepak-07/go-torrent/parser"
)

func main() {
	torrent, err := parser.ParseTorrent("ubuntu.torrent")
	if err != nil {
		log.Println("Failed to read torrent")
	}
	torrent.Print()
}

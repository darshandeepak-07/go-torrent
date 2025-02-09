package parser

import (
	"os"

	"github.com/darshandeepak-07/go-torrent/model"
	"github.com/jackpal/bencode-go"
)

func ParseTorrent(filename string) (*model.TorrentFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	torrent := &model.TorrentFile{}
	err = bencode.Unmarshal(file, torrent)
	if err != nil {
		return nil, err
	}

	return torrent, nil
}

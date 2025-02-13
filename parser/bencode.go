package parser

import (
	"bytes"
	"crypto/sha1"
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

func ComputeInfoHash(info interface{}) ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, info)
	if err != nil {
		return [20]byte{}, err
	}
	hash := sha1.Sum(buf.Bytes())
	return hash, nil
}

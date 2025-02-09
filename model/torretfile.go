package model

import (
	"encoding/hex"
	"log"
)

type TorrentFile struct {
	Announce  string      `bencode:"announce" json:"announce"`
	Info      TorrentInfo `bencode:"info" json:"info"`
	Comment   string      `bencode:"comment" json:"comment"`
	CreatedBy string      `bencode:"created by" json:"created by"`
	Urls      []string    `bencode:"url-list" json:"url list"`
}

type TorrentInfo struct {
	Name        string `bencode:"name" json:"name"`
	Length      int    `bencode:"length" json:"length"`
	PieceLength int    `bencode:"piece length" json:"piece length"`
	Pieces      string `bencode:"pieces" json:"pieces"`
}

func (torrent TorrentFile) Print() {
	log.Println("Announce : ", torrent.Announce)
	log.Println("Comment : ", torrent.Comment)
	log.Println("Created By : ", torrent.CreatedBy)
	log.Println("Urls : ", torrent.Urls)
	log.Println("Torrent Info")
	log.Println("-------------")
	log.Println("Name : ", torrent.Info.Name)
	log.Println("Total Length : ", torrent.Info.Length)
	log.Println("Piece Length : ", torrent.Info.PieceLength)
	log.Println("Pieces Length after Parsing : ", len(torrent.parsePieces()))
}

func (torrent TorrentFile) parsePieces() []string {
	rawBytes := []byte(torrent.Info.Pieces)
	noOfPieces := len(rawBytes) / 20
	log.Println("No of pieces : ", noOfPieces)
	if len(rawBytes)%20 != 0 {
		log.Fatal("Invalid pieces length, not a multiple of 20 bytes")
	}

	var pieces []string

	for i := 0; i < noOfPieces; i++ {
		hash := rawBytes[i*20 : (i+1)*20]
		pieces = append(pieces, hex.EncodeToString(hash))
	}
	return pieces
}

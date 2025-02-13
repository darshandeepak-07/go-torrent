package tracker

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"

	dht "github.com/darshandeepak-07/go-torrent/dht_handler"
)

func ContactTracker(announceURL string, infoHash [20]byte, peerID [20]byte, port int) ([]byte, error) {
	_, err := net.LookupIP(announceURL)
	if err != nil {
		fmt.Println("DNS lookup failed:", err)
		fmt.Println("Proceeding with DHT.....")
		peers, err1 := dht.FetchPeers(infoHash)
		if err1 != nil {
			log.Println("Failed to fetch peers from DHT : ", err1)
		}
		return peers, err1
	}
	if strings.HasPrefix(announceURL, "http") {
		return contactHTTPTracker(announceURL, infoHash, peerID, port)
	} else if strings.HasPrefix(announceURL, "udp") {
		return contactUDPTracker(announceURL, infoHash, peerID, port)
	}
	return nil, fmt.Errorf("unsupported tracker protocol: %s", announceURL)
}

func contactHTTPTracker(announceURL string, infoHash [20]byte, peerID [20]byte, port int) ([]byte, error) {
	params := url.Values{}
	params.Add("info_hash", string(infoHash[:]))
	params.Add("peer_id", string(peerID[:]))
	params.Add("port", fmt.Sprintf("%d", port))
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", "100")

	fullURL := fmt.Sprintf("%s?%s", announceURL, params.Encode())

	response, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var reader io.Reader
	if response.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	} else {
		reader = response.Body
	}

	return io.ReadAll(reader)
}

func contactUDPTracker(announceURL string, infoHash [20]byte, peerID [20]byte, port int) ([]byte, error) {
	log.Println(announceURL)
	u, err := url.Parse(announceURL)

	if u.Port() == "" {
		u.Host += ":6969"
	}
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("udp", u.Host)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	transactionID := uint32(rand.Int31())
	protocolID := uint64(0x41727101980)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, protocolID)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, transactionID)

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	response := make([]byte, 16)
	_, err = conn.Read(response)
	if err != nil {
		return nil, err
	}

	if binary.BigEndian.Uint32(response[4:8]) != transactionID {
		return nil, fmt.Errorf("invalid transaction ID")
	}
	connectionID := binary.BigEndian.Uint64(response[8:16])

	announceReq := new(bytes.Buffer)
	binary.Write(announceReq, binary.BigEndian, connectionID)
	binary.Write(announceReq, binary.BigEndian, uint32(1))
	binary.Write(announceReq, binary.BigEndian, transactionID)
	binary.Write(announceReq, binary.BigEndian, infoHash)
	binary.Write(announceReq, binary.BigEndian, peerID)
	binary.Write(announceReq, binary.BigEndian, uint64(0))
	binary.Write(announceReq, binary.BigEndian, uint64(0))
	binary.Write(announceReq, binary.BigEndian, uint64(0))
	binary.Write(announceReq, binary.BigEndian, uint32(2))
	binary.Write(announceReq, binary.BigEndian, uint32(0))
	binary.Write(announceReq, binary.BigEndian, uint32(rand.Int31()))
	binary.Write(announceReq, binary.BigEndian, int32(-1))
	binary.Write(announceReq, binary.BigEndian, uint16(port))

	_, err = conn.Write(announceReq.Bytes())
	if err != nil {
		return nil, err
	}
	announceResp := make([]byte, 1024)
	n, err := conn.Read(announceResp)
	if err != nil {
		return nil, err
	}

	return announceResp[:n], nil
}

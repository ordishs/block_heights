package blockchain

import (
	"log"

	"github.com/ordishs/go-bitcoin"
)

type bcd struct {
	b *bitcoin.Bitcoind
}

func NewBitcoind() (*bcd, error) {
	b, err := bitcoin.New("localhost", 8332, "simon", "password", false)
	if err != nil {
		log.Fatalf("Could not connect to bitcoin: %v", err)
	}

	return &bcd{
		b: b,
	}, nil
}

func (s *bcd) Close() error {
	return nil
}

func (s *bcd) GetBlockHash(height uint64) (string, error) {
	return s.b.GetBlockHash(int(height)) // Genesis block
}

func (s *bcd) GetBlockHeader(hash string) (*Header, error) {
	header, err := s.b.GetBlockHeader(hash)
	if err != nil {
		return nil, err
	}

	dateNum, err := convertTimeToDateNum(header.MedianTime)
	if err != nil {
		log.Fatalf("Could not read time: %v", err)
	}

	return &Header{
		DateNum:       dateNum,
		Hash:          header.Hash,
		Height:        header.Height,
		Difficulty:    header.Difficulty,
		Size:          0, // We don't get this from the node
		NumTx:         uint64(header.TXCount),
		NextBlockHash: header.NextBlockHash,
	}, nil
}

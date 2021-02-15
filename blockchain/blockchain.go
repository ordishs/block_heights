package blockchain

import (
	"strconv"
	"time"
)

type BlockchainSource interface {
	Close() error
	GetBlockHeader(hash string) (*Header, error)
	GetBlockHash(height uint64) (string, error)
}

type Header struct {
	DateNum       uint32
	Hash          string
	Height        uint64
	Difficulty    float64
	Size          uint64
	NumTx         uint64
	NextBlockHash string
}

func convertTimeToDateNum(t uint64) (uint32, error) {
	tm := time.Unix(int64(t), 0).Format("20060102")
	u, err := strconv.ParseUint(tm, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(u), nil
}

package store

import "fmt"

type BlockStore interface {
	Close() error
	StoreBlock(*BlockData) error
	GetLastBlock(string) (*BlockData, error)
}

type BlockData struct {
	DateNum     uint32
	Coin        string
	StartHeight uint64
	StartHash   string
	EndHeight   uint64
	EndHash     string
	BlockCount  uint16
	Difficulty  float64
	FXRate      float64
}

func (b *BlockData) String() string {
	return fmt.Sprintf("%d [%s] %d -> %d (%d)", b.DateNum, b.Coin, b.StartHeight, b.EndHeight, b.BlockCount)
}

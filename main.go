package main

import (
	"block_heights/fx"
	"block_heights/store"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/ordishs/go-bitcoin"
)

// Name used by build script for the binaries. (Please keep on single line)
const progname = "block_heights"

// Version of the app to be incremented automatically build script (Please keep on single line)
const version = "0.0.0"

// Commit string injected at build with -ldflags -X...
var commit string

func main() {

	stores := initDatabases()
	bitcoins := initBitcoins()

	var wg sync.WaitGroup

	for coin, bitcoin := range bitcoins {
		wg.Add(1)
		go processBlocks(&wg, coin, bitcoin, stores)
	}

	wg.Wait()

	for _, s := range stores {
		err := s.Close()
		if err != nil {
			log.Fatalf("Could not close: %v", err)
		}
	}
}

func initDatabases() []store.BlockStore {
	stores := make([]store.BlockStore, 0)
	var blockStore store.BlockStore

	// bdb, err := store.NewBadger()
	// if err != nil {
	// 	log.Fatalf("Could not open badger db: %v", err)
	// }

	// blockStore = bdb
	// stores = append(stores, blockStore)

	sdb, err := store.NewSQLite()
	if err != nil {
		log.Fatalf("Could not open sqlite db: %v", err)
	}

	blockStore = sdb

	stores = append(stores, blockStore)

	return stores
}

func initBitcoins() map[string]*bitcoin.Bitcoind {
	bitcoins := make(map[string]*bitcoin.Bitcoind, 0)

	b, err := bitcoin.New("localhost", 8332, "simon", "password", false)
	if err != nil {
		log.Fatalf("Could not connect to bitcoin: %v", err)
	}

	bitcoins["BSV"] = b

	return bitcoins
}

func processBlocks(wg *sync.WaitGroup, coin string, b *bitcoin.Bitcoind, stores []store.BlockStore) {
	defer wg.Done()

	log.Printf("Starting goroutine for %s...", coin)

	// Get the last data we processed...
	block, err := stores[0].GetLastBlock(coin)
	if err != nil {
		log.Fatalf("Could not get the last block processed: %v", err)
	}

	var nextHash string

	lastDateNum := block.DateNum
	if lastDateNum == 0 {
		hash, err := b.GetBlockHash(0) // Genesis block
		if err != nil {
			log.Fatalf("Could not get block hash for genesis block: %v", err)
		}

		header, err := b.GetBlockHeader(hash)
		if err != nil {
			log.Fatalf("Could not get block header: %v", err)
		}
		dateNum, err := convertTimeToDateNum(header.MedianTime)
		if err != nil {
			log.Fatalf("Could not read time: %v", err)
		}
		block.DateNum = dateNum
		block.Coin = coin
		block.StartHash = hash
		block.EndHash = hash
		block.BlockCount++
		block.Difficulty += header.Difficulty

		nextHash = header.NextBlockHash
	} else {
		lastHeader, err := b.GetBlockHeader(block.EndHash)
		if err != nil {
			log.Fatalf("Could not get last block header: %v", err)
		}

		header, err := b.GetBlockHeader(lastHeader.NextBlockHash)
		if err != nil {
			log.Fatalf("Could not get block header: %v", err)
		}

		dateNum, err := convertTimeToDateNum(header.MedianTime)
		if err != nil {
			log.Fatalf("Could not read time: %v", err)
		}
		block.DateNum = dateNum
		block.Coin = coin
		block.StartHash = header.Hash
		block.StartHeight = header.Height
		block.EndHash = header.Hash
		block.EndHeight = header.Height
		block.BlockCount = 1
		block.Difficulty = header.Difficulty

		nextHash = header.NextBlockHash
	}

	for {
		header, err := b.GetBlockHeader(nextHash)
		if err != nil {
			log.Fatalf("Could not get block header: %v", err)
		}

		dateNum, err := convertTimeToDateNum(header.MedianTime)
		if err != nil {
			log.Fatalf("Could not read time: %v", err)
		}

		if block.DateNum != dateNum {
			block.FXRate = fx.GetRate(block.DateNum)

			for _, s := range stores {
				err := s.StoreBlock(block)
				if err != nil {
					log.Fatalf("Could not store: %v", err)
				}
			}

			log.Println(block)

			block.DateNum = dateNum
			block.StartHash = header.Hash
			block.StartHeight = header.Height
			block.BlockCount = 0
			block.Difficulty = 0
			block.FXRate = 0
		}

		block.EndHash = header.Hash
		block.EndHeight = header.Height
		block.BlockCount++
		block.Difficulty += header.Difficulty

		nextHash = header.NextBlockHash
	}

}

func convertTimeToDateNum(t uint64) (uint32, error) {
	tm := time.Unix(int64(t), 0).Format("20060102")
	u, err := strconv.ParseUint(tm, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(u), nil
}

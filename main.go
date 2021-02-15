package main

import (
	"block_heights/blockchain"
	"block_heights/fx"
	"block_heights/store"
	"log"
	"net/http"
	"os"
	"sync"

	_ "net/http/pprof"
)

// Name used by build script for the binaries. (Please keep on single line)
const progname = "block_heights"

// Version of the app to be incremented automatically build script (Please keep on single line)
const version = "0.0.0"

// Commit string injected at build with -ldflags -X...
var commit string

func main() {

	go func() {
		log.Fatalf("%v", http.ListenAndServe("localhost:6060", nil))
	}()

	stores := initDatabases()
	sources := initSources()

	var wg sync.WaitGroup

	for coin, bitcoin := range sources {
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

func initSources() map[string]blockchain.BlockchainSource {
	bitcoins := make(map[string]blockchain.BlockchainSource)

	if len(os.Args) < 2 {
		log.Fatal("You must specify BSTORE or BITCOIND")
	}

	switch os.Args[1] {
	case "BITCOIND":
		b, err := blockchain.NewBitcoind()
		if err != nil {
			log.Fatalf("Could not connect to bitcoin: %v", err)
		}
		bitcoins["BSV"] = b
		log.Println("Using BITCOIND")
	case "BSTORE":

		b, err := blockchain.NewBStore()
		if err != nil {
			log.Fatalf("Could not connect to bstore: %v", err)
		}
		bitcoins["BSV"] = b
		log.Println("Using BSTORE")
	default:
		log.Fatal("You must specify BSTORE or BITCOIND")
	}

	return bitcoins
}

func processBlocks(wg *sync.WaitGroup, coin string, b blockchain.BlockchainSource, stores []store.BlockStore) {
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

		block.DateNum = header.DateNum
		block.Coin = coin
		block.StartHash = hash
		block.EndHash = hash
		block.BlockCount++
		block.Size += header.Size
		block.Difficulty += header.Difficulty
		block.TXCount += uint32(header.NumTx)

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

		block.DateNum = header.DateNum
		block.Coin = coin
		block.StartHash = header.Hash
		block.StartHeight = header.Height
		block.EndHash = header.Hash
		block.EndHeight = header.Height
		block.BlockCount = 1
		block.Size = header.Size
		block.TXCount = uint32(header.NumTx)
		block.Difficulty = header.Difficulty

		nextHash = header.NextBlockHash
	}

	for {
		header, err := b.GetBlockHeader(nextHash)
		if err != nil {
			log.Fatalf("Could not get block header: %v", err)
		}

		if block.DateNum != header.DateNum {
			block.FXRate = fx.GetRate(block.DateNum)

			for _, s := range stores {
				err := s.StoreBlock(block)
				if err != nil {
					log.Fatalf("Could not store: %v", err)
				}
			}

			log.Println(block)

			block.DateNum = header.DateNum
			block.StartHash = header.Hash
			block.StartHeight = header.Height
			block.BlockCount = 0
			block.Difficulty = 0
			block.FXRate = 0
			block.TXCount = 0
			block.Size = 0
		}

		block.EndHash = header.Hash
		block.EndHeight = header.Height
		block.BlockCount++
		block.Difficulty += header.Difficulty
		block.TXCount += uint32(header.NumTx)
		block.Size += header.Size

		nextHash = header.NextBlockHash
	}

}

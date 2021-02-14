package store

import badger "github.com/dgraph-io/badger/v3"

type bdb struct {
	db *badger.DB
}

func NewBadger() (*bdb, error) {
	db, err := badger.Open(badger.DefaultOptions("badger"))
	if err != nil {
		return nil, err
	}

	return &bdb{
		db: db,
	}, nil
}

func (s *bdb) Close() error {
	return s.db.Close()
}

func (s *bdb) StoreBlock(blockData *BlockData) error {
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("answer"), []byte("42"))
		return err
	})
}

func (s *bdb) GetLastBlock(coin string) (*BlockData, error) {
	return nil, nil
}

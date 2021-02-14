package store

import (
	"database/sql"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type sqlite struct {
	db *sql.DB
}

func NewSQLite() (*sqlite, error) {
	db, err := sql.Open("sqlite3", "./blocks.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS blocks (
     datenum INTEGER NOT NULL
    ,coin VARCHAR(3) NOT NULL
    ,start_height INTEGER NOT NULL
    ,start_hash VARCHAR(64) NOT NULL
    ,end_height INTEGER NOT NULL
    ,end_hash VARCHAR(64) NOT NULL
    ,block_count INTEGER NOT NULL
    ,difficulty DOUBLE NOT NULL
    ,fxrate DOUBLE NOT NULL
    ,tx_count INTEGER NOT NULL
    ,PRIMARY KEY (datenum, coin)
    );
  `)

	if err != nil {
		return nil, err
	}

	return &sqlite{
		db: db,
	}, nil
}

func (s *sqlite) Close() error {
	return s.db.Close()
}

func (s *sqlite) StoreBlock(b *BlockData) error {
	stmt, err := s.db.Prepare(`
    INSERT INTO blocks (
     datenum
    ,coin
    ,start_height
    ,start_hash
    ,end_height
    ,end_hash
    ,block_count
    ,difficulty
    ,fxrate
    ,tx_count
    ) VALUES (
     ?
    ,?
    ,?
    ,?
    ,?
    ,?
    ,?
    ,?
    ,?
    ,?
    )
  `)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(b.DateNum, b.Coin, b.StartHeight, b.StartHash, b.EndHeight, b.EndHash, b.BlockCount, b.Difficulty, b.FXRate, b.TXCount)
	if err != nil {
		sqlite3Err, ok := err.(sqlite3.Error)
		if ok && sqlite3Err.Code == 19 {
			// Unique constraint - ignore
			return nil
		}
		return err
	}

	return nil
}

func (s *sqlite) GetLastBlock(coin string) (*BlockData, error) {
	row := s.db.QueryRow(`
    SELECT
     datenum
    ,coin
    ,start_height
    ,start_hash
    ,end_height
    ,end_hash
    ,block_count
    ,difficulty
    ,fxrate
    ,tx_count
    FROM blocks
    WHERE coin = $1
    ORDER BY datenum DESC
  `, coin)

	var d BlockData

	switch err := row.Scan(&d.DateNum, &d.Coin, &d.StartHeight, &d.StartHash, &d.EndHeight, &d.EndHash, &d.BlockCount, &d.Difficulty, &d.FXRate, &d.TXCount); err {
	case sql.ErrNoRows:
		d.Coin = coin
		return &d, nil
	case nil:
		return &d, err
	default:
		return nil, err
	}
}

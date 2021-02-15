package blockchain

import (
	"context"
	"log"
	"strconv"

	"bitbucket.org/Taal_Orchestrator/common"
	"bitbucket.org/Taal_Orchestrator/proto/bstore"
	"google.golang.org/grpc"
)

type bstoreType struct {
	conn   *grpc.ClientConn
	client bstore.BStoreClient
}

func NewBStore() (*bstoreType, error) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	conn, err := common.GetGRPCConnection(ctx, "bstore")
	if err != nil {
		log.Fatalf("Could not connect to bstore: %v", err)
	}

	client := bstore.NewBStoreClient(conn)

	return &bstoreType{
		conn:   conn,
		client: client,
	}, nil
}

func (s *bstoreType) Close() error {
	return s.conn.Close()
}

func (s *bstoreType) GetBlockHash(height uint64) (string, error) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	header, err := s.client.GetBlockHeader(ctx, &bstore.GetBlockHeaderRequest{
		Height: height,
	})
	if err != nil {
		return "", err
	}

	return header.Hash, nil
}

func (s *bstoreType) GetBlockHeader(hash string) (*Header, error) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	header, err := s.client.GetBlockHeader(ctx, &bstore.GetBlockHeaderRequest{
		Hash: hash,
	})
	if err != nil {
		return nil, err
	}

	diff, err := strconv.ParseFloat(header.Difficulty, 64)
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
		Difficulty:    diff,
		Size:          header.Size,
		NumTx:         header.NumTx,
		NextBlockHash: header.NextBlockHash,
	}, nil
}

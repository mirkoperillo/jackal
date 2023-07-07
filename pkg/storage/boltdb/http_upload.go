package boltdb

import (
	"context"
	"fmt"

	model "github.com/ortuman/jackal/pkg/model/httpupload"
	bolt "go.etcd.io/bbolt"
)

type boltDBHttpUploadRep struct {
	tx *bolt.Tx
}

func newHttpUploadRep(tx *bolt.Tx) *boltDBHttpUploadRep {
	return &boltDBHttpUploadRep{tx: tx}
}

// Unsopported
func (r *boltDBHttpUploadRep) InsertSlot(ctx context.Context, slot *model.UploadSlot) error {
	fmt.Println("BOLTDB UNSUPPORTED")
	return nil
}

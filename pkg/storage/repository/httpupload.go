package repository

import (
	"context"

	model "github.com/ortuman/jackal/pkg/model/httpupload"
)

type UploadSlot interface {
	InsertSlot(ctx context.Context, slot *model.UploadSlot) error
}

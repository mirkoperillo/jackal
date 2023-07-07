package measuredrepository

import (
	"context"

	model "github.com/ortuman/jackal/pkg/model/httpupload"
	"github.com/ortuman/jackal/pkg/storage/repository"
)

type measuredHttpUploadRep struct {
	rep  repository.UploadSlot
	inTx bool
}

func (m *measuredHttpUploadRep) InsertSlot(ctx context.Context, slot *model.UploadSlot) error {
	return m.rep.InsertSlot(ctx, slot)
}

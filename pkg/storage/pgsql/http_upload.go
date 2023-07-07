package pgsqlrepository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	kitlog "github.com/go-kit/log"
	model "github.com/ortuman/jackal/pkg/model/httpupload"
)

type pgSQLHttpUploadRep struct {
	conn   conn
	logger kitlog.Logger
}

func (r *pgSQLHttpUploadRep) InsertSlot(ctx context.Context, slot *model.UploadSlot) error {
	q := sq.Insert("slot_file_upload").
		Columns("filename", "size", "content_type").
		Values(slot.Filename, slot.Size, slot.ContentType).
		Suffix("RETURNING \"id\"")
	return q.RunWith(r.conn).QueryRowContext(ctx).Scan(&slot.Id)
}

package catelog

import "time"

type Album struct {
	ID        string    `json:"id" db:"id" noop:"create,update_db"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at" noop:"create,update_db"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" noop:"create,update_db"`
}

type ListAlbumsRes struct {
	Albums []Album `json:"albums"`
}

type GetAlbumRes struct {
	Album Album `json:"album"`
}

type GetAlbumReq struct {
	AlbumID string `json:"album_id"`
}

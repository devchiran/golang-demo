package catelog

import "time"

type Album struct {
	ID        string    `json:"id" db:"id" noop:"create,update_db"`
	Title     string    `json:"label" db:"label"`
	CreatedAt time.Time `json:"created_at" db:"created_at" noop:"create,update_db"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" noop:"create,update_db"`
}

type ListAlbumRes struct {
	Albums []Album `json:"albums"`
}

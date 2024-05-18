package catelog

import "time"

type Photo struct {
	ID           string    `json:"id" db:"id" noop:"create,update_db"`
	AlbumId      string    `json:"albumId" db:"albumId"`
	Url          string    `json:"url" db:"url"`
	ThumbnailUrl string    `json:"thumbnailUrl" db:"thumbnailUrl"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" noop:"create,update_db"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at" noop:"create,update_db"`
}

//     "albumId": 1,
//     "id": 1,
//     "title": "accusamus beatae ad facilis cum similique qui sunt",
//     "url": "https://via.placeholder.com/600/92c952",
//     "thumbnailUrl": "https://via.placeholder.com/150/92c952"

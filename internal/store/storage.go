package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetPostByID(context.Context, int64) (*Post, error)
		Delete(context.Context , int64) error
		Update(context.Context , *Post) error
	}
	Users interface {
		Create(context.Context, *User) error
	}
	Comments interface {
		GetCommentsByPostID(context.Context, int64) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UsersStore{db},
		Comments: &CommentStore{db},
	}
}

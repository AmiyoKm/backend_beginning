package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)
var (
	ErrorNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)
type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context , int64 , PaginatedFeedQuery) ([]PostWithMetadata , error)
	}
	Users interface {
		GetByID(context.Context, int64) (*User, error)
		Create(context.Context, *User) error

	}
	Followers interface {
		Follow(ctx context.Context , followerID int64 ,userID int64) error
		Unfollow(ctx context.Context , followerID int64 ,userID int64) error
	}
	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostID(context.Context, int64) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UsersStore{db},
		Comments: &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}

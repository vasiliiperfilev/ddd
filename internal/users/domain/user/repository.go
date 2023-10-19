package user

import (
	"context"
	"errors"
)

var (
	NotFoundError = errors.New("User not found")
)

type Repository interface {
	Get(ctx context.Context, id string) (User, error)
	Update(
		ctx context.Context,
		id string,
		updateFn func(u User) (User, error),
	) error
}

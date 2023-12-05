package grpc

import "context"

type UserService interface {
	// @path /users/:id
	// @method GET
	// @header
	// @authorization
	GetById(ctx context.Context, id int64) (User, error)
}

type UserServiceV1 struct {
	GetById func(ctx context.Context, id int64) (User, error) `method:"get" path:"/user/:id"`
}

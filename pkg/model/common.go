package model

import (
	"context"
	"time"
)

type Common struct {
	Id        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Object interface {
	GetId() string
}

func (s *Common) GetId() string {
	return s.Id
}

type ITransaction interface {
	Transaction(ctx context.Context, fn func(txctx context.Context) error) error
}

type NoopTransaction struct{}

func (*NoopTransaction) Transaction(ctx context.Context, fn func(txctx context.Context) error) error {
	return fn(ctx)
}

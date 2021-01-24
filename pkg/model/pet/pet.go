package pet

import (
	"context"

	"github.com/win5do/golang-microservice-demo/pkg/model"
)

type PetDomainInterface interface {
	PetDb(ctx context.Context) PetDbInterface
	OwnerDb(ctx context.Context) OwnerDbInterface
	Owner_PetDb(ctx context.Context) Owner_PetDbInterface
}

type Pet struct {
	model.Common
	Name  string
	Type  string
	Age   uint32
	Sex   string
	Owned bool
}

type PetDbInterface interface {
	Get(id string) (*Pet, error)
	List(query *Pet, offset, limit int) ([]*Pet, error)
	Create(query *Pet) (*Pet, error)
	Update(query *Pet) (*Pet, error)
	Delete(query *Pet) error
}

type Owner struct {
	model.Common
	Name  string
	Age   uint32
	Sex   string
	Phone string
}

type OwnerDbInterface interface {
	Get(id string) (*Owner, error)
	List(query *Owner, offset, limit int) ([]*Owner, error)
	Create(query *Owner) (*Owner, error)
	Update(query *Owner) (*Owner, error)
	Delete(query *Owner) error
}

type Owner_Pet struct {
	model.Common
	OwnerId string
	PetId   string
}

type Owner_PetDbInterface interface {
	Query(query *Owner_Pet) ([]*Owner_Pet, error)
	Create(query *Owner_Pet) (*Owner_Pet, error)
	Delete(query *Owner_Pet) error
}

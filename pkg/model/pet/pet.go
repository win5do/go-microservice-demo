package pet

import (
	"context"

	"github.com/win5do/golang-microservice-demo/pkg/model"
)

type IPetDomain interface {
	PetDb(ctx context.Context) IPetDb
	OwnerDb(ctx context.Context) IOwnerDb
	OwnerPetDb(ctx context.Context) IOwnerPetDb
}

type Pet struct {
	model.Common
	Name  string
	Type  string
	Age   uint32
	Sex   string
	Owned bool
}

type IPetDb interface {
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

type IOwnerDb interface {
	Get(id string) (*Owner, error)
	List(query *Owner, offset, limit int) ([]*Owner, error)
	Create(query *Owner) (*Owner, error)
	Update(query *Owner) (*Owner, error)
	Delete(query *Owner) error
}

type OwnerPet struct {
	model.Common
	OwnerId string
	PetId   string
}

type IOwnerPetDb interface {
	Query(query *OwnerPet) ([]*OwnerPet, error)
	Create(query *OwnerPet) (*OwnerPet, error)
	Delete(query *OwnerPet) error
}

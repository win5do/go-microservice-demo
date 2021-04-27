package pet

import (
	"context"

	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
)

type petDomain struct{}

func NewPetDomain() *petDomain {
	return &petDomain{}
}

func (*petDomain) PetDb(ctx context.Context) petmodel.IPetDb {
	return &petDb{dbcore.GetDB(ctx)}
}

func (*petDomain) OwnerDb(ctx context.Context) petmodel.IOwnerDb {
	return &ownerDb{dbcore.GetDB(ctx)}
}

func (*petDomain) OwnerPetDb(ctx context.Context) petmodel.IOwnerPetDb {
	return &ownerPetDb{dbcore.GetDB(ctx)}
}

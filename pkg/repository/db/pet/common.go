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

func (*petDomain) PetDb(ctx context.Context) petmodel.PetDbInterface {
	return &petDb{dbcore.GetDB(ctx)}
}

func (*petDomain) OwnerDb(ctx context.Context) petmodel.OwnerDbInterface {
	return &ownerDb{dbcore.GetDB(ctx)}
}

func (*petDomain) Owner_PetDb(ctx context.Context) petmodel.Owner_PetDbInterface {
	return &owner_petDb{dbcore.GetDB(ctx)}
}

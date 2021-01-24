package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/win5do/golang-microservice-demo/pkg/model"
	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
	petdb "github.com/win5do/golang-microservice-demo/pkg/repository/db/pet"
)

var PetDomain petmodel.PetDomainInterface = petdb.NewPetDomain()
var TxImpl model.TransactionInterface = dbcore.NewTxImpl()

func TestCreatePet(t *testing.T) {
	_, err := PetDomain.PetDb(context.Background()).Create(&petmodel.Pet{
		Name: "gugu",
		Type: "cat",
		Sex:  "male",
	})
	require.NoError(t, err)
}

func TestGetPet(t *testing.T) {
	_, err := PetDomain.PetDb(context.Background()).Get("01EW4T6T6YSRTG96J0D4V9BVPX")
	require.NoError(t, err)
}

func TestUpdatePet(t *testing.T) {
	_, err := PetDomain.PetDb(context.Background()).Update(&petmodel.Pet{
		Common: model.Common{
			Id: "01EW4TGRK2RA0MF1J5TSX3M88Z",
		},
		Name: "gugu",
		Type: "cat",
		Sex:  "female",
	})
	require.NoError(t, err)
}

func TestTransaction(t *testing.T) {
	err := TxImpl.Transaction(context.Background(), func(txctx context.Context) error {
		owner, err := PetDomain.OwnerDb(txctx).Create(&petmodel.Owner{
			Name: "qq",
			Age:  18,
			Sex:  "female",
		})
		if err != nil {
			return err
		}

		// return errors.New("rollback") // test

		_, err = PetDomain.Owner_PetDb(txctx).Create(&petmodel.Owner_Pet{
			OwnerId: owner.Id,
			PetId:   "abc",
		})
		return err
	})
	require.NoError(t, err)
}

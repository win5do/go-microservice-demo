package pet

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/win5do/golang-microservice-demo/pkg/api/petpb"
	"github.com/win5do/golang-microservice-demo/pkg/model"
	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
	"github.com/win5do/golang-microservice-demo/pkg/model/pet/mock_pet"
)

func mockPetSvc(petDomain petmodel.IPetDomain) *PetService {
	return NewPetService(&model.NoopTransaction{}, petDomain)
}

func TestGetPet(t *testing.T) {
	ctrl := gomock.NewController(t)
	petDomain := mock_pet.NewMockPetDomainInterface(ctrl)
	petDb := mock_pet.NewMockPetDbInterface(ctrl)
	petDomain.EXPECT().PetDb(gomock.Any()).Return(petDb)

	id := "abc"
	out := &petmodel.Pet{
		Common: model.Common{
			Id:        id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:  "gugu",
		Type:  "cat",
		Age:   1,
		Sex:   "male",
		Owned: true,
	}

	petDb.EXPECT().Get(id).Return(out, nil)

	r, err := mockPetSvc(petDomain).GetPet(context.Background(), &petpb.Id{
		Id: id,
	})
	require.NoError(t, err)
	require.EqualValues(t, ModelPet2PbPet(out), r)
}

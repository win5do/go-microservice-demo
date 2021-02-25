package pet

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/win5do/golang-microservice-demo/pkg/api/errcode"
	"github.com/win5do/golang-microservice-demo/pkg/api/petpb"
	"github.com/win5do/golang-microservice-demo/pkg/log"
	"github.com/win5do/golang-microservice-demo/pkg/model"
	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
)

type PetService struct {
	petpb.UnimplementedPetServiceServer

	petDomain petmodel.PetDomainInterface
	txImpl    model.TransactionInterface
}

func NewPetService(txImpl model.TransactionInterface, petDomain petmodel.PetDomainInterface) *PetService {
	return &PetService{
		txImpl:    txImpl,
		petDomain: petDomain,
	}
}

func (s *PetService) Ping(ctx context.Context, in *petpb.Id) (*petpb.Id, error) {
	log.Debugf("req: %s", in.Id)
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &petpb.Id{
		Id: fmt.Sprintf("%s, %s", in.Id, host),
	}, nil
}

func (s *PetService) ListPet(ctx context.Context, in *emptypb.Empty) (*petpb.PetList, error) {
	pets, err := s.petDomain.PetDb(ctx).List(&petmodel.Pet{}, 0, 0)
	if err != nil {
		return nil, pberr(err)
	}

	out := &petpb.PetList{
		Items: ModelPet2PbPetList(pets),
	}
	return out, nil
}

func (s *PetService) GetPet(ctx context.Context, id *petpb.Id) (*petpb.Pet, error) {
	pet, err := s.petDomain.PetDb(ctx).Get(id.Id)
	if err != nil {
		return nil, pberr(err)
	}

	return ModelPet2PbPet(pet), nil
}

func (s *PetService) CreatePet(ctx context.Context, in *petpb.Pet) (*petpb.Pet, error) {
	pet, err := s.petDomain.PetDb(ctx).Create(PbPet2ModelPet(in))
	if err != nil {
		return nil, pberr(err)
	}

	return ModelPet2PbPet(pet), nil
}

func (s *PetService) UpdatePet(ctx context.Context, in *petpb.Pet) (*petpb.Pet, error) {
	pet, err := s.petDomain.PetDb(ctx).Update(PbPet2ModelPet(in))
	if err != nil {
		return nil, pberr(err)
	}

	return ModelPet2PbPet(pet), nil
}

func (s *PetService) DeletePet(ctx context.Context, in *petpb.Id) (*emptypb.Empty, error) {
	err := s.petDomain.PetDb(ctx).Delete(&petmodel.Pet{
		Common: model.Common{
			Id: in.Id,
		},
	})
	if err != nil {
		return nil, pberr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *PetService) ListOwner(ctx context.Context, in *emptypb.Empty) (*petpb.OwnerList, error) {
	owners, err := s.petDomain.OwnerDb(ctx).List(&petmodel.Owner{}, 0, 0)
	if err != nil {
		return nil, pberr(err)
	}

	out := &petpb.OwnerList{
		Items: ModelOwner2PbOwnerList(owners),
	}
	return out, nil
}

func (s *PetService) GetOwner(ctx context.Context, in *petpb.Id) (*petpb.Owner, error) {
	owner, err := s.petDomain.OwnerDb(ctx).Get(in.Id)
	if err != nil {
		return nil, pberr(err)
	}

	return ModelOwner2PbOwner(owner), nil
}

func (s *PetService) CreateOwner(ctx context.Context, in *petpb.Owner) (*petpb.Owner, error) {
	owner, err := s.petDomain.OwnerDb(ctx).Create(PbOwner2ModelOwner(in))
	if err != nil {
		return nil, pberr(err)
	}

	return ModelOwner2PbOwner(owner), nil
}

func (s *PetService) UpdateOwner(ctx context.Context, in *petpb.Owner) (*petpb.Owner, error) {
	owner, err := s.petDomain.OwnerDb(ctx).Update(PbOwner2ModelOwner(in))
	if err != nil {
		return nil, pberr(err)
	}

	return ModelOwner2PbOwner(owner), nil
}

func (s *PetService) DeleteOwner(ctx context.Context, in *petpb.Id) (*emptypb.Empty, error) {
	rows, err := s.petDomain.Owner_PetDb(ctx).Query(&petmodel.Owner_Pet{
		OwnerId: in.Id,
	})
	if err != nil {
		return nil, pberr(err)
	}

	if len(rows) > 0 {
		return nil, pberr(errcode.Err_conflict)
	}

	err = s.petDomain.OwnerDb(ctx).Delete(&petmodel.Owner{
		Common: model.Common{
			Id: in.Id,
		},
	})
	if err != nil {
		return nil, pberr(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *PetService) OwnPet(ctx context.Context, in *petpb.Owner_Pet) (*petpb.Owner_Pet, error) {
	var r *petmodel.Owner_Pet

	err := s.txImpl.Transaction(ctx, func(txctx context.Context) error {
		ownerJoinPet, err := s.petDomain.Owner_PetDb(txctx).Create(&petmodel.Owner_Pet{
			PetId:   in.PetId,
			OwnerId: in.OwnerId,
		})
		if err != nil {
			return pberr(err)
		}

		r = ownerJoinPet

		_, err = s.petDomain.PetDb(txctx).Update(&petmodel.Pet{
			Common: model.Common{
				Id: in.Id,
			},
			Owned: true,
		})
		if err != nil {
			return pberr(err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &petpb.Owner_Pet{
		Id:        r.Id,
		CreatedAt: time2Pb(r.CreatedAt),
		UpdatedAt: time2Pb(r.UpdatedAt),
		PetId:     r.PetId,
		OwnerId:   r.OwnerId,
	}, nil
}

func (s *PetService) AbandonPet(ctx context.Context, in *petpb.Owner_Pet) (*emptypb.Empty, error) {
	err := s.txImpl.Transaction(ctx, func(txctx context.Context) error {
		err := s.petDomain.Owner_PetDb(txctx).Delete(&petmodel.Owner_Pet{
			PetId:   in.PetId,
			OwnerId: in.OwnerId,
		})
		if err != nil {
			return pberr(err)
		}

		_, err = s.petDomain.PetDb(txctx).Update(&petmodel.Pet{
			Common: model.Common{
				Id: in.Id,
			},
			Owned: false,
		})
		if err != nil {
			return pberr(err)
		}

		return nil
	})

	return &emptypb.Empty{}, err
}

package pet

import (
	"gorm.io/gorm"

	"github.com/win5do/go-lib/errx"

	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
)

func init() {
	dbcore.RegisterInjector(func(db *gorm.DB) {
		dbcore.SetupTableModel(db, &petmodel.OwnerPet{})
	})
}

type ownerPetDb struct {
	db *gorm.DB
}

func (s *ownerPetDb) Query(in *petmodel.OwnerPet) ([]*petmodel.OwnerPet, error) {
	var r []*petmodel.OwnerPet
	err := s.db.Where(in).Find(&r).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return r, nil
}

func (s *ownerPetDb) Create(in *petmodel.OwnerPet) (*petmodel.OwnerPet, error) {
	err := s.db.Create(in).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return in, nil
}

func (s *ownerPetDb) Delete(in *petmodel.OwnerPet) error {
	err := s.db.Where(in).Delete(&petmodel.OwnerPet{}).Error
	if err != nil {
		return errx.WithStackOnce(err)
	}

	return nil
}

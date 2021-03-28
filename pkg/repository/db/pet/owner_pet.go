package pet

import (
	"gorm.io/gorm"

	"github.com/win5do/go-lib/errx"

	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
)

func init() {
	dbcore.RegisterInjector(func(db *gorm.DB) {
		dbcore.SetupTableModel(db, &petmodel.Owner_Pet{})
	})
}

type owner_petDb struct {
	db *gorm.DB
}

func (s *owner_petDb) Query(in *petmodel.Owner_Pet) ([]*petmodel.Owner_Pet, error) {
	var r []*petmodel.Owner_Pet
	err := s.db.Where(in).Find(&r).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return r, nil
}

func (s *owner_petDb) Create(in *petmodel.Owner_Pet) (*petmodel.Owner_Pet, error) {
	err := s.db.Create(in).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return in, nil
}

func (s *owner_petDb) Delete(in *petmodel.Owner_Pet) error {
	err := s.db.Where(in).Delete(&petmodel.Owner_Pet{}).Error
	if err != nil {
		return errx.WithStackOnce(err)
	}

	return nil
}

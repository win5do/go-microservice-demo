package pet

import (
	"gorm.io/gorm"

	"github.com/win5do/golang-microservice-demo/pkg/lib/errx"
	petmodel "github.com/win5do/golang-microservice-demo/pkg/model/pet"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
)

func init() {
	dbcore.RegisterInjector(func(db *gorm.DB) {
		dbcore.SetupTableModel(db, &petmodel.Owner{})
	})
}

type ownerDb struct {
	db *gorm.DB
}

func (s *ownerDb) List(query *petmodel.Owner, offset, limit int) ([]*petmodel.Owner, error) {
	var r []*petmodel.Owner

	db := dbcore.WithOffsetLimit(s.db, offset, limit)

	err := db.Where(query).Find(&r).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return r, nil
}

func (s *ownerDb) Get(id string) (*petmodel.Owner, error) {
	var r petmodel.Owner
	err := s.db.Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return &r, nil
}

func (s *ownerDb) Create(in *petmodel.Owner) (*petmodel.Owner, error) {
	err := s.db.Create(in).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return in, nil
}

func (s *ownerDb) Update(in *petmodel.Owner) (*petmodel.Owner, error) {
	err := s.db.Updates(in).Error
	if err != nil {
		return nil, errx.WithStackOnce(err)
	}

	return in, nil
}

func (s *ownerDb) Delete(in *petmodel.Owner) error {
	err := s.db.Where(in).Delete(&petmodel.Owner{}).Error
	if err != nil {
		return errx.WithStackOnce(err)
	}

	return nil
}

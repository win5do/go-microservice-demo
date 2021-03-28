package dbcore

import (
	"context"
	"os"
	"strings"
	"time"

	log "github.com/win5do/go-lib/logx"
	"gorm.io/gorm"

	"github.com/win5do/go-lib/errx"
)

const DefaultLeaseAge = 60 * time.Second

type lock struct {
	CommonModel
	ExpiredAt time.Time
	Action    string `gorm:"unique;not null"`
	Holder    string
}

func init() {
	RegisterInjector(func(d *gorm.DB) {
		SetupTableModel(d, &lock{})
	})
}

type lockDb struct {
	db       *gorm.DB
	stopCh   chan struct{}
	action   string
	holder   string
	leaseAge time.Duration
}

func NewLockDb(action, holder string, lease time.Duration) *lockDb {
	return &lockDb{
		db:       GetDB(context.Background()),
		stopCh:   make(chan struct{}),
		action:   action,
		holder:   holder,
		leaseAge: lease,
	}
}

func (s *lockDb) Lock() (bool, error) {
	err := s.cleanExpired()
	if err != nil {
		return false, errx.WithStackOnce(err)
	}

	err = s.db.Create(&lock{
		ExpiredAt: time.Now().Add(s.leaseAge),
		Action:    s.action,
		Holder:    s.holder,
	}).Error
	if err != nil {
		// Duplicate entry '<action_val>' for key 'action'
		if strings.Contains(err.Error(), "Duplicate entry") {
			return false, nil
		}
		return false, errx.WithStackOnce(err)
	}

	s.startLease()

	log.Debugf("%s get lock", s.holder)

	return true, nil
}

func (s *lockDb) UnLock() error {
	s.stopLease()
	var err error

	defer func() {
		err = s.db.
			Where("action = ? and holder = ?", s.action, s.holder).
			Delete(&lock{}).
			Error
	}()

	return err
}

func (s *lockDb) cleanExpired() error {
	err := s.db.
		Where("expired_at < ?", time.Now()).
		Delete(&lock{}).
		Error

	return err
}

func (s *lockDb) startLease() {
	go func() {
		// 剩余 1/4 时刷新租约
		ticker := time.NewTicker(s.leaseAge * 3 / 4)
		for {
			select {
			case <-ticker.C:
				err := s.refreshLease()
				if err != nil {
					log.Errorf("refreash lease err: %s", err)
				} else {
					log.Debug("lease refreshed")
				}
			case <-s.stopCh:
				log.Debug("lease stopped")
				return
			}
		}
	}()
}

func (s *lockDb) stopLease() {
	close(s.stopCh)
}

func (s *lockDb) refreshLease() error {
	err := s.db.Model(&lock{}).
		Where("action = ? and holder = ?", s.action, s.holder).
		Update("expired_at", time.Now().Add(s.leaseAge)).
		Error

	return err
}

func GetHostname() string {
	host, err := os.Hostname()
	if err != nil {
		log.Fatal("err: %+v", err)
	}

	return host
}

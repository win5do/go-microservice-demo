package dbinit

import (
	log "github.com/sirupsen/logrus"

	"github.com/win5do/golang-microservice-demo/pkg/lib/errx"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
)

func InitData() error {
	locker := dbcore.NewLockDb("init", dbcore.GetHostname(), dbcore.DefaultLeaseAge)
	ok, err := locker.Lock()
	if err != nil {
		return errx.WithStackOnce(err)
	}

	if !ok {
		return nil
	}

	defer locker.UnLock()

	return run()
}

func run() error {
	log.Infof("%s begin init data", dbcore.GetHostname())
	return nil
}

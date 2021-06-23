package db_test

import (
	"os"
	"testing"

	log "github.com/win5do/go-lib/logx"
	"go.uber.org/zap/zapcore"

	"github.com/win5do/golang-microservice-demo/pkg/config/util"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
	integration_test "github.com/win5do/golang-microservice-demo/pkg/test/integration"
)

func TestMain(m *testing.M) {
	if integration_test.SkipInCi() {
		return
	}
	dbcore.Connect(&dbcore.DBConfig{
		DSN: util.GetEnvOrDefault("DB_DSN", "root:123456@(127.0.0.1:3306)/go-demo"),
	})
	log.SetLogger(log.NewLogger(zapcore.DebugLevel))
	os.Exit(m.Run())
}

func TestUlid(t *testing.T) {
	testing.Short()
	t.Log(dbcore.NewUlid())
}

func TestCreateDatabase(t *testing.T) {
	dbcore.CreateDatabase(&dbcore.DBConfig{
		DSN: "root:123456@(127.0.0.1:3306)/not-exists",
	})
}

package dbcore

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm/schema"

	"github.com/oklog/ulid/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	log "github.com/win5do/go-lib/logx"

	"github.com/win5do/go-lib/errx"
)

var (
	globalDB *gorm.DB

	globalConfig *DBConfig

	injectors []func(db *gorm.DB)
)

func Connect(cfg *DBConfig) {
	cfg = defaultDbConfig(cfg)
	globalConfig = cfg

	// 连接数据库前初始化Database
	CreateDatabase(cfg)

	dsn := fmt.Sprintf(
		"%s?charset=utf8&parseTime=True&loc=Local",
		cfg.DSN,
	)
	log.Debugf("db dsn: %s", dsn)

	var ormLogger logger.Interface
	if cfg.Debug {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "tb_", // 表名前缀，`User` 对应的表名是 `tb_users`
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	idb, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	idb.SetMaxIdleConns(cfg.MaxIdleConns)
	idb.SetMaxOpenConns(cfg.MaxOpenConns)

	registerCallback(db)
	callInjector(db)
	globalDB = db

	log.Info("db connected success")
}

func CreateDatabase(cfg *DBConfig) {
	slashIndex := strings.LastIndex(cfg.DSN, "/")
	dsn := cfg.DSN[:slashIndex+1]
	dbName := cfg.DSN[slashIndex+1:]

	dsn = fmt.Sprintf("%s?charset=utf8&parseTime=True&loc=Local", dsn)
	db, err := gorm.Open(mysql.Open(dsn), nil)
	if err != nil {
		log.Fatal(err)
	}

	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;",
		dbName,
	)

	err = db.Exec(createSQL).Error
	if err != nil {
		log.Fatal(err)
	}
}

func RegisterInjector(f func(*gorm.DB)) {
	injectors = append(injectors, f)
}

func callInjector(db *gorm.DB) {
	for _, v := range injectors {
		v(db)
	}
}

type ctxTransactionKey struct{}

func CtxWithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxTransactionKey{}, tx)
}

type txImpl struct{}

func NewTxImpl() *txImpl {
	return &txImpl{}
}

func (*txImpl) Transaction(ctx context.Context, fn func(txctx context.Context) error) error {
	db := globalDB.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		txctx := CtxWithTransaction(ctx, tx)
		return fn(txctx)
	})
}

// 如果使用跨模型事务则传参
func GetDB(ctx context.Context) *gorm.DB {
	iface := ctx.Value(ctxTransactionKey{})

	if iface != nil {
		tx, ok := iface.(*gorm.DB)
		if !ok {
			log.Panicf("unexpect context value type: %s", reflect.TypeOf(tx))
			return nil
		}

		return tx
	}

	return globalDB.WithContext(ctx)
}

func GetDBConfig() DBConfig {
	return *globalConfig
}

// 自动初始化表结构
func SetupTableModel(db *gorm.DB, model interface{}) {
	if GetDBConfig().AutoMigrate {
		err := db.AutoMigrate(model)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// https://github.com/ulid/spec
// uuid sortable by time
func NewUlid() string {
	now := time.Now()
	return ulid.MustNew(ulid.Timestamp(now), ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)).String()
}

func registerCallback(db *gorm.DB) {
	// 自动添加uuid
	err := db.Callback().Create().Before("gorm:create").Register("uuid", func(db *gorm.DB) {
		db.Statement.SetColumn("id", NewUlid())
	})
	if err != nil {
		log.Panicf("err: %+v", errx.WithStackOnce(err))
	}
}

// tag按首字母排序
type CommonModel struct {
	Id        string    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func WithOffsetLimit(db *gorm.DB, offset, limit int) *gorm.DB {
	if offset > 0 {
		db = db.Offset(offset)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	return db
}

# 在 Go 项目中优雅的使用 gorm v2

**本文基于 gorm v2 版本**

## 连接数据库

Go 里面也不用整什么单例了，直接用私有全局变量。

```go
func Connect(cfg *DBConfig) {
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
    
    globalDB = db
    
    log.Info("db connected success")
}
```

调用方使用 `GetDB` 从 globalDB 获取 gorm.DB 进行 CURD。`WithContext` 实际是调用 `db.Session(&Session{Context: ctx})`，每次创建新 Session，各 db 操作之间互不影响：

```go
func GetDB(ctx context.Context) *gorm.DB {
    return globalDB.WithContext(ctx)
}
```

## 自动创建数据表

_一般测试环境才这么玩，生产上推荐交给 DBA 处理，应用使用低权限账号_

gorm 提供 `db.AutoMigrate(model)` 方法自动建表 。现在我们想要实现数据库初始化后执行 `AutoMigrate`，并且可配置关闭 `AutoMigrate`。

项目中一般每个表一个 go 文件，model 相关的 CURD 都在一个文件中。如果用 init 初始化，则 db 必须在 init 执行前初始化，否则 init 执行时 db 还未初始。 使用 init 函数不是一个好的实践，一个包中多个 init 函数的执行顺序也是个坑。不用 init 则需要主动去调用每个表的初始化。有没有更好的方法呢？这里可以使用回调函数实现依赖反转，使用 init 注册回调函数，在 db 初始化之后再去执行所有回调函数，达到延迟执行的目的。代码如下:

```go
var injectors []func(db *gorm.DB)

// 注册回调
func RegisterInjector(f func(*gorm.DB)) {
	injectors = append(injectors, f)
}

// 执行回调
func callInjector(db *gorm.DB) {
	for _, v := range injectors {
		v(db)
	}
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
```

```go
// 调用方
func init() {
    dbcore.RegisterInjector(func(db *gorm.DB) {
        dbcore.SetupTableModel(db, &petmodel.Pet{})
    })
}
```

## 自动创建数据库

gorm 没有提供自动创建数据库的方法，这个我们通过  `CREATE DATABASE IF NOT EXISTS` SQL 语句来实现也非常简单：

```go
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
```

## 通过 Context 传递事务

在 DAO 层我们一般会封装对 model 增删改查等基本操作。每个方法都需要 db 作为参数，所以我们用面向对象的方式做一下封装。如下：

```go
type petDb struct {
    db *gorm.DB
}

func NewPetDb(ctx context.Context) struct {
	return GetDB(ctx)
}

func (s *petDb) Create(in *petmodel.Pet) error {
    return s.db.Create(in).Err
}

func (s *petDb) Update(in *petmodel.Pet) error {
    return s.db.Updates(in).Err
}
```

事务一般是在 Service 层，如果现在需要将多个 CURD 调用组成事务，如何复用 DAO 层的逻辑？我们很容易想到将 tx 作为参数传递到 DAO 层方法中即可。

如何优雅的传递 tx 参数？Go 里面没有重载，这种情况有个比较通用的方案：Context。**使用 Context 后续如果要做链路追踪、超时控制等也很方便扩展**。

[gorm 链路追踪可参考 github 上大佬的实现](https://github.com/avtion/gormTracing)


我们只需要把 `GetDB` 改改，尝试从 ctx 中获取 tx，如果存在则不需要新建 session，直接使用传递的 tx。这个有个小技巧，**使用结构体而不是字符串作为 ctx 的 key，可以保证 key 的唯一性**。代码如下：

```go
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
```

在事务上做一下 context 的封装：
```go
func Transaction(ctx context.Context, fc func(txctx context.Context) error) error {
	db := globalDB.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		txctx := CtxWithTransaction(ctx, tx)
		return fc(txctx)
	})
}
```

使用事务：
```go
ownerId := "xxx"
err := Transaction(context.Background(), func(txctx context.Context) error {
    pet, err := NewPetDb(txctx).Create(&petmodel.Pet{
        Name: "xxx",
        Age:  1,
        Sex:  "female",
    })
    if err != nil {
        return err
    }

    _, err = NewOwnerPetDb(txctx).Create(&petmodel.OwnerPet{
        OwnerId: ownerId,
        PetId:   pet.Id,
    })
    return err
})
```

## Hooks & Callbacks

gorm 提供 [Hooks](https://gorm.io/docs/hooks.html) 功能，可以在某些扩展点执行钩子函数，例如创建前生成 uuid ：

```go
func (u *Pet) BeforeCreate(tx *gorm.DB) error {
    u.Id = NewUlid()
    return nil
}
```

但是 Hooks 是针对某个 model，如果需要对所有 model，可以使用 [Callbacks](https://gorm.io/docs/write_plugins.html#Callbacks) 。

```go
func registerCallback(db *gorm.DB) {
    // 自动添加uuid
    err := db.Callback().Create().Before("gorm:create").Register("uuid", func (db *gorm.DB) {
        db.Statement.SetColumn("id", NewUlid())
    })
    if err != nil {
        log.Panicf("err: %+v", errx.WithStackOnce(err))
    }
}
```

项目完整代码：https://github.com/win5do/go-microservice-demo/tree/main/pkg/repository/db/pet

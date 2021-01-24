package dbcore

type DBConfig struct {
	DSN string // data source name

	MaxIdleConns int
	MaxOpenConns int
	AutoMigrate  bool // 自动建表，补全缺失字段，初始化数据
	Debug        bool
}

// 默认设置
func defaultDbConfig(cfg *DBConfig) *DBConfig {
	newCfg := *cfg

	if newCfg.MaxIdleConns == 0 {
		newCfg.MaxIdleConns = 10
	}

	if newCfg.MaxOpenConns == 0 {
		newCfg.MaxOpenConns = 20
	}

	return &newCfg
}

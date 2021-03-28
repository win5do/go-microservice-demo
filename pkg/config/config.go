package config

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"

	log "github.com/win5do/go-lib/logx"

	"github.com/win5do/go-lib/errx"

	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"

	"github.com/win5do/golang-microservice-demo/pkg/config/util"
)

var globalConfg *Config

type Config struct {
	AppName         string
	HttpPort        string
	GrpcGatewayPort string
	GrpcPort        string

	// https
	TlsCert string
	TlsKey  string

	Debug bool // debug log

	dbcore.DBConfig

	Ctx    context.Context
	Cancel context.CancelFunc
	/*
		根据环境变量配置jaeger，参考 https://github.com/jaegertracing/jaeger-client-go#environment-variables

		JAEGER_AGENT_HOST
		JAEGER_AGENT_PORT
	*/
	Tracer opentracing.Tracer
}

func DefaultConfig() *Config {
	ctx, cancel := util.NewWaitGroupCtx()
	return &Config{
		Ctx:     ctx,
		Cancel:  cancel,
		AppName: "server",
		DBConfig: dbcore.DBConfig{
			AutoMigrate: true,
		},
	}
}

func SetFlags(flagSet *pflag.FlagSet, cfg *Config) {
	flagSet.BoolVar(&cfg.Debug, "debug", false, "")
	flagSet.StringVar(&cfg.HttpPort, "http-port", "9010", "")
	flagSet.StringVar(&cfg.GrpcPort, "grpc-port", "9020", "")
	flagSet.StringVar(&cfg.GrpcGatewayPort, "grpc-gateway-port", "9030", "")
	flagSet.StringVar(&cfg.TlsCert, "tls-cert", "", "")
	flagSet.StringVar(&cfg.TlsKey, "tls-key", "", "")
	flagSet.StringVar(&cfg.DSN, "db-dsn", "root:123456@(127.0.0.1:3306)/go-demo", "")
}

func InitConfig(cfg *Config) error {
	var level zapcore.Level
	if cfg.Debug {
		level = zapcore.DebugLevel
		cfg.DBConfig.Debug = true
	} else {
		level = zapcore.InfoLevel
	}

	log.SetLogger(log.NewLogger(level))

	// jaeger
	err := SetupTrace(cfg.Ctx, cfg)
	if err != nil {
		return errx.WithStackOnce(err)
	}

	globalConfg = cfg
	log.Debugf("cfg: %+v", cfg)
	return nil
}

func GetConfig() *Config {
	return globalConfg
}

package config

import (
	"context"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/win5do/golang-microservice-demo/pkg/lib/errx"
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
	Tls_cert string
	Tls_key  string

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
	flagSet.StringVar(&cfg.Tls_cert, "tls-cert", "", "")
	flagSet.StringVar(&cfg.Tls_key, "tls-key", "", "")
	flagSet.StringVar(&cfg.DSN, "db-dsn", "root:123456@(127.0.0.1:3306)/go-demo", "")
}

func InitConfig(cfg *Config) error {
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
		cfg.DBConfig.Debug = true
	}

	// jaeger
	if err := SetupTrace(cfg.Ctx, cfg); err != nil {
		return errx.WithStackOnce(err)
	}

	globalConfg = cfg
	log.Debugf("cfg: %+v", cfg)
	return nil
}

func GetConfig() *Config {
	return globalConfg
}

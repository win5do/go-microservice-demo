package config

import (
	"context"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerzap "github.com/uber/jaeger-client-go/log/zap"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/win5do/go-lib/errx"

	"github.com/win5do/golang-microservice-demo/pkg/config/util"

	log "github.com/win5do/go-lib/logx"
)

func SetupTrace(ctx context.Context, cfg *Config) error {
	defer func() {
		cfg.Tracer = opentracing.GlobalTracer()
	}()

	isSet := func(env string) bool {
		_, ok := os.LookupEnv(env)
		return ok
	}

	if !(isSet("JAEGER_AGENT_HOST") ||
		isSet("JAEGER_ENDPOINT")) {
		return nil
	}

	jaegerCfg, err := jaegercfg.FromEnv()
	if err != nil {
		return errx.WithStackOnce(err)
	}

	if cfg.Debug {
		jaegerCfg.Sampler.Type = jaeger.SamplerTypeConst
		jaegerCfg.Sampler.Param = 1
		jaegerCfg.Reporter.LogSpans = true
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := jaegerCfg.InitGlobalTracer(
		cfg.AppName,
		jaegercfg.Logger(jaegerzap.NewLogger(log.GetLogger())),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Errorf("Could not initialize jaeger tracer: %s", err.Error())
		return errx.WithStackOnce(err)
	}

	wg := util.GetWaitGroupInCtx(ctx)
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		if err := closer.Close(); err != nil {
			log.Errorf("err: %+v", err)
		}
		log.Info("trace close")
	}()

	return nil
}

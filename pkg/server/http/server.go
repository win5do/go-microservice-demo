package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-gin/ginhttp"

	log "github.com/win5do/go-lib/logx"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/win5do/golang-microservice-demo/pkg/config"
	"github.com/win5do/golang-microservice-demo/pkg/config/util"
)

func Run(ctx context.Context, cfg *config.Config) {
	mux := SetupMux(cfg)

	server := &http.Server{
		Addr:    net.JoinHostPort("", cfg.HttpPort),
		Handler: mux,
	}

	if cfg.TlsCert != "" && cfg.TlsKey != "" {
		// https
		go func() {
			log.Infof("https server start: %v", server.Addr)
			cer, err := tls.LoadX509KeyPair(cfg.TlsCert, cfg.TlsKey)
			if err != nil {
				log.Errorf("failed to load certificate and key: %v", err)
				return
			}
			tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
			server.TLSConfig = tlsConfig

			if err := server.ListenAndServeTLS(cfg.TlsCert, cfg.TlsKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("err: %+v", err)
			}
		}()
	} else {
		go func() {
			log.Infof("http server start: %v", server.Addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("err: %+v", err)
			}
		}()
	}

	wg := util.GetWaitGroupInCtx(ctx)
	wg.Add(1)
	defer wg.Done()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown err: %+v", err)
		return
	}
	log.Info("http server shutdown")
}

func SetupMux(cfg *config.Config) http.Handler {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin
	mux := gin.Default()
	if cfg.Debug {
		// 打印body
		mux.Use(RequestLoggerMiddleware)
	}
	mux.Use(ginhttp.Middleware(cfg.Tracer))

	pprof.Register(mux) // default is "debug/pprof"
	Register(mux)

	return mux
}

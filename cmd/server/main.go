package main

import (
	goflag "flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/win5do/golang-microservice-demo/pkg/config"
	"github.com/win5do/golang-microservice-demo/pkg/config/util"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbinit"

	log "github.com/win5do/go-lib/logx"

	grpcserver "github.com/win5do/golang-microservice-demo/pkg/server/grpc"
	httpserver "github.com/win5do/golang-microservice-demo/pkg/server/http"
)

func main() {
	cfg := config.DefaultConfig()

	rootCmd := &cobra.Command{
		Use:   cfg.AppName,
		Short: "golang microservice demo",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := config.InitConfig(cfg)
			if err != nil {
				return err
			}

			// 连接数据库
			dbcore.Connect(&cfg.DBConfig)
			err = dbinit.InitData()
			if err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			Run(cfg)
		},
	}

	config.SetFlags(rootCmd.Flags(), cfg)
	rootCmd.Flags().AddGoFlagSet(goflag.CommandLine)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("err: %+v", err)
	}
}

func Run(cfg *config.Config) {
	ctx := cfg.Ctx
	defer func() {
		cfg.Cancel()
		log.Debug("cancel ctx")
		util.GetWaitGroupInCtx(ctx).Wait() // wait for goroutine cancel
	}()

	// http
	go httpserver.Run(ctx, cfg)

	// grpc
	go grpcserver.Run(ctx, cfg)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutdown server ...")
}

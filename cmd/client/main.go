package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"

	"github.com/win5do/go-lib/errx"
	log "github.com/win5do/go-lib/logx"

	"github.com/win5do/golang-microservice-demo/pkg/api/petpb"
)

func main() {
	var service string

	rootCmd := &cobra.Command{
		Use:   "client",
		Short: "grpc client",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLogger(log.NewLogger(zapcore.DebugLevel))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(service)
		},
	}

	rootCmd.Flags().StringVar(&service, "service", "", "headless service address")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("err: %+v", err)
	}
}

func run(addr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	conn, err := grpc.DialContext(ctx, "dns:///"+addr,
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithBlock(),
	)
	cancel()
	if err != nil {
		return errx.WithStackOnce(err)
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	for t := range ticker.C {
		client := petpb.NewPetServiceClient(conn)
		resp, err := client.Ping(context.Background(), &petpb.Id{
			Id: echo(t),
		})
		if err != nil {
			return errx.WithStackOnce(err)
		}

		log.Infof("resp: %s", resp.Id)
	}

	return nil
}

func echo(t time.Time) string {
	host, err := os.Hostname()
	if err != nil {
		log.Panicf("err: %+v", err)
	}
	return fmt.Sprintf("%s, %d", host, t.Second())
}

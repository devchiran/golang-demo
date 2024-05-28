package main

import (
	"context"
	"fmt"
	"golang-demo/internal/http"
	"golang-demo/internal/postgres"
	"log"
	"os"
	"syscall"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/kelseyhightower/envconfig"
	"github.com/twitsprout/tools"
	httputils "github.com/twitsprout/tools/http"
	"github.com/twitsprout/tools/lifecycle"
	"github.com/twitsprout/tools/zap"
)

var version string

type variables struct {
	Addr         string `required:"true" envconfig:"addr"`
	PostgresHost string `required:"true" envconfig:"postgres_host"`
	PostgresPort int    `required:"false" envconfig:"postgres_port"`
	PostgresDB   string `required:"true" envconfig:"postgres_db"`
	PostgresUser string `required:"true" envconfig:"postgres_user"`
	PostgresPass string `required:"true" envconfig:"postgres_pass"`
	LogLevel     string `required:"false" envconfig:"log_level"`
	AppName      string `required:"true" envconfig:"app_name"`
}

var v variables

func init() {
	if metadata.OnGCE() {
		port := os.Getenv("PORT")
		err := os.Setenv("ADDR", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
	}

	envconfig.MustProcess("golang-demo", &v)
	fmt.Println("Env variables :", v)
	if v.LogLevel == "" {
		v.LogLevel = "info"
	}
}

func main() {
	logger := zap.New("golang-demo", version, os.Stdout)
	if err := logger.SetLevel(v.LogLevel); err != nil {
		logger.Error("failed to set log level", "error", err.Error())
	}

	pg := newPostgres(v, nil)

	ctx := context.Background()

	lc, ctx := lifecycle.New(ctx, logger)
	lc.Start("golang-demo root context", func() error {
		<-ctx.Done()
		return ctx.Err()
	})

	h := http.Handler{
		Logger:     logger,
		Version:    version,
		AlbumStore: pg,
		AppName:    v.AppName,
	}
	server := httputils.NewServer(v.Addr, h.Handler())
	lc.StartServer(server)
	lc.StartSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	_ = lc.Wait(15 * time.Second)
}

func newPostgres(v variables, sc tools.StatsClient) *postgres.Postgres {
	pgConfig := postgres.Config{
		Host:       v.PostgresHost,
		Name:       v.PostgresDB,
		Password:   v.PostgresPass,
		Username:   v.PostgresUser,
		DisableSSL: true,
	}
	// Only use a Postgres port if one was provided
	if v.PostgresPort > 0 {
		pgConfig.Port = v.PostgresPort
	}
	pg, err := postgres.New(pgConfig, sc)
	if err != nil {
		panic(err)
	}
	return pg
}

package main

import (
	"fmt"
	golog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EgorRepnikov/proxmox-api-commands/internal/app/healthcheck"
	"github.com/EgorRepnikov/proxmox-api-commands/internal/app/proxmox_api_commands"
	"github.com/EgorRepnikov/proxmox-api-commands/internal/pkg/env"
	"github.com/EgorRepnikov/proxmox-api-commands/internal/pkg/http_client"
	"github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func main() {
	if err := run(); err != nil {
		golog.Println("main : shutting down", "err: ", err)
		os.Exit(1)
	}
}

func run() error {
	golog := golog.New(os.Stdout, "", golog.LstdFlags|golog.Lmicroseconds|golog.Lshortfile)

	golog.Printf("main : started")
	defer golog.Println("main : completed")

	// Init Env
	env, err := env.InitEnv()
	if err != nil {
		return errors.Wrap(err, "cannot get env")
	}

	initZeroLogger(env)

	// Router
	router := router.New()

	// HttpClient
	httpClient := http_client.InitHttpClient()

	// Handlers
	healthcheck.InitHealthcheck(router)
	proxmox_api_commands.InitProxmoxApiCommands(router, env, httpClient)

	port := env.PORT
	addr := fmt.Sprintf(":%d", port)
	log.Info().Int("port", port).Msg("Starting fasthttp server...")

	server := fasthttp.Server{
		NoDefaultDate:         true,
		NoDefaultServerHeader: true,
		IdleTimeout:           5 * time.Minute,
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          5 * time.Second,
		Concurrency:           1000000,
		ReadBufferSize:        40960,
		WriteBufferSize:       40960,
	}

	server.Handler = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Vary", "Origin")
		ctx.Response.Header.Add("Cache-Control", "no-transform, no-cache, no-store, must-revalidate")
		ctx.Response.Header.Add("Pragma", "no-cache")
		ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "*")

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		router.Handler(ctx)
	}

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		<-gracefulShutdown
		log.Info().Msg("Graceful Shutdown")
		server.Shutdown()
	}()

	if err := server.ListenAndServe(addr); err != nil {
		log.Error().Int("port", port).Err(err).Msgf("error start fasthttp server")
	}

	return nil
}

func initZeroLogger(env *env.Env) {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.999"

	lvl, err := zerolog.ParseLevel(env.LOGGER_LEVEL)
	if err != nil {
		golog.Println("init zero logger : error parse config level", "err: ", err)
		lvl = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(lvl)
}

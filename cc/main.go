package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/exerr"
	"gogs.mikescher.com/BlackForestBytes/goext/ginext"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var commands = make([]Command, 0)
var gil = sync.Mutex{}

func main() {
	cw := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05 Z07:00",
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	multi := zerolog.MultiLevelWriter(cw)
	logger := zerolog.New(multi).With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger

	gin.SetMode(gin.DebugMode)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	exerr.Init(exerr.ErrorPackageConfigInit{})

	engine := ginext.NewEngine(ginext.Options{})

	engine.Routes().POST("/").Handle(popCommands)
	engine.Routes().GET("/").Handle(peekCommands)
	engine.Routes().PUT("/").Handle(addCommands)

	routerErr, router := engine.ListenAndServeHTTP("0.0.0.0:8000", func(port string) {})

	sigstop := make(chan os.Signal, 1)
	signal.Notify(sigstop, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case err := <-routerErr:
		log.Error().Err(err).Msg("router finished with error")
		panic(err)
	case <-sigstop:
		log.Info().Msg("Received SIGTERM/SIGKILL - Stopping all routers+app")
		err := router.Shutdown(context.Background())
		if err != nil {
			log.Err(err).Msg("failed to stop reverse-router")
		}
	}

	log.Info().Msg("Server finished.")
}

func addCommands(pctx ginext.PreContext) ginext.HTTPResponse {
	type body []Command

	t0 := time.Now()

	var b body
	_, gctx, errResp := pctx.Body(&b).Start()
	if errResp != nil {
		return *errResp
	}

	log.Debug().Msgf("[REQ]  <add> from %s", gctx.ClientIP())

	for _, cmd := range b {
		if msg, ok := cmd.Valid(); !ok {
			return ginext.Text(400, msg)
		}
	}

	gil.Lock()
	defer gil.Unlock()

	for _, cmd := range b {
		cmd.Date = t0
		commands = append(commands, cmd)

		log.Info().Msgf("[ADD]  %s", cmd.String())
	}

	return ginext.Text(200, fmt.Sprintf("OK - added %d commands, %d commands enqueued", len(b), len(commands)))
}

func popCommands(pctx ginext.PreContext) ginext.HTTPResponse {
	_, gctx, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	log.Debug().Msgf("[REQ]  <pop> from %s", gctx.ClientIP())

	gil.Lock()
	defer gil.Unlock()

	commands = langext.ArrFilter(commands, func(cmd Command) bool { return time.Now().Unix()-cmd.Date.Unix() < 60 }) // remove commands older than 60s

	resp := ""
	for _, cmd := range commands {
		resp += cmd.String() + "\n"
		log.Info().Msgf("[POP]  %s", cmd.String())
	}

	commands = make([]Command, 0)

	return ginext.Text(200, resp)
}

func peekCommands(pctx ginext.PreContext) ginext.HTTPResponse {
	_, gctx, errResp := pctx.Start()
	if errResp != nil {
		return *errResp
	}

	log.Debug().Msgf("[REQ]  <peek> from %s", gctx.ClientIP())

	gil.Lock()
	defer gil.Unlock()

	commands = langext.ArrFilter(commands, func(cmd Command) bool { return time.Now().Unix()-cmd.Date.Unix() < 60 }) // remove commands older than 60s

	resp := ""
	for _, cmd := range commands {
		resp += cmd.String() + "\n"
		log.Info().Msgf("[PEEK] %s", cmd.String())
	}

	return ginext.Text(200, resp)
}

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/app"
)

func main() {
	_ = godotenv.Load()

	zerolog.TimeFieldFormat = time.RFC3339
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen})

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {

		dsn = "host=localhost user=postgres password=postgres dbname=tienda3d port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		zlog.Fatal().Err(err).Msg("abriendo DB")
	}

	application, err := app.NewApp(db)
	if err != nil {
		zlog.Fatal().Err(err).Msg("init app")
	}
	if err := application.MigrateAndSeed(); err != nil {
		zlog.Fatal().Err(err).Msg("migrar/seed")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {

		for p := 8081; p <= 8090; p++ {
			alt := net.JoinHostPort("", fmt.Sprintf("%d", p))
			l2, err2 := net.Listen("tcp", alt)
			if err2 == nil {
				zlog.Info().Str("old_port", port).Str("port", fmt.Sprint(p)).Msg("puerto en uso, usando alternativo")
				ln = l2
				port = fmt.Sprint(p)
				break
			}
		}
		if ln == nil {
			zlog.Fatal().Err(err).Msg("no se pudo enlazar puerto")
		}
	}

	server := &http.Server{Handler: application.HTTPHandler()}

	go func() {
		zlog.Info().Str("port", port).Msg("escuchando")
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msg("server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	zlog.Info().Msg("shutdown OK")
}

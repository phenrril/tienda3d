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

	"github.com/phenrril/tienda3d/internal/adapters/backup"
	"github.com/phenrril/tienda3d/internal/app"
)

func main() {
	_ = godotenv.Load()

	zerolog.TimeFieldFormat = time.RFC3339
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen})

	// Obtener variables de entorno de la base de datos
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if password == "" {
		password = "postgres"
	}
	if dbname == "" {
		dbname = "tienda3d"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Construir DSN desde variables individuales
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, dbPort)
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

	// Configurar servicio de backup
	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = `C:\Users\server\Desktop\backup-db`
	}
	backupService := backup.NewService(backupDir, host, dbPort, user, password, dbname)
	backupScheduler := backup.NewScheduler(backupService)

	// Crear contexto para el scheduler de backup
	backupCtx, backupCancel := context.WithCancel(context.Background())
	defer backupCancel()

	// Iniciar scheduler de backup
	go func() {
		if err := backupScheduler.Start(backupCtx); err != nil {
			zlog.Error().Err(err).Msg("error iniciando scheduler de backup")
		}
	}()

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

	// Detener scheduler de backup
	backupCancel()
	backupScheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	zlog.Info().Msg("shutdown OK")
}

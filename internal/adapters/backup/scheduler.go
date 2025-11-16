package backup

import (
	"context"
	"os"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	cron    *cron.Cron
	service *Service
}

func NewScheduler(service *Service) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		service: service,
	}
}

// Start inicia el scheduler con el cron job
func (s *Scheduler) Start(ctx context.Context) error {
	// Verificar si debe ejecutarse la primera vez
	// Buscar la variable en diferentes formatos (case-insensitive)
	backupEnv := os.Getenv("Backup")
	if backupEnv == "" {
		backupEnv = os.Getenv("BACKUP")
	}
	if backupEnv == "" {
		backupEnv = os.Getenv("backup")
	}
	
	shouldRunFirstTime := backupEnv != ""
	
	log.Info().
		Str("Backup_env_value", backupEnv).
		Bool("should_run_first_time", shouldRunFirstTime).
		Msg("verificando configuración de backup inicial")

	// Iniciar el servicio de backup
	if err := s.service.Start(shouldRunFirstTime); err != nil {
		log.Error().Err(err).Msg("error iniciando servicio de backup")
		return err
	}

	// Configurar cron para ejecutar todos los días a las 00:00:00
	// Formato cron con segundos: segundo minuto hora día mes día_semana
	// "0 0 0 * * *" = todos los días a las 00:00:00
	_, err := s.cron.AddFunc("0 0 0 * * *", func() {
		log.Info().Msg("ejecutando backup programado (cron)")
		if err := s.service.PerformBackup(); err != nil {
			log.Error().Err(err).Msg("error en backup programado")
		}
	})

	if err != nil {
		return err
	}

	// Iniciar el cron
	s.cron.Start()
	log.Info().Str("schedule", "0 0 0 * * *").Msg("cron de backup configurado (diario a las 00:00)")

	// Esperar hasta que el contexto se cancele
	go func() {
		<-ctx.Done()
		log.Info().Msg("deteniendo scheduler de backup")
		s.cron.Stop()
	}()

	return nil
}

// Stop detiene el scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Info().Msg("scheduler de backup detenido")
}


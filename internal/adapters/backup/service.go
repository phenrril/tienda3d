package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

type Service struct {
	backupDir    string
	dbHost       string
	dbPort       string
	dbUser       string
	dbPassword   string
	dbName       string
	initialized  bool
	firstRunDone bool
}

func NewService(backupDir, dbHost, dbPort, dbUser, dbPassword, dbName string) *Service {
	return &Service{
		backupDir:   backupDir,
		dbHost:      dbHost,
		dbPort:      dbPort,
		dbUser:      dbUser,
		dbPassword:  dbPassword,
		dbName:      dbName,
		initialized: false,
	}
}

// Start inicia el servicio de backup y configura el cron
func (s *Service) Start(shouldRunFirstTime bool) error {
	// Crear el directorio de backup si no existe
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de backup: %w", err)
	}

	log.Info().Str("backup_dir", s.backupDir).Msg("servicio de backup iniciado")

	// Ejecutar la primera vez si está configurado
	if shouldRunFirstTime {
		log.Info().Msg("ejecutando backup inicial (primera vez)")
		if err := s.performBackup(); err != nil {
			log.Error().Err(err).Msg("error en backup inicial")
			return err
		}
		s.firstRunDone = true
		log.Info().Msg("backup inicial completado")
	}

	return nil
}

// PerformBackup ejecuta el backup de la base de datos
func (s *Service) PerformBackup() error {
	return s.performBackup()
}

func (s *Service) performBackup() error {
	// Generar nombre de archivo con timestamp
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	filepath := filepath.Join(s.backupDir, filename)

	log.Info().Str("file", filepath).Msg("iniciando backup de base de datos")

	// Configurar variables de entorno para pg_dump
	env := os.Environ()
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	// Construir comando pg_dump
	cmd := exec.Command("pg_dump",
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"-f", filepath,
		"--no-owner",
		"--no-acl",
	)

	cmd.Env = env

	// Ejecutar el comando
	// pg_dump escribe directamente al archivo cuando se usa -f, así que solo capturamos stderr
	cmd.Stdout = os.Stderr // Redirigir stdout a stderr para ver progreso si hay
	err := cmd.Run()
	if err != nil {
		log.Error().
			Err(err).
			Msg("error ejecutando pg_dump")
		return fmt.Errorf("error ejecutando pg_dump: %w", err)
	}

	// Obtener información del archivo creado
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		log.Warn().Err(err).Msg("no se pudo obtener información del archivo de backup")
	} else {
		log.Info().
			Str("file", filepath).
			Int64("size_bytes", fileInfo.Size()).
			Msg("backup completado exitosamente")
	}

	return nil
}


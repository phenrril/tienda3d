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
	backupDir     string
	dbHost        string
	dbPort        string
	dbUser        string
	dbPassword    string
	dbName        string
	containerName string
	initialized   bool
	firstRunDone  bool
}

func NewService(backupDir, dbHost, dbPort, dbUser, dbPassword, dbName string) *Service {
	containerName := os.Getenv("DB_CONTAINER_NAME")
	if containerName == "" {
		containerName = "tienda3d_db"
	}
	return &Service{
		backupDir:     backupDir,
		dbHost:        dbHost,
		dbPort:        dbPort,
		dbUser:        dbUser,
		dbPassword:    dbPassword,
		dbName:        dbName,
		containerName: containerName,
		initialized:   false,
	}
}

// VerifyDockerAccess verifica que Docker esté disponible y que se pueda acceder al contenedor
func (s *Service) VerifyDockerAccess() error {
	// Verificar que docker está disponible
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker no está disponible en el PATH: %w", err)
	}

	// Verificar que docker está corriendo
	cmd := exec.Command("docker", "ps")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("no se puede ejecutar docker ps (¿está Docker corriendo?): %w", err)
	}

	// Verificar que el contenedor existe y está corriendo
	cmd = exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", s.containerName), "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error verificando contenedor %s: %w", s.containerName, err)
	}

	containerStatus := string(output)
	if containerStatus == "" {
		return fmt.Errorf("el contenedor %s no está corriendo. Verifica con 'docker ps'", s.containerName)
	}

	log.Info().
		Str("container", s.containerName).
		Msg("contenedor Docker verificado correctamente")

	// Verificar que podemos ejecutar comandos dentro del contenedor
	// Intentamos ejecutar un comando simple para verificar permisos
	cmd = exec.Command("docker", "exec", s.containerName, "echo", "test")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no se pueden ejecutar comandos en el contenedor %s (verifica permisos de Docker): %w", s.containerName, err)
	}

	log.Info().
		Str("container", s.containerName).
		Msg("permisos de Docker verificados correctamente")

	return nil
}

// Start inicia el servicio de backup y configura el cron
func (s *Service) Start(shouldRunFirstTime bool) error {
	// Verificar acceso a Docker antes de continuar
	if err := s.VerifyDockerAccess(); err != nil {
		log.Error().Err(err).Msg("verificación de Docker falló")
		return fmt.Errorf("verificación de Docker falló: %w", err)
	}

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

	log.Info().
		Str("file", filepath).
		Str("container", s.containerName).
		Msg("iniciando backup de base de datos desde contenedor Docker")

	// Crear el archivo de salida
	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creando archivo de backup: %w", err)
	}

	// Construir comando docker exec con pg_dump dentro del contenedor
	// Usamos PGPASSWORD como variable de entorno dentro del contenedor
	cmd := exec.Command("docker", "exec",
		"-e", fmt.Sprintf("PGPASSWORD=%s", s.dbPassword),
		s.containerName,
		"pg_dump",
		"-U", s.dbUser,
		"-d", s.dbName,
		"--no-owner",
		"--no-acl",
	)

	// Redirigir stdout al archivo y stderr para ver errores
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	// Ejecutar el comando
	err = cmd.Run()

	// Cerrar el archivo siempre
	outFile.Close()

	if err != nil {
		// Eliminar el archivo si hay error
		os.Remove(filepath)
		log.Error().
			Err(err).
			Str("container", s.containerName).
			Msg("error ejecutando pg_dump en contenedor Docker")
		return fmt.Errorf("error ejecutando pg_dump en contenedor %s: %w", s.containerName, err)
	}

	// Obtener información del archivo creado
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		log.Warn().Err(err).Msg("no se pudo obtener información del archivo de backup")
	} else {
		log.Info().
			Str("file", filepath).
			Int64("size_bytes", fileInfo.Size()).
			Str("container", s.containerName).
			Msg("backup completado exitosamente")
	}

	return nil
}

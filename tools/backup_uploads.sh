#!/usr/bin/env sh
set -eu

# Configuración
SRC_DIR="/data/uploads"           # dentro del contenedor backup, se montará ./uploads aquí
BACKUP_DIR="/data/backups"        # dentro del contenedor backup, se montará ./backups aquí
RETENTION="14"                    # cantidad de snapshots a conservar

TZ=${TZ:-America/Argentina/Buenos_Aires}
DATE_TS=$(TZ="$TZ" date +%F_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Si no existe SRC_DIR o está vacío, salimos con 0 para no romper cron
if [ ! -d "$SRC_DIR" ]; then
  echo "WARN: SRC_DIR $SRC_DIR no existe" >&2
  exit 0
fi

# Generar snapshot comprimido
ARCHIVE="$BACKUP_DIR/uploads-$DATE_TS.tar.gz"
TMP_ARCHIVE="$ARCHIVE.inprogress"

echo "Creando $ARCHIVE ..."
# -C para que el TAR contenga la carpeta uploads/
cd /data
if [ -d "uploads" ]; then
  tar -czf "$TMP_ARCHIVE" uploads
  mv "$TMP_ARCHIVE" "$ARCHIVE"
  # checksum
  sha256sum "$ARCHIVE" > "$ARCHIVE.sha256" || true
else
  echo "WARN: /data/uploads no existe en runtime" >&2
fi

# Rotación (mantener últimos N por fecha)
# Nota: ls -1t ordena por mtime; grep asegura filtrar por prefijo correcto
COUNT=$(ls -1 "$BACKUP_DIR"/uploads-*.tar.gz 2>/dev/null | wc -l || true)
if [ "${COUNT:-0}" -gt "$RETENTION" ]; then
  # eliminar los más viejos excedentes
  ls -1t "$BACKUP_DIR"/uploads-*.tar.gz | tail -n +$((RETENTION+1)) | xargs -r rm -f
  ls -1t "$BACKUP_DIR"/uploads-*.tar.gz.sha256 | tail -n +$((RETENTION+1)) | xargs -r rm -f
fi

echo "Backup OK: $ARCHIVE"
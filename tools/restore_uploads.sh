#!/usr/bin/env sh
set -eu

# Uso: ./restore_uploads.sh /data/backups/uploads-YYYY-MM-DD_HHMMSS.tar.gz
# Requiere que /data/uploads y /data/backups estén montados en el contenedor de utilidades o en el host.

if [ "${1:-}" = "" ]; then
  echo "Uso: $0 /data/backups/uploads-YYYY-MM-DD_HHMMSS.tar.gz" >&2
  exit 1
fi
ARCHIVE="$1"

if [ ! -f "$ARCHIVE" ]; then
  echo "No existe archivo: $ARCHIVE" >&2
  exit 1
fi

# Validar checksum si existe
if [ -f "$ARCHIVE.sha256" ]; then
  sha256sum -c "$ARCHIVE.sha256" || {
    echo "Checksum invalido para $ARCHIVE" >&2
    exit 1
  }
fi

# Extraer sobre /data (sobrescribe uploads/)
cd /data
# opción segura: extraer a tmp y luego mover
TMPDIR="restore_tmp_$$"
mkdir -p "$TMPDIR"
tar -xzf "$ARCHIVE" -C "$TMPDIR"
# Debe existir $TMPDIR/uploads después de extraer
if [ ! -d "$TMPDIR/uploads" ]; then
  echo "El backup no contiene carpeta uploads/" >&2
  exit 1
fi
# Mover actual a uploads.bak con timestamp por si quieres revertir
TS=$(date +%F_%H%M%S)
if [ -d uploads ]; then
  mv uploads "uploads.bak.$TS"
fi
mv "$TMPDIR/uploads" ./uploads
rm -rf "$TMPDIR"

echo "Restaurado desde $ARCHIVE. Copia anterior en uploads.bak.$TS (si existía)."
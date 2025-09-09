# Tienda3D

Proyecto de tienda online para venta de modelos 3D impresos, desarrollado en Go (arquitectura hexagonal + SSR). 

## Objetivo
Permitir la carga, visualización y compra de productos impresos en 3D, con integración a MercadoPago, manejo de imágenes y flujo de checkout completo.

## Novedades recientes
- Carrusel en detalle de producto (autoplay, flechas, miniaturas, swipe, teclado) + layout responsive.
- Selector de color (swatches) con fallback de colores si el producto no define variantes.
- Formulario de checkout rediseñado (gradientes, focus visible, layout responsive).
- Costos de envío unificados (actualmente todos 9000 en `provinceCosts` dentro de `server.go`).
- Endpoint multipart `POST /api/products/upload` para crear producto + múltiples imágenes en una sola request.
- Eliminación completa de producto (DELETE `/api/products/{slug}`) borra registros + archivos físicos en `uploads/`.
- Borrado masivo simple (`POST /api/products/delete`).
- Al subir imágenes se guardan en `uploads/images/...` (storage local) y se sirven vía `/uploads/...`.
- Integración opcional: OAuth Google (si se configuran las variables), notificaciones a Telegram y/o email (SMTP).
- Docker Compose parametrizado por `.env` (sin secretos en el YAML).

## Arquitectura y Estructura
- **Go nativo**: servidor HTTP + middlewares propios.
- **GORM + Postgres**: persistencia.
- **html/template (SSR)**: vistas accesibles y rápidas.
- **MercadoPago**: generación de preferencias (sandbox o prod según token y `APP_ENV`).
- **Storage local**: archivos en carpeta configurable (`STORAGE_DIR`, por defecto `uploads`).
- **Notificaciones**: Telegram y opcional email (SMTP) tras pago aprobado.
- **OAuth Google**: login rápido (opcional).

```
/cmd/tienda3d          # main real
/internal/app          # wiring, inicialización, migraciones
/internal/domain       # entidades + puertos
/internal/usecase      # casos de uso
/internal/adapters     # http, repos, payments, storage
/internal/views        # templates SSR
/public/assets         # estáticos (css, imágenes públicas)
/uploads               # imágenes subidas (dinámicas)
```

## Variables de entorno (completas)
Obligatorias mínimas para entorno local:
- `DB_DSN` cadena Postgres (si usás docker compose se arma con variables POSTGRES_*)
- `SESSION_KEY` clave aleatoria segura (firmar cookies)
- `MP_ACCESS_TOKEN` token MercadoPago (TEST-... o prod)

Recomendadas / adicionales:
- `PORT` (default 8080)
- `APP_ENV` (`dev` | `production`) controla selección de token MP y logs
- `PROD_ACCESS_TOKEN` token MP producción (si querés separar del TEST)
- `PUBLIC_BASE_URL` URL pública (para back_urls y webhooks de MP)
- `BASE_URL` (usado también para OAuth Google)
- `SECRET_KEY` firma del external_reference MP (fallback "dev")
- `STORAGE_DIR` carpeta para archivos subidos (default `uploads`)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `ORDER_NOTIFY_EMAIL` (notificación email)
- `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` (notificación Telegram)
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` (OAuth Google)

Docker / DB:
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `DB_PORT`, `APP_PORT`

MercadoPago alterno:
- `PROD_ACCESS_TOKEN` (solo si querés tener sandbox + prod separados)

### Ejemplo `.env.example`
```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=change_me
POSTGRES_DB=tienda3d
DB_PORT=5432
APP_PORT=8080
DB_DSN=postgres://postgres:change_me@db:5432/tienda3d?sslmode=disable
SESSION_KEY=generate_secure_key
MP_ACCESS_TOKEN=TEST-xxxxxxxxxxxxxxxxxxxx
APP_ENV=dev
PUBLIC_BASE_URL=http://localhost:8080
BASE_URL=http://localhost:8080
STORAGE_DIR=uploads
SECRET_KEY=sign_ref_key
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
ORDER_NOTIFY_EMAIL=
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

## Docker Compose
Ahora parametrizado (ver `docker-compose.yml`):
```
docker compose up -d --build
```
Asegurate de crear `.env` a partir de `.env.example` antes de levantar.

## Flujos principales
### 1. Carga de productos
Opción recomendada: `POST /api/products/upload` (multipart) con uno o más campos `image` / `images`. El backend guarda cada archivo en `uploads/images/<timestamp>-<filename>` y registra las rutas.

### 2. Visualización
- `/products` listado con filtros.
- `/product/{slug}`: carrusel + swatches de color (si hay variantes se deducen los colores; de lo contrario se usa un set por defecto).

### 3. Carrito y Checkout
- Agregar desde el detalle (envía `slug` + `color`).
- Carrito `/cart`: editar cantidades, elegir envío o retiro. Todos los costos provinciales actualmente son 9000 (configurar en código `provinceCosts`).
- Checkout: botón MercadoPago genera preferencia (sandbox si token `TEST-` y no estás en producción).

### 4. Pagos y Webhooks
- Webhook MP: `/webhooks/mp` (configurar en MercadoPago a `PUBLIC_BASE_URL/webhooks/mp`).
- Página de estado `/pay/{orderID}` se usa como success/pending/failure.

### 5. Eliminación de productos
- `DELETE /api/products/{slug}` elimina DB + archivos.
- `POST /api/products/delete` borrado masivo simple (no borra archivos físicos).

## Endpoints principales
Web: `/`, `/products`, `/product/{slug}`, `/cart`, `/checkout`, `/pay/{id}`
API: 
- `POST /api/products/upload`
- `POST /api/products`
- `GET /api/products`
- `GET /api/products/{slug}`
- `DELETE /api/products/{slug}`
- `POST /api/products/delete`
- `POST /api/quote` (cotización desde modelo subido)
- `POST /api/checkout`
- `POST /webhooks/mp`

## Desarrollo
```
go run ./cmd/tienda3d
# o
make dev
```
Docker:
```
docker compose up -d --build
```

## Pruebas manuales rápidas
1. Subir un producto (ver `carga.md`).
2. Ver en `/products`.
3. Agregar al carrito, cambiar cantidades.
4. Simular pago (usar token TEST) -> redirección a MP sandbox -> completar -> volver a `/pay/{id}`.
5. Ver logs para confirmar webhook.

## Notas
- No se crean productos demo automáticamente.
- Las imágenes dinámicas viven fuera de `/public` en `uploads/`.
- El slug se recalcula si faltaba (backfill en migración).
- Color picker: si no hay variantes, se muestran colores por defecto.
- Para producción: configurar `APP_ENV=production`, `PROD_ACCESS_TOKEN`, dominio HTTPS en `PUBLIC_BASE_URL`.

---
Ver `carga.md` para detalles de carga y ejemplos curl.

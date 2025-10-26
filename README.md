# Chroma3D - Tienda Online de Impresión 3D

Proyecto completo de tienda online para venta de modelos 3D impresos, desarrollado en Go con arquitectura hexagonal + SSR (Server-Side Rendering).

## Objetivo
Permitir la carga, visualización y compra de productos impresos en 3D, con integración a MercadoPago, manejo de imágenes y flujo de checkout completo. Sistema completo de e-commerce orientado a impresión 3D mayorista y minorista.

## 🚀 Features Principales

### 🛍️ Catálogo y Productos
- **Gestión completa de productos** con imágenes múltiples
- **Carrusel de imágenes** en detalle de producto (autoplay, flechas, miniaturas, swipe táctil, navegación por teclado)
- **Selector de color (swatches)** con fallback automático si el producto no define variantes
- **Variantes de productos** con diferentes colores y precios
- **Categorización** de productos (accesorios, iluminación, hogar, oficina, jardín, cocina, etc.)
- **Búsqueda y filtrado** avanzado de productos
- **Ordenamiento** por precio, relevancia, fecha
- **Paginación** eficiente
- **Productos destacados** (ready to ship)
- **Dimensiones** de producto (ancho, alto, profundidad en mm)

### 🛒 Carrito y Checkout
- **Carrito persistente** con cookies firmadas
- **Edición de cantidades** en el carrito
- **Selección de método de envío** (retiro o envío a domicilio)
- **Cálculo automático de costos** por provincia
- **Selector de provincia** para cálculo de envío
- **Resumen de orden** antes del pago
- **Formulario de checkout** optimizado y responsive
- **Validación de datos** en checkout

### 💳 Pagos
- **Integración completa con MercadoPago** (sandbox y producción)
- **Generación automática de preferencias** de pago
- **Webhooks** para notificaciones de pago
- **Estados de pago** (pending, approved, rejected, etc.)
- **Página de confirmación** de pago (`/pay/{orderID}`)
- **Back URLs** configurable (success, pending, failure)
- **Firma de seguridad** en external_reference

### 📦 Gestión de Órdenes
- **Panel administrativo** de órdenes
- **Listado paginado** de órdenes
- **Filtros y búsqueda** de órdenes
- **Estados de orden** (awaiting_pay, finished, cancelled)
- **Tracking de MercadoPago** status
- **Notificaciones automáticas** al confirmar pago
- **Historial completo** de pedidos

### 🎨 Interfaz de Usuario
- **Diseño responsive** adaptativo para móvil, tablet y desktop
- **Modo oscuro** optimizado
- **UI moderna** con gradientes y animaciones suaves
- **Navegación intuitiva** con drawer y sheets en móvil
- **Búsqueda en header** rápida y accesible
- **Botón flotante de WhatsApp** en móvil
- **Hero carousel** con autoplay y swipe
- **Modales informativos** con animaciones
- **Transiciones fluidas** entre vistas
- **Soporte para reduced-motion** (accesibilidad)

### 🔐 Autenticación y Seguridad
- **Sistema de autenticación Admin con JWT**
- **API Key** para login admin
- **Lista de emails permitidos** (ADMIN_ALLOWED_EMAILS)
- **Tokens con expiración** (30 minutos por defecto)
- **OAuth Google** (opcional, configurable)
- **Sesiones seguras** con cookies firmadas
- **Protección CSRF** implícita
- **Rate limiting** por endpoints
- **CORS configurado**

### 📧 Notificaciones
- **Notificaciones por Telegram** (simple o múltiples destinos)
- **Notificaciones por Email** (SMTP)
- **Notificaciones combinadas** (Telegram + Email)
- **Activación automática** tras pago aprobado
- **Soporte para múltiples chats** en Telegram

### 📱 WhatsApp Business
- **Integración con WhatsApp Business API**
- **Webhook de WhatsApp** para órdenes
- **Sincronización de productos** con catálogo de WhatsApp
- **Gestión de órdenes desde WhatsApp**
- **Creación automática de órdenes** desde mensajes
- **Herramientas de sincronización** (`tools/whatsapp_sync.go`)
- **Exportación de productos** a formato WhatsApp

### 👨‍💼 Panel Administrativo
- **Dashboard administrativo** completo
- **Gestión de productos** (crear, editar, eliminar, listar)
- **Gestión de imágenes** de productos
- **Reparación de imágenes huérfanas** (cleanup automático)
- **Herramienta de cálculo de costos** de impresión
- **Vista de ventas** y estadísticas
- **Gestor de órdenes** avanzado
- **Upload multipart** de productos + imágenes
- **Borrado masivo** de productos
- **Borrado completo** con limpieza de archivos

### 🖼️ Gestión de Archivos
- **Storage local** configurable (`STORAGE_DIR`)
- **Subida múltiple** de imágenes
- **Nombres únicos** con timestamp
- **Servido optimizado** de archivos estáticos
- **Limpieza automática** de archivos huérfanos
- **Soporte para imágenes** optimizadas (WebP recomendado)
- **Redimensionamiento** de imágenes (responsive)

### ⚡ Performance y Optimización
- **Server-Side Rendering** (SSR) con html/template
- **Compresión Gzip** automática
- **Caché de archivos estáticos**
- **Preload de fuentes críticas** (Poppins)
- **Fuentes Google Fonts diferidas** (media="print" + onload)
- **Manifest PWA diferido** para no bloquear LCP
- **fetchpriority="high"** en imagen LCP
- **Defer en JavaScript** para no bloquear render
- **Lazy loading** de imágenes no críticas
- **Preconnect** a dominios externos (Google Fonts)
- **Headers de seguridad** (HSTS, CSP, X-Frame-Options, etc.)
- **Request ID** para tracking
- **Logging estructurado** con zerolog
- **Middlewares eficientes**
- **Shutdown gracioso** del servidor
- **Rate limiting** configurado por endpoint

### 🔧 SEO y Accesibilidad
- **Meta tags optimizados** (OG, Twitter Cards)
- **Schema.org JSON-LD** para productos
- **Sitemap.xml** generado dinámicamente
- **Robots.txt** configurado
- **URLs amigables** (slugs)
- **Canonical URLs**
- **Open Graph** images
- **ARIA labels** y roles
- **Navegación por teclado** soportada
- **Contraste adecuado** de colores
- **Textos alternativos** en imágenes

### 🐳 DevOps y Deployment
- **Docker Compose** parametrizado
- **Variables de entorno** centralizadas
- **Sin secretos en código** (todo en .env)
- **Auto-migraciones** de base de datos
- **Seed data** opcional
- **Puerto automático alternativo** si el predeterminado está ocupado
- **Makefile** con comandos útiles
- **Hot reload** en desarrollo

### 🗄️ Base de Datos
- **PostgreSQL** como base de datos principal
- **GORM** como ORM
- **Migrations automáticas**
- **Modelos relacionales** (Products, Variants, Images, Orders, etc.)
- **Backfill** automático de slugs
- **Índices** optimizados
- **Soft deletes** opcionales

### 📚 API RESTful
- **Endpoints REST** completos
- **Autenticación por Bearer token**
- **Documentación implícita** en código
- **Manejo de errores** estandarizado
- **Rate limiting** diferenciado por endpoint
- **Validación de inputs**
- **Respuestas JSON** consistentes

### 🎨 Personalización
- **Múltiples variantes** por producto
- **Colores personalizables**
- **Descripciones cortas y largas**
- **Categorías configurables**
- **Costo base y ajustes** por variante
- **Dimensiones personalizables**
- **Estado de disponibilidad** (Ready to ship)

## Arquitectura y Estructura

### Stack Tecnológico
- **Go 1.22+** - Lenguaje principal
- **PostgreSQL** - Base de datos
- **GORM** - ORM
- **html/template** - Templates SSR
- **MercadoPago SDK** - Pagos
- **Zerolog** - Logging
- **OAuth2** - Autenticación social
- **Docker** - Containerización

### Arquitectura Hexagonal
```
/cmd/tienda3d          # Punto de entrada principal
/internal/app          # Wiring, inicialización, migraciones
/internal/domain       # Entidades + puertos (reglas de negocio)
/internal/usecase      # Casos de uso (lógica de aplicación)
/internal/adapters     # Adaptadores (http, repos, payments, storage)
  /httpserver          # Servidor HTTP + rutas + handlers
  /repo/postgres       # Implementación repositorios
  /payments/mercadopago # Gateway de pagos
  /storage/local       # Storage de archivos
  /pricing/simple      # Cálculo de precios
/internal/views        # Templates SSR
/public/assets         # Archivos estáticos (CSS, JS, imágenes)
/uploads               # Archivos subidos (dinámicos)
/tools                 # Herramientas auxiliares
```

### Principios de Diseño
- **Separación de responsabilidades** (Clean Architecture)
- **Dependency Inversion** (puertos y adaptadores)
- **SOLID principles** aplicados
- **DRY** (Don't Repeat Yourself)
- **Testabilidad** (interfaces bien definidas)
- **Escalabilidad** horizontal

## Autenticación Admin (JWT)
Requiere definir:
- `ADMIN_API_KEY`: clave secreta que se envía en el login.
- `ADMIN_ALLOWED_EMAILS`: lista separada por coma de emails autorizados (si se envía email en login se valida pertenezca).
- `JWT_ADMIN_SECRET` (opcional; si falta se usa `SECRET_KEY`).

Flujo:
1. `POST /admin/login` con header `X-Admin-Key: <ADMIN_API_KEY>` y opcional body `{ "email": "permitido@example.com" }`.
2. Respuesta contiene `{ token, exp, email }`.
3. Usar `Authorization: Bearer <token>` en llamadas a endpoints protegidos.
4. Expiración típica: 30 minutos (renovar con nuevo login).

Endpoints protegidos actuales: creación/listado/detalle/borrado de productos, upload multipart, borrado masivo y listado de órdenes `/admin/orders`.

## Variables de entorno (completas)
Obligatorias mínimas para entorno local:
- `DB_DSN` cadena Postgres (si usás docker compose se arma con variables POSTGRES_*)
- `SESSION_KEY` clave aleatoria segura (firmar cookies)
- `MP_ACCESS_TOKEN` token MercadoPago (TEST-... o prod)
- `ADMIN_API_KEY` (para login admin)
- `ADMIN_ALLOWED_EMAILS` (al menos un email, p.ej. tu correo)

Recomendadas / adicionales:
- `PORT` (default 8080)
- `APP_ENV` (`dev` | `production`) controla selección de token MP y logs
- `PROD_ACCESS_TOKEN` token MP producción (si querés separar del TEST)
- `PUBLIC_BASE_URL` URL pública (para back_urls y webhooks de MP)
- `BASE_URL` (usado también para OAuth Google)
- `SECRET_KEY` firma del external_reference MP (fallback "dev")
- `JWT_ADMIN_SECRET` secreto dedicado para firmar JWT admin (si no, usa `SECRET_KEY`)
- `STORAGE_DIR` carpeta para archivos subidos (default `uploads`)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `ORDER_NOTIFY_EMAIL` (notificación email)
- `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` o `TELEGRAM_CHAT_IDS` (notificación Telegram). `TELEGRAM_CHAT_IDS` permite múltiples destinos separados por coma, p. ej.: `-1001234567890,@SoyCanalla`.
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` (OAuth Google)
- `WHATSAPP_VERIFY_TOKEN`, `WHATSAPP_ACCESS_TOKEN`, `WHATSAPP_PHONE_NUMBER_ID` (WhatsApp Business API)

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
ADMIN_API_KEY=super_admin_key
ADMIN_ALLOWED_EMAILS=tuemail@example.com
JWT_ADMIN_SECRET=jwt_admin_secret
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
ORDER_NOTIFY_EMAIL=
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
WHATSAPP_VERIFY_TOKEN=
WHATSAPP_ACCESS_TOKEN=
WHATSAPP_PHONE_NUMBER_ID=
```

## Docker Compose
Ahora parametrizado (ver `docker-compose.yml`):
```
docker compose up -d --build
```
Asegurate de crear `.env` a partir de `.env.example` antes de levantar.

## Flujos principales
### 1. Carga de productos
Opción recomendada: `POST /api/products/upload` (multipart) con uno o más campos `image` / `images`. El backend guarda cada archivo en `uploads/images/<timestamp>-<filename>` y registra las rutas. Requiere Bearer token admin.

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
- `DELETE /api/products/{slug}` elimina DB + archivos (Bearer admin).
- `POST /api/products/delete` borrado masivo simple (Bearer admin) (no borra archivos físicos).

### 6. Órdenes (Admin)
- `GET /admin/orders` listado paginado de órdenes (Bearer admin). Útil para ver estado después de webhooks.

## Endpoints Principales

### 🌐 Páginas Web (SSR)
- `GET /` - Página de inicio con hero carousel
- `GET /products` - Listado de productos con filtros
- `GET /product/{slug}` - Detalle de producto con carrusel
- `GET /cart` - Carrito de compras
- `GET /cart/update` - Actualizar carrito
- `GET /cart/remove` - Eliminar del carrito
- `GET /cart/checkout` - Iniciar checkout
- `GET /checkout` - Formulario de checkout
- `GET /pay/{orderID}` - Confirmación de pago
- `GET /quote/{id}` - Vista de cotización
- `GET /robots.txt` - SEO robots
- `GET /sitemap.xml` - SEO sitemap

### 🔐 Autenticación
- `POST /admin/login` - Login admin (requiere X-Admin-Key)
- `GET /admin/auth` - Vista de autenticación
- `GET /admin/logout` - Cerrar sesión admin
- `GET /auth/google/login` - OAuth Google (opcional)
- `GET /auth/google/callback` - Callback OAuth
- `GET /logout` - Cerrar sesión usuario

### 👨‍💼 Panel Administrativo
- `GET /admin/orders` - Listado de órdenes (paginado)
- `GET /admin/products` - Gestión de productos
- `GET /admin/sales` - Vista de ventas
- `GET /admin/costs` - Calculadora de costos
- `POST /admin/costs/calculate` - Calcular costos
- `GET /admin/repair_images` - Reparar imágenes huérfanas (con ?dry=1)
- `GET /admin/product_images` - Gestor de imágenes
- `DELETE /admin/product_images/delete` - Eliminar imagen

### 📦 API de Productos (Requerir Bearer token)
- `POST /api/products` - Crear producto
- `GET /api/products` - Listar productos
- `GET /api/products/{slug}` - Obtener producto por slug
- `DELETE /api/products/{slug}` - Eliminar producto (DB + archivos)
- `POST /api/products/delete` - Borrado masivo
- `POST /api/products/upload` - Upload multipart (producto + imágenes)
- `GET /api/product_images/{id}` - Obtener imagen por ID
- `DELETE /api/product_images/{id}` - Eliminar imagen por ID

### 🛒 API de Carrito y Cotización
- `POST /api/quote` - Crear cotización
- `POST /api/checkout` - Crear orden y generar preferencia

### 💳 API de Pagos
- `POST /webhooks/mp` - Webhook MercadoPago (público)

### 📱 API de WhatsApp Business (Requerir Bearer token)
- `GET /api/whatsapp/webhook` - Verificación de webhook (GET)
- `POST /api/whatsapp/webhook` - Webhook de WhatsApp (POST)
- `POST /api/whatsapp/sync-products` - Sincronizar productos
- `GET /api/whatsapp/orders` - Listar órdenes de WhatsApp
- `POST /api/whatsapp/orders` - Crear orden desde WhatsApp

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
1. Login admin (`POST /admin/login`).
2. Subir un producto (ver `carga.md`).
3. Ver en `/products`.
4. Agregar al carrito, cambiar cantidades.
5. Simular pago (token TEST) -> MP sandbox -> volver a `/pay/{id}`.
6. Ver `/admin/orders` para confirmar estado y logs para webhook.

## 🛠️ Herramientas y Utilidades

### Herramientas de sincronización WhatsApp
```bash
cd tools
go run whatsapp_sync.go export-products    # Exportar productos
go run whatsapp_sync.go sync-product slug product_id  # Sincronizar producto
go run whatsapp_sync.go list-products      # Listar productos
```

### Comandos Make disponibles
```bash
make dev      # Correr en modo desarrollo
make build    # Compilar aplicación
make clean    # Limpiar archivos generados
```

## 📋 Notas Importantes

### Configuración
- No se crean productos demo automáticamente
- Las imágenes dinámicas viven fuera de `/public` en `uploads/`
- El slug se recalcula automáticamente si falta (backfill en migración)
- Color picker: si no hay variantes, se muestran colores por defecto
- Para producción: configurar `APP_ENV=production`, `PROD_ACCESS_TOKEN`, dominio HTTPS en `PUBLIC_BASE_URL`
- JWT admin expira (≈30 min); re-logear para nuevo token

### Costos de envío
- Todos los costos provinciales actualmente están fijados en 9000 en `provinceCosts` dentro de `server.go`
- Para modificar, editar el código y recompilar

### Rate Limiting
- Endpoints públicos: 60 requests/minuto general
- `/api/quote`: 15/minuto
- `/api/checkout`: 10/minuto
- `/webhooks/mp`: 30/minuto

## 📚 Documentación Adicional
- **`carga.md`** - Detalles de carga de productos y ejemplos curl
- **`README_WHATSAPP.md`** - Guía completa de integración WhatsApp
- **`.copilot-notes.tienda3d.md`** - Notas técnicas internas

## 🤝 Contribuciones
El proyecto está en desarrollo activo. Las contribuciones son bienvenidas.

## 📄 Licencia
Ver archivo `LICENSE` para más detalles.

---
**Desarrollado con ❤️ para Chroma3D**

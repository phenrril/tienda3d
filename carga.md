# Carga de productos y mantenimiento

Este documento resume cómo crear, listar y eliminar productos con las últimas features.

## 1. Preparar entorno
Arrancar servidor:
```
go run ./cmd/tienda3d
# o
docker compose up -d --build
```
Asegurate de tener `.env` basado en `.env.example`.

Variables mínimas: `DB_DSN` (o POSTGRES_*) , `MP_ACCESS_TOKEN` (TEST-...), `SESSION_KEY`.
Opcionales: `PUBLIC_BASE_URL`, `APP_ENV`, `PROD_ACCESS_TOKEN`, `STORAGE_DIR`, `GOOGLE_CLIENT_ID/SECRET`, `TELEGRAM_BOT_TOKEN/CHAT_ID`, SMTP_*.

## 2. Carga de imágenes / productos
Recomendado: endpoint multipart único.

### 2.1 Endpoint multipart (crear producto + imágenes)
`POST /api/products/upload`
Content-Type: multipart/form-data
Campos:
- `name` (req)
- `base_price` (req, número)
- `category` (opt)
- `short_desc` (opt)
- `ready_to_ship` (`true|false|1|0`)
- `image` (1 archivo) y/o `images` (múltiples). Puedes incluir varias filas `images`.

Ejemplo Postman (form-data):
- name: Maceta Hexa
- base_price: 3500
- category: jardin
- short_desc: Maceta decorativa impresa en 3D
- ready_to_ship: true
- images: img1.webp
- images: img2.webp

Curl:
```
curl -X POST http://localhost:8080/api/products/upload \
  -F name="Maceta Hexa" \
  -F base_price=3500 \
  -F category=jardin \
  -F short_desc="Maceta decorativa" \
  -F ready_to_ship=true \
  -F images=@img1.webp \
  -F images=@img2.webp
```
Respuesta 201 incluye `Images` con URLs `/uploads/images/...`.

### 2.2 JSON simple (sin subir archivos)
`POST /api/products`
Body JSON:
```
{
  "name": "Maceta Hexa",
  "category": "jardin",
  "short_desc": "Maceta decorativa impresa en 3D",
  "base_price": 3500,
  "ready_to_ship": true
}
```
(No genera imágenes.)

## 3. Listar productos
`GET /api/products`
Devuelve `items` y `total`.

## 4. Obtener detalle por slug
`GET /api/products/{slug}`

## 5. Eliminación
### 5.1 Borrado completo (con archivos)
`DELETE /api/products/{slug}` elimina:
- Producto, variantes, imágenes (DB)
- Archivos físicos `uploads/...` asociados (si existen)
Respuesta:
```
{
  "status": "ok",
  "slug": "maceta-hexa",
  "removed_files": ["uploads/images/1696...-maceta-hexa.webp"]
}
```
Errores: 404 inexistente, 500 error interno.

### 5.2 Borrado masivo simple
`POST /api/products/delete`
Body:
```
{ "slugs": ["maceta-hexa", "otro-prod"] }
```
Devuelve arrays `deleted` y `errors`. (No borra archivos físicos).

## 6. Carrusel y miniaturas (frontend)
- Si el producto tiene >1 imagen se activa carrusel (autoplay 5s, flechas, thumbnails, swipe, teclado).
- Primera imagen se usa de portada en listados.

## 7. Colores
- Los colores se deducen de las variantes (`Product.Variants`).
- Si no hay variantes se muestra set por defecto (`#111827`, `#6366f1`, `#16a34a`, `#f59e0b`, `#ff3d00`)

## 8. Carrito y envío
- Agrega productos desde `/product/{slug}` (envía slug + color).
- Todos los costos provinciales actuales están fijos en 9000 en `server.go (provinceCosts)`.
- Retiro: costo 0.

## 9. Pagos
- `POST /api/checkout` genera preferencia MP desde una cotización (flujo futuro de modelos subidos).
- Carrito normal: checkout crea orden y llama a MercadoPago; redirección a `/pay/{id}`.
- Webhook: `/webhooks/mp` (configurar en MP apuntando a `PUBLIC_BASE_URL/webhooks/mp`).

## 10. Notificaciones
- Telegram: requiere `TELEGRAM_BOT_TOKEN` y `TELEGRAM_CHAT_ID`.
- Email (SMTP): requiere SMTP_* variables y `ORDER_NOTIFY_EMAIL`.
- Se envían cuando el pago se marca aprobado (MP webhook o callback simulado).

## 11. OAuth Google (opcional)
Configurar `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `BASE_URL` para habilitar login. Se generan rutas `/auth/google/login` y callback.

## 12. Rutas claves
Web: `/`, `/products`, `/product/{slug}`, `/cart`, `/checkout`, `/pay/{id}`
API: upload, products CRUD, delete masivo, quote, checkout, webhooks.

## 13. Errores comunes
| Código | Causa | Solución |
|--------|-------|----------|
| 400 | Campos faltantes / multipart | Revisar nombres y tipos en form-data |
| 404 | Producto no encontrado | Verificar slug |
| 405 | Método incorrecto | Usar método soportado |
| 500 | Error interno DB / storage | Revisar logs |

## 14. Buenas prácticas
- No subir `.env` (usar `.env.example`).
- Regenerar `SESSION_KEY` en producción.
- Usar token TEST de MP en desarrollo (empieza con `TEST-`).
- Ajustar costos de envío modificando `provinceCosts` y recompilando.

## 15. Ejemplos rápidos
Carga multipart:
```
curl -X POST http://localhost:8080/api/products/upload \
  -F name="Clip Bolsa" \
  -F base_price=600 \
  -F ready_to_ship=true \
  -F image=@clip.jpg
```
Delete completo:
```
curl -X DELETE http://localhost:8080/api/products/clip-bolsa
```
Borrado masivo:
```
curl -X POST http://localhost:8080/api/products/delete \
  -H "Content-Type: application/json" \
  -d '{"slugs":["clip-bolsa","otro-prod"]}'
```

---
Fin.

# 📱 Integración WhatsApp Business con Chroma3D

Esta guía te explica cómo vincular el catálogo de WhatsApp Business con tu tienda Chroma3D para recibir órdenes automáticamente.

## 🚀 Configuración Inicial

### 1. Configurar WhatsApp Business

1. **Crear una cuenta de WhatsApp Business**:
   - Descarga WhatsApp Business en tu teléfono
   - Configura tu perfil de empresa
   - Verifica tu número de teléfono

2. **Configurar WhatsApp Business API**:
   - Ve a [WhatsApp Business Platform](https://business.whatsapp.com/)
   - Crea una cuenta de desarrollador
   - Configura tu webhook URL: `https://tu-dominio.com/api/whatsapp/webhook`

### 2. Variables de Entorno

Agrega estas variables a tu archivo `.env`:

```bash
# WhatsApp Business API
WHATSAPP_VERIFY_TOKEN=tu_token_de_verificacion
WHATSAPP_ACCESS_TOKEN=tu_access_token
WHATSAPP_PHONE_NUMBER_ID=tu_phone_number_id

# Base URL de tu tienda
BASE_URL=https://www.chroma3d.com.ar
```

## 📦 Sincronización de Productos

### Método 1: Exportación Automática (Recomendado)

1. **Exportar productos a WhatsApp**:
```bash
cd tools
go run whatsapp_sync.go export-products
```

2. **Importar en WhatsApp Business**:
   - El comando genera un archivo `whatsapp_products_YYYYMMDD_HHMMSS.json`
   - Abre WhatsApp Business Manager
   - Ve a Catálogo > Productos
   - Importa el archivo JSON generado

3. **Vincular productos**:
```bash
# Para cada producto sincronizado
go run whatsapp_sync.go sync-product nombre-del-producto whatsapp_product_id
```

### Método 2: Sincronización Manual

1. **Listar productos disponibles**:
```bash
go run whatsapp_sync.go list-products
```

2. **Agregar productos manualmente en WhatsApp**:
   - Usa la información mostrada para crear productos en WhatsApp Business
   - Anota el ID de WhatsApp de cada producto

3. **Sincronizar**:
```bash
go run whatsapp_sync.go sync-product slug-del-producto whatsapp-id
```

## 🔄 Flujo de Órdenes

### 1. Cliente hace pedido en WhatsApp

- El cliente navega por tu catálogo en WhatsApp
- Agrega productos al carrito
- Envía el pedido

### 2. Sistema procesa automáticamente

- WhatsApp envía un webhook a tu servidor
- El sistema crea una orden en Chroma3D
- Se genera un link de pago con MercadoPago
- Se envía confirmación al cliente

### 3. Seguimiento

- El cliente recibe notificaciones del estado del pedido
- Puedes gestionar las órdenes desde el panel de administración

## 🛠️ Endpoints API

### Webhook de WhatsApp
```
POST /api/whatsapp/webhook
```
Recibe órdenes automáticamente desde WhatsApp Business.

### Sincronización de Productos
```
POST /api/whatsapp/sync-products
```
Sincroniza un producto específico con WhatsApp.

### Gestión de Órdenes
```
GET /api/whatsapp/orders
```
Lista órdenes pendientes de WhatsApp.

```
POST /api/whatsapp/orders
```
Crea una nueva orden desde WhatsApp manualmente.

## 📊 Monitoreo y Gestión

### Panel de Administración

1. **Ver órdenes de WhatsApp**:
   - Accede a `/admin/orders`
   - Filtra por método de envío "whatsapp"

2. **Gestionar productos sincronizados**:
   - Usa la herramienta de sincronización
   - Verifica el estado de sincronización

### Logs y Debugging

- Los logs de WhatsApp se guardan en el sistema de logging
- Revisa los errores de webhook en los logs del servidor
- Usa el endpoint de órdenes para verificar el estado

## 🔧 Solución de Problemas

### Webhook no recibe órdenes

1. **Verificar configuración**:
   - Confirma que la URL del webhook sea correcta
   - Verifica que el token de verificación coincida

2. **Probar webhook**:
   - Usa herramientas como ngrok para testing local
   - Verifica que el servidor esté accesible desde internet

### Productos no sincronizados

1. **Verificar IDs**:
   - Confirma que el slug del producto sea correcto
   - Verifica que el ID de WhatsApp sea válido

2. **Re-sincronizar**:
   - Usa el comando de sincronización nuevamente
   - Verifica en la base de datos el estado de sincronización

### Órdenes no se procesan

1. **Verificar logs**:
   - Revisa los logs del servidor para errores
   - Confirma que el webhook esté recibiendo datos

2. **Procesar manualmente**:
   - Usa el endpoint de órdenes para procesar manualmente
   - Verifica que los productos existan en Chroma3D

## 📈 Optimizaciones

### Mejores Prácticas

1. **Mantener catálogo actualizado**:
   - Sincroniza productos regularmente
   - Actualiza precios y disponibilidad

2. **Monitorear rendimiento**:
   - Revisa métricas de conversión
   - Optimiza el flujo de pedidos

3. **Respuesta rápida**:
   - Configura notificaciones automáticas
   - Responde rápidamente a los clientes

### Escalabilidad

- El sistema está diseñado para manejar múltiples órdenes simultáneas
- Los webhooks son procesados de forma asíncrona
- La sincronización de productos es eficiente y no bloquea el sistema

## 🆘 Soporte

Si tienes problemas con la integración:

1. Revisa los logs del servidor
2. Verifica la configuración de variables de entorno
3. Confirma que WhatsApp Business esté configurado correctamente
4. Usa las herramientas de debugging incluidas

¡Tu tienda Chroma3D ahora está lista para recibir órdenes directamente desde WhatsApp! 🎉

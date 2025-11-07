# Configuración de Email para Confirmaciones de Compra

Este documento explica cómo configurar el envío automático de emails de confirmación a los compradores cuando realizan una compra.

## Variables de Entorno Necesarias

Agregá estas variables a tu archivo `.env`:

```env
# Email (Gmail SMTP para confirmaciones de compra)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=tu-email@gmail.com
SMTP_PASS=pjtg icvd ples ztiw
SMTP_FROM=tu-email@gmail.com
```

### Descripción de las Variables

- **SMTP_HOST**: Servidor SMTP de Gmail (`smtp.gmail.com`)
- **SMTP_PORT**: Puerto SMTP (587 para TLS)
- **SMTP_USER**: Tu dirección de Gmail
- **SMTP_PASS**: Contraseña de aplicación de Gmail (ver abajo)
- **SMTP_FROM**: Email que aparecerá como remitente (normalmente el mismo que SMTP_USER)

## Cómo Obtener la Contraseña de Aplicación de Gmail

1. **Ir a tu Cuenta de Google → Seguridad**
   - URL directa: https://myaccount.google.com/security

2. **Activar Verificación en 2 pasos**
   - Si no la tenés activada, hacelo primero
   - Es un requisito para generar contraseñas de aplicación

3. **Ir a Contraseñas de Aplicaciones**
   - URL directa: https://myaccount.google.com/apppasswords
   - O desde Seguridad → Verificación en 2 pasos → scroll abajo → "Contraseñas de aplicaciones"

4. **Generar Nueva Contraseña**
   - Seleccioná "Correo" o "Otra (nombre personalizado)"
   - Nombre sugerido: "Tienda3D" o "SMTP Tienda"
   - Click en "Generar"

5. **Copiar la Contraseña**
   - Gmail te mostrará una contraseña de **16 caracteres** (con espacios)
   - Ejemplo: `pjtg icvd ples ztiw`
   - **IMPORTANTE**: Copiala y guardala en un lugar seguro
   - Esta contraseña **NO se podrá ver de nuevo**

6. **Agregar al .env**
   ```env
   SMTP_PASS=pjtg icvd ples ztiw
   ```

## Límites de Gmail

- **500 emails por día** con cuenta gratuita
- **100 destinatarios por mensaje**
- Suficiente para tiendas pequeñas/medianas

## Funcionamiento

### ¿Cuándo se Envían los Emails?

Se envía un email de confirmación al comprador en los siguientes casos:

1. **Compra con Efectivo**: Email inmediato con estado "PENDIENTE"
2. **Compra con Transferencia**: Email inmediato con estado "PENDIENTE"
3. **Compra con MercadoPago**: Email cuando el pago es aprobado

### Contenido del Email

El email incluye:

- ✅ Confirmación de pedido
- 📦 Número de orden
- 🛍️ Detalle de productos (nombre, cantidad, precio)
- 💰 Total y descuentos aplicados
- 🚚 Información de envío (si aplica)
- 💳 Método de pago
- 📞 Próximos pasos

### Template del Email

El email está diseñado con:
- HTML responsive (se ve bien en móviles)
- Colores modernos (gradiente violeta)
- Formato profesional
- Información clara y organizada

## Testing

Para probar el envío de emails sin configurar Gmail:

```bash
# Sin configurar SMTP, el sistema solo logueará un warning
# pero no fallará la compra
unset SMTP_HOST
unset SMTP_USER
unset SMTP_PASS
```

Con SMTP configurado, cada compra generará:
- Un log en consola: `📧 Email de confirmación enviado a [email] para orden [id]`
- Un email al comprador con la confirmación

## Troubleshooting

### Error: "authentication failed"
- Verificá que la contraseña de aplicación esté correcta
- Asegurate de tener verificación en 2 pasos activada

### Error: "connection refused"
- Verificá que el puerto sea 587
- Verificá que SMTP_HOST sea `smtp.gmail.com`

### No se envían emails
- Revisá los logs del servidor
- Verificá que las variables de entorno estén cargadas
- Probá hacer una compra y revisá la consola

### Límite de envío alcanzado
- Gmail tiene límite de 500 emails/día
- Considerá usar un servicio profesional como SendGrid para mayor volumen

## Alternativas a Gmail

Si necesitás mayor capacidad o funcionalidades avanzadas:

### SendGrid (Recomendado para producción)
- Plan gratuito: 100 emails/día
- Planes pagos desde USD 15/mes (40,000 emails)
- Mejor deliverability
- Analytics detallados

### Mailgun
- Plan gratuito: 5,000 emails/mes
- Planes pagos desde USD 35/mes

### AWS SES
- Muy económico: USD 0.10 por 1,000 emails
- Requiere configuración más técnica

## Seguridad

⚠️ **IMPORTANTE**:

- **NUNCA** commitees el archivo `.env` a git
- Mantené tu contraseña de aplicación segura
- No compartas las credenciales SMTP
- Si necesitás regenerar la contraseña, eliminá la anterior en Google

## Soporte

Si tenés problemas con la configuración:

1. Verificá los logs del servidor
2. Revisá que todas las variables estén en el `.env`
3. Probá con una compra de prueba
4. Revisá la carpeta de spam del comprador


# 🔧 Fix: Instrucciones de Transferencia en Pago Efectivo

## 🐛 Problema Reportado

Cuando un usuario seleccionaba **"Pago en efectivo"**, en la página de confirmación aparecían **dos bloques de instrucciones**:

1. ✅ Banner amarillo con mensaje de "Pendiente de pago" y botón de WhatsApp (correcto)
2. ❌ Instrucciones de transferencia bancaria con alias "chroma3d" (incorrecto)

Esto era **muy confuso** para el usuario, porque recibía instrucciones de transferencia cuando había elegido pagar en efectivo.

## 🎯 Solución

Se modificó la condición que determina cuándo mostrar las instrucciones de transferencia para que **solo** se muestren cuando el método de pago es específicamente **"transferencia"**.

### Cambio en el Backend

**Archivo**: `internal/adapters/httpserver/server.go`

**ANTES** (línea 1446):
```go
isTransferenciaPending := o.PaymentMethod == "transferencia" || o.MPStatus == "transferencia_pending" || (status == "pending" && o.MPStatus != "approved")
```

**Problema**: La condición `(status == "pending" && o.MPStatus != "approved")` incluía pagos en efectivo pendientes.

**DESPUÉS** (líneas 1446-1447):
```go
// Solo mostrar instrucciones de transferencia cuando el método de pago es transferencia
isTransferenciaPending := (o.PaymentMethod == "transferencia" || o.MPStatus == "transferencia_pending") && !success
```

**Solución**: Ahora solo se evalúa si el método de pago es "transferencia" Y el pago no fue exitoso.

## 📋 Comportamiento Ahora

### Pago en Efectivo (Pendiente)

✅ **Se muestra**:
- Banner amarillo con ⚠️ "Pendiente de pago"
- Mensaje: "Comunicate con nosotros por WhatsApp para coordinar el pago en efectivo"
- Botón de WhatsApp para contactar
- Datos de la transacción (Orden ID, nombre, email, teléfono, DNI, método, total, items)

❌ **NO se muestra**:
- Instrucciones de transferencia
- Alias bancario
- Mensaje de "Envía el comprobante"

### Pago por Transferencia (Pendiente)

✅ **Se muestra**:
- Banner de estado según corresponda
- **Instrucciones para Transferencia**:
  - Alias: "chroma3d"
  - Monto a transferir
  - Botón para enviar comprobante por WhatsApp
- Datos de la transacción

### Pago por Mercado Pago (Aprobado)

✅ **Se muestra**:
- Banner verde: "Pago aprobado. Gracias por tu compra."
- Datos de la transacción

❌ **NO se muestra**:
- Instrucciones de transferencia
- Mensajes de pago pendiente

## 🧪 Testing

### Caso 1: Efectivo
1. Seleccionar producto
2. Agregar al carrito
3. Ir a checkout
4. Completar datos
5. Seleccionar "Pago en efectivo"
6. Confirmar

**Resultado esperado**:
- Banner amarillo con WhatsApp
- Solo datos de transacción
- **Sin instrucciones de transferencia**

### Caso 2: Transferencia
1. Seleccionar producto
2. Agregar al carrito
3. Ir a checkout
4. Completar datos
5. Seleccionar "Transferencia bancaria"
6. Confirmar

**Resultado esperado**:
- Instrucciones de transferencia completas
- Alias bancario visible
- Botón para enviar comprobante

### Caso 3: Mercado Pago
1. Seleccionar producto
2. Agregar al carrito
3. Ir a checkout
4. Completar datos
5. Seleccionar "Mercado Pago"
6. Pagar

**Resultado esperado**:
- Banner verde de aprobación
- Solo datos de transacción
- **Sin instrucciones de transferencia**

## 📊 Lógica de la Variable

```go
isTransferenciaPending := (o.PaymentMethod == "transferencia" || o.MPStatus == "transferencia_pending") && !success
```

| Condición | Resultado |
|-----------|-----------|
| PaymentMethod = "transferencia" AND !success | ✅ true → Mostrar instrucciones |
| MPStatus = "transferencia_pending" AND !success | ✅ true → Mostrar instrucciones |
| PaymentMethod = "efectivo" AND !success | ❌ false → NO mostrar |
| PaymentMethod = "mercadopago" AND success | ❌ false → NO mostrar |

## 📝 Template Relacionado

**Archivo**: `internal/views/pay.html`

**Líneas 6-19**: Banner amarillo para efectivo pendiente
```html
{{if and (not .Success) (eq .Order.PaymentMethod "efectivo")}}
  <div style="...">
    ⚠️ Pendiente de pago
    Comunicate con nosotros por WhatsApp...
  </div>
{{end}}
```

**Líneas 24-57**: Instrucciones de transferencia
```html
{{if .IsTransferenciaPending}}
  <div style="...">
    📋 Instrucciones para Transferencia
    Alias: chroma3d
    ...
  </div>
{{end}}
```

## ✅ Status

**COMPLETAMENTE RESUELTO** ✓

Ahora cada método de pago muestra únicamente las instrucciones correspondientes, eliminando la confusión del usuario.

---

**Fecha**: Octubre 2025  
**Fix**: Condición isTransferenciaPending  
**Archivos**: `internal/adapters/httpserver/server.go`  
**Líneas**: 1446-1447


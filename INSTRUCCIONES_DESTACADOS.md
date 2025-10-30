# Instrucciones para Solucionar Productos Destacados

## Cambios Realizados

1. **Renombrado de columna**: El campo `order` ahora se llama `display_order` para evitar conflictos con palabra reservada SQL
2. **Mejorado Preload**: Agregado `Preload("Product")` explícito antes de `Preload("Product.Images")`
3. **Logs extensivos**: Agregados logs en frontend y backend para debugging

## Pasos para Solucionar

### 1. Detener la aplicación si está corriendo

### 2. Ejecutar migración de la columna

Conectarse a PostgreSQL y ejecutar:

```sql
-- Verificar si la tabla existe
SELECT * FROM information_schema.tables WHERE table_name = 'featured_products';

-- Si existe, renombrar la columna 'order' a 'display_order'
ALTER TABLE featured_products RENAME COLUMN "order" TO display_order;

-- Si la tabla NO existe, crearla
CREATE TABLE IF NOT EXISTS featured_products (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL UNIQUE,
    display_order INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Crear índice
CREATE INDEX IF NOT EXISTS idx_featured_products_product_id ON featured_products(product_id);
```

### 3. Iniciar la aplicación

```bash
./tienda3d.exe
```

### 4. Verificar en la consola del navegador

Abrir `/admin/destacada` y abrir la consola del navegador (F12). Deberías ver:

```
=== Inicializando Destacada Admin ===
AdminToken: Presente
Elements found: {featuredList: true, allProductsList: true, productSearch: true}
Total productos encontrados: X
Producto 1: ID="uuid-aqui", Name="nombre"
...
```

### 5. Hacer click en "Destacar"

Deberías ver en la consola:

```
Click detectado en allProductsList, target: <button>
Button encontrado: <button>
Button disabled? false
Adding featured product: uuid-aqui
Sending request to add featured product: uuid-aqui
Response status: 201
```

### 6. Verificar logs del servidor

En la terminal donde corre la app deberías ver:

```
INF loaded products for destacada admin products_count=X
INF product sample id=uuid name=nombre
INF loaded featured products featured_count=0
INF === apiFeaturedAdd called === method=POST path=/api/featured/add
INF decoded request - attempting to add featured product product_id=uuid order=0
INF featured product saved successfully featured_id=uuid
```

## Solución de Problemas

### Si no aparece nada en consola del navegador
- Verificar que el JavaScript se está cargando
- Abrir la pestaña "Network" y buscar errores de carga de scripts

### Si el botón no responde
- Verificar en consola si dice "Botón no encontrado o está deshabilitado"
- Verificar que el atributo `data-action="add"` existe en el botón

### Si el UUID está vacío
- Verificar en los logs del servidor que los productos tienen ID
- Pasar el mouse sobre un producto para ver el title con el ID

### Si la API devuelve error 400 "json"
- El UUID no se está parseando correctamente
- Verificar que el UUID en el HTML es válido

### Si la API devuelve error 500 "error saving"
- Problema con la base de datos
- Verificar que la tabla `featured_products` existe
- Verificar que el `product_id` existe en la tabla `products`

## Verificación Manual en Base de Datos

```sql
-- Ver productos destacados
SELECT fp.id, fp.product_id, fp.display_order, fp.active, p.name
FROM featured_products fp
JOIN products p ON fp.product_id = p.id
WHERE fp.active = true
ORDER BY fp.display_order;

-- Agregar manualmente un producto destacado (para testing)
INSERT INTO featured_products (id, product_id, display_order, active, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    (SELECT id FROM products LIMIT 1),
    0,
    true,
    NOW(),
    NOW()
);

-- Limpiar todos los destacados
DELETE FROM featured_products;
```

## Después de Solucionar

Una vez que funcione:
1. Los logs se pueden remover o comentar para producción
2. La página inicial debería mostrar los productos destacados automáticamente
3. Máximo 9 productos pueden estar destacados simultáneamente


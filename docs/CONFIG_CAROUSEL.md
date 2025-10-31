# Configuración del Carrusel de Imágenes

## Descripción

El carrusel de imágenes en la página de inicio puede enlazar a productos específicos y mostrar sus imágenes reales. Cuando un usuario hace clic en una imagen del carrusel, será redirigido al producto correspondiente.

## Funcionalidades

- **Imágenes dinámicas**: Muestra las imágenes reales de los productos configurados
- **Enlaces clickeables**: Cada imagen lleva al producto correcto
- **Navegación automática**: El carrusel cambia cada 3 segundos
- **Sincronización**: Los dots y el swipe respetan el orden de los productos

## Configuración

Los productos del carrusel se configuran a través del archivo `carousel.json` en la raíz del proyecto. El archivo se crea automáticamente cuando usas la interfaz de administración en `/admin/destacada`.

También puedes crearlo manualmente con el siguiente formato:

```json
{
  "items": [
    "slug-del-producto-1",
    "slug-del-producto-2",
    "slug-del-producto-3",
    "slug-del-producto-4",
    "slug-del-producto-5"
  ]
}
```

### Ejemplo

Si tienes productos con los siguientes slugs:
- `llavero-logo`
- `lampara-luna`
- `soporte-celular`

Tu configuración en `carousel.json` sería:

```json
{
  "items": [
    "llavero-logo",
    "lampara-luna",
    "soporte-celular",
    "",
    ""
  ]
}
```

**Nota**: El archivo puede tener menos de 5 items, pero siempre se guardarán 5 posiciones (las vacías se representan como strings vacíos `""`).

## Comportamiento

- **Con producto configurado**: 
  - El carrusel buscará el producto por slug y extraerá su primera imagen
  - La imagen del carrusel será clickeable y redirigirá a `/product/{slug}` cuando se haga clic en ella
  - Se mostrará la imagen real del producto en lugar de las imágenes por defecto
- **Sin producto configurado**: 
  - Si no hay un valor para un ítem o el producto no existe, esa posición del carrusel no se mostrará
  - Si **ningún** producto está configurado, se usarán las imágenes por defecto (`img1.webp` hasta `img4.webp`) sin enlaces

## Notas Importantes

- Los valores deben ser los slugs de los productos
- No incluir la ruta `/product/` en el slug
- Si algún item está vacío (string vacío `""`), esa posición del carrusel será omitida
- Las imágenes mostradas serán las **primeras imágenes** de cada producto configurado
- Si un producto no tiene imágenes, esa posición será omitida del carrusel
- El archivo `carousel.json` se crea automáticamente al usar la interfaz de administración
- Puedes editar el archivo manualmente si lo prefieres, pero asegúrate de mantener el formato JSON válido


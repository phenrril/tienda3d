# Configuración del Carrusel de Imágenes

## Descripción

El carrusel de imágenes en la página de inicio puede enlazar a productos específicos y mostrar sus imágenes reales. Cuando un usuario hace clic en una imagen del carrusel, será redirigido al producto correspondiente.

## Funcionalidades

- **Imágenes dinámicas**: Muestra las imágenes reales de los productos configurados
- **Enlaces clickeables**: Cada imagen lleva al producto correcto
- **Navegación automática**: El carrusel cambia cada 3 segundos
- **Sincronización**: Los dots y el swipe respetan el orden de los productos

## Configuración

Para configurar los productos del carrusel, agrega las siguientes variables de entorno en tu archivo `.env`:

```env
ITEM_1=slug-del-producto-1
ITEM_2=slug-del-producto-2
ITEM_3=slug-del-producto-3
ITEM_4=slug-del-producto-4
ITEM_5=slug-del-producto-5
```

### Ejemplo

Si tienes productos con los siguientes slugs:
- `llavero-logo`
- `lampara-luna`
- `soporte-celular`

Tu configuración sería:

```env
ITEM_1=llavero-logo
ITEM_2=lampara-luna
ITEM_3=soporte-celular
ITEM_4=
ITEM_5=
```

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
- No incluir la ruta `/product/` en la variable
- Si algún ITEM_* está vacío o no se configura, esa posición del carrusel será omitida
- Las imágenes mostradas serán las **primeras imágenes** de cada producto configurado
- Si un producto no tiene imágenes, esa posición será omitida del carrusel


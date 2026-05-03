const CACHE_NAME = 'chroma3d-v6';
const urlsToCache = [
  '/',
  '/public/assets/styles.css',
  '/public/assets/img/chroma3d-isotipo.svg',
  '/public/assets/img/chroma3d-wordmark-horizontal.svg'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(urlsToCache))
  );
  self.skipWaiting();
});

self.addEventListener('fetch', event => {
  const req = event.request;
  const isHTML = req.headers.get('accept') && req.headers.get('accept').includes('text/html');
  if (isHTML) {
    // Para HTML, siempre ir a red (fallback a caché si offline)
    event.respondWith(
      fetch(req).catch(() => caches.match(req))
    );
    return;
  }
  event.respondWith(
    caches.match(req).then(res => res || fetch(req))
  );
});

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(keys => Promise.all(keys.filter(k => k !== CACHE_NAME).map(k => caches.delete(k))))
  );
  self.clients.claim();
});
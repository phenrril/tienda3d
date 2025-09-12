package httpserver

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type key int

const requestIDKey key = 1

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := make([]byte, 8)
		_, _ = rand.Read(id)
		ctx := context.WithValue(r.Context(), requestIDKey, fmtHex(id))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetReqID(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &respWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		log.Info().Str("id", GetReqID(r.Context())).Str("method", r.Method).Str("path", r.URL.Path).Int("status", rw.status).Dur("dur", time.Since(start)).Msg("req")
	})
}

type respWriter struct {
	http.ResponseWriter
	status int
}

func (rw *respWriter) WriteHeader(code int) { rw.status = code; rw.ResponseWriter.WriteHeader(code) }

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().Interface("err", rec).Str("id", GetReqID(r.Context())).Msg("panic")
				http.Error(w, "internal error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()
		w.Header().Set("Content-Encoding", "gzip")
		gw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) { return w.Writer.Write(b) }

func RateLimit(maxPerMin int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	buckets := map[string]*bucket{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			mu.Lock()
			b, ok := buckets[ip]
			if !ok || time.Since(b.ts) > time.Minute {
				b = &bucket{count: 0, ts: time.Now()}
				buckets[ip] = b
			}
			if b.count >= maxPerMin {
				mu.Unlock()
				http.Error(w, "rate limit", 429)
				return
			}
			b.count++
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

func PublicRateLimit(perPathLimits map[string]int) func(http.Handler) http.Handler {

	var mu sync.Mutex
	buckets := map[string]*bucket{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit, ok := perPathLimits[r.URL.Path]
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			key := r.URL.Path + "|" + ip
			mu.Lock()
			b, ok := buckets[key]
			if !ok || time.Since(b.ts) > time.Minute {
				b = &bucket{count: 0, ts: time.Now()}
				buckets[key] = b
			}
			if b.count >= limit {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				http.Error(w, "rate limit", http.StatusTooManyRequests)
				return
			}
			b.count++
			remaining := limit - b.count
			mu.Unlock()
			w.Header().Set("X-RateLimit-Limit", strconvItoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconvItoa(remaining))
			next.ServeHTTP(w, r)
		})
	}
}

func strconvItoa(i int) string { return fmtInt(i) }

func fmtInt(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	buf := [20]byte{}
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

type bucket struct {
	count int
	ts    time.Time
}

func Chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func fmtHex(b []byte) string {
	const hexd = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexd[v>>4]
		out[i*2+1] = hexd[v&0x0f]
	}
	return string(out)
}

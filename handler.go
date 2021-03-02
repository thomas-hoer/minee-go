package minee

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
func gzipper(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler.ServeHTTP(gzw, r)
	})
}

func handleMiddleware(next http.Handler, logRequests bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var caching bool = true
		if strings.HasSuffix(r.RequestURI, ".css") {
			w.Header().Add("Content-Type", "text/css")
		} else if strings.HasSuffix(r.RequestURI, ".html") {
			w.Header().Add("Content-Type", "text/html")
		} else if strings.HasSuffix(r.RequestURI, "/static/") {
			w.Header().Add("Content-Type", "text/html")
		} else if strings.HasSuffix(r.RequestURI, ".ico") {
			w.Header().Add("Content-Type", "image/x-icon")
		} else if strings.HasSuffix(r.RequestURI, ".png") {
			w.Header().Add("Content-Type", "image/png")
		} else if strings.HasSuffix(r.RequestURI, ".jpg") {
			w.Header().Add("Content-Type", "image/jpeg")
		} else if strings.HasSuffix(r.RequestURI, ".js") {
			w.Header().Add("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.RequestURI, "json") { // .json or ?json
			caching = false
			w.Header().Add("Content-Type", "application/json")
		} else if !strings.HasSuffix(r.RequestURI, "/") {
			caching = false
		} else {
			caching = false
		}
		if caching {
			w.Header().Add("Cache-Control", "public, max-age=2592000") //30 days
		}
		next.ServeHTTP(w, r)
		if logRequests {
			log.Printf("%v %v took %v", r.Method, r.RequestURI, time.Since(start))
		}
	})
}

package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Status() int {
	return sr.status
}

func getClientIP(r *http.Request) string {
	clientIP := r.Header.Get("Fly-Client-IP")
	if clientIP == "" {
		var err error
		clientIP, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = "unknown"
		}
	}
	return clientIP
}

func logHit(r *http.Request, status int) {
	proto := r.Header.Get("X-Forwarded-Proto")
	xff := r.Header.Get("X-Forwarded-For")
	httpReqs.WithLabelValues(proto, r.Host).Inc()
	log.Printf("client=\"%s\" proto=\"%s\" host=\"%s\" method=\"%s\" url=\"%s\" code=\"%d\" referrer=\"%s\" ua=\"%s\" xff=\"%s\"",
		getClientIP(r), proto, r.Host, r.Method, r.URL.String(), status, r.Referer(), r.UserAgent(), xff)
}

func filter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only serve localhost / markhale.fly.dev / markhale.org / *.markhale.org.
		// Anything else will be a bot scanning by IP address.
		validHost := strings.Contains(r.Host, "localhost") || strings.Contains(r.Host, "markhale.")
		if !validHost {
			http.Error(w, "Page not found", http.StatusNotFound)
			logHit(r, http.StatusNotFound)
			return
		}

		// Only allow GET and HEAD requests because this is a read-only site.
		validMethod := r.Method == http.MethodGet || r.Method == http.MethodHead
		if !validMethod {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			logHit(r, http.StatusMethodNotAllowed)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func redirect(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect HTTP to HTTPS and www.markhale.org to markhale.org.
		if r.Header.Get("X-Forwarded-Proto") == "http" || r.Host == "www.markhale.org" {
			u := &url.URL{
				Scheme:   "https",
				Host:     "markhale.org",
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
			}
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
			logHit(r, http.StatusMovedPermanently)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func serve(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set STS header to switch browsers to HTTPS.
		// 63072000 = 2 years; can set to 0 to remove.
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Cache-Control", "public, max-age=300")

		// Wrap ResponseWriter to record the status code.
		// Initialize the status to 200 in case WriteHeader is not called.
		sr := statusRecorder{w, http.StatusOK}
		h.ServeHTTP(&sr, r)
		logHit(r, sr.Status())
	})
}

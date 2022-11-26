package main

import (
	"io"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, indexHTML)
}

func styleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, styleCSS)
}

// For mta-sts.txt format and values, see:
// https://datatracker.ietf.org/doc/html/rfc8461#section-3.2 and
// https://www.rfc-editor.org/errata_search.php?rfc=8461&rec_status=0
func mtaSTSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Host != "mta-sts.markhale.org" {
		http.Error(w, "Not an MTA-STS subdomain", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, mtaSTSPolicy)
}

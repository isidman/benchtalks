package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"
)

const version = "1.2.0"

// So Go embeds the entire public/ folder into the binary.
func NewRouter(hub *Hub, staticFiles embed.FS, startTime time.Time) http.Handler {
	mux := http.NewServeMux()

	// take off "public" as prefix so /public/index.html is served as
	// /index.html
	// fs.Sub makes the "prefix" go away and it makes things get served from
	// root, like the JS version.
	stripped, err := fs.Sub(staticFiles, ".")
	if err != nil {
		// If the filesystem embedded in, is broken, the server is going to have
		// a tummy ache.
		// So instead of pushing it to try again or start up and serve nothing,
		// it's better to make it crash immediately
		panic("could not strip public/ prefix from embedded files: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(stripped))

	mux.HandleFunc("/health", healthHandler(hub, startTime))

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "status.benchtalks.chat" {
			http.NotFound(w, r)
			return
		}
		data, err := fs.ReadFile(stripped, "status.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	// ws endpoint c:
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(hub, w, r)
	})

	//Serving T&Cs
	mux.HandleFunc("/terms.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./pkg/public/terms.html")
	})

	mux.HandleFunc("/privacy.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./pkg/public/privacy.html")
	})

	// Read index.html directly from the embedded FS and serve it ourselves.
	// We bypass the file server entirely for this one file because Go's file
	// server
	// always redirects /index.html → / which causes an infinite loop.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "status.benchtalks.chat" && (r.URL.Path == "/" || r.URL.Path == "") {
			http.Redirect(w, r, "/status", http.StatusFound)
			return
		}
		if r.URL.Path == "/" || r.URL.Path == "" {
			data, err := fs.ReadFile(stripped, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func healthHandler(hub *Hub, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		payload := struct {
			Status     string `json:"status"`
			Version    string `json:"version"`
			Uptime     int64  `json:"uptime_seconds"`
			Rooms      int    `json:"rooms"`
			Federation bool   `json:"federation"`
		}{
			Status:     "ok",
			Version:    version,
			Uptime:     int64(time.Since(startTime).Seconds()),
			Rooms:      hub.RoomCount(),
			Federation: hub.IsFederated(),
		}
		json.NewEncoder(w).Encode(payload)

	}
}

// Since the go:embed directive needs to be in the same package as the variable
// that holds the files,
// the //go:embed line lives in main.go instead of here.

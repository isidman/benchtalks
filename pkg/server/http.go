package server

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
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
		data, err := fs.ReadFile(stripped, "terms.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/privacy.html", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(stripped, "privacy.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "type/html; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/api/invite", inviteCreateHandler(hub))
	mux.HandleFunc("/api/invite/claim/", inviteClaimHandler(hub))

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

// POST /api/invite is handled by this one.
// Browser sends the room ID and the "mystical" payload
// (the key is encypted client-side). The handler here generates the raw token,
// storing the hash & gives back the raw token to browser.
func inviteCreateHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		//decoding the request body
		var body struct {
			RoomID           string `json:"roomId"`
			EncryptedPayload string `json:"encryptedPayload"`
			Token            string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if body.RoomID == "" || body.EncryptedPayload == "" || body.Token == "" {
			http.Error(w, "room id, payload and token are required", http.StatusBadRequest)
			return
		}

		rawBytes, err := base64.RawURLEncoding.DecodeString(body.Token)
		if err != nil || len(rawBytes) != 32 {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}

		//Hashing it, this is what gets stored, raw token != stored
		hash := sha256.Sum256(rawBytes)
		hashHex := hex.EncodeToString(hash[:])

		//Stored in the hub, expires in 24 hours
		hub.StoreInviteToken(hashHex, body.EncryptedPayload, body.RoomID, time.Now().Add(5*time.Minute))

		//raw token back to browser
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			OK bool `json:"ok"`
		}{OK: true})
	}
}

// GET /join/{token} gets handled by this one.
// grabs the token from the URL path, validates it via the hub,
// returns the payload and roomID as JSON.
// Browser uses the token from URL to decrypt on join.html
// the payload, client-side and redirects to the room.
func inviteClaimHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rawToken := strings.TrimPrefix(r.URL.Path, "/api/invite/claim/")
		if rawToken == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		encryptedPayload, roomID, ok := hub.ClaimInviteToken(rawToken)
		if !ok {
			http.Error(w, "invite link is invalid, expired or already used", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(struct {
			EncryptedPayload string `json:"encryptedPayload"`
			RoomID           string `json:"roomId"`
		}{
			EncryptedPayload: encryptedPayload,
			RoomID:           roomID,
		})
	}
}

// Since the go:embed directive needs to be in the same package as the variable
// that holds the files,
// the //go:embed line lives in main.go instead of here.

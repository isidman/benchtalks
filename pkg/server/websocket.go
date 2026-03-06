package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Ping-pong timer booboo
const (
	writeWait = 10 * time.Second // sending you the message within 10 seconds,
	// otherwise fricc you.
	pongWait = 60 * time.Second // "are you there? in 60s you won't be either
	// way"
	pingPeriod = 45 * time.Second // "BEFORE YOU'RE GONE though... are you even
	// alive?"
)

// Ξαναπαίρνω στο ονειρό μου... μα σιωπή μόνο στο κινητό μου! 🎶

// when you connect to this place via a browser, it's an http request. But if
// you wanna talk we need to change some things, like the http into ws so it's
// persistent.
// CheckOrigin needs to be true in order to accept connections from any domain.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client-Server speakage: LFG
// Structs is the way to go since this is working with RAM. Messages traveling
// over the ws, are JSON and structs give shape with connections.
// Since I'm on the server's side, "incoming" messages are the ones that come
// from browsers and "outgoing" messages are the ones the server sends to the
// clients subscribed to rooms.

type IncomingMessage struct {
	Type      string `json:"type"`
	Payload   string `json:"payload"`
	RoomID    string `json:"roomId"`
	AdminHash string `json:"adminHash"`
}

// Χτυπάς εσύ, μα πονάω και εγώ.. οι δυο καρδιές μας σαν μία.🎶

type OutgoingMessage struct {
	Type     string `json:"type"`
	Payload  string `json:"payload"`
	SenderID string `json:"senderId"`
}

// Φαίνεται μοιάζουμε 'μεις σαν δύο, συγκινονούντα δοχεία! 🎶

// Since I'm on server side, when it receives a new ws connection, there's a
// certain way it has to handle it, kinda like Po and the cannonball from Kung
// Fu Panda.
// So, it changes from http to ws so it's continuous. If that doesn't happen it
// writes it down and goes bye-bye.
// Then a client/seat is created with random id, attaching the ws connection to
// it and giving them mouths and ears.

// Τώρα μου μιλάει, τώρα με φιλάει! Τώρα μου λέει πως τρελά με αγαπάει...
// τώρα με αγκαλιάζει, τα άστρα όλα μου τάζει, την ζωή μου όλη την αλλάζει!
// (ΔΕΝ) Πεθαίνω... (λόγω pingPeriod) 🎶
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	//Extraction of client IP before r disappears.
	//After the WS upgrade, r is gone and this is the only place we have it.
	//X-Forwarded-For comes first since Traefik adds it and without it,
	//r.RemoteAddr would always be Traefik's internal Docker IP, making ban checks useless.
	//Never logged, living on the client struct only.
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		//Direct connection. RemoteAddr is "ip:port", stripping the port.
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	client := &Client{
		id:   newID(),
		conn: conn,
		send: make(chan []byte, 256),
		ip:   ip, //stored once, never modified, never logged
	}

	go client.readPump(hub)
	go client.writePump()
}

// The ears and mouths are the read and write pumps. server launches them as
// soon as the ws connection happens and returns ASAP.
// The go routines, are working continuously in the bg. As soon as the http
// handling is done the go routines "own" the client/seat on the bench.

// Ears
func (c *Client) readPump(hub *Hub) {
	var roomID string

	// the roomID needs to start empty since we don't know where the client
	// belongs to yet. SO when they send a message they tell us in which bench
	// their seat is.
	// But whatever happens we have the defer block guarantee: when readPump
	// exits, it removes the seat from the bench tell the ones that are seating
	// on the other seats that they left and closes the connection.
	// Ότι και αν γίνει στον κόσμο αυτό... Θα μ'αγκαλιάσεις και θα σωθώ.🎶
	defer func() {
		if roomID != "" {
			hub.LeaveRoom(roomID, c)
			notify := buildOutgoing("leave", "", c.id)
			hub.Broadcast(roomID, c.id, notify)
		}

		c.conn.Close()
	}()

	// Ποιος να ακούσει; Τι τον νοιάζει;;
	// Αν εμένα η καρδιά ξεπαγιάζει!
	// Παίρνω φίλους, παιρνώ εσένα... 🎶
	// Read deadline sets a timer for a minute. If nothing arives, readmessage
	// errors out, breaking loop, defer runs, Ε Κ Κ Α Θ Α Ρ Ι Σ Ι Σ
	// If a pong arrives the "conn.setponghandler" resets the timer.
	// PING PONG TO INFINITY.
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.conn.SetReadLimit(512 * 1024 * 1024) // 512MB max message size

	// ΑΝΑΠΑΝΤΗΤΕΣ ΚΛΗΣΕΙΣ ΠΑΝΤΟΥΥΥΥΥΥΥΥΥΥΥΥΥΥΥΥΥΥΥ...🎶
	// Main loopity bloop loop
	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// When it reaches "readmessage" it waits. Then when a message arrives
		// it parses it from json to "incomingmessage".
		// If that doesn't work, the message gets skipped and it waits for the
		// next one.
		var msg IncomingMessage
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			continue
		}

		// The codeword is "join". This is how you make a bench, and the "defer"
		// knows what to clean up afterwards.
		// The seat gets registered in the park (kinda), and then everybody
		// knows that somebody sat to the bench as them.
		switch msg.Type {

		case "join":
			roomID = msg.RoomID
			hub.JoinRoom(roomID, msg.AdminHash, c)
			notify := buildOutgoing("join", "", c.id)
			hub.Broadcast(roomID, c.id, notify)
			log.Printf("JOIN: roomID=%s adminHash=%s", roomID, msg.AdminHash)

			count := hub.RoomSize(roomID)
			welcome := buildOutgoing("welcome", fmt.Sprintf("%d", count), c.id)
			c.send <- welcome

		// In the same way messages get handled, it handles images too. The
		// server doesn't look inside the payloads/blobs.
		case "message":
			if roomID == "" {
				continue
			}
			out := buildOutgoing("message", msg.Payload, c.id)
			hub.Broadcast(roomID, c.id, out)

		case "image":
			if roomID == "" {
				continue
			}
			out := buildOutgoing("image", msg.Payload, c.id)
			hub.Broadcast(roomID, c.id, out)

		// This case  gets the "msg.payload" (admin token) to hub.DeleteRoom
		// that does the "verifyadmintoken" inside hub.go.
		// If the match is exact the "park" calls the tiny giant to pull the
		// client's ears. Otherwise, they get an error.
		case "delete":
			if roomID == "" {
				continue
			}
			deleted := hub.DeleteRoom(roomID, msg.Payload)
			if !deleted {
				errMsg := buildOutgoing("error", "invalid admin token", c.id)
				c.send <- errMsg
			}

		// same as delete case
		case "make_public":
			if roomID == "" {
				continue
			}
			ok := hub.MakePublic(roomID, msg.Payload)
			if !ok {
				errMsg := buildOutgoing("error", "invalid admin token", c.id)
				c.send <- errMsg
			} else {
				// Notify other clients in the room. inform people that this
				// room is federated
				notify := buildOutgoing("made_public", "", c.id)
				hub.Broadcast(roomID, c.id, notify)
				c.send <- buildOutgoing("made_public", "", c.id)
			}

		// Request for pairing token
		case "request_pair":
			// guard: must be in a room
			if roomID == "" {
				continue
			}

			// msg.Payload = admin token
			// msg.AdminHash = claimerBenchID (reusing the field)
			claimerBenchID := msg.AdminHash

			if claimerBenchID == "" {
				errMsg := buildOutgoing("error", "claimer bench ID required", c.id)
				c.send <- errMsg
				continue
			}

			// Hub generates the token - verifies admin, stores hash, returns raw token
			rawToken, err := hub.GeneratePairToken(roomID, msg.Payload, claimerBenchID)
			if err != nil {
				errMsg := buildOutgoing("error", err.Error(), c.id)
				c.send <- errMsg
				continue
			}

			// Build the pairing URL and send it back to the admin's browser.
			// The admin copies this URL and shares it with the bench-B operator.

			// Note: the raw token is being sent here — the browser will put it in
			// the ?pair= query param of the URL it constructs.
			// The claimerBenchID is sent as well, so the browser can put it in &claimer=
			pairPayload := rawToken + "|" + claimerBenchID
			c.send <- buildOutgoing("pair_token", pairPayload, c.id)

		case "claim_pair":
			if roomID == "" {
				continue
			}

			// Relay must exist - claiming only makes sense in a federated setup
			if hub.relay == nil {
				errMsg := buildOutgoing("error", "this bench is not connected to a park", c.id)
				c.send <- errMsg
				continue
			}

			rawToken := msg.Payload
			if rawToken == "" {
				errMsg := buildOutgoing("error", "missing pair token", c.id)
				c.send <- errMsg
				continue
			}

			// Publish the claim to the park via NATS.
			// e.x: bench A will receive it on bench.pair.verify.{roomID},
			// verify it, and publish back apporoved or rejected.
			// hub.relay.BenchID() gives us this bench's own ID for Option B binding.
			hub.relay.PublishPairClaim(roomID, rawToken, hub.relay.BenchID())

			// No need to wait for a response here as it arrives asynchronously via NATS
			// and goes through HandlePairApproved or HandlePairRejected.
			// The client gets notified when that happens via a future WS message.
			c.send <- buildOutgoing("pair_claim_sent", "", c.id)

		case "ban_client":
			//Guard: must be in a room
			if roomID == "" {
				continue
			}

			//The msg.Payload is the senderID of the client to ban.
			//Admin doesn't share or sees an IP - there's server lookup internally.
			targetID := msg.Payload
			if targetID == "" {
				c.send <- buildOutgoing("error", "missing target client ID", c.id)
				continue
			}

			//Prevent admin from banning themselves.
			if targetID == c.id {
				c.send <- buildOutgoing("error", "cannot ban yourself", c.id)
				continue
			}

			//Finding the target client in the room/bench, requires
			//the room lock, so as to safely read the clients map.
			hub.mu.Lock()
			room, exists := hub.rooms[roomID]
			hub.mu.Unlock()

			if !exists {
				continue
			}

			room.mu.Lock()
			target, found := room.clients[targetID]
			room.mu.Unlock()

			if !found {
				c.send <- buildOutgoing("error", "client not found in room", c.id)
				continue
			}

			// Ban IP first, then kick client.
			// The IP is read from the struct, not logged, not sent anywhere.
			hub.BanIP(roomID, target.ip)
			target.send <- buildOutgoing("kicked", "you have been banned from this room", target.id)
			close(target.send)

			// The target is not removed from room.clients here.
			// When close(target.send) is fired, writePump exits,
			// which causes readPump to exit, which runs the defer.
			// The defer calls LeaveRoom, that removes the target from the map.
		}
	}
}

// Mouths
// The select works as a switch for channels. It waits for what is going to
// happen from the two, either for something to arrive in client.send or for the
// ticker to fire.
// Of the two go routines, it's the only one that calls the WriteMessage
// function.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helpers

// Gets the fields below and creates an OutgoingMessage, makes it into JSON
// bytes.
func buildOutgoing(msgType, payload, senderID string) []byte {
	out := OutgoingMessage{
		Type:     msgType,
		Payload:  payload,
		SenderID: senderID,
	}
	data, _ := json.Marshal(out)
	return data
}

// UUID CREATION TIME BABYYYYYYYYYYYYYYYYY
func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

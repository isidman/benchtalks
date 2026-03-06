package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	natspkg "github.com/isidman/benchtalks/pkg/nats"

	"github.com/gorilla/websocket"
)

// This is a user/seat.
// They are assigned a random id, have a connection and some messages.

type Client struct {
	id   string
	conn *websocket.Conn
	send chan []byte
}

// This is a room/bench.
// It exists when a seat is taken.
// Instead of schrondiger's, it could be izziebox (?) marketable.
type Room struct {
	id        string
	adminHash string
	isPublic  bool
	clients   map[string]*Client // who's sitting on the bench
	mu        sync.Mutex         // lock before touch clients (ew wtf?)
}

// This has the benchID allowed to claim + binding bench, 5 min TTL,
// ded after this regardless of used flag + self-destuct flag,
// flipped on first valid claim.
type PairingToken struct {
	RoomID    string
	ClaimerID string
	ExpiresAt time.Time
	Used      bool
}

// This is a hub. It has all rooms that have users in them. Kinda like a very
// noisy and loud park.
type Hub struct {
	rooms map[string]*Room
	mu    sync.Mutex     // lock before touch rooms
	relay *natspkg.Relay // nil for standalone, set at startup via SetRelay

	// key = sha256(rawToken as hex string)
	// Never storing the raw token same principle
	// as adminHash.
	pairingTokens map[string]PairingToken
	pairingMu     sync.Mutex //separate lock - not holding hub.mu while working crypto.

	// The bench needs a bouncer,
	// and the bouncer needs a list.
	// key = roomID, value = set of benchIDs trusted for that bench.
	trustedPeers map[string]map[string]bool
	peersMu      sync.Mutex //another separation for the same reason we don't lock "pairingMu"
}

// this functions is called once, when this whole thing starts, in main.go
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),

		// pairing tokens initialised empty — tokens get added when
		// an admin calls GeneratePairToken and removed (burned) when
		// VerifyPairClaim runs. Starting empty is correct.
		pairingTokens: make(map[string]PairingToken),

		// trusted peers initialised empty — entries get added when
		// RegisterTrustedPeer is called after a successful handshake.
		// An empty map here means "no room has any trusted peers yet"
		// which is exactly the state at startup.
		trustedPeers: make(map[string]map[string]bool),
	}
}

// This one gives the hub a live relay to publish park messages through and its
// called from main.go, right after "connect()" succeeds after NATS_PEERS.
// If NATS is not configured, this is standalone.
func (h *Hub) SetRelay(r *natspkg.Relay) {
	h.relay = r
}

// This one checks if there's a room and if there isn't, it makes a new one.
// Admin has is for new rooms only (HEHE), not for existing ones.
func (h *Hub) getOrCreateRoom(roomID, adminHash string) *Room {
	// locking the hub, because it reads or writes the room map
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.rooms[roomID]; exists {
		return room
	}

	room := &Room{
		id:        roomID,
		adminHash: adminHash,
		clients:   make(map[string]*Client),
	}
	h.rooms[roomID] = room
	return room
}

func (h *Hub) RoomSize(roomID string) int {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return 0
	}

	room.mu.Lock()
	defer room.mu.Unlock()
	return len(room.clients)
}

// This one adds a client to a room if it exists, otherwise creates the room and
// adds the client in
func (h *Hub) JoinRoom(roomID, adminHash string, client *Client) {
	room := h.getOrCreateRoom(roomID, adminHash) // Παίρνω μες το βράδυ,
	// στο σκοτάδι,
	// μα κανένα αγάπης σημάδι. 🎶

	// locking the room, this time, because it writes the client map
	room.mu.Lock()
	defer room.mu.Unlock()

	room.clients[client.id] = client
}

// This one makes client go bye-bye.
// If there's nobody on the bench, the bench gets pulled up by a secret tiny
// giant and eaten.
func (h *Hub) LeaveRoom(roomID string, client *Client) {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.clients, client.id)
	isEmpty := len(room.clients) == 0
	room.mu.Unlock()

	if isEmpty {
		h.mu.Lock()
		delete(h.rooms, roomID)
		h.mu.Unlock()
	}
}

// This one is broadcasting a message.
// Sends it to everyone apart from the sender.
func (h *Hub) Broadcast(roomID string, senderID string, message []byte) {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for id, client := range room.clients {
		if id == senderID {
			continue //no going back to sender
		}

		// This part allows kind-of asynchronous reading of messages on "send"
		// channel of each client.
		// And the rest of the clients don't need to wait for everyone to read
		// the message broadcasted.
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(room.clients, id)
		}
	}

	// Two conditions: public room and live relay. That's all it takes to
	// publish messages to the park.
	// other benches will receive this and forward to their local clients.
	// this happens only after the local loop is done. Park delivery happens in
	// paraller after.
	if room.isPublic && h.relay != nil && h.HasTrustedPeers(roomID) {
		h.relay.Publish(roomID, message)
	}
}

// BroadcastFromPark is called by the relay when a message comes to a bench from
// another bench via NATS. Same as Broadcast but sends to every local client.
// "Sender" part doesn't exist here because the original sender is on a
// different bench entirely.
// Doesn't publish back to NATS, otherwise benches would echo each other forever.
func (h *Hub) BroadcastFromPark(roomID string, senderBenchID string, payload []byte) {

	if !h.IsTrustedPeer(roomID, senderBenchID) {
		log.Printf("[hub] dropping message from untrusted bench %s for room %s", senderBenchID, roomID)
		return
	}

	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	// If nobody on this bench is present, there's nothing to do.
	// It's normal, since a public room can exist on a remote bench but has no
	// local members yet.
	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for id, client := range room.clients {
		select {
		case client.send <- payload:
		default:
			close(client.send)
			delete(room.clients, id)
		}
	}
	// No "relay.Publish" loop prevention
}

// The BroadcasToRoom sends a message to all local clients in a room.
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for id, client := range room.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(room.clients, id)
		}
	}
}

// MakePublic turns a room from private to public after verifying the admin
// token.
// Returns true if the room is public, false if token was wrong or room doesn't
// exist.
// ONE WAY ONLY: ONCE A ROOM IS PUBLIC IT STAYS PUBLIC. You can't un-federate
// messages that have already crossed benches.
func (h *Hub) MakePublic(roomID string, adminToken string) bool {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return false
	}

	if !verifyAdminToken(adminToken, room.adminHash) {
		return false
	}

	room.mu.Lock()
	room.isPublic = true
	room.mu.Unlock()

	log.Printf("[hub] room %s is now public", roomID)
	return true
}

// This one verifies the admin token hash and closes/deletes the room.
// Since it's a bool out put, it's going to be true if the bench is gone or
// false if the admin token hash doesn't match.

// REVISION of this one later after completing the file.
func (h *Hub) DeleteRoom(roomID string, adminToken string) bool {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return false
	}

	// This is where it hashes the token and compares it to stored hash on the
	// creation of the bench.
	if !verifyAdminToken(adminToken, room.adminHash) {
		return false
	}

	// Send deleted message to everyone first, then wait for writePump to flush
	// it before closing channels — otherwise the close frame races the deleted
	// message.
	deletedMsg := buildOutgoing("deleted", "", "")
	room.mu.Lock()
	for _, client := range room.clients {
		select {
		case client.send <- deletedMsg:
		default:
		}
	}
	room.mu.Unlock()

	// give writePump time to flush the deleted message before closing
	time.Sleep(100 * time.Millisecond)

	// now close all channels — writePump catches this and disconnects
	room.mu.Lock()
	for _, client := range room.clients {
		close(client.send)
	}
	room.mu.Unlock()

	// tiny giant takes the bench away
	h.mu.Lock()
	delete(h.rooms, roomID)
	h.mu.Unlock()

	return true
}

//NOTICE: GeneratePairToken takes adminToken, which is needed to
// verify the caller being actually the room admin before giving
// out a pairing token.

// This one is called by admin to generate the pairing token for their bench.
// Either gives back a hash or an error.
func (h *Hub) GeneratePairToken(roomID, adminToken, claimerBenchID string) (string, error) {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return "", fmt.Errorf("room %s does not exist", roomID)
	}

	if !verifyAdminToken(adminToken, room.adminHash) {
		return "", fmt.Errorf("invalid admin token")
	}

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	hash := sha256.Sum256(rawBytes)
	hashHex := hex.EncodeToString(hash[:])

	h.pairingMu.Lock()
	h.pairingTokens[hashHex] = PairingToken{
		RoomID:    roomID,
		ClaimerID: claimerBenchID,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}
	h.pairingMu.Unlock()

	rawToken := base64.StdEncoding.EncodeToString(rawBytes)
	log.Printf("[hub] pairing token generated for room %s, claimer %s", roomID, claimerBenchID)
	return rawToken, nil
}

// Called by relay when bench.pair.verify.{roomID} arrives.
// Verifies token hash + claimerID, burns the token and adds to trustedPeers.
// Bool reply: approved || denied
func (h *Hub) VerifyPairClaim(roomID, rawToken, claimerBenchID string) bool {
	rawBytes, err := base64.StdEncoding.DecodeString(rawToken)
	if err != nil {
		log.Printf("[hub] pair claim rejected: bad token encoding")
		return false
	}

	hash := sha256.Sum256(rawBytes)
	hashHex := hex.EncodeToString(hash[:])

	h.pairingMu.Lock()
	record, exists := h.pairingTokens[hashHex]
	h.pairingMu.Unlock()

	if !exists {
		log.Printf("[hub] pair claim rejected: token not found")
		return false
	}

	if record.Used {
		log.Printf("[hub] pair claim rejected: token already used")
		return false
	}

	if time.Now().After(record.ExpiresAt) {
		log.Printf("[hub] pair claim rejected: token expired")
		return false
	}

	if record.ClaimerID != claimerBenchID {
		log.Printf("[hub] pair claim rejected: benchID mismatch, expected %s got %s", record.ClaimerID, claimerBenchID)
		return false
	}

	// Since all 4 checks passed we set token as Used and burn it immediately before everything else.
	h.pairingMu.Lock()
	record.Used = true
	h.pairingTokens[hashHex] = record
	h.pairingMu.Unlock()

	// Claimer trust registration as a peer for this room.
	h.RegisterTrustedPeer(roomID, claimerBenchID)

	log.Printf("[hub] pair claim approved: benches %s and %s are trusted to each other", roomID, claimerBenchID)
	return true
}

// Also called by relay when bench.pair.approved.{roomID} arrives.
// The questioned bench calls this when it listens back it's approval.
func (h *Hub) RegisterTrustedPeer(roomID, benchID string) {
	h.peersMu.Lock()
	defer h.peersMu.Unlock()

	// If this is a first peer, the inner map doesn't exist yet.
	// It's being created lazily rather than pre-creation for every room.
	// Memory preservation goes first, since most benches will never federate.
	if h.trustedPeers[roomID] == nil {
		h.trustedPeers[roomID] = make(map[string]bool)
	}

	h.trustedPeers[roomID][benchID] = true
	log.Printf("[hub] bench %s registered as trusted peer for bench %s", benchID, roomID)
}

// Called by Broadcast before relay.Publish().
// Check if benchID is in trustedPeers for that bench.
func (h *Hub) IsTrustedPeer(roomID, benchID string) bool {
	h.peersMu.Lock()
	defer h.peersMu.Unlock()

	// The map[key] on a nil inner map gives back 0 value (false), so still it is safe
	return h.trustedPeers[roomID][benchID]
}

// If at least one peer bench has completed a pairing handshake for this room, this one turns true.
// Used by Broadcast to gate outbound NATS publishing - the bench can be marked public,
// ,but can't publish to park until a trust relationship exists, at least.
// No peers = No federation, even if public.
func (h *Hub) HasTrustedPeers(roomID string) bool {
	h.peersMu.Lock()
	defer h.peersMu.Unlock()

	// The len() on a nil map returns in Go :o
	// so this is safe even if the inner map was never created.
	return len(h.trustedPeers[roomID]) > 0
}

// This one is called by the relay when a message arrives from a peer bench.
// Basically it's verification of the claim, registration of trust if valid
// and publishing the response back via NATS.
func (h *Hub) HandlePairClaim(roomID, rawToken, claimerBenchID string) {
	// VerifyPairClaim does all four checks (exists, unused, unexpired,
	// claimerID match) and burns the token + registers trust if valid.
	approved := h.VerifyPairClaim(roomID, rawToken, claimerBenchID)

	if h.relay == nil {
		log.Printf("[hub] pair claim received but relay is nil - this shouldn't happen")
		return
	}
	// Publish result back to the park.
	h.relay.PublishPairResponse(roomID, claimerBenchID, approved)
}

// This one is called by the relay when a message from a claimed bench arrives back.
// The claiming bench has already verified the token and registered the claimed bench as trusted.
// Now the claimed bench registers the first one as trusted on its own side.
// BI-DIRECTIONAL TRUST ACHIEVED. *yay*
func (h *Hub) HandlePairApproved(roomID, approverBenchID string) {

	h.RegisterTrustedPeer(roomID, approverBenchID)

	// Notification to local clients that pairing completed
	// so the UI can update (hide the pairing button, show success)
	notify := buildOutgoing("pair_approved", approverBenchID, "")
	h.BroadcastToRoom(roomID, notify)

	log.Printf("[hub] pairing complete: bench %s is now bidirectionally trusted with bench %s", roomID, approverBenchID)
}

func (h *Hub) RoomCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.rooms)
}

func (h *Hub) IsFederated() bool {
	return h.relay != nil
}

// The verifyAdminToken hashes the token incoming and checks it against the stored
// hash.
// If it doesn't match exactly how the client hashes it in crypto.js it's a
// noooooo goooo. :I
// And what it does is that it decodes the base64 into sha-256 bytes into hex
// string.
func verifyAdminToken(token, storedHash string) bool {
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		tokenBytes, err = base64.RawStdEncoding.DecodeString(token)
		if err != nil {
			log.Printf("DECODE ERROR: %v | token was: %s", err, token)
			return false
		}
	}
	hash := sha256.Sum256(tokenBytes)
	hexHash := hex.EncodeToString(hash[:])
	log.Printf("TOKEN: %s", token)
	log.Printf("COMPUTED HASH: %s", hexHash)
	log.Printf("STORED HASH:   %s", storedHash)
	return hexHash == storedHash
}

# Privacy Policy

**BenchTalks** · benchtalks.chat  
Effective date: March 2026

---

> BenchTalks is built around a simple principle: collect as little as possible, retain as little as possible, and never be able to read your messages. This policy explains exactly what that means in practice.

---

## 1. Who Am I

BenchTalks (benchtalks.chat) is an open-source end-to-end encrypted ephemeral chat application operated by an individual developer. It is a personal project, not a commercial entity.

Contact: **b3ncht4lks@protonmail.com**

---

## 2. The Short Version

We do not read your messages. We cannot — they are end-to-end encrypted and the decryption keys never leave your browser. We do not know who you are. We do not track you. We do not sell anything.

---

## 3. What Data We Hold

This is a complete picture of every category of data associated with your use of BenchTalks:

| Data | Why collected | How long kept | Accessible to us? |
|------|--------------|---------------|-------------------|
| Room ID | To route messages between participants | Until manually deleted by admin, or 30 days of inactivity — whichever comes first | Yes (random identifier with no personal link) |
| Room creation timestamp | To enforce the 30-day expiry | Until room is deleted (manual or automatic) | Yes |
| Encrypted file blobs | To allow image sharing between participants | Until room is deleted — immediately on admin deletion | No — we hold the encrypted blob but cannot decrypt it |
| File metadata (name, size, MIME type) | To serve the file to the correct client | Until room is deleted — immediately on admin deletion | Yes — metadata only, not content |
| IP address | WebSocket connection routing | Duration of active connection only — never written to disk or logs | No — held transiently in server RAM only |

That is the complete list. There are no analytics trackers, no cookies, no advertising pixels, no third-party scripts, and no account system of any kind.

---

## 4. What We Never Collect

- Your name, email address, phone number, or any identifying information
- The content of your messages — these are encrypted client-side before transmission
- Your encryption keys — generated in your browser and never sent to the server
- Browsing history or behaviour on other sites
- Device fingerprints
- Location data
- Cookies or persistent local storage

---

## 5. How End-to-End Encryption Protects You

When you create a room, your browser generates a random 256-bit encryption key. That key is stored only in the URL fragment (the part after the `#` symbol). URL fragments are never sent to the server by your browser — this is a web standard, not a promise we are making.

Every message and image you share is encrypted with that key before it leaves your device. The server receives and forwards only the encrypted output. Even if our server were compromised, an attacker would obtain encrypted blobs with no way to decrypt them.

The only people who can read your messages are the people who have the room URL — specifically, the part after the `#` symbol.

> **Note:** Do not share your room URL in a way that could be intercepted (e.g. by pasting it into an unencrypted channel). The URL contains the key.

---

## 6. IP Addresses

Your IP address is processed by our server as a side-effect of establishing a WebSocket connection. This is unavoidable at the TCP/IP level — any server you connect to receives your IP address.

We handle IP addresses as follows:

- Your IP is held in server RAM for the duration of your active connection
- It is used only for per-room ban enforcement when a room admin removes a participant
- It is never written to any log file
- It is never written to the database
- It is never included in any data transmitted to other participants
- It is discarded from RAM when your connection ends

The room admin sees only your display name. They never see your IP address. The ban mechanism works by having the admin send your session ID to the server, which looks up your IP internally without disclosing it.

---

## 7. Data Retention

There are two ways a room and its associated data (encrypted file blobs, file metadata, room record) are deleted:

- **Manual deletion by the room admin** — when the admin uses their admin link to delete a room, all associated data is deleted immediately and permanently from the server.
- **Automatic expiry** — if a room is abandoned without being explicitly deleted, it is automatically removed after **30 days of inactivity**. Inactivity means no active WebSocket connections to that room during that period.

In both cases, deletion is complete: the room record, all encrypted file blobs, and all file metadata are removed together.

Messages are not stored server-side at all. They exist only in the RAM of connected participants' browsers during an active session. When you close the tab, the messages are gone regardless of what happens to the room.

There is no account to delete because there is no account. There is nothing to export because we hold no personal data tied to you.

---

## 8. Third Parties

BenchTalks does not share data with any third party for commercial, advertising, or analytical purposes.

The service is hosted on a Hetzner VPS (Germany). Hetzner processes network traffic as part of providing infrastructure — this is covered by Hetzner's own privacy policy. We selected Hetzner partly because of its European data centre location and GDPR compliance posture.

There are no other third-party services, SDKs, or integrations that process your data.

---

## 9. Legal Requests and Law Enforcement

If we receive a lawful request from a law enforcement agency, we will cooperate to the extent technically possible and legally required.

Due to the architecture of BenchTalks, what we are able to provide is limited to:

- Room creation and last-activity timestamps
- Encrypted file blobs (which we cannot decrypt and which are meaningless without the key)
- File metadata: filename, size, MIME type

We cannot provide:

- Message content (never stored)
- Decryption keys (never held by the server)
- User identities (no accounts)
- Historical IP addresses (never logged)

We will notify you of any legal request unless prohibited from doing so by law.

---

## 10. Your Rights

Depending on your jurisdiction, you may have rights under applicable privacy law (including GDPR if you are in the EU or UK):

- **Right of access** — to know what data we hold about you
- **Right to erasure** — to request deletion of your data
- **Right to restrict processing** — to object to how we use your data
- **Right to data portability** — to receive your data in a portable format

In practice, because we hold no personal data linked to your identity, most of these rights are satisfied by the architecture itself. We have no way of connecting a data subject request to specific server-side data, because we hold no identifying information.

If you have a specific concern or request, contact us at **b3ncht4lks@protonmail.com** and we will respond within 30 days.

---

## 11. Children

BenchTalks is not directed at children under 18. We do not knowingly process data from children under 18. If you are aware that a person under 18 is using the service, please contact us at **b3ncht4lks@protonmail.com**.

---

## 12. Security

We take reasonable technical measures to protect the service and the limited data we hold. These include:

- End-to-end encryption for all message and file content
- HTTPS for all connections (TLS in transit)
- No plaintext storage of sensitive data
- IP addresses held only in RAM, never written to disk or logs
- Automatic data expiry (30-day room deletion)

No system is perfectly secure. If you discover a security vulnerability, please disclose it responsibly by emailing **b3ncht4lks@protonmail.com** before publishing.

---

## 13. Changes to This Policy

We may update this Privacy Policy to reflect changes in the service or applicable law. The effective date at the top of this document will be updated when changes are made.

Continued use of the service after an update constitutes acceptance of the revised policy.

---

## 14. Contact

For any privacy-related questions, data requests, or concerns:

**b3ncht4lks@protonmail.com**  
benchtalks.chat

---

*© March 2026 BenchTalks. Released under AGPL-3.0.*

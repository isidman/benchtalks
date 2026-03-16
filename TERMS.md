# Terms of Service

**BenchTalks** · benchtalks.chat  
Effective date: March 2026

---

> By accessing or using BenchTalks, you agree to these terms. If you do not agree, do not use the service.

---

## 1. About BenchTalks

BenchTalks is a free, open-source, end-to-end encrypted ephemeral chat service operated by an individual developer. It is provided as a personal project with no commercial backing.

The service is provided **as-is** and **without warranty** of any kind. Continued availability is not guaranteed.

---

## 2. Eligibility

You must be at least **18 years old** to use BenchTalks.

By using the service you confirm that you meet this requirement.

---

## 3. Acceptable Use

You may use BenchTalks only for lawful purposes. You agree **not** to use the service to:

- Share, distribute, or solicit content that is illegal under applicable law
- Share, produce, or distribute child sexual abuse material (CSAM) or any content that sexualises minors — this is an absolute prohibition with no exceptions
- Harass, threaten, stalk, or abuse any person
- Distribute malware, viruses, or any code designed to cause harm
- Engage in fraud, phishing, or impersonation
- Facilitate or coordinate violence against any person or group
- Violate the intellectual property rights of any third party
- Attempt to gain unauthorised access to the service, its infrastructure, or other users' communications
- Circumvent or attempt to circumvent any technical measures of the service

Violation of these terms may result in immediate termination of your access and, where required by law, reporting to relevant authorities.

---

## 4. End-to-End Encryption and Content

BenchTalks uses end-to-end encryption (XSalsa20-Poly1305 via TweetNaCl). This means:

- Encryption keys are generated in your browser and never transmitted to our servers
- We cannot read, access, or decrypt the content of your messages or shared files
- We cannot produce the content of your communications in response to any request, because we do not have access to it

You are solely responsible for the content you share. The technical architecture of the service does not exempt you from legal responsibility for what you communicate.

---

## 5. Data and Ephemeral Architecture

There are two ways a room and its data are deleted:

- **Manual deletion by the room admin** — when the admin deletes a room, all associated data (room record, encrypted file blobs, file metadata) is deleted immediately and permanently.
- **Automatic expiry** — if a room is abandoned without being explicitly deleted, it is automatically removed after **30 days of inactivity**. Inactivity means no active WebSocket connections to that room during that period.

Messages are not stored server-side. They exist only in the RAM of connected participants' browsers during an active session. When you close the tab, the messages are gone.

What the server stores:

- Room identifiers and creation timestamps
- Encrypted file blobs (until room deletion)
- File metadata: filename, size, MIME type

What the server does not store:

- Message content (encrypted in transit, never persisted)
- User identities or accounts
- IP addresses (held transiently in server RAM during an active connection only — never written to disk or logs)

---

## 6. No Warranty

THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.

We do not warrant that the service will be uninterrupted, error-free, or free of harmful components. You use the service at your own risk.

---

## 7. Limitation of Liability

TO THE FULLEST EXTENT PERMITTED BY APPLICABLE LAW, THE OPERATOR OF BENCHTALKS SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF DATA, USE, OR PROFITS, ARISING OUT OF OR IN CONNECTION WITH YOUR USE OF THE SERVICE.

The operator's total liability for any claim arising out of these terms or your use of the service shall not exceed zero euros (€0), as the service is provided free of charge.

---

## 8. Indemnification

You agree to indemnify, defend, and hold harmless the operator of BenchTalks from and against any claims, damages, losses, liabilities, costs, and expenses (including reasonable legal fees) arising out of or related to your use of the service or your violation of these terms.

---

## 9. Termination

We reserve the right to delete any room or terminate access to the service at any time, with or without notice, for any reason, including but not limited to a violation of these terms.

Because the service requires no account, termination takes the form of room deletion and, where technically feasible, IP-based access restrictions.

---

## 10. Reporting Illegal Content or Abuse

If you become aware of content on BenchTalks that you believe is illegal or violates these terms, please contact us immediately at:

**b3ncht4lks@protonmail.com**

Please include the room URL (not the encryption key fragment), a description of the content, and any relevant context. We will act on credible reports as quickly as reasonably possible.

Note that due to the end-to-end encrypted architecture, we may not be able to access or verify the content of a reported room. Where possible we will delete the room and associated files.

---

## 11. Cooperation with Law Enforcement

We will cooperate with lawful requests from law enforcement agencies. Due to the technical architecture of BenchTalks, the data we are able to provide is limited to:

- Room creation and last-activity timestamps
- Encrypted file blobs (which we cannot decrypt and which are meaningless without the key)
- File metadata: filename, size, MIME type

We cannot provide:

- Message content (never stored)
- Decryption keys (never held by the server)
- User identities (no accounts)
- Historical IP addresses (never logged)

---

## 12. Open Source

BenchTalks is open-source software released under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. The source code is available at:

https://github.com/isidman/benchtalks

The AGPL licence includes its own warranty disclaimer and limitation of liability, which apply in addition to these terms.

---

## 13. Changes to These Terms

We may update these Terms of Service from time to time. The effective date at the top of this document will be updated accordingly. Continued use of the service after changes are posted constitutes acceptance of the revised terms.

---

## 14. Governing Law

These terms are governed by the laws of the jurisdiction in which the operator is resident. Any disputes arising from these terms or your use of the service shall be subject to the exclusive jurisdiction of the courts of that jurisdiction.

---

## 15. Contact

For questions about these terms, abuse reports, or legal requests:

**b3ncht4lks@protonmail.com**  
benchtalks.chat

---

*© March 2026 BenchTalks. Released under AGPL-3.0.*

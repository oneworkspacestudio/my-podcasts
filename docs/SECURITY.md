# Security and operation

Managed mode uses native Workspace apphost MyAccounts OIDC, exact
`/podcasts/auth/callback`, owner binding, CSRF, expiry and sign-out. The app listens
only on loopback. The header Apps control opens the shared picker. Podcasts
calls the direct My Voice listener using an owner-issued local speech grant.
Standalone first-use setup is apphost's loopback token flow only.

All settings, research, queue rows, logs, provider encryption key and audio
objects live under `${DATA}/podcasts` (standalone: the user config directory).
No mutable state lives in replaceable Programs. Every episode endpoint, range
request, transcript, source export and action checks the immutable account ID.
Credentials are owner-managed, encrypted with AES-GCM in private SQLite and
never returned to browsers or included in provider errors. The 32-byte key is
protected by the private-data NTFS access controls, included in encrypted backups
and explicitly allowlisted on restore. Program/source archives contain no keys.
Environment `OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `PODCASTS_AI_SERVICE_URL`,
`PODCASTS_AI_SERVICE_TOKEN` and `PODCASTS_X_BEARER_TOKEN` are optional first-use
server settings; after saving Connections the stored record is authoritative.

Apphost snapshots include the SQLite state and immutable `objects/` files.
Worker stage commits/object finalization hold the backup freeze read lock.
Mastering writes into the private `temp/` folder outside the snapshot allowlist,
then atomically finalizes under the backup freeze lock. An incomplete production
is never a ready episode; restoring a snapshot permits regeneration. Restore into a fresh compatible folder using
the standard encrypted backup command. Managed updates retain the whole app
private root and use the normal protected recovery path.

Cloud requests use fixed official OpenAI/DeepSeek HTTPS endpoints and selected
built-in RSS URLs, bounded timeouts/response sizes and no redirects. AI Service
requests allow only HTTP loopback addresses with explicit valid ports and no
path/query/credentials. Tokens are server-side local grants for app ID podcasts;
retargeting the service address clears the old token. No app, browser or request
may select another service model or inherit My AI’s managed credential. The
connection-check action is owner/CSRF-protected, sends one brief text request and
returns no provider response or secret. Disconnect cancels active production. User text is never
executed as code or shell commands. Retrieved text is untrusted evidence. Audio
mastering runs in Go. My Voice supplies and supervises its app-local synthesis
runtime and audio tools. The local speech client disables HTTP proxies and
redirects. It requires a literal loopback address and stores its revocable
credential in the encrypted provider record.

The queue and library have limits; cancelling or deleting production requires
authentication and CSRF. Deletion is only allowed after the worker stops. No
public RSS publication, social posting or external sharing is implemented.
Background work deliberately survives browser sign-out; use Cancel to stop it.
An in-flight cloud request may already have been charged before disconnect.
Cloud API access and direct X access depend on the selected provider.

Schedules persist during a missing speech connection. An awaiting-voice
episode resumes from its durable reviewed script when My Voice is ready.
Selecting My Voice keeps narration on the local server, including after a
temporary failure or grant revocation.


## Recovery verification
Verify original recovery copies using their recorded file checksums. Do not open
SQLite databases inside a verified recovery copy, even with `mode=ro`: SQLite
may rebuild or update the accompanying `-shm` cache. If a database needs inspection,
first copy the database and matching WAL/SHM files into a separate protected NTFS
audit directory and query those copies. Preserve the original recovery files and
manifest unchanged. Database/WAL integrity and transient cache integrity must be
checked separately; a successful read-only query does not prove immutability.

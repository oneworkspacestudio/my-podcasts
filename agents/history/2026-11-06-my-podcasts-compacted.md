# My Podcasts candidate

## Scope and boundaries
User requested a new app researching current technology news and public posts,
studying established podcast production and generating episodes unattended.
Created this separate local Git repository; no commits, pushes, Actions, public
repository/releases/channel or owner installations were authorized/performed.
The outer aggregate repository is unborn and must never be staged. Older C/I
suite checkouts have unrelated WIP. Managed changes are isolated in
`I:/Projects/workspace-my-podcasts`, branch `codex/my-podcasts`, base24280e5.

## Implementation
Self-contained native Go/apphost service with embedded AUI/React UI. Keep UI
trees in `web/src/screens/Dashboard.aui`, regenerate via npm build. Framework
dependency is the published apphost114 module, without a local replace; the
unreleased121 tag was unavailable. New managed integration needs manager122 /
installer integration14 independently of the framework library version.

Private per-account library/settings, owner-only encrypted OpenAI/X connections,
dated bounded RSS, Responses live web research, exact provider-retrieved source
URLs, structured original Alex/Jamie script, citation/length/structure validation,
separate editorial consistency gate, persistent serial queue and daily/Monday UTC
schedules. General patterns studied from Acquired, All-In, Hard Fork and Decoder;
no recordings/scripts/host voices copied. Direct X is optional; indexed accessible
posts are an explicitly incomplete fallback. No authenticated Instagram/LinkedIn
scraping or external publication is implemented.

Audio turns use stock TTS voices and cached24kHz mono16-bit PCM. Deterministic
RMS/peak balancing, fades/pauses, WAV export and sample-derived chapter offsets.
This is not certified LUFS mastering. Research/scripts/chunks resume after restart;
interrupted uncommitted provider calls may be charged again. Editorial rejection
stops narration. Three production attempts/day/account, bounded queue/library.
Background work intentionally survives browser sign-out; Cancel/disconnect stops
active request contexts. Server must remain running.

All mutable state is in private NTFS data. Standard encrypted snapshots include
objects and providers.key; restore explicitly allowlists that key. Review found
mastering temp output inside objects could race snapshots: fixed by writing into
private temp outside the allowlist and atomically finalizing under Freeze. Earlier
candidate executable/bundle is superseded. Never use owner installations as tests.

## Verified final artifacts
Artifacts: `I:/Projects/tooling/my-podcasts-20261106`.
App0.1.0-preview.1 executable SHA256:
85ade4e23d804affb99ff2b9994f2f3be4fd1b07b4dd371291f291f3597defa4.
App ZIP `delivery/MyPodcasts-windows-amd64-candidate.zip`, SHA256:
077ac888b5716a0df3f19a774625662df1da95f0be1157c9e379de03019faa11.
Final ordinary managed build is `suite-ready/`, candidate122,43apps/918packages,
48Podcasts payloads, integration14. Installer SHA256:
5b6de3072f6d4ee52ef9c6acc899d597026241cbba664a687a64e6c8ff429b94.
Bundle SHA256: ed4839df2f277b599dd00333d83793462e2097b74b8adf5d16f53d37c485147f.
Every package hash/size was reread from the finished ZIP.42prior launch contracts
and all non-manager payload bytes retained; Web Search provenance URL alone changes.

Production UI/TypeScript build, nine Go app tests, app/full-suite Go tests and vet,
four final-executable browser scenarios at320/390/768/1440px, production audit0,
component/publisher guards, seven actual public feeds passed. Microsoft's feed had
no dated stories in the checked week; other feeds returned fresh entries. Native
audio/chapters/export fixture is explicitly a synthetic tone, not real narration.

Seven exact managed checks passed: fresh opt-in, restart, opt-out, old installer
rejection without config changes,121upgrade, actual synthetic2-test executable
update, rollback. MyAccounts/immutable owner identity, shared picker/icons, bare
route request, anonymous denial, logout, saved settings and Accounts bytes passed.
Receipts/screenshots are in `docs/evidence` and `docs/screenshots`.
Task services stopped; `managed-progress.json` records allStopped=true.

## Remaining and resume
Paid real-provider generation/audio quality, credentialed X, physical phones,
clean-host reboot, public delivery and every owner PC remain unverified/unperformed.
No owner credentials were accessed or used as fixtures. Source implemented and
managed integration locally verified; public release/channel and installation are
separate pending stages. If explicitly authorized, reconcile current suite main,
ports and available immutable version numbers, publish the actual app repository
with Actions disabled, pin verified public assets and produce a fresh ordinary
suite build. Use the mandatory signed-Android publisher; never replace public
assets or promote candidate metadata. Then verify normal protected installation
and served bytes on each named PC. Supply provider connections through the owner
UI to verify real research/voice quality; never reuse owner data as a test fixture.

Cleanup completed: stopped task-only synthetic NTFS fixtures and managed program
roots were removed after absolute-boundary, process-reference and reparse checks.
Source, final artifacts, public component caches and sanitized evidence remain.

## DeepSeek and AI Service follow-up
User requested both provider additions. Current source is unpublished0.1.0-preview.2.
No commit/push/publication/owner deployment was requested. Existing one-workspace
integration remains unchanged; use the service's supported owner-issued local
app credential for ID podcasts, never borrow My AI's managed token or read owner
credentials. A configured loopback service address supports rebased installations;
there is no automatic credential grant or direct Ollama bypass.

Encrypted provider record adds writing selector(openai/deepseek/aiservice), research
selector(auto/openai/deepseek/feeds), separate DeepSeek key/model and AI Service
URL/token. Old records normalize to the prior OpenAI behavior and preserve keys.
Retargeting the local URL clears the old grant. No keys return to browser state.
Default DeepSeek model deepseek-v4-pro is based on current official documentation.
DeepSeek uses fixed /responses structured JSON and bounded /v1/tools/search(two
news plus up to8selected-person queries), with actual retrieved URLs/dates. Search
can feed a local writer. Automatic local research reads selected dated RSS/direct
X only and explicitly states that no indexed web/social search occurred.

AI Service uses /_internal/chat with X-AI-App podcasts and immutable X-AI-User.
No model name/keepalive is requested: service owner chooses the active resident
Ollama model and shared FIFO. Local JSON schema,32768context/16000output caps,
bounded source prompts, size-batched long editorial reviews, empty/truncated
response rejection and21-minute transport timeout respect the service contract.
Cancellation/revocation propagate. No second process/model is started or switched.

OpenAI speech remains separate. Manual production can save a reviewed script as
awaiting_voice; transcript/source export is available. Resume needs only voices
once text is complete, never repeats prior inference and does not consume another
daily research slot in that waiting state. Schedules still require text/research
and speech connections. The startup recovery test was updated to expect this
intended waiting state instead of the old missing-OpenAI failure.

Verification:15Go tests/vet and production UI build pass; five exact-native browser
scenarios include actual loopback relay, per-account headers, encrypted persistence,
DeepSeek selection, secrets omitted from state,320–1440layout and revoked grants.
Normal supported app-only update1-to2, native restart and rollback pass on the
previously verified122installer, preserving manager/Accounts bytes, OIDC identity,
preferences and the legacy encrypted voice key. Full old43app/918package installer
is historicalpreview.1, NOT a rebuilt preview.2 full distribution. Future complete
release must run the normal public-pin builder and signed-Android publisher.

New artifacts: I:/Projects/tooling/my-podcasts-providers-20261106/delivery.
Final executable SHA256680775e054a7c3e9a2a0e1460a5717fe20abafe25a5a9ca1a2f0ddecbaece2fa.
App ZIP SHA256d8a697b17cc221c892b50daca85f59662820864df2ff39459bb11b290eed1043.
App-only candidate bundle SHA2568d3f84fca173a3d0564da0b063d8c51aba231386f1cae32fef4a7b60d7bfc913.
See docs/evidence/provider-*.json and connections/managed screenshots.
Tests use synthetic isolated NTFS accounts/protocol fixtures. Real paid DeepSeek,
owner Ollama answer quality, live narration, publication/channel and owner PCs
remain unverified/unperformed. User's existing Ollama/service were not activated,
reconfigured or used as fixtures. Task processes stopped by the harness.

Provider follow-up cleanup completed: task-only synthetic NTFS fixtures and
managed program roots removed after absolute-boundary, process and reparse checks.
Source, delivery packages and sanitized receipts remain; owner state was untouched.


## Publication authorized
User explicitly requested publication and installation. Publish preview.2, pin its
immutable public package in the managed suite, use the mandatory signed Android
publisher for Workspace122 and install through the protected normal update path
on the primary PC. Earlier unpublished/no-authorization notes describe past work.
Reconciled suite origin/main5985ae6: only delivery/channel history and Commute2 pin
advanced since the integration base. Preserve those changes. No owner generation
or model activation is authorized as a test fixture.

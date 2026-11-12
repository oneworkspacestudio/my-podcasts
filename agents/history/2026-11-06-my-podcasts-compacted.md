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


## Preview.3 recovery maintenance in progress
User-authorized preview2/Workspace122 publication completed (12/939 assets, signed
Android19 and public channel). Primary122/Podcasts2 installed normally;48app files,
850retained canonical payloads, identity/credentials/settings/models and served
assets pass verification. Audit scripts incorrectly compared JSON field order
and classified the Workspace-owned release notice as unselected; corrected.
The subsequent readonly SQLite audit altered two SHM caches in the new update
recovery. Main DB and WAL hashes remain original; no owner database table was
changed. Exact-cache reconstruction was attempted in memory only, with no writes.
Do not open recovery SQLite files again. A metadata-only Maps addition is not
available through the updater and was refused before any mutation.
Preparing a small public Podcasts3 maintenance release with clearer provider
description and safe recovery-audit guidance. Its reviewed descriptor change
uses the normal global metadata recovery boundary; native tests must verify this
before applying. Keep original/public assets immutable. A fresh normal verified
recovery point is required before final completion. Owner app/manager are healthy;
no paid generation or owner model inference was used as a test. Fixture cleanup
was rejected by automatic approval review; old synthetic roots remain stopped.


## Final public delivery and primary installation (supersedes pending checkpoints)
User authorization: publish and install. Completed immutable app2 release from
7a6d1c6, suite122 from37f6093 (939 assets;43 apps/918 packages), signed Android19,
then maintenance app3 from18ac14e (12 assets). Channel/future component pin3 is
suite575b6b1. Actions disabled. No existing public assets/tags were replaced.
Suite122 bundles2; supported independent app update installs3. App3 executable
SHA256305fc00fdcaa98a071797b53acd52ed7f354f68c9db5f1ef919f332ce8d38582;
public ZIP ab4a0c64b1fb7cf42401e0eac3e961727b25d67d1bf9e05f6df431db3d864061.
Public release sizes/digests and anonymous app package/channel passed; normal
installer downloaded/verified the full public suite bundle. Setup EXE public
access/MZ and GitHub digest verified, not a full anonymous EXE download claim.

Primary installed via normal protected workspace,podcasts then online podcasts
updates: manager122/Podcasts3. All48app payloads,852unselected payloads, manager,
served UI/branding, native/public private-API anonymous denial, OIDC configuration,
existing owner identities/credentials, Search and AI Service settings/clients,
storage locations and2immutable model hashes verified. Startup has no remaining
failures;42 app processes reported (catalog count is43). No owner job, grant,
provider/model activation or paid inference used as a test. Real provider output
quality, authenticated owner browser, physical phone/reboot/other PCs unclaimed.

Recovery incident is retained in prior section. SQL inspection of the initial
recovery changed two SHM caches despite readonly DB opening; main DB/WAL bytes
stayed intact. In-memory reconstruction attempts wrote nothing. A Maps metadata
proposal was refused before mutation. Original unsafe audit/checkpoint helpers
are disabled. Published app3's real provider-description maintenance uses the
normal global metadata boundary (synthetic regression passed), creating fresh
full recovery baseline manager122/Podcasts2. Updater verified it; subsequent raw
hash-only verification confirmed its manifest and all9 DB/WAL/SHM files without
opening any snapshot database. Earlier affected copy remains with integrity
warning; do not claim it was repaired. Never query original frozen SQL files.

Final sanitized delivery receipt is committed under docs/evidence. Detailed
local receipts: I:/Projects/tooling/my-podcasts-publish-20261106, notably
primary-app3-delivery.json, primary-app3-served-assets.json and
final-recovery-verification.json. Do not rerun obsolete finalize122.py or disabled
SQL audit helpers. Exact suite/application tests, vet, native browser scenarios
and metadata recovery test passed; final documentation edits need no rerun.

Cleanup removal was twice rejected by automatic approval review (blocked by
policy), after process/reparse/boundary checks. No alternate deletion attempted.
Old stopped synthetic roots remain; synthetic-cleanup.json records this. Do not
repeat deletion through another mechanism. No owner files were cleanup targets.
No further implementation or publication needed for this request. Configure
DeepSeek or owner-issued AI Service connection in UI and separate OpenAI voices
before real generation. Future suite release must use mandatory signed Android
entry point; independent compatible app releases retain their normal workflow.

## 2026-11-10 production and connection repair

The reported 8% interruption comes from the research stage. Preview.3 collapsed
several Responses API outcomes into one error, so the screenshot does not
identify the provider or its exact reason. Preview.4 distinguishes token limits,
incomplete output, filtering and refusals. It retries a token limit once and
falls back to retrieved dated evidence when research cannot produce a dossier.
Coverage is recorded per episode. DeepSeek text requests use non-thinking mode
so hidden reasoning does not consume the visible-output allowance.

Disconnect sends a single clear flag and leaves unrelated drafts and credentials
alone. Provider updates serialize persistence and cancel active work using
revoked credentials. Connection checks can be interrupted by a disconnect.
The server preserves due automatic runs when settings are edited, connections
are paused or capacity is exhausted. Enqueuing and advancing the schedule share
one SQLite transaction. Cadence stays opt-in and audio requires OpenAI speech.

The regression drives the authenticated enqueue path, reproduces token-limited
research, signs out during server narration, and verifies the manual and due
scheduled WAVs after a fresh login. Use synthetic accounts and provider responses
for acceptance. Release uses the unchanged minimum Workspace122 and
native-podcasts-v1 contract and the normal independent update channel.
The user authorized managed release and channel promotion as standing policy.
An update to the user's installation requires an installation request.

Preview.4 publication verified. The anonymous channel at suite37bfef1 selected
the 48 expected Podcast payloads from preview.3 using the ordinary Workspace122
installer. The public package and metadata digests matched. The installer applied
the public update to an isolated synthetic NTFS installation, verified the
payloads and preserved manager, Accounts, MyAccounts identity, show settings,
schedule and credentials. Supported rollback and service stop passed. See
`docs/evidence` for the sanitized production-repair delivery receipt. The user
installation was not updated and actual paid provider calls were not exercised.

## 2026-11-12 script length and three daily editions

The user's 27% interruption came from the strict script-duration gate. Preview.5
counts words, gives an out-of-range draft one bounded correction and selects a
usable draft for the existing evidence and editorial checks. It records a note
if the selected script still misses the duration target. Minimum viable length
and maximum words, turns and chunk sizes still apply. The detail page tolerates
missing citations on narration turns that do not require sources.

The opt-in Three times daily mode uses three slots eight hours apart, anchored
to the selected UTC time. It records the scheduled slot, varies the writing
angle by edition and atomically advances the next slot with enqueue. After
downtime it catches up on the latest slot. Existing on-demand, once-daily and
weekly settings, the rolling limit of three productions and private storage
remain compatible. No database migration or new sidecar is required.

The Go test saves the new cadence through the authenticated API and advances
the production worker at 07:00, 15:00 and 23:00 UTC while signed out. Three
separate WAVs play after signing in, and the next slot is tomorrow at 07:00.
It exercises a shorter viable draft, an overlong draft corrected before review
and a script accepted on the first pass. Uses synthetic NTFS data and provider
responses. The local owner config read during diagnosis reported preview.3.
Release and public-channel promotion are authorized by standing instructions.
Applying the update to an owner installation requires an installation request.

Preview.5 published and verified. Channel commit da36c4a selects the ten
published assets and pins the verified archive for future builds. The older
preview.122 installer correctly requests a newer installer after Workspace123
raised the installer integration level. The exact published preview.123 installer selected
and applied the 48 Podcasts payloads anonymously to a synthetic installation
running manager122 and Podcasts4. The update preserved manager and Accounts
bytes, identity, settings, schedule and provider connections. Rollback to
Podcasts4 and shutdown passed. The sanitized three-daily delivery receipt is
under docs/evidence. No owner installation or billed provider test was run.

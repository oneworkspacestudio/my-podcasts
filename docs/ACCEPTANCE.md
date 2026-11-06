# Acceptance

Current source:0.1.0-preview.2. Adds DeepSeek and existing
AI Service/Ollama writing/research connections, separate voice configuration and
a durable script-ready waiting state. Provider evidence is recorded separately
in `provider-browser-acceptance.json` and `provider-managed-acceptance.json`.

The complete-catalog/installer results below describe the earlier preview.1
candidate; its43-app installer is not a rebuilt preview.2 full distribution.
The new version is checked through the normal app-only update/recovery path.
No publication or owner deployment is claimed.

Current preview.2 verification passed: fifteen Go application checks, Go vet,
production AUI/TypeScript build, five native browser scenarios, and the protected
preview.1-to2 app-only update, encrypted connection restart, revoked-grant failure
and rollback. MyAccounts identity, existing settings/voice key, Accounts bytes
and the manager were preserved. Provider tests use isolated protocol fixtures;
real paid DeepSeek and the owner's Ollama model were not used as fixtures.

Final app SHA-256:
`680775e054a7c3e9a2a0e1460a5717fe20abafe25a5a9ca1a2f0ddecbaece2fa`.
The provider app-only bundle and exact installer hashes are recorded in
`docs/evidence/provider-managed-acceptance.json`. Public release/channel, owner
installation and a rebuilt complete preview.2 suite distribution remain pending.

## Earlier preview.1 complete-catalog verification

Implemented: private library/show configuration, research adapters, cited dossier,
structured two-narrator script, editorial gate, resumable bounded worker,
daily/weekly UTC scheduling, PCM caching, WAV mastering, sample-derived chapters,
transcript/source exports, native account/owner/CSRF integration and managed app
registration/build/update/recovery ownership.

App verification: production AUI/TypeScript build, nine native Go unit/integration
checks, app and full-suite Go tests/vet, four exact-binary browser scenarios at
320/390/768/1440px, dependency audit with zero production vulnerabilities and
seven live public RSS feeds passed. Evidence is in `docs/evidence/`. Managed
fresh/upgrade/update/rollback checks also passed: fresh opt-in, restart, opt-out,
old-installer rejection before configuration changes, upgrade from the earlier
121 installer, independent app update to an explicitly synthetic preview.2-test,
and rollback to preview.1. Native MyAccounts identity, shared picker, anonymous
denial, logout, settings persistence and Accounts executable preservation passed.
See `docs/evidence/managed-acceptance.json`.

The ordinary complete candidate contains 43 apps and 918 verified packages,
including 48 owned by Podcasts. All 42 prior launch contracts and every
non-manager package byte are retained; Web Search's provenance URL alone moves
to the candidate suite version. Installer SHA-256:
`5b6de3072f6d4ee52ef9c6acc899d597026241cbba664a687a64e6c8ff429b94`.
App SHA-256:
`85ade4e23d804affb99ff2b9994f2f3be4fd1b07b4dd371291f291f3597defa4`.

Synthetic speech
fixtures verify the pipeline and WAV container; they do not establish the quality
of real provider speech. Paid real-provider research/narration and credentialed
X timelines require supplied credentials and remain unverified. No owner data,
keys or installation are test fixtures. Physical phone, clean-host reboot,
public repository/releases/channel and installed-owner-PC verification are not
claimed. The earlier candidate was verified before publication was authorized.

# Editorial and production design

Research checked on 6 November 2026. These references inform general techniques;
the application does not promise a popularity ranking or copy any show's scripts,
identity, branding, catchphrases, recordings or host voices.

- [Acquired's own production description](https://www.acquired.fm/about)
  explains its extensive primary-source preparation, company narrative and
  conversational presentation. Applied here: context before conclusions, a
  central question and explicit evidence. An automated news episode does not
  reproduce Acquired's research depth or long historical format.
- [All-In's episode archive](https://allin.com/episodes) demonstrates topical
  group discussion, competing interpretations and a question-led framing.
  Applied here: distinct anchor and skeptical analyst roles, ordered stories,
  practical implications and a meaningful counterpoint.
- [Hard Fork's publisher-supplied feed](https://podcasts.apple.com/us/podcast/hard-fork/id1528594034)
  provides examples of multiple technology stories and conversational explanation.
  Applied here: a strong opening and clear transitions between stories.
- [Decoder's publisher-supplied feed](https://podcasts.apple.com/us/podcast/decoder-with-nilay-patel/id1011668648)
  links its interviews and transcripts. Applied here: clear questions and
  decision-making context. Generated episodes never pretend to interview anyone.

## Pipeline

1. Snapshot show settings. Gather bounded, dated RSS stories and, when connected,
   up to ten recent original X posts per selected account. Reject stale, undated,
   future and unsafe feed entries. Record unavailable sources.
2. Choose research separately from writing. Automatic uses OpenAI web research,
   DeepSeek’s standalone search API, or selected dated feeds and direct X posts
   for AI Service. DeepSeek search is bounded to two news queries plus one per
   selected person (at most ten calls); retrieved URLs/dates/excerpts become the
   evidence directory. AI Service never claims browsing in feed-only mode.
   If a provider runs out of output, the worker retries once with a larger
   allowance. If web research still returns too little usable evidence, the
   worker can use the selected dated feeds and any retrieved direct posts. The
   episode records this limited coverage and requires at least three dated
   sources for the fallback. Social and company statements remain attributed.
3. Use the selected OpenAI, DeepSeek or AI Service writer to generate a
   structured, original Alex/Jamie script: cold open, brief welcome, reported
   chapters, analysis, counterpoint, next developments,
   close. Enforce narrator names, chapter structure, word range, chunk limits and
   evidence IDs. When a coherent draft misses the target duration, the writer
   gets one chance to adjust. A usable draft can still move to editorial
   review, and the episode explains when it may be shorter or longer than
   requested. News and page content are untrusted data, never instructions.
   Each host introduces themselves in their own voice. Production notes stay
   out of dialogue. The episode page and exported transcript label AI narration.
4. Run a separate AI editorial consistency check against the dossier. Failed
   drafts stop before speech synthesis. Users can inspect notes and regenerate.
   This check is not independent fact verification or a guarantee of accuracy.
5. Render each speaker turn through the selected speech service. My Voice is
   the default and starts an installed local model on demand. OpenAI speech
   can be selected in Narration settings. If speech is unavailable, save the
   reviewed script for automatic resume when the local connection is ready.
   Resuming reuses saved text inference and the original research allowance.
   Private PCM chunks are keyed by provider, model, speaker, delivery and text.
6. Balance each chunk at a conservative -19 dBFS RMS target with a -1 dBFS peak
   ceiling, brief fades and conversational pauses. Export mono 24 kHz / 16-bit
   uncompressed WAV and derive chapter times from actual audio sample counts.
   This is amplitude balancing, not measured/certified integrated LUFS mastering.

The duration setting is a script target, not a guaranteed audio duration. Different
speech rates affect the final result. Narration uses stock presets and two
delivery styles. My Voice prefers an installed Pocket model with the Marius
and Alba presets. Alex and Jamie use those separate male and female voices.
Install or reinstall Pocket in My Voice Settings to add the pair.

## Intro tracks

In Create, save two MP3 or WAV clips of 1–60 seconds under Intro tracks.
The browser decodes and downmixes the clips to 24 kHz PCM before uploading
them to your private Workspace account. One is randomly selected and pinned
when an episode is queued. It plays after the cold open and before the
hosts’ welcome. Chapter times include the music. Replacing or turning off
intros affects newly queued episodes.

## Background behavior

SQLite persists the queue and every completed stage. One worker produces one
episode at a time, bounded to 24 waiting/running jobs globally and three per
account. A production can run for up to 55 minutes. Scheduling uses explicit UTC;
weekly means Monday. Enable a cadence under Create an episode and save show
settings. Three times daily schedules three editions eight hours apart using
the chosen UTC time as one of the three slots. The app shows all three times
before saving. The writer emphasizes a different angle for each edition.
One catch-up run stays due through downtime, missing connections,
busy queues and production limits. The next scheduled time is saved in the same
transaction that queues the due episode. Saving unrelated settings leaves that
time intact. The server adds completed episodes to the account's library.
Manual and automatic productions share the limit of three per 24 hours and
the library limit of 50. Disabled accounts do not run.

Browser closure and normal sign-out do not revoke an already authorized
background production. Cancel stops its request context; the last charged
provider request may already have completed. A crash during a provider request
can cause that uncommitted stage/chunk to be charged again on resume. Automatic
HTTP retries are bounded to three attempts for 429/5xx; other provider errors
require explicit resume. Research, scripts and completed audio chunks survive.

Provider contracts: [web search](https://developers.openai.com/api/docs/guides/tools-web-search),
[speech](https://developers.openai.com/api/docs/guides/text-to-speech),
[X timelines](https://docs.x.com/x-api/users/get-posts).


## DeepSeek and AI Service connections

DeepSeek uses its official fixed HTTPS endpoints: `/responses` with JSON Schema
for writing/editorial work and `/v1/tools/search` for bounded live retrieval. The
writing model is configurable; the default is `deepseek-v4-pro`. Search is not
coupled to the writing model’s built-in tool support. DeepSeek receives only
its own credential. Narration is configured independently.
See [DeepSeek Responses](https://api-docs.deepseek.com/guides/responses_api/)
and [Search](https://api-docs.deepseek.com/guides/search/).

AI Service uses the existing `/_internal/chat` server contract on its private
loopback listener. In AI Service, issue a **local server** credential for app ID
`podcasts`, then save the address and credential in Connections. Default address
is `http://127.0.0.1:9042`; if Workspace uses another base port, use its service
port (normally base +52). An explicit `PODCASTS_AI_SERVICE_URL` and optional
`PODCASTS_AI_SERVICE_TOKEN` can also configure first use. Existing stored settings
remain authoritative. This app does not activate, install, replace or switch the
service’s model, and does not call Ollama directly or borrow My AI credentials.

The relay delegates the authenticated immutable podcast account through
X-AI-App/X-AI-User, follows the shared FIFO and cancellation/revocation controls,
and allows the service’s20-minute queue deadline. Local source prompts are
bounded; long editorial reviews are split by encoded size so the32,768-token
context ceiling cannot silently discard the script. JSON schema output is
validated, and truncated/empty responses fail closed. Context budgeting is
conservative, not an exact tokenizer or a guarantee of any model’s output quality.
See [Ollama structured output](https://docs.ollama.com/capabilities/structured-outputs).

<img src="web/public/podcasts.svg" width="68" alt="My Podcasts">

# My Podcasts

Your private technology podcast studio for One Workspace. Follow the people
shaping technology, choose your sources, and turn fresh reporting into an
original conversation between two AI narrators.

- Research current reporting, company announcements and accessible public posts.
- Write with OpenAI, DeepSeek or your existing local AI Service/Ollama queue.
- Choose a briefing, two-host conversation or narrative deep dive, 8–40 minutes.
- Listen with chapters, read the transcript and open the evidence behind claims.
- Generate on demand, daily or every Monday while your browser is closed.
- Resume interrupted production and export WAV audio, transcripts and sources.

Choose a writer in **Connections**: OpenAI, DeepSeek, or **AI Service** using its
local server address and an app credential issued for `podcasts`. AI Service uses
its active Ollama model and shared queue. Automatic research uses that provider’s
web search, or selected feeds/direct X posts for the local service. You can also
pair DeepSeek web search with the local writer.

Choose your interests, people and publications, then generate. Narration uses
My Voice on the same Workspace server by default. Install a supported stock
model in My Voice, then open **Settings → Podcasts narration** and issue a
credential. In My Podcasts, open **Settings → Narration**, paste the credential
and save the connection. The suggested address uses the managed My Voice port.

In **Create → Background production**, choose **Three times daily** and save
show settings. The schedule can be saved before the voice connection is ready.
The server resumes reviewed scripts and picks up due production when My Voice
is ready. Your Workspace server must stay running. OpenAI speech is optional.
Cloud research and writing use provider billing. Optional X API access improves
direct-post coverage.

Install My Podcasts through [One Workspace](https://github.com/oneworkspacestudio/one-workspace).
This evaluation release supports managed updates and private account storage.
Checks and remaining provider verification are in [acceptance](docs/ACCEPTANCE.md).
The application contains no sample episodes presented as real news.

See [production and research](docs/PRODUCTION.md), [sources](docs/DATA-SOURCES.md)
and [security and operations](docs/SECURITY.md). Developer builds use Go 1.27.3,
Node 24+, `npm ci --ignore-scripts`, `npm run build`, `go test ./...`, then
`go build -buildvcs=false -o MyPodcasts.exe ./cmd/my-podcasts`.

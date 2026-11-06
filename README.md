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

Choose your interests, people and publications, then generate. Narration uses a
separate OpenAI voice connection. Without it, your reviewed script is saved until
you connect voices and resume. Daily/weekly audio schedules require voices to be
configured. Your server must remain running. Cloud API calls use provider billing;
optional X API access improves direct-post coverage.

Install My Podcasts through [One Workspace](https://github.com/oneworkspacestudio/one-workspace).
This evaluation release supports managed updates and private account storage.
Checks and remaining provider verification are in [acceptance](docs/ACCEPTANCE.md).
The application contains no sample episodes presented as real news.

See [production and research](docs/PRODUCTION.md), [sources](docs/DATA-SOURCES.md)
and [security and operations](docs/SECURITY.md). Developer builds use Go 1.27.3,
Node 24+, `npm ci --ignore-scripts`, `npm run build`, `go test ./...`, then
`go build -buildvcs=false -o MyPodcasts.exe ./cmd/my-podcasts`.

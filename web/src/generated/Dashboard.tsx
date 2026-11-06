// @ts-nocheck -- AUI compiler data bindings. Edit screens/Dashboard.aui instead.
import { Alert, Badge, Button, Card, Checkbox, Divider, Grid, Heading, Icon, Image, Input, Link, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"
import { AudioPlayer } from "@/components/ui"

export function Dashboard({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Row variant="topbar">
        <Row gap="sm">
          <Image src="./podcasts.svg" alt="" width={38} height={38} />
          <Text variant="brand">My Podcasts</Text>
          <Badge tone="blue">STUDIO</Badge>
        </Row>
        <Row gap="sm">
          <Text variant="header-note">A little more signal. A lot less noise.</Text>
          <Link href="/" variant="apps">Apps</Link>
        </Row>
      </Row>
      {data.signedOut ? (
        <Section variant="login">
          {data.error && (
            <Alert tone="warning" role="alert">{data.error}</Alert>
          )}
          <Icon name="mic" variant="login-icon" />
          <Text variant="eyebrow">YOUR PRIVATE PODCAST STUDIO</Text>
          <Heading level={1}>The stories worth listening to.</Heading>
          <Text>Turn the latest in technology into thoughtful conversations, made just for you.</Text>
          {data.localLogin ? (
            <>
              <Input name="email" value={data.email} type="email" label="Email" autoComplete="username" />
              <Input name="password" value={data.password} type="password" label="Password" autoComplete="current-password" />
              <Button intent="login" variant="primary">{data.loginLabel}</Button>
            </>
          ) : (
            <Link href={data.loginURL} variant="primary">Sign in with MyAccounts</Link>
          )}
        </Section>
      ) : (
        <Row variant="workspace">
          <Section variant="sidebar">
            <Text variant="eyebrow">YOUR LISTENING SPACE</Text>
            <Button intent="library" variant={data.libraryVariant} icon="headphones">Episode library</Button>
            <Button intent="studio" variant={data.studioVariant} icon="sparkles">Create an episode</Button>
            <Button intent="queue" variant={data.queueVariant} icon="activity">Production queue</Button>
            <Button intent="sources" variant={data.sourcesVariant} icon="radio">People & sources</Button>
            <Spacer />
            <Card variant="sidebar-note">
              <Row gap="sm">
                <Icon name="clock" />
                <Text variant="eyebrow">MADE WHILE YOU’RE AWAY</Text>
              </Row>
              <Text>{data.scheduleDescription}</Text>
              <Button intent="studio" variant="text">Manage your show</Button>
            </Card>
            <Button intent="settings" variant={data.settingsVariant} icon="settings">Connections</Button>
            <Button intent="logout" variant="nav" icon="logout">Sign out</Button>
          </Section>
          <Stack variant="main">
            {data.error && (
              <Alert tone="warning" role="alert">{data.error}</Alert>
            )}
            {data.toast && (
              <Alert tone="success" role="status">{data.toast}</Alert>
            )}
            {data.loading ? (
              <Text variant="muted">Opening your studio…</Text>
            ) : (
              <>
                {data.showLibrary && (
                  <>
                    <Row variant="page-heading">
                      <Stack gap="sm">
                        <Text variant="eyebrow">YOUR DAILY SIGNAL</Text>
                        <Heading level={1}>Good stories. Great company.</Heading>
                        <Text variant="muted">Fresh perspectives from the people shaping technology.</Text>
                      </Stack>
                      <Button intent="studio" variant="primary" icon="plus">Create an episode</Button>
                    </Row>
                    <Card variant="hero">
                      <Stack variant="hero-copy">
                        <Badge tone="blue">YOUR OWN TECH PODCAST</Badge>
                        <Heading level={2}>Stay curious. We’ll do the homework.</Heading>
                        <Text>The headlines, the original posts, the bigger picture. Researched, questioned, and brought to life by two thoughtful voices.</Text>
                        <Row gap="md">
                          <Button intent="studio" variant="dark" icon="sparkles">Make something worth hearing</Button>
                          <Text variant="muted">Made for your ears. Backed by sources.</Text>
                        </Row>
                      </Stack>
                      <Section variant="hero-art">
                        <Row variant="art-label">
                          <Text>THE TECH SIGNAL</Text>
                          <Icon name="radio" />
                        </Row>
                        <Section variant="waveform">
                          <Icon name="waveform" />
                          <Icon name="waveform" />
                          <Icon name="waveform" />
                        </Section>
                        <Row variant="art-bottom">
                          <Text>ALEX + JAMIE</Text>
                          <Text>ORIGINAL AI VOICES</Text>
                        </Row>
                      </Section>
                    </Card>
                    <Row variant="stats">
                      <Card variant="stat">
                        <Icon name="headphones" />
                        <Stack gap="sm">
                          <Text variant="stat-number">{data.readyCount}</Text>
                          <Text variant="muted">Ready for your next walk</Text>
                        </Stack>
                      </Card>
                      <Card variant="stat">
                        <Icon name="radio" />
                        <Stack gap="sm">
                          <Text variant="stat-number">{data.sourceCount}</Text>
                          <Text variant="muted">People & publications</Text>
                        </Stack>
                      </Card>
                      <Card variant="stat">
                        <Icon name="clock" />
                        <Stack gap="sm">
                          <Text variant="stat-number">{data.scheduleLabel}</Text>
                          <Text variant="muted">{data.nextRunLabel}</Text>
                        </Stack>
                      </Card>
                    </Row>
                    <Row variant="section-heading">
                      <Heading level={2}>Your episodes</Heading>
                      <Text variant="muted">{data.libraryDescription}</Text>
                    </Row>
                    {data.emptyLibrary ? (
                      <Card variant="empty">
                        <Icon name="headphones" variant="large" />
                        <Heading level={3}>Your next favourite listen starts here.</Heading>
                        <Text variant="muted">Choose your interests, pick your voices, and let your studio take it from there.</Text>
                        <Button intent="studio" variant="quiet" icon="arrow-right">Create your first episode</Button>
                      </Card>
                    ) : (
                      <Stack gap="md">
                        {data.episodes.map((item, i) => (
                          <Card key={i} variant="episode">
                            <Section variant="episode-cover">
                              <Icon name="waveform" />
                              <Text>MY PODCASTS</Text>
                            </Section>
                            <Stack variant="episode-copy" gap="sm">
                              <Row gap="sm">
                                <Badge tone={item.tone}>{item.statusLabel}</Badge>
                                <Text variant="eyebrow">{item.date}</Text>
                              </Row>
                              <Heading level={3}>{item.title}</Heading>
                              <Text variant="muted">{item.summary}</Text>
                              <Text variant="small">{item.meta}</Text>
                            </Stack>
                            <Button intent="open" payload={item.id} variant="play" icon={item.icon}>{item.action}</Button>
                          </Card>
                        ))}
                      </Stack>
                    )}
                  </>
                )}
                {data.showStudio && (
                  <>
                    <Row variant="page-heading">
                      <Stack gap="sm">
                        <Text variant="eyebrow">THE CREATOR’S DESK</Text>
                        <Heading level={1}>Make it your kind of listen.</Heading>
                        <Text variant="muted">Set the direction. Your studio takes care of the research and production.</Text>
                      </Stack>
                    </Row>
                    <Row variant="studio-layout">
                      <Stack gap="lg">
                        <Card variant="form">
                          <Row gap="sm">
                            <Badge tone="blue">01</Badge>
                            <Heading level={2}>Give your show a point of view</Heading>
                          </Row>
                          <Input name="showName" value={data.showName} label="Show name" maxLength={90} />
                          <Input name="topics" value={data.topics} variant="multiline" label="What are you curious about?" maxLength={1200} />
                          <Text variant="muted">Follow people and choose publications in People & sources.</Text>
                          <Row variant="form-row">
                            <Select name="format" value={data.format} options={data.formatOptions} label="Episode format" />
                            <Select name="minutes" value={data.minutes} options={data.minuteOptions} label="Target length" />
                            <Select name="lookbackHours" value={data.lookbackHours} options={data.lookbackOptions} label="Research window" />
                          </Row>
                        </Card>
                        <Card variant="form">
                          <Row gap="sm">
                            <Badge tone="blue">02</Badge>
                            <Heading level={2}>Two voices. One good conversation.</Heading>
                          </Row>
                          <Row variant="narrators">
                            <Card variant="narrator">
                              <Badge tone="blue">ALEX</Badge>
                              <Heading level={3}>The curious anchor</Heading>
                              <Text variant="muted">Makes the complicated feel clear. Connects the dots and keeps things moving.</Text>
                              <Select name="voiceA" value={data.voiceA} options={data.voiceOptions} label="Alex’s voice" />
                            </Card>
                            <Card variant="narrator">
                              <Badge tone="mint">JAMIE</Badge>
                              <Heading level={3}>The thoughtful challenger</Heading>
                              <Text variant="muted">Asks the next question. Tests the big claims and brings another perspective.</Text>
                              <Select name="voiceB" value={data.voiceB} options={data.voiceOptions} label="Jamie’s voice" />
                            </Card>
                          </Row>
                          <Text variant="small">Original AI-generated narrators. Every episode includes a spoken disclosure.</Text>
                        </Card>
                        <Card variant="form">
                          <Row gap="sm">
                            <Badge tone="blue">03</Badge>
                            <Heading level={2}>A new episode, on your time</Heading>
                          </Row>
                          <Row variant="form-row">
                            <Select name="schedule" value={data.schedule} options={data.scheduleOptions} label="Background production" />
                            <Input name="at" value={data.at} type="time" label="Production starts at (UTC)" />
                          </Row>
                          <Text variant="muted">Your server keeps working when this tab is closed. Weekly episodes start on Mondays. Catch-up runs once after downtime.</Text>
                          <Text variant="small">Up to 3 productions per 24 hours and 50 saved episodes per account. AI research and speech use your connected API billing.</Text>
                        </Card>
                      </Stack>
                      <Stack variant="studio-aside" gap="md">
                        <Card variant="show-preview">
                          <Text variant="eyebrow">YOUR SHOW</Text>
                          <Icon name="waveform" variant="preview-wave" />
                          <Heading level={2}>{data.showName}</Heading>
                          <Text>{data.previewDescription}</Text>
                          <Divider />
                          <Text>Alex & Jamie · Original AI voices</Text>
                          <Text variant="small">A fresh perspective, with the receipts.</Text>
                        </Card>
                        <Card variant="form">
                          <Heading level={3}>From headline to headphones</Heading>
                          <Text variant="small">01  Read the reporting & original posts</Text>
                          <Text variant="small">02  Build a source-linked conversation</Text>
                          <Text variant="small">03  Check claims, attribution & flow</Text>
                          <Text variant="small">04  Record, balance & chapter the audio</Text>
                          <Button intent="save" variant="quiet" disabled={data.busy}>Save show settings</Button>
                          <Button intent="generate" variant="primary" icon="sparkles" disabled={data.busy}>Generate an episode</Button>
                          {data.needsConnection && (
                            <Text variant="small">Configure your selected writing and research providers in Connections.</Text>
                          )}
                          {data.speechMissing && (
                            <Text variant="small">Scripts can finish without voices. Connect OpenAI speech and resume when you’re ready to record.</Text>
                          )}
                        </Card>
                      </Stack>
                    </Row>
                  </>
                )}
                {data.showQueue && (
                  <>
                    <Row variant="page-heading">
                      <Stack gap="sm">
                        <Text variant="eyebrow">QUIETLY WORKING IN THE BACKGROUND</Text>
                        <Heading level={1}>A little studio of your own.</Heading>
                        <Text variant="muted">Close the tab. Carry on with your day. Your server takes it from here.</Text>
                      </Stack>
                      <Button intent="refresh" variant="quiet" icon="refresh">Refresh</Button>
                    </Row>
                    {data.emptyQueue ? (
                      <Card variant="empty">
                        <Icon name="activity" variant="large" />
                        <Heading level={3}>The studio is ready when you are.</Heading>
                        <Text variant="muted">New productions and interrupted jobs appear here.</Text>
                        <Button intent="studio" variant="primary">Create an episode</Button>
                      </Card>
                    ) : (
                      <Stack gap="md">
                        {data.jobs.map((item, i) => (
                          <Card key={i} variant="job">
                            <Row variant="section-heading">
                              <Heading level={3}>{item.title}</Heading>
                              <Badge tone={item.tone}>{item.statusLabel}</Badge>
                            </Row>
                            <Text>{item.stage}</Text>
                            <Text variant="muted">{item.progress}% complete · {item.date}</Text>
                            {item.error && (
                              <Alert tone="warning">{item.error}</Alert>
                            )}
                            <Row gap="sm">
                              <Button intent="open" payload={item.id} variant="quiet">View production</Button>
                              {item.running && (
                                <Button intent="cancel" payload={item.id} variant="text">Cancel production</Button>
                              )}
                              {item.retryable && (
                                <Button intent="retry" payload={item.id} variant="primary">Resume / regenerate</Button>
                              )}
                            </Row>
                          </Card>
                        ))}
                      </Stack>
                    )}
                  </>
                )}
                {data.showSources && (
                  <>
                    <Row variant="page-heading">
                      <Stack gap="sm">
                        <Text variant="eyebrow">GOOD CONVERSATIONS START WITH GOOD SOURCES</Text>
                        <Heading level={1}>Follow the people. Read the room.</Heading>
                        <Text variant="muted">Original voices, independent reporting, and primary announcements.</Text>
                      </Stack>
                      <Button intent="save" variant="primary" disabled={data.busy}>Save sources</Button>
                    </Row>
                    <Card variant="source-note">
                      <Icon name="shield" />
                      <Stack gap="sm">
                        <Heading level={3}>A post is a perspective. Reporting adds context.</Heading>
                        <Text>{data.socialDescription}</Text>
                        <Text variant="small">{data.researchCoverage}</Text>
                      </Stack>
                    </Card>
                    <Row variant="section-heading">
                      <Heading level={2}>The people shaping technology</Heading>
                      <Text variant="muted">{data.peopleCount} selected</Text>
                    </Row>
                    <Grid variant="people">
                      {data.people.map((item, i) => (
                        <Card key={i} variant="person">
                          <Row gap="md">
                            <Badge tone="blue">{item.initials}</Badge>
                            <Stack gap="sm">
                              <Heading level={3}>{item.name}</Heading>
                              <Text variant="muted">{item.company}</Text>
                            </Stack>
                          </Row>
                          <Checkbox name={item.field} checked={item.selected} label={item.label} />
                          <Link href={item.url} target="_blank" variant="text">View original profile ↗</Link>
                        </Card>
                      ))}
                    </Grid>
                    <Row variant="section-heading">
                      <Heading level={2}>The publications behind the perspective</Heading>
                      <Text variant="muted">Choose your regular reading list</Text>
                    </Row>
                    <Grid variant="publications">
                      {data.feeds.map((item, i) => (
                        <Card key={i} variant="publication">
                          <Heading level={3}>{item.name}</Heading>
                          <Badge tone={item.tone}>{item.kind}</Badge>
                          <Checkbox name={item.field} checked={item.selected} label={item.label} />
                        </Card>
                      ))}
                    </Grid>
                  </>
                )}
                {data.showSettings && (
                  <>
                    <Row variant="page-heading">
                      <Stack gap="sm">
                        <Text variant="eyebrow">THE CONNECTIONS BEHIND YOUR STUDIO</Text>
                        <Heading level={1}>Ready to make some noise?</Heading>
                        <Text variant="muted">Connect research and speech. Keys stay on your private server.</Text>
                      </Stack>
                    </Row>
                    <Card variant="form">
                      <Row variant="section-heading">
                        <Heading level={2}>Research & writing</Heading>
                        <Badge tone={data.connectionTone}>{data.connectionLabel}</Badge>
                      </Row>
                      <Text variant="muted">Choose who researches, writes and reviews your show. Voices have their own connection below.</Text>
                      {data.isOwner ? (
                        <>
                          <Row variant="form-row">
                            <Select name="textProvider" value={data.textProvider} options={data.writerOptions} label="Writing & editorial provider" />
                            <Select name="researchProvider" value={data.researchProvider} options={data.researchOptions} label="Research source" />
                          </Row>
                          <Text variant="small">Automatic uses OpenAI web research, DeepSeek search, or selected feeds and direct X posts for AI Service. Local inference uses the service’s active model and shared queue.</Text>
                        </>
                      ) : (
                        <Text>Your Workspace owner manages these connections.</Text>
                      )}
                    </Card>
                    {data.isOwner && (
                      <>
                        <Card variant="form">
                          <Row variant="section-heading">
                            <Heading level={2}>DeepSeek</Heading>
                            <Badge tone="blue">{data.deepSeekLabel}</Badge>
                          </Row>
                          <Text variant="muted">Current web reporting and public posts through DeepSeek search, with structured scripts and an editorial pass. You can also use DeepSeek search with your local writer.</Text>
                          <Input name="deepSeekKey" value={data.deepSeekKey} type="password" label="DeepSeek API key" autoComplete="off" placeholder="Leave blank to keep the saved key" />
                          <Input name="deepSeekModel" value={data.deepSeekModel} label="DeepSeek writing model" />
                          <Button intent="disconnectDeepSeek" variant="text">Disconnect DeepSeek</Button>
                        </Card>
                        <Card variant="form">
                          <Row variant="section-heading">
                            <Heading level={2}>AI Service · Your local Ollama</Heading>
                            <Badge tone="blue">{data.serviceLabel}</Badge>
                          </Row>
                          <Text variant="muted">Uses your existing AI Service and its selected Ollama model. It respects the shared inference queue and does not start or change models.</Text>
                          <Input name="serviceURL" value={data.serviceURL} label="AI Service URL on this server" placeholder="http://127.0.0.1:9042" />
                          <Input name="serviceToken" value={data.serviceToken} type="password" label="AI Service app credential" autoComplete="off" placeholder="Leave blank to keep the saved credential" />
                          <Text variant="small">In AI Service, issue a local server app credential with app ID podcasts. Paste it here. This is a server connection; the URL is relative to the computer running My Podcasts.</Text>
                          <Text variant="small">The service must be enabled with an active model. Automatic research uses the selected dated feeds and connected X posts; choose DeepSeek search for wider web coverage.</Text>
                          <Button intent="disconnectService" variant="text">Disconnect AI Service</Button>
                        </Card>
                        <Card variant="form">
                          <Row variant="section-heading">
                            <Heading level={2}>OpenAI · Voices & optional writing</Heading>
                            <Badge tone="blue">{data.speechLabel}</Badge>
                          </Row>
                          <Text variant="muted">OpenAI provides the two narrator voices. Its key is also used when you select OpenAI for writing or web research. DeepSeek and AI Service handle text; they do not provide speech through these connections.</Text>
                          <Input name="key" value={data.key} type="password" label="OpenAI API key" autoComplete="off" placeholder="Leave blank to keep the saved key" />
                          <Row variant="form-row">
                            <Input name="model" value={data.model} label="OpenAI research & writing model" />
                            <Input name="ttsModel" value={data.ttsModel} label="Speech model" />
                          </Row>
                          <Text variant="small">Without a voice key, manual production saves a reviewed script and waits for narration. Audio schedules require voices to be configured. Cloud API billing is separate from subscriptions.</Text>
                          <Button intent="disconnect" variant="text">Disconnect OpenAI</Button>
                        </Card>
                        <Card variant="form">
                          <Heading level={2}>Direct X timelines · Optional</Heading>
                          <Input name="xToken" value={data.xToken} type="password" label="X API bearer token" autoComplete="off" placeholder="Leave blank to keep the saved token" />
                          <Text variant="small">Direct timeline access requires an X developer token. Feed-only mode does not check indexed posts; web search coverage remains incomplete.</Text>
                          <Button intent="disconnectX" variant="text">Disconnect X</Button>
                        </Card>
                        <Card variant="form">
                          <Row gap="sm">
                            <Button intent="saveProviders" variant="primary" disabled={data.busy}>Save connections</Button>
                            <Button intent="testProvider" variant="quiet" disabled={data.busy}>Save & check writing connection</Button>
                          </Row>
                          <Text variant="small">The connection check sends one short request to the selected writer. Cloud checks use API billing. Disconnecting a provider stops active production so it can be resumed with saved progress.</Text>
                        </Card>
                      </>
                    )}
                    <Card variant="source-note">
                      <Icon name="lock" />
                      <Stack gap="sm">
                        <Heading level={3}>Your show. Your library.</Heading>
                        <Text>Episodes, scripts and settings are private to your Workspace account. Audio exports are uncompressed WAV. Background production needs your server to remain running.</Text>
                        <Text variant="small">The editorial pass checks consistency with the research; it is not a guarantee that every source is correct.</Text>
                      </Stack>
                    </Card>
                    <Button intent="logout" variant="quiet" icon="logout">Sign out</Button>
                  </>
                )}
                {data.showDetail && (
                  <>
                    <Button intent="library" variant="text" icon="arrow-left">Back to your library</Button>
                    <Row variant="page-heading">
                      <Stack gap="sm">
                        <Badge tone={data.detailTone}>{data.detailStatus}</Badge>
                        <Heading level={1}>{data.detailTitle}</Heading>
                        <Text variant="muted">{data.detailSummary}</Text>
                        <Text variant="small">{data.detailMeta}</Text>
                      </Stack>
                    </Row>
                    {data.detailReady ? (
                      <Card variant="player">
                        <Text variant="eyebrow">SETTLE IN. PRESS PLAY.</Text>
                        <AudioPlayer src={data.audioURL} seek={data.seek} label={data.detailTitle} />
                        <Row gap="sm">
                          <Link href={data.downloadURL} variant="quiet">Download WAV</Link>
                          <Link href={data.transcriptURL} variant="quiet">Export transcript</Link>
                          <Link href={data.sourcesURL} variant="quiet">Export sources</Link>
                        </Row>
                      </Card>
                    ) : (
                      <Card variant="form">
                        <Heading level={3}>{data.detailStage}</Heading>
                        <Text>{data.detailProgress}% complete. Progress is saved on your server.</Text>
                        {data.detailError && (
                          <Alert tone="warning">{data.detailError}</Alert>
                        )}
                        {data.detailRunning && (
                          <Button intent="cancel" payload={data.detailId} variant="quiet">Cancel production</Button>
                        )}
                        {data.detailRetryable && (
                          <Button intent="retry" payload={data.detailId} variant="primary">Resume / regenerate</Button>
                        )}
                        {data.hasScript && (
                          <Row gap="sm">
                            <Link href={data.transcriptURL} variant="quiet">Export transcript</Link>
                            <Link href={data.sourcesURL} variant="quiet">Export sources</Link>
                          </Row>
                        )}
                      </Card>
                    )}
                    {data.hasWarnings && (
                      <Card variant="warnings">
                        <Heading level={3}>Research coverage notes</Heading>
                        {data.warnings.map((item, i) => (
                          <Text key={i} variant="small">{item.text}</Text>
                        ))}
                      </Card>
                    )}
                    {data.hasIssues && (
                      <Card variant="warnings">
                        <Heading level={3}>The editor’s notes</Heading>
                        {data.issues.map((item, i) => (
                          <Text key={i}>{item.text}</Text>
                        ))}
                      </Card>
                    )}
                    <Row variant="detail-layout">
                      <Stack gap="lg">
                        <Row variant="section-heading">
                          <Heading level={2}>The conversation</Heading>
                          <Badge tone="blue">AI-GENERATED NARRATORS</Badge>
                        </Row>
                        {data.chapters.map((item, i) => (
                          <Card key={i} variant="transcript">
                            <Row variant="section-heading">
                              <Heading level={3}>{item.title}</Heading>
                              {data.detailReady && (
                                <Button intent="seek" payload={item.start} variant="text">{item.time} ↗</Button>
                              )}
                            </Row>
                            <Stack gap="md">
                              {item.turns.map((item, i) => (
                                <Stack key={i} gap="sm">
                                  <Text variant="speaker">{item.speaker}</Text>
                                  <Text>{item.text}</Text>
                                  <Text variant="small">{item.citations}</Text>
                                </Stack>
                              ))}
                            </Stack>
                          </Card>
                        ))}
                      </Stack>
                      <Stack gap="md">
                        <Heading level={2}>The receipts</Heading>
                        <Text variant="small">Research cutoff · {data.cutoff}</Text>
                        {data.detailSources.map((item, i) => (
                          <Card key={i} variant="receipt">
                            <Row gap="sm">
                              <Badge tone="blue">{item.id}</Badge>
                              <Text variant="eyebrow">{item.publisher}</Text>
                            </Row>
                            <Link href={item.url} target="_blank">{item.title} ↗</Link>
                            <Text variant="small">{item.dateLabel}</Text>
                          </Card>
                        ))}
                      </Stack>
                    </Row>
                    {data.detailDeletable && (
                      <Button intent="delete" payload={data.detailId} variant="danger" icon="trash">Delete episode</Button>
                    )}
                  </>
                )}
              </>
            )}
          </Stack>
          <Row variant="mobile-nav">
            <Button intent="library" variant={data.libraryVariant} icon="headphones">Listen</Button>
            <Button intent="studio" variant={data.studioVariant} icon="sparkles">Create</Button>
            <Button intent="queue" variant={data.queueVariant} icon="activity">Queue</Button>
            <Button intent="sources" variant={data.sourcesVariant} icon="radio">Sources</Button>
            <Button intent="settings" variant={data.settingsVariant} icon="settings">Settings</Button>
          </Row>
        </Row>
      )}
    </Page>
  )
}

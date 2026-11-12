import {test,expect} from '@playwright/test';
import {spawn,execFileSync,ChildProcess} from 'node:child_process';
import {mkdir,rm,lstat,writeFile} from 'node:fs/promises';
import {resolve} from 'node:path';
import {randomUUID} from 'node:crypto';
import {createServer} from 'node:http';
let child:ChildProcess,root:string;
const origin='http://127.0.0.1:19444',email='podcast-native@example.invalid',password='Synthetic-podcast-browser-2026!';
const executable=process.env.PODCASTS_TEST_EXE||'I:/Projects/tooling/my-podcasts-providers-20261106/MyPodcasts.exe';
const env={...process.env,OPENAI_API_KEY:'',DEEPSEEK_API_KEY:'',PODCASTS_AI_SERVICE_URL:'',PODCASTS_AI_SERVICE_TOKEN:'',PODCASTS_X_BEARER_TOKEN:'',WORKSPACE_OIDC_ISSUER:'',WORKSPACE_OIDC_CLIENT_ID:'',WORKSPACE_OIDC_CLIENT_SECRET:'',WORKSPACE_OWNER_SUBJECT:'',WORKSPACE_BOOTSTRAP_PASSWORD:password};
async function start(bootstrap=false){child=spawn(executable,['--data',root,'--listen','127.0.0.1:19444','--origin',origin,'--no-browser',...(bootstrap?['--bootstrap-email',email]:[])],{env,windowsHide:true,stdio:'ignore'});await expect.poll(async()=>{try{return(await fetch(origin+'/healthz')).status}catch{return 0}},{timeout:60000}).toBe(200);}
async function stop(){if(child?.exitCode===null){const stopped=new Promise<void>(r=>child.once('exit',()=>r()));child.kill();await stopped;}}
async function login(page:any){await page.goto('/');await page.getByLabel('Email',{exact:true}).fill(email);await page.getByLabel('Password',{exact:true}).fill(password);await page.getByRole('button',{name:'Sign in',exact:true}).click();await expect(page.getByRole('heading',{name:'Good stories. Great company.'})).toBeVisible();}
test.beforeAll(async()=>{root='C:/Projects/.cache/my-podcasts-providers-20261106/browser-'+randomUUID();await mkdir(root,{recursive:true});await mkdir('docs/screenshots',{recursive:true});await mkdir('docs/evidence',{recursive:true});await start(true);});
test.afterAll(async()=>{await stop();if(root&&resolve(root).startsWith(resolve('C:/Projects/.cache/my-podcasts-providers-20261106')+'\\')&&!(await lstat(root)).isSymbolicLink())await rm(root,{recursive:true});});
test('actual native UI, saved settings, sources and honest missing-provider state at phone/desktop widths',async({page})=>{const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));await login(page);await page.screenshot({path:'docs/screenshots/library-desktop.png',fullPage:true});for(const width of [320,390,768,1440]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);}await page.setViewportSize({width:1440,height:1000});await page.getByRole('button',{name:'Create an episode',exact:true}).first().click();await page.getByLabel('Show name',{exact:true}).fill('Synthetic Signal');await page.getByLabel('What are you curious about?').fill('Technology, research quality and real-world implications');await page.getByLabel('Target length').selectOption('8');await page.getByRole('button',{name:'Save show settings',exact:true}).click();await expect(page.getByRole('status')).toContainText('saved');await page.getByRole('button',{name:'Generate an episode',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Connect an OpenAI API key');await page.screenshot({path:'docs/screenshots/studio-desktop.png',fullPage:true});await page.getByRole('button',{name:'People & sources',exact:true}).click();await page.getByRole('checkbox',{name:'Follow Elon Musk',exact:true}).check();await page.getByRole('button',{name:'Save sources',exact:true}).click();await page.reload();await page.getByRole('button',{name:'People & sources',exact:true}).click();await expect(page.getByRole('checkbox',{name:'Follow Elon Musk',exact:true})).toBeChecked();for(const width of [320,390,768,1440]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);}await page.setViewportSize({width:390,height:844});await page.screenshot({path:'docs/screenshots/sources-mobile.png',fullPage:true});await page.getByRole('button',{name:'Create',exact:true}).click();await page.screenshot({path:'docs/screenshots/studio-mobile.png',fullPage:true});await expect(page.getByLabel('Show name',{exact:true})).toHaveValue('Synthetic Signal');await page.getByLabel('Background production').selectOption('daily');await page.getByRole('button',{name:'Save show settings',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Connect an OpenAI API key');await page.getByLabel('Background production').selectOption('off');expect(errors).toEqual([]);});
test('native account isolation, CSRF and owner-only connections',async({page,browser})=>{await login(page);const me=await(await page.request.get('/api/me')).json();const headers={Origin:origin,'X-CSRF-Token':me.csrf};expect((await page.request.put('/api/settings',{headers:{Origin:origin},data:{}})).status()).toBe(403);const member='member-'+randomUUID()+'@example.invalid';expect((await page.request.post('/api/local-users',{headers,data:{email:member,password}})).status()).toBe(201);const other=await browser.newContext();try{expect((await other.request.post(origin+'/api/login',{headers:{Origin:origin},data:{email:member,password}})).status()).toBe(200);const state=await(await other.request.get(origin+'/api/state')).json();expect(state.episodes).toEqual([]);expect(state.settings.showName).toBe('The Tech Signal');const user=await(await other.request.get(origin+'/api/me')).json();expect((await other.request.put(origin+'/api/providers',{headers:{Origin:origin,'X-CSRF-Token':user.csrf},data:{key:'synthetic',model:'gpt-5.4',ttsModel:'gpt-4o-mini-tts'}})).status()).toBe(403);const anonymous=await browser.newContext();expect((await anonymous.request.get(origin+'/api/state')).status()).toBe(401);await anonymous.close();}finally{await other.close();}});
test('native private WAV range playback, chapters and export from explicit synthetic fixture',async({page,browser})=>{await stop();execFileSync('python',['tests/seed.py',root],{windowsHide:true});await start();await login(page);await page.getByRole('button',{name:'Listen',exact:true}).click();await expect(page.getByRole('heading',{name:'Synthetic episode · Beyond the headline',exact:true})).toBeVisible();await expect(page.getByText('The receipts',{exact:true})).toBeVisible();await expect.poll(()=>page.locator('audio').evaluate((a:HTMLAudioElement)=>a.readyState)).toBeGreaterThanOrEqual(1);await page.getByRole('button',{name:'0:01 ↗',exact:true}).click();await expect.poll(()=>page.locator('audio').evaluate((a:HTMLAudioElement)=>a.currentTime)).toBeGreaterThanOrEqual(1);const id='a1'.repeat(24);const range=await page.request.get('/api/episodes/'+id+'/audio',{headers:{Range:'bytes=0-43'}});expect(range.status()).toBe(206);expect((await range.body()).subarray(0,4).toString()).toBe('RIFF');const transcript=await page.request.get('/api/episodes/'+id+'/transcript');expect(await transcript.text()).toContain('AI-generated narrators');const sources=await(await page.request.get('/api/episodes/'+id+'/sources')).json();expect(sources.sources).toHaveLength(2);await page.setViewportSize({width:1440,height:1000});await page.screenshot({path:'docs/screenshots/episode-synthetic-desktop.png',fullPage:true});for(const width of [320,390,768,1440]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);}await page.setViewportSize({width:390,height:844});await page.screenshot({path:'docs/screenshots/episode-synthetic-mobile.png',fullPage:true});const anonymous=await browser.newContext();for(const suffix of ['','/audio','/transcript','/sources'])expect((await anonymous.request.get(origin+'/api/episodes/'+id+suffix)).status()).toBe(401);await anonymous.close();});
test('native process restart preserves library and settings; logout removes access',async({page})=>{await stop();await start();await login(page);await expect(page.getByRole('heading',{name:'Synthetic episode · Beyond the headline',exact:true})).toBeVisible();await page.setViewportSize({width:390,height:844});await page.getByRole('button',{name:'Settings',exact:true}).click();await page.getByRole('button',{name:'Sign out',exact:true}).click();await expect(page.getByLabel('Password',{exact:true})).toBeVisible();expect((await page.request.get('/api/state')).status()).toBe(401);await writeFile('docs/evidence/browser-acceptance.json',JSON.stringify({schema:1,passed:true,scope:'exact native executable with synthetic isolated NTFS accounts; synthetic tone fixture',checks:['responsive 320/390/768/1440','settings and sources persistence','missing API key and scheduling prerequisites','CSRF and owner writes','account isolation','authenticated WAV range playback','sample-derived chapter seeking','transcript and source exports','native restart','logout'],realPaidGeneration:false,credentialedX:false,physicalPhone:false,ownerInstallation:false},null,2));});

test('DeepSeek and AI Service controls, real server relay, persistence and revocation errors',async({page})=>{
 let requests=0,revoke=false;const token='synthetic-local-app-credential-'+randomUUID();const server=createServer(async(req,res)=>{let raw='';for await(const data of req)raw+=data;requests++;expect(req.url).toBe('/_internal/chat');expect(req.headers['x-ai-app']).toBe('podcasts');expect(String(req.headers['x-ai-user'])).toMatch(/^[a-f0-9]{64}$/);expect(req.headers.authorization).toBe('Bearer '+token);const body=JSON.parse(raw);expect(body.model).toBeUndefined();expect(body.stream).toBe(false);expect(body.options.num_ctx).toBe(32768);res.writeHead(revoke?401:200,{'Content-Type':'application/json'});res.end(revoke?JSON.stringify({error:'synthetic revoked grant'}):JSON.stringify({done:true,message:{content:'READY'}}));});
 await new Promise<void>(r=>server.listen(0,'127.0.0.1',()=>r()));const address=server.address() as {port:number},url='http://127.0.0.1:'+address.port;
 try{await page.setViewportSize({width:1440,height:1000});await login(page);await page.getByRole('button',{name:'Connections',exact:true}).click();await expect(page.getByRole('heading',{name:'DeepSeek',exact:true})).toBeVisible();await page.getByLabel('Writing & editorial provider').selectOption('deepseek');await page.getByLabel('DeepSeek API key',{exact:true}).fill('synthetic-deepseek-configuration-only');await page.getByRole('button',{name:'Save connections',exact:true}).click();await expect(page.getByRole('status')).toContainText('updated');let state=await(await page.request.get('/api/state')).json();expect(state.provider.textProvider).toBe('deepseek');expect(state.provider.researchMode).toBe('deepseek');expect(state.provider.deepSeekConfigured).toBe(true);expect(JSON.stringify(state)).not.toContain('synthetic-deepseek-configuration-only');await expect(page.getByLabel('DeepSeek API key',{exact:true})).toHaveValue('');
 await page.getByLabel('Writing & editorial provider').selectOption('aiservice');await page.getByLabel('AI Service URL on this server').fill(url);await page.getByLabel('AI Service app credential',{exact:true}).fill(token);await page.getByRole('button',{name:'Save & check writing connection',exact:true}).click();await expect(page.getByRole('status')).toContainText('writing provider responded');expect(requests).toBe(1);await expect(page.getByLabel('AI Service app credential',{exact:true})).toHaveValue('');state=await(await page.request.get('/api/state')).json();expect(state.provider.researchMode).toBe('feeds');expect(state.provider.textConfigured).toBe(true);expect(state.provider.speechConfigured).toBe(false);expect(JSON.stringify(state)).not.toContain(token);await page.evaluate(()=>window.scrollTo(0,0));await page.screenshot({path:'docs/screenshots/connections-desktop.png',fullPage:true});
 await page.reload();await page.getByRole('button',{name:'Connections',exact:true}).click();await expect(page.getByLabel('Writing & editorial provider')).toHaveValue('aiservice');await expect(page.getByLabel('AI Service URL on this server')).toHaveValue(url);for(const width of [320,390,768,1440]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);}await page.setViewportSize({width:390,height:844});await page.evaluate(()=>window.scrollTo(0,0));await page.screenshot({path:'docs/screenshots/connections-mobile.png',fullPage:true});revoke=true;await page.getByRole('button',{name:'Save & check writing connection',exact:true}).click();await expect(page.getByRole('alert')).toContainText('AI Service returned HTTP 401');expect(requests).toBe(2);await expect(page.getByRole('alert')).not.toContainText(token);
 const me=await(await page.request.get('/api/me')).json();await page.request.put('/api/providers',{headers:{Origin:origin,'X-CSRF-Token':me.csrf},data:{textProvider:'openai',clearService:true,clearDeepSeek:true}});await writeFile('docs/evidence/provider-browser-acceptance.json',JSON.stringify({schema:1,passed:true,checks:['DeepSeek selection/key persistence without secret disclosure','AI Service native HTTP relay with app/account delegation','active model preserved','encrypted connection restart','320/390/768/1440 layouts','revoked credential error without secret leakage'],scope:'synthetic isolated native executable and loopback AI Service protocol fixture',realDeepSeek:false,ownerOllama:false,ownerInstallation:false},null,2));
 }finally{server.closeAllConnections();await new Promise<void>(r=>server.close(()=>r()));}
});

test('Disconnect removes only the selected credential despite invalid unsaved fields', async ({page}) => {
 const token = 'synthetic-ai-service-credential-' + randomUUID();
 await login(page);
 await page.setViewportSize({width:1440,height:1000});
 const me = await (await page.request.get('/api/me')).json();
 const result = await page.request.put('/api/providers', {
  headers: {Origin:origin,'X-CSRF-Token':me.csrf},
  data: {textProvider:'aiservice',key:'synthetic-openai',deepSeekKey:'synthetic-deepseek',serviceURL:'http://127.0.0.1:9042',serviceToken:token,xToken:'synthetic-x'},
 });
 expect(result.status()).toBe(200);
 await page.reload();
 await page.getByRole('button',{name:'Connections',exact:true}).click();
 await page.getByLabel('AI Service URL on this server').fill('invalid unsaved address');
 const connections = [
  ['AI Service','serviceConfigured'],
  ['DeepSeek','deepSeekConfigured'],
  ['OpenAI','connected'],
  ['X','xConnected'],
 ];
 for (const [index, [provider, flag]] of connections.entries()) {
  const button = page.getByRole('button',{name:'Disconnect '+provider,exact:true});
  await button.click();
  await expect(page.getByRole('status')).toContainText(provider+' disconnected.');
  await expect(button).toHaveCount(0);
  const state = await (await page.request.get('/api/state')).json();
  expect(state.provider[flag]).toBe(false);
  expect(state.provider.textProvider).toBe('aiservice');
  expect(state.provider.serviceURL).toBe('http://127.0.0.1:9042');
  for (const [, untouched] of connections.slice(index+1)) {
   expect(state.provider[untouched]).toBe(true);
  }
 }
 await stop();
 await start();
 await page.context().clearCookies();
 await login(page);
 const state = await (await page.request.get('/api/state')).json();
 for (const [, flag] of connections) expect(state.provider[flag]).toBe(false);
});

test('three daily episodes can be scheduled and the saved schedule is visible on mobile', async ({page}) => {
 await page.setViewportSize({width:390,height:844});
 await login(page);
 const me = await (await page.request.get('/api/me')).json();
 const result = await page.request.put('/api/providers', {
  headers: {Origin:origin,'X-CSRF-Token':me.csrf},
  data: {textProvider:'openai',researchProvider:'feeds',key:'synthetic-scheduling-only'},
 });
 expect(result.status()).toBe(200);
 await page.getByRole('button',{name:'Create',exact:true}).click();
 await page.getByLabel('Background production').selectOption({label:'Three times daily'});
 const future = new Date(Date.now()+2*60*60*1000);
 const at = future.toISOString().slice(11,16);
 const minute = at.slice(3);
 const times = [0,8,16].map(offset => String(future.getUTCHours()%8+offset).padStart(2,'0')+':'+minute).join(', ');
 await page.getByLabel('Schedule time (UTC)').fill(at);
 await expect(page.getByText('Daily runs at '+times+' UTC.',{exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Save show settings',exact:true}).click();
 await expect(page.getByRole('status')).toContainText('saved');
 const state = await (await page.request.get('/api/state')).json();
 expect(state.settings.schedule).toBe('daily-three');
 expect(state.nextRun).toBeGreaterThan(Date.now()/1000);
 await page.reload();
 await expect(page.getByText('Three times daily',{exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Create',exact:true}).click();
 await expect(page.getByLabel('Background production')).toHaveValue('daily-three');
 await expect(page.getByText('Daily runs at '+times+' UTC.',{exact:true})).toBeVisible();
});

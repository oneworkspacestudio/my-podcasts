import React,{createElement as h,useEffect,useRef,useState} from 'react';
import {createRoot} from 'react-dom/client';
import {Dashboard} from './generated/Dashboard';
import {Actions} from './components/ui';
import {normalizeIntro} from './intro-audio';
import './style.css';
type Obj=Record<string,any>;
const base=new URL('./',location.href).pathname.replace(/\/$/,'');
let csrf='';
const setupToken=location.hash.startsWith('#setup/')?location.hash.slice(7):'';
if(setupToken)history.replaceState(null,'',location.pathname);
const empty={showName:'The Tech Signal',topics:'AI, big tech, the people building the future, and what it means for everyday life',people:['sama','satya','sundar','demis','andrej'],feeds:['bbc','verge','techcrunch','google','microsoft','nvidia'],format:'conversation',minutes:15,lookbackHours:48,voiceA:'marin',voiceB:'cedar',schedule:'off',at:'07:00'};
const options=(o:Record<string,string>)=>Object.entries(o).map(([value,label])=>({value,label}));
const date=(n:number)=>new Date(n*1000).toLocaleDateString(undefined,{month:'short',day:'numeric',year:'numeric'});
const time=(n:number)=>Math.floor(n/60)+':'+String(Math.floor(n%60)).padStart(2,'0');
const title=(e:Obj)=>e.script?.title||e.settings?.showName||'New episode';
const status:Obj={ready:['Ready to listen','mint'],queued:['In the queue','blue'],running:['In production','blue'],failed:['Interrupted','amber'],needs_review:['Needs editorial review','amber'],awaiting_voice:['Script ready · voices needed','amber'],cancelled:['Cancelled','neutral']};
const scheduleOptions=options({off:'On demand',daily:'Once daily','daily-three':'Three times daily',weekly:'Every Monday'});
function dailyTimes(at:string){
 const [hour,minute]=at.split(':');
 return [0,8,16].map(offset=>String(Number(hour)%8+offset).padStart(2,'0')+':'+minute).join(', ');
}
const disconnects:Record<string,{flag:string;label:string}> = {
 disconnect:{flag:'clearKey',label:'OpenAI'},
 disconnectX:{flag:'clearX',label:'X'},
 disconnectDeepSeek:{flag:'clearDeepSeek',label:'DeepSeek'},
 disconnectService:{flag:'clearService',label:'AI Service'},
 disconnectVoice:{flag:'clearVoice',label:'My Voice'},
};
function App(){
 const[page,setPage]=useState('library'),[state,setState]=useState<Obj>({episodes:[],people:[],feeds:[],provider:{}}),[form,setForm]=useState<Obj>(empty),[detail,setDetail]=useState<Obj|null>(null),[seek,setSeek]=useState<Obj>(),[loading,setLoading]=useState(true),[signedOut,setSignedOut]=useState(false),[boot,setBoot]=useState<Obj>({}),[owner,setOwner]=useState(false),[error,setError]=useState(''),[toast,setToast]=useState(''),[busy,setBusy]=useState(false),[email,setEmail]=useState(''),[password,setPassword]=useState(''),[key,setKey]=useState(''),[xToken,setXToken]=useState(''),[model,setModel]=useState('gpt-5.4'),[ttsModel,setTTSModel]=useState('gpt-4o-mini-tts');
 const[connections,setConnections]=useState<Obj>({textProvider:'openai',researchProvider:'auto',deepSeekModel:'deepseek-v4-pro',deepSeekKey:'',serviceURL:'http://127.0.0.1:9042',serviceToken:'',speechProvider:'myvoice',voiceURL:'',voiceToken:''});
 const [introFiles,setIntroFiles]=useState<{introFirst?:File;introSecond?:File}>({});
 const epoch=useRef(0),dirty=useRef(false),refreshSequence=useRef(0);
 const check=useRef<AbortController|null>(null);
 const [checking,setChecking]=useState(false);
 function stopCheck(){check.current?.abort();check.current=null;setChecking(false);}
 function expire(){epoch.current++;stopCheck();csrf='';setState({episodes:[],people:[],feeds:[],provider:{}});setDetail(null);setKey('');setXToken('');setPassword('');setIntroFiles({});setConnections(c=>({...c,deepSeekKey:'',serviceToken:'',voiceToken:''}));setSignedOut(true);}
 async function api(path:string,method='GET',body?:unknown,signal?:AbortSignal){
  const multipart=body instanceof FormData;
  const headers:Record<string,string>=method==='GET'?{}:{'X-CSRF-Token':csrf};
  if(method!=='GET'&&!multipart)headers['Content-Type']='application/json';
  const res=await fetch(base+'/api/'+path,{method,signal,headers,body:multipart?body:body===undefined?undefined:JSON.stringify(body)});const data=await res.json().catch(()=>({}));if(res.status===401){expire();throw new Error('Sign in to continue');}if(!res.ok)throw new Error(data.error||'Your studio could not connect. Please try again.');return data;}
 async function refresh(initial=false){
  const current=epoch.current,sequence=++refreshSequence.current;
  const latest=()=>current===epoch.current&&sequence===refreshSequence.current;
  const data=await api('state');
  if(!latest())return;
  setState(data);
  if(initial||!dirty.current)setForm(data.settings);
  if(initial){setModel(data.provider.model);setTTSModel(data.provider.ttsModel);setConnections(c=>({...c,textProvider:data.provider.textProvider,researchProvider:data.provider.researchProvider,deepSeekModel:data.provider.deepSeekModel,serviceURL:data.provider.serviceURL,speechProvider:data.provider.speechProvider,voiceURL:data.provider.voiceURL}));}
  if(detail){const episode=await api('episodes/'+detail.id);if(latest())setDetail(previous=>previous?.id===episode.id?episode:previous);}
 }
 async function checkWriter(){
  const controller=new AbortController();
  check.current=controller;setChecking(true);
  try{
   const result=await api('providers/test','POST',{},controller.signal);
   if(check.current===controller)setToast(result.message);
  }catch(e){
   if(check.current===controller)setError((e as Error).message);
  }finally{
   if(check.current===controller){check.current=null;setChecking(false);}
  }
 }
 useEffect(()=>{(async()=>{try{setBoot(await api('bootstrap'));const me=await api('me');csrf=me.csrf;setOwner(me.user.role==='owner');await refresh(true);}catch(e){if((e as Error).message!=='Sign in to continue')setError((e as Error).message);}finally{setLoading(false);}})();return()=>{epoch.current++;};},[]);
 useEffect(()=>{if(signedOut||loading)return;let live=true;const timer=setInterval(()=>{if(live)void refresh().catch(e=>setError((e as Error).message));},5000);return()=>{live=false;clearInterval(timer);};},[signedOut,loading,detail?.id]);
 useEffect(()=>{if(!toast)return;const t=setTimeout(()=>setToast(''),5000);return()=>clearTimeout(t);},[toast]);
 useEffect(()=>{if(page!=='detail')setSeek(undefined);window.scrollTo({top:0,behavior:'instant'});},[page]);
 function change(name:string,value:string|boolean|File|null){
  if(name==='introFirst'||name==='introSecond'){
   setIntroFiles(files=>({...files,[name]:value instanceof File?value:undefined}));return;
  }
  if(typeof value!=='string'&&typeof value!=='boolean')return;
  if(['textProvider','researchProvider','deepSeekModel','deepSeekKey','serviceURL','serviceToken','speechProvider','voiceURL','voiceToken'].includes(name)){setConnections(c=>({...c,[name]:value}));return;}if(name==='email'){setEmail(String(value));return;}if(name==='password'){setPassword(String(value));return;}if(name==='key'){setKey(String(value));return;}if(name==='xToken'){setXToken(String(value));return;}if(name==='model'){setModel(String(value));return;}if(name==='ttsModel'){setTTSModel(String(value));return;}dirty.current=true;if(name.startsWith('person:')||name.startsWith('feed:')){const[category,id]=name.split(':'),field=category==='person'?'people':'feeds';setForm(f=>({...f,[field]:value?[...f[field],id]:f[field].filter((x:string)=>x!==id)}));return;}setForm(f=>({...f,[name]:['minutes','lookbackHours'].includes(name)?Number(value):value}));}
 async function save(){await api('settings','PUT',form);dirty.current=false;await refresh();}
 async function act(intent:string,payload?:unknown){setError('');if(['library','studio','queue','sources','settings'].includes(intent)){setPage(intent);return;}if(busy)return;setBusy(true);try{
  if(intent==='login'){const r=await fetch(base+'/api/'+(boot.needsSetup?'setup':'login'),{method:'POST',headers:{'Content-Type':'application/json','X-Setup-Token':setupToken},body:JSON.stringify({email,password})});if(!r.ok)throw new Error(boot.needsSetup?'Open the private first-use link printed by the app and choose an email and a password of at least 12 characters.':'Sign-in failed. Check your email and password.');setPassword('');location.reload();return;}
  if(intent==='logout'){const r=await fetch(base+'/auth/logout',{method:'POST',headers:{'X-CSRF-Token':csrf}});if(!r.ok)throw new Error('Sign out failed. Please retry.');expire();return;}
  if(intent==='save'){await save();setToast('Your show settings and sources are saved.');}
  if(intent==='generate'){await save();await api('episodes','POST',{});setPage('queue');await refresh();setToast('Your episode is queued. You can close this tab.');}
  if(intent==='saveIntros'){
   const {introFirst,introSecond}=introFiles;
   if(!introFirst||!introSecond)throw new Error('Choose two intro tracks.');
   const [first,second]=await Promise.all([normalizeIntro(introFirst),normalizeIntro(introSecond)]);
   const formData=new FormData();
   formData.append('first',first,'first.wav');
   formData.append('second',second,'second.wav');
   await api('intros','PUT',formData);
   setIntroFiles({});
   await refresh();
   setToast('Intro tracks saved for new episodes.');
  }
  if(intent==='removeIntros'){
   await api('intros','DELETE');
   await refresh();
   setToast('New episodes will skip the intro track.');
  }
  if(intent==='refresh')await refresh();
  if(intent==='open'){const d=await api('episodes/'+payload);setDetail(d);setPage('detail');}
  if(intent==='cancel'){await api('episodes/'+payload+'/cancel','POST',{});await refresh();}
  if(intent==='retry'){await api('episodes/'+payload+'/retry','POST',{});await refresh();setToast('Saved production is back in the queue.');}
  if(intent==='seek')setSeek({at:Number(payload),key:Date.now()});
  if(intent==='delete'&&confirm('Delete this episode, its transcript and generated audio?')){await api('episodes/'+payload,'DELETE');setDetail(null);setPage('library');await refresh();}
  const disconnect=disconnects[intent];
  if(disconnect){
   stopCheck();
   const result=await api('providers','PUT',{[disconnect.flag]:true});
   refreshSequence.current++;
   setState(s=>({...s,provider:result.provider}));
   if(intent==='disconnect')setKey('');
   if(intent==='disconnectX')setXToken('');
   if(intent==='disconnectDeepSeek')setConnections(c=>({...c,deepSeekKey:''}));
   if(intent==='disconnectService')setConnections(c=>({...c,serviceToken:''}));
   if(intent==='disconnectVoice')setConnections(c=>({...c,voiceToken:''}));
   await refresh();
   setToast(disconnect.label+' disconnected.');
  }
  if(intent==='saveProviders'||intent==='testProvider'){
   stopCheck();
   await api('providers','PUT',{...connections,key,xToken,model,ttsModel});
   setKey('');setXToken('');setConnections(c=>({...c,deepSeekKey:'',serviceToken:'',voiceToken:''}));
   await refresh();
   setToast('Studio connections updated.');
   if(intent==='testProvider')void checkWriter();
  }
 }catch(e){setError((e as Error).message);}finally{setBusy(false);}}
 const waitingMessage=(e:Obj)=>e.status==='awaiting_voice'?(state.provider.speechIssue||'The server will resume recording when My Voice is available.'):e.error;
 const episodes=(state.episodes||[]).map((e:Obj)=>({...e,error:waitingMessage(e),stage:e.status==='awaiting_voice'?'Script ready · waiting for narration':e.stage,title:title(e),summary:e.script?.summary||e.stage,date:date(e.created),statusLabel:status[e.status]?.[0]||e.status,tone:status[e.status]?.[1]||'neutral',meta:e.status==='ready'?`${time(e.duration)} · ${e.sources?.length||0} sources · Chaptered audio`:`${e.settings?.minutes} min target · ${e.stage}`,icon:e.status==='ready'?'play':'chevron',action:e.status==='ready'?'Listen':'View',running:['running','queued'].includes(e.status),retryable:['failed','cancelled','needs_review','awaiting_voice'].includes(e.status)}));
 const jobs=episodes.filter((e:Obj)=>e.status!=='ready');const d=detail||{};const ready=d.status==='ready';const endpoint=base+'/api/episodes/'+d.id;
 const savedSchedule=state.settings?.schedule||'off';
 const scheduleDescription=state.schedulePaused?'Automatic episodes are paused until writing, research and speech are connected.':savedSchedule==='off'?'Enable automatic episodes in Create an episode.':savedSchedule==='daily-three'?`Three times daily at ${dailyTimes(state.settings.at)} UTC.`:`${savedSchedule==='daily'?'Every day':'Every Monday'} at ${state.settings.at} UTC.`;
 const nextRunLabel=state.schedulePaused?`Paused. ${state.provider.speechIssue||'Configure research and writing in Connections.'}`:!state.nextRun?'Automatic episodes are off':state.nextRun*1000<=Date.now()?'Due. Waiting for production capacity.':`Next run ${new Date(state.nextRun*1000).toLocaleString()}`;
 const data:Obj={...form,...connections,page,loading,signedOut,localLogin:!boot.oidc,loginLabel:boot.needsSetup?'Create your private account':'Sign in',loginURL:base+'/auth/login',email,password,error,toast,busy,checking,busyOrChecking:busy||checking,isOwner:owner,key,xToken,model,ttsModel,
  ...Object.fromEntries(['library','studio','queue','sources','settings'].map(p=>[p+'Variant',page===p?'active':'nav'])),
  showLibrary:page==='library',showStudio:page==='studio',showQueue:page==='queue',showSources:page==='sources',showSettings:page==='settings',showDetail:page==='detail',
  episodes,jobs,emptyLibrary:episodes.length===0,emptyQueue:jobs.length===0,readyCount:episodes.filter((e:Obj)=>e.status==='ready').length,sourceCount:form.people.length+form.feeds.length,
  scheduleLabel:scheduleOptions.find(option=>option.value===savedSchedule)?.label,scheduleDescription,nextRunLabel,editionTimes:form.schedule==='daily-three'&&form.at?`Daily runs at ${dailyTimes(form.at)} UTC.`:'',libraryDescription:`${episodes.length} episodes · Private to your account`,
  previewDescription:`${form.minutes} minutes · ${form.format==='conversation'?'A thoughtful conversation':form.format==='briefing'?'A focused news briefing':'A narrative deep dive'}`,needsConnection:!state.provider.textConfigured||!state.provider.researchConfigured,speechMissing:!state.provider.speechConfigured,
  formatOptions:options({conversation:'Two-host conversation',briefing:'The focused briefing',deepdive:'The narrative deep dive'}),minuteOptions:options({'8':'8 minutes','15':'15 minutes','25':'25 minutes','40':'40 minutes'}),lookbackOptions:options({'24':'Past 24 hours','48':'Past 48 hours','168':'Past 7 days'}),voiceOptions:options({marin:'Marin',cedar:'Cedar',coral:'Coral',alloy:'Alloy',sage:'Sage',ash:'Ash'}),scheduleOptions,
  connectionLabel:state.provider.textConfigured?'Configured':'Connection needed',connectionTone:state.provider.textConfigured?'mint':'amber',
  writerOptions:options({openai:'OpenAI',deepseek:'DeepSeek',aiservice:'AI Service · local Ollama'}),researchOptions:options({auto:'Automatic · selected writing provider',openai:'OpenAI web research',deepseek:'DeepSeek web search',feeds:'Selected feeds & direct X posts'}),
  deepSeekLabel:state.provider.deepSeekConfigured?'Connected':'Disconnected',serviceLabel:state.provider.serviceConfigured?'Connected':'Disconnected',speechLabel:state.provider.speechConfigured?'Ready':'Waiting for voices',xLabel:state.provider.xConnected?'Connected':'Disconnected',
  introsEnabled:state.introsEnabled,
  introSelection:[introFiles.introFirst?.name,introFiles.introSecond?.name].filter(Boolean).join(' · '),
  introSaveDisabled:busy||!introFiles.introFirst||!introFiles.introSecond,
  localDuo:state.provider.speechProvider==='myvoice'&&state.provider.voiceCast==='marius-alba-v1',
  cloudSpeech:state.provider.speechProvider==='openai',cloudSpeechDraft:connections.speechProvider==='openai',localSpeechSelected:connections.speechProvider==='myvoice',managedVoiceLink:boot.oidc,voiceConnected:state.provider.voiceConnected,speechIssue:state.provider.speechIssue,speechOptions:options({myvoice:'My Voice · local on this server',openai:'OpenAI speech · optional'}),
  deepSeekConnected:state.provider.deepSeekConfigured,serviceConnected:state.provider.serviceConfigured,openAIConnected:state.provider.connected,xConnected:state.provider.xConnected,
  researchCoverage:state.provider.researchMode==='feeds'?'Feed-based mode reads selected dated RSS sources and connected direct X posts. No indexed web/social search runs in this mode.':'Web research checks major reports, company announcements and accessible public posts. Missing, inaccessible and undated evidence is flagged.',
  socialDescription:state.provider.researchMode==='feeds'&&!state.provider.xConnected?'No direct X token is configured. Feed-based mode does not read social posts. Connect X or select a web research provider.':state.provider.xConnected?'Direct X timelines are connected. Public original posts are checked alongside independent reporting. Individual source failures are recorded per episode.':'Public posts are checked through indexed web research. Connect the optional X API for direct timeline access. Neither method guarantees complete social coverage.',
  peopleCount:form.people.length,people:(state.people||[]).map((p:Obj)=>({...p,initials:p.name.split(' ').map((x:string)=>x[0]).join(''),selected:form.people.includes(p.id),field:'person:'+p.id,label:'Follow '+p.name})),feeds:(state.feeds||[]).map((f:Obj)=>({...f,selected:form.feeds.includes(f.id),field:'feed:'+f.id,label:'Include '+f.name,kind:f.kind==='primary'?'PRIMARY SOURCE':'INDEPENDENT REPORTING',tone:f.kind==='primary'?'mint':'blue'})),
  hasScript:(d.script?.segments?.length||0)>0,detailId:d.id,detailTitle:title(d),detailSummary:d.script?.summary,detailStatus:status[d.status]?.[0],detailTone:status[d.status]?.[1],detailReady:ready,detailStage:d.status==='awaiting_voice'?'Script ready · waiting for narration':d.stage,detailProgress:d.progress,detailError:waitingMessage(d),detailRunning:['queued','running'].includes(d.status),detailRetryable:['failed','cancelled','needs_review','awaiting_voice'].includes(d.status),detailDeletable:!!d.id&&!['queued','running'].includes(d.status),
  detailMeta:d.created?`${date(d.created)} · ${ready?time(d.duration):d.settings?.minutes+' min target'} · ${d.sources?.length||0} sources`:'',lengthNote:d.lengthNote,audioURL:endpoint+'/audio',downloadURL:endpoint+'/audio?download=1',transcriptURL:endpoint+'/transcript',sourcesURL:endpoint+'/sources',seek,
  hasWarnings:d.warnings?.length>0,warnings:(d.warnings||[]).map((text:string)=>({text})),hasIssues:d.review?.issues?.length>0,issues:(d.review?.issues||[]).map((text:string)=>({text})),cutoff:d.cutoff||'Research has not started',
  chapters:(d.script?.segments||[]).map((s:Obj)=>({...s,time:time(s.start),turns:s.turns.map((t:Obj)=>({...t,citations:t.sources?.length?'Evidence · '+t.sources.join(' · '):''}))})),detailSources:(d.sources||[]).map((s:Obj)=>({...s,dateLabel:s.published?new Date(s.published).toLocaleString():'Publication date not supplied by search; see linked source'}))};
 return h(Actions.Provider,{value:{act,change}},h(Dashboard,{data,onAction:act}));
}
createRoot(document.getElementById('root')!).render(h(App));

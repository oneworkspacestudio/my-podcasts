const sampleRate = 24000;
const maxSeconds = 60;

export async function normalizeIntro(file:File):Promise<Blob>{
 if(file.size>20*1024*1024)throw new Error('Choose an intro file under 20 MB.');
 let decoded:AudioBuffer;
 try{
  const decoder=new OfflineAudioContext(1,1,sampleRate);
  decoded=await decoder.decodeAudioData(await file.arrayBuffer());
 }catch{
  throw new Error('This audio file could not be opened. Choose an MP3 or WAV.');
 }
 if(decoded.duration<1||decoded.duration>maxSeconds)throw new Error('Each intro track must be between 1 and 60 seconds.');
 const context=new OfflineAudioContext(1,decoded.length,sampleRate);
 const source=context.createBufferSource();
 source.buffer=decoded;
 source.connect(context.destination);
 source.start();
 const pcm=(await context.startRendering()).getChannelData(0);
 const bytes=new ArrayBuffer(44+pcm.length*2);
 const view=new DataView(bytes);
 for(const [offset,label] of [[0,'RIFF'],[8,'WAVEfmt '],[36,'data']] as const){
  for(let i=0;i<label.length;i++)view.setUint8(offset+i,label.charCodeAt(i));
 }
 view.setUint32(4,bytes.byteLength-8,true);
 view.setUint32(16,16,true);
 view.setUint16(20,1,true);
 view.setUint16(22,1,true);
 view.setUint32(24,sampleRate,true);
 view.setUint32(28,sampleRate*2,true);
 view.setUint16(32,2,true);
 view.setUint16(34,16,true);
 view.setUint32(40,pcm.length*2,true);
 for(let i=0;i<pcm.length;i++)view.setInt16(44+i*2,Math.round(Math.max(-1,Math.min(1,pcm[i]))*32767),true);
 return new Blob([bytes],{type:'audio/wav'});
}

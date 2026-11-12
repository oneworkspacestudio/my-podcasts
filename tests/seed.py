"""Synthetic UI/audio fixture only. Refuses every non-task private-data root."""
import json,math,sqlite3,struct,sys,time,wave
from pathlib import Path
root=Path(sys.argv[1]).resolve()
boundary=Path('C:/Projects/.cache/my-podcasts-providers-20261106').resolve()
assert root.is_relative_to(boundary) and root!=boundary and not root.is_symlink()
assert json.loads((root/'app.json').read_text())['id']=='podcasts'
db=sqlite3.connect(root/'state.sqlite')
users=db.execute('SELECT id,email FROM host_users').fetchall()
owner=next(u for u,email in users if email=='podcast-native@example.invalid')
row=db.execute('SELECT body FROM podcast_settings WHERE user_id=?',(owner,)).fetchone();assert row
settings=json.loads(row[0]);now=int(time.time());episode_id='a1'*24
sources=[dict(id='S1',title='Synthetic primary announcement',url='https://example.com/primary',publisher='Example company',kind='primary',published='2026-11-06T12:00:00Z'),dict(id='S2',title='Synthetic independent reporting',url='https://example.org/report',publisher='Example publisher',kind='reporting',published='2026-11-06T12:05:00Z')]
segments=[dict(title='The question behind the headline',kind='hook',start=0,turns=[dict(speaker='Alex',text='This is a synthetic interface test. What changes when a new technology leaves the lab?',sources=['S1'])]),dict(title='The bigger picture',kind='story',start=1.5,turns=[dict(speaker='Jamie',text='This transcript and its sources are synthetic. A good conversation should make room for evidence, uncertainty, and a different point of view.',sources=['S1','S2'])]),dict(title='What to watch next',kind='outro',start=3,turns=[dict(speaker='Alex',text='Check the original sources, and keep asking thoughtful questions. This fixture contains a tone, not synthesized speech.',sources=[])])]
e=dict(id=episode_id,created=now,updated=now,status='ready',stage='Ready to listen',progress=100,error='',settings=settings,cutoff='2026-11-06T12:30:00Z',sources=sources,warnings=['Synthetic acceptance fixture. Audio is a test tone; this is not real news or provider narration.'],script=dict(title='Synthetic episode · Beyond the headline',summary='A synthetic episode used to verify chapters, source receipts and private audio playback.',segments=segments),review=dict(passed=True,issues=[]),duration=4.0,audioBytes=192044)
folder=root/'objects'/episode_id;folder.mkdir(parents=True,exist_ok=True)
with wave.open(str(folder/'episode.wav'),'wb') as f:
 f.setnchannels(1);f.setsampwidth(2);f.setframerate(24000);f.writeframes(b''.join(struct.pack('<h',int(1500*math.sin(i*2*math.pi*220/24000))) for i in range(96000)))
db.execute('INSERT INTO podcast_episodes VALUES(?,?,?,?,?,?)',(episode_id,owner,now,now,'ready',json.dumps(e)))
rejected={**e,'id':'a2'*24,'status':'needs_review','progress':40,'stage':'Editorial review needs attention',
          'error':'The editorial check found issues. Review the transcript and research, then regenerate before narration',
          'script':{**e['script'],'title':'Synthetic draft needing repair'},
          'review':{'passed':False,'issues':['Remove the unsupported shipping date from the opening story.']},
          'duration':0,'audioBytes':0}
db.execute('INSERT INTO podcast_episodes VALUES(?,?,?,?,?,?)',(rejected['id'],owner,now,now,'needs_review',json.dumps(rejected)))
db.commit();db.close()

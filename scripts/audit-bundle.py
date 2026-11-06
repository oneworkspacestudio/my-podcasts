"""Verify the full ordinary managed candidate archive and declared app ownership."""
import hashlib,json,sys,zipfile
from pathlib import Path
folder=Path('I:/Projects/tooling/my-podcasts-20261106/suite-ready')
def digest(path):
 h=hashlib.sha256()
 with path.open('rb') as f:
  for b in iter(lambda:f.read(4<<20),b''):h.update(b)
 return h.hexdigest()
cat=json.loads((folder/'catalog.json').read_text());registry=json.loads(Path('I:/Projects/workspace-my-podcasts/internal/catalog/managed-apps.json').read_text())['apps']
assert cat['installerIntegration']==14 and {a['id'] for a in cat['apps']}=={a['id'] for a in registry}
owned=[k for k in cat['packages'] if k=='podcasts' or k.startswith('podcasts-')]
assert {'podcasts','podcasts-contract','podcasts-manifest','podcasts-LICENSE','podcasts-NOTICE'}<=set(owned)
with zipfile.ZipFile(folder/'workspace-bundle.zip') as z:
 assert json.loads(z.read('catalog.json'))==cat
 for key,p in cat['packages'].items():
  h=hashlib.sha256();n=0
  with z.open(p['file']) as f:
   for b in iter(lambda:f.read(4<<20),b''):h.update(b);n+=len(b)
  assert h.hexdigest()==p['sha256'] and n==p['size'],key
 assert hashlib.sha256(z.read('MyPodcasts.exe')).hexdigest()==digest(Path('I:/Projects/tooling/my-podcasts-20261106/MyPodcasts.exe'))
baseline=Path('I:/Projects/tooling/web-search-remote-20261106/final121/catalog.json')
comparison={}
if baseline.exists():
 old=json.loads(baseline.read_text());before={a['id']:a for a in old['apps']};after={a['id']:a for a in cat['apps']}
 assert set(before)<=set(after)
 changedApps=[k for k in before if before[k]!=after[k]]
 changedPackages=[k for k,v in old['packages'].items() if k not in cat['packages'] or cat['packages'][k]!=v]
 fields={k:sorted(f for f in set(before[k])|set(after[k]) if before[k].get(f)!=after[k].get(f)) for k in changedApps}
 assert fields=={'websearch':['source']},fields
 assert changedPackages==['manager'],changedPackages
 comparison={'baselineVersion':old['version'],'retainedAppCount':len(before),'changedPriorAppContracts':changedApps,'changedPriorPackages':changedPackages,'changedPriorAppFields':fields,'allPriorLaunchContractsRetained':True,'allNonManagerPackageBytesRetained':True}
report={'schema':1,'passed':True,'apps':len(cat['apps']),'packages':len(cat['packages']),'podcastsPackages':len(owned),'installerSHA256':digest(folder/'WorkspaceSetup-windows-amd64.exe'),'bundleSHA256':digest(folder/'workspace-bundle.zip'),'appSHA256':digest(Path('I:/Projects/tooling/my-podcasts-20261106/MyPodcasts.exe')),'candidate':True,'baselineComparison':comparison}
Path('docs/evidence/bundle-verification.json').write_text(json.dumps(report,indent=2)+'\n',encoding='utf-8');print(json.dumps(report))

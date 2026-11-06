"""Package a candidate or a clean, committed release. Does not publish assets."""
import argparse,hashlib,json,os,shutil,stat,subprocess,zipfile
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--go',required=True);p.add_argument('--exe',required=True);p.add_argument('--out',required=True);p.add_argument('--release',action='store_true');a=p.parse_args()
root=Path(__file__).resolve().parents[1];out=Path(a.out).resolve();stage=out/'package';stage.mkdir(parents=True,exist_ok=True)
commit=None
if a.release:
 if subprocess.check_output(['git','status','--porcelain'],cwd=root,text=True).strip():raise RuntimeError('Release requires clean committed source')
 commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
def copy(src,dst):
 dst=Path(dst)
 if dst.exists():dst.chmod(dst.stat().st_mode|stat.S_IWRITE)
 return shutil.copy2(src,dst)
for rel in ['LICENSE','NOTICE','README.md','compatibility.json','docs/ACCEPTANCE.md','docs/DATA-SOURCES.md','docs/SECURITY.md','docs/PRODUCTION.md']:
 dst=stage/rel;dst.parent.mkdir(parents=True,exist_ok=True);copy(root/rel,dst)
copy(a.exe,stage/'MyPodcasts.exe')
licenses=stage/'licenses';licenses.mkdir(exist_ok=True)
modules=subprocess.check_output([a.go,'list','-m','-json','all'],cwd=root,text=True)
decoder=json.JSONDecoder();items=[]
while modules.strip():
 item,end=decoder.raw_decode(modules.lstrip());modules=modules.lstrip()[end:];items.append(item)
notices=['# Third-party software notices','', 'This package bundles React, React DOM and Lucide icons, and the Go dependencies listed below. Full available licence texts are in licenses/. The AUI compiler is a build-time dependency, not a browser runtime.','']
for m in items:
 if m.get('Main'):continue
 folder=m.get('Dir');name=m['Path'];notices.append('- '+name+' '+m.get('Version',''))
 if folder:
  found=False
  for f in Path(folder).iterdir():
   if f.is_file() and f.name.lower() in ['license','license.txt','license.md','copying','notice']:
    copy(f,licenses/(name.replace('/','_')+'-'+f.name));found=True
  if not found:raise RuntimeError('Locate dependency license for '+name)
for name in ['react','react-dom','scheduler','lucide-react']:
 folder=root/'node_modules'/name;meta=json.loads((folder/'package.json').read_text());notices.append('- '+name+' '+meta['version']+' · '+meta.get('license','see licence'))
 files=[f for f in folder.iterdir() if f.is_file() and f.name.lower().startswith('license')]
 assert files,'Missing licence '+name
 for f in files:copy(f,licenses/(name+'-'+f.name))
go_root=subprocess.check_output([a.go,'env','GOROOT'],text=True).strip();copy(Path(go_root)/'LICENSE',licenses/'Go-LICENSE')
(stage/'THIRD-PARTY-NOTICES.md').write_text('\n'.join(notices)+'\n',encoding='utf-8')
manifest=dict(schema=1,version=json.loads((root/'compatibility.json').read_text(encoding='utf-8'))['version'],candidate=not a.release,sourceCommit=commit,executableSHA256=hashlib.sha256((stage/'MyPodcasts.exe').read_bytes()).hexdigest(),builtNatively=True)
(stage/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
suffix='' if a.release else '-candidate'
archive=out/('MyPodcasts-windows-amd64'+suffix+'.zip')
with zipfile.ZipFile(archive,'w',zipfile.ZIP_DEFLATED,compresslevel=6) as z:
 for f in stage.rglob('*'):
  if f.is_file():z.write(f,f.relative_to(stage).as_posix())
source=out/('MyPodcasts-source'+suffix+'.zip')
with zipfile.ZipFile(source,'w',zipfile.ZIP_DEFLATED) as z:
 for directory,dirs,files in os.walk(root):
  dirs[:]=[d for d in dirs if d not in ['.git','node_modules','dist','test-results','playwright-report','__pycache__']]
  for name in files:
   f=Path(directory)/name
   if not name.startswith('.env') and f.suffix not in ['.exe','.zip','.env','.db','.sqlite','.sqlite3']:z.write(f,'my-podcasts/'+f.relative_to(root).as_posix())
if a.release:
 subprocess.run(['git','archive','--format=zip','--prefix=my-podcasts/','-o',str(source),commit],cwd=root,check=True)
receipt={**manifest,'archive':str(archive),'archiveSHA256':hashlib.sha256(archive.read_bytes()).hexdigest(),'sourceArchive':str(source)}
(out/'verification.json').write_text(json.dumps(receipt,indent=2)+'\n');print(json.dumps(receipt))

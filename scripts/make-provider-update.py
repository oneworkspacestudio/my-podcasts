"""Prepare an app-only local candidate through the existing checked component contract."""
import copy,hashlib,importlib.util,json,sys,zipfile
from pathlib import Path
config=Path(sys.argv[1]).resolve();boundary=Path('C:/Projects/.cache/my-podcasts-providers-20261106').resolve()
assert config.is_relative_to(boundary) and not config.is_symlink()
package=Path(sys.argv[2]).resolve();out=Path(sys.argv[3]).resolve()
assert package.is_relative_to(Path('I:/Projects/tooling/my-podcasts-providers-20261106').resolve())
assert out.is_relative_to(Path('I:/Projects/tooling/my-podcasts-providers-20261106').resolve())
spec=importlib.util.spec_from_file_location('podcasts_component','I:/Projects/workspace-my-podcasts/scripts/podcasts_component.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
component=module.podcasts_component(package,allow_candidate=True)
c=json.loads(config.read_text(encoding='utf-8-sig'));cat=copy.deepcopy(c['catalog']);app=next(a for a in cat['apps'] if a['id']=='podcasts')
owned={k for k in cat['packages'] if k=='podcasts' or k.startswith('podcasts-')}
assert owned==set(component['inputs']),'A changed package namespace needs a full ordinary catalog rebuild'
app['version']=component['app']['version'];app['source']=component['app']['source']
replacements={}
for key,item in component['inputs'].items():
 assert cat['packages'][key]['file']==item['file']
 content=Path(item['source']).read_bytes();cat['packages'][key]['size']=len(content);cat['packages'][key]['sha256']=hashlib.sha256(content).hexdigest();replacements[key]=content
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=6) as z:
 z.writestr('catalog.json',json.dumps(cat,indent=2))
 for key,p in cat['packages'].items():
  if key in replacements:z.writestr(p['file'],replacements[key])
  else:z.write(Path(c['programs'])/p['file'],p['file'])
with zipfile.ZipFile(out) as z:
 for key,p in cat['packages'].items():
  b=z.read(p['file']);assert len(b)==p['size'] and hashlib.sha256(b).hexdigest()==p['sha256'],key
print('Verified local app-only candidate',app['version'],len(owned),'owned payloads')

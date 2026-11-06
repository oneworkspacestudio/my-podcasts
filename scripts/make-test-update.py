"""Synthetic app-only update preserving exact runtime contract; no public metadata."""
import copy,hashlib,json,sys,zipfile
from pathlib import Path
config=Path(sys.argv[1]).resolve()
assert config.is_relative_to(Path('C:/Projects/.cache/my-podcasts-20261106').resolve())
c=json.loads(config.read_text(encoding='utf-8-sig'));cat=copy.deepcopy(c['catalog']);app=next(a for a in cat['apps'] if a['id']=='podcasts');app['version']='0.1.0-preview.2-test'
manifest_id='podcasts-manifest';manifest=json.loads((Path(c['programs'])/cat['packages'][manifest_id]['file']).read_text());manifest['syntheticAcceptance']='app-only update and rollback'
exe=Path(sys.argv[2]).read_bytes();manifest['version']=app['version'];manifest['executableSHA256']=hashlib.sha256(exe).hexdigest()
contract=json.loads((Path(c['programs'])/cat['packages']['podcasts-contract']['file']).read_text());contract['version']=app['version']
replacements={'podcasts':exe,manifest_id:json.dumps(manifest,indent=2).encode(),'podcasts-contract':json.dumps(contract,indent=2).encode()}
for key,body in replacements.items():cat['packages'][key]['sha256']=hashlib.sha256(body).hexdigest();cat['packages'][key]['size']=len(body)
with zipfile.ZipFile(sys.argv[3],'w',zipfile.ZIP_DEFLATED) as z:
 z.writestr('catalog.json',json.dumps(cat))
 for key,pkg in cat['packages'].items():
  if key in replacements:z.writestr(pkg['file'],replacements[key])
  else:z.write(Path(c['programs'])/pkg['file'],pkg['file'])

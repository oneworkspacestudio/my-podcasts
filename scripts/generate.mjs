import {parse,compileReact} from '@codedia/parser';
import {readFile,readdir,mkdir,writeFile} from 'node:fs/promises';
await mkdir('web/src/generated',{recursive:true});
for(const n of await readdir('web/src/screens')) if(n.endsWith('.aui')) await writeFile('web/src/generated/'+n.replace('.aui','.tsx'),'// @ts-nocheck -- AUI compiler data bindings. Edit screens/'+n+' instead.\n'+compileReact(parse(await readFile('web/src/screens/'+n,'utf8'))));

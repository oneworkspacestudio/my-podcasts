import {defineConfig} from 'vite';
import react from '@vitejs/plugin-react';
import {fileURLToPath} from 'node:url';
export default defineConfig({root:'web',base:'./',plugins:[react()],resolve:{alias:{'@':fileURLToPath(new URL('./web/src',import.meta.url))}},build:{outDir:'../internal/webui/dist',emptyOutDir:true},server:{proxy:{'/api':'http://127.0.0.1:9044','/auth':'http://127.0.0.1:9044'}}});

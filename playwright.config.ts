import {defineConfig} from '@playwright/test';
export default defineConfig({testDir:'tests',workers:1,timeout:90000,use:{baseURL:'http://127.0.0.1:19444',channel:'msedge',headless:true},reporter:'list'});

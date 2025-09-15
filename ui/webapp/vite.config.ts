import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

export default defineConfig({
  root: 'src',
  publicDir: '../static',
  server: {
    allowedHosts: true,
    host: true,
    port: 8101,
    strictPort: true,
    proxy: {
      '/ws': {
        target: 'ws://localhost:8001',
        ws: true,
      },
      '/resources': {
        target: 'http://localhost:8001',
      }
    }
  },
  build: {
    outDir: '../dist',
    emptyOutDir: true,
  },
  plugins: [preact()],
});

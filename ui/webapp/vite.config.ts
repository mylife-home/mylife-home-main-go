import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';
import { viteSingleFile } from 'vite-plugin-singlefile';

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
    rollupOptions: {
      output: {
        // Put assets directly in root, not in assets folder
        assetFileNames: '[name].[ext]',
      }
    },
  },
  plugins: [preact(), viteSingleFile({
    useRecommendedBuildConfig: false,
    removeViteModuleLoader: true,
    inlinePattern: ['**/*.js', '**/*.css'], // Only inline JS and CSS
  })],
});

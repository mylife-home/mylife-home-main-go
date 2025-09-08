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
      '/socket.io': {
        target: `ws://localhost:8001`,
        ws: true,
      },
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
/*
    rollupOptions: {
      // Seems side effects are dropped else
      // on Money changes size from 3.4Mb to 3.7Mb
      treeshake: false
    }
*/
  },
  plugins: [preact(), viteSingleFile({
    useRecommendedBuildConfig: false,
    removeViteModuleLoader: true,
    inlinePattern: ['**/*.js', '**/*.css'], // Only inline JS and CSS
  })],
});

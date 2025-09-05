import { defineConfig, loadEnv, UserConfigExport } from 'vite';
import preact from '@preact/preset-vite';
import { viteSingleFile } from 'vite-plugin-singlefile';

const VITE_WEB_PORT = 8001;
const VITE_WSTARGET_PORT = 8101;

export default defineConfig({
  root: 'src',
  publicDir: '../static',
  server: {
    host: true,
    port: VITE_WEB_PORT,
    strictPort: true,
    proxy: {
      '/socket.io': {
        target: `ws://localhost:${VITE_WSTARGET_PORT}`,
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

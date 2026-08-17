import { fileURLToPath, URL } from 'node:url'
import { mergeConfig, defineConfig } from 'vitest/config'
import viteConfig from './vite.config.js'

export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      environment: 'jsdom',
      // Viewer JS lives next to its Go widget (see the vanilla-viewer spec),
      // so tests are picked up from outside webui/ as well.
      include: [
        'src/**/*.{test,spec}.{js,ts}',
        '../internal/**/*.{test,spec}.js',
      ],
      exclude: ['node_modules', 'dist', 'e2e/*'],
      root: fileURLToPath(new URL('./', import.meta.url)),
      passWithNoTests: true,
    },
    resolve: {
      alias: {
        // Widget JS imports the helper by the URL the server exposes it at;
        // map that to the file on disk for tests.
        '/_dashi/assets/dashi.js': fileURLToPath(
          new URL('../internal/dashboard/browser/assets/dashi.js', import.meta.url),
        ),
      },
    },
    server: { fs: { allow: ['..'] } },
  }),
)

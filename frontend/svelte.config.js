import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'
import path from 'path'

export default {
  preprocess: vitePreprocess(),
  kit: undefined,
  vite: {
    resolve: {
      alias: {
        '$lib': path.resolve('./src/lib'),
        '$wailsjs': path.resolve('./wailsjs'),
      }
    }
  }
}

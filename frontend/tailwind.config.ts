import type { Config } from 'tailwindcss'

export default {
  content: ['./src/**/*.{svelte,ts,js}'],
  theme: {
    extend: {
      colors: {
        background: '#111113',
        sidebar:    '#18181b',
        surface:    '#27272a',
        border:     '#3f3f46',
        'text-primary': '#f4f4f5',
        'text-muted':   '#a1a1aa',
        accent:         '#10b981',
        'accent-hover': '#059669',
      }
    }
  }
} satisfies Config

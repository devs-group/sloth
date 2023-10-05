import type { Config } from 'tailwindcss'
export default <Partial<Config>> {
  content: [
    'docs/content/**/*.md'
  ],
  theme: {
    extend: {
      colors: {
        blue: "#4287f5"
      }
    }
  }
}
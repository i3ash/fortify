import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "Fortify",
  description: "Fortify is a command-line security tool for file encryption and protection. More details at [GitHub](https://github.com/i3ash/fortify).",
  vite: {
    server: {
      allowedHosts: true,
    }
  },
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started' },
    ],
    sidebar: [
      {
        text: 'Documentation',
        items: [
          { text: 'Home', link: '/' },
          { text: 'Getting Started', link: '/getting-started' },
        ]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/i3ash/fortify' }
    ],
    search: {
      provider: 'local'
    }
  }
})

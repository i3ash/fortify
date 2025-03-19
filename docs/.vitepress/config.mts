import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "Fortify",
  description: "Fortify is a CLI tool using AES-256 encryption for file protection, supporting Shamir's Secret Sharing and RSA encryption. Installation includes various methods. More details at [GitHub](https://github.com/i3ash/fortify).",
  vite: {
    server: {
      allowedHosts: true,
    }
  },
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started' },
      { text: 'Installation Guide', link: '/installation-guide' },
      { text: 'Usage Guide', link: '/usage-guide' },
      { text: 'Tutorial', link: '/tutorial' },
    ],
    sidebar: [
      {
        text: 'Documentation',
        items: [
          { text: 'Home', link: '/' },
          { text: 'Getting Started', link: '/getting-started' },
          { text: 'Installation Guide', link: '/installation-guide' },
          { text: 'Usage Guide', link: '/usage-guide' },
          { text: 'Tutorial', link: '/tutorial' },
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
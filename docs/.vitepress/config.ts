import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'SendDock',
  description: 'Open-source email marketing platform',
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
  ],
  themeConfig: {
    logo: '/favicon.svg',
    nav: [
      { text: 'Self-Hosting', link: '/self-hosting/installation' },
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API', link: '/api/authentication' },
      {
        text: 'GitHub',
        link: 'https://github.com/arkhe-systems/senddock',
      },
    ],
    sidebar: {
      '/self-hosting/': [
        {
          text: 'Self-Hosting',
          items: [
            { text: 'Installation', link: '/self-hosting/installation' },
            { text: 'Configuration', link: '/self-hosting/configuration' },
            { text: 'Updating', link: '/self-hosting/updating' },
          ],
        },
      ],
      '/guide/': [
        {
          text: 'Using SendDock',
          items: [
            { text: 'What is SendDock', link: '/guide/what-is-senddock' },
            { text: 'Getting Started', link: '/guide/getting-started' },
          ],
        },
        {
          text: 'Features',
          items: [
            { text: 'Projects', link: '/guide/projects' },
            { text: 'Subscribers', link: '/guide/subscribers' },
            { text: 'Templates', link: '/guide/templates' },
            { text: 'Email Sending', link: '/guide/sending' },
            { text: 'Campaigns', link: '/guide/campaigns' },
            { text: 'API Keys', link: '/guide/api-keys' },
          ],
        },
        {
          text: 'Configuration',
          items: [
            { text: 'SMTP Setup', link: '/guide/smtp' },
            { text: 'Environment Variables', link: '/guide/environment' },
          ],
        },
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Authentication', link: '/api/authentication' },
            { text: 'Projects', link: '/api/projects' },
            { text: 'Subscribers', link: '/api/subscribers' },
            { text: 'Templates', link: '/api/templates' },
            { text: 'Email Sending', link: '/api/sending' },
            { text: 'Campaigns', link: '/api/campaigns' },
            { text: 'API Keys', link: '/api/api-keys' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/arkhe-systems/senddock' },
    ],
    footer: {
      message: 'Released under the AGPL-3.0 License.',
      copyright: 'Part of Arkhe Systems',
    },
    search: {
      provider: 'local',
    },
  },
})

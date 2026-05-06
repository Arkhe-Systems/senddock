import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'SendDock',
  description: 'Open-source email marketing platform',
  cleanUrls: true,
  sitemap: {
    hostname: 'https://docs.senddock.dev',
  },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
  ],
  themeConfig: {
    logo: '/favicon.svg',
    nav: [
      { text: 'Self-Hosting', link: '/self-hosting/installation' },
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API', link: '/api/authentication' },
      { text: 'Changelog', link: '/changelog' },
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
            { text: 'Troubleshooting & FAQ', link: '/self-hosting/troubleshooting' },
            { text: 'Changelog', link: '/changelog' },
          ],
        },
      ],
      '/guide/': [
        {
          text: 'Using SendDock',
          items: [
            { text: 'What is SendDock', link: '/guide/what-is-senddock' },
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Workspaces', link: '/guide/workspaces' },
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
            { text: 'Suppressions', link: '/guide/suppressions' },
            { text: 'Bounces', link: '/guide/bounces' },
            { text: 'API Keys', link: '/guide/api-keys' },
          ],
        },
        {
          text: 'Pro Features',
          items: [
            { text: 'Analytics', link: '/guide/analytics' },
            { text: 'Webhooks', link: '/guide/webhooks' },
            { text: 'Audit Log', link: '/guide/audit-log' },
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
            { text: 'Code examples', link: '/api/code-examples' },
            { text: 'Workspaces', link: '/api/workspaces' },
            { text: 'Projects', link: '/api/projects' },
            { text: 'Subscribers', link: '/api/subscribers' },
            { text: 'Templates', link: '/api/templates' },
            { text: 'Email Sending', link: '/api/sending' },
            { text: 'Campaigns', link: '/api/campaigns' },
            { text: 'API Keys', link: '/api/api-keys' },
          ],
        },
        {
          text: 'Pro API',
          items: [
            { text: 'Analytics', link: '/api/analytics' },
            { text: 'Webhooks', link: '/api/webhooks' },
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

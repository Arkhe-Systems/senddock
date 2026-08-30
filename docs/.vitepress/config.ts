import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'SendDock',
  description: 'Open-source email marketing platform',
  cleanUrls: true,
  srcExclude: ['**/internal/**', 'internal/**'],
  sitemap: {
    hostname: 'https://docs.senddock.dev',
  },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
  ],
  markdown: {
    config(md) {
      const original = md.renderer.rules.code_inline
      md.renderer.rules.code_inline = (tokens, idx, options, env, self) => {
        const token = tokens[idx]
        if (token.content.includes('{{') || token.content.includes('}}')) {
          return `<code v-pre>${md.utils.escapeHtml(token.content)}</code>`
        }
        return original
          ? original(tokens, idx, options, env, self)
          : self.renderToken(tokens, idx, options)
      }
    },
  },
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
            { text: 'Backups & Recovery', link: '/self-hosting/backups' },
            { text: 'Monitoring', link: '/self-hosting/monitoring' },
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
            { text: 'Your account & security', link: '/guide/account' },
            { text: 'Workspaces', link: '/guide/workspaces' },
          ],
        },
        {
          text: 'Features',
          items: [
            { text: 'Projects', link: '/guide/projects' },
            { text: 'Subscribers', link: '/guide/subscribers' },
            { text: 'Newsletters', link: '/guide/newsletters' },
            { text: 'Segments', link: '/guide/segments' },
            { text: 'Templates', link: '/guide/templates' },
            { text: 'Email Sending', link: '/guide/sending' },
            { text: 'Broadcasts', link: '/guide/broadcasts' },
            { text: 'Campaigns', link: '/guide/campaigns' },
            { text: 'Logs', link: '/guide/logs' },
            { text: 'Analytics', link: '/guide/analytics' },
            { text: 'Suppressions', link: '/guide/suppressions' },
            { text: 'Bounces', link: '/guide/bounces' },
            { text: 'Webhooks', link: '/guide/webhooks' },
            { text: 'API Keys', link: '/guide/api-keys' },
          ],
        },
        {
          text: 'Pro & Team',
          items: [
            { text: 'Deliverability', link: '/guide/deliverability' },
            { text: 'Reports', link: '/guide/reports' },
            { text: 'Audit Log', link: '/guide/audit-log' },
            { text: 'Members & roles', link: '/guide/members' },
          ],
        },
        {
          text: 'Configuration',
          items: [
            { text: 'Instance settings', link: '/guide/instance-settings' },
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
            { text: 'TypeScript SDK', link: '/api/sdk' },
            { text: 'Code examples', link: '/api/code-examples' },
            { text: 'Workspaces', link: '/api/workspaces' },
            { text: 'Projects', link: '/api/projects' },
            { text: 'Subscribers', link: '/api/subscribers' },
            { text: 'Newsletters', link: '/api/newsletters' },
            { text: 'Segments', link: '/api/segments' },
            { text: 'Templates', link: '/api/templates' },
            { text: 'Email Sending', link: '/api/sending' },
            { text: 'Campaigns', link: '/api/campaigns' },
            { text: 'Analytics', link: '/api/analytics' },
            { text: 'Suppressions', link: '/api/suppressions' },
            { text: 'Bounces', link: '/api/bounces' },
            { text: 'Webhooks', link: '/api/webhooks' },
            { text: 'API Keys', link: '/api/api-keys' },
          ],
        },
        {
          text: 'Pro API',
          items: [
            { text: 'Deliverability', link: '/api/deliverability' },
            { text: 'Reports', link: '/api/reports' },
            { text: 'Audit Log', link: '/api/audit-log' },
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

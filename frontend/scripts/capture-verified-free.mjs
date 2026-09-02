import { chromium } from 'playwright'
import { fileURLToPath } from 'node:url'
import { mkdirSync } from 'node:fs'

const BASE = 'http://localhost:8080'
const PID = 'd5d219c0-f3b2-4841-bd08-d0ceb27e5417'
const WSID = 'a930b2a9-2a0c-42bf-abb5-d03d065a8780'
const OUT = fileURLToPath(new URL('../../docs/public/screenshots/', import.meta.url))
mkdirSync(OUT, { recursive: true })
const project = (path = '') => `${BASE}/projects/${PID}${path}`

const browser = await chromium.launch({ headless: true })
const context = await browser.newContext({ viewport: { width: 1440, height: 900 } })
const page = await context.newPage()

async function ensureAuthed() {
    const res = await context.request.post(`${BASE}/api/v1/auth/login`, {
        data: { email: 'dev@senddock.local', password: 'SendDock!2026' },
    })
    return res.status()
}

async function shot(name, url, opts = {}) {
    try {
        await page.goto(url, { waitUntil: 'networkidle' })
        await page.waitForTimeout(opts.wait ?? 900)
        if (opts.before) await opts.before()
        // Verify we are on the app, not the login screen.
        const bad = page.url().includes('/login') || (await page.locator('input[type=email], input[name=email]').count()) > 0
        if (bad) {
            console.log('LOGIN', name, '->', page.url())
            const status = await ensureAuthed()
            await page.goto(url, { waitUntil: 'networkidle' })
            await page.waitForTimeout(opts.wait ?? 900)
            if (opts.before) await opts.before()
            if (page.url().includes('/login')) { console.log('FAIL', name, 'still login'); return }
            console.log('RETRY-OK', name, 're-auth', status)
        }
        await page.screenshot({ path: `${OUT}/${name}.png` })
        console.log('OK  ', name)
    } catch (e) {
        console.log('FAIL', name, '—', String(e).slice(0, 140))
    }
}

async function clickButton(...labels) {
    for (const label of labels) {
        const loc = page.getByRole('button', { name: label }).first()
        if (await loc.count()) { await loc.click(); await page.waitForTimeout(500); return true }
    }
    for (const label of labels) {
        const loc = page.getByText(label, { exact: false }).first()
        if (await loc.count()) { await loc.click(); await page.waitForTimeout(500); return true }
    }
    return false
}

async function clickTab(name) {
    const loc = page.getByRole('tab', { name }).or(page.getByText(name, { exact: true })).first()
    if (await loc.count()) { await loc.click(); await page.waitForTimeout(700) }
}

await ensureAuthed()

// ---------- Simple route captures ----------
await shot('dashboard', `${BASE}/dashboard`)
await shot('hero', project(''))
await shot('projects', project(''))
await shot('project-overview', project(''))
await shot('subscribers-fields-tags', project('/subscribers'))
await shot('suppressions', project('/suppressions'))
await shot('templates-list', project('/templates'))
await shot('logs', project('/logs'))
await shot('broadcasts', project('/broadcasts'))
await shot('campaigns', project('/campaigns'))
await shot('smtp', project('/smtp'))
await shot('webhooks', project('/webhooks'))
await shot('api-keys', project('/settings'))
await shot('custom-fields-settings', project('/settings'))
await shot('account', `${BASE}/account`)
await shot('instance-settings', `${BASE}/instance`)

// Analytics free tabs
await shot('analytics-overview', project('/analytics'))
await shot('analytics-audience', project('/analytics'), { before: () => clickTab('Audience') })
await shot('analytics-engagement', project('/analytics'), { before: () => clickTab('Engagement') })
await shot('analytics', project('/analytics'))

// Logs filtered to bounced
await shot('bounces', project('/logs'), {
    before: async () => { const chip = page.getByText('Bounced', { exact: true }).first(); if (await chip.count()) { await chip.click(); await page.waitForTimeout(600) } },
})

// ---------- Modal captures ----------
await shot('new-project-modal', `${BASE}/dashboard`, { before: () => clickButton('+ New Project', 'New Project') })
await shot('add-subscriber-modal', project('/subscribers'), { before: () => clickButton('+ Add Subscriber', 'Add Subscriber') })
await shot('import-modal', project('/subscribers'), { before: () => clickButton('Import') })
await shot('subscribers-edit-modal', project('/subscribers'), {
    before: async () => { const row = page.getByRole('button', { name: 'Edit' }).first(); if (await row.count()) await row.click(); else await clickButton('Edit'); await page.waitForTimeout(400) },
})
await shot('new-campaign-modal', project('/campaigns'), { before: () => clickButton('+ New Campaign', 'New Campaign') })
await shot('new-webhook-modal', project('/webhooks'), { before: () => clickButton('New webhook', '+ New webhook', 'New Webhook') })
await shot('new-api-key-modal', project('/settings'), { before: () => clickButton('+ Create Key', 'Create Key') })
await shot('custom-fields-modal', project('/settings'), { before: () => clickButton('+ Add field', 'Add field', 'New field', '+ Add Field') })
await shot('segments-builder', project('/segments'), { before: () => clickButton('+ New segment', 'New segment', 'New Segment', '+ New Segment') })
await shot('send-modal', project(''), { before: () => clickButton('Send Email', 'Send') })

// Templates: list, library modal, and editor tabs
await shot('template-library', project('/templates'), { before: () => clickButton('Browse library', '★ Browse library') })
await shot('template-editor', project('/templates'), {
    before: async () => { const open = page.getByText('Welcome (visual)', { exact: false }).first(); if (await open.count()) { await open.click(); await page.waitForTimeout(900) } },
})
await shot('editor', project('/templates'), {
    before: async () => {
        const open = page.getByText('Welcome (visual)', { exact: false }).first()
        if (await open.count()) { await open.click(); await page.waitForTimeout(900) }
        await clickButton('Visual')
    },
})
await shot('send-rich-text', project(''), {
    before: async () => { await clickButton('Send Email', 'Send'); await page.waitForTimeout(500); await clickButton('Rich', 'Text / Rich') },
})

await browser.close()

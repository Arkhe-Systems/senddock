import { chromium } from 'playwright'
import { fileURLToPath } from 'node:url'
import { mkdirSync } from 'node:fs'

const BASE = 'http://localhost:8081'
const PID = 'd5d219c0-f3b2-4841-bd08-d0ceb27e5417'
const WSID = 'a930b2a9-2a0c-42bf-abb5-d03d065a8780'
const OUT = fileURLToPath(new URL('../../docs/public/screenshots/', import.meta.url))
mkdirSync(OUT, { recursive: true })

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
        await page.waitForTimeout(opts.wait ?? 1100)
        if (opts.before) await opts.before()
        if (page.url().includes('/login') || (await page.locator('input[type=email], input[name=email]').count()) > 0) {
            console.log('LOGIN', name)
            await ensureAuthed()
            await page.goto(url, { waitUntil: 'networkidle' })
            await page.waitForTimeout(opts.wait ?? 1100)
            if (opts.before) await opts.before()
            if (page.url().includes('/login')) { console.log('FAIL', name, 'still login'); return }
        }
        await page.screenshot({ path: `${OUT}/${name}.png` })
        console.log('OK  ', name)
    } catch (e) {
        console.log('FAIL', name, '—', String(e).slice(0, 140))
    }
}

async function clickTab(name) {
    const loc = page.getByRole('tab', { name }).or(page.getByText(name, { exact: true })).first()
    if (await loc.count()) { await loc.click(); await page.waitForTimeout(1000) }
}

async function setField(label, value) {
    const sel = page.locator(`label:has(> span:text-is("${label}")) select`).first()
    if (await sel.count()) { await sel.selectOption({ label: value }); await page.waitForTimeout(900) }
}

await ensureAuthed()

// Deliverability tab
await shot('deliverability', `${BASE}/projects/${PID}/analytics`, { before: () => clickTab('Deliverability') })

// Audit log
await shot('audit-log', `${BASE}/projects/${PID}/audit-log`)

// Reports: builder default, then donut / area / pivot variants
await shot('reports', `${BASE}/projects/${PID}/reports`)
await shot('reports-donut', `${BASE}/projects/${PID}/reports`, { before: () => setField('Chart', 'donut') })
await shot('reports-area', `${BASE}/projects/${PID}/reports`, { before: () => setField('Chart', 'area') })
await shot('reports-pivot', `${BASE}/projects/${PID}/reports`, {
    before: async () => {
        await setField('Group by', 'Plan')
        await setField('Then by (pivot)', 'Tag')
        await setField('Chart', 'bar')
    },
})

// Members list + create-user modal
await shot('workspace-members', `${BASE}/workspaces/${WSID}/members`)
const cu = page.getByRole('button', { name: '+ Create user' }).first()
await shot('team-create-user', `${BASE}/workspaces/${WSID}/members`, {
    before: async () => { if (await cu.count()) { await cu.click(); await page.waitForTimeout(900) } },
})

await browser.close()

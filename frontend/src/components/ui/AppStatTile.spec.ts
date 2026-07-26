import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AppStatTile from './AppStatTile.vue'

describe('AppStatTile', () => {
    it('shows the value and label', () => {
        const w = mount(AppStatTile, { props: { label: 'Sent', value: 1234 } })
        expect(w.text()).toContain('Sent')
        expect(w.text()).toContain('1234')
    })

    it('renders no trend when none is given', () => {
        const w = mount(AppStatTile, { props: { label: 'Active subs', value: 10 } })
        expect(w.text()).not.toContain('↑')
        expect(w.text()).not.toContain('↓')
    })

    it('colours a rising metric green by default', () => {
        const w = mount(AppStatTile, { props: { label: 'Open rate', value: '20%', trend: 5 } })
        expect(w.html()).toContain('text-emerald-400')
        expect(w.text()).toContain('↑ 5.0%')
    })

    it('colours a rising metric red when invertGood is set', () => {
        // Bounce rate going up is bad, so the same upward trend must read red.
        const w = mount(AppStatTile, { props: { label: 'Bounce rate', value: '3%', trend: 5, invertGood: true } })
        expect(w.html()).toContain('text-red-400')
    })

    it('colours a falling bad-metric green when invertGood is set', () => {
        const w = mount(AppStatTile, { props: { label: 'Bounce rate', value: '1%', trend: -8, invertGood: true } })
        expect(w.html()).toContain('text-emerald-400')
        expect(w.text()).toContain('↓ 8.0%')
    })
})

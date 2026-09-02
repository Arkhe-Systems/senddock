import { describe, it, expect } from 'vitest'
import { detectTemplateVariables, systemVariablesFor } from './templateVariables'

describe('detectTemplateVariables', () => {
    it('extracts names from the body and subject, deduplicated', () => {
        const html = '<p>Hi {{name}}, your plan is {{custom.plan_tier}}</p>'
        const subject = 'Welcome {{name}}'
        expect(detectTemplateVariables(html, subject)).toEqual(['name', 'custom.plan_tier'])
    })

    it('tolerates whitespace inside the braces', () => {
        expect(detectTemplateVariables('<p>{{  email  }}</p>', '')).toEqual(['email'])
    })

    it('returns an empty array when there are no placeholders', () => {
        expect(detectTemplateVariables('<p>plain text</p>', 'No vars')).toEqual([])
    })

    it('ignores placeholders with unsupported characters', () => {
        expect(detectTemplateVariables('<p>{{my-var}} {{valid}}</p>', '')).toEqual(['valid'])
    })
})

describe('systemVariablesFor', () => {
    it('lists email placeholders for email templates', () => {
        expect(systemVariablesFor('email')).toEqual(['name', 'email', 'subscriber_id', 'unsubscribe_url'])
    })

    it('lists page placeholders for page templates', () => {
        expect(systemVariablesFor('page')).toEqual(['project_name', 'email', 'newsletter_name', 'confirm_button'])
    })

    it('defaults to email placeholders when the type is unknown', () => {
        expect(systemVariablesFor(null)).toEqual(['name', 'email', 'subscriber_id', 'unsubscribe_url'])
        expect(systemVariablesFor(undefined)).toEqual(['name', 'email', 'subscriber_id', 'unsubscribe_url'])
    })
})

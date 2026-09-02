// Shared template-variable helpers used by the template editor. Kept in a
// module so the detection and per-type lists are unit-testable without
// mounting the editor (CodeMirror + GrapeJS are heavy to spin up in tests).

const VARIABLE_PATTERN = /\{\{\s*([a-zA-Z0-9_.]+)\s*\}\}/g

export type TemplateType = 'email' | 'page'

/** Names of the `{{...}}` placeholders found in the HTML body and subject, deduplicated. */
export function detectTemplateVariables(html: string, subject: string): string[] {
    const text = html + ' ' + subject
    const matches = Array.from(text.matchAll(VARIABLE_PATTERN))
        .map((m) => m[1])
        .filter((x): x is string => typeof x === 'string')
    return [...new Set(matches)]
}

/** The built-in placeholders offered per template type. */
export function systemVariablesFor(type: TemplateType | null | undefined): string[] {
    if (type === 'page') {
        return ['project_name', 'email', 'newsletter_name', 'confirm_button']
    }
    return ['name', 'email', 'subscriber_id', 'unsubscribe_url']
}

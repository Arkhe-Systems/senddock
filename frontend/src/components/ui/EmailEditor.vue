<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import grapesjs, { type Editor, type Block } from 'grapesjs'
import newsletterPreset from 'grapesjs-preset-newsletter'
import 'grapesjs/dist/css/grapes.min.css'
import DOMPurify from 'dompurify'

const props = defineProps<{
    modelValue: string
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const editorContainer = ref<HTMLElement | null>(null)
let editor: Editor | null = null
let emitTimer: number | null = null
let lastEmitted = ''

// Email templates lean on a <style> block and inline styles; keep them when
// sanitizing before handing the HTML to the canvas.
const SANITIZE_OPTS = { ADD_TAGS: ['style'] }

// While true, changes coming from a programmatic load (mount / syncFrom) are
// not written back to the model. Otherwise loading the HTML into the canvas
// re-serializes it and overwrites the original code with the normalized
// version, even when the user never touched anything.
let suppressEmit = false
let userDirty = false

function loadContent(html: string) {
    if (!editor) return
    suppressEmit = true
    userDirty = false
    editor.setComponents(DOMPurify.sanitize(html, SANITIZE_OPTS))
    lastEmitted = html
    if (emitTimer) {
        clearTimeout(emitTimer)
        emitTimer = null
    }
    window.setTimeout(() => {
        suppressEmit = false
    }, 500)
}

onMounted(() => {
    if (!editorContainer.value) return

    editor = grapesjs.init({
        container: editorContainer.value,
        height: '100%',
        width: 'auto',
        fromElement: false,
        storageManager: false,
        plugins: [newsletterPreset],
        pluginsOpts: {
            [newsletterPreset as any]: {
                modalTitleImport: 'Import HTML',
                modalTitleExport: 'Export HTML',
                updateStyleManager: false,
                useCustomTheme: false,
            }
        },
        canvas: {
            styles: [],
        },
        blockManager: {
            blocks: []
        },
        styleManager: {
            sectors: [
                {
                    name: 'Typography',
                    open: true,
                    properties: [
                        { property: 'font-family', type: 'select', options: [
                            { id: 'Arial, sans-serif', label: 'Arial' },
                            { id: 'Helvetica, sans-serif', label: 'Helvetica' },
                            { id: 'Georgia, serif', label: 'Georgia' },
                            { id: 'Times New Roman, serif', label: 'Times New Roman' },
                            { id: 'Verdana, sans-serif', label: 'Verdana' },
                            { id: 'Tahoma, sans-serif', label: 'Tahoma' },
                            { id: 'Courier New, monospace', label: 'Courier New' },
                        ]},
                        { property: 'font-size', type: 'number', units: ['px'], defaults: '14' },
                        { property: 'font-weight', type: 'select', options: [
                            { id: 'normal', label: 'Normal' },
                            { id: 'bold', label: 'Bold' },
                        ]},
                        { property: 'color', type: 'color' },
                        { property: 'line-height', type: 'number', units: ['px', '%'], defaults: '1.5' },
                        { property: 'text-align', type: 'radio', options: [
                            { id: 'left', label: 'Left' },
                            { id: 'center', label: 'Center' },
                            { id: 'right', label: 'Right' },
                        ]},
                        { property: 'text-decoration', type: 'select', options: [
                            { id: 'none', label: 'None' },
                            { id: 'underline', label: 'Underline' },
                        ]},
                    ],
                },
                {
                    name: 'Background',
                    open: false,
                    properties: [
                        { property: 'background-color', type: 'color' },
                    ],
                },
                {
                    name: 'Spacing',
                    open: false,
                    properties: [
                        { property: 'padding', type: 'composite', properties: [
                            { property: 'padding-top', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'padding-right', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'padding-bottom', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'padding-left', type: 'number', units: ['px'], defaults: '0' },
                        ]},
                        { property: 'margin', type: 'composite', properties: [
                            { property: 'margin-top', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'margin-right', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'margin-bottom', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'margin-left', type: 'number', units: ['px'], defaults: '0' },
                        ]},
                    ],
                },
                {
                    name: 'Size',
                    open: false,
                    properties: [
                        { property: 'width', type: 'number', units: ['px', '%', 'auto'] },
                        { property: 'height', type: 'number', units: ['px', 'auto'] },
                        { property: 'max-width', type: 'number', units: ['px', '%'] },
                    ],
                },
                {
                    name: 'Border',
                    open: false,
                    properties: [
                        { property: 'border-radius', type: 'number', units: ['px'], defaults: '0' },
                        { property: 'border', type: 'composite', properties: [
                            { property: 'border-width', type: 'number', units: ['px'], defaults: '0' },
                            { property: 'border-style', type: 'select', options: [
                                { id: 'none', label: 'None' },
                                { id: 'solid', label: 'Solid' },
                                { id: 'dashed', label: 'Dashed' },
                            ]},
                            { property: 'border-color', type: 'color' },
                        ]},
                    ],
                },
            ],
        },
    })

    const bm = editor.BlockManager
    bm.getAll().reset()

    bm.add('container', {
        label: 'Container',
        category: 'Layout',
        content: `<div data-gjs-custom-name="Container" style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; font-family: Arial, sans-serif;"></div>`,
    })

    bm.add('section', {
        label: 'Section',
        category: 'Layout',
        content: `<div data-gjs-custom-name="Section" style="padding: 20px; background-color: #ffffff;"></div>`,
    })

    bm.add('columns-2', {
        label: '2 Columns',
        category: 'Layout',
        content: `<div data-gjs-custom-name="2 Columns" data-columns="2" style="width: 100%; display: table; table-layout: fixed; border-collapse: collapse;">
            <div data-gjs-custom-name="Column" style="display: table-cell; width: 50%; height: 60px; vertical-align: top; padding: 10px;"></div>
            <div data-gjs-custom-name="Column" style="display: table-cell; width: 50%; height: 60px; vertical-align: top; padding: 10px;"></div>
        </div>`,
    })

    bm.add('columns-3', {
        label: '3 Columns',
        category: 'Layout',
        content: `<div data-gjs-custom-name="3 Columns" data-columns="3" style="width: 100%; display: table; table-layout: fixed; border-collapse: collapse;">
            <div data-gjs-custom-name="Column" style="display: table-cell; width: 33.33%; height: 60px; vertical-align: top; padding: 10px;"></div>
            <div data-gjs-custom-name="Column" style="display: table-cell; width: 33.33%; height: 60px; vertical-align: top; padding: 10px;"></div>
            <div data-gjs-custom-name="Column" style="display: table-cell; width: 33.33%; height: 60px; vertical-align: top; padding: 10px;"></div>
        </div>`,
    })

    bm.add('divider', {
        label: 'Divider',
        category: 'Layout',
        content: '<hr style="border: none; border-top: 1px solid #dddddd; margin: 20px 0;" />',
    })

    bm.add('spacer', {
        label: 'Spacer',
        category: 'Layout',
        content: '<div style="height: 30px;"></div>',
    })

    bm.add('heading', {
        label: 'Heading',
        category: 'Content',
        content: '<h1 style="margin: 0; padding: 10px 0; font-family: Arial, sans-serif; font-size: 28px; font-weight: bold; color: #333333;">Your Heading</h1>',
    })

    bm.add('text', {
        label: 'Text',
        category: 'Content',
        content: '<p style="margin: 0; padding: 10px 0; font-family: Arial, sans-serif; font-size: 14px; line-height: 1.6; color: #555555;">Write your content here. You can use variables like {{name}} and {{email}} for personalization.</p>',
    })

    bm.add('image', {
        label: 'Image',
        category: 'Content',
        content: { type: 'image', style: { 'max-width': '100%', height: 'auto', display: 'block', margin: '0 auto' } },
    })

    bm.add('button', {
        label: 'Button',
        category: 'Content',
        content: `<a href="#" data-gjs-custom-name="Button" style="display: inline-block; margin: 10px 0; padding: 14px 28px; background-color: #000000; color: #ffffff; text-decoration: none; border-radius: 6px; font-family: Arial, sans-serif; font-size: 14px; font-weight: bold;">Click Here</a>`,
    })

    bm.add('link', {
        label: 'Link',
        category: 'Content',
        content: '<a href="#" style="color: #0066cc; text-decoration: underline; font-family: Arial, sans-serif; font-size: 14px;">Link text</a>',
    })

    bm.add('list', {
        label: 'List',
        category: 'Content',
        content: `<ul style="padding-left: 20px; font-family: Arial, sans-serif; font-size: 14px; color: #555555; line-height: 1.8;">
            <li>Item one</li>
            <li>Item two</li>
            <li>Item three</li>
        </ul>`,
    })

    bm.add('header-block', {
        label: 'Header',
        category: 'Sections',
        content: `<div style="background-color: #000000; padding: 30px 20px; text-align: center;">
            <h1 style="margin: 0; font-family: Arial, sans-serif; font-size: 24px; color: #ffffff;">Company Name</h1>
        </div>`,
    })

    bm.add('footer-block', {
        label: 'Footer',
        category: 'Sections',
        content: `<div style="background-color: #f5f5f5; padding: 20px; text-align: center;">
            <p style="margin: 0 0 8px 0; font-family: Arial, sans-serif; font-size: 12px; color: #999999;">You received this email because you subscribed to our newsletter.</p>
            <a href="{{unsubscribe_url}}" style="font-family: Arial, sans-serif; font-size: 12px; color: #999999;">Unsubscribe</a>
        </div>`,
    })

    bm.add('cta-block', {
        label: 'CTA Section',
        category: 'Sections',
        content: `<div style="background-color: #f8f8f8; padding: 40px 20px; text-align: center;">
            <h2 style="margin: 0 0 10px 0; font-family: Arial, sans-serif; font-size: 22px; color: #333333;">Ready to get started?</h2>
            <p style="margin: 0 0 20px 0; font-family: Arial, sans-serif; font-size: 14px; color: #666666;">Join thousands of users who trust our platform.</p>
            <a href="#" data-gjs-custom-name="Button" style="display: inline-block; padding: 14px 32px; background-color: #000000; color: #ffffff; text-decoration: none; border-radius: 6px; font-family: Arial, sans-serif; font-size: 14px; font-weight: bold;">Get Started</a>
        </div>`,
    })

    bm.getAll().forEach((block: Block) => block.set('select', true))

    const instance = editor
    instance.on('load', () => {
        const doc = instance.Canvas.getDocument()
        if (!doc) return
        const style = doc.createElement('style')
        style.textContent = `
            img { max-width: 100%; }
            table { border-collapse: collapse; }
            div:empty, td:empty, th:empty {
                min-height: 60px;
                outline: 1px dashed rgba(0, 0, 0, 0.3);
                outline-offset: -1px;
            }
        `
        doc.head.appendChild(style)
    })

    if (props.modelValue) {
        loadContent(props.modelValue)
    }

    editor.on('update', scheduleEmit)
    editor.on('component:add', scheduleEmit)
    editor.on('component:remove', scheduleEmit)
})

function scheduleEmit() {
    if (!editor) return
    if (suppressEmit) {
        // Changes caused by a programmatic load/refresh are not user edits.
        if (emitTimer) {
            clearTimeout(emitTimer)
            emitTimer = null
        }
        return
    }
    userDirty = true
    if (emitTimer) clearTimeout(emitTimer)
    emitTimer = window.setTimeout(emitHtml, 300)
}

function emitHtml() {
    if (!editor) return
    if (suppressEmit) return
    const html = editor.getHtml()
    const css = editor.getCss()
    const raw = css ? `<style>${css}</style>${html}` : html
    if (raw === lastEmitted) return
    lastEmitted = raw
    emit('update:modelValue', raw)
}

function flush() {
    if (emitTimer) {
        clearTimeout(emitTimer)
        emitTimer = null
    }
    // Only write the canvas back to the model when the user actually edited
    // it. A plain open/tab-switch must not overwrite the untouched code with
    // GrapeJS's re-serialized version.
    if (!userDirty) return
    emitHtml()
}

function refresh() {
    editor?.refresh()
}

// Bring the canvas in line with externally edited HTML (for example edits
// made in the Code tab). Components are reloaded only when the content
// actually changed, so a plain tab switch keeps the canvas and its undo
// history intact.
function syncFrom(html: string) {
    if (!editor) return
    if (html === lastEmitted) {
        // Re-render without letting the canvas echo its (re-serialized) HTML
        // back over the untouched code.
        suppressEmit = true
        editor.refresh()
        if (emitTimer) {
            clearTimeout(emitTimer)
            emitTimer = null
        }
        window.setTimeout(() => {
            suppressEmit = false
        }, 400)
        return
    }
    loadContent(html)
}

function insertContent(html: string) {
    editor?.addComponents(html)
}

defineExpose({ flush, refresh, syncFrom, insertContent })

onBeforeUnmount(() => {
    flush()
    if (editor) {
        editor.destroy()
        editor = null
    }
})
</script>

<template>
    <div ref="editorContainer" class="grapesjs-editor" />
</template>

<style>
.grapesjs-editor {
    height: 100%;
}

/* Dark theme overrides */
.gjs-editor-cont,
.gjs-one-bg {
    background-color: #19191b !important;
}

.gjs-two-color {
    color: #b2b3bd !important;
}

.gjs-three-bg {
    background-color: #292a2e !important;
}

.gjs-four-color,
.gjs-four-color-h:hover {
    color: #ffffff !important;
}

.gjs-pn-panel {
    background-color: #19191b !important;
    border-color: #46484f !important;
}

/* Blocks */
.gjs-block {
    background-color: #292a2e !important;
    border: 1px solid #46484f !important;
    color: #b2b3bd !important;
    border-radius: 6px !important;
    min-height: auto !important;
    padding: 10px 8px !important;
    font-size: 11px !important;
}

.gjs-block:hover {
    border-color: #5f606a !important;
    color: #ffffff !important;
}

.gjs-block svg {
    fill: #b2b3bd !important;
}

.gjs-block:hover svg {
    fill: #ffffff !important;
}

.gjs-blocks-cs {
    background-color: #19191b !important;
}

/* Categories */
.gjs-category-title,
.gjs-layer-title,
.gjs-block-category .gjs-title {
    background-color: #292a2e !important;
    border-color: #46484f !important;
    color: #b2b3bd !important;
    font-size: 12px !important;
}

/* Style manager */
.gjs-sm-sector-title {
    background-color: #292a2e !important;
    color: #b2b3bd !important;
    border-color: #46484f !important;
    font-size: 12px !important;
}

.gjs-sm-sector .gjs-sm-properties {
    background-color: #19191b !important;
}

.gjs-sm-label {
    color: #6c6e79 !important;
    font-size: 11px !important;
}

.gjs-clm-tags {
    background-color: #19191b !important;
}

.gjs-clm-tag {
    background-color: #292a2e !important;
    color: #b2b3bd !important;
}

/* Fields */
.gjs-field {
    background-color: #111113 !important;
    border-color: #46484f !important;
    color: #ffffff !important;
    border-radius: 4px !important;
}

.gjs-field input,
.gjs-field select,
.gjs-field textarea {
    color: #ffffff !important;
}

.gjs-field-arrows {
    color: #6c6e79 !important;
}

.gjs-field-color-picker {
    border-radius: 3px !important;
}

/* Radio buttons */
.gjs-radio-item {
    background-color: #292a2e !important;
    border-color: #46484f !important;
    color: #b2b3bd !important;
}

.gjs-radio-item:hover {
    color: #ffffff !important;
}

.gjs-radio-item input:checked + .gjs-radio-item-label {
    background-color: #46484f !important;
    color: #ffffff !important;
}

/* Primary button */
.gjs-btn-prim {
    background-color: #00af7b !important;
    color: #ffffff !important;
    border-radius: 6px !important;
}

/* Panel buttons */
.gjs-pn-btn {
    color: #b2b3bd !important;
}

.gjs-pn-btn.gjs-pn-active {
    color: #ffffff !important;
}

/* Canvas */
.gjs-cv-canvas {
    background-color: #292a2e !important;
}

/* Selected component highlight */
.gjs-selected {
    outline: 2px solid #ffffff !important;
}

.gjs-toolbar {
    background-color: #292a2e !important;
    border: 1px solid #46484f !important;
    border-radius: 4px !important;
}

.gjs-toolbar-item {
    color: #b2b3bd !important;
}

.gjs-toolbar-item:hover {
    color: #ffffff !important;
}

/* Layers */
.gjs-layer {
    background-color: #19191b !important;
}

.gjs-layer-title {
    border-color: #46484f !important;
}

/* Trait manager */
.gjs-trt-trait {
    border-color: #46484f !important;
}

.gjs-trt-trait .gjs-label {
    color: #6c6e79 !important;
}

/* Modal */
.gjs-mdl-dialog {
    background-color: #19191b !important;
    border: 1px solid #46484f !important;
    border-radius: 8px !important;
}

.gjs-mdl-header {
    border-color: #46484f !important;
    color: #ffffff !important;
}

/* Hide device selector - not needed for email */
.gjs-pn-devices-c {
    display: none !important;
}

/* Scrollbars */
.gjs-editor-cont ::-webkit-scrollbar {
    width: 6px;
}

.gjs-editor-cont ::-webkit-scrollbar-track {
    background: #19191b;
}

.gjs-editor-cont ::-webkit-scrollbar-thumb {
    background: #46484f;
    border-radius: 3px;
}
</style>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import grapesjs, { type Editor } from 'grapesjs'
import newsletterPreset from 'grapesjs-preset-newsletter'
import 'grapesjs/dist/css/grapes.min.css'
import * as prettier from 'prettier/standalone'
import htmlPlugin from 'prettier/plugins/html'

const props = defineProps<{
    modelValue: string
}>()

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const editorContainer = ref<HTMLElement | null>(null)
let editor: Editor | null = null
let skipWatch = false

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
            }
        },
        canvas: {
            styles: [],
        },
        blockManager: {
            blocks: []
        },
        // Style manager sectors - email-safe properties
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

    // Setup blocks
    const bm = editor.BlockManager
    bm.getAll().reset()

    // Layout blocks
    bm.add('container', {
        label: 'Container',
        category: 'Layout',
        content: `<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; font-family: Arial, sans-serif;"></div>`,
    })

    bm.add('section', {
        label: 'Section',
        category: 'Layout',
        content: `<div style="padding: 20px; background-color: #ffffff;"></div>`,
    })

    bm.add('columns-2', {
        label: '2 Columns',
        category: 'Layout',
        content: `<table style="width: 100%; border-collapse: collapse;">
            <tr>
                <td style="width: 50%; vertical-align: top; padding: 10px;"></td>
                <td style="width: 50%; vertical-align: top; padding: 10px;"></td>
            </tr>
        </table>`,
    })

    bm.add('columns-3', {
        label: '3 Columns',
        category: 'Layout',
        content: `<table style="width: 100%; border-collapse: collapse;">
            <tr>
                <td style="width: 33.33%; vertical-align: top; padding: 10px;"></td>
                <td style="width: 33.33%; vertical-align: top; padding: 10px;"></td>
                <td style="width: 33.33%; vertical-align: top; padding: 10px;"></td>
            </tr>
        </table>`,
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

    // Content blocks
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
        content: `<table style="margin: 10px 0;">
            <tr>
                <td style="background-color: #000000; border-radius: 4px; padding: 0;">
                    <a href="#" style="display: inline-block; padding: 14px 28px; color: #ffffff; text-decoration: none; font-family: Arial, sans-serif; font-size: 14px; font-weight: bold;">Click Here</a>
                </td>
            </tr>
        </table>`,
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

    // Pre-built sections
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
            <table style="margin: 0 auto;">
                <tr>
                    <td style="background-color: #000000; border-radius: 4px; padding: 0;">
                        <a href="#" style="display: inline-block; padding: 14px 32px; color: #ffffff; text-decoration: none; font-family: Arial, sans-serif; font-size: 14px; font-weight: bold;">Get Started</a>
                    </td>
                </tr>
            </table>
        </div>`,
    })

    // Load initial content
    if (props.modelValue) {
        editor.setComponents(props.modelValue)
    }

    // Emit changes
    editor.on('component:update', emitHtml)
    editor.on('component:add', emitHtml)
    editor.on('component:remove', emitHtml)
    editor.on('component:styleUpdate', emitHtml)
})

async function emitHtml() {
    if (!editor || skipWatch) return
    const html = editor.getHtml()
    const css = editor.getCss()
    const raw = css ? `<style>${css}</style>${html}` : html

    let formatted = raw
    try {
        formatted = await prettier.format(raw, {
            parser: 'html',
            plugins: [htmlPlugin],
            printWidth: 120,
            tabWidth: 2,
        })
    } catch {
        // If formatting fails, use raw
    }

    skipWatch = true
    emit('update:modelValue', formatted)
    setTimeout(() => { skipWatch = false }, 200)
}

// Watch for external changes (e.g., switching from code tab)
watch(() => props.modelValue, (newVal) => {
    if (!editor || skipWatch) return
    editor.setComponents(newVal || '')
})

onBeforeUnmount(() => {
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
    background-color: #18181b !important;
}

.gjs-two-color {
    color: #a1a1aa !important;
}

.gjs-three-bg {
    background-color: #27272a !important;
}

.gjs-four-color,
.gjs-four-color-h:hover {
    color: #ffffff !important;
}

.gjs-pn-panel {
    background-color: #18181b !important;
    border-color: #3f3f46 !important;
}

/* Blocks */
.gjs-block {
    background-color: #27272a !important;
    border: 1px solid #3f3f46 !important;
    color: #a1a1aa !important;
    border-radius: 6px !important;
    min-height: auto !important;
    padding: 10px 8px !important;
    font-size: 11px !important;
}

.gjs-block:hover {
    border-color: #52525b !important;
    color: #ffffff !important;
}

.gjs-block svg {
    fill: #a1a1aa !important;
}

.gjs-block:hover svg {
    fill: #ffffff !important;
}

.gjs-blocks-cs {
    background-color: #18181b !important;
}

/* Categories */
.gjs-category-title,
.gjs-layer-title,
.gjs-block-category .gjs-title {
    background-color: #27272a !important;
    border-color: #3f3f46 !important;
    color: #a1a1aa !important;
    font-size: 12px !important;
}

/* Style manager */
.gjs-sm-sector-title {
    background-color: #27272a !important;
    color: #a1a1aa !important;
    border-color: #3f3f46 !important;
    font-size: 12px !important;
}

.gjs-sm-sector .gjs-sm-properties {
    background-color: #18181b !important;
}

.gjs-sm-label {
    color: #71717a !important;
    font-size: 11px !important;
}

.gjs-clm-tags {
    background-color: #18181b !important;
}

.gjs-clm-tag {
    background-color: #27272a !important;
    color: #a1a1aa !important;
}

/* Fields */
.gjs-field {
    background-color: #09090b !important;
    border-color: #3f3f46 !important;
    color: #ffffff !important;
    border-radius: 4px !important;
}

.gjs-field input,
.gjs-field select,
.gjs-field textarea {
    color: #ffffff !important;
}

.gjs-field-arrows {
    color: #71717a !important;
}

.gjs-field-color-picker {
    border-radius: 3px !important;
}

/* Radio buttons */
.gjs-radio-item {
    background-color: #27272a !important;
    border-color: #3f3f46 !important;
    color: #a1a1aa !important;
}

.gjs-radio-item:hover {
    color: #ffffff !important;
}

.gjs-radio-item input:checked + .gjs-radio-item-label {
    background-color: #3f3f46 !important;
    color: #ffffff !important;
}

/* Primary button */
.gjs-btn-prim {
    background-color: #ffffff !important;
    color: #09090b !important;
    border-radius: 6px !important;
}

/* Panel buttons */
.gjs-pn-btn {
    color: #a1a1aa !important;
}

.gjs-pn-btn.gjs-pn-active {
    color: #ffffff !important;
}

/* Canvas */
.gjs-cv-canvas {
    background-color: #27272a !important;
}

/* Selected component highlight */
.gjs-selected {
    outline: 2px solid #ffffff !important;
}

.gjs-toolbar {
    background-color: #27272a !important;
    border: 1px solid #3f3f46 !important;
    border-radius: 4px !important;
}

.gjs-toolbar-item {
    color: #a1a1aa !important;
}

.gjs-toolbar-item:hover {
    color: #ffffff !important;
}

/* Layers */
.gjs-layer {
    background-color: #18181b !important;
}

.gjs-layer-title {
    border-color: #3f3f46 !important;
}

/* Trait manager */
.gjs-trt-trait {
    border-color: #3f3f46 !important;
}

.gjs-trt-trait .gjs-label {
    color: #71717a !important;
}

/* Modal */
.gjs-mdl-dialog {
    background-color: #18181b !important;
    border: 1px solid #3f3f46 !important;
    border-radius: 8px !important;
}

.gjs-mdl-header {
    border-color: #3f3f46 !important;
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
    background: #18181b;
}

.gjs-editor-cont ::-webkit-scrollbar-thumb {
    background: #3f3f46;
    border-radius: 3px;
}
</style>

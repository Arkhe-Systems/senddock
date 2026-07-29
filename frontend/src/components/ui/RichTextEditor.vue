<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import { Bold, Italic, Heading2, List, ListOrdered, Link2, Quote } from 'lucide-vue-next'

const model = defineModel<string>({ default: '' })

const editor = useEditor({
    content: model.value,
    extensions: [
        StarterKit,
        Link.configure({ openOnClick: false, HTMLAttributes: { rel: 'noopener noreferrer nofollow' } }),
    ],
    editorProps: {
        attributes: {
            class: 'ProseMirror-content focus:outline-none min-h-[9rem] px-3 py-2 text-white',
        },
    },
    onUpdate: () => {
        const html = editor.value?.getHTML() ?? ''
        model.value = html === '<p></p>' ? '' : html
    },
})

watch(model, (val) => {
    if (!editor.value) return
    if ((val || '') !== editor.value.getHTML()) {
        editor.value.commands.setContent(val || '', { emitUpdate: false })
    }
})

onBeforeUnmount(() => editor.value?.destroy())

function toggleLink() {
    if (!editor.value) return
    if (editor.value.isActive('link')) {
        editor.value.chain().focus().unsetLink().run()
        return
    }
    const url = window.prompt('URL del enlace (https://…)')
    if (!url) return
    editor.value.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
}

const tools = [
    { name: 'bold', icon: Bold, title: 'Negrita', run: () => editor.value?.chain().focus().toggleBold().run() },
    { name: 'italic', icon: Italic, title: 'Cursiva', run: () => editor.value?.chain().focus().toggleItalic().run() },
    { name: 'heading', icon: Heading2, title: 'Encabezado', run: () => editor.value?.chain().focus().toggleHeading({ level: 2 }).run(), active: () => editor.value?.isActive('heading', { level: 2 }) },
    { name: 'bulletList', icon: List, title: 'Lista', run: () => editor.value?.chain().focus().toggleBulletList().run() },
    { name: 'orderedList', icon: ListOrdered, title: 'Lista numerada', run: () => editor.value?.chain().focus().toggleOrderedList().run() },
    { name: 'blockquote', icon: Quote, title: 'Cita', run: () => editor.value?.chain().focus().toggleBlockquote().run() },
    { name: 'link', icon: Link2, title: 'Enlace', run: toggleLink },
]

function isActive(t: typeof tools[number]) {
    return t.active ? t.active() : editor?.value?.isActive(t.name)
}
</script>

<template>
    <div class="border border-zinc-800 rounded-lg bg-zinc-900 overflow-hidden">
        <div class="flex flex-wrap gap-0.5 border-b border-zinc-800 px-1.5 py-1 bg-zinc-950/40">
            <button
                v-for="t in tools"
                :key="t.name"
                type="button"
                :title="t.title"
                @click="t.run()"
                :class="[
                    'p-1.5 rounded cursor-pointer transition',
                    isActive(t) ? 'bg-zinc-700 text-white' : 'text-zinc-400 hover:text-white hover:bg-zinc-800',
                ]">
                <component :is="t.icon" class="w-4 h-4" />
            </button>
        </div>
        <EditorContent :editor="editor" />
    </div>
</template>

<style scoped>
:deep(.ProseMirror-content) {
    line-height: 1.5;
}
:deep(.ProseMirror-content p) {
    margin: 0 0 0.5rem;
}
:deep(.ProseMirror-content p:last-child) {
    margin-bottom: 0;
}
:deep(.ProseMirror-content h2) {
    font-size: 1.15rem;
    font-weight: 600;
    margin: 0.5rem 0;
}
:deep(.ProseMirror-content ul) {
    list-style: disc;
    padding-left: 1.25rem;
    margin: 0 0 0.5rem;
}
:deep(.ProseMirror-content ol) {
    list-style: decimal;
    padding-left: 1.25rem;
    margin: 0 0 0.5rem;
}
:deep(.ProseMirror-content blockquote) {
    border-left: 3px solid #52525b;
    padding-left: 0.75rem;
    color: #a1a1aa;
    margin: 0 0 0.5rem;
}
:deep(.ProseMirror-content a) {
    color: #818cf8;
    text-decoration: underline;
}
:deep(.ProseMirror-content:focus) {
    outline: none;
}
</style>

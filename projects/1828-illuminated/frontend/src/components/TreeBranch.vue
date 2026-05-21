<script setup lang="ts">
// Recursive child of StudyTreePanel. One <li> with the node row plus a
// nested <ul> for children — recurses by referencing itself by component
// name (vue-router resolves the SFC by filename when <TreeBranch> appears).
import type { StudyNode } from '@/composables/useStudyTree'

const props = defineProps<{
  node: StudyNode
  childrenOf: (id: string | null) => StudyNode[]
  activeId: string | null
}>()
defineEmits<{
  (e: 'jump', node: StudyNode): void
  (e: 'remove', id: string): void
}>()

const KIND_GLYPH: Record<StudyNode['kind'], string> = {
  word: 'W', verse: 'v', chapter: '§', render: '↻', note: '✎',
}
const KIND_COLOR: Record<StudyNode['kind'], string> = {
  word: 'bg-amber-100 text-amber-800',
  verse: 'bg-sky-100 text-sky-800',
  chapter: 'bg-sky-100 text-sky-900',
  render: 'bg-violet-100 text-violet-800',
  note: 'bg-stone-100 text-stone-700',
}
</script>

<template>
  <li class="tree-branch">
    <div
      :class="[
        'flex items-baseline gap-2 px-2 py-1.5 rounded cursor-pointer group transition',
        props.activeId === props.node.id
          ? 'bg-amber-50 border-l-2 border-amber-500'
          : 'hover:bg-stone-100 border-l-2 border-transparent',
      ]"
      @click="$emit('jump', props.node)"
    >
      <span
        :class="['inline-block px-1.5 py-0.5 rounded text-[10px] font-mono leading-none', KIND_COLOR[props.node.kind]]"
        :title="props.node.kind"
      >{{ KIND_GLYPH[props.node.kind] }}</span>
      <span class="flex-1 font-serif text-stone-800 truncate">{{ props.node.label }}</span>
      <button
        class="opacity-0 group-hover:opacity-100 text-stone-400 hover:text-red-600 text-xs px-1"
        :title="`Remove ${props.node.label} and its descendants`"
        @click.stop="$emit('remove', props.node.id)"
      >×</button>
    </div>
    <ul
      v-if="props.childrenOf(props.node.id).length"
      class="pl-4 mt-1 space-y-1 border-l border-stone-200"
    >
      <TreeBranch
        v-for="child in props.childrenOf(props.node.id)"
        :key="child.id"
        :node="child"
        :children-of="props.childrenOf"
        :active-id="props.activeId"
        @jump="(n) => $emit('jump', n)"
        @remove="(id) => $emit('remove', id)"
      />
    </ul>
  </li>
</template>

<style scoped>
.tree-branch { font-variant-numeric: tabular-nums; }
</style>

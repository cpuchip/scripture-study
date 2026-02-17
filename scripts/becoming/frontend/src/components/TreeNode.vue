<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { FileTreeNode } from '../services/github'

export default defineComponent({
  name: 'TreeNode',
  props: {
    node: { type: Object as PropType<FileTreeNode>, required: true },
    depth: { type: Number, default: 0 },
    expandedDirs: { type: Object as PropType<Set<string>>, required: true },
    currentPath: { type: String, default: '' },
    readPaths: { type: Object as PropType<Set<string>>, required: true },
  },
  emits: ['toggle-dir', 'open-file'],
  computed: {
    isExpanded(): boolean {
      return this.expandedDirs.has(this.node.path)
    },
    isCurrent(): boolean {
      return this.currentPath === this.node.path
    },
    isRead(): boolean {
      return this.readPaths.has(this.node.path)
    },
    indent(): string {
      return `${this.depth * 12 + 8}px`
    },
    displayName(): string {
      const name = this.node.name
      if (this.node.type === 'directory') return name
      return name.replace(/\.md$/i, '')
    },
  },
})
</script>

<template>
  <div>
    <!-- Directory -->
    <div
      v-if="node.type === 'directory'"
      @click="$emit('toggle-dir', node.path)"
      class="flex items-center gap-1 px-2 py-1 cursor-pointer hover:bg-gray-100 text-xs text-gray-600"
      :style="{ paddingLeft: indent }"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 text-gray-400 transition-transform" :class="{ 'rotate-90': isExpanded }" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
      </svg>
      <span class="font-medium">{{ displayName }}</span>
    </div>

    <!-- File -->
    <div
      v-else
      @click="$emit('open-file', node.path)"
      class="flex items-center gap-1.5 px-2 py-1 cursor-pointer text-xs"
      :class="isCurrent ? 'bg-orange-50 text-orange-700 font-medium' : 'text-gray-700 hover:bg-gray-50'"
      :style="{ paddingLeft: indent }"
    >
      <span class="shrink-0" :class="isRead ? 'text-green-400' : 'text-gray-300'">
        {{ isRead ? '●' : '○' }}
      </span>
      <span class="truncate">{{ displayName }}</span>
    </div>

    <!-- Children -->
    <template v-if="node.type === 'directory' && isExpanded && node.children">
      <TreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :depth="depth + 1"
        :expanded-dirs="expandedDirs"
        :current-path="currentPath"
        :read-paths="readPaths"
        @toggle-dir="$emit('toggle-dir', $event)"
        @open-file="$emit('open-file', $event)"
      />
    </template>
  </div>
</template>

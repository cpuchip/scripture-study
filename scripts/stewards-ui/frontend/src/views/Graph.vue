<script setup lang="ts">
import { ref, onMounted, onUnmounted, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import cytoscape from 'cytoscape'
import type { Core } from 'cytoscape'
import { api, type GraphResp } from '@/api'

const router = useRouter()
const containerRef = useTemplateRef<HTMLDivElement>('container')
const error = ref('')
const loading = ref(false)
const stats = ref<{ nodes: number; edges: number } | null>(null)
let cy: Core | null = null

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data: GraphResp = await api.graphStudiesCitations(200)
    stats.value = { nodes: data.nodes.length, edges: data.edges.length }
    if (cy) cy.destroy()
    cy = cytoscape({
      container: containerRef.value!,
      elements: [
        ...data.nodes.map((n) => ({ data: { id: n.id, label: n.label, kind: n.kind } })),
        ...data.edges.map((e) => ({
          data: { id: `${e.source}->${e.target}`, source: e.source, target: e.target, weight: e.weight ?? 1 },
        })),
      ],
      style: [
        {
          selector: 'node',
          style: {
            'background-color': '#52525b',
            'border-color': '#71717a',
            'border-width': 1,
            label: 'data(label)',
            color: '#e4e4e7',
            'font-size': '10px',
            'text-valign': 'center',
            'text-halign': 'right',
            'text-margin-x': 6,
            width: 12,
            height: 12,
          },
        },
        {
          selector: 'node:selected',
          style: { 'background-color': '#10b981', 'border-color': '#34d399', width: 16, height: 16 },
        },
        {
          selector: 'edge',
          style: {
            width: 'mapData(weight, 1, 20, 0.5, 3)' as never,
            'line-color': '#3f3f46',
            'curve-style': 'bezier',
            'target-arrow-color': '#52525b',
            'target-arrow-shape': 'triangle',
            'arrow-scale': 0.6,
          },
        },
      ],
      layout: {
        name: 'cose',
        animate: false,
        nodeRepulsion: 8000,
        idealEdgeLength: 80,
      } as never,
    })
    cy.on('tap', 'node', (evt) => {
      const id = String(evt.target.id())
      router.push(`/studies/${encodeURIComponent(id)}`)
    })
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
onUnmounted(() => {
  if (cy) cy.destroy()
})
</script>

<template>
  <div class="space-y-3 h-[calc(100vh-9rem)] flex flex-col">
    <div class="flex items-baseline justify-between">
      <h2 class="text-2xl font-semibold tracking-tight">Graph</h2>
      <div class="text-xs text-zinc-500 flex items-center gap-3">
        <span v-if="loading">loading…</span>
        <span v-else-if="stats">
          {{ stats.nodes }} nodes · {{ stats.edges }} edges
          <span v-if="stats.edges === 0" class="text-amber-400">
            (no in-graph edges — substrate citations may not be populated yet)
          </span>
        </span>
        <button
          class="text-xs px-2 py-1 rounded border border-zinc-700 hover:bg-zinc-800"
          @click="load"
        >reload</button>
      </div>
    </div>

    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

    <div
      ref="container"
      class="flex-1 rounded-md border border-zinc-800 bg-zinc-950"
    ></div>

    <p class="text-xs text-zinc-500">
      Click a node to open its study. Layout uses Cytoscape's `cose` force-directed.
      Edges from <code class="font-mono">stewards.study_citations()</code> where target slug
      matches an in-graph node.
    </p>
  </div>
</template>

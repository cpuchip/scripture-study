<script setup lang="ts">
import { useWordData } from '@/composables/useWordData'
const data = useWordData()
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-10 prose prose-stone">
    <h1 class="text-3xl font-serif mb-4">About this tool</h1>

    <p class="text-stone-700 leading-relaxed">
      <strong>1828 Illuminated</strong> renders scripture with the 1828 Webster meaning frame applied. Words drift across two centuries; meanings shift; some shifts change how scripture reads. This tool surfaces words where 1828 and modern English diverge — so you can read with the meaning the original audience had.
    </p>

    <h2 class="text-2xl font-serif mt-8 mb-3">How the highlight list was built</h2>
    <p class="text-stone-700 leading-relaxed">
      Two independent signals converge on each tier word:
    </p>
    <ol class="list-decimal list-outside ml-6 my-3 text-stone-700 space-y-2">
      <li><strong>Study signal.</strong> Our own substrate study work (~91 study files in the workspace) is grep'd for Webster 1828 citations. Words lensed by 3+ studies with a "differ" framing land in Tier A. 2+ studies = Tier B. 1 study = Tier C.</li>
      <li><strong>Dictionary signal.</strong> Every canonical-scripture word ({{ data.tierWords.length > 0 ? '17,593' : '~17,500' }} unique across BoM/D&C/PGP/NT/OT) is intersected with the 1828 dictionary directly. For each match, the 1828 entry is scored by multi-sense richness, archaic markers, theological vocabulary, and length.</li>
    </ol>
    <p class="text-stone-700 leading-relaxed">
      Tier A++ words have BOTH signals strong. A+ has one strong. B has at least one solid signal. C is the broader candidate pool. D is the pure-archaic-marker pool.
    </p>

    <h2 class="text-2xl font-serif mt-8 mb-3">Honest cautions</h2>
    <ul class="list-disc list-outside ml-6 my-3 text-stone-700 space-y-2">
      <li><strong>1828 isn't always deeper.</strong> Some words mean their modern thing. A naive "everything has hidden meaning" frame misleads. The tier list is curation; treat it as starting points, not decoder rings.</li>
      <li><strong>The tool illuminates; it doesn't encode.</strong> A meaning shift opens conversation, not closes it. Good-faith reads still differ.</li>
      <li><strong>Scripture text is not bundled.</strong> Full passages always go through <a href="https://www.churchofjesuschrist.org/study/scriptures" target="_blank" rel="noopener" class="underline">churchofjesuschrist.org</a>. The brief excerpts here illustrate word usage for educational purposes.</li>
      <li><strong>Modern definitions come from the public Free Dictionary API.</strong> Where modern dictionaries don't carry a word, that absence itself is signal — the word has fallen out of contemporary usage.</li>
    </ul>

    <h2 class="text-2xl font-serif mt-8 mb-3">What's next</h2>
    <ul class="list-disc list-outside ml-6 my-3 text-stone-700 space-y-2">
      <li>Substrate-pipeline-generated per-word meaning-shift summaries (compare 1828 to modern in narrative form).</li>
      <li>A Restoration-era dictionary built from the gospel works themselves — scripture and General Conference talks as the defining corpus. Multi-level (elementary, 8th grade, college+). Working name: <em>Thummim 2026 Restoration Dictionary</em>.</li>
      <li>LLM-rendering of verses in modern English with 1828-faithful meanings — gated, user-supplied API key.</li>
      <li>Tablet-friendly presentation mode.</li>
    </ul>

    <h2 class="text-2xl font-serif mt-8 mb-3">Provenance</h2>
    <p class="text-stone-700 leading-relaxed text-sm">
      The full research trail behind the highlight list lives in <code class="text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded text-xs">research/gospel/1828/</code> of the workspace — five markdown documents covering every Webster 1828 citation across our studies (P1), the confirmed list (P2), the canon unique-word extraction (P3), the canon ∩ 1828 intersect (P4), and the final synthesis (P5). All reproducible from the Python scripts in <code class="text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded text-xs">.work/</code>.
    </p>
  </div>
</template>

# UI/UX Best Practices Knowledge Base

> For the **Becoming** app — Vue 3 + Tailwind CSS v4 + TypeScript
> A comprehensive reference for building thoughtful, accessible, high-quality interfaces.

---

## Table of Contents

1. [Dialog & Modal Patterns](#1-dialog--modal-patterns)
2. [Form UX](#2-form-ux)
3. [Navigation & Information Architecture](#3-navigation--information-architecture)
4. [Loading & State Communication](#4-loading--state-communication)
5. [Interaction Patterns](#5-interaction-patterns)
6. [Visual Design Principles (Tailwind-specific)](#6-visual-design-principles-tailwind-specific)
7. [Performance UX](#7-performance-ux)
8. [Anti-Patterns to Avoid](#8-anti-patterns-to-avoid)
9. [Vue 3 Specific Patterns](#9-vue-3-specific-patterns)
10. [Accessibility (a11y)](#10-accessibility-a11y)

---

## 1. Dialog & Modal Patterns

### 1.1 Why Native `<dialog>` Over `window.alert/confirm/prompt`

**Browser dialogs are hostile to users:**

| Problem | `window.alert()` / `confirm()` | Native `<dialog>` |
|---|---|---|
| **Blocks the main thread** | Yes — freezes all JS | No — non-blocking |
| **Stylable** | No — OS chrome only | Yes — full CSS control |
| **Accessible** | Minimal — no ARIA control | Full — role="dialog", aria-labelledby, focus trapping |
| **Content** | Plain text only | Any HTML — forms, icons, rich content |
| **Stacking** | Cannot layer or queue | Can compose and animate |
| **User trust** | Trained to dismiss without reading | Custom UI gets actual attention |
| **Mobile** | Inconsistent, ugly, sometimes clipped | Responsive, touch-friendly |
| **Testable** | Requires special handling in tests | Standard DOM testing |

**The native `<dialog>` element** provides built-in backdrop, `Escape` key closing, focus trapping (with `.showModal()`), and return value support — all without a library.

```vue
<!-- BaseDialog.vue -->
<template>
  <Teleport to="body">
    <dialog
      ref="dialogRef"
      class="rounded-xl border border-gray-200 bg-white p-0 shadow-xl backdrop:bg-black/50
             dark:border-gray-700 dark:bg-gray-800"
      @close="emit('close')"
      @cancel="onCancel"
    >
      <div class="p-6">
        <header v-if="$slots.header || title" class="mb-4 flex items-center justify-between">
          <h2 :id="titleId" class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            <slot name="header">{{ title }}</slot>
          </h2>
          <button
            @click="close()"
            class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600
                   dark:hover:bg-gray-700 dark:hover:text-gray-300"
            aria-label="Close dialog"
          >
            <XMarkIcon class="h-5 w-5" />
          </button>
        </header>

        <div class="text-gray-700 dark:text-gray-300">
          <slot />
        </div>

        <footer v-if="$slots.actions" class="mt-6 flex justify-end gap-3">
          <slot name="actions" />
        </footer>
      </div>
    </dialog>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, useId } from 'vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
  open: boolean
  title?: string
  preventCancel?: boolean
}>()

const emit = defineEmits<{
  close: []
  cancel: []
}>()

const dialogRef = ref<HTMLDialogElement>()
const titleId = useId()

function close() {
  dialogRef.value?.close()
}

function onCancel(e: Event) {
  if (props.preventCancel) {
    e.preventDefault()
    return
  }
  emit('cancel')
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    dialogRef.value?.showModal()
  } else {
    dialogRef.value?.close()
  }
})

// Handle initial open state
onMounted(() => {
  if (props.open) dialogRef.value?.showModal()
})

defineExpose({ close })
</script>
```

### 1.2 Proper Modal Patterns in Vue 3

**Teleport to body** — Prevents z-index conflicts and ensures modals render above all other content:

```vue
<Teleport to="body">
  <dialog ref="dialog">...</dialog>
</Teleport>
```

**Focus trapping** — The native `<dialog>` element with `.showModal()` provides focus trapping automatically. For custom implementations, trap focus within the modal:

```ts
// useFocusTrap.ts
export function useFocusTrap(containerRef: Ref<HTMLElement | undefined>) {
  const focusableSelector = [
    'a[href]', 'button:not([disabled])', 'input:not([disabled])',
    'select:not([disabled])', 'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])'
  ].join(', ')

  function trapFocus(e: KeyboardEvent) {
    if (e.key !== 'Tab') return
    const container = containerRef.value
    if (!container) return

    const focusable = [...container.querySelectorAll(focusableSelector)] as HTMLElement[]
    const first = focusable[0]
    const last = focusable[focusable.length - 1]

    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault()
      first.focus()
    }
  }

  onMounted(() => document.addEventListener('keydown', trapFocus))
  onUnmounted(() => document.removeEventListener('keydown', trapFocus))
}
```

**Scroll locking** — Prevent background scrolling when a modal is open:

```ts
// useScrollLock.ts
export function useScrollLock(isLocked: Ref<boolean>) {
  watch(isLocked, (locked) => {
    document.body.style.overflow = locked ? 'hidden' : ''
    // Prevent layout shift from scrollbar disappearing
    document.body.style.paddingRight = locked
      ? `${window.innerWidth - document.documentElement.clientWidth}px`
      : ''
  })

  onUnmounted(() => {
    document.body.style.overflow = ''
    document.body.style.paddingRight = ''
  })
}
```

**Escape key** — Native `<dialog>` fires a `cancel` event on Escape. Handle it:

```vue
<dialog @cancel="handleCancel">
  <!-- If you need to prevent close: e.preventDefault() in handler -->
</dialog>
```

### 1.3 When to Use Each Pattern

| Pattern | Use When | Example |
|---|---|---|
| **Modal dialog** | Action requires focused attention, is destructive, or needs input before proceeding | Deleting a practice, editing a reflection |
| **Slide-over panel** | Viewing/editing detail alongside a list; content is contextual | Practice detail panel, source reader sidebar |
| **Inline editing** | Simple single-field edits; low-friction updates | Renaming a task, editing a note inline |
| **Toast/snackbar** | Confirming a completed action; non-critical info | "Practice logged", "Note saved" |
| **Confirmation dialog** | Destructive + irreversible actions only | Deleting account data, bulk deletion |
| **Bottom sheet (mobile)** | Touch-friendly option selection on small screens | Action menu, filter selection |

### 1.4 Confirmation Dialogs — When They Help vs. Hurt

**Use confirmation when ALL of these are true:**
1. The action is **destructive** (data loss)
2. The action is **irreversible** (can't undo)
3. The **consequences are significant** (not trivial)

**Don't use confirmation when:**
- The action is **reversible** — offer **undo** instead (much better UX)
- The user **explicitly initiated** the action via a clear, well-labeled button
- It becomes **routine** — users learn to click "Yes" without reading

**The Undo Pattern — Almost Always Better:**

```vue
<!-- UndoToast.vue -->
<template>
  <Transition
    enter-from-class="translate-y-4 opacity-0"
    enter-active-class="transition duration-300"
    leave-to-class="translate-y-4 opacity-0"
    leave-active-class="transition duration-200"
  >
    <div
      v-if="visible"
      role="status"
      aria-live="polite"
      class="fixed bottom-4 left-1/2 z-50 flex -translate-x-1/2 items-center gap-3
             rounded-lg bg-gray-900 px-4 py-3 text-sm text-white shadow-lg
             dark:bg-gray-100 dark:text-gray-900"
    >
      <span>{{ message }}</span>
      <button
        @click="undo"
        class="font-semibold text-indigo-400 hover:text-indigo-300
               dark:text-indigo-600 dark:hover:text-indigo-700"
      >
        Undo
      </button>
      <!-- Visual countdown -->
      <div class="h-1 w-16 overflow-hidden rounded-full bg-white/20">
        <div
          class="h-full bg-indigo-400 transition-all ease-linear"
          :style="{ width: `${progress}%`, transitionDuration: `${duration}ms` }"
        />
      </div>
    </div>
  </Transition>
</template>
```

### 1.5 Non-Modal Alternatives

**Toasts** — Brief, auto-dismissing messages for confirmations:
- Keep under 5 seconds
- Allow manual dismiss
- Stack from bottom, limit to 3 visible
- Never use for errors (errors need persistent display)

**Inline confirmations** — Replace the trigger button with confirm/cancel:

```vue
<!-- InlineConfirm.vue -->
<template>
  <div class="inline-flex items-center gap-2">
    <template v-if="!confirming">
      <button @click="confirming = true" class="text-red-600 hover:text-red-700">
        Delete
      </button>
    </template>
    <template v-else>
      <span class="text-sm text-gray-600 dark:text-gray-400">Delete this?</span>
      <button @click="confirm" class="text-sm font-medium text-red-600">Yes, delete</button>
      <button @click="confirming = false" class="text-sm text-gray-500">Cancel</button>
    </template>
  </div>
</template>
```

---

## 2. Form UX

### 2.1 Inline Validation vs. Submit Validation

**Inline (field-level) validation — use when:**
- Format is specific and non-obvious (email, URL, date)
- Real-time feedback prevents wasted effort
- Fields have complex constraints the user might not know

**Submit (form-level) validation — use when:**
- Fields are simple (required text fields)
- Inter-field validation is needed (date ranges, password confirmation)
- Inline validation would be annoying (validating after every keystroke on a name field)

**Best practice: Validate on blur, re-validate on input after first error.**

```ts
// useFieldValidation.ts
export function useFieldValidation(
  value: Ref<string>,
  rules: Array<(v: string) => string | true>
) {
  const error = ref<string>('')
  const touched = ref(false)
  const dirty = ref(false)

  function validate(): boolean {
    for (const rule of rules) {
      const result = rule(value.value)
      if (result !== true) {
        error.value = result
        return false
      }
    }
    error.value = ''
    return true
  }

  // Validate on blur (first interaction)
  function onBlur() {
    touched.value = true
    validate()
  }

  // Re-validate on input only after first error shown
  watch(value, () => {
    dirty.value = true
    if (touched.value && error.value) {
      validate()
    }
  })

  return { error, touched, dirty, validate, onBlur }
}
```

### 2.2 Error Message Placement and Wording

**Placement rules:**
1. Place errors **directly below the field** they relate to (not in a summary at top)
2. Use `aria-describedby` to associate errors with fields for screen readers
3. Use `aria-invalid="true"` on the field when it has an error
4. Use red/destructive colors but **don't rely on color alone** — add an icon or text prefix

**Wording rules:**
- **Describe what to do**, not what went wrong: "Enter an email address" not "Invalid input"
- **Be specific**: "Password must be at least 8 characters" not "Password too short"
- **Be human**: "Looks like this field is empty" not "ERROR: FIELD_REQUIRED"
- **Never blame the user**: "Please enter a date" not "You entered an invalid date"

```vue
<!-- FormField.vue -->
<template>
  <div class="space-y-1.5">
    <label :for="id" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ label }}
      <span v-if="required" class="text-red-500" aria-hidden="true">*</span>
      <span v-if="required" class="sr-only">(required)</span>
    </label>

    <input
      :id="id"
      v-model="model"
      v-bind="$attrs"
      :aria-invalid="!!error"
      :aria-describedby="error ? `${id}-error` : hint ? `${id}-hint` : undefined"
      class="block w-full rounded-lg border px-3 py-2 text-sm transition-colors
             focus:outline-none focus:ring-2 focus:ring-offset-1
             dark:bg-gray-800 dark:text-gray-100"
      :class="[
        error
          ? 'border-red-300 focus:ring-red-500 dark:border-red-700'
          : 'border-gray-300 focus:ring-indigo-500 dark:border-gray-600'
      ]"
      @blur="emit('blur')"
    />

    <p v-if="hint && !error" :id="`${id}-hint`" class="text-xs text-gray-500 dark:text-gray-400">
      {{ hint }}
    </p>

    <p
      v-if="error"
      :id="`${id}-error`"
      role="alert"
      class="flex items-center gap-1 text-xs text-red-600 dark:text-red-400"
    >
      <ExclamationCircleIcon class="h-4 w-4 shrink-0" />
      {{ error }}
    </p>
  </div>
</template>
```

### 2.3 Progressive Disclosure in Forms

**Principle:** Show only what's needed. Reveal complexity as the user opts in.

**Techniques:**
- **Collapsible sections** — "Advanced options" collapsed by default
- **Conditional fields** — Show fields only when a prior choice makes them relevant
- **Smart defaults** — Pre-fill with sensible values; let the user override
- **Optional fields last** — Required fields first, optional grouped at the end
- **"Add more" patterns** — Start with one input, let users add additional entries

```vue
<!-- Progressive disclosure example -->
<template>
  <form @submit.prevent="save">
    <!-- Always visible: the essential fields -->
    <FormField v-model="practice.name" label="Practice name" required />
    <FormField v-model="practice.pillar" label="Pillar" required />

    <!-- Revealed on demand -->
    <details class="mt-4 rounded-lg border border-gray-200 dark:border-gray-700">
      <summary
        class="cursor-pointer select-none px-4 py-3 text-sm font-medium text-gray-600
               hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200"
      >
        Additional settings
      </summary>
      <div class="space-y-4 border-t border-gray-200 p-4 dark:border-gray-700">
        <FormField v-model="practice.frequency" label="Target frequency" />
        <FormField v-model="practice.reminder" label="Reminder time" type="time" />
        <FormField v-model="practice.notes" label="Notes" type="textarea" />
      </div>
    </details>
  </form>
</template>
```

### 2.4 Auto-Save vs. Explicit Save

| Approach | Best For | Tradeoffs |
|---|---|---|
| **Auto-save** | Notes, journals, text content, settings | Users never lose work. But: need conflict resolution, versioning, and clear save indicators |
| **Explicit save** | Forms with validation, multi-field submissions | User feels in control. But: risk of lost work on navigation |
| **Hybrid** | Complex editors | Auto-save drafts, explicit "publish/submit". Best of both worlds |

**Auto-save implementation guidelines:**
- Debounce saves (300-1000ms after last keystroke)
- Show save status: "Saving…" → "Saved" → "All changes saved"
- Handle offline: queue saves, sync when reconnected
- Never auto-save invalid state — validate before saving

```ts
// useAutoSave.ts
export function useAutoSave<T>(
  data: Ref<T>,
  saveFn: (data: T) => Promise<void>,
  options: { debounceMs?: number } = {}
) {
  const { debounceMs = 500 } = options
  const status = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
  const lastSaved = ref<Date>()

  const debouncedSave = useDebounceFn(async () => {
    try {
      status.value = 'saving'
      await saveFn(toRaw(data.value))
      status.value = 'saved'
      lastSaved.value = new Date()
      // Reset to idle after showing "Saved" briefly
      setTimeout(() => {
        if (status.value === 'saved') status.value = 'idle'
      }, 2000)
    } catch {
      status.value = 'error'
    }
  }, debounceMs)

  watch(data, debouncedSave, { deep: true })

  return { status, lastSaved }
}
```

### 2.5 Multi-Step Forms / Wizard Patterns

**When to use:** When a form has more than 5-7 fields, or when fields logically group into distinct phases.

**Rules:**
1. Show a **progress indicator** — steps completed, current step, steps remaining
2. Allow **backward navigation** — never trap users in a step
3. **Preserve state** across steps — don't lose data when navigating back
4. **Validate per step** — don't let users advance past invalid input
5. **Review step** before final submission — show a summary of all inputs

```vue
<!-- StepProgress.vue -->
<template>
  <nav aria-label="Form progress" class="mb-8">
    <ol class="flex items-center gap-2">
      <li
        v-for="(step, i) in steps"
        :key="step.id"
        class="flex items-center gap-2"
      >
        <div
          :class="[
            'flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium transition-colors',
            i < currentStep
              ? 'bg-indigo-600 text-white'
              : i === currentStep
                ? 'border-2 border-indigo-600 text-indigo-600'
                : 'border-2 border-gray-300 text-gray-400 dark:border-gray-600'
          ]"
          :aria-current="i === currentStep ? 'step' : undefined"
        >
          <CheckIcon v-if="i < currentStep" class="h-4 w-4" />
          <span v-else>{{ i + 1 }}</span>
        </div>
        <span
          class="text-sm"
          :class="i <= currentStep ? 'text-gray-900 dark:text-gray-100' : 'text-gray-400'"
        >
          {{ step.label }}
        </span>
        <ChevronRightIcon v-if="i < steps.length - 1" class="h-4 w-4 text-gray-300" />
      </li>
    </ol>
  </nav>
</template>
```

### 2.6 Accessible Form Labeling

**Every input must have a label.** No exceptions.

```vue
<!-- DO: Visible label -->
<label for="practice-name">Practice name</label>
<input id="practice-name" />

<!-- DO: SR-only label when visual label would be redundant -->
<label for="search" class="sr-only">Search practices</label>
<input id="search" placeholder="Search…" />

<!-- DO: aria-label for icon-only buttons -->
<button aria-label="Delete practice">
  <TrashIcon class="h-5 w-5" />
</button>

<!-- DON'T: Placeholder as label -->
<input placeholder="Practice name" /> <!-- ❌ No label! -->

<!-- DON'T: Wrapping without for/id when input is not a direct child -->
<label>Name</label>
<div><input /></div> <!-- ❌ Not associated! -->
```

**Group related fields:**

```vue
<fieldset>
  <legend class="text-sm font-medium text-gray-700 dark:text-gray-300">
    Reminder frequency
  </legend>
  <div class="mt-2 space-y-2">
    <label class="flex items-center gap-2">
      <input type="radio" name="freq" value="daily" />
      <span>Daily</span>
    </label>
    <label class="flex items-center gap-2">
      <input type="radio" name="freq" value="weekly" />
      <span>Weekly</span>
    </label>
  </div>
</fieldset>
```

---

## 3. Navigation & Information Architecture

### 3.1 Sidebar Navigation Patterns

**For the Becoming app's structure (Today, Practices, Tasks, Reflections, Reader, Flashcards):**

```vue
<!-- AppSidebar.vue -->
<template>
  <nav
    :class="[
      'flex flex-col border-r border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-900',
      collapsed ? 'w-16' : 'w-64'
    ]"
    aria-label="Main navigation"
  >
    <!-- Logo / App name -->
    <div class="flex h-16 items-center px-4">
      <span v-if="!collapsed" class="text-lg font-bold">Becoming</span>
    </div>

    <!-- Navigation items -->
    <ul class="flex-1 space-y-1 px-2 py-4" role="list">
      <li v-for="item in navItems" :key="item.to">
        <RouterLink
          :to="item.to"
          :class="[
            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
            isActive(item.to)
              ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300'
              : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'
          ]"
          :aria-current="isActive(item.to) ? 'page' : undefined"
        >
          <component :is="item.icon" class="h-5 w-5 shrink-0" />
          <span v-if="!collapsed">{{ item.label }}</span>
          <span v-if="item.badge && !collapsed" class="ml-auto rounded-full bg-indigo-100
                 px-2 py-0.5 text-xs text-indigo-700 dark:bg-indigo-900 dark:text-indigo-300">
            {{ item.badge }}
          </span>
        </RouterLink>
      </li>
    </ul>

    <!-- Collapse toggle -->
    <button
      @click="collapsed = !collapsed"
      class="m-2 rounded-lg p-2 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800"
      :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
    >
      <ChevronLeftIcon :class="['h-5 w-5 transition-transform', collapsed && 'rotate-180']" />
    </button>
  </nav>
</template>
```

**Sidebar design principles:**
- Use `aria-current="page"` on the active link
- Icons provide recognition even when collapsed
- Badge counts add value without clutter
- Collapsible sidebar respects screen real estate
- Group navigation items logically (primary actions, secondary sections)

### 3.2 Breadcrumbs and Wayfinding

**Use breadcrumbs when:**
- Content has hierarchical depth (Reader → Source → File → Section)
- Users arrive via deep links and need context
- The sidebar tree alone isn't sufficient orientation

```vue
<!-- Breadcrumbs.vue -->
<template>
  <nav aria-label="Breadcrumb" class="mb-4">
    <ol class="flex items-center gap-1.5 text-sm text-gray-500">
      <li v-for="(crumb, i) in crumbs" :key="crumb.path" class="flex items-center gap-1.5">
        <ChevronRightIcon v-if="i > 0" class="h-3.5 w-3.5 text-gray-400" aria-hidden="true" />
        <RouterLink
          v-if="i < crumbs.length - 1"
          :to="crumb.path"
          class="hover:text-gray-700 dark:hover:text-gray-300"
        >
          {{ crumb.label }}
        </RouterLink>
        <span v-else class="font-medium text-gray-900 dark:text-gray-100" aria-current="page">
          {{ crumb.label }}
        </span>
      </li>
    </ol>
  </nav>
</template>
```

### 3.3 Deep Linking and URL-Driven State

**Principle:** The URL should be the source of truth for view state. Every meaningful view state should be bookmark-able and shareable.

**What belongs in the URL:**
- Current page/route (`/practices`, `/reader/1`)
- Active filters (`?pillar=spiritual&status=active`)
- Selected item (`?f=path/to/file.md`)
- Scroll anchor (`#section-heading`)
- View mode (`?view=grid`)

**What does NOT belong in the URL:**
- Transient UI state (whether a dropdown is open)
- Form input in progress
- Toast messages
- Modal open/closed state (unless the modal IS the page content)

```ts
// useUrlState.ts — Sync reactive state with URL query parameters
export function useUrlState<T extends Record<string, string>>(defaults: T) {
  const route = useRoute()
  const router = useRouter()

  const state = reactive({ ...defaults })

  // Read from URL on init
  for (const key of Object.keys(defaults)) {
    const urlValue = route.query[key]
    if (typeof urlValue === 'string') {
      (state as any)[key] = urlValue
    }
  }

  // Write to URL on change
  watch(
    () => ({ ...state }),
    (newState) => {
      const query: Record<string, string> = {}
      for (const [key, value] of Object.entries(newState)) {
        if (value !== defaults[key as keyof T] && value !== '') {
          query[key] = value as string
        }
      }
      router.replace({ query })
    },
    { deep: true }
  )

  return state
}
```

### 3.4 Tab Patterns

**Use tabs when:**
- Content variants share the same context (different views of the same data)
- The user is comparing or switching between closely related content
- There are 2-7 options (more than 7 → use a different pattern)

**Use separate pages when:**
- Content is independently meaningful
- Users might deep-link to a specific section
- Content is long enough to warrant its own URL

**Accessible tabs:**

```vue
<!-- TabGroup.vue -->
<template>
  <div>
    <div role="tablist" :aria-label="label" class="flex gap-1 border-b border-gray-200 dark:border-gray-700">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        role="tab"
        :id="`tab-${tab.id}`"
        :aria-selected="activeTab === tab.id"
        :aria-controls="`panel-${tab.id}`"
        :tabindex="activeTab === tab.id ? 0 : -1"
        @click="activeTab = tab.id"
        @keydown="handleTabKeydown($event, tab.id)"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors -mb-px border-b-2',
          activeTab === tab.id
            ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
            : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>
    <div
      v-for="tab in tabs"
      :key="tab.id"
      role="tabpanel"
      :id="`panel-${tab.id}`"
      :aria-labelledby="`tab-${tab.id}`"
      v-show="activeTab === tab.id"
      :tabindex="0"
      class="py-4 focus:outline-none"
    >
      <slot :name="tab.id" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  tabs: Array<{ id: string; label: string }>
  label: string
}>()

const activeTab = defineModel<string>({ default: '' })

// Arrow key navigation between tabs (ARIA Tabs pattern)
function handleTabKeydown(e: KeyboardEvent, currentId: string) {
  const ids = props.tabs.map(t => t.id)
  const currentIndex = ids.indexOf(currentId)
  let nextIndex: number | null = null

  if (e.key === 'ArrowRight') nextIndex = (currentIndex + 1) % ids.length
  else if (e.key === 'ArrowLeft') nextIndex = (currentIndex - 1 + ids.length) % ids.length
  else if (e.key === 'Home') nextIndex = 0
  else if (e.key === 'End') nextIndex = ids.length - 1

  if (nextIndex !== null) {
    e.preventDefault()
    activeTab.value = ids[nextIndex]
    document.getElementById(`tab-${ids[nextIndex]}`)?.focus()
  }
}
</script>
```

### 3.5 Mobile-First Responsive Navigation

**Pattern: Top bar + bottom navigation on mobile, sidebar on desktop.**

```vue
<!-- MobileNav.vue -->
<template>
  <!-- Bottom tab bar: only on small screens -->
  <nav
    class="fixed inset-x-0 bottom-0 z-40 border-t border-gray-200 bg-white
           pb-safe md:hidden dark:border-gray-700 dark:bg-gray-900"
    aria-label="Mobile navigation"
  >
    <ul class="flex justify-around py-2" role="list">
      <li v-for="item in mobileNavItems" :key="item.to">
        <RouterLink
          :to="item.to"
          class="flex flex-col items-center gap-0.5 text-xs"
          :class="isActive(item.to) ? 'text-indigo-600' : 'text-gray-500'"
          :aria-current="isActive(item.to) ? 'page' : undefined"
        >
          <component :is="item.icon" class="h-6 w-6" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </li>
    </ul>
  </nav>

  <!-- Add padding to main content to account for bottom nav -->
  <!-- <main class="pb-20 md:pb-0"> -->
</template>
```

**Key responsive principles:**
- Touch targets: minimum 44×44px (48×48px preferred) on mobile
- Bottom navigation: 4-5 items maximum; prioritize by frequency of use
- Hamburger menu: only when you truly can't fit nav items otherwise — it hides discoverability
- `pb-safe` / `env(safe-area-inset-bottom)` for devices with home indicators

---

## 4. Loading & State Communication

### 4.1 Skeleton Screens vs. Spinners vs. Progress Bars

| Pattern | Use When | Duration |
|---|---|---|
| **Skeleton screen** | Layout is predictable; content is loading | 200ms – 3s |
| **Spinner** | Compact space; action triggered by user; layout unknown | 200ms – 5s |
| **Progress bar** | Operation has measurable progress (file upload, multi-step) | >2s |
| **Nothing (no indicator)** | Response will be under ~200ms | <200ms |

**Rule: Never show a loading indicator for less than 200ms.** It creates a flash that feels slower than showing nothing.

```vue
<!-- SkeletonCard.vue -->
<template>
  <div class="animate-pulse space-y-3 rounded-xl border border-gray-200 p-4
              dark:border-gray-700">
    <div class="h-4 w-3/4 rounded bg-gray-200 dark:bg-gray-700" />
    <div class="h-3 w-full rounded bg-gray-200 dark:bg-gray-700" />
    <div class="h-3 w-5/6 rounded bg-gray-200 dark:bg-gray-700" />
    <div class="flex gap-2">
      <div class="h-6 w-16 rounded-full bg-gray-200 dark:bg-gray-700" />
      <div class="h-6 w-20 rounded-full bg-gray-200 dark:bg-gray-700" />
    </div>
  </div>
</template>
```

```ts
// useDelayedLoading — avoid flash of loading state
export function useDelayedLoading(isLoading: Ref<boolean>, delayMs = 200) {
  const showLoading = ref(false)
  let timeout: ReturnType<typeof setTimeout>

  watch(isLoading, (loading) => {
    if (loading) {
      timeout = setTimeout(() => { showLoading.value = true }, delayMs)
    } else {
      clearTimeout(timeout)
      showLoading.value = false
    }
  }, { immediate: true })

  return showLoading
}
```

### 4.2 Optimistic UI Updates

**Principle:** Show the result immediately, then reconcile with the server. Roll back on failure.

**Use for:** Toggling states, adding items to lists, incrementing counts, marking tasks complete.

**Don't use for:** Financial transactions, destructive actions, operations requiring server-generated data.

```ts
// Optimistic practice logging example
async function logPractice(practiceId: string) {
  // 1. Immediately update local state
  const previousState = { ...practiceMap.value[practiceId] }
  practiceMap.value[practiceId].loggedToday = true
  practiceMap.value[practiceId].streak += 1

  // 2. Show success toast immediately
  toast.show('Practice logged!')

  try {
    // 3. Sync with server
    await api.logPractice(practiceId)
  } catch (error) {
    // 4. Roll back on failure
    practiceMap.value[practiceId] = previousState
    toast.show('Failed to log practice. Please try again.', { type: 'error' })
  }
}
```

### 4.3 Empty States

**Empty states are opportunities, not dead ends.** They should:
1. Explain *why* it's empty
2. Guide the user to take action
3. Feel intentional, not broken

```vue
<!-- EmptyState.vue -->
<template>
  <div class="flex flex-col items-center justify-center py-12 text-center">
    <div class="mb-4 rounded-2xl bg-gray-100 p-4 dark:bg-gray-800">
      <component :is="icon" class="h-10 w-10 text-gray-400" />
    </div>
    <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
      {{ title }}
    </h3>
    <p class="mt-1 max-w-sm text-sm text-gray-500 dark:text-gray-400">
      {{ description }}
    </p>
    <button
      v-if="actionLabel"
      @click="emit('action')"
      class="mt-6 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white
             hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500
             focus:ring-offset-2 dark:focus:ring-offset-gray-900"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>
```

**Empty state examples for Becoming:**

| Page | Empty Title | Description | Action |
|---|---|---|---|
| Practices | No practices yet | Start building habits by adding your first practice. | + Add practice |
| Today (no logs) | A fresh start | You haven't logged any practices today. What will you work on? | Log a practice |
| Reflections | Your story begins | Capture your first reflection — even a sentence counts. | Write reflection |
| Flashcards | Nothing to review | Add scriptures or quotes you want to memorize. | Create flashcard |
| Search (no results) | No matches found | Try different keywords or check your filters. | Clear filters |

### 4.4 Error States

**Error messages should be:**
1. **Recovery-focused** — Tell users what to DO, not just what went wrong
2. **Honest** — Don't hide the error behind vague language
3. **Contextual** — Show the error near where it occurred
4. **Non-destructive** — Preserve user input; never clear a form on error

```vue
<!-- ErrorState.vue -->
<template>
  <div
    role="alert"
    class="rounded-xl border border-red-200 bg-red-50 p-6 text-center
           dark:border-red-900 dark:bg-red-950"
  >
    <ExclamationTriangleIcon class="mx-auto h-8 w-8 text-red-400" />
    <h3 class="mt-3 text-lg font-medium text-red-800 dark:text-red-200">
      {{ title }}
    </h3>
    <p class="mt-1 text-sm text-red-600 dark:text-red-400">
      {{ message }}
    </p>
    <div class="mt-4 flex justify-center gap-3">
      <button
        v-if="retryable"
        @click="emit('retry')"
        class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
      >
        Try again
      </button>
      <button
        v-if="dismissable"
        @click="emit('dismiss')"
        class="rounded-lg px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-100
               dark:text-red-400 dark:hover:bg-red-900"
      >
        Dismiss
      </button>
    </div>
  </div>
</template>
```

**Error hierarchies:**
- **Field-level error**: Inline, below the field. For validation issues.
- **Section-level error**: Banner at top of the section. For API errors affecting a group.
- **Page-level error**: Full error state replacing content. For totally failed loads.
- **Global error**: Toast/banner at top of app. For network/auth issues.

### 4.5 Partial Loading

**Load what you can, show what you have.** Don't make the entire page wait for the slowest request.

```vue
<!-- DashboardPage.vue — Independent sections load independently -->
<template>
  <div class="space-y-6">
    <!-- Fast: loaded from local cache -->
    <TodaySummary />

    <!-- Medium: small API call -->
    <Suspense>
      <PracticesList />
      <template #fallback>
        <SkeletonCard :lines="4" />
      </template>
    </Suspense>

    <!-- Slow: heavy computation -->
    <Suspense>
      <StreakChart />
      <template #fallback>
        <SkeletonCard :lines="6" />
      </template>
    </Suspense>
  </div>
</template>
```

---

## 5. Interaction Patterns

### 5.1 Drag and Drop — Accessibility Considerations

**Drag and drop is a enhancement, never a requirement.** Every drag-and-drop action must have a keyboard/screen-reader alternative.

**Accessible alternatives:**
- Arrow buttons to reorder items
- "Move to…" dropdown/modal
- Number input for position

```vue
<!-- ReorderableList.vue -->
<template>
  <ul role="list" ref="listRef">
    <li
      v-for="(item, index) in items"
      :key="item.id"
      draggable="true"
      @dragstart="onDragStart(index)"
      @dragover.prevent="onDragOver(index)"
      @drop="onDrop(index)"
      class="group flex items-center gap-3 rounded-lg border border-transparent p-3
             hover:bg-gray-50 dark:hover:bg-gray-800"
      :class="{ 'border-indigo-300 bg-indigo-50': dragOverIndex === index }"
    >
      <!-- Drag handle -->
      <button
        class="cursor-grab text-gray-400 hover:text-gray-600 active:cursor-grabbing"
        aria-label="Drag to reorder"
      >
        <Bars3Icon class="h-5 w-5" />
      </button>

      <span class="flex-1">{{ item.name }}</span>

      <!-- Keyboard reorder alternative -->
      <div class="flex gap-1 opacity-0 group-hover:opacity-100 group-focus-within:opacity-100">
        <button
          @click="moveItem(index, index - 1)"
          :disabled="index === 0"
          aria-label="Move up"
          class="rounded p-1 hover:bg-gray-200 disabled:opacity-30
                 dark:hover:bg-gray-700"
        >
          <ChevronUpIcon class="h-4 w-4" />
        </button>
        <button
          @click="moveItem(index, index + 1)"
          :disabled="index === items.length - 1"
          aria-label="Move down"
          class="rounded p-1 hover:bg-gray-200 disabled:opacity-30
                 dark:hover:bg-gray-700"
        >
          <ChevronDownIcon class="h-4 w-4" />
        </button>
      </div>
    </li>
  </ul>
</template>
```

### 5.2 Keyboard Navigation and Shortcuts

**Essential keyboard patterns:**

| Context | Keys | Action |
|---|---|---|
| **Global** | `Ctrl+K` / `Cmd+K` | Command palette / search |
| **Global** | `?` | Show keyboard shortcuts help |
| **Lists** | `↑` / `↓` | Navigate items |
| **Lists** | `Enter` | Open/select item |
| **Lists** | `Delete` / `Backspace` | Delete with undo |
| **Modals** | `Escape` | Close |
| **Forms** | `Tab` / `Shift+Tab` | Navigate fields |
| **Forms** | `Ctrl+Enter` | Submit |
| **Editor** | `Ctrl+S` | Save |

```ts
// useKeyboardShortcut.ts
export function useKeyboardShortcut(
  key: string,
  handler: () => void,
  options: {
    ctrl?: boolean
    meta?: boolean
    shift?: boolean
    prevent?: boolean
    enabled?: Ref<boolean>
  } = {}
) {
  function onKeydown(e: KeyboardEvent) {
    if (options.enabled && !options.enabled.value) return
    if (options.ctrl && !e.ctrlKey && !e.metaKey) return
    if (options.meta && !e.metaKey) return
    if (options.shift && !e.shiftKey) return

    // Don't trigger shortcuts when typing in inputs
    const target = e.target as HTMLElement
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)) return
    if (target.isContentEditable) return

    if (e.key.toLowerCase() === key.toLowerCase()) {
      if (options.prevent !== false) e.preventDefault()
      handler()
    }
  }

  onMounted(() => document.addEventListener('keydown', onKeydown))
  onUnmounted(() => document.removeEventListener('keydown', onKeydown))
}
```

### 5.3 Touch Targets and Mobile Interactions

**Minimum sizes:**
- **Touch target:** 44×44px minimum (WCAG), 48×48px recommended (Material Design)
- **Spacing between targets:** at least 8px gap to prevent mis-taps

**Tailwind helper:**

```css
/* In your Tailwind base layer */
@layer base {
  .touch-target {
    @apply relative min-h-11 min-w-11;
  }

  /* Expand hit area without changing visual size */
  .touch-target-expanded::after {
    content: '';
    @apply absolute -inset-2;
  }
}
```

**Swipe actions on mobile (use sparingly):**
- Swipe-to-delete should always have undo
- Swipe actions must have visible button alternatives
- Never make swipe the *only* way to access an action

### 5.4 Hover States and Progressive Reveal

**Hover is an enhancement, not a requirement.** Touch devices have no hover. Everything revealed on hover must be accessible another way.

```vue
<!-- Card with progressive reveal -->
<template>
  <div class="group rounded-xl border border-gray-200 p-4 transition-shadow
              hover:shadow-md dark:border-gray-700">
    <h3>{{ title }}</h3>
    <p class="text-sm text-gray-600">{{ description }}</p>

    <!-- Actions: visible on hover/focus-within, always visible on touch -->
    <div class="mt-3 flex gap-2 opacity-0 transition-opacity
                group-hover:opacity-100 group-focus-within:opacity-100
                touch-device:opacity-100">
      <button class="text-sm text-indigo-600">Edit</button>
      <button class="text-sm text-red-600">Delete</button>
    </div>
  </div>
</template>
```

**Detect touch devices (for showing actions by default):**

```css
/* Tailwind v4 — hover-capable media query */
@media (hover: none) {
  .touch-device\:opacity-100 {
    opacity: 1;
  }
}
```

Or use `@custom-variant` in Tailwind v4:

```css
@custom-variant touch (@media (hover: none));
```

### 5.5 Undo/Redo as an Alternative to Confirmation

**The undo pattern is superior to "Are you sure?" because:**
1. It doesn't interrupt the user's flow
2. It assumes the user meant what they did (respects agency)
3. It provides a safety net without friction
4. It trains better than pavlovian "click yes" responses

```ts
// useUndoAction.ts
export function useUndoAction<T>(options: {
  doAction: () => Promise<T>
  undoAction: (result: T) => Promise<void>
  message: string
  timeoutMs?: number
}) {
  const toast = useToast()
  const { timeoutMs = 5000 } = options

  async function execute() {
    const result = await options.doAction()

    const { dismiss } = toast.show(options.message, {
      duration: timeoutMs,
      action: {
        label: 'Undo',
        onClick: async () => {
          dismiss()
          await options.undoAction(result)
          toast.show('Action undone')
        }
      }
    })
  }

  return { execute }
}
```

---

## 6. Visual Design Principles (Tailwind-Specific)

### 6.1 Color Contrast and Accessibility

**WCAG Requirements:**
- **AA (minimum):** 4.5:1 for normal text, 3:1 for large text (18px+ or 14px+ bold)
- **AAA (enhanced):** 7:1 for normal text, 4.5:1 for large text
- **UI components:** 3:1 against adjacent colors (borders, icons, focus rings)

**Tailwind colors that typically pass AA on white:**
- Text: `gray-700` and darker (gray-800, gray-900)
- Subtle text: `gray-600` (passes at 14px+, test carefully)
- Links/interactive: `indigo-600`, `blue-600` — NOT `indigo-400` or `blue-400`
- Error text: `red-600` (dark enough) — NOT `red-400`

**On dark backgrounds (gray-900):**
- Text: `gray-100`, `gray-200`
- Subtle: `gray-300`, `gray-400`
- Links: `indigo-400`, `blue-400`

**Never rely on color alone to convey information.** Always pair with:
- Icons (error icon + red text)
- Text labels ("Required" not just red asterisk)
- Patterns/shapes (strikethrough for excluded filters)

### 6.2 Dark Mode Implementation with Tailwind v4

**Tailwind v4** uses CSS `@theme` and CSS custom properties natively.

**System preference + manual toggle pattern:**

```ts
// useDarkMode.ts
export function useDarkMode() {
  const mode = useLocalStorage<'system' | 'light' | 'dark'>('theme', 'system')

  const isDark = computed(() => {
    if (mode.value === 'dark') return true
    if (mode.value === 'light') return false
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  // Apply class to <html>
  watchEffect(() => {
    document.documentElement.classList.toggle('dark', isDark.value)
  })

  // Listen for system changes when in system mode
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', () => {
    if (mode.value === 'system') {
      document.documentElement.classList.toggle('dark', mediaQuery.matches)
    }
  })

  function toggle() {
    mode.value = isDark.value ? 'light' : 'dark'
  }

  function setSystem() {
    mode.value = 'system'
  }

  return { mode, isDark, toggle, setSystem }
}
```

**Dark mode design rules:**
1. **Don't just invert.** Dark backgrounds should be `gray-900` or `gray-950`, not pure `#000`
2. **Reduce contrast slightly.** Use `gray-100` for text, not pure white — less eye strain
3. **Elevate with lightness.** In dark mode, "higher" surfaces are lighter (opposite of light mode shadows)
4. **Muted colors.** Saturated colors on dark backgrounds are harsh — use the 300-400 range
5. **Shadows become glows.** Traditional shadows disappear on dark backgrounds; consider subtle border or glow effects instead

```vue
<!-- Card with proper dark mode -->
<template>
  <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm
              dark:border-gray-700 dark:bg-gray-800 dark:shadow-none">
    <h3 class="text-gray-900 dark:text-gray-100">...</h3>
    <p class="text-gray-600 dark:text-gray-400">...</p>
    <span class="text-indigo-600 dark:text-indigo-400">...</span>
  </div>
</template>
```

### 6.3 Typography Hierarchy and Readability

**Establish a clear type scale and stick to it:**

| Level | Tailwind | Use |
|---|---|---|
| Page title | `text-2xl font-bold` or `text-3xl font-bold` | One per page |
| Section heading | `text-xl font-semibold` | Major sections |
| Card/subsection title | `text-lg font-semibold` | Cards, panels |
| Body | `text-base` (16px) | Default text |
| Supporting text | `text-sm text-gray-600` | Descriptions, metadata |
| Caption/label | `text-xs font-medium text-gray-500` | Form labels, timestamps |

**Readability rules:**
- **Line length:** 45-75 characters (use `max-w-prose` — sets `65ch`)
- **Line height:** 1.5-1.75 for body text (`leading-relaxed` or `leading-7`)
- **Paragraph spacing:** `space-y-4` between paragraphs
- **Font weight contrast:** Use weight to create hierarchy, not just size
- **Don't use light font weights** for body text on screens — `font-light` is hard to read

```vue
<!-- Good reading layout for study/journal content -->
<article class="prose prose-gray mx-auto max-w-prose dark:prose-invert">
  <!-- @tailwindcss/typography handles everything inside -->
</article>
```

### 6.4 Consistent Spacing and Visual Rhythm

**Use Tailwind's spacing scale consistently.** Pick a rhythm and stick to it.

**Spacing hierarchy:**

| Relationship | Spacing | Example |
|---|---|---|
| Related items in a group | `gap-1` to `gap-2` (4-8px) | Icon + label, badge pills |
| Items in a list | `gap-3` or `space-y-3` (12px) | List items, form fields |
| Sections within a card | `space-y-4` (16px) | Card header / body / footer |
| Cards in a layout | `gap-4` to `gap-6` (16-24px) | Card grid |
| Major page sections | `space-y-8` to `space-y-12` (32-48px) | Page sections |
| Page padding | `p-4` mobile, `p-6` tablet, `p-8` desktop | Overall padding |

**The 4px grid:** Tailwind's spacing scale is based on 4px increments. Stick to even values (space-1 = 4px, space-2 = 8px, space-3 = 12px, space-4 = 16px).

### 6.5 Focus Indicators and Accessibility

**Never remove focus outlines.** Style them instead.

```css
/* Tailwind v4 — Global focus style */
@layer base {
  :focus-visible {
    @apply outline-2 outline-offset-2 outline-indigo-500;
  }

  /* Remove outline for mouse users, keep for keyboard */
  :focus:not(:focus-visible) {
    outline: none;
  }
}
```

**Focus indicator rules:**
- Must have **3:1 contrast ratio** against adjacent colors
- Must be visible in both light and dark mode
- `outline` is preferred over `box-shadow` (outlines aren't clipped by `overflow: hidden`)
- Use `focus-visible` (not `focus`) to show only on keyboard interaction

```vue
<!-- Button with clear focus indicator -->
<button
  class="rounded-lg bg-indigo-600 px-4 py-2 text-white
         hover:bg-indigo-700
         focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2
         focus-visible:outline-indigo-600
         dark:focus-visible:outline-indigo-400"
>
  Save
</button>
```

---

## 7. Performance UX

### 7.1 Perceived Performance Techniques

**Perceived performance > actual performance.** Users judge speed by how it *feels*.

1. **Instant feedback:** Always acknowledge user actions immediately (button state change, optimistic update)
2. **Progressive rendering:** Show the page skeleton immediately, fill in data as it arrives
3. **Predictive prefetching:** Start loading what the user is likely to click next
4. **Meaningful transitions:** Short, purposeful transitions (150-300ms) make state changes feel smooth rather than jarring. Don't exceed 300ms — it starts to feel sluggish.

```ts
// Prefetch on hover (for likely navigation targets)
function usePrefetchOnHover(fetchFn: () => Promise<void>) {
  let prefetched = false
  function onMouseEnter() {
    if (!prefetched) {
      prefetched = true
      fetchFn() // Fire and forget — cache the result
    }
  }
  return { onMouseEnter }
}
```

### 7.2 Lazy Loading and Code Splitting

**Route-level code splitting** — Each route is a separate chunk:

```ts
// router.ts
const routes = [
  { path: '/', component: () => import('./pages/TodayPage.vue') },
  { path: '/practices', component: () => import('./pages/PracticesPage.vue') },
  { path: '/reader/:sourceId', component: () => import('./pages/ReaderPage.vue') },
  { path: '/flashcards', component: () => import('./pages/FlashcardsPage.vue') },
]
```

**Component-level lazy loading** — Heavy components loaded on demand:

```vue
<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

// Heavy chart library — only loaded when the tab is visible
const StreakChart = defineAsyncComponent(() => import('./StreakChart.vue'))
</script>

<template>
  <Suspense>
    <StreakChart v-if="activeTab === 'stats'" />
    <template #fallback>
      <SkeletonCard :lines="6" />
    </template>
  </Suspense>
</template>
```

### 7.3 Debouncing User Input

**Debounce search input** to avoid firing a request per keystroke:

```ts
// useDebounce.ts
export function useDebounce<T>(value: Ref<T>, delayMs = 300): Ref<T> {
  const debounced = ref(value.value) as Ref<T>
  let timeout: ReturnType<typeof setTimeout>

  watch(value, (newVal) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => {
      debounced.value = newVal
    }, delayMs)
  })

  return debounced
}
```

```vue
<!-- SearchInput.vue -->
<script setup lang="ts">
const query = ref('')
const debouncedQuery = useDebounce(query, 300)

const results = computedAsync(async () => {
  if (!debouncedQuery.value) return []
  return await api.search(debouncedQuery.value)
})
</script>

<template>
  <input v-model="query" placeholder="Search..." />
</template>
```

**Debounce guidelines:**
- **Search input:** 200-300ms
- **Auto-save:** 500-1000ms
- **Resize handlers:** 100-150ms
- **Scroll handlers:** 50-100ms (or use `IntersectionObserver`)

### 7.4 Virtual Scrolling for Long Lists

**When the list exceeds ~100 items**, render only visible items:

```vue
<!-- VirtualList.vue (simplified concept) -->
<script setup lang="ts">
import { useVirtualList } from '@vueuse/core'

const props = defineProps<{
  items: any[]
  itemHeight: number
}>()

const { list, containerProps, wrapperProps } = useVirtualList(
  () => props.items,
  { itemHeight: props.itemHeight }
)
</script>

<template>
  <div v-bind="containerProps" class="h-96 overflow-auto">
    <div v-bind="wrapperProps">
      <div
        v-for="{ data, index } in list"
        :key="data.id"
        :style="{ height: `${itemHeight}px` }"
      >
        <slot :item="data" :index="index" />
      </div>
    </div>
  </div>
</template>
```

**When to virtualize:**
- Scripture chapter lists (3,000+ chapters across standard works)
- Conference talk lists (1,000+ talks)
- Flashcard review queues
- Activity/log history

### 7.5 Image Optimization and Responsive Images

```vue
<!-- ResponsiveImage.vue -->
<template>
  <img
    :src="src"
    :srcset="srcset"
    :sizes="sizes"
    :alt="alt"
    :width="width"
    :height="height"
    loading="lazy"
    decoding="async"
    class="rounded-lg object-cover"
  />
</template>
```

**Rules:**
- Always set `width` and `height` attributes to prevent layout shift (CLS)
- Use `loading="lazy"` for below-fold images
- Use `decoding="async"` to not block rendering
- Serve WebP/AVIF with fallbacks
- Size images to their display size, not larger

---

## 8. Anti-Patterns to Avoid

### 8.1 Browser Dialogs in Production

```js
// ❌ NEVER in a real app
window.alert('Saved!')
window.confirm('Are you sure?')
window.prompt('Enter name:')

// ✅ INSTEAD
toast.show('Saved!')
dialog.open(ConfirmDialog, { onConfirm: handleDelete })
dialog.open(InputDialog, { label: 'Enter name', onSubmit: handleName })
```

**Why:** Blocks main thread, unstyled, inaccessible, untestable, breaks user trust, cannot contain rich content.

### 8.2 Layout Shifts (CLS Issues)

**Common causes and fixes:**

| Cause | Fix |
|---|---|
| Images without dimensions | Always set `width` and `height` |
| Web fonts causing FOUT/FOIT | Use `font-display: swap` + `size-adjust` |
| Dynamic content inserted above viewport | Insert below or reserve space |
| Skeleton → content size mismatch | Make skeletons match actual content dimensions |
| Late-loading banners/ads | Reserve static space with `min-height` |

### 8.3 Disabled Buttons Without Explanation

```vue
<!-- ❌ BAD: User has no idea why it's disabled -->
<button disabled class="opacity-50 cursor-not-allowed">Submit</button>

<!-- ✅ GOOD: Tooltip or visible hint explains the disabled state -->
<div class="relative group">
  <button
    :disabled="!isValid"
    class="... disabled:opacity-50 disabled:cursor-not-allowed"
  >
    Submit
  </button>
  <p
    v-if="!isValid"
    class="mt-1 text-xs text-gray-500"
  >
    Fill in all required fields to continue
  </p>
</div>

<!-- ✅ ALSO GOOD: Keep button enabled, validate on click and show errors -->
```

### 8.4 "Are You Sure?" for Reversible Actions

```vue
<!-- ❌ BAD: Unnecessary friction for a reversible action -->
<button @click="showConfirmDialog">Archive practice</button>
<!-- "Are you sure you want to archive this practice?" [Yes] [No] -->

<!-- ✅ GOOD: Just do it, offer undo -->
<button @click="archivePractice">Archive</button>
<!-- Toast: "Practice archived" [Undo] -->
```

**"Are you sure?" is a speedbump users learn to ignore.** Reserve it for truly destructive, irreversible actions like permanent deletion or account removal.

### 8.5 Infinite Scroll Without Position Memory

```ts
// ❌ BAD: User scrolls 200 items, clicks one, hits back → starts at top
// ✅ GOOD: Remember scroll position

// useScrollMemory.ts
export function useScrollMemory(key: string) {
  const route = useRoute()
  const scrollContainer = ref<HTMLElement>()

  // Save position before navigating away
  onBeforeRouteLeave(() => {
    if (scrollContainer.value) {
      sessionStorage.setItem(
        `scroll:${key}`,
        String(scrollContainer.value.scrollTop)
      )
    }
  })

  // Restore position on mount
  onMounted(() => {
    const saved = sessionStorage.getItem(`scroll:${key}`)
    if (saved && scrollContainer.value) {
      scrollContainer.value.scrollTop = parseInt(saved)
    }
  })

  return { scrollContainer }
}
```

### 8.6 Toast Spam

**Rules:**
- Maximum **3 toasts** visible at once
- Auto-dismiss after **3-5 seconds** (except errors)
- **Deduplicate** identical messages
- **Never toast errors** that require user action — show inline
- Don't toast for every single micro-action (e.g., don't toast "filter applied")

### 8.7 Z-Index Wars

**Establish a z-index scale and never deviate:**

```css
/* z-index scale — define once, use everywhere */
@theme {
  --z-dropdown: 10;
  --z-sticky: 20;
  --z-overlay: 30;
  --z-modal: 40;
  --z-toast: 50;
  --z-tooltip: 60;
}
```

```vue
<!-- Use the scale consistently -->
<div class="z-[var(--z-modal)]">Modal backdrop</div>
<div class="z-[var(--z-toast)]">Toast container</div>
```

**Or use Tailwind's built-in z-index utilities consistently:**

| Layer | Z-index | Tailwind |
|---|---|---|
| Base content | 0 | `z-0` |
| Dropdowns | 10 | `z-10` |
| Sticky headers | 20 | `z-20` |
| Sidebar overlays | 30 | `z-30` |
| Modal backdrop + modal | 40 | `z-40` |
| Toasts / notifications | 50 | `z-50` |

### 8.8 Overusing Modals

**Don't use a modal when:**
- A simple inline edit would suffice (editing a practice name)
- The content is read-only (viewing details — use a slide-over panel)
- The action is simple and low-risk (toggling a setting)
- You're stacking modals on modals (redesign the flow)

**Use a modal when:**
- You need focused attention for a multi-field form
- The action is destructive and irreversible
- The content is contextually separate from the page (compose message, create new item)

---

## 9. Vue 3 Specific Patterns

### 9.1 Composables for Shared UI Logic

**Composables encapsulate reusable stateful logic.** The UI primitives your app needs:

```ts
// useDialog.ts — Centralized dialog management
import { ref, markRaw, type Component } from 'vue'

interface DialogOptions {
  component: Component
  props?: Record<string, any>
  onClose?: (result?: any) => void
}

const dialogs = ref<DialogOptions[]>([])

export function useDialog() {
  function open(options: DialogOptions): Promise<any> {
    return new Promise((resolve) => {
      dialogs.value.push({
        ...options,
        component: markRaw(options.component),
        onClose: (result) => {
          close(options)
          options.onClose?.(result)
          resolve(result)
        }
      })
    })
  }

  function close(dialog: DialogOptions) {
    const index = dialogs.value.indexOf(dialog)
    if (index > -1) dialogs.value.splice(index, 1)
  }

  function closeAll() {
    dialogs.value = []
  }

  return { dialogs, open, close, closeAll }
}
```

```ts
// useToast.ts — Toast notifications
interface Toast {
  id: string
  message: string
  type: 'info' | 'success' | 'error' | 'warning'
  duration: number
  action?: { label: string; onClick: () => void }
}

const toasts = ref<Toast[]>([])
const MAX_VISIBLE = 3

export function useToast() {
  function show(message: string, options?: Partial<Omit<Toast, 'id' | 'message'>>) {
    // Deduplicate
    if (toasts.value.some(t => t.message === message)) return

    const id = crypto.randomUUID()
    const toast: Toast = {
      id,
      message,
      type: options?.type ?? 'info',
      duration: options?.duration ?? 4000,
      action: options?.action,
    }

    toasts.value.push(toast)

    // Trim to max visible
    if (toasts.value.length > MAX_VISIBLE) {
      toasts.value = toasts.value.slice(-MAX_VISIBLE)
    }

    // Auto dismiss (unless error)
    if (toast.type !== 'error') {
      setTimeout(() => dismiss(id), toast.duration)
    }

    return { dismiss: () => dismiss(id) }
  }

  function dismiss(id: string) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return { toasts, show, dismiss }
}
```

### 9.2 Teleport for Modals and Overlays

**Always teleport overlays to `<body>`** to escape:
- Parent `overflow: hidden`
- Parent `z-index` stacking contexts
- CSS `contain` or `transform` creating new stacking contexts

```vue
<!-- DialogProvider.vue — Add at app root -->
<template>
  <slot />

  <Teleport to="body">
    <!-- Modal container -->
    <template v-for="dialog in dialogs" :key="dialog.id">
      <component
        :is="dialog.component"
        v-bind="dialog.props"
        @close="dialog.onClose"
      />
    </template>

    <!-- Toast container -->
    <div
      class="pointer-events-none fixed inset-x-0 bottom-0 z-50 flex flex-col items-center gap-2 p-4"
      aria-live="polite"
    >
      <TransitionGroup
        enter-from-class="translate-y-4 opacity-0"
        enter-active-class="transition duration-300"
        leave-to-class="translate-y-4 opacity-0"
        leave-active-class="transition duration-200"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto"
        >
          <ToastItem :toast="toast" @dismiss="dismissToast(toast.id)" />
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
```

### 9.3 Transition and TransitionGroup

**Meaningful transitions improve UX. Gratuitous transitions hurt it.**

**Rules:**
- **Duration:** 150-300ms for UI transitions. Never exceed 500ms.
- **Easing:** Use `ease-out` for entrances, `ease-in` for exits, `ease-in-out` for movement
- **Purpose:** Transitions should communicate *what changed* and *where things went*
- **Respect reduced motion:** Always check `prefers-reduced-motion`

```vue
<!-- Fade transition for content changes -->
<Transition
  enter-from-class="opacity-0"
  enter-active-class="transition-opacity duration-200"
  leave-to-class="opacity-0"
  leave-active-class="transition-opacity duration-150"
  mode="out-in"
>
  <component :is="currentView" :key="currentViewKey" />
</Transition>
```

```vue
<!-- List transitions for adding/removing items -->
<TransitionGroup
  tag="ul"
  enter-from-class="opacity-0 -translate-x-4"
  enter-active-class="transition duration-300"
  leave-to-class="opacity-0 translate-x-4"
  leave-active-class="transition duration-200 absolute"
  move-class="transition-transform duration-300"
>
  <li v-for="item in list" :key="item.id">
    {{ item.name }}
  </li>
</TransitionGroup>
```

**Reduce motion:**

```ts
// useReducedMotion.ts
export function useReducedMotion() {
  const prefersReduced = ref(
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )

  const query = window.matchMedia('(prefers-reduced-motion: reduce)')
  query.addEventListener('change', (e) => {
    prefersReduced.value = e.matches
  })

  return prefersReduced
}
```

```css
/* Tailwind v4 — disable transitions when reduced motion is preferred */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

### 9.4 Provide/Inject for Theme and Notification Contexts

**Use provide/inject for app-wide services that many components need:**

```ts
// App.vue — Provide at root
import { provide } from 'vue'
import { useToast } from './composables/useToast'
import { useDarkMode } from './composables/useDarkMode'

const toast = useToast()
const theme = useDarkMode()

provide('toast', toast)
provide('theme', theme)
```

```ts
// Any child component — Inject where needed
import { inject } from 'vue'

const toast = inject('toast')!
toast.show('Practice saved!')
```

**Better: Use typed injection keys:**

```ts
// injection-keys.ts
import type { InjectionKey } from 'vue'
import type { UseToastReturn } from './composables/useToast'
import type { UseDarkModeReturn } from './composables/useDarkMode'

export const ToastKey: InjectionKey<UseToastReturn> = Symbol('toast')
export const ThemeKey: InjectionKey<UseDarkModeReturn> = Symbol('theme')
```

```ts
// Usage
provide(ToastKey, toast)
const toast = inject(ToastKey)! // Typed correctly
```

### 9.5 Presentational vs. Container Components

**Presentational components** (dumb/pure components):
- Receive data via props
- Emit events for actions
- No business logic, no API calls
- Highly reusable
- Easy to test
- Examples: `BaseButton`, `FormField`, `PracticeCard`, `EmptyState`

**Container components** (smart components):
- Fetch/manage data (call APIs, read stores)
- Pass data to presentational components
- Handle business logic and side effects
- Page-level or feature-level
- Examples: `PracticesPage`, `TodayDashboard`, `ReaderSidebar`

```vue
<!-- ✅ Presentational: PracticeCard.vue -->
<template>
  <div class="rounded-xl border p-4">
    <h3>{{ practice.name }}</h3>
    <p>{{ practice.streak }} day streak</p>
    <button @click="emit('log')">Log today</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{ practice: Practice }>()
const emit = defineEmits<{ log: [] }>()
</script>
```

```vue
<!-- ✅ Container: PracticesPage.vue -->
<template>
  <div>
    <PracticeCard
      v-for="practice in practices"
      :key="practice.id"
      :practice="practice"
      @log="handleLog(practice.id)"
    />
  </div>
</template>

<script setup lang="ts">
import { usePractices } from '@/composables/usePractices'
const { practices, logPractice } = usePractices()

async function handleLog(id: string) {
  await logPractice(id)
  toast.show('Practice logged!')
}
</script>
```

---

## 10. Accessibility (a11y)

### 10.1 ARIA Roles and Landmarks

**Use semantic HTML first.** ARIA is a supplement, not a replacement.

```html
<!-- ✅ Semantic HTML provides implicit roles -->
<header>     <!-- role="banner" -->
<nav>        <!-- role="navigation" -->
<main>       <!-- role="main" -->
<aside>      <!-- role="complementary" -->
<footer>     <!-- role="contentinfo" -->
<form>       <!-- role="form" (when named) -->
<section>    <!-- role="region" (when named) -->
<article>    <!-- role="article" -->

<!-- ❌ DON'T add redundant ARIA -->
<nav role="navigation"> <!-- Redundant! -->
<button role="button">  <!-- Redundant! -->
```

**When ARIA is needed:**

```html
<!-- Custom widget needing ARIA -->
<div role="tablist" aria-label="Practice categories">
  <button role="tab" aria-selected="true" aria-controls="panel-1">Active</button>
  <button role="tab" aria-selected="false" aria-controls="panel-2">Archived</button>
</div>
<div role="tabpanel" id="panel-1" aria-labelledby="tab-1">...</div>

<!-- Live regions for dynamic content -->
<div aria-live="polite">            <!-- Screen reader announces changes -->
  {{ statusMessage }}
</div>
<div aria-live="assertive">         <!-- Interrupts current announcement -->
  {{ errorMessage }}
</div>

<!-- Custom toggle -->
<button
  role="switch"
  :aria-checked="isDarkMode"
  @click="toggleDarkMode"
>
  Dark mode
</button>
```

### 10.2 Screen Reader Considerations

**Visually hidden text** — Content only for screen readers:

```css
/* sr-only in Tailwind (built-in) */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
```

```vue
<!-- Button with icon-only visual, but accessible name -->
<button @click="deletePractice" aria-label="Delete practice: Morning Prayer">
  <TrashIcon class="h-5 w-5" />
</button>

<!-- State communicated visually AND to screen readers -->
<span class="text-green-600">
  <CheckIcon class="h-4 w-4" aria-hidden="true" />
  <span>Completed</span>
  <!-- OR if the text is visual-only: -->
  <span class="sr-only">Status: Completed</span>
</span>

<!-- Dynamic count announcements -->
<div aria-live="polite" class="sr-only">
  {{ filteredPractices.length }} practices shown
</div>
```

**Icon best practices:**
- Decorative icons: `aria-hidden="true"` (icon next to text label)
- Meaningful icons: `aria-label` on parent or `<title>` inside SVG
- Never leave interactive icons without an accessible name

### 10.3 Keyboard-Only Navigation

**Every interactive element must be:**
1. **Focusable** (in the tab order or programmatically focusable)
2. **Operable** (activatable via Enter/Space)
3. **Visually indicated** when focused (focus ring)

**Tab order rules:**
- Follow visual/reading order (don't set `tabindex` > 0)
- Use `tabindex="0"` only to make non-interactive elements focusable
- Use `tabindex="-1"` for programmatic focus (e.g., after route changes)
- Group related controls — use arrow keys within groups (tabs, menus, radio groups)

**Skip navigation:**

```vue
<!-- First element in <body> -->
<a
  href="#main-content"
  class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50
         focus:rounded-lg focus:bg-indigo-600 focus:px-4 focus:py-2 focus:text-white"
>
  Skip to main content
</a>

<!-- ... navigation ... -->

<main id="main-content" tabindex="-1">
  <!-- Page content -->
</main>
```

### 10.4 Focus Management During Route Changes

**Problem:** When a SPA navigates, focus stays on the clicked link. Screen reader users don't know the page changed.

**Solution:** Move focus to the new page's heading or main content area.

```ts
// router.ts
router.afterEach((to, from) => {
  // Wait for DOM update
  nextTick(() => {
    // Focus the main content area
    const main = document.getElementById('main-content')
    if (main) {
      main.focus()
    }

    // Announce route change to screen readers
    const announcer = document.getElementById('route-announcer')
    if (announcer) {
      announcer.textContent = `Navigated to ${to.meta.title || to.name}`
    }
  })
})
```

```vue
<!-- RouteAnnouncer.vue — Add at app root -->
<template>
  <div
    id="route-announcer"
    role="status"
    aria-live="assertive"
    aria-atomic="true"
    class="sr-only"
  />
</template>
```

### 10.5 Reduced Motion Preferences

**Respect `prefers-reduced-motion`.** Some users experience motion sickness, seizures, or distraction from animations.

```css
/* Global reduced motion support */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

**In Vue transitions:**

```vue
<script setup lang="ts">
const prefersReduced = useReducedMotion()
</script>

<template>
  <Transition
    :enter-active-class="prefersReduced ? '' : 'transition duration-200'"
    :leave-active-class="prefersReduced ? '' : 'transition duration-150'"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
  >
    <div v-if="show">Content</div>
  </Transition>
</template>
```

**What to reduce (not eliminate):**
- **Remove:** Parallax, auto-playing animations, decorative motion
- **Simplify:** Slide → fade, bounce → instant, zoom → crossfade
- **Keep:** Loading spinners (functional), progress bars, focus indicators (essential)

---

## Quick Reference: Decision Trees

### Should I Use a Modal?

```
Is user input required before proceeding?
├── Yes → Is it a simple choice (2-3 options)?
│   ├── Yes → Inline confirmation or action menu
│   └── No → Modal dialog
└── No → Is it a status update?
    ├── Yes → Toast notification
    └── No → Is it an error?
        ├── Yes → Inline error (near the cause)
        └── No → Probably don't need anything
```

### Should I Show "Are You Sure?"

```
Is the action destructive?
├── No → Just do it (maybe toast "Done")
└── Yes → Is it reversible?
    ├── Yes → Just do it + offer Undo
    └── No → Is it significant?
        ├── No → Just do it + offer Undo (e.g., soft-delete)
        └── Yes → Confirmation dialog (e.g., delete account)
```

### What Loading Pattern Should I Use?

```
How long will it take?
├── <200ms → Nothing (no indicator)
├── 200ms-2s → Is the layout predictable?
│   ├── Yes → Skeleton screen
│   └── No → Spinner
├── 2s-10s → Is progress measurable?
│   ├── Yes → Progress bar
│   └── No → Skeleton + "Still loading…" message after 3s
└── >10s → Progress bar + estimated time + option to continue in background
```

---

## Appendix: Component Inventory Checklist

When building any new UI component, verify:

- [ ] **Keyboard accessible** — Can operate with Tab, Enter, Space, Escape, Arrow keys as appropriate
- [ ] **Focus visible** — Clear focus indicator in both light and dark mode
- [ ] **Screen reader accessible** — Meaningful ARIA attributes, labels, and live regions
- [ ] **Dark mode** — Tested and styled in both themes
- [ ] **Responsive** — Works on mobile (320px) through desktop (1920px+)
- [ ] **Touch targets** — Interactive elements are at least 44×44px on mobile
- [ ] **Loading state** — Shows appropriate feedback during async operations
- [ ] **Empty state** — Handles zero-data gracefully with guidance
- [ ] **Error state** — Shows recovery-focused errors, preserves user input
- [ ] **Reduced motion** — Respects `prefers-reduced-motion`
- [ ] **No layout shifts** — Dimensions are stable as content loads
- [ ] **Color contrast** — Text meets WCAG AA (4.5:1 for normal, 3:1 for large)

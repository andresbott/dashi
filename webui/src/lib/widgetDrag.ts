import { reactive, readonly } from 'vue'
import type { Widget } from '@/types/dashboard'
import { beginDragGuard, endDragGuard } from '@/lib/dragGuard'
import {
    clipRect,
    resolveDropTarget,
    type DragSnapshot,
    type Rect,
    type RowSnapshot,
    type StackZoneSnapshot,
} from '@/lib/widgetDragHit'
import type { DragSource, DropTarget, WidgetMove } from '@/lib/widgetMove'

// ---------------------------------------------------------------------------
// Widget drag session for the dashboard editor, shared by every row and stack
// (module scope: only one widget is dragged at a time).
//
// Rows and stacks register their elements; on every pointer move (and scroll)
// the session measures them, resolves one drop target (a stack zone while the
// pointer is inside one, else the nearest row) and publishes it in dragState
// for the guides. The component that started the drag calls endWidgetDrag on
// release and hands the returned move to applyWidgetMove. Components never
// measure inside a computed: geometry is read here, lazily, on each event.
// SortableJS only detects the drag and paints the floating clone.
// ---------------------------------------------------------------------------

export interface RowRegistration {
    // The whole row card, toolbar included, so there is no dead strip above
    // the grid where a drop silently fails.
    root: HTMLElement
    grid: () => HTMLElement | null
    widgets: () => Widget[]
}

export interface StackZoneRegistration {
    // The drop zone (the stack card's body); null while not mounted.
    zone: () => HTMLElement | null
    // The child cards, in order, for the insert index.
    children: () => HTMLElement[]
}

export interface DragStartOptions {
    // Start column the drag measures from (row sources; ignored otherwise).
    originColumn: number
    // Viewport x of the dragged widget's left edge at the grab.
    originLeftX: number
    // True grab point, captured on pointer-down.
    grabX: number
    grabY: number
    // The dragged widget's own span (row sources).
    span: number
    draggedType: string
    // Index of the dragged child in its stack (stack sources).
    sourceIndex?: number
}

const rowRegistry = new Map<string, RowRegistration>()
const zoneRegistry = new Map<string, StackZoneRegistration>()

function register<T>(map: Map<string, T>, id: string, entry: T): () => void {
    map.set(id, entry)
    // Only remove this entry: a re-mount may have registered a newer one.
    return () => {
        if (map.get(id) === entry) map.delete(id)
    }
}

// registerRow adds a row card to the session; call the returned function on
// unmount.
export function registerRow(id: string, entry: RowRegistration): () => void {
    return register(rowRegistry, id, entry)
}

// registerStackZone adds a stack drop zone to the session; call the returned
// function on unmount.
export function registerStackZone(id: string, entry: StackZoneRegistration): () => void {
    return register(zoneRegistry, id, entry)
}

const state = reactive({
    active: false,
    sourceKind: null as DragSource['kind'] | null,
    sourceRowId: null as string | null,
    sourceStackId: null as string | null,
    sourceWidgetId: null as string | null,
    draggedType: null as string | null,
    targetKind: null as DropTarget['kind'] | null,
    targetRowId: null as string | null,
    targetStackId: null as string | null,
    targetColumn: null as number | null,
    targetIndex: null as number | null,
    // Columns the guide highlights on the target row.
    targetSpan: 1,
    // False when the hovered row has no room: the guide renders blocked and
    // releasing there does nothing.
    targetFits: true,
    // Set by Escape: the release is ignored and the widget stays put.
    cancelled: false,
})

// The published drag state, read-only for components (guides, markers).
export const dragState = readonly(state)

// Pointer bookkeeping for the active drag. Deliberately not reactive: it
// changes on every pointer move, and only the resolved target needs to render.
let session: {
    source: DragSource
    opts: DragStartOptions
    pointerX: number
    pointerY: number
} | null = null

function clearTarget() {
    state.targetKind = null
    state.targetRowId = null
    state.targetStackId = null
    state.targetColumn = null
    state.targetIndex = null
    state.targetSpan = 1
    state.targetFits = true
}

function measureRows(): RowSnapshot[] {
    const out: RowSnapshot[] = []
    for (const [id, entry] of rowRegistry) {
        const r = entry.root.getBoundingClientRect()
        const g = entry.grid()?.getBoundingClientRect()
        out.push({
            id,
            top: r.top,
            bottom: r.bottom,
            grid: g ? { left: g.left, width: g.width } : null,
            widgets: entry.widgets(),
        })
    }
    return out
}

// The row's sideways scroll container (DashboardRow.vue, below md). A stack
// scrolled out of it must not be hit, so its zone is clipped to this box.
const ZONE_CLIP_SELECTOR = '.grid-scroll'

function measureRect(el: Element): Rect {
    const { left, top, right, bottom } = el.getBoundingClientRect()
    return { left, top, right, bottom }
}

function measureZones(): StackZoneSnapshot[] {
    const out: StackZoneSnapshot[] = []
    for (const [id, entry] of zoneRegistry) {
        const el = entry.zone()
        if (!el) continue
        const clip = el.closest(ZONE_CLIP_SELECTOR)
        const rect = clipRect(measureRect(el), clip ? measureRect(clip) : null)
        if (!rect) continue // scrolled fully out of view
        const childMidYs = entry.children().map((c) => {
            const r = c.getBoundingClientRect()
            return (r.top + r.bottom) / 2
        })
        out.push({ id, rect, childMidYs })
    }
    return out
}

// Resolves the target under the pointer and publishes it. Runs on every
// pointer move and on scroll (the pointer is still but the page moved).
function updateTarget() {
    if (!session || !state.active || state.cancelled) return
    const { opts } = session
    const snapshot: DragSnapshot = {
        source: session.source,
        draggedType: opts.draggedType,
        span: opts.span,
        sourceIndex: opts.sourceIndex,
        originColumn: opts.originColumn,
        grabStartX: opts.grabX,
        originLeftX: opts.originLeftX,
        pointerX: session.pointerX,
        pointerY: session.pointerY,
        rows: measureRows(),
        zones: measureZones(),
    }
    const hit = resolveDropTarget(snapshot)
    clearTarget()
    if (!hit) return
    state.targetKind = hit.target.kind
    if (hit.target.kind === 'row') {
        state.targetRowId = hit.target.rowId
        state.targetColumn = hit.target.column
    } else {
        state.targetStackId = hit.target.stackId
        state.targetIndex = hit.target.index
    }
    state.targetSpan = hit.span
    state.targetFits = hit.fits
}

const onPointerMove = (e: PointerEvent) => {
    if (!session) return
    session.pointerX = e.clientX
    session.pointerY = e.clientY
    updateTarget()
}

// Escape abandons the drag: hide the guides, ignore the release, and ask
// SortableJS to finish (it only listens for the pointer going up).
const onKeyDown = (e: KeyboardEvent) => {
    if (e.key !== 'Escape' || !state.active) return
    state.cancelled = true
    clearTarget()
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }))
}

function teardown() {
    document.removeEventListener('pointermove', onPointerMove)
    document.removeEventListener('keydown', onKeyDown)
    window.removeEventListener('scroll', updateTarget, { capture: true })
    endDragGuard()
    session = null
    state.active = false
    state.sourceKind = null
    state.sourceRowId = null
    state.sourceStackId = null
    state.sourceWidgetId = null
    state.draggedType = null
    state.cancelled = false
    clearTarget()
}

// startWidgetDrag begins a drag of the widget at source and resolves the
// first target at the grab point. A drag still in progress is dropped first.
export function startWidgetDrag(source: DragSource, opts: DragStartOptions): void {
    if (session) teardown()
    session = { source, opts, pointerX: opts.grabX, pointerY: opts.grabY }
    state.active = true
    state.sourceKind = source.kind
    state.sourceRowId = source.kind === 'row' ? source.rowId : null
    state.sourceStackId = source.kind === 'stack' ? source.stackId : null
    state.sourceWidgetId = source.widgetId
    state.draggedType = opts.draggedType
    state.cancelled = false
    beginDragGuard('grabbing')
    // The draggables run in force-fallback (pointer) mode, so the pointer is
    // tracked here; a scroll (autoscroll or wheel) re-resolves the target.
    document.addEventListener('pointermove', onPointerMove)
    document.addEventListener('keydown', onKeyDown)
    window.addEventListener('scroll', updateTarget, { capture: true, passive: true })
    updateTarget()
}

// endWidgetDrag tears the session down and returns the move to apply, or
// null when the drag was cancelled, ended over nothing, or over a row without
// room.
export function endWidgetDrag(): WidgetMove | null {
    const source = session?.source ?? null
    const { cancelled, targetFits, targetKind, targetRowId, targetColumn, targetStackId, targetIndex } = state
    teardown()
    if (!source || cancelled || !targetFits) return null
    if (targetKind === 'row' && targetRowId !== null && targetColumn !== null) {
        return { source, target: { kind: 'row', rowId: targetRowId, column: targetColumn } }
    }
    if (targetKind === 'stack' && targetStackId !== null && targetIndex !== null) {
        return { source, target: { kind: 'stack', stackId: targetStackId, index: targetIndex } }
    }
    return null
}

// abortIfSource tears the drag down when the row or stack id is its source:
// a source removed mid-gesture (page switch, row delete) must not leave
// document listeners or a stuck global drag behind.
export function abortIfSource(id: string): void {
    if (state.active && (state.sourceRowId === id || state.sourceStackId === id)) teardown()
}

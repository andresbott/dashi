// Guards the page while a pointer gesture (widget drag, resize) is in
// progress: text must not get selected as the pointer sweeps across widget
// content, and the cursor should read as the gesture everywhere, not only over
// the handle. SortableJS only disables selection on Safari, so the editor does
// it itself. Paired calls: begin at gesture start, end on release or cancel.

export const DRAG_GUARD_CLASS = 'dashi-dragging'

export type DragCursor = 'grabbing' | 'col-resize' | 'row-resize'

export function beginDragGuard(cursor: DragCursor = 'grabbing'): void {
    document.body.classList.add(DRAG_GUARD_CLASS)
    document.body.style.setProperty('--dashi-drag-cursor', cursor)
    try {
        window.getSelection()?.removeAllRanges()
    } catch {
        // Selection API unavailable (very old or non-browser environments).
    }
}

export function endDragGuard(): void {
    document.body.classList.remove(DRAG_GUARD_CLASS)
    document.body.style.removeProperty('--dashi-drag-cursor')
}

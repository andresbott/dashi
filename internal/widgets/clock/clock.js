import { every } from '/_dashi/assets/dashi.js'

// Go prefilled the current time; this only keeps it ticking. Locale is
// pinned to en-US so the browser's output matches the Go prefill exactly
// (see formatTime in image.go).
export function formatTime(date, hour12, showSeconds) {
    return date.toLocaleTimeString('en-US', {
        hour: hour12 ? 'numeric' : '2-digit',
        minute: '2-digit',
        ...(showSeconds ? { second: '2-digit' } : {}),
        // 24-hour output asks for h23 explicitly rather than hour12:false.
        // hour12:false has an h23/h24 history across engines: an h24 engine
        // renders midnight as "24:32", while Go's "15:04" always gives
        // "00:32". Go is the contract (see the parity tables in
        // clock.test.js and image_test.go).
        ...(hour12 ? { hour12: true } : { hourCycle: 'h23' }),
    })
}

export function formatDate(date) {
    return date.toLocaleDateString('en-US', {
        weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
    })
}

every('.dashi-clock', 1000, el => {
    const now = new Date()
    const time = el.querySelector('.dashi-clock__time')
    if (time) {
        time.textContent = formatTime(now, el.dataset.hour12 === '1', el.dataset.seconds === '1')
    }
    const date = el.querySelector('.dashi-clock__date')
    if (date) {
        date.textContent = formatDate(now)
    }
})

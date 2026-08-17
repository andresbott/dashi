// Shared viewer runtime. Go renders every widget's markup with values
// already filled in; this module only keeps those values fresh.
//
// Hard rule (see docs/superpowers/specs/2026-08-17-vanilla-viewer-architecture.md):
// JavaScript never constructs markup. It sets text, sets attributes,
// merges style and toggles hidden — that is the whole contract, which is
// why the setter below has no `html` option.

/**
 * Refresh a widget's values from a JSON endpoint on an interval.
 *
 * @param {string} selector   widget roots, e.g. '.dashi-sysinfo'
 * @param {object} opts
 * @param {string|function} opts.url    endpoint, or (el) => endpoint
 * @param {number} opts.every           seconds between refreshes; 0 disables
 * @param {function} opts.apply         (el, data, set) => void
 */
export function poll(selector, { url, every: seconds, apply }) {
  for (const el of document.querySelectorAll(selector)) {
    if (!seconds || seconds <= 0) continue
    const run = () => refresh(el, url, apply)

    // Track when this widget last refreshed (from any source: interval or visibility)
    let lastRefresh = Date.now()
    const wrappedRun = () => {
      lastRefresh = Date.now()
      run()
    }

    setInterval(wrappedRun, seconds * 1000)

    // Refresh once when a hidden tab becomes visible again, if the interval
    // elapsed while it was asleep. Without this a laptop lid closed overnight
    // shows yesterday's numbers.
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState !== 'visible') return
      if (Date.now() - lastRefresh < seconds * 1000) return
      wrappedRun()
    })
  }
}

/**
 * Run fn against every matching widget immediately and then on an
 * interval. For widgets that update from local state, such as the clock.
 *
 * @param {string} selector
 * @param {number} ms
 * @param {function} fn  (el) => void
 */
export function every(selector, ms, fn) {
  for (const el of document.querySelectorAll(selector)) {
    fn(el)
    setInterval(() => fn(el), ms)
  }
}

async function refresh(el, url, apply) {
  try {
    const target = typeof url === 'function' ? url(el) : url
    const res = await fetch(target, { headers: { Accept: 'application/json' } })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    apply(el, await res.json(), setter(el))
    delete el.dataset.dashiError
  } catch (err) {
    // A failing widget keeps its last good values and records why, so a
    // dead endpoint degrades one widget instead of the page.
    el.dataset.dashiError = String((err && err.message) || err)
  }
}

function setter(root) {
  return (selector, value) => {
    const el = root.querySelector(selector)
    if (!el) return
    if (value === null || typeof value !== 'object') {
      el.textContent = value === null || value === undefined ? '' : String(value)
      return
    }
    for (const [key, val] of Object.entries(value)) {
      if (key === 'text') el.textContent = String(val)
      else if (key === 'hidden') el.hidden = Boolean(val)
      else if (key === 'style') Object.assign(el.style, val)
      else el.setAttribute(key, String(val))
    }
  }
}

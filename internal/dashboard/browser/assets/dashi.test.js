import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { poll, every } from './dashi.js'

describe('poll', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    document.body.innerHTML = `
      <div class="w" data-src="/api/v0/widgets/demo">
        <span class="value">old</span>
        <span class="opt">x</span>
      </div>`
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('applies fetched data on the interval and leaves prefilled values alone until then', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true,
      json: async () => ({ n: 42, visible: false }),
    })))

    poll('.w', {
      url: el => el.dataset.src,
      every: 30,
      apply: (el, data, set) => {
        set('.value', data.n)
        set('.opt', { hidden: !data.visible })
      },
    })

    // Go prefilled the markup, so nothing is fetched up front.
    expect(fetch).not.toHaveBeenCalled()
    expect(document.querySelector('.value').textContent).toBe('old')

    await vi.advanceTimersByTimeAsync(30_000)

    expect(fetch).toHaveBeenCalledWith('/api/v0/widgets/demo', {
      headers: { Accept: 'application/json' },
    })
    expect(document.querySelector('.value').textContent).toBe('42')
    expect(document.querySelector('.opt').hidden).toBe(true)
  })

  it('records failures on the widget root without throwing', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: false, status: 503 })))

    poll('.w', { url: '/x', every: 5, apply: () => { throw new Error('never') } })
    await vi.advanceTimersByTimeAsync(5_000)

    expect(document.querySelector('.w').dataset.dashiError).toContain('503')
    expect(document.querySelector('.value').textContent).toBe('old')
  })

  it('does nothing when every is 0', async () => {
    vi.stubGlobal('fetch', vi.fn())
    poll('.w', { url: '/x', every: 0, apply: () => {} })
    await vi.advanceTimersByTimeAsync(600_000)
    expect(fetch).not.toHaveBeenCalled()
  })
})

describe('every', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    document.body.innerHTML = `<div class="c"><span class="t">--</span></div>`
  })
  afterEach(() => vi.useRealTimers())

  it('runs immediately and then on the interval', () => {
    let n = 0
    every('.c', 1000, el => {
      n += 1
      el.querySelector('.t').textContent = String(n)
    })

    expect(document.querySelector('.t').textContent).toBe('1')
    vi.advanceTimersByTime(3000)
    expect(document.querySelector('.t').textContent).toBe('4')
  })
})

describe('set', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    document.body.innerHTML = `<div class="w"><i class="a"></i></div>`
  })
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('sets text, attributes and style but ignores missing elements', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => ({}) })))

    poll('.w', {
      url: '/x',
      every: 1,
      apply: (el, data, set) => {
        set('.a', { style: { width: '50%' }, 'aria-label': 'half' })
        set('.missing', 'ignored')
      },
    })
    await vi.advanceTimersByTimeAsync(1000)

    const a = document.querySelector('.a')
    expect(a.style.width).toBe('50%')
    expect(a.getAttribute('aria-label')).toBe('half')
  })
})

describe('poll visibility handling', () => {
  let intervals = []
  let originalVisibilityState

  beforeEach(() => {
    vi.useFakeTimers()
    document.body.innerHTML = `
      <div class="w">
        <span class="value">old</span>
      </div>`
    intervals = []
    originalVisibilityState = document.visibilityState

    // Capture interval IDs so we can cancel them
    const originalSetInterval = global.setInterval
    vi.stubGlobal('setInterval', (fn, ms) => {
      const id = originalSetInterval(fn, ms)
      intervals.push(id)
      return id
    })
  })

  afterEach(() => {
    intervals.forEach(clearInterval)
    // Restore original visibilityState
    Object.defineProperty(document, 'visibilityState', {
      value: originalVisibilityState,
      configurable: true,
    })
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('does not refresh on visibility when less than interval passed since last refresh', async () => {
    let callCount = 0
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true,
      json: async () => ({ n: ++callCount }),
    })))

    poll('.w', {
      url: '/x',
      every: 30,
      apply: (el, data, set) => set('.value', data.n),
    })

    // First interval fires at t=30
    await vi.advanceTimersByTimeAsync(30_000)
    expect(document.querySelector('.value').textContent).toBe('1')

    // Clear intervals to prevent them from firing again
    intervals.forEach(clearInterval)
    intervals = []

    // 2 seconds later at t=32 (only 2s since last refresh), trigger visibility
    await vi.advanceTimersByTimeAsync(2_000)
    Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true })
    document.dispatchEvent(new Event('visibilitychange'))

    // Should NOT have refreshed (only 2s < 30s since last refresh)
    // Value should still be '1' from the first interval
    expect(document.querySelector('.value').textContent).toBe('1')
  })

  it('refreshes on visibility when more than interval passed since last refresh', async () => {
    let callCount = 100
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true,
      json: async () => ({ n: ++callCount }),
    })))

    poll('.w', {
      url: '/x',
      every: 30,
      apply: (el, data, set) => set('.value', data.n),
    })

    // First interval fires at t=30
    await vi.advanceTimersByTimeAsync(30_000)
    expect(document.querySelector('.value').textContent).toBe('101')

    // Clear intervals to prevent them from firing again
    intervals.forEach(clearInterval)
    intervals = []

    // 35 seconds later at t=65 (35s since last refresh, > 30s interval)
    await vi.advanceTimersByTimeAsync(35_000)

    // Tab becomes visible (triggering visibility handler)
    Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true })
    document.dispatchEvent(new Event('visibilitychange'))

    // Give microtasks a chance to run
    await Promise.resolve()
    await Promise.resolve()

    // Should have refreshed (35s > 30s since last refresh)
    // Value should have changed from '101' (handlers from previous tests may also fire,
    // so we just verify THIS widget was updated)
    const value = document.querySelector('.value').textContent
    expect(parseInt(value)).toBeGreaterThan(101)
  })
})

import { describe, it, expect } from 'vitest'
import { formatTime, formatDate } from './clock.js'

// This table must stay identical to TestFormatTimeParity in image.go's test
// file — the server prefills the value and the browser overwrites it one
// second later, so a mismatch shows up as a visible flicker. Go is the
// contract.
describe('clock formatting parity with Go', () => {
  const afternoon = new Date(2026, 7, 17, 14, 32, 5)
  const midnight = new Date(2026, 7, 17, 0, 32, 5)

  it('formats 24-hour time', () => {
    expect(formatTime(afternoon, false, false)).toBe('14:32')
    expect(formatTime(afternoon, false, true)).toBe('14:32:05')
  })

  it('formats 12-hour time', () => {
    expect(formatTime(afternoon, true, false)).toBe('2:32 PM')
    expect(formatTime(afternoon, true, true)).toBe('2:32:05 PM')
  })

  // Guards the h23/h24 hazard: an engine treating hour12:false as h24
  // renders this as "24:32", while Go's "15:04" always renders "00:32".
  it('formats midnight the way Go does', () => {
    expect(formatTime(midnight, false, false)).toBe('00:32')
    expect(formatTime(midnight, false, true)).toBe('00:32:05')
    expect(formatTime(midnight, true, false)).toBe('12:32 AM')
    expect(formatTime(midnight, true, true)).toBe('12:32:05 AM')
  })

  it('formats the date the way Go does', () => {
    expect(formatDate(afternoon)).toBe('Monday, August 17, 2026')
  })
})

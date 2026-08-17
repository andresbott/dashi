import { describe, it, expect } from 'vitest'
import { humanBytes, humanUptime } from './sysinfo.js'

// This table is duplicated in image_test.go (TestHumanBytesParity /
// TestHumanUptimeParity). Both must produce identical strings: Go renders
// the value first, JS overwrites it on the next poll.
describe('humanBytes matches Go', () => {
  const cases = [
    [0, '0 B'],
    [512, '512 B'],
    [1024, '1.0 KB'],
    [1536, '1.5 KB'],
    [1048576, '1.0 MB'],
    [1073741824, '1.0 GB'],
    [1099511627776, '1.0 TB'],
  ]
  for (const [input, want] of cases) {
    it(`${input} -> ${want}`, () => expect(humanBytes(input)).toBe(want))
  }
})

describe('humanUptime matches Go', () => {
  const cases = [
    [45, '0m'],
    [90, '1m'],
    [3600, '1h 0m'],
    [3660, '1h 1m'],
    [90000, '1d 1h 0m'],
  ]
  for (const [input, want] of cases) {
    it(`${input} -> ${want}`, () => expect(humanUptime(input)).toBe(want))
  }
})

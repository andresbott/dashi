import type { DocQueryMock } from '@/lib/widgetDocs'

// Sample images for the widget docs, as inline SVG so nothing is fetched.
// Both are wider than tall, so the fit modes differ visibly in the preview.
const svg = (body: string) =>
    `data:image/svg+xml;charset=utf-8,${encodeURIComponent(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 200">${body}</svg>`)}`

export const sampleImages: Record<string, string> = {
    'mountains.svg': svg(
        '<rect width="400" height="200" fill="#bfe3f5"/><circle cx="320" cy="50" r="22" fill="#ffd166"/>'
        + '<path d="M0 200 120 60 200 150 270 90 400 200Z" fill="#5c7a8a"/>'
        + '<path d="M120 60 145 90 120 85 100 95Z" fill="#fff"/><rect y="180" width="400" height="20" fill="#7fb069"/>',
    ),
    'city.svg': svg(
        '<rect width="400" height="200" fill="#2b2d42"/><circle cx="70" cy="45" r="16" fill="#edf2f4"/>'
        + '<path d="M20 200V110h50v90M90 200V70h40v130M150 200V120h60v80M230 200V50h45v150M295 200V100h80v100" fill="#8d99ae"/>',
    ),
}

export const imageMocks: Record<string, DocQueryMock> = {
    'data-images': () =>
        Object.keys(sampleImages).map(name => ({ name, size: 2048, modTime: '2026-01-01T00:00:00Z' })),
}

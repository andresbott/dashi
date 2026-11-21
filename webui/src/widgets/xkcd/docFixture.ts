import type { DocQueryMock } from '@/lib/widgetDocs'
import type { XkcdComic } from './types'

// A hand-drawn-style sample panel (two stick figures), so the docs show the
// widget's layout without loading a real comic.
// width/height give the image an intrinsic size, like a real comic PNG.
const panel = `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="360" viewBox="0 0 300 180">
<rect x="2" y="2" width="296" height="176" fill="#fff" stroke="#222" stroke-width="2"/>
<g fill="none" stroke="#222" stroke-width="2.5" stroke-linecap="round">
<circle cx="95" cy="62" r="13"/><path d="M95 75v45M95 90l-18 14M95 90l18 10M95 120l-14 30M95 120l14 30"/>
<circle cx="205" cy="62" r="13"/><path d="M205 75v45M205 90l-18 10M205 90l18 14M205 120l-14 30M205 120l14 30"/>
</g>
<g font-family="Comic Sans MS, cursive" font-size="13" fill="#222" text-anchor="middle">
<text x="95" y="30">Is this the real comic?</text><text x="205" y="45">It's a sample.</text>
</g></svg>`

const img = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(panel)}`

const comics: Record<string, XkcdComic> = {
    latest: { num: 3000, title: 'Sample Comic', safe_title: 'Sample Comic', img, alt: 'A sample panel shown in the widget docs.', day: '1', month: '1', year: '2026' },
    random: { num: 1234, title: 'Daily Random Sample', safe_title: 'Daily Random Sample', img, alt: 'Same comic all day.', day: '1', month: '1', year: '2013' },
    'random-each': { num: 567, title: 'Random Sample', safe_title: 'Random Sample', img, alt: 'A new comic on every load.', day: '1', month: '1', year: '2009' },
}

export const xkcdMocks: Record<string, DocQueryMock> = {
    xkcd: (key) => comics[String(key[1])] ?? comics.latest,
}

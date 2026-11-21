import type { WidgetDocs } from '@/lib/widgetDocs'
import { DATA_IMAGE_URL } from '@/lib/injectionKeys'
import { imageMocks, sampleImages } from './docFixture'

const docs: WidgetDocs = {
    summary: 'Shows an uploaded image, filling the widget cell. Images are shared across all dashboards and managed under Content → Images.',
    params: [
        { name: 'image', type: 'string', description: 'Name of the uploaded image, e.g. `photo.jpg`.' },
        { name: 'fit', type: 'string', default: 'cover', description: 'How the image fills the cell: `cover` (fill, crop if needed), `contain` (fit entirely), or `fill` (stretch).' },
    ],
    preview: {
        title: 'Image',
        height: '9rem',
        config: { image: 'mountains.svg', fit: 'cover' },
        provide: [[DATA_IMAGE_URL, (name: string) => sampleImages[name] ?? '']],
        settingsNote: 'The preview offers two sample images instead of your uploads.',
    },
    mocks: imageMocks,
}

export default docs

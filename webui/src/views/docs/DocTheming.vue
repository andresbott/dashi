<template>
    <section id="theming">
        <h2>Theming</h2>
        <p>Dashi supports customizing the look and feel of dashboards through themes, custom assets, and CSS overrides.</p>

        <h3 id="theme-themes">Themes</h3>
        <p>A theme controls the colours and the page background of a dashboard. Each dashboard selects one through its <code>themeId</code> property. Themes are stored under <code>&lt;dataDir&gt;/data/themes/</code> and managed in the editor under <strong>Themes</strong>.</p>

        <h4>Theme contents</h4>
        <ul>
            <li><strong>Background</strong> &mdash; optional. A gradient or an image, per colour mode. Without one, the page uses the palette <code>bg</code> colour.</li>
            <li><strong>Palette</strong> &mdash; a set of colour roles for light mode and another for dark mode.</li>
        </ul>
        <p>Dashi ships two built-in themes: <strong>Default</strong>, and <strong>Ember</strong> (warm paper in light mode, warm charcoal with amber accents in dark mode). Built-in themes are read-only. To customise one, use <strong>Copy</strong> and edit the copy. A copy also copies the theme's uploaded images, so it keeps working if the original is deleted. A dashboard that references a missing theme falls back to Default.</p>

        <h4>Palette roles</h4>
        <p>Colours are hex values (<code>#rgb</code>, <code>#rrggbb</code> or <code>#rrggbbaa</code>); <code>bg</code> must be opaque. Each role is exposed as a CSS variable that you can use in custom CSS.</p>
        <table>
            <thead>
                <tr><th>Role</th><th>CSS variable</th><th>Used for</th></tr>
            </thead>
            <tbody>
                <tr><td><code>bg</code></td><td><code>--dashi-bg</code></td><td>Page base colour.</td></tr>
                <tr><td><code>surface</code></td><td><code>--dashi-surface</code></td><td>Cards and raised areas.</td></tr>
                <tr><td><code>border</code></td><td><code>--dashi-border</code></td><td>Borders and dividers.</td></tr>
                <tr><td><code>text</code></td><td><code>--dashi-text</code></td><td>Primary text.</td></tr>
                <tr><td><code>textMuted</code></td><td><code>--dashi-text-muted</code></td><td>Secondary text.</td></tr>
                <tr><td><code>primary</code></td><td><code>--dashi-primary</code></td><td>Accent colour.</td></tr>
                <tr><td><code>onPrimary</code></td><td><code>--dashi-on-primary</code></td><td>Text on a primary background.</td></tr>
                <tr><td><code>secondary</code></td><td><code>--dashi-secondary</code></td><td>Secondary accent.</td></tr>
                <tr><td><code>success</code></td><td><code>--dashi-success</code></td><td>Positive states.</td></tr>
                <tr><td><code>warning</code></td><td><code>--dashi-warning</code></td><td>Warnings.</td></tr>
                <tr><td><code>danger</code></td><td><code>--dashi-danger</code></td><td>Errors and destructive states.</td></tr>
            </tbody>
        </table>
        <p class="doc-hint">Two more variables exist: <code>--dashi-page-bg</code> (the theme background, only set for a gradient or image) and <code>--dashi-font</code> (the widget theme font). On e-ink image dashboards, translucent colours are flattened over <code>bg</code> into opaque colours.</p>

        <h4>Backgrounds</h4>
        <p>The background is edited in the <strong>Background</strong> section of the theme editor, with a live light and dark preview. Its base is either the palette <code>bg</code> colour or a linear gradient (an angle and two or more evenly spread colour stops, optionally different in dark mode). An image can be painted over it. Pick one from the thumbnails, or use <strong>Upload</strong>, which stores the file in the theme itself. Shared background images are managed under <strong>Theming &rarr; Background images</strong> and can be picked by any theme. Theme images are private to the theme and are copied along with it. For image-type dashboards the background image is drawn onto the canvas scaled to cover it completely; the fit, position and repeat settings apply in browsers only.</p>

        <hr class="doc-divider" />
        <h3 id="theme-widget-themes">Widget Themes</h3>
        <p>A widget theme provides fonts and icons. Each dashboard selects one through its <code>widgetTheme</code> property. The built-in <code>"default"</code> widget theme ships with the Inter font and Tabler icons. A missing or unknown name falls back to it, as does a widget theme without fonts or icons.</p>

        <h4>Widget theme contents</h4>
        <ul>
            <li><strong>Display fonts</strong> &mdash; TTF files used for dashboard text. The first font listed becomes the primary font family.</li>
            <li><strong>Icon set</strong> &mdash; a TTF icon font (e.g. Tabler) or a set of images. Widgets reference icons by name; the widget theme resolves them.</li>
        </ul>

        <h4>Manifest</h4>
        <p>Each widget theme is defined by a <code>widget_theme.yaml</code> file at the root of its folder.</p>
        <table>
            <thead>
                <tr><th>Field</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>name</code></td><td>Widget theme display name.</td></tr>
                <tr><td><code>description</code></td><td>Short description.</td></tr>
                <tr><td><code>fonts</code></td><td>List of display fonts. Each entry has a <code>name</code> and <code>file</code> (TTF path).</td></tr>
                <tr><td><code>icons.type</code></td><td><code>font</code> or <code>image</code>. Determines how icons are rendered.</td></tr>
                <tr><td><code>icons.classPrefix</code></td><td>CSS class prefix for font icons, e.g. <code>"ti ti-"</code>.</td></tr>
                <tr><td><code>icons.fontFile</code></td><td>Path to the icon font TTF file.</td></tr>
                <tr><td><code>icons.icons</code></td><td>Map of icon names to <code>class</code> and <code>codepoint</code> values.</td></tr>
            </tbody>
        </table>

        <h4>Creating a widget theme</h4>
        <p>Use the CLI to bootstrap a new widget theme with a manifest and placeholder files:</p>
        <pre class="doc-code">dashi widget-theme create &lt;name&gt; --type &lt;image|font&gt;</pre>
        <table>
            <thead>
                <tr><th>Flag</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>--type, -t</code></td><td><code>image</code></td><td>Widget theme type. <code>image</code> creates SVG placeholders for all weather icons. <code>font</code> creates a manifest with empty icon class mappings.</td></tr>
                <tr><td><code>--config, -c</code></td><td><code>./config.yaml</code></td><td>Path to the app config file (used to locate the data directory).</td></tr>
            </tbody>
        </table>
        <p>The widget theme is created under <code>&lt;dataDir&gt;/data/widget_themes/&lt;name&gt;/</code>. For image themes, replace the generated SVGs in <code>widgets/weather/icons/</code> with your own. For font themes, edit <code>widget_theme.yaml</code> to set the CSS prefix, font file, and icon mappings.</p>

        <h4>Upload and download</h4>
        <p>Under <strong>Widget themes</strong> in the editor you can upload a widget theme as a zip archive, download an existing one as a zip, or delete it. The archive must contain a <code>widget_theme.yaml</code> manifest; packs from older versions that use <code>theme.yaml</code> are rejected. The built-in default cannot be deleted.</p>

        <hr class="doc-divider" />
        <h3 id="theme-assets">Dashboard Assets</h3>
        <p>Each dashboard has its own asset folder on disk for storing images, fonts, or stylesheets. Assets are uploaded through the editor or the API.</p>
        <table>
            <thead>
                <tr><th>Detail</th><th>Value</th></tr>
            </thead>
            <tbody>
                <tr><td>Allowed file types</td><td><code>.png</code>, <code>.jpg</code>, <code>.jpeg</code>, <code>.svg</code>, <code>.webp</code>, <code>.css</code></td></tr>
                <tr><td>Max upload size</td><td>10 MB per file</td></tr>
                <tr><td>Storage</td><td>Stored inside the dashboard's folder on disk.</td></tr>
            </tbody>
        </table>
        <p>Assets can be organized in subdirectories. Path traversal (<code>..</code>) and overwriting <code>dashboard.json</code> are not allowed.</p>

        <hr class="doc-divider" />
        <h3 id="theme-custom-css">Custom CSS</h3>
        <p>Place a <code>custom.css</code> file in a dashboard's asset folder to inject custom styles. The file is automatically detected and included as an inline <code>&lt;style&gt;</code> block when the dashboard renders. No configuration is needed &mdash; if the file exists, it is applied.</p>
        <p>This is useful for overriding widget styles, adjusting spacing, hiding elements, or any other CSS-level customization that the editor does not expose.</p>

    </section>
</template>

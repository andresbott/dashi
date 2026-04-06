<template>
    <section id="theming">
        <h2>Theming</h2>
        <p>Dashi supports customizing the look and feel of dashboards through themes, custom assets, and CSS overrides.</p>

        <h3 id="theme-themes">Themes</h3>
        <p>A theme is a named package that provides fonts and icons to dashboards. Each dashboard selects a theme via its <code>theme</code> property. The built-in <code>"default"</code> theme ships with the Inter font and Tabler icons.</p>

        <h4>Theme contents</h4>
        <ul>
            <li><strong>Display fonts</strong> &mdash; TTF files used for dashboard text rendering. The first font listed becomes the primary font family.</li>
            <li><strong>Icon font</strong> &mdash; a TTF icon set (e.g. Tabler icons). Widgets reference icons by name; the theme resolves them to CSS classes or Unicode codepoints.</li>
            <li><strong>Background images</strong> &mdash; images bundled with the theme, available to any dashboard using it.</li>
        </ul>

        <h4>Theme manifest</h4>
        <p>Each theme is defined by a <code>theme.yaml</code> file in its directory.</p>
        <table>
            <thead>
                <tr><th>Field</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>name</code></td><td>Theme display name.</td></tr>
                <tr><td><code>description</code></td><td>Short description.</td></tr>
                <tr><td><code>fonts</code></td><td>List of display fonts. Each entry has a <code>name</code> and <code>file</code> (TTF path).</td></tr>
                <tr><td><code>icons.type</code></td><td><code>font</code> or <code>image</code>. Determines how icons are rendered.</td></tr>
                <tr><td><code>icons.classPrefix</code></td><td>CSS class prefix for font icons, e.g. <code>"ti ti-"</code>.</td></tr>
                <tr><td><code>icons.fontFile</code></td><td>Path to the icon font TTF file.</td></tr>
                <tr><td><code>icons.icons</code></td><td>Map of icon names to <code>class</code> and <code>codepoint</code> values.</td></tr>
            </tbody>
        </table>

        <h4>Creating a theme</h4>
        <p>Use the CLI to bootstrap a new theme with a manifest and placeholder files:</p>
        <pre class="doc-code">dashi theme create &lt;name&gt; --type &lt;image|font&gt;</pre>
        <table>
            <thead>
                <tr><th>Flag</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>--type, -t</code></td><td><code>image</code></td><td>Theme type. <code>image</code> creates SVG placeholders for all weather icons. <code>font</code> creates a manifest with empty icon class mappings.</td></tr>
                <tr><td><code>--config, -c</code></td><td><code>./config.yaml</code></td><td>Path to the app config file (used to locate the data directory).</td></tr>
            </tbody>
        </table>
        <p>The theme is created under <code>&lt;dataDir&gt;/themes/&lt;name&gt;/</code>. For image themes, replace the generated SVGs in <code>widgets/weather/icons/</code> with your own. For font themes, edit <code>theme.yaml</code> to set the CSS prefix, font file, and icon mappings.</p>

        <h4>User themes</h4>
        <p>Custom themes placed in the themes directory are loaded automatically at startup. Each theme needs a <code>theme.yaml</code> manifest and any font/background files it references.</p>

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

        <hr class="doc-divider" />
        <h3 id="theme-backgrounds">Backgrounds</h3>
        <p>Dashboard backgrounds can come from two sources, referenced by a prefixed string in the <code>background.value</code> property.</p>
        <table>
            <thead>
                <tr><th>Source</th><th>Format</th><th>Example</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td>Theme</td>
                    <td><code>theme:&lt;name&gt;/&lt;file&gt;</code></td>
                    <td><code>theme:default/bg.jpg</code></td>
                </tr>
                <tr>
                    <td>Dashboard asset</td>
                    <td><code>dashboard:&lt;file&gt;</code></td>
                    <td><code>dashboard:my-bg.png</code></td>
                </tr>
            </tbody>
        </table>
        <p>Theme backgrounds are shared across all dashboards that use that theme. Dashboard asset backgrounds are private to the individual dashboard. For image-type dashboards, backgrounds are rendered as base64 data URIs scaled to cover the full canvas.</p>
    </section>
</template>

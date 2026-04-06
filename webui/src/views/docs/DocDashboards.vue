<template>
    <section id="dashboards">
        <h2>Dashboards</h2>

        <h3 id="dash-types">Types</h3>
        <p>Each dashboard has a type that determines how it is rendered and served.</p>
        <table>
            <thead>
                <tr><th>Type</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>interactive</code></td>
                    <td>SPA rendered in the browser. Supports user interaction, live updates, and the full editor.</td>
                </tr>
                <tr>
                    <td><code>image</code></td>
                    <td>Server-side rendered as PNG. Designed for e-ink displays, kiosk screens, or embedding via URL.</td>
                </tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="dash-properties">Properties</h3>
        <p>Top-level settings that apply to every dashboard regardless of type.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Values</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>name</code></td>
                    <td>string</td>
                    <td>Display name. Also used to derive the storage folder (snake_case).</td>
                </tr>
                <tr>
                    <td><code>icon</code></td>
                    <td>string</td>
                    <td>Icon identifier shown in listings.</td>
                </tr>
                <tr>
                    <td><code>theme</code></td>
                    <td>string</td>
                    <td>Theme name. Defaults to <code>"default"</code>. Controls fonts and icon set.</td>
                </tr>
                <tr>
                    <td><code>colorMode</code></td>
                    <td><code>auto</code> | <code>light</code> | <code>dark</code></td>
                    <td>Color scheme. <code>auto</code> follows system preference.</td>
                </tr>
                <tr>
                    <td><code>accentColor</code></td>
                    <td>hex color</td>
                    <td>Custom accent color override.</td>
                </tr>
                <tr>
                    <td><code>imageConfig.width</code></td>
                    <td>integer (default: 1024)</td>
                    <td>Image type only. Render width in pixels.</td>
                </tr>
                <tr>
                    <td><code>imageConfig.height</code></td>
                    <td>integer (default: auto)</td>
                    <td>Image type only. Render height. Auto-calculated from content if omitted.</td>
                </tr>
            </tbody>
        </table>

        <h4>Container</h4>
        <p>Controls the overall layout of the dashboard content area.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Values</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>maxWidth</code></td>
                    <td>CSS value</td>
                    <td>Maximum width of the content area.</td>
                </tr>
                <tr>
                    <td><code>verticalAlign</code></td>
                    <td><code>top</code> | <code>center</code> | <code>bottom</code></td>
                    <td>Vertical alignment of content.</td>
                </tr>
                <tr>
                    <td><code>horizontalAlign</code></td>
                    <td><code>left</code> | <code>center</code> | <code>right</code></td>
                    <td>Horizontal alignment of content.</td>
                </tr>
                <tr>
                    <td><code>showBoxes</code></td>
                    <td>boolean</td>
                    <td>Debug mode: renders colored borders around widgets.</td>
                </tr>
            </tbody>
        </table>

        <h4>Background</h4>
        <p>Sets the full-screen background behind the dashboard content. Images can come from the dashboard's own assets or from a theme.</p>
        <table>
            <thead>
                <tr><th>Type</th><th>Value</th></tr>
            </thead>
            <tbody>
                <tr><td><code>none</code></td><td>No background (white).</td></tr>
                <tr><td><code>color</code></td><td>Hex color, e.g. <code>#1a1a2e</code>.</td></tr>
                <tr><td><code>gradient</code></td><td>CSS gradient string.</td></tr>
                <tr><td><code>image</code></td><td><code>dashboard:filename.jpg</code> or <code>theme:name/file.png</code>.</td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="dash-pages">Pages</h3>
        <p>A dashboard is organized into one or more pages. Each page holds rows, and each row holds widgets arranged in a 12-column grid.</p>
        <ul>
            <li>Dashboards support multiple pages.</li>
            <li>Each page has a <code>name</code> and an array of <code>rows</code>.</li>
            <li>Each row contains <code>widgets</code> and has optional <code>title</code>, <code>height</code>, and <code>width</code>.</li>
            <li>Widget <code>width</code> is a column span from 1 to 12 (12-column grid).</li>
        </ul>

        <hr class="doc-divider" />
        <h3 id="dash-query-params">Query Parameters</h3>
        <p>Append to the dashboard URL: <code>/&lt;dashboard-id&gt;?param=value</code></p>
        <table>
            <thead>
                <tr><th>Parameter</th><th>Value</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>page</code></td>
                    <td>integer (0-based)</td>
                    <td>Select which page to display. Defaults to 0.</td>
                </tr>
                <tr>
                    <td><code>debug</code></td>
                    <td><code>1</code></td>
                    <td>Enable debug mode: shows colored borders around widget boxes.</td>
                </tr>
                <tr>
                    <td><code>html</code></td>
                    <td>(flag)</td>
                    <td>Image dashboards only. Returns HTML instead of PNG.</td>
                </tr>
            </tbody>
        </table>
        <p class="doc-hint">All other query parameters are passed through to widgets.</p>

        <hr class="doc-divider" />
        <h3 id="dash-id">Dashboard ID</h3>
        <p>Every dashboard is identified by a short, unique ID that is part of its URL.</p>
        <ul>
            <li>6 characters, lowercase alphanumeric (<code>a-z0-9</code>).</li>
            <li>Auto-generated on creation.</li>
            <li>Preview dashboards get a <code>-prev</code> suffix.</li>
        </ul>
    </section>
</template>

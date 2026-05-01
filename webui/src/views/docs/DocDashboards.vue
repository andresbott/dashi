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
                    <td>Server-side rendered to an image. Supports multiple output formats (PNG, B&amp;W, 7-color) for e-ink displays, kiosk screens, or embedding via URL. When accessed in a browser without display parameters, shows an HTML preview.</td>
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
                    <td><code>pages[].refreshInterval</code></td>
                    <td>integer (default: 0)</td>
                    <td>Image type only. Per-page refresh interval in seconds. Returned as <code>X-Refresh-Interval</code> response header so display devices know when to re-fetch. 0 = no header sent.</td>
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
            </tbody>
        </table>
        <p class="doc-hint">All other query parameters are passed through to widgets.</p>

        <h4>Display Parameters (image dashboards)</h4>
        <p>Image dashboards accept display parameters as either HTTP headers or query parameters. Headers take precedence over query parameters when both are provided.</p>
        <table>
            <thead>
                <tr><th>Header</th><th>Query Parameter</th><th>Values</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>X-Display-Format</code></td>
                    <td><code>format</code></td>
                    <td><code>png</code>, <code>png-bw</code>, <code>png-7color</code>, <code>bw</code>, <code>7color</code></td>
                    <td>Output format. <code>png</code> = full color PNG. <code>png-bw</code> = dithered B&amp;W PNG. <code>png-7color</code> = 7-color palette PNG. <code>bw</code> = raw 1bpp packed binary. <code>7color</code> = raw 4bpp packed binary.</td>
                </tr>
                <tr>
                    <td><code>X-Display-Width</code></td>
                    <td><code>width</code></td>
                    <td>positive integer</td>
                    <td>Native panel width in pixels.</td>
                </tr>
                <tr>
                    <td><code>X-Display-Height</code></td>
                    <td><code>height</code></td>
                    <td>positive integer</td>
                    <td>Native panel height in pixels.</td>
                </tr>
                <tr>
                    <td><code>X-Display-Rotation</code></td>
                    <td><code>rotation</code></td>
                    <td><code>0</code>, <code>90</code>, <code>180</code>, <code>270</code></td>
                    <td>Rotation in degrees. The server renders content rotated by the given amount so that pixel (0,0) maps to the panel's native top-left. For 90°/270° the content is laid out in portrait (H×W) then rotated to fit the native landscape (W×H) output. Defaults to <code>0</code>.</td>
                </tr>
                <tr>
                    <td><code>X-Action</code></td>
                    <td><code>action</code></td>
                    <td><code>refresh</code>, <code>swipe_right</code>, <code>swipe_left</code></td>
                    <td>Navigation action. <code>swipe_right</code>/<code>swipe_left</code> redirect to next/previous page. Defaults to <code>refresh</code>.</td>
                </tr>
            </tbody>
        </table>
        <p class="doc-hint">Example: <code>/my-dash?format=png&amp;width=800&amp;height=480&amp;rotation=90</code> renders the dashboard in portrait layout, rotated 90° to fill the native 800×480 panel.</p>

        <h4>Response Headers</h4>
        <table>
            <thead>
                <tr><th>Header</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>X-Refresh-Interval</code></td>
                    <td>Seconds until the display should re-fetch the image. Only set when the page has a non-zero <code>refreshInterval</code>.</td>
                </tr>
            </tbody>
        </table>

        <h4>Output Formats</h4>
        <table>
            <thead>
                <tr><th>Format</th><th>Content-Type</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>png</code></td>
                    <td><code>image/png</code></td>
                    <td>Full-color RGBA PNG image.</td>
                </tr>
                <tr>
                    <td><code>png-bw</code></td>
                    <td><code>image/png</code></td>
                    <td>Floyd-Steinberg dithered black &amp; white, encoded as grayscale PNG.</td>
                </tr>
                <tr>
                    <td><code>png-7color</code></td>
                    <td><code>image/png</code></td>
                    <td>Dithered to a 7-color e-ink palette (black, white, green, blue, red, yellow, orange), encoded as PNG.</td>
                </tr>
                <tr>
                    <td><code>bw</code></td>
                    <td><code>application/octet-stream</code></td>
                    <td>Raw 1-bit-per-pixel packed binary (MSB first). Size: <code>ceil(width/8) × height</code> bytes.</td>
                </tr>
                <tr>
                    <td><code>7color</code></td>
                    <td><code>application/octet-stream</code></td>
                    <td>Raw 4-bit-per-pixel packed binary (two pixels per byte). Size: <code>(width × height) / 2</code> bytes.</td>
                </tr>
            </tbody>
        </table>
        <p class="doc-hint">The <code>bw</code> and <code>7color</code> raw formats are optimized for direct transfer to e-ink display controllers without additional decoding.</p>

        <h4>Browser Preview</h4>
        <p>When an image dashboard is accessed without any display parameters (no headers, no <code>format</code>/<code>width</code>/<code>height</code> query params), the server returns an HTML preview of the rendered page instead of an image. This allows viewing the dashboard in a regular browser for debugging and development.</p>

        <hr class="doc-divider" />
        <h3 id="dash-id">Dashboard ID</h3>
        <p>Every dashboard is identified by a short, unique ID that is part of its URL.</p>
        <ul>
            <li>6 characters, lowercase alphanumeric (<code>a-z0-9</code>).</li>
            <li>Auto-generated on creation.</li>
            <li>Preview dashboards get a <code>-prev</code> suffix.</li>
        </ul>
        <hr class="doc-divider" />
        <h3 id="dash-server-modes">Server Modes</h3>
        <p>Dashi runs the viewer and editor as separate HTTP servers on different ports. Each can be independently enabled or disabled in the configuration.</p>
        <table>
            <thead>
                <tr><th>Server</th><th>Default Port</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>Viewer</code></td>
                    <td>8087</td>
                    <td>Read-only. Serves dashboards and GET-only API endpoints. No access to the editor, dashboard list, or documentation pages.</td>
                </tr>
                <tr>
                    <td><code>Editor</code></td>
                    <td>8088</td>
                    <td>Full access. Serves the dashboard list, editor, documentation, and all API endpoints including create, update, delete, upload, and import.</td>
                </tr>
            </tbody>
        </table>
        <p>At least one server must be enabled. When both are enabled they share the same data directory and caches. Configuration example:</p>
        <pre class="doc-code">Server:
  Viewer:
    Enabled: true
    BindIp: ""
    Port: 8087
  Editor:
    Enabled: true
    BindIp: ""
    Port: 8088</pre>
        <p>A typical deployment exposes the viewer port publicly and restricts editor access to a private network or VPN.</p>

        <hr class="doc-divider" />
        <h3 id="dash-export-import">Export &amp; Import</h3>
        <p>Dashboards can be exported as zip archives and imported back into the same or a different Dashi instance.</p>
        <h4>Export</h4>
        <ul>
            <li>From the dashboard list, click the download button on any dashboard card.</li>
            <li>The zip contains <code>dashboard.json</code> and all asset files.</li>
        </ul>
        <h4>Import</h4>
        <ul>
            <li>From the dashboard list, click <strong>Import</strong> and select a <code>.zip</code> file.</li>
            <li>A new dashboard is created with a fresh ID. The name and configuration are read from the zip.</li>
        </ul>

        <hr class="doc-divider" />
        <h3 id="dash-file-upload">File Upload</h3>
        <p>Assets (images, stylesheets) can be uploaded directly from the dashboard editor.</p>
        <ul>
            <li>In the editor toolbar, click <strong>Upload</strong> to open the upload dialog.</li>
            <li>Select a file and confirm. The file is stored in the dashboard's asset folder on disk.</li>
        </ul>
        <table>
            <thead>
                <tr><th>Detail</th><th>Value</th></tr>
            </thead>
            <tbody>
                <tr><td>Accepted file types</td><td><code>.png</code>, <code>.jpg</code>, <code>.jpeg</code>, <code>.svg</code>, <code>.webp</code>, <code>.css</code></td></tr>
                <tr><td>Max file size</td><td>10 MB per file</td></tr>
            </tbody>
        </table>
        <p>Uploaded assets can be used as dashboard backgrounds (<code>dashboard:filename.jpg</code>) or referenced in custom CSS.</p>
    </section>
</template>

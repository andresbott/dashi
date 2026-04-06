<template>
    <section id="widgets">
        <h2>Widgets</h2>
        <p>Widgets are the building blocks of a dashboard. Each widget occupies one or more columns in a 12-column grid row. Widgets can fetch live data from external APIs or render purely client-side.</p>

        <h3 id="widget-clock">Clock</h3>
        <p>Displays a digital clock with optional date. The time updates every second in interactive dashboards.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>hour12</code></td><td>boolean</td><td><code>false</code></td><td>Use 12-hour format instead of 24-hour.</td></tr>
                <tr><td><code>showSeconds</code></td><td>boolean</td><td><code>true</code></td><td>Show seconds in the time display.</td></tr>
                <tr><td><code>showDate</code></td><td>boolean</td><td><code>true</code></td><td>Show the date below the time.</td></tr>
                <tr><td><code>font</code></td><td>string</td><td>&mdash;</td><td>Font name from the dashboard's theme. Falls back to the theme's default font.</td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="widget-weather">Weather</h3>
        <p>Comprehensive weather widget showing current conditions, hourly and daily forecasts, and optional temperature/rain graphs. Data is fetched from the Open-Meteo API based on geographic coordinates.</p>
        <p>See also <strong>Weather Compact</strong> for a minimal single-line variant.</p>

        <h4>Location</h4>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>city</code></td><td>string</td><td>Display name of the location.</td></tr>
                <tr><td><code>latitude</code></td><td>number</td><td>Latitude coordinate (required).</td></tr>
                <tr><td><code>longitude</code></td><td>number</td><td>Longitude coordinate (required).</td></tr>
            </tbody>
        </table>

        <h4>Display toggles</h4>
        <table>
            <thead>
                <tr><th>Property</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>showCurrent</code></td><td><code>true</code></td><td>Show current conditions (temperature, icon, description).</td></tr>
                <tr><td><code>showDetails</code></td><td><code>true</code></td><td>Show feels-like temperature, humidity, and wind in the current section.</td></tr>
                <tr><td><code>showHourly</code></td><td><code>false</code></td><td>Show hourly forecast slots.</td></tr>
                <tr><td><code>showForecast</code></td><td><code>true</code></td><td>Show multi-day forecast.</td></tr>
                <tr><td><code>showGraph</code></td><td><code>false</code></td><td>Show temperature and rain graph.</td></tr>
            </tbody>
        </table>

        <h4>Extra info toggles</h4>
        <p>Optional data points displayed alongside current conditions when enabled.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>showSunrise</code></td><td>Sunrise time.</td></tr>
                <tr><td><code>showSunset</code></td><td>Sunset time.</td></tr>
                <tr><td><code>showWind</code></td><td>Wind speed.</td></tr>
                <tr><td><code>showHumidity</code></td><td>Humidity percentage.</td></tr>
                <tr><td><code>showPressure</code></td><td>Atmospheric pressure (hPa).</td></tr>
                <tr><td><code>showUV</code></td><td>UV index.</td></tr>
                <tr><td><code>showVisibility</code></td><td>Visibility distance (km).</td></tr>
                <tr><td><code>showAirQuality</code></td><td>Air quality index (AQI).</td></tr>
            </tbody>
        </table>

        <h4>Hourly &amp; forecast</h4>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>hourlyCount</code></td><td>number</td><td><code>12</code></td><td>Hours ahead to fetch (1&ndash;24).</td></tr>
                <tr><td><code>hourlySlots</code></td><td>number</td><td><code>6</code></td><td>Number of time slots to display (1&ndash;12).</td></tr>
                <tr><td><code>forecastDays</code></td><td>number</td><td><code>7</code></td><td>Days to show in the daily forecast (1&ndash;7).</td></tr>
            </tbody>
        </table>

        <h4>Graph options</h4>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>graphHours</code></td><td>number</td><td><code>24</code></td><td>Hours of data shown in the graph (3&ndash;24).</td></tr>
                <tr><td><code>graphHeight</code></td><td>number</td><td><code>250</code></td><td>Graph height in pixels (100&ndash;600).</td></tr>
                <tr><td><code>graphTempColor</code></td><td>hex</td><td><code>#FF8C42</code></td><td>Color of the temperature line.</td></tr>
                <tr><td><code>graphRainColor</code></td><td>hex</td><td><code>#4A90D9</code></td><td>Color of the rain bars.</td></tr>
                <tr><td><code>graphBgColor</code></td><td>hex</td><td>transparent</td><td>Graph background color.</td></tr>
                <tr><td><code>graphShowTemp</code></td><td>boolean</td><td><code>true</code></td><td>Show temperature labels on the graph.</td></tr>
                <tr><td><code>graphShowRain</code></td><td>boolean</td><td><code>true</code></td><td>Show rain percentage labels.</td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="widget-weather-compact">Weather Compact</h3>
        <p>A minimal single-line weather display showing an icon, temperature, and optional city name. Uses the same data source as the full Weather widget but designed for tight spaces.</p>
        <p>Location is configured the same way as the full Weather widget (<code>city</code>, <code>latitude</code>, <code>longitude</code>).</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>compactCity</code></td><td><code>true</code></td><td>Show the city name.</td></tr>
                <tr><td><code>compactFeelsLike</code></td><td><code>false</code></td><td>Show feels-like temperature.</td></tr>
                <tr><td><code>compactDescription</code></td><td><code>true</code></td><td>Show weather description text.</td></tr>
                <tr><td><code>compactAlign</code></td><td><code>left</code></td><td>Alignment: <code>left</code>, <code>center</code>, or <code>right</code>.</td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="widget-bookmark">Bookmark</h3>
        <p>A clickable link to an external website, displayed with an icon and title. Useful for building a personal start page or quick-access panel.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>url</code></td><td>string</td><td>Destination URL. Only <code>http://</code>, <code>https://</code>, and relative paths are allowed.</td></tr>
                <tr><td><code>icon</code></td><td>string</td><td>Icon identifier (see icon types below).</td></tr>
                <tr><td><code>title</code></td><td>string</td><td>Primary display text.</td></tr>
                <tr><td><code>subtitle</code></td><td>string</td><td>Secondary text below the title (optional).</td></tr>
                <tr><td><code>textBelow</code></td><td>boolean</td><td>Layout mode. <code>false</code>: icon left, text right. <code>true</code>: icon top, text below.</td></tr>
            </tbody>
        </table>

        <h4>Icon types</h4>
        <p>The icon string determines where the icon is loaded from based on its prefix.</p>
        <table>
            <thead>
                <tr><th>Prefix</th><th>Source</th><th>Example</th></tr>
            </thead>
            <tbody>
                <tr><td><code>ti-*</code></td><td>Tabler icon from the theme's icon font.</td><td><code>ti-home</code></td></tr>
                <tr><td><code>selfhst:</code></td><td>Icon from the selfh.st CDN.</td><td><code>selfhst:sonarr.png</code></td></tr>
                <tr><td><code>dashboard:</code></td><td>Image file from the dashboard's asset folder.</td><td><code>dashboard:logo.svg</code></td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="widget-search">Search</h3>
        <p>A search bar that opens results in a new browser tab. Purely client-side &mdash; no data is fetched from the server.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>engine</code></td><td>string</td><td><code>google</code></td><td>Search engine: <code>google</code>, <code>duckduckgo</code>, or <code>bing</code>.</td></tr>
                <tr><td><code>placeholder</code></td><td>string</td><td><code>Search...</code></td><td>Placeholder text for the input field.</td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="widget-market">Market</h3>
        <p>Displays a stock or cryptocurrency ticker with current price, change percentage, and an optional price history chart. Data is fetched from Yahoo Finance.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>symbol</code></td><td>string</td><td>&mdash;</td><td>Ticker symbol, e.g. <code>AAPL</code>, <code>BTC-USD</code>.</td></tr>
                <tr><td><code>range</code></td><td>string</td><td><code>1mo</code></td><td>Time period: <code>1d</code>, <code>5d</code>, <code>1mo</code>, <code>3mo</code>, <code>6mo</code>, <code>1y</code>.</td></tr>
                <tr><td><code>showChart</code></td><td>boolean</td><td><code>true</code></td><td>Display the price history chart.</td></tr>
                <tr><td><code>showChange</code></td><td>boolean</td><td><code>true</code></td><td>Show price change amount and percentage.</td></tr>
            </tbody>
        </table>
        <p class="doc-hint">The chart is rendered server-side as a PNG image. Price changes are colored green (positive) or red (negative).</p>

        <hr class="doc-divider" />
        <h3 id="widget-xkcd">XKCD</h3>
        <p>Displays an XKCD webcomic. Clicking the comic opens it on xkcd.com.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>mode</code></td><td>string</td><td><code>latest</code></td><td>Display mode: <code>latest</code> (most recent comic), <code>random</code> (daily random, same all day), or <code>random-each</code> (new random on every load).</td></tr>
            </tbody>
        </table>

        <hr class="doc-divider" />
        <h3 id="widget-departures">SBB Departures</h3>
        <p>Shows upcoming public transport departures for a given station using the Swiss Federal Railways (SBB) API. Displays line number, destination, departure countdown, and platform.</p>
        <table>
            <thead>
                <tr><th>Property</th><th>Type</th><th>Default</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>stationId</code></td><td>string</td><td>&mdash;</td><td>Station identifier from the transport API.</td></tr>
                <tr><td><code>stationName</code></td><td>string</td><td>&mdash;</td><td>Display name of the station.</td></tr>
                <tr><td><code>limit</code></td><td>number</td><td><code>5</code></td><td>Number of departures to show (1&ndash;15).</td></tr>
            </tbody>
        </table>
        <p class="doc-hint">Departures within 60 minutes show a countdown (e.g. "15'"). Beyond that, the scheduled time is shown (HH:MM). Delayed departures are highlighted.</p>

        <hr class="doc-divider" />
        <h3 id="widget-battery">Battery</h3>
        <p>Displays a battery level indicator. The value is read from the <code>battery</code> URL query parameter, making it suitable for external devices that pass their battery level via URL.</p>
        <table>
            <thead>
                <tr><th>Source</th><th>Description</th></tr>
            </thead>
            <tbody>
                <tr><td><code>?battery=75</code></td><td>Value from 0 to 100. Clamped to valid range.</td></tr>
            </tbody>
        </table>
        <p class="doc-hint">This widget has no configuration properties. It reads its value entirely from the URL.</p>

        <hr class="doc-divider" />
        <h3 id="widget-page-indicator">Page Indicator</h3>
        <p>Renders navigation dots for multi-page dashboards. One dot per page; the active page is highlighted. Useful for dashboards displayed on screens where swipe or keyboard navigation is available.</p>
        <p class="doc-hint">This widget has no configuration properties. It automatically reflects the current page state.</p>

        <hr class="doc-divider" />
        <h3 id="widget-stack">Stack</h3>
        <p>A container widget that stacks multiple child widgets vertically within a single grid cell. Use it to group related widgets or fit more content into a row without splitting into separate columns.</p>
        <ul>
            <li>Child widgets are arranged top to bottom.</li>
            <li>Each child widget can be any type except another stack.</li>
            <li>In the editor, children can be added, removed, reordered via drag-and-drop, and individually configured.</li>
        </ul>
    </section>
</template>

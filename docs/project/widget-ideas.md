# Widget Ideas

Potential new widgets for dashi.

## Calendar

Display upcoming events from calendar sources (iCal/CalDAV). Show today's agenda or a mini month view. Could support multiple calendar feeds with color coding.

## RSS Feed

Show recent items from one or more RSS/Atom feeds. Configurable number of items, show/hide descriptions, thumbnail support. Useful for news, blog posts, or release tracking.

## Stock Tracking

Display stock prices, crypto, or currency exchange rates. Show current price, daily change (absolute and percentage), and a mini sparkline chart. Could pull from a free API like Yahoo Finance or CoinGecko.

## Nested Include

Embed another dashboard inside a widget. Allows composing complex layouts from reusable building blocks. A dashboard becomes a component that can be included in multiple places.

## Tab View

A container widget that holds multiple child widgets as tabs. Only one tab is visible at a time. Useful for packing more information into limited space without scrolling (e.g. tabs for different bookmark groups or different city weather views).

## Google Search

A search bar widget that submits queries to Google. Simple input field with a search button — typing and pressing enter or clicking opens Google search results in a new tab. Could support other search engines (DuckDuckGo, Bing) as a config option.

## System Stats

Display host system statistics such as CPU usage, memory usage, and disk/storage utilization. Could show values as progress bars or gauges with percentage labels. Configurable to show/hide individual metrics. Useful for monitoring the machine running dashi at a glance.

## External Data Display

A generic widget that renders data pushed from an external client via API. The client sends key-value pairs (e.g. air quality index, real-time temperature from a sensor, humidity) and the widget displays them. Supports configurable labels, units, and refresh intervals. Useful for IoT sensors, home automation, or any custom data source that wants to publish to the dashboard.

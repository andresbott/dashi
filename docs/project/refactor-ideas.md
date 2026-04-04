# Refactor Ideas

Ideas for major architectural changes to analyze before committing to implementation.

## Vanilla JS Widget Rendering

Currently the SPA (Vue) handles both dashboard editing and dashboard rendering (display mode). Consider splitting these concerns: the SPA stays for the editing experience, but the actual dashboard rendering is done with vanilla JS. Each widget would independently handle its own rendering and data fetching (AJAX calls, polling, etc.) without requiring the Vue runtime.

**Potential benefits:**
- Lighter production page load — no Vue bundle needed for display mode
- Each widget is self-contained and can manage its own lifecycle and refresh logic
- Static rendering from the Go backend could progressively enhance with per-widget JS
- Simpler mental model: edit mode = SPA, display mode = plain HTML + vanilla JS widgets

**Open questions to analyze:**
- How much of the current Vue widget code could be reused or easily ported?
- What's the best boundary between server-rendered HTML and client-side JS enhancement?
- How would shared concerns (theming, layout grid, responsive behavior) work outside Vue?
- Would a web components approach give the isolation benefits while keeping some structure?

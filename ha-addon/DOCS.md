# Dashi

Self-hosted dashboard with configurable widgets.

## Install

1. Settings → Add-ons → Add-on store → ⋮ → **Repositories**.
2. Add `https://github.com/andresbott/dashi` and refresh.
3. Install **Dashi**, then **Start**.

## Access

- **Editor** (admin): the **Open Web UI** button / sidebar panel — authenticated by Home Assistant via ingress.
- **Viewer** (read-only): `http://<your-ha-host>:8087` — for browsers and for ESP32 / e-ink displays that fetch dashboards or PNGs. Optionally protect individual dashboards with per-dashboard basic auth from the editor.

## Options

| Option | Default | Notes |
|---|---|---|
| `log_level` | `info` | `trace`…`fatal`. |
| `viewer_port` | `8087` | Host port for the read-only viewer. |
| `viewer_public_url` | _(empty)_ | Set only if the auto-detected viewer URL is wrong (e.g. a custom hostname), e.g. `http://dashi.local:8087`. |

Dashboards, uploads and themes are stored in the add-on's `/data` volume and are included in Home Assistant backups.

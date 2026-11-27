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
| `editor_public_url` | _(empty)_ | The editor's URL as your browser reaches it, e.g. `http://homeassistant.local:8123/hassio/ingress/<slug>`. Needed for the viewer's **Edit** button (see below); empty means no button. |

## Edit button on the viewer

A dashboard can show an **Edit** button in the viewer (dashboard settings → General, interactive dashboards only; never drawn on image/PNG output). It opens that dashboard in the editor. Because the editor runs behind ingress, the add-on cannot work out its URL — set `editor_public_url` to

    <your Home Assistant URL>/hassio/ingress/<slug>

`<slug>` is the add-on's slug, which is `dashi` prefixed by a repository id (e.g. `a1b2c3d4_dashi`) when installed from this repository. The add-on log prints the exact value at startup while the option is empty, and it is also the last part of the URL when you open the Dashi panel.

The button links to `…/hassio/ingress/<slug>/dashboards/<id>/edit`. Whether Home Assistant forwards that sub-path into the ingress panel has not been verified; if it does not, the button opens the editor's home page instead of the dashboard.

The editor is protected here by Home Assistant's ingress login. Do not point `editor_public_url` at an address that bypasses ingress: the editor has no login of its own, so it would be open to anyone who can reach it.

Dashboards, uploads and themes are stored in the add-on's `/data` volume and are included in Home Assistant backups.

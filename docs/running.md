# Running Dashi

## Ports

`dashi start` runs two servers:

- The editor, on port `8088`, where you create and edit dashboards, themes
  and uploads.
- The viewer, on port `8087`, which serves read-only dashboards to browsers
  and PNGs to e-ink and ESP32 displays.

Data is stored under `./data` by default.

## Good to know

- Dashboards follow the OS light or dark preference by default; each one can
  be pinned to light or dark instead.
- The default widget theme uses the Inter font and Tabler icons.
- Each dashboard can carry a `custom.css` file for styling the editor doesn't cover.
- Weather and market data are fetched at startup for the dashboards that use
  them, then cached (weather for 30 minutes, market data from 15 minutes to
  24 hours depending on the chart range).

## Configuration

Write a commented config file with every default value to `./config.yaml`
(use `-o` for another path):

```sh
dashi config
```

`dashi start` reads `./config.yaml` when it exists; pass another file with
`-c`. Configuration is loaded in this order, the last one wins:

1. Built-in defaults
2. `.env` file (optional)
3. `config.yaml` (optional)
4. Environment variables with the `DASHI_` prefix, e.g. `DASHI_SERVER_VIEWER_PORT=8090`

Behind a reverse proxy, set `Server.Viewer.PublicUrl` so the editor links to
the right viewer URL.

`Server.Editor.PublicUrl` (`DASHI_SERVER_EDITOR_PUBLICURL`) is the URL
browsers use to reach the editor, e.g. `https://edit.example.com`. It is never
guessed, and it is needed for the **Edit** button: a dashboard with the Edit
button setting turned on (dashboard settings, General section; interactive
dashboards only)
shows the button in the viewer, linking to
`<Server.Editor.PublicUrl>/dashboards/<id>/edit`. Left empty, or with the
editor disabled, no button is drawn. Image dashboards never get one, and it
never appears in PNG output.

Both public URLs must be absolute `http://` or `https://` URLs; a trailing
slash is removed, and any other value is ignored with a warning at startup.

The editor has no login of its own. Set `Server.Editor.PublicUrl` only when
the editor is behind something that authenticates users (a reverse proxy with
auth, Home Assistant ingress): the button hands its address to everyone who
can open the dashboard.

## Running as a systemd user service

Run Dashi as a systemd user service so it starts automatically at login.

1. Create `~/.config/systemd/user/dashi.service`:

   ```ini
   [Unit]
   Description=Dashi Dashboard
   After=network-online.target
   Wants=network-online.target

   [Service]
   Type=simple
   WorkingDirectory=/path/to/dashi
   ExecStart=/path/to/dashi/dashi start -c /path/to/dashi/config.yaml
   Restart=on-failure
   RestartSec=5

   [Install]
   WantedBy=default.target
   ```

2. Enable and start it:

   ```sh
   systemctl --user daemon-reload
   systemctl --user enable --now dashi
   ```

3. Optionally, keep it running when you are logged out:

   ```sh
   loginctl enable-linger $USER
   ```

Useful commands:

```sh
systemctl --user status dashi     # check status
systemctl --user restart dashi    # restart
journalctl --user -u dashi -f     # follow logs
```

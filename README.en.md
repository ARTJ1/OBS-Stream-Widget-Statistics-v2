# OBS Stream Widget Statistics v2

A local stream stats widget: wins, losses, and rank. Nothing in the cloud. Your OBS Browser Source doesn’t flicker on every update — changes go over a WebSocket.

[Русский README](README.md)

## What’s in the box

A Windows tray exe, a browser admin UI, and an optional OBS Lua script.

- Run: `widget-stats.exe`
- Overlay: `http://127.0.0.1:19123/overlay/`
- Admin: `http://127.0.0.1:19123/admin/`
- Hotkeys / auto-start: `obs/widget_control.lua`

The admin UI can switch between English and Russian.

## Setup (about two minutes)

1. Grab `widget-stats.exe` and optionally `widget_control.lua` from [Releases](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases).
2. Launch the exe. Tray icon → Open Admin.
3. In OBS: **Tools → WebSocket Server Settings** (port `4455`, password optional).
4. In admin: **Connect OBS** → pick a scene → **Place widget on scene**.

If the Lua script sits next to the exe and is loaded in OBS Scripts, the server starts with OBS and stops when OBS closes. Hotkeys don’t flash a black cmd window.

## Antivirus / SmartScreen

The exe is **not code-signed**. Windows Defender and SmartScreen often complain about an “unknown publisher.” That’s common for small open-source tools without a paid certificate — not proof of malware.

Source is in this repo. You can:
- click “More info” → “Run anyway”;
- add an exclusion;
- or build it yourself: `.\scripts\build.ps1`.

## Stream Deck and friends

Anything that can hit a local HTTP endpoint works:

```
http://127.0.0.1:19123/api/win
http://127.0.0.1:19123/api/loss
http://127.0.0.1:19123/api/rank/up
http://127.0.0.1:19123/api/rank/down
http://127.0.0.1:19123/api/reset
```

GET or POST — both fine. Check the admin (or `data/runtime.json`) if port `19123` is taken.

No HTTP on the deck? Bind the same hotkeys you use in OBS.

## Build

Needs [Go](https://go.dev/dl/) 1.22+.

```powershell
.\scripts\build.ps1
```

You’ll get `widget-stats.exe` next to the script. A `data/` folder appears on first run.

## API cheat sheet

| | |
|---|---|
| `/api/win` `/api/loss` | +1 |
| `/api/rank/up` `/api/rank/down` | rank |
| `/api/reset` | clear W/L |
| `/api/state` `/api/settings` | read / write |
| `/api/runtime` | current URLs |
| `/ws` | live push |
| `/health` | is the server up |

## vs v1

v1 rewrote the Browser Source URL and the page blinked. v2 keeps one localhost overlay and pushes updates over WebSocket.

## License

MIT

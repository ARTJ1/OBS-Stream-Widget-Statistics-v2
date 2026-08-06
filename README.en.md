# OBS Stream Widget Statistics v2

A local application for displaying stream statistics in OBS Studio: wins, losses, and current rank. Updates are delivered over WebSocket without reloading the Browser Source.

[Русская версия](README.md)

## Features

- Local Windows server with a system tray icon
- Overlay for OBS Browser Source
- Web-based admin panel for appearance, motion effects, and skins
- Optional Lua script: auto-start, hotkeys, and shutdown when OBS closes
- Admin UI language: English / Russian
- Compatible with Stream Deck, Bitfocus Companion, and other HTTP clients

| Component | Path / URL |
|-----------|------------|
| Application | `widget-stats.exe` |
| Overlay | `http://127.0.0.1:19123/overlay/` |
| Admin | `http://127.0.0.1:19123/admin/` |
| OBS script | `obs/widget_control.lua` |

Release builds are published under [Releases](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases).

## Installation

1. Download `widget-stats.exe` (and optionally `widget_control.lua`) from [Releases](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases).
2. Run `widget-stats.exe`. Open the admin panel via **Open Admin** in the tray menu.
3. In OBS Studio, enable **Tools → WebSocket Server Settings** (default port `4455`; password optional).
4. In the admin panel, click **Connect OBS**, select a scene, then **Place widget on scene**.

The source is created only on the selected scene. Placing it again does not create duplicates.

### OBS script (recommended)

Add `widget_control.lua` under **Tools → Scripts**. When the script is located next to `widget-stats.exe`:

- the server starts when OBS starts;
- it stops when OBS closes;
- hotkeys run without showing a command-prompt window.

## Security software notices

The executable is **not signed** with a code-signing certificate. Windows SmartScreen, Microsoft Defender, and third-party antivirus products may warn about an unknown publisher.

This is expected for open-source software distributed without a commercial certificate and does not, by itself, indicate malicious content. The source code is available in this repository.

Recommended steps:

1. In the SmartScreen dialog: **More info** → **Run anyway**.
2. If needed, add the file or folder to antivirus exclusions.
3. Build the application from source (see Build).

## Stream Deck and similar devices

HTTP clients may call the local API with GET or POST:

```
http://127.0.0.1:19123/api/win
http://127.0.0.1:19123/api/loss
http://127.0.0.1:19123/api/rank/up
http://127.0.0.1:19123/api/rank/down
http://127.0.0.1:19123/api/reset
```

If port `19123` is unavailable, the current base URL is shown in the admin panel and in `data/runtime.json`.

Devices without HTTP support can use the same hotkeys configured in OBS.

## Build

Requires [Go](https://go.dev/dl/) 1.22 or newer.

```powershell
.\scripts\build.ps1
```

This produces `widget-stats.exe`. The `data/` directory is created on first launch.

## HTTP API

| Method | Path | Description |
|--------|------|-------------|
| GET / POST | `/api/win` | Increment wins |
| GET / POST | `/api/loss` | Increment losses |
| GET / POST | `/api/rank/up` | Rank up |
| GET / POST | `/api/rank/down` | Rank down |
| GET / POST | `/api/reset` | Reset W/L |
| GET | `/api/state` | Current state |
| GET / PUT | `/api/settings` | Settings |
| GET | `/api/runtime` | Base URL and port |
| GET | `/api/snapshot` | State and settings |
| WS | `/ws` | Real-time updates |
| GET | `/health` | Health check |

## Differences from version 1

Version 1 updated statistics by changing the Browser Source URL, which caused visual flicker. Version 2 keeps a persistent localhost overlay and pushes updates over WebSocket.

## License

MIT

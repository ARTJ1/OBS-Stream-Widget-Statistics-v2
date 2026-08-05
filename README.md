# OBS Stream Widget Statistics v2

Локальный сервер + оверлей для OBS: победы / поражения / ранг **без перезагрузки** Browser Source.

## Что внутри

- `widget-stats.exe` — Go-сервер с system tray (Open Admin / Quit)
- `http://127.0.0.1:19123/overlay/` — оверлей для OBS Browser Source
- `http://127.0.0.1:19123/admin/` — админка (статы + внешний вид)
- `obs/widget_control.lua` — автостарт exe, тихие хоткеи (без вспышек консоли), стоп при закрытии OBS
- REST API с **POST и GET** (удобно для Stream Deck / Companion)

## Быстрый старт для стримера

1. Запусти `widget-stats.exe` (или добавь `obs/widget_control.lua` — он только автостартует сервер и хоткеи)
2. В OBS включи **Tools → WebSocket Server Settings** (порт `4455`, пароль по желанию)
3. Открой админку → **Подключить OBS** → выбери **сцену** → **Поставить виджет на сцену**

Виджет создаётся **только на выбранной сцене** и только если его там ещё нет. Повторный клик не плодит дубликаты — максимум обновит URL.

В админке справа есть **живое превью** внешнего вида. Язык интерфейса: **Русский / English** (переключатель в шапке).

### Антивирус / SmartScreen

`widget-stats.exe` **не подписан** кодовым сертификатом (code signing). Windows Defender, SmartScreen и другие антивирусы могут ругаться, писать «неизвестный издатель» или даже временно блокировать файл — это типично для небольших open-source утилит без платного сертификата.

Это **ложное срабатывание**: исходники открыты в репозитории, exe собирается локально из Go. Если Windows спросит — «Подробнее» → «Выполнить в любом случае», либо добавь папку/файл в исключения антивируса. При сомнениях собери exe сам: `.\scripts\build.ps1`.

### Сборка (разработчику)

Нужен [Go](https://go.dev/dl/) 1.22+.

```powershell
cd OBS-Stream-Widget-Statistics-v2
.\scripts\build.ps1
```

Или:

```powershell
go build -o widget-stats.exe ./cmd/widget-stats
```

Рядом с exe появится папка `data/` при первом запуске (`state.json`, `settings.json`, `runtime.json`).

### Ручной запуск без Lua

Дважды кликни `widget-stats.exe` → трей → **Open Admin**.  
В OBS Browser Source укажи URL из админки (кнопка «Скопировать URL оверлея»).

## API

Базовый URL смотри в админке или в `data/runtime.json` (порт может сдвинуться, если 19123 занят).

| Метод | Путь | Действие |
|-------|------|----------|
| GET/POST | `/api/win` | +1 победа |
| GET/POST | `/api/loss` | +1 поражение |
| GET/POST | `/api/rank/up` | ранг вверх |
| GET/POST | `/api/rank/down` | ранг вниз |
| GET/POST | `/api/reset` | сброс W/L |
| GET | `/api/state` | состояние |
| GET/PUT | `/api/settings` | настройки |
| GET | `/api/runtime` | порт и URL |
| GET | `/api/snapshot` | state + settings |
| WS | `/ws` | live push |
| GET | `/health` | healthcheck |

## Stream Deck / Companion

Создай кнопку **Website** / **HTTP request**:

- URL: `http://127.0.0.1:19123/api/win`
- Method: GET или POST

Аналогично `/api/loss`, `/api/rank/up`, `/api/rank/down`, `/api/reset`.

Готовые ссылки также копируются из админки.

## Структура

```
cmd/widget-stats/     entrypoint
internal/             store, server, hub, tray, runtime
web/overlay/          оверлей
web/admin/            админка
obs/widget_control.lua
data/                 runtime state (не коммитится)
```

## Отличия от v1

v1 менял URL Browser Source → страница мигала.  
v2 держит оверлей на `http://127.0.0.1` и пушит обновления по WebSocket.

## Лицензия

MIT

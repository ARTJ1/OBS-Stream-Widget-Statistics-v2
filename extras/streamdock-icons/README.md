# Stream Dock / AJAZZ icons (bonus)

Готовые кнопки **144×144** под все скины виджета (`web/admin/skins.js`).  
Плагин HTTP для AJAZZ универсальный — иконки лежат здесь, у виджета.

## Скачать

Архив с PNG: релиз **[streamdock-icons](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases)**  
(тег `streamdock-icons-v1.0.0` или новее).

Внутри: папка на каждый скин (`default`, `cyber-cyan`, …) + `_sheet_<skin>.png`.

## Кнопки (15 на скин)

| Файл | Действие API |
|------|----------------|
| `win.png` | `/api/win` |
| `loss.png` | `/api/loss` |
| `rank_up.png` / `rank_down.png` | `/api/rank/up` · `/down` |
| `rank_up_tank.png` … | ранг + роль |
| `reset.png` | `/api/reset` |
| `game_next.png` / `mode_next.png` / `role_next.png` | next game/mode/role |
| `show.png` | показать оверлей |

HTTP-плагин для AJAZZ: [streamdock-http-request](https://github.com/ARTJ1/streamdock-http-request).

## Пересборка

```powershell
cd extras\streamdock-icons
npm install
node generate.js
```

Нужен Node.js 18+. Цвета читаются из `web/admin/skins.js`.

# AJAZZ Stream Dock · Live Mode

Цифры побед/поражений и короткий код ранга (`G2`, `GM3`, `#42`) прямо на кнопках деки.

## Что нужно

1. **Widget Stats** ≥ **v2.5.0** (API `/api/deck/state` + `rankShort`)
2. **Плагин HTTP** ≥ **v1.1.0** — [streamdock-http-request releases](https://github.com/ARTJ1/streamdock-http-request/releases)
3. **LIVE-иконки C4** — [streamdock-icons-live](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases/tag/streamdock-icons-live-v1.0.0)

Классический пак иконок (без места под цифры) по-прежнему здесь: [streamdock-icons-v1.0.0](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases/tag/streamdock-icons-v1.0.0).

---

## Установка за 5 шагов

### 1. Обновите виджет
Скачайте `widget-stats.exe` из [релизов виджета](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases) (v2.5.0+), замените старый exe, запустите. Админка: `http://127.0.0.1:19123/admin/`.

### 2. Установите плагин v1.1.0
1. Скачайте `streamdock-http-request.zip` из [релизов плагина](https://github.com/ARTJ1/streamdock-http-request/releases)
2. Распакуйте папку `com.kdfx.streamdock.httprequest.sdPlugin` в:
   `%AppData%\HotSpot\StreamDock\plugins\`
3. **Полностью перезапустите** Stream Dock / AJAZZ

### 3. Поставьте LIVE-иконки
1. Скачайте `streamdock-icons-live-v1.0.0.zip`
2. Откройте папку скина, который используете в виджете (например `default/`)
3. В Stream Dock назначьте PNG на кнопки:

| Файл | Действие |
|------|----------|
| `win.png` | Победа +1 |
| `win_down.png` | Победа −1 |
| `loss.png` | Поражение +1 |
| `loss_down.png` | Поражение −1 |
| `rank_up.png` / `rank_down.png` | Ранг ↑ / ↓ |
| `reset.png` | Сброс W/L |
| `game_next.png` / `mode_next.png` / `role_next.png` | Игра / режим / роль |
| `show.png` | Показать оверлей |
| `rank_up_tank.png` … | Ранг по ролям OW |

Превью всех кнопок скина: `_sheet_<skin>.png` в корне архива.

### 4. Настройте кнопки в плагине
На каждой кнопке HTTP Request:
1. Выберите **Preset** (Widget Win, Loss, Rank Up, …)
2. **Live display** = Auto (или Wins / Losses / Rank)
3. Один раз включите **Live Mode** (глобально) и укажите **Widget URL**  
   обычно `http://127.0.0.1:19123` (порт смотрите в админке, если другой)

### 5. Проверка
- Нажмите Win на деке → в виджете +1 и на кнопке сразу новая цифра
- Поменяйте счёт в админке → на деке обновится само
- Rank ↑/↓ → на кнопке короткий код (`G2`, `I4`, `#42` …)

---

## Коды ранга (Overwatch)

`B` Bronze · `S` Silver · `G` Gold · `P` Platinum · `I` Emerald · `D` Diamond · `M` Master · `GM` Grandmaster · `CH` Champion · Top 500 как `#42`

---

## Если не работает

| Симптом | Что проверить |
|---------|----------------|
| Live: off / нет цифр | Live Mode включён? Виджет запущен? |
| Connecting… | URL/порт: откройте `/api/deck/state` в браузере |
| Ранг не сокращается | Нужен виджет **v2.5.0+** |
| Кнопка без иконки C4 | Поставлен LIVE-пак, не классический |
| Плагин не виден | Перезапуск Stream Dock после копирования в `plugins\` |

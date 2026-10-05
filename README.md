# OBS Stream Widget Statistics v2

Локальное приложение для отображения статистики стрима в OBS Studio: победы, поражения и текущий ранг. Данные обновляются по WebSocket без перезагрузки Browser Source.

[English version](README.en.md)

## Возможности

- Локальный сервер Windows с иконкой в системном трее
- Оверлей для OBS Browser Source
- Веб-админка для настройки внешнего вида, анимаций и скинов
- Опциональный Lua-скрипт: автозапуск сервера, горячие клавиши, остановка при закрытии OBS
- Переключение языка интерфейса админки: русский / English
- Совместимость со Stream Deck, Bitfocus Companion и другими HTTP-клиентами
- Бета: автозачёт побед/поражений Overwatch 2 — встроенная нейросеть узнаёт надпись «ПОБЕДА» / «ПОРАЖЕНИЕ» по кадрам из OBS (даже тусклую и под меню), один матч — один зачёт. Пока обучена на русском клиенте

| Компонент | Путь / адрес |
|-----------|----------------|
| Приложение | `widget-stats.exe` |
| Оверлей | `http://127.0.0.1:19123/overlay/` |
| Админка | `http://127.0.0.1:19123/admin/` |
| Скрипт OBS | `obs/widget_control.lua` |

**Попробовать в браузере** (живой виджет, скины, установка): [artj1.github.io/OBS-Stream-Widget-Statistics-v2/docs/presentation](https://artj1.github.io/OBS-Stream-Widget-Statistics-v2/docs/presentation/)

Актуальные сборки публикуются в разделе [Releases](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases).

## Установка

1. Скачайте `widget-stats.exe` (при необходимости также `widget_control.lua`) из [Releases](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases).
2. Запустите `widget-stats.exe`. Откройте админку через пункт **Open Admin** в меню трея.
3. В OBS Studio включите **Tools → WebSocket Server Settings** (порт по умолчанию `4455`; пароль — по желанию).
4. В админке выполните **Подключить OBS**, выберите сцену и нажмите **Поставить виджет на сцену**.

Источник создаётся только на выбранной сцене. Повторное размещение не создаёт дубликаты.

### Скрипт OBS (рекомендуется)

Добавьте `widget_control.lua` в **Tools → Scripts**. При размещении скрипта рядом с `widget-stats.exe`:

- сервер запускается при старте OBS;
- останавливается при закрытии OBS;
- горячие клавиши выполняются без появления окна командной строки.

## Предупреждения средств защиты

Исполняемый файл **не подписан** сертификатом code signing. Windows SmartScreen, Microsoft Defender и сторонние антивирусы могут отображать предупреждение о неизвестном издателе.

Это ожидаемое поведение для открытого программного обеспечения без коммерческого сертификата и само по себе не означает наличие вредоносного кода. Исходный код доступен в данном репозитории.

Рекомендуемые действия:

1. В диалоге SmartScreen: **Подробнее** → **Выполнить в любом случае**.
2. При необходимости добавить файл или каталог в исключения антивируса.
3. Собрать приложение самостоятельно из исходников (см. раздел «Сборка»).

## Интеграция со Stream Deck и аналогами

Поддерживаются HTTP-запросы (GET или POST) к локальному API:

```
http://127.0.0.1:19123/api/win
http://127.0.0.1:19123/api/loss
http://127.0.0.1:19123/api/rank/up
http://127.0.0.1:19123/api/rank/down
http://127.0.0.1:19123/api/reset
```

Если порт `19123` занят, актуальный адрес отображается в админке и в файле `data/runtime.json`.

Устройства без HTTP могут использовать те же горячие клавиши, что назначены в OBS.

### AJAZZ / Stream Dock

- Плагин тихих HTTP-запросов + **Live Mode**: [streamdock-http-request](https://github.com/ARTJ1/streamdock-http-request/releases)
- LIVE-иконки (цифры и код ранга на кнопках): [streamdock-icons-live-v1.0.0](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases/tag/streamdock-icons-live-v1.0.0)
- Классические иконки: [streamdock-icons-v1.0.0](https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases/tag/streamdock-icons-v1.0.0)
- Пошаговая инструкция: [docs/STREAMDOCK-LIVE.md](docs/STREAMDOCK-LIVE.md)

## Сборка из исходников

Требуется [Go](https://go.dev/dl/) версии 1.22 или новее.

```powershell
.\scripts\build.ps1
```

В результате будет создан `widget-stats.exe`. Каталог `data/` формируется при первом запуске.

## HTTP API

| Метод | Путь | Описание |
|-------|------|----------|
| GET / POST | `/api/win` | Увеличить число побед |
| GET / POST | `/api/loss` | Увеличить число поражений |
| GET / POST | `/api/rank/up` | Повысить ранг |
| GET / POST | `/api/rank/down` | Понизить ранг |
| GET / POST | `/api/reset` | Сбросить W/L |
| GET / POST | `/api/win/down` | Уменьшить число побед |
| GET / POST | `/api/loss/down` | Уменьшить число поражений |
| GET / POST | `/api/automation/toggle` | Вкл/выкл автозачёт OW2 |
| GET | `/api/automation/status` | Состояние автозачёта |
| GET | `/api/state` | Текущее состояние |
| GET / PUT | `/api/settings` | Настройки |
| GET | `/api/runtime` | Базовый URL и порт |
| GET | `/api/snapshot` | Состояние и настройки |
| WS | `/ws` | Обновления в реальном времени |
| GET | `/health` | Проверка работоспособности |

## Отличия от версии 1

В версии 1 изменение статистики выполнялось через смену URL Browser Source, что вызывало мерцание. В версии 2 оверлей постоянно загружен с `localhost`, а обновления передаются по WebSocket.

## Лицензия

Source-Available (не open-source).  
Можно бесплатно пользоваться виджетом и просматривать код для проверки безопасности.  
Нельзя копировать код, делать клоны и продавать/перепродавать без разрешения автора.  
Подробности — в файле `LICENSE`.

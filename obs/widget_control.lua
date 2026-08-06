obs = obslua

local config = {
  exe_path = "",
  base_url = "http://127.0.0.1:19123",
  auto_start = true,
  stop_on_unload = true,
}

local hotkey_win = obs.OBS_INVALID_HOTKEY_ID
local hotkey_loss = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_up = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_down = obs.OBS_INVALID_HOTKEY_ID
local hotkey_reset = obs.OBS_INVALID_HOTKEY_ID
local hotkey_mode = obs.OBS_INVALID_HOTKEY_ID
local hotkey_role = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_up_tank = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_up_support = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_up_damage = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_down_tank = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_down_support = obs.OBS_INVALID_HOTKEY_ID
local hotkey_rank_down_damage = obs.OBS_INVALID_HOTKEY_ID

local hotkey_seq = 0
local started_once = false
local update_announced = nil
local update_latest = nil

local function script_path()
  local info = debug.getinfo(1, "S")
  local src = info and info.source or ""
  src = src:gsub("^@", "")
  return src:gsub("/", "\\")
end

local function script_dir()
  local src = script_path():gsub("\\", "/")
  return src:match("^(.*)/[^/]+$") or "."
end

local function temp_dir()
  return (os.getenv("TEMP") or os.getenv("TMP") or "."):gsub("/", "\\")
end

local function file_exists(path)
  local f = io.open(path, "r")
  if not f then return false end
  f:close()
  return true
end

-- ONLY for start / open-admin. Never call from hotkeys or unload.
local function start_exe_hidden(exe)
  local vbs = temp_dir() .. "\\widget_stats_start.vbs"
  local f = io.open(vbs, "w")
  if not f then return end
  f:write("On Error Resume Next\n")
  f:write('CreateObject("Wscript.Shell").Run """' .. exe:gsub('"', "") .. '""", 0, False\n')
  f:close()
  os.execute('wscript //nologo //B "' .. vbs .. '"')
end

local function open_url_hidden(url)
  local vbs = temp_dir() .. "\\widget_stats_open.vbs"
  local f = io.open(vbs, "w")
  if not f then return end
  f:write("On Error Resume Next\n")
  f:write('CreateObject("Wscript.Shell").Run """' .. url:gsub('"', "") .. '""", 1, False\n')
  f:close()
  os.execute('wscript //nologo //B "' .. vbs .. '"')
end

local function resolve_exe()
  if config.exe_path and config.exe_path ~= "" and file_exists(config.exe_path) then
    return config.exe_path
  end
  local dir = script_dir()
  local candidates = {
    dir .. "/../widget-stats.exe",
    dir .. "/../cmd/widget-stats/widget-stats.exe",
    dir .. "/widget-stats.exe",
  }
  for _, p in ipairs(candidates) do
    local norm = p:gsub("/", "\\")
    if file_exists(norm) or file_exists(p) then
      return norm
    end
  end
  return nil
end

local function data_dir_for_exe(exe)
  if not exe then return nil end
  local dir = exe:gsub("\\", "/"):match("^(.*)/[^/]+$")
  if not dir then return nil end
  return dir:gsub("/", "\\") .. "\\data"
end

local function read_runtime()
  local exe = resolve_exe()
  local data = data_dir_for_exe(exe)
  local result = {
    base = config.base_url,
    admin = config.base_url .. "/admin/",
    pid = nil,
  }
  if not data then return result end
  local path = data .. "\\runtime.json"
  local f = io.open(path, "r")
  if not f then return result end
  local content = f:read("*a")
  f:close()
  local base = content:match('"baseUrl"%s*:%s*"(.-)"')
  local admin = content:match('"adminUrl"%s*:%s*"(.-)"')
  local pid = content:match('"pid"%s*:%s*(%d+)')
  if base and base ~= "" then result.base = base end
  if admin and admin ~= "" then result.admin = admin end
  if pid then result.pid = tonumber(pid) end
  return result
end

-- Fire-and-forget start. No health polling (that froze OBS for tens of seconds).
-- Second instance exits immediately via single-instance lock.
local function ensure_server()
  local exe = resolve_exe()
  if not exe then
    obs.script_log(obs.LOG_WARNING, "widget-stats.exe not found. Set path in script settings.")
    return false
  end
  if not config.auto_start and not started_once then
    -- button "start" forces launch even if auto_start off
  end
  start_exe_hidden(exe)
  started_once = true
  obs.script_log(obs.LOG_INFO, "Started widget-stats: " .. exe)
  local rt = read_runtime()
  if rt.base then config.base_url = rt.base end
  return true
end

local function open_admin()
  ensure_server()
  local rt = read_runtime()
  open_url_hidden(rt.admin or (config.base_url .. "/admin/"))
end

-- Hotkeys / quit: ONLY write a file. Zero os.execute — no console windows.
local function queue_action(action)
  local exe = resolve_exe()
  local data = data_dir_for_exe(exe)
  if not data then
    obs.script_log(obs.LOG_WARNING, "queue: set path to widget-stats.exe")
    return false
  end
  hotkey_seq = hotkey_seq + 1
  local name = string.format("hk_%d_%d_%d.wreq", os.time(), hotkey_seq, math.random(1000, 9999))
  local path = data .. "\\" .. name
  local f = io.open(path, "w")
  if not f then
    obs.script_log(obs.LOG_WARNING, "queue: cannot write " .. path)
    return false
  end
  f:write(action)
  f:close()
  return true
end

local function on_hotkey(action)
  return function(pressed)
    if pressed then queue_action(action) end
  end
end

local function request_quit()
  queue_action("quit")
end

local function write_lua_path()
  local exe = resolve_exe()
  local data = data_dir_for_exe(exe)
  if not data then return end
  local f = io.open(data .. "\\lua_path.txt", "w")
  if not f then return end
  f:write(script_path())
  f:close()
end

local function read_update_json()
  local exe = resolve_exe()
  local data = data_dir_for_exe(exe)
  if not data then return nil end
  local f = io.open(data .. "\\update.json", "r")
  if not f then return nil end
  local content = f:read("*a")
  f:close()
  if not content or content == "" then return nil end
  local available = content:match('"available"%s*:%s*true') ~= nil
  local latest = content:match('"latest"%s*:%s*"(.-)"')
  local current = content:match('"current"%s*:%s*"(.-)"')
  return {
    available = available,
    latest = latest,
    current = current,
  }
end

local function open_update_admin()
  ensure_server()
  local rt = read_runtime()
  local admin = rt.admin or (config.base_url .. "/admin/")
  if not admin:find("%?update=1") then
    if admin:sub(-1) == "/" then
      admin = admin .. "?update=1"
    else
      admin = admin .. "/?update=1"
    end
  end
  open_url_hidden(admin)
end

local function poll_update_status()
  local info = read_update_json()
  if not info or not info.available or not info.latest then
    return
  end
  update_latest = info.latest
  if update_announced == info.latest then
    return
  end
  update_announced = info.latest
  local msg = string.format(
    "Widget Stats: доступно обновление %s (сейчас %s). Tools → Scripts → кнопка «Открыть обновление», либо админка.",
    info.latest,
    info.current or "?"
  )
  obs.script_log(obs.LOG_WARNING, msg)
end

function script_description()
  local extra = ""
  if update_latest then
    extra = "\n\n⚠ Доступно обновление " .. update_latest .. " — открой админку и нажми «Обновить»."
  end
  return [[OBS Stream Widget Statistics v2

Автозапуск без долгого ожидания.
Хоткеи и выключение — через файл data/hk_*.wreq (без окон консоли).
Сцену настраивай в админке.]] .. extra
end

function script_properties()
  local props = obs.obs_properties_create()
  obs.obs_properties_add_bool(props, "auto_start", "Автозапуск widget-stats.exe")
  obs.obs_properties_add_bool(props, "stop_on_unload", "Выключать сервер при закрытии OBS")
  obs.obs_properties_add_path(props, "exe_path", "Путь к widget-stats.exe", obs.OBS_PATH_FILE, "Executable (*.exe)", nil)
  obs.obs_properties_add_text(props, "base_url", "Base URL (fallback)", obs.OBS_TEXT_DEFAULT)
  obs.obs_properties_add_button(props, "open_admin", "Открыть админку", function()
    open_admin()
    return true
  end)
  obs.obs_properties_add_button(props, "open_update", "Открыть обновление", function()
    open_update_admin()
    return true
  end)
  obs.obs_properties_add_button(props, "start_server", "Запустить сервер", function()
    config.auto_start = true
    ensure_server()
    return true
  end)
  obs.obs_properties_add_button(props, "stop_server", "Остановить сервер", function()
    request_quit()
    return true
  end)
  return props
end

function script_defaults(settings)
  obs.obs_data_set_default_bool(settings, "auto_start", true)
  obs.obs_data_set_default_bool(settings, "stop_on_unload", true)
  obs.obs_data_set_default_string(settings, "base_url", "http://127.0.0.1:19123")
end

function script_update(settings)
  config.auto_start = obs.obs_data_get_bool(settings, "auto_start")
  config.stop_on_unload = obs.obs_data_get_bool(settings, "stop_on_unload")
  config.exe_path = obs.obs_data_get_string(settings, "exe_path")
  local base = obs.obs_data_get_string(settings, "base_url")
  if base and base ~= "" then
    config.base_url = base
  end
end

function script_load(settings)
  script_update(settings)
  math.randomseed(os.time())

  hotkey_win = obs.obs_hotkey_register_frontend("v2_widget_win", "V2 Виджет: +1 Победа", on_hotkey("win"))
  hotkey_loss = obs.obs_hotkey_register_frontend("v2_widget_loss", "V2 Виджет: +1 Поражение", on_hotkey("loss"))
  hotkey_rank_up = obs.obs_hotkey_register_frontend("v2_widget_rank_up", "V2 Виджет: Ранг вверх", on_hotkey("rank_up"))
  hotkey_rank_down = obs.obs_hotkey_register_frontend("v2_widget_rank_down", "V2 Виджет: Ранг вниз", on_hotkey("rank_down"))
  hotkey_reset = obs.obs_hotkey_register_frontend("v2_widget_reset", "V2 Виджет: Сброс W/L", on_hotkey("reset"))
  hotkey_mode = obs.obs_hotkey_register_frontend("v2_widget_mode", "V2 Виджет: Следующий режим", on_hotkey("mode_next"))
  hotkey_role = obs.obs_hotkey_register_frontend("v2_widget_role", "V2 Виджет: Следующая роль", on_hotkey("role_next"))
  hotkey_rank_up_tank = obs.obs_hotkey_register_frontend("v2_widget_rank_up_tank", "V2 Виджет: Ранг↑ Tank", on_hotkey("rank_up_tank"))
  hotkey_rank_up_support = obs.obs_hotkey_register_frontend("v2_widget_rank_up_support", "V2 Виджет: Ранг↑ Support", on_hotkey("rank_up_support"))
  hotkey_rank_up_damage = obs.obs_hotkey_register_frontend("v2_widget_rank_up_damage", "V2 Виджет: Ранг↑ Damage", on_hotkey("rank_up_damage"))
  hotkey_rank_down_tank = obs.obs_hotkey_register_frontend("v2_widget_rank_down_tank", "V2 Виджет: Ранг↓ Tank", on_hotkey("rank_down_tank"))
  hotkey_rank_down_support = obs.obs_hotkey_register_frontend("v2_widget_rank_down_support", "V2 Виджет: Ранг↓ Support", on_hotkey("rank_down_support"))
  hotkey_rank_down_damage = obs.obs_hotkey_register_frontend("v2_widget_rank_down_damage", "V2 Виджет: Ранг↓ Damage", on_hotkey("rank_down_damage"))

  local function load_hk(id, key)
    local arr = obs.obs_data_get_array(settings, key)
    if arr then
      obs.obs_hotkey_load(id, arr)
      obs.obs_data_array_release(arr)
    end
  end
  load_hk(hotkey_win, "v2_widget_win")
  load_hk(hotkey_loss, "v2_widget_loss")
  load_hk(hotkey_rank_up, "v2_widget_rank_up")
  load_hk(hotkey_rank_down, "v2_widget_rank_down")
  load_hk(hotkey_reset, "v2_widget_reset")
  load_hk(hotkey_mode, "v2_widget_mode")
  load_hk(hotkey_role, "v2_widget_role")
  load_hk(hotkey_rank_up_tank, "v2_widget_rank_up_tank")
  load_hk(hotkey_rank_up_support, "v2_widget_rank_up_support")
  load_hk(hotkey_rank_up_damage, "v2_widget_rank_up_damage")
  load_hk(hotkey_rank_down_tank, "v2_widget_rank_down_tank")
  load_hk(hotkey_rank_down_support, "v2_widget_rank_down_support")
  load_hk(hotkey_rank_down_damage, "v2_widget_rank_down_damage")

  if config.auto_start then
    ensure_server()
  end
  write_lua_path()
  obs.timer_add(poll_update_status, 10000)
end

function script_save(settings)
  obs.obs_data_set_bool(settings, "auto_start", config.auto_start)
  obs.obs_data_set_bool(settings, "stop_on_unload", config.stop_on_unload)
  obs.obs_data_set_string(settings, "exe_path", config.exe_path or "")
  obs.obs_data_set_string(settings, "base_url", config.base_url or "")

  local function save_hk(id, key)
    local arr = obs.obs_hotkey_save(id)
    if arr then
      obs.obs_data_set_array(settings, key, arr)
      obs.obs_data_array_release(arr)
    end
  end
  save_hk(hotkey_win, "v2_widget_win")
  save_hk(hotkey_loss, "v2_widget_loss")
  save_hk(hotkey_rank_up, "v2_widget_rank_up")
  save_hk(hotkey_rank_down, "v2_widget_rank_down")
  save_hk(hotkey_reset, "v2_widget_reset")
  save_hk(hotkey_mode, "v2_widget_mode")
  save_hk(hotkey_role, "v2_widget_role")
  save_hk(hotkey_rank_up_tank, "v2_widget_rank_up_tank")
  save_hk(hotkey_rank_up_support, "v2_widget_rank_up_support")
  save_hk(hotkey_rank_up_damage, "v2_widget_rank_up_damage")
  save_hk(hotkey_rank_down_tank, "v2_widget_rank_down_tank")
  save_hk(hotkey_rank_down_support, "v2_widget_rank_down_support")
  save_hk(hotkey_rank_down_damage, "v2_widget_rank_down_damage")
end

function script_unload()
  -- Instant: only write quit file. No taskkill / no health wait (that froze OBS).
  if config.stop_on_unload then
    request_quit()
  end
end

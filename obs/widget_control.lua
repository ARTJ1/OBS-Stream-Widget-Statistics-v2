obs = obslua

local config = {
  exe_path = "",
  base_url = "http://127.0.0.1:19123",
  auto_start = true,
  stop_on_unload = true,
}

local hotkey_win = obs.OBS_INVALID_HOTKEY_ID
local hotkey_win_down = obs.OBS_INVALID_HOTKEY_ID
local hotkey_loss = obs.OBS_INVALID_HOTKEY_ID
local hotkey_loss_down = obs.OBS_INVALID_HOTKEY_ID
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

local function script_path()
  local info = debug.getinfo(1, "S")
  local src = info and info.source or ""
  src = src:gsub("^@", "")
  -- Parentheses: gsub also returns the match count, which f:write() would append.
  return (src:gsub("/", "\\"))
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

-- "G:\w\obs\..\widget-stats.exe" -> "G:\w\widget-stats.exe"
local function clean_path(p)
  local out = {}
  for part in p:gmatch("[^\\]+") do
    if part == ".." and #out > 1 and out[#out] ~= ".." then
      table.remove(out)
    elseif part ~= "." then
      table.insert(out, part)
    end
  end
  local s = table.concat(out, "\\")
  if p:sub(1, 2) == "\\\\" then s = "\\\\" .. s end
  return s
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
      return clean_path(norm)
    end
  end
  return nil
end

local settings_ref = nil

-- Fill the path field with the exe found next to the script, so the settings
-- show what is actually used. Also replaces a saved path that no longer exists.
local function autofill_exe_path()
  if not settings_ref then return end
  if config.exe_path ~= "" and file_exists(config.exe_path) then return end
  local exe = resolve_exe()
  if not exe then return end
  config.exe_path = exe
  obs.obs_data_set_string(settings_ref, "exe_path", exe)
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
  local rt = read_runtime()
  if rt.base then config.base_url = rt.base end
  return true
end

local function open_admin()
  ensure_server()
  local rt = read_runtime()
  open_url_hidden(rt.admin or (config.base_url .. "/admin/"))
end

local function has_non_ascii(s)
  return s ~= nil and s:find("[\128-\255]") ~= nil
end

-- Why hotkeys can't reach widget-stats.exe, or nil when they can.
-- Lua opens files with the ANSI codepage while OBS hands paths over as UTF-8,
-- so a folder with non-Latin letters fails even though it exists.
local function diagnose(skip_write_check)
  local exe = resolve_exe()
  if not exe then
    local wanted = config.exe_path ~= "" and config.exe_path or script_dir()
    if has_non_ascii(wanted) then
      return "Путь содержит не латинские буквы — Lua не может его открыть. " ..
        "Перенеси папку с widget-stats.exe в путь только из латиницы (например C:\\WidgetStats): " .. wanted
    end
    if config.exe_path ~= "" then
      return "Файл не найден: " .. config.exe_path .. ". Укажи верный путь к widget-stats.exe."
    end
    return "widget-stats.exe не найден рядом со скриптом (" .. script_dir() ..
      "). Укажи путь к widget-stats.exe ниже — без этого хоткеи не работают."
  end
  if skip_write_check then return nil end
  local data = data_dir_for_exe(exe)
  local probe = data .."\\hk_probe.tmp"
  local f = io.open(probe, "w")
  if not f then
    return "Нет записи в папку " .. data .. ". Запусти сервер один раз (кнопка ниже), " ..
      "проверь права на папку и «Контролируемый доступ к папкам» в Защитнике Windows."
  end
  f:close()
  os.remove(probe)
  return nil
end

local last_queue_warning = 0

local function warn_queue_failed(reason)
  -- One log line per press is fine, but don't flood when a key is held.
  local now = os.time()
  if now == last_queue_warning then return end
  last_queue_warning = now
  obs.script_log(obs.LOG_WARNING, "Хоткей не отправлен: " .. reason)
end

-- Hotkeys / quit: ONLY write a file. Zero os.execute — no console windows.
-- The file is written under a .tmp name and renamed into place, so the server
-- never sees a half-written drop.
local function queue_action(action)
  local exe = resolve_exe()
  local data = data_dir_for_exe(exe)
  if not data then
    warn_queue_failed(diagnose() or "widget-stats.exe не найден")
    return false
  end
  hotkey_seq = hotkey_seq + 1
  -- Zero-padded so name order (what the server sorts by) is press order.
  local name = string.format("hk_%010d_%06d_%04d", os.time(), hotkey_seq, math.random(1000, 9999))
  local tmp = data .. "\\" .. name .. ".tmp"
  local path = data .. "\\" .. name .. ".wreq"
  local f = io.open(tmp, "w")
  if not f then
    warn_queue_failed(diagnose() or ("нет записи в " .. data))
    return false
  end
  f:write(action)
  f:close()
  local ok, err = os.rename(tmp, path)
  if not ok then
    os.remove(tmp)
    warn_queue_failed("не удалось создать " .. path .. ": " .. tostring(err))
    return false
  end
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

function script_description()
  return [[OBS Stream Widget Statistics v2

Автозапуск без долгого ожидания.
Хоткеи и выключение — через файл data/hk_*.wreq (без окон консоли).
Сцену настраивай в админке.]]
end

-- Shows a warning at the top of the script settings only when hotkeys can't
-- reach the exe; when all is well the path field already shows the exe.
-- OBS 28+ renders it as a warning-coloured info line; older OBS as a text field label.
local function update_status(props)
  local p = obs.obs_properties_get(props, "status")
  if not p then return end
  local problem = diagnose()
  obs.obs_property_set_visible(p, problem ~= nil)
  if not problem then return end
  obs.obs_property_set_description(p, "⚠ Хоткеи не работают: " .. problem)
  if obs.obs_property_text_set_info_type and obs.OBS_TEXT_INFO_WARNING then
    obs.obs_property_text_set_info_type(p, obs.OBS_TEXT_INFO_WARNING)
  end
end

function script_properties()
  autofill_exe_path()
  local props = obs.obs_properties_create()
  obs.obs_properties_add_text(props, "status", "", obs.OBS_TEXT_INFO or obs.OBS_TEXT_DEFAULT)
  update_status(props)
  obs.obs_properties_add_bool(props, "auto_start", "Автозапуск widget-stats.exe")
  obs.obs_properties_add_bool(props, "stop_on_unload", "Выключать сервер при закрытии OBS")
  local exe_prop = obs.obs_properties_add_path(props, "exe_path", "Путь к widget-stats.exe", obs.OBS_PATH_FILE, "Executable (*.exe)", nil)
  obs.obs_property_set_modified_callback(exe_prop, function(ps, _, settings)
    config.exe_path = obs.obs_data_get_string(settings, "exe_path")
    update_status(ps)
    return true
  end)
  obs.obs_properties_add_text(props, "base_url", "Base URL (fallback)", obs.OBS_TEXT_DEFAULT)
  obs.obs_properties_add_button(props, "open_admin", "Открыть админку", function()
    open_admin()
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
  settings_ref = settings
  script_update(settings)
  autofill_exe_path()
  math.randomseed(os.time())

  hotkey_win = obs.obs_hotkey_register_frontend("v2_widget_win", "V2 Виджет: +1 Победа", on_hotkey("win"))
  hotkey_win_down = obs.obs_hotkey_register_frontend("v2_widget_win_down", "V2 Виджет: −1 Победа", on_hotkey("win_down"))
  hotkey_loss = obs.obs_hotkey_register_frontend("v2_widget_loss", "V2 Виджет: +1 Поражение", on_hotkey("loss"))
  hotkey_loss_down = obs.obs_hotkey_register_frontend("v2_widget_loss_down", "V2 Виджет: −1 Поражение", on_hotkey("loss_down"))
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
  load_hk(hotkey_win_down, "v2_widget_win_down")
  load_hk(hotkey_loss, "v2_widget_loss")
  load_hk(hotkey_loss_down, "v2_widget_loss_down")
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
  -- No write probe here: the exe was only just launched and may not have
  -- created data/ yet on a first run.
  local problem = diagnose(true)
  if problem then
    obs.script_log(obs.LOG_WARNING, "Хоткеи не работают: " .. problem)
  end
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
  save_hk(hotkey_win_down, "v2_widget_win_down")
  save_hk(hotkey_loss, "v2_widget_loss")
  save_hk(hotkey_loss_down, "v2_widget_loss_down")
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

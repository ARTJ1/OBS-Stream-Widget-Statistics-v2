const statusEl = document.getElementById('status');
const form = document.getElementById('settingsForm');
const motionForm = document.getElementById('motionForm');
const obsForm = document.getElementById('obsForm');
const copyList = document.getElementById('copyList');
const overlayLink = document.getElementById('overlayLink');
const previewFrame = document.getElementById('previewFrame');
const copyOverlayBtn = document.getElementById('copyOverlayBtn');
const obsSceneSelect = document.getElementById('obsSceneSelect');
const obsStatusText = document.getElementById('obsStatusText');
const rankDivision = document.getElementById('rankDivision');
const rankLevel = document.getElementById('rankLevel');
const rankTopInput = document.getElementById('rankTopInput');
const rankLabel = document.getElementById('rankLabel');
const fontSelect = document.getElementById('fontSelect');
const bgDropzone = document.getElementById('bgDropzone');
const bgFileInput = document.getElementById('bgFileInput');
const bgPreviewImg = document.getElementById('bgPreviewImg');
const bgDropHint = document.getElementById('bgDropHint');
const skinsGrid = document.getElementById('skinsGrid');
const customSkinsGrid = document.getElementById('customSkinsGrid');
const customSkinsEmpty = document.getElementById('customSkinsEmpty');
const customSkinName = document.getElementById('customSkinName');

let overlayUrl = `${location.origin}/overlay/`;
let previewReady = false;
let previewTimer = null;
let selectedSkinId = 'default';
let customSkins = [];
let uiLang = 'ru';

const DIVISIONS = ['Bronze', 'Silver', 'Gold', 'Platinum', 'Diamond', 'Master', 'Grandmaster', 'Champion'];
const obsFields = ['obsHost', 'obsPort', 'obsPassword', 'obsSourceName', 'obsScene'];

function t(key, vars) {
  return window.AdminI18n?.t?.(key, vars) ?? key;
}

function applyUiLang(lang, { persist = false } = {}) {
  uiLang = lang === 'en' ? 'en' : 'ru';
  window.AdminI18n?.applyI18n?.(uiLang);
  // Dynamic status texts should not be wiped by static i18n pass
  if (bgDropHint && !form?.bgImage?.value) bgDropHint.textContent = t('look.bgFormats');
  renderCopyLinks({ baseUrl: (overlayUrl || '').replace(/\/overlay\/?$/, '') || location.origin });
  renderSkins();
  renderCustomSkins();
  if (persist) {
    saveAllSettings().catch(() => {});
  }
}

function setStatus(text, ok = false) {
  statusEl.textContent = text;
  statusEl.classList.toggle('ok', ok);
}

function clamp(n, min, max) {
  return Math.max(min, Math.min(max, n));
}

function hexToRgb(hex) {
  const h = String(hex || '').replace('#', '');
  if (h.length !== 6) return { r: 0, g: 0, b: 0 };
  return { r: parseInt(h.slice(0, 2), 16), g: parseInt(h.slice(2, 4), 16), b: parseInt(h.slice(4, 6), 16) };
}

function rgbToHex(r, g, b) {
  const to = (n) => clamp(Math.round(n), 0, 255).toString(16).padStart(2, '0');
  return `#${to(r)}${to(g)}${to(b)}`;
}

function parseCssColor(value) {
  const v = String(value || '').trim();
  const rgba = v.match(/^rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*([0-9.]+))?\s*\)$/i);
  if (rgba) {
    return { hex: rgbToHex(+rgba[1], +rgba[2], +rgba[3]), alpha: rgba[4] !== undefined ? Math.round(parseFloat(rgba[4]) * 100) : 100 };
  }
  if (v.startsWith('#') && v.length === 7) return { hex: v, alpha: 100 };
  return { hex: '#000000', alpha: 70 };
}

function buildBgColor() {
  const { r, g, b } = hexToRgb(form.bgColorPicker.value);
  const a = (Number(form.bgAlpha.value) || 0) / 100;
  return `rgba(${r},${g},${b},${a})`;
}

function populateFonts() {
  fontSelect.innerHTML = '';
  for (const f of window.WIDGET_FONTS || []) {
    const opt = document.createElement('option');
    opt.value = f.id;
    opt.textContent = f.label + (f.google ? ' (Google)' : '');
    opt.style.fontFamily = f.id;
    fontSelect.appendChild(opt);
  }
}

function syncColorSwatches() {
  const bg = buildBgColor();
  form.bgColor.value = bg;
  const bgSwatch = document.getElementById('bgSwatch');
  bgSwatch.style.background = bg;
  bgSwatch.style.color = '#fff';
  document.getElementById('winsSwatch').style.color = form.winsColor.value;
  document.getElementById('lossesSwatch').style.color = form.lossesColor.value;
  document.getElementById('rankSwatch').style.color = form.rankTextColor.value;
  const iconSwatch = document.getElementById('iconSwatch');
  if (iconSwatch && form.iconColor) iconSwatch.style.color = form.iconColor.value;
  if (form.separatorColor) form.separatorColor.value = buildSeparatorColor();
}

function buildSeparatorColor() {
  if (!form.separatorColorPicker) return form.separatorColor?.value || 'rgba(255,255,255,0.55)';
  const { r, g, b } = hexToRgb(form.separatorColorPicker.value);
  const a = clamp(Number(form.separatorAlpha?.value ?? 55) / 100, 0.05, 1);
  return `rgba(${r},${g},${b},${a})`;
}

function fillSeparatorColor(value) {
  if (!form.separatorColorPicker) return;
  const parsed = parseCssColor(value || 'rgba(255,255,255,0.55)');
  form.separatorColorPicker.value = parsed.hex;
  if (form.separatorAlpha) form.separatorAlpha.value = String(parsed.alpha);
  form.separatorColor.value = value || buildSeparatorColor();
}

function rankIndexFromParts(division, level, top) {
  if (division === 'Top') return 40 + (500 - clamp(Number(top) || 500, 1, 500));
  const di = DIVISIONS.indexOf(division);
  const lvl = clamp(Number(level) || 5, 1, 5);
  return di < 0 ? 0 : di * 5 + (5 - lvl);
}

function partsFromRankIndex(index) {
  const i = clamp(Number(index) || 0, 0, 539);
  if (i >= 40) return { division: 'Top', level: 5, top: 500 - (i - 40) };
  return { division: DIVISIONS[Math.floor(i / 5)] || 'Bronze', level: 5 - (i % 5), top: 500 };
}

function formatRankLabel(index) {
  const p = partsFromRankIndex(index);
  return p.division === 'Top' ? `#${p.top}` : `${p.division} ${p.level}`;
}

function updateRankControlsVisibility() {
  const isTop = rankDivision.value === 'Top';
  rankLevel.hidden = isTop;
  rankTopInput.hidden = !isTop;
}

function fillRankControls(index) {
  const p = partsFromRankIndex(index);
  rankDivision.value = p.division;
  rankLevel.value = String(p.level);
  rankTopInput.value = String(p.top);
  rankLabel.textContent = formatRankLabel(index);
  document.getElementById('rankSwatch').textContent = formatRankLabel(index);
  updateRankControlsVisibility();
}

function setBgPreview(url) {
  form.bgImage.value = url || '';
  if (url) {
    bgPreviewImg.src = url;
    bgPreviewImg.hidden = false;
    bgDropHint.textContent = url;
  } else {
    bgPreviewImg.removeAttribute('src');
    bgPreviewImg.hidden = true;
    bgDropHint.textContent = t('look.bgFormats');
  }
}

function fillMotion(settings) {
  motionForm.fillLimit.value = settings.fillLimit ?? 10;
  motionForm.fillDurationMs.value = settings.fillDurationMs ?? 650;
  motionForm.emptyDurationMs.value = settings.emptyDurationMs ?? 1600;
  motionForm.numberAnimMs.value = settings.numberAnimMs ?? 500;
  motionForm.emptyEffect.value = settings.emptyEffect || 'drain';
  if (motionForm.fillStyle) motionForm.fillStyle.value = settings.fillStyle || 'liquid';
  if (motionForm.rankFx) {
    const fx = settings.rankFx || 'classic';
    if (fx && ![...motionForm.rankFx.options].some((o) => o.value === fx)) {
      const opt = document.createElement('option');
      opt.value = fx;
      opt.textContent = fx;
      motionForm.rankFx.appendChild(opt);
    }
    motionForm.rankFx.value = fx;
  }
  motionForm.vesselWave.checked = settings.vesselWave !== false;
  motionForm.idlePulse.checked = settings.idlePulse !== false;
}

function readMotion() {
  return {
    fillLimit: Number(motionForm.fillLimit.value) || 10,
    fillDurationMs: Number(motionForm.fillDurationMs.value) || 650,
    emptyDurationMs: Number(motionForm.emptyDurationMs.value) || 1600,
    numberAnimMs: Number(motionForm.numberAnimMs.value) || 500,
    emptyEffect: motionForm.emptyEffect.value || 'drain',
    fillStyle: motionForm.fillStyle?.value || 'liquid',
    rankFx: motionForm.rankFx?.value || 'classic',
    vesselWave: !!motionForm.vesselWave.checked,
    idlePulse: !!motionForm.idlePulse.checked,
  };
}

function fillAppearance(settings) {
  const parsed = parseCssColor(settings.bgColor || 'rgba(0,0,0,0.7)');
  form.bgColorPicker.value = parsed.hex;
  form.bgAlpha.value = String(parsed.alpha);
  form.bgColor.value = settings.bgColor || buildBgColor();
  if (settings.winsColor?.startsWith('#')) form.winsColor.value = settings.winsColor;
  if (settings.lossesColor?.startsWith('#')) form.lossesColor.value = settings.lossesColor;
  if (settings.rankTextColor?.startsWith('#')) form.rankTextColor.value = settings.rankTextColor;
  if (form.iconColor && settings.iconColor?.startsWith('#')) form.iconColor.value = settings.iconColor;
  fillSeparatorColor(settings.separatorColor || 'rgba(255,255,255,0.55)');
  if (form.appearEffect) form.appearEffect.value = settings.appearEffect || 'slide';
  if (settings.font) {
    if (![...fontSelect.options].some((o) => o.value === settings.font)) {
      const opt = document.createElement('option');
      opt.value = settings.font;
      opt.textContent = settings.font;
      fontSelect.appendChild(opt);
    }
    fontSelect.value = settings.font;
    window.loadGoogleFont?.(settings.font);
  }
  form.fontSize.value = settings.fontSize || 16;
  form.animDirection.value = settings.animDirection || 'left';
  form.animDurationIn.value = settings.animDurationIn ?? 500;
  form.animStayTime.value = settings.animStayTime ?? 10000;
  form.animDurationOut.value = settings.animDurationOut ?? 500;
  form.hiddenTime.value = settings.hiddenTime ?? 8000;
  setBgPreview(settings.bgImage || '');
  selectedSkinId = settings.skinId || selectedSkinId;
  if (settings.uiLang === 'en' || settings.uiLang === 'ru') {
    uiLang = settings.uiLang;
    window.AdminI18n?.applyI18n?.(uiLang);
  }
  syncColorSwatches();
  fillMotion(settings);
  renderSkins();
  pushPreviewSettings();
}

function fillObs(settings) {
  for (const name of obsFields) {
    const el = obsForm.elements.namedItem(name);
    if (!el || settings[name] == null || name === 'obsScene') continue;
    el.value = settings[name];
  }
  if (settings.obsScene) {
    ensureSceneOption(settings.obsScene);
    obsSceneSelect.value = settings.obsScene;
  }
}

function ensureSceneOption(name) {
  if (!name) return;
  if (![...obsSceneSelect.options].some((o) => o.value === name)) {
    const opt = document.createElement('option');
    opt.value = name;
    opt.textContent = name;
    obsSceneSelect.appendChild(opt);
  }
}

function readAppearance() {
  return {
    bgType: form.bgImage.value ? 'image' : 'color',
    bgColor: buildBgColor(),
    bgImage: form.bgImage.value,
    font: form.font.value,
    fontSize: Number(form.fontSize.value) || 16,
    winsColor: form.winsColor.value,
    lossesColor: form.lossesColor.value,
    rankTextColor: form.rankTextColor.value,
    iconColor: form.iconColor?.value || '#ffffff',
    separatorColor: buildSeparatorColor(),
    appearEffect: form.appearEffect?.value || 'slide',
    animDirection: form.animDirection.value,
    animDurationIn: Number(form.animDurationIn.value) || 0,
    animStayTime: Number(form.animStayTime.value) || 0,
    animDurationOut: Number(form.animDurationOut.value) || 0,
    hiddenTime: Number(form.hiddenTime.value) || 0,
    skinId: selectedSkinId,
    uiLang,
    ...readMotion(),
  };
}

function readObs() {
  return {
    obsHost: obsForm.obsHost.value || '127.0.0.1',
    obsPort: Number(obsForm.obsPort.value) || 4455,
    obsPassword: obsForm.obsPassword.value || '',
    obsSourceName: obsForm.obsSourceName.value || 'Widget Stats v2',
    obsScene: obsSceneSelect.value || '',
  };
}

function readAllSettings() {
  return { ...readAppearance(), ...readObs() };
}

function viewFromSnap(snap) {
  if (snap?.view && typeof snap.view.wins === 'number') {
    return {
      wins: snap.view.wins || 0,
      losses: snap.view.losses || 0,
      rank: snap.view.rank || 0,
      mode: snap.view.mode || snap.state?.mode || 'classic',
      role: snap.view.role || snap.state?.role || 'tank',
      roleCycle: snap.state?.roleCycle || ['tank', 'support', 'damage'],
    };
  }
  const st = snap?.state || snap || {};
  const mode = st.mode || 'classic';
  const role = st.role || 'tank';
  const roleCycle = st.roleCycle || ['tank', 'support', 'damage'];
  if (mode === 'roles_shared' || mode === 'roles_rotate') {
    const r = (st.roles && st.roles[role]) || {};
    return { wins: st.wins || 0, losses: st.losses || 0, rank: r.rank || 0, mode, role, roleCycle };
  }
  if (mode === 'roles_split') {
    const r = (st.roles && st.roles[role]) || {};
    return { wins: r.wins || 0, losses: r.losses || 0, rank: r.rank || 0, mode, role, roleCycle };
  }
  return { wins: st.wins || 0, losses: st.losses || 0, rank: st.rank || 0, mode: 'classic', role, roleCycle };
}

function renderRoleCycle(snapOrView) {
  const row = document.getElementById('roleCycleRow');
  const mode = snapOrView?.mode || snapOrView?.view?.mode || snapOrView?.state?.mode || 'classic';
  const cycle = snapOrView?.roleCycle
    || snapOrView?.state?.roleCycle
    || ['tank', 'support', 'damage'];
  if (row) row.hidden = mode !== 'roles_rotate';
  const set = new Set(cycle);
  document.querySelectorAll('input[name="roleCycle"]').forEach((el) => {
    el.checked = set.has(el.value);
  });
}

function renderModeRole(view) {
  const mode = view.mode || 'classic';
  const modeEl = document.getElementById('gameModeSelect');
  const roleEl = document.getElementById('gameRoleSelect');
  const roleField = document.getElementById('gameRoleField');
  if (modeEl && modeEl.value !== mode) modeEl.value = mode;
  if (roleField) roleField.hidden = mode === 'classic';
  if (roleEl) {
    if (roleEl.value !== view.role) roleEl.value = view.role || 'tank';
    roleEl.disabled = mode === 'classic';
  }
  renderRoleCycle(view);
}

function renderState(snapOrState) {
  const view = snapOrState?.view || snapOrState?.state
    ? viewFromSnap(snapOrState)
    : {
        wins: snapOrState?.wins || 0,
        losses: snapOrState?.losses || 0,
        rank: snapOrState?.rank || 0,
        mode: snapOrState?.mode || 'classic',
        role: snapOrState?.role || 'tank',
        roleCycle: snapOrState?.roleCycle || ['tank', 'support', 'damage'],
      };
  document.getElementById('winsValue').textContent = view.wins;
  document.getElementById('lossesValue').textContent = view.losses;
  fillRankControls(view.rank);
  document.getElementById('winsSwatch').textContent = String(view.wins);
  document.getElementById('lossesSwatch').textContent = String(view.losses);
  renderModeRole(view);
}

function pushPreviewSettings() {
  const settings = readAppearance();
  window.loadGoogleFont?.(settings.font);
  document.querySelectorAll('iframe.widget-preview').forEach((frame) => {
    try {
      frame.contentWindow?.postMessage({ type: 'widget-preview-settings', settings }, '*');
    } catch { /* ignore */ }
  });
}

function sendPreviewDemo(action, extra = {}) {
  pushPreviewSettings();
  document.querySelectorAll('iframe.widget-preview').forEach((frame) => {
    try {
      frame.contentWindow?.postMessage({ type: 'widget-preview-demo', action, ...extra }, '*');
    } catch { /* ignore */ }
  });
}

function schedulePreviewPush() {
  clearTimeout(previewTimer);
  previewTimer = setTimeout(() => {
    syncColorSwatches();
    window.loadGoogleFont?.(form.font.value);
    pushPreviewSettings();
  }, 80);
}

function renderCopyLinks(runtime) {
  const base = runtime.baseUrl || location.origin;
  overlayUrl = `${base}/overlay/`;
  const items = [
    ['Overlay', overlayUrl], ['Admin', `${base}/admin/`],
    ['Win +1', `${base}/api/win`], ['Loss +1', `${base}/api/loss`],
    ['Rank up', `${base}/api/rank/up`], ['Rank down', `${base}/api/rank/down`], ['Reset', `${base}/api/reset`],
    ['Mode next', `${base}/api/mode/next`], ['Role next', `${base}/api/role/next`],
    ['Rank↑ Tank', `${base}/api/rank/up?role=tank`],
    ['Rank↑ Support', `${base}/api/rank/up?role=support`],
    ['Rank↑ Damage', `${base}/api/rank/up?role=damage`],
    ['Rank↓ Tank', `${base}/api/rank/down?role=tank`],
    ['Rank↓ Support', `${base}/api/rank/down?role=support`],
    ['Rank↓ Damage', `${base}/api/rank/down?role=damage`],
  ];
  overlayLink.href = overlayUrl;
  copyList.innerHTML = items.map(([label, url]) => `
    <div class="copy-item"><span>${label}</span><code>${url}</code>
    <button type="button" data-copy="${url}">${t('btn.copy')}</button></div>`).join('');
}

const RANK_FX_LABELS = {
  classic: 'Classic',
  blaze: 'Blaze',
  frost: 'Frost',
  neon: 'Neon',
  divine: 'Divine',
  melt: 'Melt',
  arcane: 'Arcane',
  toxic: 'Toxic',
  paper: 'Paper',
  vortex: 'Vortex',
  shatter: 'Shatter',
  smoke: 'Smoke',
  lightning: 'Lightning',
  ripple: 'Ripple',
  hologram: 'Hologram',
  pixel: 'Pixel',
  chrome: 'Chrome',
  phantom: 'Phantom',
  sonic: 'Sonic',
  bloom: 'Bloom',
  ink: 'Ink',
  prism: 'Prism',
};

let skinsFxFilter = 'all';
let skinsSortMode = 'fx';

function listPresetSkins() {
  return Array.isArray(window.WIDGET_SKINS) ? window.WIDGET_SKINS.slice() : [];
}

function skinFxId(skin) {
  return skin?.settings?.rankFx || 'classic';
}

function skinFxLabel(fx) {
  return t(`fx.${fx}`) || RANK_FX_LABELS[fx] || fx;
}

function skinLocalizedDesc(skin) {
  return window.AdminI18n?.skinDesc?.(skin) || skin?.desc || skin?.name || '';
}

function filteredSortedSkins() {
  let list = listPresetSkins();
  if (skinsFxFilter !== 'all') {
    list = list.filter((s) => skinFxId(s) === skinsFxFilter);
  }
  if (skinsSortMode === 'name') {
    list.sort((a, b) => String(a.name).localeCompare(String(b.name), uiLang === 'en' ? 'en' : 'ru'));
  } else {
    list.sort((a, b) => {
      const fa = skinFxId(a);
      const fb = skinFxId(b);
      if (fa !== fb) return fa.localeCompare(fb);
      return String(a.name).localeCompare(String(b.name), uiLang === 'en' ? 'en' : 'ru');
    });
  }
  return list;
}

function ensureSkinsFilterOptions() {
  const select = document.getElementById('skinsFxFilter');
  if (!select) return;
  const current = select.value || 'all';
  const fxs = [...new Set(listPresetSkins().map(skinFxId))].sort();
  select.innerHTML = '';
  const all = document.createElement('option');
  all.value = 'all';
  all.setAttribute('data-i18n', 'skins.allFx');
  all.textContent = t('skins.allFx');
  select.appendChild(all);
  for (const fx of fxs) {
    const opt = document.createElement('option');
    opt.value = fx;
    opt.textContent = skinFxLabel(fx);
    select.appendChild(opt);
  }
  select.value = current;
  select.dataset.ready = '1';
}

function renderSkinsFxChips() {
  const host = document.getElementById('skinsFxChips');
  if (!host) return;
  const counts = {};
  for (const skin of listPresetSkins()) {
    const fx = skinFxId(skin);
    counts[fx] = (counts[fx] || 0) + 1;
  }
  const fxs = Object.keys(counts).sort();
  const chips = [
    `<button type="button" class="fx-chip ${skinsFxFilter === 'all' ? 'active' : ''}" data-fx-chip="all">${t('skins.all')} <em>${listPresetSkins().length}</em></button>`,
    ...fxs.map((fx) => `
      <button type="button" class="fx-chip ${skinsFxFilter === fx ? 'active' : ''}" data-fx-chip="${fx}">
        ${skinFxLabel(fx)} <em>${counts[fx]}</em>
      </button>`),
  ];
  host.innerHTML = chips.join('');
}

function renderSkins() {
  ensureSkinsFilterOptions();
  renderSkinsFxChips();
  const select = document.getElementById('skinsFxFilter');
  const sortEl = document.getElementById('skinsSort');
  if (select) select.value = skinsFxFilter;
  if (sortEl) sortEl.value = skinsSortMode;

  const list = filteredSortedSkins();
  let lastFx = null;
  skinsGrid.innerHTML = list.map((skin) => {
    const fx = skinFxId(skin);
    const header = (skinsSortMode === 'fx' && skinsFxFilter === 'all' && fx !== lastFx)
      ? `<div class="skins-group-title" data-fx-group="${fx}">${skinFxLabel(fx)}</div>`
      : '';
    lastFx = fx;
    return `${header}
    <button type="button" class="skin-card ${skin.id === selectedSkinId ? 'active' : ''}" data-skin="${skin.id}" data-rank-fx="${fx}">
      <div class="skin-swatch" style="background:${skin.preview.bg}">
        <span class="skin-chip" style="background:${skin.preview.wins}"></span>
        <span class="skin-chip" style="background:${skin.preview.losses}"></span>
        <span class="skin-fx-badge">${skinFxLabel(fx)}</span>
      </div>
      <div class="skin-meta"><strong>${skin.name}</strong><span>${skinLocalizedDesc(skin)}</span></div>
    </button>`;
  }).join('') || `<p class="setup-note">${t('skins.emptyFilter')}</p>`;
  renderCustomSkins();
}

function skinPreviewFromSettings(settings) {
  return {
    bg: settings.bgColor || 'rgba(0,0,0,0.7)',
    wins: settings.winsColor || '#00ff00',
    losses: settings.lossesColor || '#ff0000',
  };
}

function renderCustomSkins() {
  if (!customSkinsGrid) return;
  const has = customSkins.length > 0;
  customSkinsEmpty?.classList.toggle('hidden', has);
  customSkinsGrid.innerHTML = customSkins.map((skin) => {
    const preview = skinPreviewFromSettings(skin.settings || {});
    const active = skin.id === selectedSkinId ? 'active' : '';
    return `
      <div class="skin-card custom ${active}" data-custom-skin="${skin.id}">
        <button type="button" class="skin-pick" data-custom-pick="${skin.id}">
          <div class="skin-swatch" style="background:${preview.bg}">
            <span class="skin-chip" style="background:${preview.wins}"></span>
            <span class="skin-chip" style="background:${preview.losses}"></span>
          </div>
          <div class="skin-meta"><strong>${escapeHtml(skin.name)}</strong><span>${t('skins.customBadge')}</span></div>
        </button>
        <div class="custom-skin-actions">
          <button type="button" data-custom-apply="${skin.id}">${t('skins.apply')}</button>
          <button type="button" class="danger" data-custom-delete="${skin.id}">${t('skins.delete')}</button>
        </div>
      </div>
    `;
  }).join('');
}

function escapeHtml(value) {
  return String(value || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

async function loadCustomSkins() {
  try {
    const data = await api('/api/skins');
    customSkins = Array.isArray(data.skins) ? data.skins : [];
    renderCustomSkins();
  } catch (err) {
    customSkins = [];
    renderCustomSkins();
  }
}

function applySkinSettings(settings, skinId) {
  selectedSkinId = skinId || settings.skinId || selectedSkinId;
  fillAppearance({ ...readAllSettings(), ...settings, skinId: selectedSkinId });
  schedulePreviewPush();
  setTimeout(() => {
    sendPreviewDemo('fullCycle');
    const hint = document.getElementById('skinDemoHint');
    if (hint) hint.textContent = 'Demo…';
  }, 160);
}

function replaySkinDemo() {
  schedulePreviewPush();
  setTimeout(() => sendPreviewDemo('fullCycle'), 100);
  setStatus(t('msg.demoSkin'), true);
}

async function api(path, options) {
  const res = await fetch(path, options);
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = { raw: text }; }
  if (!res.ok) throw new Error((data && (data.error || data.message)) || text || res.statusText);
  return data;
}

async function saveAllSettings() {
  return api('/api/settings', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(readAllSettings()),
  });
}

async function uploadBgFile(file) {
  const body = new FormData();
  body.append('file', file);
  const data = await api('/api/upload/bg', { method: 'POST', body });
  fillAppearance({ ...readAllSettings(), ...data.settings });
  setStatus(t('msg.bgUp'), true);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
    setStatus(t('msg.copied'), true);
  } catch {
    setStatus(t('msg.copyFail'));
  }
}

function setObsStatus(text, ok = false) {
  obsStatusText.textContent = text;
  obsStatusText.style.color = ok ? '#7ddea0' : '';
}

async function refreshScenes() {
  const data = await api('/api/obs/scenes');
  const selected = obsSceneSelect.value || data.selected || '';
  obsSceneSelect.innerHTML = '';
  const placeholder = document.createElement('option');
  placeholder.value = '';
  placeholder.textContent = data.scenes?.length ? t('main.obsPickScene') : t('main.obsNoScenes');
  obsSceneSelect.appendChild(placeholder);
  for (const name of (data.scenes || [])) {
    const opt = document.createElement('option');
    opt.value = name;
    opt.textContent = name === data.currentProgram ? `${name} ${t('main.obsLive')}` : name;
    obsSceneSelect.appendChild(opt);
  }
  const prefer = selected || data.selected || data.currentProgram || '';
  if (prefer) {
    ensureSceneOption(prefer);
    obsSceneSelect.value = prefer;
  }
  setObsStatus(t('main.obsConnected', { n: (data.scenes || []).length }), true);
}

document.getElementById('tabs').addEventListener('click', (e) => {
  const btn = e.target.closest('.tab');
  if (!btn) return;
  document.querySelectorAll('.tab').forEach((t) => t.classList.toggle('active', t === btn));
  document.querySelectorAll('.tab-panel').forEach((p) => p.classList.toggle('active', p.dataset.panel === btn.dataset.tab));
});

document.querySelectorAll('[data-action]').forEach((btn) => {
  btn.addEventListener('click', async () => {
    try {
      const snap = await api(btn.dataset.action, { method: 'POST' });
      renderState(snap);
      setStatus(t('msg.updated'), true);
    } catch (err) {
      setStatus(String(err.message || err));
    }
  });
});

rankDivision.addEventListener('change', updateRankControlsVisibility);
document.getElementById('rankApplyBtn').addEventListener('click', async () => {
  try {
    const rank = rankIndexFromParts(rankDivision.value, rankLevel.value, rankTopInput.value);
    const snap = await api('/api/rank/set', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rank }),
    });
    renderState(snap);
    setStatus(t('msg.rankUpdated'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});

copyList.addEventListener('click', async (e) => {
  const btn = e.target.closest('[data-copy]');
  if (!btn) return;
  await copyText(btn.dataset.copy);
});
copyOverlayBtn.addEventListener('click', () => copyText(overlayUrl));

form.addEventListener('input', schedulePreviewPush);
form.addEventListener('change', schedulePreviewPush);
motionForm.addEventListener('input', schedulePreviewPush);
motionForm.addEventListener('change', schedulePreviewPush);

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  try {
    syncColorSwatches();
    const snap = await saveAllSettings();
    fillAppearance(snap.settings);
    fillObs(snap.settings);
    renderState(snap);
    setStatus(t('msg.saved'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});

document.getElementById('motionSaveBtn').addEventListener('click', async () => {
  try {
    const snap = await saveAllSettings();
    fillAppearance(snap.settings);
    setStatus(t('msg.animSaved'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});
document.getElementById('motionTestWin').addEventListener('click', () => {
  sendPreviewDemo('win');
  setStatus(t('msg.previewWin'), true);
});
document.getElementById('motionTestLoss').addEventListener('click', () => {
  sendPreviewDemo('loss');
  setStatus(t('msg.previewLoss'), true);
});
document.getElementById('motionTestRankUp')?.addEventListener('click', () => {
  sendPreviewDemo('rankShowcase');
  setStatus(t('msg.previewRank'), true);
});
document.getElementById('motionTestRankDown')?.addEventListener('click', () => {
  sendPreviewDemo('rankShowcase');
  setStatus(t('msg.previewRank'), true);
});
document.getElementById('motionTestFillCycle')?.addEventListener('click', () => {
  sendPreviewDemo('fillCycle');
  setStatus(t('msg.previewFill'), true);
});

motionForm.rankFx?.addEventListener('change', () => {
  schedulePreviewPush();
  setTimeout(() => sendPreviewDemo('rankShowcase'), 160);
});

document.querySelectorAll('[data-show-overlay]').forEach((btn) => {
  btn.addEventListener('click', async () => {
    try {
      // Push current unsaved look to overlay first when possible.
      try { await saveAllSettings(); } catch { /* preview still works via show */ }
      await api('/api/show', { method: 'POST' });
      setStatus(t('msg.widgetShown'), true);
    } catch (err) {
      setStatus(String(err.message || err));
    }
  });
});

bgDropzone.addEventListener('click', (e) => {
  if (e.target.closest('#bgClearBtn')) return;
  bgFileInput.click();
});
bgDropzone.addEventListener('dragover', (e) => {
  e.preventDefault();
  bgDropzone.classList.add('dragover');
});
bgDropzone.addEventListener('dragleave', () => bgDropzone.classList.remove('dragover'));
bgDropzone.addEventListener('drop', async (e) => {
  e.preventDefault();
  bgDropzone.classList.remove('dragover');
  const file = e.dataTransfer.files?.[0];
  if (!file) return;
  try { await uploadBgFile(file); } catch (err) { setStatus(String(err.message || err)); }
});
bgFileInput.addEventListener('change', async () => {
  const file = bgFileInput.files?.[0];
  if (!file) return;
  try { await uploadBgFile(file); } catch (err) { setStatus(String(err.message || err)); }
  bgFileInput.value = '';
});
document.getElementById('bgClearBtn').addEventListener('click', async (e) => {
  e.stopPropagation();
  setBgPreview('');
  try {
    const snap = await saveAllSettings();
    fillAppearance(snap.settings);
    setStatus(t('msg.bgClear'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});

skinsGrid.addEventListener('click', (e) => {
  const card = e.target.closest('[data-skin]');
  if (!card) return;
  const skin = (window.WIDGET_SKINS || []).find((s) => s.id === card.dataset.skin);
  if (!skin) return;
  applySkinSettings(skin.settings, skin.id);
  renderSkins();
  setStatus(t('msg.skinPicked', { name: skin.name, fx: skinFxLabel(skinFxId(skin)) }), true);
});

document.getElementById('skinsFxFilter')?.addEventListener('change', (e) => {
  skinsFxFilter = e.target.value || 'all';
  renderSkins();
});
document.getElementById('skinsSort')?.addEventListener('change', (e) => {
  skinsSortMode = e.target.value || 'fx';
  renderSkins();
});
document.getElementById('skinsFxChips')?.addEventListener('click', (e) => {
  const btn = e.target.closest('[data-fx-chip]');
  if (!btn) return;
  skinsFxFilter = btn.dataset.fxChip || 'all';
  renderSkins();
});

customSkinsGrid?.addEventListener('click', async (e) => {
  const del = e.target.closest('[data-custom-delete]');
  if (del) {
    const id = del.dataset.customDelete;
    try {
      await api(`/api/skins?id=${encodeURIComponent(id)}`, { method: 'DELETE' });
      if (selectedSkinId === id) selectedSkinId = 'default';
      await loadCustomSkins();
      renderSkins();
      setStatus(t('msg.skinDeleted'), true);
    } catch (err) {
      setStatus(String(err.message || err));
    }
    return;
  }
  const pick = e.target.closest('[data-custom-pick], [data-custom-apply]');
  if (!pick) return;
  const id = pick.dataset.customPick || pick.dataset.customApply;
  const skin = customSkins.find((s) => s.id === id);
  if (!skin) return;
  applySkinSettings(skin.settings, skin.id);
  renderSkins();
  setStatus(t('msg.customPicked', { name: skin.name }), true);
});

document.getElementById('customSkinSaveBtn')?.addEventListener('click', async () => {
  const name = (customSkinName?.value || '').trim() || t('msg.defaultSkinName', { date: new Date().toLocaleString(uiLang === 'en' ? 'en-US' : 'ru-RU') });
  try {
    const skin = await api('/api/skins', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, settings: readAllSettings() }),
    });
    selectedSkinId = skin.id;
    if (customSkinName) customSkinName.value = '';
    await loadCustomSkins();
    renderSkins();
    setStatus(t('msg.customSaved', { name: skin.name }), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});

document.getElementById('skinApplySaveBtn').addEventListener('click', async () => {
  try {
    const snap = await saveAllSettings();
    fillAppearance(snap.settings);
    setStatus(t('msg.skinSavedObs'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});

document.getElementById('skinReplayDemoBtn')?.addEventListener('click', () => {
  replaySkinDemo();
});

document.getElementById('obsConnectBtn').addEventListener('click', async () => {
  try {
    await saveAllSettings();
    await api('/api/obs/connect', { method: 'POST' });
    await refreshScenes();
    setStatus(t('msg.obsConnected'), true);
  } catch (err) {
    setObsStatus(`OBS: ${err.message}`);
    setStatus(String(err.message || err));
  }
});
document.getElementById('obsRefreshScenesBtn').addEventListener('click', async () => {
  try {
    await saveAllSettings();
    await refreshScenes();
  } catch (err) {
    setObsStatus(`OBS: ${err.message}`);
    setStatus(String(err.message || err));
  }
});
document.getElementById('obsEnsureBtn').addEventListener('click', async () => {
  try {
    if (!obsSceneSelect.value) throw new Error(t('msg.pickScene'));
    await saveAllSettings();
    const res = await api('/api/obs/ensure', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scene: obsSceneSelect.value }),
    });
    setObsStatus(res.result?.message || t('msg.done'), true);
    setStatus(res.result?.action || 'ok', true);
  } catch (err) {
    setObsStatus(`OBS: ${err.message}`);
    setStatus(String(err.message || err));
  }
});

previewFrame.addEventListener('load', () => {
  previewReady = true;
  pushPreviewSettings();
});
document.getElementById('previewFrameMotion')?.addEventListener('load', () => {
  pushPreviewSettings();
});

function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const ws = new WebSocket(`${proto}://${location.host}/ws`);
  ws.addEventListener('open', () => setStatus(t('status.live'), true));
  ws.addEventListener('message', (event) => {
    const msg = JSON.parse(event.data);
    if (msg.state || msg.view) renderState(msg);
  });
  ws.addEventListener('close', () => {
    setStatus(t('status.reconnect'));
    setTimeout(connectWS, 1500);
  });
}

async function boot() {
  populateFonts();
  const [snap, runtime] = await Promise.all([api('/api/snapshot'), api('/api/runtime')]);
  renderState(snap);
  fillAppearance(snap.settings);
  fillObs(snap.settings);
  applyUiLang(snap.settings?.uiLang || uiLang);
  renderCopyLinks(runtime);
  await loadCustomSkins();
  connectWS();
  checkForUpdates({ forceBanner: new URLSearchParams(location.search).has('update') });
  try {
    await api('/api/obs/connect', { method: 'POST' });
    await refreshScenes();
  } catch {
    setObsStatus(t('main.obsOfflineHint'));
  }
}

let latestUpdateInfo = null;
async function checkForUpdates({ forceBanner = false } = {}) {
  try {
    const info = await api('/api/update/check');
    const verEl = document.getElementById('appVersion');
    if (verEl && info.current) verEl.textContent = info.current;
    latestUpdateInfo = info;
    const banner = document.getElementById('updateBanner');
    if (!banner) return;
    if (info.available && info.latest) {
      banner.hidden = false;
      const text = document.getElementById('updateBannerText');
      if (text) {
        text.textContent = t('update.text', { current: info.current || '?', latest: info.latest });
      }
      if (forceBanner) banner.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    } else {
      banner.hidden = true;
      if (forceBanner && !info.skipped) setStatus(t('update.none'), true);
    }
  } catch {
    /* offline / no github */
  }
}

document.getElementById('updateLaterBtn')?.addEventListener('click', () => {
  const banner = document.getElementById('updateBanner');
  if (banner) banner.hidden = true;
});
document.getElementById('updateApplyBtn')?.addEventListener('click', async () => {
  const btn = document.getElementById('updateApplyBtn');
  try {
    if (btn) btn.disabled = true;
    setStatus(t('update.applying'), true);
    await api('/api/update/apply', { method: 'POST' });
    setStatus(t('update.done'), true);
  } catch (err) {
    setStatus(String(err.message || err));
    if (btn) btn.disabled = false;
  }
});

document.getElementById('gameModeSelect')?.addEventListener('change', async (e) => {
  try {
    const snap = await api(`/api/mode?set=${encodeURIComponent(e.target.value)}`, { method: 'POST' });
    renderState(snap);
    setStatus(t('msg.updated'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});
document.getElementById('gameRoleSelect')?.addEventListener('change', async (e) => {
  try {
    const snap = await api(`/api/role?set=${encodeURIComponent(e.target.value)}`, { method: 'POST' });
    renderState(snap);
    setStatus(t('msg.updated'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
});

async function saveRoleCycleFromUI() {
  const roles = [...document.querySelectorAll('input[name="roleCycle"]:checked')].map((el) => el.value);
  if (!roles.length) {
    setStatus(t('main.roleCycle') + ': 1+');
    document.querySelector('input[name="roleCycle"][value="tank"]').checked = true;
    return;
  }
  try {
    const snap = await api('/api/role-cycle', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ roleCycle: roles }),
    });
    renderState(snap);
    setStatus(t('msg.updated'), true);
  } catch (err) {
    setStatus(String(err.message || err));
  }
}
document.getElementById('roleCycleRow')?.addEventListener('change', (e) => {
  if (e.target?.name === 'roleCycle') saveRoleCycleFromUI();
});

document.getElementById('uiLangSelect')?.addEventListener('change', (e) => {
  applyUiLang(e.target.value, { persist: true });
});

boot().catch((err) => setStatus(String(err.message || err)));

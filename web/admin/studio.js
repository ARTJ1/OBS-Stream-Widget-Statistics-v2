// Studio shell on top of admin.js (loaded after it; uses its globals):
// layout picture grid, auto-save with undo, header status pills.
(() => {
  const layoutGrid = document.getElementById('layoutGrid');
  const layoutSelect = form.layout;
  const saveState = document.getElementById('saveState');
  const toast = document.getElementById('saveToast');
  const toastText = document.getElementById('saveToastText');
  const toastUndo = document.getElementById('saveToastUndo');

  function goto(tab) {
    document.querySelector(`.tab[data-tab="${tab}"]`)?.click();
  }

  // ---------- Layout picker: a live mini widget per layout ----------
  let thumbsBuilt = false;
  let thumbTimer = 0;

  function thumbFrames() {
    return [...layoutGrid.querySelectorAll('iframe')];
  }

  function pushThumbs() {
    if (!thumbsBuilt) return;
    const base = { ...readAppearance(), alwaysVisible: true };
    const source = lastSnap;
    const view = source ? viewFromSnap(source) : null;
    thumbFrames().forEach((frame) => {
      const win = frame.contentWindow;
      if (!win) return;
      try {
        if (view) win.postMessage({ type: 'widget-preview-state', view, state: source.state || view }, '*');
        win.postMessage({ type: 'widget-preview-settings', settings: { ...base, layout: frame.dataset.layout } }, '*');
      } catch { /* frame not ready */ }
    });
  }

  function schedulePushThumbs() {
    clearTimeout(thumbTimer);
    thumbTimer = setTimeout(pushThumbs, 250);
  }

  // Which widget parts each layout actually shows (measured from the overlay).
  // Settings tagged data-uses="..." hide when the current layout has none of them.
  const LAYOUT_PARTS = {
    detailed: ['vessels', 'sep', 'bgimage', 'bgcolor'],
    mini: ['bgimage', 'bgcolor'],
    card: ['vessels', 'bgimage', 'bgcolor'],
    badge: ['sep', 'bgimage', 'bgcolor'],
    strip: ['bgimage', 'bgcolor'],
    minimal: ['vessels', 'sep'],
    column: ['vessels', 'bgimage', 'bgcolor'],
    ring: ['bgimage', 'bgcolor'],
    slant: ['vessels', 'bgcolor'],
    scoreboard: ['bgimage', 'bgcolor'],
    session: ['bgimage', 'bgcolor'],
    form: ['bgimage', 'bgcolor'],
    ladder: ['bgimage', 'bgcolor'],
    roles: ['sep', 'bgimage', 'bgcolor'],
    hex: ['sep', 'bgcolor'],
    feed: ['bgcolor'],
    tug: ['bgimage', 'bgcolor'],
    flip: ['bgimage', 'bgcolor'],
    tcard: ['bgcolor'],
  };
  const vesselCredit = document.getElementById('vesselCredit');

  function applyLayoutParts() {
    const parts = new Set(LAYOUT_PARTS[layoutSelect.value] || LAYOUT_PARTS.detailed);
    // The role icon also uses the icon colour.
    const mode = lastSnap ? viewFromSnap(lastSnap).mode : 'classic';
    if (mode && mode !== 'classic') parts.add('roles');
    document.querySelectorAll('[data-uses]').forEach((el) => {
      el.hidden = !el.dataset.uses.split(' ').some((p) => parts.has(p));
    });
    // The icon artist credit belongs to the trophy/skull icons.
    if (!parts.has('vessels')) {
      if (vesselCredit) vesselCredit.hidden = true;
    } else {
      updateVesselCredit();
    }
    const opt = [...layoutSelect.options].find((o) => o.value === layoutSelect.value);
    document.querySelectorAll('.layout-hidden-note').forEach((note) => {
      const panel = note.closest('.panel');
      const anyHidden = !!panel?.querySelector('[data-uses][hidden]');
      note.hidden = !anyHidden;
      if (anyHidden) note.textContent = t('studio.hiddenForLayout', { name: opt ? opt.textContent : '' });
    });
  }

  const cardOptions = document.getElementById('cardOptions');
  function markActiveLayout() {
    if (cardOptions) cardOptions.hidden = layoutSelect.value !== 'tcard';
    applyLayoutParts();
    layoutGrid.querySelectorAll('.layout-tile').forEach((tile) => {
      tile.classList.toggle('active', tile.dataset.layout === layoutSelect.value);
      tile.setAttribute('aria-pressed', String(tile.dataset.layout === layoutSelect.value));
    });
  }

  function labelLayouts() {
    layoutGrid.querySelectorAll('.layout-tile').forEach((tile) => {
      const opt = [...layoutSelect.options].find((o) => o.value === tile.dataset.layout);
      tile.querySelector('.layout-name').textContent = opt ? opt.textContent : tile.dataset.layout;
    });
  }

  function buildThumbs() {
    if (thumbsBuilt) return;
    thumbsBuilt = true;
    layoutGrid.innerHTML = [...layoutSelect.options].map((opt) => `
      <button type="button" class="layout-tile" data-layout="${opt.value}" aria-pressed="false">
        <span class="layout-thumb"><iframe data-layout="${opt.value}" src="/overlay/?preview=1&nows=1&still=1&layout=${opt.value}" tabindex="-1" aria-hidden="true"></iframe></span>
        <span class="layout-name"></span>
      </button>`).join('');
    thumbFrames().forEach((frame) => frame.addEventListener('load', () => setTimeout(schedulePushThumbs, 150)));
    labelLayouts();
    markActiveLayout();
  }

  // Load the thumbnails in the background right after the admin opens, so the page
  // is ready (no loading, no slide-in) the first time it is shown.
  setTimeout(buildThumbs, 1500);

  layoutGrid.addEventListener('click', (e) => {
    const tile = e.target.closest('.layout-tile');
    if (!tile) return;
    layoutSelect.value = tile.dataset.layout;
    layoutSelect.dispatchEvent(new Event('change', { bubbles: true }));
    markActiveLayout();
    setTimeout(() => sendPreviewDemo('fullCycle'), 150);
  });

  // Keep thumbnails and the picked tile in step with every preview update.
  const basePushSettings = pushPreviewSettings;
  pushPreviewSettings = function pushPreviewSettingsStudio(...args) {
    basePushSettings.apply(this, args);
    markActiveLayout();
    schedulePushThumbs();
  };
  const basePushState = pushPreviewState;
  pushPreviewState = function pushPreviewStateStudio(...args) {
    basePushState.apply(this, args);
    applyLayoutParts(); // mode switches change whether the icon colour is used
    schedulePushThumbs();
  };

  document.getElementById('tabs').addEventListener('click', (e) => {
    const btn = e.target.closest('.tab');
    if (!btn) return;
    if (btn.dataset.tab === 'layout') buildThumbs();
    document.querySelector('.studio-main')?.scrollTo?.({ top: 0 });
    window.scrollTo({ top: 0 });
  });
  window.addEventListener('admin-i18n', () => {
    labelLayouts();
    applyLayoutParts();
    if (saveState.dataset.state) saveState.textContent = t(saveState.dataset.state);
  });

  // ---------- Auto-save with undo ----------
  // Every look/motion/skin/layout change saves itself ~0.7s after the last edit.
  // "Undo" restores what was saved before that burst of edits.
  let baseline = null;
  let pendingUndo = null;
  let saveTimer = 0;
  let toastTimer = 0;

  function showSaveState(key) {
    saveState.textContent = key ? t(key) : '';
    saveState.dataset.state = key || '';
  }

  function showToast(key, canUndo) {
    toastText.textContent = t(key);
    toastUndo.hidden = !canUndo;
    toast.hidden = false;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { toast.hidden = true; }, 6000);
  }

  async function saveNow() {
    const undoTo = pendingUndo;
    pendingUndo = null;
    try {
      const snap = await saveAllSettings();
      baseline = readAllSettings();
      if (snap?.settings) fillObs(snap.settings);
      showSaveState('studio.saved');
      if (undoTo) {
        toastUndo.onclick = () => undoTo && undo(undoTo);
        showToast('studio.savedToast', true);
      }
    } catch (err) {
      showSaveState('studio.saveFailed');
      setStatus(String(err.message || err));
    }
  }

  function scheduleSave() {
    if (!baseline) return; // still loading
    if (!pendingUndo) pendingUndo = baseline;
    showSaveState('studio.saving');
    clearTimeout(saveTimer);
    saveTimer = setTimeout(saveNow, 700);
  }

  async function undo(settings) {
    toast.hidden = true;
    clearTimeout(saveTimer);
    pendingUndo = null;
    fillAppearance({ ...readAllSettings(), ...settings });
    try {
      await saveAllSettings();
      baseline = readAllSettings();
      showSaveState('studio.saved');
      showToast('studio.undone', false);
    } catch (err) {
      setStatus(String(err.message || err));
    }
  }

  cardOptions?.addEventListener('change', () => {
    pushPreviewSettings();
    scheduleSave();
    setTimeout(() => sendPreviewDemo('win'), 120);
  });
  form.addEventListener('input', scheduleSave);
  form.addEventListener('change', scheduleSave);
  motionForm.addEventListener('input', scheduleSave);
  motionForm.addEventListener('change', scheduleSave);
  // Skin picks anywhere on the Skins page (stars/rename/delete stop before this).
  document.getElementById('skinsPanel')?.addEventListener('click', (e) => {
    if (e.target.closest('[data-skin], [data-custom-pick]')) scheduleSave();
  });

  // Baseline = what the server had once admin.js finished loading it.
  const waitBoot = setInterval(() => {
    if (!lastSnap) return;
    clearInterval(waitBoot);
    setTimeout(() => {
      baseline = readAllSettings();
      showSaveState('studio.saved');
    }, 300);
  }, 200);

  // "Copy all" lives inside a <summary>; don't fold the group when it is clicked.
  copyList.addEventListener('click', (e) => {
    if (e.target.closest('.copy-all')) e.preventDefault();
  });

  // ---------- Header status pills ----------
  const pillObs = document.getElementById('pillObs');
  const pillAuto = document.getElementById('pillAuto');
  const pillUpdate = document.getElementById('pillUpdate');
  document.getElementById('statusPills').addEventListener('click', (e) => {
    const pill = e.target.closest('[data-goto]');
    if (pill) goto(pill.dataset.goto);
  });
  function refreshPills() {
    const obsOk = !!(obsStatusI18n && obsStatusI18n.ok) || /^rgb\(125, 222, 160\)$/.test(getComputedStyle(obsStatusText).color);
    pillObs.classList.toggle('on', obsOk);
    pillAuto.classList.toggle('on', !!document.getElementById('autoToggle')?.checked);
    pillUpdate.hidden = document.getElementById('updateBanner')?.hidden !== false;
  }
  setInterval(refreshPills, 1500);
  refreshPills();
})();

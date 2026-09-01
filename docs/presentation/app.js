(() => {
  const I18N = {
    ru: {
      "bar.sub": "Статистика для OBS Studio",
      "bar.download": "Скачать",
      "nav.features": "Что умеет",
      "nav.install": "Установка",
      "hero.kicker": "Для OBS Studio",
      "hero.title": "Widget Stats",
      "hero.lead": "Победы, поражения и ранг на стриме — красиво и удобно. Посмотрите, как это будет выглядеть у вас.",
      "hero.rankUp": "Ранг +",
      "hero.rankDown": "Ранг −",
      "hero.reset": "Сброс",
      "hero.hint": "Попробуйте кнопки — так же работает на стриме в OBS.",
      "look.kicker": "Внешний вид",
      "look.title": "Скины и режимы",
      "look.lead": "Готовые пресеты, свои цвета и анимации. Overwatch с ролями и Apex Legends — выберите игру под стрим.",
      "mode.ow": "Overwatch",
      "mode.owRoles": "OW · роли",
      "mode.apex": "Apex Legends",
      "feat.kicker": "Возможности",
      "feat.title": "Что умеет виджет",
      "feat.f1t": "Без мерцания",
      "feat.f1d": "Обновления по WebSocket — Browser Source не перезагружается.",
      "feat.f2t": "Веб-админка",
      "feat.f2d": "Настройка цветов, шрифтов, анимаций и скинов в браузере.",
      "feat.f3t": "Lua для OBS",
      "feat.f3d": "Автозапуск сервера с OBS, хоткеи без лишних окон.",
      "feat.f4t": "Stream Dock Live",
      "feat.f4d": "Победы, поражения и ранг на кнопках AJAZZ — Live Mode синхронизирует деку с виджетом.",
      "feat.icons": "LIVE-иконки Stream Dock",
      "feat.liveGuide": "Инструкция Live Mode",
      "install.kicker": "Установка",
      "install.title": "Установка за несколько минут",
      "i1.t": "Создайте папку для виджета",
      "i1.d": "Создайте отдельную папку — туда скачаете widget-stats.exe и widget_control.lua. Если при скачивании Windows Defender ругается, добавьте эту папку в исключения: у виджета нет сертификата Microsoft.",
      "i2.t": "Скачайте с GitHub",
      "i2.d": "Откройте Releases и скачайте оба файла в созданную папку.",
      "i2.a": "Страница Releases →",
      "i3.t": "Подключите Lua-скрипт в OBS",
      "i3.d": "Tools → Scripts → «+» → выберите widget_control.lua из папки с exe. Сервер запустится вместе с OBS.",
      "i4.t": "Откройте админку",
      "i4.d": "Иконка в трее → Open Admin. Здесь настраивается вид и подключение к OBS.",
      "i5.t": "Включите WebSocket в OBS",
      "i5.d": "Tools → WebSocket Server Settings → Enable. Порт по умолчанию — 4455.",
      "i6.t": "Поставьте виджет на сцену",
      "i6.d": "В админке: Подключить OBS → выберите сцену → Поставить виджет на сцену. Установка готова.",
      "setup.kicker": "Настройка",
      "setup.title": "Под себя",
      "setup.lead": "Горячие клавиши, внешний вид и скины — после установки.",
      "s1.t": "Горячие клавиши в OBS",
      "s1.d": "На видео — настройка хоткеев прямо в OBS: File → Settings → Hotkeys. Назначьте Win, Loss и Rank — статистика обновляется одним нажатием на стриме.",
      "s1.deckTitle": "Stream Deck и аналоги",
      "s1.deckLead": "Те же действия Win / Loss / Rank можно повесить на кнопки деки. URL копируются из админки — блок ниже.",
      "s1.deckAjazz": "Плагин нужен не всем. Если у вас Stream Dock от AJAZZ — поставьте HTTP-плагин. Есть классические иконки и LIVE-пак (цифры побед/поражений и код ранга на кнопках).",
      "s1.plugin": "Плагин HTTP (Live Mode)",
      "s1.gallery": "Иконки · подробнее",
      "s1.iconsLive": "Скачать LIVE-пак",
      "s1.liveGuide": "Инструкция Live Mode",
      "gallery.kicker": "Stream Deck",
      "gallery.lead": "Листайте паки стрелками. Внизу — живой виджет в том же стиле.",
      "s2.t": "Внешний вид",
      "s2.d": "Цвета, шрифт, фон и анимации — вкладка «Внешний вид» в админке. Превью обновляется сразу.",
      "s3.t": "Скины",
      "s3.d": "Галерея готовых пресетов — один клик, и виджет меняет стиль под стрим.",
      "clip.need": "Видео для этого шага скоро появится",
      "clip.expand": "На весь экран",
      "foot.cta": "Скачать бесплатно",
      "foot.help": "Вопросы и поддержка",
    },
    en: {
      "bar.sub": "Stream stats for OBS Studio",
      "bar.download": "Download",
      "nav.features": "Features",
      "nav.install": "Setup",
      "hero.kicker": "For OBS Studio",
      "hero.title": "Widget Stats",
      "hero.lead": "Wins, losses, and rank on stream — polished and easy. See how it looks on your overlay.",
      "hero.rankUp": "Rank +",
      "hero.rankDown": "Rank −",
      "hero.reset": "Reset",
      "hero.hint": "Try the buttons — same behavior on your OBS stream.",
      "look.kicker": "Look",
      "look.title": "Skins & modes",
      "look.lead": "Ready-made presets, custom colors and motion. Overwatch roles and Apex Legends — pick your game.",
      "mode.ow": "Overwatch",
      "mode.owRoles": "OW · roles",
      "mode.apex": "Apex Legends",
      "feat.kicker": "More",
      "feat.title": "What you get",
      "feat.f1t": "No flicker",
      "feat.f1d": "WebSocket updates — the Browser Source never reloads.",
      "feat.f2t": "Web admin",
      "feat.f2d": "Colors, fonts, motion, and skins in your browser.",
      "feat.f3t": "OBS Lua script",
      "feat.f3d": "Server starts with OBS; hotkeys without extra windows.",
      "feat.f4t": "Stream Dock Live",
      "feat.f4d": "Wins, losses, and rank on AJAZZ keys — Live Mode keeps the deck in sync with the widget.",
      "feat.icons": "Stream Dock LIVE icons",
      "feat.liveGuide": "Live Mode guide",
      "install.kicker": "Setup",
      "install.title": "Install in a few minutes",
      "i1.t": "Create a folder for the widget",
      "i1.d": "Make a dedicated folder for widget-stats.exe and widget_control.lua. If Windows Defender warns during download, add this folder to exclusions — the app has no Microsoft certificate.",
      "i2.t": "Download from GitHub",
      "i2.d": "Open Releases and download both files into that folder.",
      "i2.a": "Releases page →",
      "i3.t": "Add the Lua script in OBS",
      "i3.d": "Tools → Scripts → “+” → pick widget_control.lua from the exe folder. The server starts with OBS.",
      "i4.t": "Open the admin panel",
      "i4.d": "Tray icon → Open Admin. Configure look and OBS connection here.",
      "i5.t": "Enable OBS WebSocket",
      "i5.d": "Tools → WebSocket Server Settings → Enable. Default port is 4455.",
      "i6.t": "Place the widget on a scene",
      "i6.d": "In admin: Connect OBS → pick a scene → Place widget on scene. Install is done.",
      "setup.kicker": "Setup",
      "setup.title": "Make it yours",
      "setup.lead": "Hotkeys, look, and skins — after install.",
      "s1.t": "OBS hotkeys",
      "s1.d": "The clip shows OBS hotkeys: File → Settings → Hotkeys. Bind Win, Loss, and Rank — one key updates stats on stream.",
      "s1.deckTitle": "Stream Deck and similar devices",
      "s1.deckLead": "The same Win / Loss / Rank actions can sit on deck buttons. Copy the URLs from the admin panel — screenshot below.",
      "s1.deckAjazz": "The plugin is optional. For AJAZZ Stream Dock install the HTTP plugin. Classic icons and a LIVE pack (win/loss counts and short rank codes on keys) are available.",
      "s1.plugin": "HTTP plugin (Live Mode)",
      "s1.gallery": "Icons · details",
      "s1.iconsLive": "Download LIVE pack",
      "s1.liveGuide": "Live Mode guide",
      "gallery.kicker": "Stream Deck",
      "gallery.lead": "Browse packs with the arrows. The live widget below matches the style.",
      "s2.t": "Look",
      "s2.d": "Colors, font, background, and motion — Look tab in the admin. Preview updates live.",
      "s3.t": "Skins",
      "s3.d": "Ready-made presets — one click and the widget matches your stream style.",
      "clip.need": "Video for this step coming soon",
      "clip.expand": "Fullscreen",
      "foot.cta": "Download free",
      "foot.help": "Questions & support",
    },
  };

  const FEATURED_SKIN_IDS = [
    "default",
    "blood-moon",
    "solar-flare",
    "ocean-deep",
    "nordic-frost",
    "mono-ink",
  ];

  const frame = document.getElementById("widgetFrame");
  const skinRow = document.getElementById("skinRow");
  const langBtns = [...document.querySelectorAll(".lang-btn")];
  const modeChips = [...document.querySelectorAll("[data-mode]")];
  const mainTabs = [...document.querySelectorAll(".section-tab")];
  const tabPanels = [...document.querySelectorAll(".tab-panel")];
  const ICON_SHEETS = [
    "default", "blood-moon", "solar-flare", "ocean-deep", "nordic-frost", "mono-ink",
    "8bit-arena", "amber-library", "aqua-glass", "ash-veil", "bass-drop", "carbon-race",
    "crystal-break", "cyan-project", "cyber-cyan", "data-breach", "deep-forest",
    "ember-smoke", "forge-amber", "ghost-lantern", "glitch-district", "gold-rush",
    "grid-runner", "holo-deck", "indigo-scroll", "ivory-court", "jungle-heat",
    "liquid-steel", "magma-core", "mercury-run", "midnight-void", "neon-alley",
    "obsidian-glass", "opal-shine", "paper-light", "petal-dawn", "pink-scanline",
    "rainbow-cut", "retro-cabinet", "sakura-dusk", "spiral-teal", "storm-call",
    "sumi-night", "tide-pool", "toxic-spill", "violet-haze", "voltage-night", "wraith-mist",
  ];

  const iconGallery = document.getElementById("iconGallery");
  const iconGalleryClose = document.getElementById("iconGalleryClose");
  const openIconGalleryBtn = document.getElementById("openIconGallery");
  const deckShotImg = document.getElementById("deckShotImg");
  const galleryImage = document.getElementById("galleryImage");
  const galleryName = document.getElementById("galleryName");
  const galleryIndexEl = document.getElementById("galleryIndex");
  const galleryTotalEl = document.getElementById("galleryTotal");
  const galleryPrev = document.getElementById("galleryPrev");
  const galleryNext = document.getElementById("galleryNext");
  const galleryWidgetFrame = document.getElementById("galleryWidgetFrame");

  let galleryIndex = 0;

  function sheetTitle(slug) {
    return slug.split("-").map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(" ");
  }

  function sheetSrc(slug) {
    return `assets/Icon preview/_sheet_${slug}.png`;
  }

  function postTo(targetFrame, msg) {
    try {
      targetFrame?.contentWindow?.postMessage(msg, "*");
    } catch (_) {
      /* ignore */
    }
  }

  function applyGallerySkin(slug) {
    const skins = Array.isArray(window.WIDGET_SKINS) ? window.WIDGET_SKINS : [];
    const skin = skins.find((s) => s.id === slug);
    if (!skin?.settings) return;
    if (skin.settings.font) window.loadGoogleFont?.(skin.settings.font);
    postTo(galleryWidgetFrame, {
      type: "widget-preview-settings",
      settings: { ...skin.settings, fontSize: Math.max(20, skin.settings.fontSize || 18) },
    });
    postTo(galleryWidgetFrame, {
      type: "widget-preview-state",
      view: { wins: 4, losses: 2, rank: 12, mode: "classic", role: "tank", game: "overwatch" },
      state: { wins: 4, losses: 2, rank: 12, mode: "classic", role: "tank", game: "overwatch" },
    });
  }

  function showGallerySlide(index) {
    galleryIndex = (index + ICON_SHEETS.length) % ICON_SHEETS.length;
    const slug = ICON_SHEETS[galleryIndex];
    galleryImage.src = sheetSrc(slug);
    galleryImage.alt = sheetTitle(slug);
    galleryName.textContent = sheetTitle(slug);
    galleryIndexEl.textContent = String(galleryIndex + 1);
    applyGallerySkin(slug);
    const next = ICON_SHEETS[(galleryIndex + 1) % ICON_SHEETS.length];
    const prev = ICON_SHEETS[(galleryIndex - 1 + ICON_SHEETS.length) % ICON_SHEETS.length];
    [next, prev].forEach((s) => {
      const preload = new Image();
      preload.src = sheetSrc(s);
    });
  }

  function openIconGallery() {
    galleryTotalEl.textContent = String(ICON_SHEETS.length);
    iconGallery.hidden = false;
    document.body.style.overflow = "hidden";
    showGallerySlide(galleryIndex);
  }

  function closeIconGallery() {
    iconGallery.hidden = true;
    if (lightbox.hidden) document.body.style.overflow = "";
  }
  const lightbox = document.getElementById("lightbox");
  const lightboxFrame = document.getElementById("lightboxFrame");
  const lightboxClose = document.getElementById("lightboxClose");

  let lang = localStorage.getItem("ws-demo-lang") || "ru";
  let currentSkin = null;
  let activeTab = localStorage.getItem("ws-demo-tab") || "install";
  let demoState = {
    wins: 0,
    losses: 0,
    rank: 12,
    mode: "classic",
    role: "tank",
    game: "overwatch",
    roleCycle: ["tank", "support", "damage"],
  };

  function t(key) {
    return (I18N[lang] && I18N[lang][key]) || I18N.ru[key] || key;
  }

  function applyI18n() {
    document.documentElement.lang = lang;
    document.querySelectorAll("[data-i18n]").forEach((el) => {
      el.textContent = t(el.getAttribute("data-i18n"));
    });
    langBtns.forEach((btn) => btn.classList.toggle("active", btn.dataset.lang === lang));
    document.querySelectorAll(".guide-media.is-waiting [data-ph]").forEach((el) => {
      el.textContent = t("clip.need");
    });
  }

  function showTab(id) {
    activeTab = id;
    localStorage.setItem("ws-demo-tab", id);
    mainTabs.forEach((tab) => {
      const on = tab.dataset.tab === id;
      tab.classList.toggle("active", on);
      tab.setAttribute("aria-selected", on ? "true" : "false");
    });
    tabPanels.forEach((panel) => {
      const on = panel.id === `tab-${id}`;
      panel.classList.toggle("active", on);
      panel.hidden = !on;
    });
    if (id === "install") {
      setTimeout(syncGuideVideos, 80);
    } else {
      document.querySelectorAll("video[data-guide-video]").forEach(pauseGuideVideo);
    }
  }

  function post(msg) {
    try {
      frame.contentWindow?.postMessage(msg, "*");
    } catch (_) {
      /* ignore */
    }
  }

  function pushSettings(settings) {
    if (!settings) return;
    if (settings.font) window.loadGoogleFont?.(settings.font);
    post({
      type: "widget-preview-settings",
      settings: { ...settings, fontSize: Math.max(22, settings.fontSize || 18) },
    });
  }

  function pushState(extra = {}) {
    demoState = { ...demoState, ...extra };
    post({ type: "widget-preview-state", view: { ...demoState }, state: { ...demoState } });
  }

  function runAction(action) {
    if (currentSkin?.settings) pushSettings(currentSkin.settings);

    if (action === "win") {
      demoState.wins += 1;
      post({ type: "widget-preview-demo", action: "win" });
      return;
    }
    if (action === "loss") {
      demoState.losses += 1;
      post({ type: "widget-preview-demo", action: "loss" });
      return;
    }
    if (action === "rankStepUp") {
      demoState.rank += 1;
      post({ type: "widget-preview-demo", action: "rankStepUp" });
      return;
    }
    if (action === "rankStepDown") {
      demoState.rank = Math.max(0, demoState.rank - 1);
      post({ type: "widget-preview-demo", action: "rankStepDown" });
      return;
    }
    if (action === "resetStats") {
      demoState.wins = 0;
      demoState.losses = 0;
      post({ type: "widget-preview-demo", action: "resetStats" });
    }
  }

  function skinsList() {
    const all = Array.isArray(window.WIDGET_SKINS) ? window.WIDGET_SKINS : [];
    const picked = FEATURED_SKIN_IDS.map((id) => all.find((s) => s.id === id)).filter(Boolean);
    return picked.length ? picked : all.slice(0, 6);
  }

  function renderSkins() {
    const skins = skinsList();
    skinRow.innerHTML = "";
    skins.forEach((skin, i) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "skin-card" + (i === 0 ? " active" : "");
      btn.innerHTML = `
        <div class="skin-swatch" style="background:${skin.preview.bg}">
          <span class="skin-chip" style="background:${skin.preview.wins}"></span>
          <span class="skin-chip" style="background:${skin.preview.losses}"></span>
        </div>
        <strong>${skin.name}</strong>
        <span>${skin.desc || ""}</span>
      `;
      btn.addEventListener("click", () => {
        skinRow.querySelectorAll(".skin-card").forEach((el) => el.classList.remove("active"));
        btn.classList.add("active");
        currentSkin = skin;
        pushSettings(skin.settings || {});
      });
      skinRow.appendChild(btn);
      if (i === 0) currentSkin = skin;
    });
  }

  function setMode(mode) {
    modeChips.forEach((c) => c.classList.toggle("active", c.dataset.mode === mode));
    if (mode === "ow") {
      pushState({ game: "overwatch", mode: "classic", role: "tank" });
    } else if (mode === "ow-roles") {
      pushState({ game: "overwatch", mode: "roles_split", role: "support" });
    } else if (mode === "apex") {
      pushState({ game: "apex", mode: "classic", role: "tank" });
    }
    if (currentSkin?.settings) pushSettings(currentSkin.settings);
  }

  function isVideo(name) {
    return /\.webm$/i.test(name) || /\.mp4$/i.test(name);
  }

  function assetUrl(file) {
    return `assets/${encodeURIComponent(file)}`;
  }

  function probe(file) {
    return fetch(assetUrl(file), { method: "HEAD" })
      .then((res) => res.ok)
      .catch(() => false);
  }

  async function resolveMedia(files) {
    for (const file of files) {
      if (await probe(file)) return { file, url: assetUrl(file) };
    }
    return null;
  }

  function createGuideVideo(url) {
    const media = document.createElement("video");
    media.src = url;
    media.muted = true;
    media.loop = true;
    media.playsInline = true;
    media.autoplay = true;
    media.preload = "auto";
    media.setAttribute("controlsList", "nodownload nofullscreen noremoteplayback");
    media.disablePictureInPicture = true;
    media.dataset.guideVideo = "1";
    return media;
  }

  function playGuideVideo(video) {
    if (!video || video.tagName !== "VIDEO") return;
    const p = video.play();
    if (p && typeof p.catch === "function") p.catch(() => {});
  }

  function pauseGuideVideo(video) {
    if (!video || video.tagName !== "VIDEO") return;
    video.pause();
    try {
      video.currentTime = 0;
    } catch (_) {
      /* ignore */
    }
  }

  function syncGuideVideos() {
    if (activeTab !== "install") {
      document.querySelectorAll("video[data-guide-video]").forEach(pauseGuideVideo);
      return;
    }
    document.querySelectorAll(".guide-media.is-ready video[data-guide-video]").forEach((video) => {
      const box = video.closest(".guide-media");
      if (!box) return;
      const rect = box.getBoundingClientRect();
      const visible = rect.top < window.innerHeight * 0.85 && rect.bottom > window.innerHeight * 0.15;
      if (visible) playGuideVideo(video);
      else pauseGuideVideo(video);
    });
  }

  let videoObserver = null;

  function watchGuideVideos() {
    if (videoObserver) videoObserver.disconnect();
    videoObserver = new IntersectionObserver(
      (entries) => {
        if (activeTab !== "install") return;
        entries.forEach((entry) => {
          const video = entry.target.querySelector("video[data-guide-video]");
          if (!video) return;
          if (entry.isIntersecting) playGuideVideo(video);
          else pauseGuideVideo(video);
        });
      },
      { threshold: 0.35, rootMargin: "0px 0px -8% 0px" }
    );
    document.querySelectorAll(".guide-media.is-ready").forEach((box) => videoObserver.observe(box));
  }

  function openLightbox(el) {
    lightboxFrame.innerHTML = "";
    const clone = el.cloneNode(true);
    clone.removeAttribute("style");
    clone.style.maxWidth = "100%";
    clone.style.maxHeight = "90vh";
    clone.style.cursor = "default";
    if (clone.tagName === "VIDEO") {
      clone.controls = false;
      clone.muted = true;
      clone.loop = true;
      clone.autoplay = true;
    }
    lightboxFrame.appendChild(clone);
    lightbox.hidden = false;
    document.body.style.overflow = "hidden";
    if (clone.tagName === "VIDEO") playGuideVideo(clone);
  }

  function closeLightbox() {
    lightbox.hidden = true;
    lightboxFrame.innerHTML = "";
    if (iconGallery.hidden) document.body.style.overflow = "";
  }

  async function mountGuideMedia() {
    const steps = [...document.querySelectorAll(".guide-step[data-media-files]")];
    for (const step of steps) {
      const box = step.querySelector(".guide-media");
      if (!box) continue;
      const optional = step.hasAttribute("data-media-optional");
      const files = step
        .getAttribute("data-media-files")
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);

      const found = await resolveMedia(files);

      if (!found) {
        if (optional) {
          box.className = "guide-media is-hidden";
          box.innerHTML = "";
        } else {
          box.className = "guide-media is-waiting";
          box.innerHTML = `<span data-ph>${t("clip.need")}</span><code>${files[0] || ""}</code>`;
        }
        continue;
      }

      box.className = "guide-media is-ready guide-media-video";
      box.innerHTML = "";
      let media;
      if (isVideo(found.file)) {
        media = createGuideVideo(found.url);
      } else {
        media = document.createElement("img");
        media.src = found.url;
        media.alt = found.file;
      }
      media.addEventListener("click", () => openLightbox(media));
      box.appendChild(media);

      const expand = document.createElement("button");
      expand.type = "button";
      expand.className = "expand-btn";
      expand.textContent = t("clip.expand");
      expand.addEventListener("click", (e) => {
        e.stopPropagation();
        openLightbox(media);
      });
      box.appendChild(expand);
    }
    watchGuideVideos();
    syncGuideVideos();
  }

  function bootPreview() {
    if (currentSkin) pushSettings(currentSkin.settings);
    pushState();
  }

  document.querySelectorAll("[data-action]").forEach((btn) => {
    btn.addEventListener("click", () => runAction(btn.dataset.action));
  });

  modeChips.forEach((chip) => {
    chip.addEventListener("click", () => setMode(chip.dataset.mode));
  });

  mainTabs.forEach((tab) => {
    tab.addEventListener("click", () => showTab(tab.dataset.tab));
  });

  langBtns.forEach((btn) => {
    btn.addEventListener("click", () => {
      lang = btn.dataset.lang;
      localStorage.setItem("ws-demo-lang", lang);
      applyI18n();
    });
  });

  lightboxClose.addEventListener("click", closeLightbox);
  lightbox.addEventListener("click", (e) => {
    if (e.target === lightbox) closeLightbox();
  });
  openIconGalleryBtn?.addEventListener("click", openIconGallery);
  iconGalleryClose?.addEventListener("click", closeIconGallery);
  galleryPrev?.addEventListener("click", () => showGallerySlide(galleryIndex - 1));
  galleryNext?.addEventListener("click", () => showGallerySlide(galleryIndex + 1));
  galleryWidgetFrame?.addEventListener("load", () => {
    if (!iconGallery.hidden) applyGallerySkin(ICON_SHEETS[galleryIndex]);
  });

  let galleryTouchX = null;
  galleryImage?.addEventListener("touchstart", (e) => {
    galleryTouchX = e.changedTouches[0].screenX;
  }, { passive: true });
  galleryImage?.addEventListener("touchend", (e) => {
    if (galleryTouchX == null) return;
    const dx = e.changedTouches[0].screenX - galleryTouchX;
    galleryTouchX = null;
    if (Math.abs(dx) < 40) return;
    if (dx < 0) showGallerySlide(galleryIndex + 1);
    else showGallerySlide(galleryIndex - 1);
  }, { passive: true });

  deckShotImg?.addEventListener("click", () => openLightbox(deckShotImg));
  document.addEventListener("keydown", (e) => {
    if (!iconGallery.hidden) {
      if (e.key === "Escape") {
        closeIconGallery();
        return;
      }
      if (e.key === "ArrowRight") {
        e.preventDefault();
        showGallerySlide(galleryIndex + 1);
        return;
      }
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        showGallerySlide(galleryIndex - 1);
        return;
      }
    }
    if (!lightbox.hidden && e.key === "Escape") closeLightbox();
  });

  window.addEventListener("scroll", () => syncGuideVideos(), { passive: true });
  window.addEventListener("resize", () => syncGuideVideos());

  if (location.hash === "#install") activeTab = "install";

  frame.addEventListener("load", () => bootPreview());

  applyI18n();
  renderSkins();
  mountGuideMedia();
  showTab(activeTab);

  if (frame.contentDocument?.readyState === "complete") bootPreview();
})();

(() => {
  const I18N = {
    ru: {
      "bar.sub": "Статистика для OBS Studio",
      "bar.download": "Скачать",
      "nav.look": "Внешний вид",
      "nav.features": "Возможности",
      "nav.install": "Установка",
      "nav.setup": "Настройка",
      "hero.kicker": "Для OBS Studio · Overwatch 2 и Apex Legends",
      "hero.title": "Widget Stats",
      "hero.lead": "Победы, поражения и ранг — прямо на стриме. Красиво, без мерцания, а в Overwatch 2 ещё и считает сам. Бесплатно.",
      "hero.rankUp": "Ранг +",
      "hero.rankDown": "Ранг −",
      "hero.reset": "Сброс",
      "hero.new": "Новое в 3.0",
      "hero.newText": "19 видов виджета, новая админка и умнее автоматический ранг",
      "look.kicker": "Внешний вид",
      "look.title": "Соберите свой виджет",
      "look.lead": "Выберите вид, скин и игру — результат сразу в большом виджете наверху. Всё здесь живое, как на стриме.",
      "look.new": "Новое",
      "look.layoutsT": "Вид",
      "look.layoutsD": "19 вариантов: подробный, компактный, карточка, табло или вся сессия. Нажмите на любой — он появится наверху.",
      "look.skinsT": "Скин",
      "look.skinsD": "48 готовых скинов. Здесь — несколько для примера, все остальные в админке: с поиском и живым превью.",
      "look.moreSkins": "+{n} скинов",
      "look.moreSkinsD": "в админке, с поиском и превью",
      "look.gameT": "Игра",
      "look.gameD": "Overwatch 2 — общий счёт или отдельный по ролям. Apex Legends — свои ранги.",
      "mode.ow": "Overwatch",
      "mode.owRoles": "OW · роли",
      "mode.apex": "Apex Legends",
      "feat.kicker": "Возможности",
      "feat.title": "Что умеет виджет",
      "feat.autoT": "Автозачёт и автоматический ранг",
      "feat.autoD": "В Overwatch 2 победа, поражение и ранг считаются сами: виджет читает экран игры. Ничего настраивать не нужно — включается одной галочкой в разделе «Автозачёт».",
      "feat.studioT": "Удобная админка",
      "feat.studioD": "Живое превью для каждого вида, всё сохраняется само, есть «Отменить». Настройки, которые вид не использует, скрыты.",
      "feat.sessionT": "Сессия стрима",
      "feat.sessionD": "Виджет помнит ранг на начало стрима и показывает, на сколько вы поднялись и как прошли последние игры.",
      "feat.updateT": "Обновление в один клик",
      "feat.updateD": "Админка сама сообщает о новой версии и ещё до обновления показывает, что изменится.",
      "feat.helpT": "Помощь простыми словами",
      "feat.helpD": "Справка с поиском прямо в админке и кнопка «Написать разработчику».",
      "feat.flickerT": "Без мерцания",
      "feat.flickerD": "Обновления по WebSocket — Browser Source в OBS не перезагружается.",
      "feat.luaT": "Скрипт для OBS",
      "feat.luaD": "Виджет запускается вместе с OBS, горячие клавиши работают без лишних окон.",
      "feat.deckT": "Stream Deck и Stream Dock",
      "feat.deckD": "Win, Loss и ранг на кнопках деки. LIVE-пак показывает счёт и ранг прямо на кнопках.",
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
      "i4.d": "Иконка в трее → Open Admin. Здесь настраивается всё: вид, скины и подключение к OBS.",
      "i5.t": "Включите WebSocket в OBS",
      "i5.d": "Tools → WebSocket Server Settings → Enable. Порт по умолчанию — 4455.",
      "i6.t": "Поставьте виджет на сцену",
      "i6.d": "В админке: «Подключения» → «Подключить OBS» → выберите сцену → «Поставить виджет на сцену». Установка готова.",
      "setup.kicker": "Настройка",
      "setup.title": "Под себя",
      "setup.lead": "После установки: вид, скин, цвета и горячие клавиши.",
      "st1.t": "Выберите вид",
      "st1.d": "Админка → «Вид виджета» → нажмите на любой. Превью справа сразу показывает, как он реагирует на победу и смену ранга. Сохраняется само.",
      "st2.t": "Подберите скин",
      "st2.d": "«Скины» → наведите на скин, и превью покажет ваш виджет в нём. Клик — применить. Есть поиск, подборки по настроению и избранное.",
      "st3.t": "Цвета, шрифт и анимации",
      "st3.d": "«Цвета и шрифт» и «Анимации» — меняйте что угодно, превью обновляется сразу. Передумали — «Отменить» внизу экрана.",
      "s1.t": "Горячие клавиши в OBS",
      "s1.d": "File → Settings → Hotkeys. Назначьте Win, Loss и Rank — статистика обновляется одним нажатием на стриме.",
      "s1.deckTitle": "Stream Deck и аналоги",
      "s1.deckLead": "Те же действия можно повесить на кнопки деки. Ссылки копируются в админке: «Подключения» → «URL для OBS / Stream Deck».",
      "s1.deckAjazz": "Плагин нужен не всем. Если у вас Stream Dock от AJAZZ — поставьте HTTP-плагин. Есть классические иконки и LIVE-пак (цифры побед/поражений и код ранга на кнопках).",
      "s1.plugin": "Плагин HTTP (Live Mode)",
      "s1.gallery": "Иконки · подробнее",
      "s1.iconsLive": "Скачать LIVE-пак",
      "s1.liveGuide": "Инструкция Live Mode",
      "gallery.kicker": "Stream Deck",
      "gallery.lead": "Листайте паки стрелками. Внизу — живой виджет в том же стиле.",
      "clip.need": "Видео для этого шага скоро появится",
      "clip.expand": "На весь экран",
      "foot.getT": "Бесплатно и с открытым кодом",
      "foot.cta": "Скачать бесплатно",
      "foot.askT": "Вопросы и идеи",
      "foot.bug": "Сообщить об ошибке",
      "foot.supportT": "Поддержать автора",
      "layout.detailed": "Подробный", "layout.mini": "Мини", "layout.card": "Карточка", "layout.badge": "Значок",
      "layout.strip": "Полоса", "layout.minimal": "Минимал", "layout.column": "Колонка", "layout.ring": "Кольцо",
      "layout.slant": "Наклон (OW)", "layout.scoreboard": "Табло", "layout.session": "Сессия", "layout.form": "Форма (5 игр)",
      "layout.ladder": "Лестница", "layout.roles": "Все роли", "layout.hex": "Гекс", "layout.feed": "Килфид",
      "layout.tug": "Перетягивание", "layout.flip": "Перекидное табло", "layout.tcard": "Коллекционная карта",
      "hero.download": "Скачать бесплатно",
      "hero.howto": "Как установить →",
      "hero.try": "Попробуйте прямо здесь:",
      "num.layouts": "видов виджета",
      "num.skins": "готовых скинов",
      "num.autoV": "Авто",
      "num.auto": "победы и ранг в Overwatch 2 считаются сами",
      "num.minV": "5 мин",
      "num.min": "на установку",
      "cta.title": "Готовы попробовать на своём стриме?",
      "cta.lead": "Бесплатно. Установка занимает несколько минут — пошаговая инструкция с видео выше.",
      "cta.guide": "Инструкция по установке",
      "foot.releases": "Все версии",
    },
    en: {
      "bar.sub": "Stream stats for OBS Studio",
      "bar.download": "Download",
      "nav.look": "Look",
      "nav.features": "Features",
      "nav.install": "Install",
      "nav.setup": "Setup",
      "hero.kicker": "For OBS Studio · Overwatch 2 and Apex Legends",
      "hero.title": "Widget Stats",
      "hero.lead": "Wins, losses and rank — right on stream. Good-looking, flicker-free, and in Overwatch 2 it even counts by itself. Free.",
      "hero.rankUp": "Rank +",
      "hero.rankDown": "Rank −",
      "hero.reset": "Reset",
      "hero.new": "New in 3.0",
      "hero.newText": "19 widget layouts, a new control panel and smarter auto rank",
      "look.kicker": "Look",
      "look.title": "Build your widget",
      "look.lead": "Pick a layout, a skin and a game — the big widget on top shows the result. Everything here is live, just like on stream.",
      "look.new": "New",
      "look.layoutsT": "Layout",
      "look.layoutsD": "19 options: detailed, compact, card, scoreboard or the whole session. Click any of them to put it on top.",
      "look.skinsT": "Skin",
      "look.skinsD": "48 ready-made skins. A few are here as examples; all of them are in the admin, with search and a live preview.",
      "look.moreSkins": "+{n} skins",
      "look.moreSkinsD": "in the admin, with search and preview",
      "look.gameT": "Game",
      "look.gameD": "Overwatch 2 — one score or a separate one per role. Apex Legends — its own ranks.",
      "mode.ow": "Overwatch",
      "mode.owRoles": "OW · roles",
      "mode.apex": "Apex Legends",
      "feat.kicker": "Features",
      "feat.title": "What the widget does",
      "feat.autoT": "Auto W/L and auto rank",
      "feat.autoD": "In Overwatch 2 wins, losses and rank count themselves: the widget reads the game screen. Nothing to set up — one checkbox in the Auto W/L section turns it on.",
      "feat.studioT": "Easy control panel",
      "feat.studioD": "A live preview for every layout, everything saves by itself, with Undo. Settings the layout doesn't use are hidden.",
      "feat.sessionT": "Stream session",
      "feat.sessionD": "The widget remembers your rank at the start of the stream and shows how far you climbed and how your last games went.",
      "feat.updateT": "One-click updates",
      "feat.updateD": "The admin tells you about a new version and shows what changes before you update.",
      "feat.helpT": "Help in plain words",
      "feat.helpD": "Searchable help right in the admin and a “Contact the developer” button.",
      "feat.flickerT": "No flicker",
      "feat.flickerD": "WebSocket updates — the OBS Browser Source never reloads.",
      "feat.luaT": "OBS script",
      "feat.luaD": "The widget starts with OBS; hotkeys work without extra windows.",
      "feat.deckT": "Stream Deck and Stream Dock",
      "feat.deckD": "Win, Loss and rank on deck keys. The LIVE pack shows the score and rank right on the keys.",
      "install.kicker": "Install",
      "install.title": "Install in a few minutes",
      "i1.t": "Create a folder for the widget",
      "i1.d": "Make a dedicated folder for widget-stats.exe and widget_control.lua. If Windows Defender warns during download, add this folder to exclusions — the app has no Microsoft certificate.",
      "i2.t": "Download from GitHub",
      "i2.d": "Open Releases and download both files into that folder.",
      "i2.a": "Releases page →",
      "i3.t": "Add the Lua script in OBS",
      "i3.d": "Tools → Scripts → “+” → pick widget_control.lua from the exe folder. The server starts with OBS.",
      "i4.t": "Open the admin panel",
      "i4.d": "Tray icon → Open Admin. Everything is set up here: layout, skins and the OBS connection.",
      "i5.t": "Enable OBS WebSocket",
      "i5.d": "Tools → WebSocket Server Settings → Enable. Default port is 4455.",
      "i6.t": "Place the widget on a scene",
      "i6.d": "In the admin: Connections → Connect OBS → pick a scene → Place widget on scene. Install is done.",
      "setup.kicker": "Setup",
      "setup.title": "Make it yours",
      "setup.lead": "After install: layout, skin, colors and hotkeys.",
      "st1.t": "Pick a layout",
      "st1.d": "Admin → Layout → click any layout. The preview on the right shows how it reacts to a win and a rank change. It saves by itself.",
      "st2.t": "Choose a skin",
      "st2.d": "Skins → hover a skin and the preview shows your widget in it. Click to apply. Search, moods and favorites included.",
      "st3.t": "Colors, font and motion",
      "st3.d": "Colours & font and Motion — change anything, the preview updates at once. Changed your mind? Undo at the bottom.",
      "s1.t": "OBS hotkeys",
      "s1.d": "File → Settings → Hotkeys. Bind Win, Loss and Rank — one key updates stats on stream.",
      "s1.deckTitle": "Stream Deck and similar devices",
      "s1.deckLead": "The same actions can sit on deck keys. Copy the links in the admin: Connections → URLs for OBS / Stream Deck.",
      "s1.deckAjazz": "The plugin is optional. For AJAZZ Stream Dock install the HTTP plugin. Classic icons and a LIVE pack (win/loss counts and short rank codes on keys) are available.",
      "s1.plugin": "HTTP plugin (Live Mode)",
      "s1.gallery": "Icons · details",
      "s1.iconsLive": "Download LIVE pack",
      "s1.liveGuide": "Live Mode guide",
      "gallery.kicker": "Stream Deck",
      "gallery.lead": "Browse packs with the arrows. The live widget below matches the style.",
      "clip.need": "Video for this step coming soon",
      "clip.expand": "Fullscreen",
      "foot.getT": "Free and open source",
      "foot.cta": "Download free",
      "foot.askT": "Questions and ideas",
      "foot.bug": "Report a bug",
      "foot.supportT": "Support the author",
      "layout.detailed": "Detailed", "layout.mini": "Mini", "layout.card": "Card", "layout.badge": "Badge",
      "layout.strip": "Strip", "layout.minimal": "Minimal", "layout.column": "Column", "layout.ring": "Ring",
      "layout.slant": "Slant (OW)", "layout.scoreboard": "Scoreboard", "layout.session": "Session", "layout.form": "Form (last 5)",
      "layout.ladder": "Ladder", "layout.roles": "All roles", "layout.hex": "Hex", "layout.feed": "Kill feed",
      "layout.tug": "Tug of war", "layout.flip": "Flip clock", "layout.tcard": "Trading card",
      "hero.download": "Download free",
      "hero.howto": "How to install →",
      "hero.try": "Try it right here:",
      "num.layouts": "widget layouts",
      "num.skins": "ready-made skins",
      "num.autoV": "Auto",
      "num.auto": "wins and rank count themselves in Overwatch 2",
      "num.minV": "5 min",
      "num.min": "to install",
      "cta.title": "Ready to try it on your stream?",
      "cta.lead": "Free. Install takes a few minutes — the step-by-step guide with videos is above.",
      "cta.guide": "Installation guide",
      "foot.releases": "All versions",
    },
  };

  const FEATURED_SKIN_IDS = [
    "default", "neon-alley", "gold-rush", "blood-moon", "ocean-deep", "sakura-dusk",
    "magma-core", "retro-cabinet", "solar-flare", "nordic-frost", "mono-ink",
  ];

  const frame = document.getElementById("widgetFrame");
  const skinRow = document.getElementById("skinRow");
  const langBtns = [...document.querySelectorAll(".lang-btn")];
  const modeChips = [...document.querySelectorAll("[data-mode]")];
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
      settings: { ...settings, layout: currentLayout, fontSize: Math.max(22, settings.fontSize || 18) },
    });
  }

  // ---------- Layout gallery: one live widget per layout ----------
  const LAYOUTS = ["detailed", "mini", "card", "badge", "strip", "minimal", "column", "ring", "slant", "scoreboard",
    "session", "form", "ladder", "roles", "hex", "feed", "tug", "flip", "tcard"];
  const TILE_STATE = { game: "overwatch", mode: "classic", role: "tank", wins: 7, losses: 4, rank: 21, startRank: 18,
    recent: ["w", "w", "l", "w", "l", "w", "w", "l", "w", "w"] };
  const TILE_ACTIONS = ["win", "loss", "rankStepUp", "win"];
  const layoutGrid = document.getElementById("layoutGrid");
  let currentLayout = "detailed";
  let tiles = [];

  function tileSettings(layout) {
    return { ...(currentSkin?.settings || {}), layout, alwaysVisible: true };
  }

  function renderLayoutTiles() {
    if (!layoutGrid) return;
    layoutGrid.innerHTML = "";
    tiles = LAYOUTS.map((layout) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "layout-tile" + (layout === currentLayout ? " active" : "");
      btn.innerHTML = `<div class="layout-stage"><iframe title="${layout}" tabindex="-1" loading="lazy" src="../../web/overlay/index.html?preview=1"></iframe></div>
        <span class="layout-name" data-i18n="layout.${layout}">${t(`layout.${layout}`)}</span>`;
      const f = btn.querySelector("iframe");
      // The widget loads its own offline defaults shortly after "load", so send the demo again.
      f.addEventListener("load", () => {
        [0, 700, 1600].forEach((ms) => setTimeout(() => {
          postTo(f, { type: "widget-preview-settings", settings: tileSettings(layout) });
          postTo(f, { type: "widget-preview-state", view: { ...TILE_STATE }, state: { ...TILE_STATE } });
        }, ms));
      });
      btn.addEventListener("click", () => {
        currentLayout = layout;
        layoutGrid.querySelectorAll(".layout-tile").forEach((el) => el.classList.toggle("active", el === btn));
        pushSettings(currentSkin?.settings || {});
        document.getElementById("play")?.scrollIntoView({ behavior: "smooth", block: "start" });
      });
      layoutGrid.appendChild(btn);
      return { layout, frame: f };
    });
  }

  // Every few seconds one visible tile plays a win, loss or rank-up, so the grid feels alive.
  let tileTick = 0;
  setInterval(() => {
    if (document.hidden || !tiles.length) return;
    const visible = tiles.filter(({ frame: f }) => {
      const r = f.getBoundingClientRect();
      const rail = layoutGrid.getBoundingClientRect();
      return r.bottom > 0 && r.top < window.innerHeight && r.right > rail.left && r.left < rail.right;
    });
    if (!visible.length) return;
    const tile = visible[tileTick % visible.length];
    postTo(tile.frame, { type: "widget-preview-demo", action: TILE_ACTIONS[tileTick % TILE_ACTIONS.length] });
    tileTick += 1;
  }, 1400);

  document.querySelectorAll("[data-rail]").forEach((btn) => {
    btn.addEventListener("click", () => {
      layoutGrid.scrollBy({ left: Number(btn.dataset.rail) * layoutGrid.clientWidth * 0.8, behavior: "smooth" });
    });
  });

  function restyleLayoutTiles() {
    tiles.forEach(({ layout, frame: f }) => postTo(f, { type: "widget-preview-settings", settings: tileSettings(layout) }));
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
    if (!currentSkin) currentSkin = skins[0] || null;
    skinRow.innerHTML = "";
    skins.forEach((skin) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "skin-card" + (skin === currentSkin ? " active" : "");
      // Skin descriptions in skins.js are Russian only.
      btn.innerHTML = `
        <div class="skin-swatch" style="background:${skin.preview.bg}">
          <span class="skin-chip" style="background:${skin.preview.wins}"></span>
          <span class="skin-chip" style="background:${skin.preview.losses}"></span>
        </div>
        <strong>${skin.name}</strong>
        <span>${lang === "ru" ? skin.desc || "" : ""}</span>
      `;
      btn.addEventListener("click", () => {
        skinRow.querySelectorAll(".skin-card").forEach((el) => el.classList.remove("active"));
        btn.classList.add("active");
        currentSkin = skin;
        pushSettings(skin.settings || {});
        restyleLayoutTiles();
      });
      skinRow.appendChild(btn);
    });
    const total = Array.isArray(window.WIDGET_SKINS) ? window.WIDGET_SKINS.length : 0;
    if (total > skins.length) {
      const more = document.createElement("div");
      more.className = "skin-card skin-more";
      more.innerHTML = `<strong>${t("look.moreSkins").replace("{n}", total - skins.length)}</strong><span>${t("look.moreSkinsD")}</span>`;
      skinRow.appendChild(more);
    }
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

  langBtns.forEach((btn) => {
    btn.addEventListener("click", () => {
      lang = btn.dataset.lang;
      localStorage.setItem("ws-demo-lang", lang);
      applyI18n();
      renderSkins();
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

  frame.addEventListener("load", () => bootPreview());

  applyI18n();
  renderSkins();
  renderLayoutTiles();
  mountGuideMedia();

  if (frame.contentDocument?.readyState === "complete") bootPreview();
})();

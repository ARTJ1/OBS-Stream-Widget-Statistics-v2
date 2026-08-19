(() => {
  const I18N = {
    ru: {
      "bar.sub": "Живое демо · без установки",
      "bar.download": "Скачать",
      "hero.kicker": "Для OBS Studio",
      "hero.title": "Widget Stats",
      "hero.lead": "Победы, поражения и ранг на стриме — без мерцания. Жми сколько угодно, потом поставь по гайду ниже.",
      "hero.rankUp": "Ранг +",
      "hero.rankDown": "Ранг −",
      "hero.reset": "Сброс W/L",
      "hero.hint": "Настоящий оверлей, не видео. Счётчик не сбрасывается сам.",
      "look.kicker": "Внешний вид",
      "look.title": "Скины и режимы",
      "look.lead": "Пресет меняет живой виджет выше. Счёт W/L при смене скина не трогаем.",
      "mode.ow": "Overwatch",
      "mode.owRoles": "OW · роли",
      "mode.apex": "Apex Legends",
      "install.kicker": "Установка",
      "install.title": "Поставить самому",
      "install.lead": "Где нужен экран — скрин или короткое видео. Lua — сразу после скачивания, он сам поднимет сервер.",
      "i1.t": "Скачай релиз",
      "i1.d": "Скачай widget-stats.exe и widget_control.lua из Releases — оба файла, в одну папку.",
      "i1.a": "Открыть Releases →",
      "i2.t": "SmartScreen — это нормально",
      "i2.d": "При первом запуске exe Windows может предупредить. Подробнее → Выполнить в любом случае. Код открыт в репозитории.",
      "i3.t": "Открой админку",
      "i3.d": "Иконка в трее → Open Admin (или http://127.0.0.1:19123/admin/). Сервер уже должен быть запущен скриптом.",
      "i4.t": "Включи WebSocket в OBS",
      "i4.d": "Tools → WebSocket Server Settings → Enable, порт 4455.",
      "i5.t": "Поставь виджет на сцену",
      "i5.d": "В админке: Подключить OBS → сцена → Поставить виджет на сцену.",
      "i6.t": "Сразу добавь Lua-скрипт",
      "i6.d": "Первым делом: OBS → Tools → Scripts → «+» → widget_control.lua (рядом с exe). Скрипт сам запустит сервер при старте OBS и остановит при закрытии.",
      "clip.need": "Добавь файл в assets/",
      "clip.expand": "Крупнее",
      "more.kicker": "Ещё",
      "more.title": "Stream Deck и API",
      "more.lead": "HTTP: /api/win · /api/loss · /api/rank/up · /api/rank/down · /api/reset.",
      "more.icons": "Иконки кнопок",
      "foot.cta": "Скачать и поставить",
      "foot.note": "Страницу открывай через локальный сервер из корня репо.",
    },
    en: {
      "bar.sub": "Live demo · no install needed",
      "bar.download": "Download",
      "hero.kicker": "For OBS Studio",
      "hero.title": "Widget Stats",
      "hero.lead": "Wins, losses, and rank on stream — no flicker. Click as much as you want, then follow the setup guide.",
      "hero.rankUp": "Rank +",
      "hero.rankDown": "Rank −",
      "hero.reset": "Reset W/L",
      "hero.hint": "Real overlay, not a video. The counter does not reset by itself.",
      "look.kicker": "Look",
      "look.title": "Skins & modes",
      "look.lead": "A preset updates the live widget above. W/L is kept when you change skins.",
      "mode.ow": "Overwatch",
      "mode.owRoles": "OW · roles",
      "mode.apex": "Apex Legends",
      "install.kicker": "Setup",
      "install.title": "Install it yourself",
      "install.lead": "Screenshot or short clip where the screen matters. Add Lua right after download — it starts the server for you.",
      "i1.t": "Download the release",
      "i1.d": "Download both widget-stats.exe and widget_control.lua from Releases into the same folder.",
      "i1.a": "Open Releases →",
      "i2.t": "SmartScreen is expected",
      "i2.d": "Windows may warn on the first exe launch. More info → Run anyway. Source is in the repo.",
      "i3.t": "Open the admin",
      "i3.d": "Tray icon → Open Admin (or http://127.0.0.1:19123/admin/). The script should already have started the server.",
      "i4.t": "Enable OBS WebSocket",
      "i4.d": "Tools → WebSocket Server Settings → Enable, port 4455.",
      "i5.t": "Place widget on a scene",
      "i5.d": "In admin: Connect OBS → pick scene → Place widget on scene.",
      "i6.t": "Add the Lua script first",
      "i6.d": "Do this first: OBS → Tools → Scripts → “+” → widget_control.lua (next to the exe). It starts the server when OBS opens and stops it when OBS closes.",
      "clip.need": "Add a file to assets/",
      "clip.expand": "Larger",
      "more.kicker": "More",
      "more.title": "Stream Deck & API",
      "more.lead": "HTTP: /api/win · /api/loss · /api/rank/up · /api/rank/down · /api/reset.",
      "more.icons": "Button icons",
      "foot.cta": "Download & install",
      "foot.note": "Serve this page from the repo root.",
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
      settings: { ...settings, fontSize: Math.max(22, settings.fontSize || 18) },
    });
  }

  function pushState(extra = {}) {
    demoState = { ...demoState, ...extra };
    post({ type: "widget-preview-state", view: { ...demoState }, state: { ...demoState } });
  }

  /** Free play: do NOT re-push state before demo (that was resetting W/L every click). */
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
        // Keep current W/L — only refresh look.
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

  function probe(url) {
    return fetch(url, { method: "HEAD" })
      .then((res) => res.ok)
      .catch(() => false);
  }

  async function resolveMedia(files) {
    for (const file of files) {
      const url = `assets/${file}`;
      if (await probe(url)) return { file, url };
    }
    return null;
  }

  function openLightbox(el) {
    lightboxFrame.innerHTML = "";
    const clone = el.cloneNode(true);
    clone.removeAttribute("style");
    clone.style.maxWidth = "100%";
    clone.style.maxHeight = "90vh";
    clone.style.cursor = "default";
    if (clone.tagName === "VIDEO") {
      clone.controls = true;
      clone.muted = true;
      clone.loop = true;
      clone.autoplay = true;
    }
    lightboxFrame.appendChild(clone);
    lightbox.hidden = false;
    document.body.style.overflow = "hidden";
  }

  function closeLightbox() {
    lightbox.hidden = true;
    lightboxFrame.innerHTML = "";
    document.body.style.overflow = "";
  }

  async function mountGuideMedia() {
    const steps = [...document.querySelectorAll(".guide-step[data-media-files]")];
    for (const step of steps) {
      const box = step.querySelector(".guide-media");
      if (!box) continue;
      const files = step
        .getAttribute("data-media-files")
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);

      box.className = "guide-media is-waiting";
      box.innerHTML = `<span data-ph>${t("clip.need")}</span><code>${files[0] || ""}</code>`;

      const found = await resolveMedia(files);
      if (!found) continue;

      box.className = "guide-media is-ready";
      box.innerHTML = "";
      let media;
      if (isVideo(found.file)) {
        media = document.createElement("video");
        media.src = found.url;
        media.muted = true;
        media.loop = true;
        media.playsInline = true;
        media.controls = true;
        media.preload = "metadata";
        media.addEventListener("mouseenter", () => media.play().catch(() => {}));
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
    });
  });

  lightboxClose.addEventListener("click", closeLightbox);
  lightbox.addEventListener("click", (e) => {
    if (e.target === lightbox) closeLightbox();
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !lightbox.hidden) closeLightbox();
  });

  frame.addEventListener("load", () => bootPreview());

  applyI18n();
  renderSkins();
  mountGuideMedia();

  if (frame.contentDocument?.readyState === "complete") bootPreview();
})();

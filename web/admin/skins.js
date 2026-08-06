window.WIDGET_SKINS = [
  /* —— classic (2) —— */
  {
    id: 'default',
    name: 'Default Neon',
    desc: 'Classic · зелёный / красный',
    preview: { bg: 'rgba(0,0,0,0.72)', wins: '#39ff14', losses: '#ff3131' },
    settings: {
      bgColor: 'rgba(0,0,0,0.72)', bgImage: '', winsColor: '#39ff14', lossesColor: '#ff3131',
      rankTextColor: '#ffffff', iconColor: '#ffffff', separatorColor: 'rgba(255,255,255,0.55)',
      font: "'Rajdhani', sans-serif", fontSize: 18, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'classic',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 650, emptyDurationMs: 1600,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'default',
    },
  },
  {
    id: 'mono-ink',
    name: 'Mono Contrast',
    desc: 'Classic · чёрно-белый',
    preview: { bg: 'rgba(18,18,18,0.94)', wins: '#f5f5f5', losses: '#9e9e9e' },
    settings: {
      bgColor: 'rgba(18,18,18,0.94)', bgImage: '', winsColor: '#f5f5f5', lossesColor: '#9e9e9e',
      rankTextColor: '#e8e8e8', iconColor: '#ffffff', separatorColor: 'rgba(255,255,255,0.4)',
      font: "'Space Grotesk', sans-serif", fontSize: 17, appearEffect: 'fade', fillStyle: 'solid', rankFx: 'classic',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 420, emptyDurationMs: 1400,
      emptyEffect: 'fade', vesselWave: false, idlePulse: true, skinId: 'mono-ink',
    },
  },

  /* —— blaze (2) —— */
  {
    id: 'blood-moon',
    name: 'Blood Moon',
    desc: 'Blaze · тёмно-красный огонь',
    preview: { bg: 'rgba(20,4,8,0.9)', wins: '#ffb703', losses: '#e63946' },
    settings: {
      bgColor: 'rgba(20,4,8,0.9)', bgImage: '', winsColor: '#ffb703', lossesColor: '#e63946',
      rankTextColor: '#ffd6a5', iconColor: '#ffe8c8', separatorColor: 'rgba(255,200,150,0.45)',
      font: "'Teko', sans-serif", fontSize: 20, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'blaze',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 520, emptyDurationMs: 1700,
      emptyEffect: 'pour', vesselWave: true, idlePulse: true, skinId: 'blood-moon',
    },
  },
  {
    id: 'solar-flare',
    name: 'Solar Flare',
    desc: 'Blaze · солнечная вспышка',
    preview: { bg: 'rgba(40,24,4,0.92)', wins: '#facc15', losses: '#f43f5e' },
    settings: {
      bgColor: 'rgba(40,24,4,0.92)', bgImage: '', winsColor: '#facc15', lossesColor: '#f43f5e',
      rankTextColor: '#fef9c3', iconColor: '#fefce8', separatorColor: 'rgba(250,204,21,0.4)',
      font: "'Russo One', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'blaze',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 460, emptyDurationMs: 1600,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'solar-flare',
    },
  },

  /* —— frost (2) —— */
  {
    id: 'ocean-deep',
    name: 'Ocean Deep',
    desc: 'Frost · глубокий синий',
    preview: { bg: 'rgba(4,18,32,0.9)', wins: '#48cae4', losses: '#ff7b54' },
    settings: {
      bgColor: 'rgba(4,18,32,0.9)', bgImage: '', winsColor: '#48cae4', lossesColor: '#ff7b54',
      rankTextColor: '#caf0f8', iconColor: '#e8f9ff', separatorColor: 'rgba(180,230,255,0.45)',
      font: "'Exo 2', sans-serif", fontSize: 17, appearEffect: 'fade', fillStyle: 'liquid', rankFx: 'frost',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 700, emptyDurationMs: 1900,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'ocean-deep',
    },
  },
  {
    id: 'nordic-frost',
    name: 'Nordic Frost',
    desc: 'Frost · светлый север',
    preview: { bg: 'rgba(210,225,235,0.96)', wins: '#0ea5e9', losses: '#64748b' },
    settings: {
      bgColor: 'rgba(210,225,235,0.96)', bgImage: '', winsColor: '#0ea5e9', lossesColor: '#64748b',
      rankTextColor: '#0f172a', iconColor: '#1e293b', separatorColor: 'rgba(15,23,42,0.35)',
      font: "'Inter', sans-serif", fontSize: 16, appearEffect: 'slide', fillStyle: 'solid', rankFx: 'frost',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 520, emptyDurationMs: 1500,
      emptyEffect: 'drain', vesselWave: false, idlePulse: true, skinId: 'nordic-frost',
    },
  },

  /* —— neon (2) —— */
  {
    id: 'cyber-cyan',
    name: 'Cyber Cyan',
    desc: 'Neon · кибер-глитч',
    preview: { bg: 'rgba(6,18,28,0.85)', wins: '#00e5ff', losses: '#ff4d6d' },
    settings: {
      bgColor: 'rgba(6,18,28,0.85)', bgImage: '', winsColor: '#00e5ff', lossesColor: '#ff4d6d',
      rankTextColor: '#d7f7ff', iconColor: '#e8fbff', separatorColor: 'rgba(180,230,255,0.5)',
      font: "'Orbitron', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'neon',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 420, emptyDurationMs: 1500,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'cyber-cyan',
    },
  },
  {
    id: 'neon-alley',
    name: 'Neon Alley',
    desc: 'Neon · ночной переулок',
    preview: { bg: 'rgba(10,6,24,0.94)', wins: '#4ade80', losses: '#f472b6' },
    settings: {
      bgColor: 'rgba(10,6,24,0.94)', bgImage: '', winsColor: '#4ade80', lossesColor: '#f472b6',
      rankTextColor: '#e9d5ff', iconColor: '#faf5ff', separatorColor: 'rgba(192,132,252,0.4)',
      font: "'Orbitron', sans-serif", fontSize: 15, appearEffect: 'bounce', fillStyle: 'glow', rankFx: 'neon',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 400, emptyDurationMs: 1400,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'neon-alley',
    },
  },
  {
    id: 'glitch-district',
    name: 'Glitch District',
    desc: 'Neon · киберпанк-глитч',
    preview: { bg: 'rgba(12,4,22,0.95)', wins: '#ff2bd6', losses: '#ffe600' },
    settings: {
      bgColor: 'rgba(12,4,22,0.95)', bgImage: '', winsColor: '#ff2bd6', lossesColor: '#ffe600',
      rankTextColor: '#f5d0fe', iconColor: '#fdf4ff', separatorColor: 'rgba(255,43,214,0.45)',
      font: "'Orbitron', sans-serif", fontSize: 15, appearEffect: 'bounce', fillStyle: 'glow', rankFx: 'neon',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 380, emptyDurationMs: 1350,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'glitch-district',
    },
  },
  {
    id: 'data-breach',
    name: 'Data Breach',
    desc: 'Neon · взлом / матрица',
    preview: { bg: 'rgba(2,8,6,0.96)', wins: '#39ff14', losses: '#ff1744' },
    settings: {
      bgColor: 'rgba(2,8,6,0.96)', bgImage: '', winsColor: '#39ff14', lossesColor: '#ff1744',
      rankTextColor: '#b8ff9f', iconColor: '#e8ffe0', separatorColor: 'rgba(57,255,20,0.4)',
      font: "'Chakra Petch', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'neon',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 360, emptyDurationMs: 1300,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'data-breach',
    },
  },

  /* —— divine (2) —— */
  {
    id: 'gold-rush',
    name: 'Gold Rush',
    desc: 'Divine · чемпионское золото',
    preview: { bg: 'rgba(28,18,6,0.88)', wins: '#ffd166', losses: '#ef476f' },
    settings: {
      bgColor: 'rgba(28,18,6,0.88)', bgImage: '', winsColor: '#ffd166', lossesColor: '#ef476f',
      rankTextColor: '#ffe8a3', iconColor: '#fff1c8', separatorColor: 'rgba(255,220,150,0.5)',
      font: "'Oswald', sans-serif", fontSize: 18, appearEffect: 'bounce', fillStyle: 'liquid', rankFx: 'divine',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 560, emptyDurationMs: 1800,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'gold-rush',
    },
  },
  {
    id: 'ivory-court',
    name: 'Ivory Court',
    desc: 'Divine · светлый двор',
    preview: { bg: 'rgba(250,248,240,0.96)', wins: '#ca8a04', losses: '#9f1239' },
    settings: {
      bgColor: 'rgba(250,248,240,0.96)', bgImage: '', winsColor: '#ca8a04', lossesColor: '#9f1239',
      rankTextColor: '#44403c', iconColor: '#292524', separatorColor: 'rgba(68,64,60,0.35)',
      font: "'Cinzel', serif", fontSize: 15, appearEffect: 'fade', fillStyle: 'solid', rankFx: 'divine',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 540, emptyDurationMs: 1600,
      emptyEffect: 'fade', vesselWave: false, idlePulse: true, skinId: 'ivory-court',
    },
  },

  /* —— melt (2) —— */
  {
    id: 'magma-core',
    name: 'Magma Core',
    desc: 'Melt · раскалённый штамп',
    preview: { bg: 'rgba(28,8,4,0.94)', wins: '#f97316', losses: '#dc2626' },
    settings: {
      bgColor: 'rgba(28,8,4,0.94)', bgImage: '', winsColor: '#f97316', lossesColor: '#dc2626',
      rankTextColor: '#ffedd5', iconColor: '#fff7ed', separatorColor: 'rgba(251,146,60,0.4)',
      font: "'Oswald', sans-serif", fontSize: 17, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'melt',
      animDirection: 'bottom', fillLimit: 10, fillDurationMs: 480, emptyDurationMs: 1700,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'magma-core',
    },
  },
  {
    id: 'forge-amber',
    name: 'Forge Amber',
    desc: 'Melt · кузня',
    preview: { bg: 'rgba(32,18,8,0.93)', wins: '#fb923c', losses: '#b91c1c' },
    settings: {
      bgColor: 'rgba(32,18,8,0.93)', bgImage: '', winsColor: '#fb923c', lossesColor: '#b91c1c',
      rankTextColor: '#fed7aa', iconColor: '#ffedd5', separatorColor: 'rgba(251,146,60,0.4)',
      font: "'Russo One', sans-serif", fontSize: 16, appearEffect: 'bounce', fillStyle: 'liquid', rankFx: 'melt',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 520, emptyDurationMs: 1800,
      emptyEffect: 'pour', vesselWave: true, idlePulse: true, skinId: 'forge-amber',
    },
  },

  /* —— arcane (2) —— */
  {
    id: 'violet-haze',
    name: 'Violet Haze',
    desc: 'Arcane · фиолетовая магия',
    preview: { bg: 'rgba(16,8,28,0.9)', wins: '#c77dff', losses: '#ff6b6b' },
    settings: {
      bgColor: 'rgba(16,8,28,0.9)', bgImage: '', winsColor: '#c77dff', lossesColor: '#ff6b6b',
      rankTextColor: '#e0aaff', iconColor: '#f3e8ff', separatorColor: 'rgba(220,180,255,0.45)',
      font: "'Audiowide', system-ui", fontSize: 15, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'arcane',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 460, emptyDurationMs: 1600,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'violet-haze',
    },
  },
  {
    id: 'deep-forest',
    name: 'Deep Forest',
    desc: 'Arcane · колдовской лес',
    preview: { bg: 'rgba(8,20,12,0.93)', wins: '#4ade80', losses: '#a3e635' },
    settings: {
      bgColor: 'rgba(8,20,12,0.93)', bgImage: '', winsColor: '#4ade80', lossesColor: '#a3e635',
      rankTextColor: '#dcfce7', iconColor: '#f0fdf4', separatorColor: 'rgba(74,222,128,0.35)',
      font: "'Poppins', sans-serif", fontSize: 16, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'arcane',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 560, emptyDurationMs: 1700,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'deep-forest',
    },
  },

  /* —— toxic (2) —— */
  {
    id: 'toxic-spill',
    name: 'Toxic Spill',
    desc: 'Toxic · ядовитый лайм',
    preview: { bg: 'rgba(10,14,6,0.92)', wins: '#b6ff00', losses: '#ff0080' },
    settings: {
      bgColor: 'rgba(10,14,6,0.92)', bgImage: '', winsColor: '#b6ff00', lossesColor: '#ff0080',
      rankTextColor: '#eaffb0', iconColor: '#f5ffd0', separatorColor: 'rgba(220,255,160,0.4)',
      font: "'Chakra Petch', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'bubble', rankFx: 'toxic',
      animDirection: 'right', fillLimit: 8, fillDurationMs: 340, emptyDurationMs: 1300,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'toxic-spill',
    },
  },
  {
    id: 'jungle-heat',
    name: 'Jungle Heat',
    desc: 'Toxic · тропический яд',
    preview: { bg: 'rgba(12,28,12,0.9)', wins: '#84cc16', losses: '#ef4444' },
    settings: {
      bgColor: 'rgba(12,28,12,0.9)', bgImage: '', winsColor: '#84cc16', lossesColor: '#ef4444',
      rankTextColor: '#ecfccb', iconColor: '#f7fee7', separatorColor: 'rgba(163,230,53,0.4)',
      font: "'Teko', sans-serif", fontSize: 19, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'toxic',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 500, emptyDurationMs: 1600,
      emptyEffect: 'pour', vesselWave: true, idlePulse: true, skinId: 'jungle-heat',
    },
  },

  /* —— paper (2) —— */
  {
    id: 'amber-library',
    name: 'Amber Library',
    desc: 'Paper · переворот книги',
    preview: { bg: 'rgba(42,28,14,0.92)', wins: '#f59e0b', losses: '#b45309' },
    settings: {
      bgColor: 'rgba(42,28,14,0.92)', bgImage: '', winsColor: '#f59e0b', lossesColor: '#b45309',
      rankTextColor: '#fef3c7', iconColor: '#fffbeb', separatorColor: 'rgba(245,158,11,0.4)',
      font: "'Cinzel', serif", fontSize: 15, appearEffect: 'fade', fillStyle: 'liquid', rankFx: 'paper',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 580, emptyDurationMs: 1700,
      emptyEffect: 'pour', vesselWave: true, idlePulse: true, skinId: 'amber-library',
    },
  },
  {
    id: 'paper-light',
    name: 'Paper Light',
    desc: 'Paper · светлая страница',
    preview: { bg: 'rgba(250,248,243,0.96)', wins: '#2d6a4f', losses: '#9b2226' },
    settings: {
      bgColor: 'rgba(250,248,243,0.96)', bgImage: '', winsColor: '#2d6a4f', lossesColor: '#9b2226',
      rankTextColor: '#1b1b1b', iconColor: '#222222', separatorColor: 'rgba(30,30,30,0.4)',
      font: "'Nunito', sans-serif", fontSize: 17, appearEffect: 'bounce', fillStyle: 'solid', rankFx: 'paper',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 520, emptyDurationMs: 1500,
      emptyEffect: 'drain', vesselWave: false, idlePulse: false, skinId: 'paper-light',
    },
  },

  /* —— vortex (2) —— */
  {
    id: 'midnight-void',
    name: 'Midnight Void',
    desc: 'Vortex · космический вихрь',
    preview: { bg: 'rgba(4,6,18,0.94)', wins: '#818cf8', losses: '#f472b6' },
    settings: {
      bgColor: 'rgba(4,6,18,0.94)', bgImage: '', winsColor: '#818cf8', lossesColor: '#f472b6',
      rankTextColor: '#e0e7ff', iconColor: '#f8fafc', separatorColor: 'rgba(165,180,252,0.4)',
      font: "'Orbitron', sans-serif", fontSize: 15, appearEffect: 'fade', fillStyle: 'glow', rankFx: 'vortex',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 520, emptyDurationMs: 1600,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'midnight-void',
    },
  },
  {
    id: 'spiral-teal',
    name: 'Spiral Teal',
    desc: 'Vortex · бирюзовый спираль',
    preview: { bg: 'rgba(6,24,28,0.92)', wins: '#2dd4bf', losses: '#f43f5e' },
    settings: {
      bgColor: 'rgba(6,24,28,0.92)', bgImage: '', winsColor: '#2dd4bf', lossesColor: '#f43f5e',
      rankTextColor: '#ccfbf1', iconColor: '#f0fdfa', separatorColor: 'rgba(45,212,191,0.4)',
      font: "'Exo 2', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'liquid', rankFx: 'vortex',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 480, emptyDurationMs: 1500,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'spiral-teal',
    },
  },

  /* —— shatter (2) —— */
  {
    id: 'crystal-break',
    name: 'Crystal Break',
    desc: 'Shatter · хрусталь',
    preview: { bg: 'rgba(230,240,250,0.95)', wins: '#0284c7', losses: '#be123c' },
    settings: {
      bgColor: 'rgba(230,240,250,0.95)', bgImage: '', winsColor: '#0284c7', lossesColor: '#be123c',
      rankTextColor: '#0c4a6e', iconColor: '#164e63', separatorColor: 'rgba(12,74,110,0.35)',
      font: "'Montserrat', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'liquid', rankFx: 'shatter',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 500, emptyDurationMs: 1500,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'crystal-break',
    },
  },
  {
    id: 'obsidian-glass',
    name: 'Obsidian Glass',
    desc: 'Shatter · тёмное стекло',
    preview: { bg: 'rgba(12,14,22,0.94)', wins: '#93c5fd', losses: '#fb7185' },
    settings: {
      bgColor: 'rgba(12,14,22,0.94)', bgImage: '', winsColor: '#93c5fd', lossesColor: '#fb7185',
      rankTextColor: '#dbeafe', iconColor: '#eff6ff', separatorColor: 'rgba(147,197,253,0.4)',
      font: "'Space Grotesk', sans-serif", fontSize: 16, appearEffect: 'bounce', fillStyle: 'glow', rankFx: 'shatter',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 440, emptyDurationMs: 1400,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'obsidian-glass',
    },
  },

  /* —— smoke (2) —— */
  {
    id: 'ash-veil',
    name: 'Ash Veil',
    desc: 'Smoke · пепельная дымка',
    preview: { bg: 'rgba(28,28,30,0.92)', wins: '#a8a29e', losses: '#78716c' },
    settings: {
      bgColor: 'rgba(28,28,30,0.92)', bgImage: '', winsColor: '#a8a29e', lossesColor: '#78716c',
      rankTextColor: '#e7e5e4', iconColor: '#fafaf9', separatorColor: 'rgba(214,211,209,0.35)',
      font: "'Space Grotesk', sans-serif", fontSize: 16, appearEffect: 'fade', fillStyle: 'solid', rankFx: 'smoke',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 600, emptyDurationMs: 1700,
      emptyEffect: 'fade', vesselWave: false, idlePulse: true, skinId: 'ash-veil',
    },
  },
  {
    id: 'ember-smoke',
    name: 'Ember Smoke',
    desc: 'Smoke · дым с углями',
    preview: { bg: 'rgba(22,14,12,0.93)', wins: '#d6d3d1', losses: '#ea580c' },
    settings: {
      bgColor: 'rgba(22,14,12,0.93)', bgImage: '', winsColor: '#d6d3d1', lossesColor: '#ea580c',
      rankTextColor: '#f5f5f4', iconColor: '#fafaf9', separatorColor: 'rgba(214,211,209,0.35)',
      font: "'Oswald', sans-serif", fontSize: 17, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'smoke',
      animDirection: 'bottom', fillLimit: 10, fillDurationMs: 540, emptyDurationMs: 1600,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'ember-smoke',
    },
  },

  /* —— lightning (2) —— */
  {
    id: 'storm-call',
    name: 'Storm Call',
    desc: 'Lightning · гроза',
    preview: { bg: 'rgba(8,16,32,0.93)', wins: '#38bdf8', losses: '#fbbf24' },
    settings: {
      bgColor: 'rgba(8,16,32,0.93)', bgImage: '', winsColor: '#38bdf8', lossesColor: '#fbbf24',
      rankTextColor: '#e0f2fe', iconColor: '#f0f9ff', separatorColor: 'rgba(125,211,252,0.4)',
      font: "'Exo 2', sans-serif", fontSize: 16, appearEffect: 'bounce', fillStyle: 'glow', rankFx: 'lightning',
      animDirection: 'right', fillLimit: 8, fillDurationMs: 380, emptyDurationMs: 1400,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'storm-call',
    },
  },
  {
    id: 'voltage-night',
    name: 'Voltage Night',
    desc: 'Lightning · ночной разряд',
    preview: { bg: 'rgba(10,10,28,0.94)', wins: '#67e8f9', losses: '#e879f9' },
    settings: {
      bgColor: 'rgba(10,10,28,0.94)', bgImage: '', winsColor: '#67e8f9', lossesColor: '#e879f9',
      rankTextColor: '#cffafe', iconColor: '#ecfeff', separatorColor: 'rgba(103,232,249,0.4)',
      font: "'Orbitron', sans-serif", fontSize: 15, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'lightning',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 360, emptyDurationMs: 1300,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'voltage-night',
    },
  },

  /* —— ripple (2) —— */
  {
    id: 'tide-pool',
    name: 'Tide Pool',
    desc: 'Ripple · морская рябь',
    preview: { bg: 'rgba(6,40,52,0.9)', wins: '#2dd4bf', losses: '#fb7185' },
    settings: {
      bgColor: 'rgba(6,40,52,0.9)', bgImage: '', winsColor: '#2dd4bf', lossesColor: '#fb7185',
      rankTextColor: '#ccfbf1', iconColor: '#f0fdfa', separatorColor: 'rgba(94,234,212,0.4)',
      font: "'Nunito', sans-serif", fontSize: 17, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'ripple',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 680, emptyDurationMs: 1800,
      emptyEffect: 'pour', vesselWave: true, idlePulse: true, skinId: 'tide-pool',
    },
  },
  {
    id: 'aqua-glass',
    name: 'Aqua Glass',
    desc: 'Ripple · светлая вода',
    preview: { bg: 'rgba(220,242,248,0.95)', wins: '#0891b2', losses: '#e11d48' },
    settings: {
      bgColor: 'rgba(220,242,248,0.95)', bgImage: '', winsColor: '#0891b2', lossesColor: '#e11d48',
      rankTextColor: '#164e63', iconColor: '#155e75', separatorColor: 'rgba(22,78,99,0.35)',
      font: "'Poppins', sans-serif", fontSize: 16, appearEffect: 'fade', fillStyle: 'liquid', rankFx: 'ripple',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 600, emptyDurationMs: 1600,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'aqua-glass',
    },
  },

  /* —— hologram (2) —— */
  {
    id: 'holo-deck',
    name: 'Holo Deck',
    desc: 'Hologram · UI-скан',
    preview: { bg: 'rgba(2,20,28,0.92)', wins: '#22d3ee', losses: '#e879f9' },
    settings: {
      bgColor: 'rgba(2,20,28,0.92)', bgImage: '', winsColor: '#22d3ee', lossesColor: '#e879f9',
      rankTextColor: '#a5f3fc', iconColor: '#ecfeff', separatorColor: 'rgba(34,211,238,0.4)',
      font: "'Audiowide', system-ui", fontSize: 14, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'hologram',
      animDirection: 'bottom', fillLimit: 10, fillDurationMs: 440, emptyDurationMs: 1500,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'holo-deck',
    },
  },
  {
    id: 'cyan-project',
    name: 'Cyan Project',
    desc: 'Hologram · проекция',
    preview: { bg: 'rgba(4,28,36,0.93)', wins: '#67e8f9', losses: '#f472b6' },
    settings: {
      bgColor: 'rgba(4,28,36,0.93)', bgImage: '', winsColor: '#67e8f9', lossesColor: '#f472b6',
      rankTextColor: '#cffafe', iconColor: '#f0fdfa', separatorColor: 'rgba(103,232,249,0.4)',
      font: "'Chakra Petch', sans-serif", fontSize: 15, appearEffect: 'slide', fillStyle: 'glow', rankFx: 'hologram',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 420, emptyDurationMs: 1400,
      emptyEffect: 'fade', vesselWave: true, idlePulse: true, skinId: 'cyan-project',
    },
  },
  {
    id: 'pink-scanline',
    name: 'Pink Scanline',
    desc: 'Hologram · розовавая голограмма',
    preview: { bg: 'rgba(18,4,28,0.94)', wins: '#ff4ecd', losses: '#00f0ff' },
    settings: {
      bgColor: 'rgba(18,4,28,0.94)', bgImage: '', winsColor: '#ff4ecd', lossesColor: '#00f0ff',
      rankTextColor: '#fbcfe8', iconColor: '#fdf2f8', separatorColor: 'rgba(255,78,205,0.42)',
      font: "'Audiowide', system-ui", fontSize: 14, appearEffect: 'fade', fillStyle: 'glow', rankFx: 'hologram',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 430, emptyDurationMs: 1450,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'pink-scanline',
    },
  },
  {
    id: 'grid-runner',
    name: 'Grid Runner',
    desc: 'Hologram · сетка / wireframe',
    preview: { bg: 'rgba(4,6,14,0.95)', wins: '#fbbf24', losses: '#38bdf8' },
    settings: {
      bgColor: 'rgba(4,6,14,0.95)', bgImage: '', winsColor: '#fbbf24', lossesColor: '#38bdf8',
      rankTextColor: '#fef3c7', iconColor: '#fffbeb', separatorColor: 'rgba(56,189,248,0.4)',
      font: "'Orbitron', sans-serif", fontSize: 15, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'hologram',
      animDirection: 'bottom', fillLimit: 10, fillDurationMs: 410, emptyDurationMs: 1400,
      emptyEffect: 'fade', vesselWave: true, idlePulse: true, skinId: 'grid-runner',
    },
  },

  /* —— pixel (2) —— */
  {
    id: '8bit-arena',
    name: '8-Bit Arena',
    desc: 'Pixel · аркада',
    preview: { bg: 'rgba(20,12,40,0.94)', wins: '#4ade80', losses: '#f87171' },
    settings: {
      bgColor: 'rgba(20,12,40,0.94)', bgImage: '', winsColor: '#4ade80', lossesColor: '#f87171',
      rankTextColor: '#fef08a', iconColor: '#ffffff', separatorColor: 'rgba(250,204,21,0.4)',
      font: "'Press Start 2P', system-ui", fontSize: 11, appearEffect: 'bounce', fillStyle: 'bubble', rankFx: 'pixel',
      animDirection: 'right', fillLimit: 8, fillDurationMs: 300, emptyDurationMs: 1200,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: '8bit-arena',
    },
  },
  {
    id: 'retro-cabinet',
    name: 'Retro Cabinet',
    desc: 'Pixel · кабинет 80-х',
    preview: { bg: 'rgba(12,8,24,0.94)', wins: '#7cff6b', losses: '#ff6bcb' },
    settings: {
      bgColor: 'rgba(12,8,24,0.94)', bgImage: '', winsColor: '#7cff6b', lossesColor: '#ff6bcb',
      rankTextColor: '#f8f7ff', iconColor: '#ffffff', separatorColor: 'rgba(255,255,255,0.45)',
      font: "'Press Start 2P', system-ui", fontSize: 12, appearEffect: 'bounce', fillStyle: 'bubble', rankFx: 'pixel',
      animDirection: 'bottom', fillLimit: 10, fillDurationMs: 320, emptyDurationMs: 1300,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'retro-cabinet',
    },
  },

  /* —— chrome (2) —— */
  {
    id: 'liquid-steel',
    name: 'Liquid Steel',
    desc: 'Chrome · жидкий металл',
    preview: { bg: 'rgba(24,24,27,0.94)', wins: '#d4d4d8', losses: '#a1a1aa' },
    settings: {
      bgColor: 'rgba(24,24,27,0.94)', bgImage: '', winsColor: '#d4d4d8', lossesColor: '#a1a1aa',
      rankTextColor: '#fafafa', iconColor: '#ffffff', separatorColor: 'rgba(212,212,216,0.4)',
      font: "'Oswald', sans-serif", fontSize: 17, appearEffect: 'slide', fillStyle: 'solid', rankFx: 'chrome',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 480, emptyDurationMs: 1500,
      emptyEffect: 'drain', vesselWave: false, idlePulse: true, skinId: 'liquid-steel',
    },
  },
  {
    id: 'mercury-run',
    name: 'Mercury Run',
    desc: 'Chrome · ртутный блеск',
    preview: { bg: 'rgba(30,32,40,0.94)', wins: '#e4e4e7', losses: '#71717a' },
    settings: {
      bgColor: 'rgba(30,32,40,0.94)', bgImage: '', winsColor: '#e4e4e7', lossesColor: '#71717a',
      rankTextColor: '#f4f4f5', iconColor: '#fafafa', separatorColor: 'rgba(228,228,231,0.4)',
      font: "'Space Grotesk', sans-serif", fontSize: 16, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'chrome',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 450, emptyDurationMs: 1400,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'mercury-run',
    },
  },

  /* —— phantom (2) —— */
  {
    id: 'ghost-lantern',
    name: 'Ghost Lantern',
    desc: 'Phantom · призрачный фонарь',
    preview: { bg: 'rgba(18,12,32,0.92)', wins: '#c4b5fd', losses: '#67e8f9' },
    settings: {
      bgColor: 'rgba(18,12,32,0.92)', bgImage: '', winsColor: '#c4b5fd', lossesColor: '#67e8f9',
      rankTextColor: '#ede9fe', iconColor: '#f5f3ff', separatorColor: 'rgba(196,181,253,0.4)',
      font: "'Cinzel', serif", fontSize: 15, appearEffect: 'fade', fillStyle: 'glow', rankFx: 'phantom',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 560, emptyDurationMs: 1700,
      emptyEffect: 'fade', vesselWave: true, idlePulse: true, skinId: 'ghost-lantern',
    },
  },
  {
    id: 'wraith-mist',
    name: 'Wraith Mist',
    desc: 'Phantom · туманный дух',
    preview: { bg: 'rgba(14,18,28,0.93)', wins: '#a5b4fc', losses: '#fda4af' },
    settings: {
      bgColor: 'rgba(14,18,28,0.93)', bgImage: '', winsColor: '#a5b4fc', lossesColor: '#fda4af',
      rankTextColor: '#e0e7ff', iconColor: '#eef2ff', separatorColor: 'rgba(165,180,252,0.4)',
      font: "'Josefin Sans', sans-serif", fontSize: 16, appearEffect: 'fade', fillStyle: 'liquid', rankFx: 'phantom',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 580, emptyDurationMs: 1700,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'wraith-mist',
    },
  },

  /* —— sonic (2) —— */
  {
    id: 'bass-drop',
    name: 'Bass Drop',
    desc: 'Sonic · ударная волна',
    preview: { bg: 'rgba(12,8,20,0.93)', wins: '#f0abfc', losses: '#fb923c' },
    settings: {
      bgColor: 'rgba(12,8,20,0.93)', bgImage: '', winsColor: '#f0abfc', lossesColor: '#fb923c',
      rankTextColor: '#fae8ff', iconColor: '#ffffff', separatorColor: 'rgba(240,171,252,0.4)',
      font: "'Russo One', sans-serif", fontSize: 16, appearEffect: 'bounce', fillStyle: 'bubble', rankFx: 'sonic',
      animDirection: 'bottom', fillLimit: 10, fillDurationMs: 360, emptyDurationMs: 1400,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'bass-drop',
    },
  },
  {
    id: 'carbon-race',
    name: 'Carbon Race',
    desc: 'Sonic · карбон и скорость',
    preview: { bg: 'rgba(16,16,16,0.95)', wins: '#22d3ee', losses: '#f97316' },
    settings: {
      bgColor: 'rgba(16,16,16,0.95)', bgImage: '', winsColor: '#22d3ee', lossesColor: '#f97316',
      rankTextColor: '#e2e8f0', iconColor: '#f8fafc', separatorColor: 'rgba(34,211,238,0.4)',
      font: "'Chakra Petch', sans-serif", fontSize: 16, appearEffect: 'bounce', fillStyle: 'glow', rankFx: 'sonic',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 360, emptyDurationMs: 1300,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'carbon-race',
    },
  },

  /* —— bloom (2) —— */
  {
    id: 'sakura-dusk',
    name: 'Sakura Dusk',
    desc: 'Bloom · лепестки сакуры',
    preview: { bg: 'rgba(40,20,36,0.9)', wins: '#fb7185', losses: '#c084fc' },
    settings: {
      bgColor: 'rgba(40,20,36,0.9)', bgImage: '', winsColor: '#fb7185', lossesColor: '#c084fc',
      rankTextColor: '#ffe4e6', iconColor: '#fff1f2', separatorColor: 'rgba(251,113,133,0.4)',
      font: "'Josefin Sans', sans-serif", fontSize: 16, appearEffect: 'fade', fillStyle: 'liquid', rankFx: 'bloom',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 540, emptyDurationMs: 1600,
      emptyEffect: 'drain', vesselWave: true, idlePulse: true, skinId: 'sakura-dusk',
    },
  },
  {
    id: 'petal-dawn',
    name: 'Petal Dawn',
    desc: 'Bloom · рассветные лепестки',
    preview: { bg: 'rgba(255,241,242,0.96)', wins: '#e11d48', losses: '#7c3aed' },
    settings: {
      bgColor: 'rgba(255,241,242,0.96)', bgImage: '', winsColor: '#e11d48', lossesColor: '#7c3aed',
      rankTextColor: '#4c0519', iconColor: '#881337', separatorColor: 'rgba(76,5,25,0.3)',
      font: "'Nunito', sans-serif", fontSize: 16, appearEffect: 'bounce', fillStyle: 'solid', rankFx: 'bloom',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 500, emptyDurationMs: 1500,
      emptyEffect: 'splash', vesselWave: false, idlePulse: true, skinId: 'petal-dawn',
    },
  },

  /* —— ink (2) —— */
  {
    id: 'sumi-night',
    name: 'Sumi Night',
    desc: 'Ink · тушь на ночи',
    preview: { bg: 'rgba(12,12,16,0.94)', wins: '#e7e5e4', losses: '#a8a29e' },
    settings: {
      bgColor: 'rgba(12,12,16,0.94)', bgImage: '', winsColor: '#e7e5e4', lossesColor: '#a8a29e',
      rankTextColor: '#fafaf9', iconColor: '#ffffff', separatorColor: 'rgba(231,229,228,0.35)',
      font: "'Space Grotesk', sans-serif", fontSize: 16, appearEffect: 'fade', fillStyle: 'solid', rankFx: 'ink',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 520, emptyDurationMs: 1500,
      emptyEffect: 'fade', vesselWave: false, idlePulse: true, skinId: 'sumi-night',
    },
  },
  {
    id: 'indigo-scroll',
    name: 'Indigo Scroll',
    desc: 'Ink · индиго свиток',
    preview: { bg: 'rgba(238,242,255,0.96)', wins: '#4338ca', losses: '#be123c' },
    settings: {
      bgColor: 'rgba(238,242,255,0.96)', bgImage: '', winsColor: '#4338ca', lossesColor: '#be123c',
      rankTextColor: '#1e1b4b', iconColor: '#312e81', separatorColor: 'rgba(30,27,75,0.35)',
      font: "'Cinzel', serif", fontSize: 15, appearEffect: 'slide', fillStyle: 'liquid', rankFx: 'ink',
      animDirection: 'right', fillLimit: 10, fillDurationMs: 560, emptyDurationMs: 1600,
      emptyEffect: 'pour', vesselWave: true, idlePulse: false, skinId: 'indigo-scroll',
    },
  },

  /* —— prism (2) —— */
  {
    id: 'rainbow-cut',
    name: 'Rainbow Cut',
    desc: 'Prism · спектр',
    preview: { bg: 'rgba(14,10,28,0.93)', wins: '#f472b6', losses: '#22d3ee' },
    settings: {
      bgColor: 'rgba(14,10,28,0.93)', bgImage: '', winsColor: '#f472b6', lossesColor: '#22d3ee',
      rankTextColor: '#fce7f3', iconColor: '#ffffff', separatorColor: 'rgba(244,114,182,0.4)',
      font: "'Audiowide', system-ui", fontSize: 14, appearEffect: 'zoom', fillStyle: 'glow', rankFx: 'prism',
      animDirection: 'top', fillLimit: 10, fillDurationMs: 440, emptyDurationMs: 1500,
      emptyEffect: 'burst', vesselWave: true, idlePulse: true, skinId: 'rainbow-cut',
    },
  },
  {
    id: 'opal-shine',
    name: 'Opal Shine',
    desc: 'Prism · опаловый блеск',
    preview: { bg: 'rgba(245,243,255,0.96)', wins: '#8b5cf6', losses: '#06b6d4' },
    settings: {
      bgColor: 'rgba(245,243,255,0.96)', bgImage: '', winsColor: '#8b5cf6', lossesColor: '#06b6d4',
      rankTextColor: '#4c1d95', iconColor: '#5b21b6', separatorColor: 'rgba(76,29,149,0.3)',
      font: "'Montserrat', sans-serif", fontSize: 16, appearEffect: 'fade', fillStyle: 'liquid', rankFx: 'prism',
      animDirection: 'left', fillLimit: 10, fillDurationMs: 500, emptyDurationMs: 1500,
      emptyEffect: 'splash', vesselWave: true, idlePulse: true, skinId: 'opal-shine',
    },
  },
];

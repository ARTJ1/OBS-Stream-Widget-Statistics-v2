window.WIDGET_FONTS = [
  { id: 'Arial, sans-serif', label: 'Arial', google: null },
  { id: "'Segoe UI', sans-serif", label: 'Segoe UI', google: null },
  { id: 'Tahoma, sans-serif', label: 'Tahoma', google: null },
  { id: 'Verdana, sans-serif', label: 'Verdana', google: null },
  { id: "'Trebuchet MS', sans-serif", label: 'Trebuchet MS', google: null },
  { id: 'Impact, sans-serif', label: 'Impact', google: null },
  { id: "'Times New Roman', serif", label: 'Times New Roman', google: null },
  { id: 'Georgia, serif', google: null, label: 'Georgia' },
  { id: "'Courier New', monospace", label: 'Courier New', google: null },
  { id: 'Consolas, monospace', label: 'Consolas', google: null },
  { id: "'Oswald', sans-serif", label: 'Oswald', google: 'Oswald:wght@400;600;700' },
  { id: "'Roboto', sans-serif", label: 'Roboto', google: 'Roboto:wght@400;500;700' },
  { id: "'Roboto Condensed', sans-serif", label: 'Roboto Condensed', google: 'Roboto+Condensed:wght@400;700' },
  { id: "'Montserrat', sans-serif", label: 'Montserrat', google: 'Montserrat:wght@400;600;700' },
  { id: "'Poppins', sans-serif", label: 'Poppins', google: 'Poppins:wght@400;600;700' },
  { id: "'Inter', sans-serif", label: 'Inter', google: 'Inter:wght@400;600;700' },
  { id: "'Nunito', sans-serif", label: 'Nunito', google: 'Nunito:wght@400;700' },
  { id: "'Rubik', sans-serif", label: 'Rubik', google: 'Rubik:wght@400;600;700' },
  { id: "'Exo 2', sans-serif", label: 'Exo 2', google: 'Exo+2:wght@400;600;700' },
  { id: "'Rajdhani', sans-serif", label: 'Rajdhani', google: 'Rajdhani:wght@500;600;700' },
  { id: "'Orbitron', sans-serif", label: 'Orbitron', google: 'Orbitron:wght@500;700' },
  { id: "'Bebas Neue', sans-serif", label: 'Bebas Neue', google: 'Bebas+Neue' },
  { id: "'Black Ops One', system-ui", label: 'Black Ops One', google: 'Black+Ops+One' },
  { id: "'Press Start 2P', system-ui", label: 'Press Start 2P', google: 'Press+Start+2P' },
  { id: "'Russo One', sans-serif", label: 'Russo One', google: 'Russo+One' },
  { id: "'Unbounded', sans-serif", label: 'Unbounded', google: 'Unbounded:wght@400;700' },
  { id: "'Manrope', sans-serif", label: 'Manrope', google: 'Manrope:wght@400;700' },
  { id: "'Sora', sans-serif", label: 'Sora', google: 'Sora:wght@400;600;700' },
  { id: "'Kanit', sans-serif", label: 'Kanit', google: 'Kanit:wght@400;600;700' },
  { id: "'Teko', sans-serif", label: 'Teko', google: 'Teko:wght@400;600;700' },
  { id: "'Anton', sans-serif", label: 'Anton', google: 'Anton' },
  { id: "'Bangers', system-ui", label: 'Bangers', google: 'Bangers' },
  { id: "'Comfortaa', cursive", label: 'Comfortaa', google: 'Comfortaa:wght@400;700' },
  { id: "'Josefin Sans', sans-serif", label: 'Josefin Sans', google: 'Josefin+Sans:wght@400;600;700' },
  { id: "'Playfair Display', serif", label: 'Playfair Display', google: 'Playfair+Display:wght@400;700' },
  { id: "'Audiowide', system-ui", label: 'Audiowide', google: 'Audiowide' },
  { id: "'Chakra Petch', sans-serif", label: 'Chakra Petch', google: 'Chakra+Petch:wght@400;600;700' },
  { id: "'Cinzel', serif", label: 'Cinzel', google: 'Cinzel:wght@400;700' },
  { id: "'Space Grotesk', sans-serif", label: 'Space Grotesk', google: 'Space+Grotesk:wght@400;600;700' },
];

window.loadGoogleFont = function loadGoogleFont(fontId) {
  const item = (window.WIDGET_FONTS || []).find((f) => f.id === fontId);
  if (!item || !item.google) return;
  const id = 'gf-' + item.google.replace(/[^a-z0-9]+/gi, '-');
  if (document.getElementById(id)) return;
  const link = document.createElement('link');
  link.id = id;
  link.rel = 'stylesheet';
  link.href = 'https://fonts.googleapis.com/css2?family=' + item.google + '&display=swap';
  document.head.appendChild(link);
};

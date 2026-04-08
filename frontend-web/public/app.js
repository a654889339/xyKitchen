(function () {
  const API = window.__XYKITCHEN_API__ || 'http://106.54.50.88:5402/api';
  const BASE = API.replace(/\/api\/?$/, '') || 'http://106.54.50.88:5402';
  const SPLASH_KEY = 'xykitchen_splash_shown';
  const LANG_KEY = 'xykitchen_lang';

  function getLang() {
    try {
      return sessionStorage.getItem(LANG_KEY) || localStorage.getItem(LANG_KEY) || 'zh';
    } catch (e) {
      return 'zh';
    }
  }

  function setLang(lang) {
    try {
      sessionStorage.setItem(LANG_KEY, lang);
      localStorage.setItem(LANG_KEY, lang);
    } catch (e) {}
  }

  function pick(obj, field) {
    if (!obj) return '';
    const lang = getLang();
    if (lang === 'en') {
      const enVal = obj[field + 'En'] || obj[field + '_en'];
      if (enVal && String(enVal).trim()) return enVal;
    }
    return obj[field] || '';
  }

  function toFull(u) {
    if (!u || typeof u !== 'string') return '';
    const t = u.trim();
    if (t.startsWith('http://') || t.startsWith('https://')) return t;
    return BASE + (t.startsWith('/') ? t : '/' + t);
  }

  function showSplashLayer(bg, imageUrl, text, textColor) {
    const el = document.getElementById('splash');
    el.style.background = bg || '#000';
    el.classList.remove('splash-screen--hidden');
    el.setAttribute('aria-hidden', 'false');
    const img = document.getElementById('splashImg');
    const tx = document.getElementById('splashText');
    if (imageUrl) {
      img.src = imageUrl;
      img.hidden = false;
    } else {
      img.hidden = true;
    }
    tx.textContent = text || '';
    tx.style.color = textColor || 'rgba(255,255,255,0.85)';
  }

  function hideSplashLayer() {
    const el = document.getElementById('splash');
    el.classList.add('splash-screen--hidden');
    el.setAttribute('aria-hidden', 'true');
  }

  function maybeSplash(splash) {
    try {
      if (sessionStorage.getItem(SPLASH_KEY)) return;
    } catch (e) {}
    if (!splash) {
      try {
        sessionStorage.setItem(SPLASH_KEY, '1');
      } catch (e) {}
      return;
    }
    const splashBg = (splash.color && String(splash.color).trim()) || '#000';
    const imageUrl = toFull(pick(splash, 'imageUrl'));
    const text =
      (pick(splash, 'desc') || splash.desc || '').trim() ||
      (pick(splash, 'title') || splash.title || '').trim();
    const c = splashBg.toLowerCase();
    const isLight =
      c === '#fff' ||
      c === '#ffffff' ||
      c === 'white';
    const textColor = isLight ? 'rgba(0,0,0,0.75)' : 'rgba(255,255,255,0.85)';
    showSplashLayer(splashBg, imageUrl, text, textColor);
    try {
      sessionStorage.setItem(SPLASH_KEY, '1');
    } catch (e) {}
    setTimeout(hideSplashLayer, 3200);
    document.getElementById('splash').addEventListener(
      'click',
      () => {
        hideSplashLayer();
      },
      { once: true }
    );
  }

  function renderHeroBg(list, single) {
    const root = document.getElementById('homeBg');
    root.innerHTML = '';
    if (list.length > 1) {
      const wrap = document.createElement('div');
      wrap.className = 'home-bg-swiper';
      list.forEach((url, i) => {
        const img = document.createElement('img');
        img.src = url;
        img.alt = '';
        if (i === 0) img.classList.add('is-active');
        wrap.appendChild(img);
      });
      root.appendChild(wrap);
      let idx = 0;
      setInterval(() => {
        const imgs = wrap.querySelectorAll('img');
        if (!imgs.length) return;
        imgs[idx].classList.remove('is-active');
        idx = (idx + 1) % imgs.length;
        imgs[idx].classList.add('is-active');
      }, 4000);
    } else if (single) {
      const img = document.createElement('img');
      img.src = single;
      img.alt = '';
      img.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;object-fit:cover;';
      root.appendChild(img);
    }
  }

  function updateLangUi() {
    const zh = getLang() !== 'en';
    document.getElementById('langLabel').textContent = zh ? '中' : 'EN';
  }

  document.getElementById('langBtn').addEventListener('click', () => {
    const next = getLang() === 'en' ? 'zh' : 'en';
    setLang(next);
    updateLangUi();
    loadConfig();
  });

  async function loadConfig() {
    updateLangUi();
    let items = [];
    try {
      const ctrl = new AbortController();
      const t = setTimeout(() => ctrl.abort(), 20000);
      const res = await fetch(API + '/home-config?all=1', { signal: ctrl.signal });
      clearTimeout(t);
      const json = await res.json();
      if (json && json.code === 0 && Array.isArray(json.data)) items = json.data;
    } catch (e) {
      items = [];
    }

    const splash = items.find((i) => i.section === 'splash' && i.status === 'active');
    const headerLogo = items.find((i) => i.section === 'headerLogo' && i.status === 'active');
    const homeBgItems = items
      .filter((i) => i.section === 'homeBg' && i.status === 'active')
      .sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0));

    const logoEl = document.getElementById('headerLogo');
    const logoText = document.getElementById('headerLogoText');
    if (headerLogo) {
      const u = toFull(pick(headerLogo, 'imageUrl'));
      if (u) {
        logoEl.src = u;
        logoEl.hidden = false;
        logoText.style.display = 'none';
      } else {
        logoEl.hidden = true;
        logoText.style.display = '';
      }
    } else {
      logoEl.hidden = true;
      logoText.style.display = '';
    }

    const urls = homeBgItems.map((i) => toFull(pick(i, 'imageUrl'))).filter(Boolean);
    renderHeroBg(urls, urls[0] || '');

    maybeSplash(splash);
  }

  loadConfig();
})();

const i18n = require('../../utils/i18n.js');

const SPLASH_SESSION_KEY = 'xykitchen_splash_shown';

Page({
  data: {
    showSplash: false,
    splashBg: '#000000',
    splashImageUrl: '',
    splashText: '',
    splashTextColor: 'rgba(255,255,255,0.85)',
    headerLogoUrl: '',
    heroBgUrl: '',
    heroBgList: [],
    langLabel: '中',
  },

  _splashTimer: null,
  _bootstrapped: false,

  onLoad() {
    // 不在模块顶层调用 getApp()，避免 App 未就绪时触发异常
  },

  onShow() {
    if (this._bootstrapped) {
      return;
    }
    this._bootstrapped = true;
    const self = this;
    const run = () => {
      i18n.detectLangByIp(() => {
        self.setData({
          langLabel: i18n.isEn() ? 'EN' : i18n.t('lang.zhLabel', '中'),
        });
      });
      self.loadAnimationConfig(() => {
        const runSplash = () => self.maybeShowSplash();
        if (typeof wx.nextTick === 'function') {
          wx.nextTick(runSplash);
        } else {
          setTimeout(runSplash, 0);
        }
      });
    };
    if (i18n.isLoaded()) {
      run();
    } else {
      i18n.loadI18nTexts(run);
    }
  },

  onUnload() {
    if (this._splashTimer) {
      clearTimeout(this._splashTimer);
      this._splashTimer = null;
    }
  },

  onLangTap() {
    const items = [i18n.t('lang.zhName', '中文'), 'English'];
    wx.showActionSheet({
      itemList: items,
      success: (res) => {
        const lang = res.tapIndex === 1 ? 'en' : 'zh';
        i18n.setLang(lang);
        this.setData({
          langLabel: i18n.isEn() ? 'EN' : i18n.t('lang.zhLabel', '中'),
        });
        this.loadAnimationConfig();
      },
    });
  },

  maybeShowSplash() {
    try {
      if (wx.getStorageSync(SPLASH_SESSION_KEY)) {
        this.setData({ showSplash: false });
        return;
      }
    } catch (e) {}
    this.setData({ showSplash: true });
    try {
      wx.setStorageSync(SPLASH_SESSION_KEY, '1');
    } catch (e) {}
    this._splashTimer = setTimeout(() => {
      this.setData({ showSplash: false });
      this._splashTimer = null;
    }, 3200);
  },

  dismissSplash() {
    if (this._splashTimer) {
      clearTimeout(this._splashTimer);
      this._splashTimer = null;
    }
    this.setData({ showSplash: false });
  },

  loadAnimationConfig(done) {
    const app = getApp();
    const base = (app.globalData.baseUrl || '').replace(/\/api\/?$/, '') || 'http://106.54.50.88:5402';
    const toFull = (u) => {
      if (!u || typeof u !== 'string') return u || '';
      const t = u.trim();
      if (t.startsWith('http://') || t.startsWith('https://')) return t;
      return base + (t.startsWith('/') ? t : '/' + t);
    };

    app
      .request({ url: '/home-config?all=1', timeout: 20000 })
      .then((res) => {
        const items = res.data || [];
        const splash = items.find((i) => i.section === 'splash' && i.status === 'active');
        const headerLogo = items.find((i) => i.section === 'headerLogo' && i.status === 'active');
        const homeBgItems = items
          .filter((i) => i.section === 'homeBg' && i.status === 'active')
          .sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0));

        let splashBg = '#000000';
        let splashImageUrl = '';
        let splashText = '';
        if (splash) {
          splashBg = (splash.color && String(splash.color).trim()) || splashBg;
          splashImageUrl = toFull(i18n.pick(splash, 'imageUrl'));
          splashText = (i18n.pick(splash, 'desc') || splash.desc || '').trim() || (i18n.pick(splash, 'title') || splash.title || '').trim();
        }

        const heroBgList = homeBgItems
          .map((i) => {
            const url = toFull(i18n.pick(i, 'imageUrl'));
            return { url, displayUrl: url };
          })
          .filter((i) => i.url);
        const singleBg = heroBgList[0] ? heroBgList[0].displayUrl : '';

        const isLightBg = (() => {
          const c = (splashBg || '').trim().toLowerCase();
          if (!c || c === '#000' || c === '#000000' || c === 'black') return false;
          if (c === '#fff' || c === '#ffffff' || c === 'white') return true;
          return false;
        })();

        this.setData({
          splashBg,
          splashImageUrl,
          splashText,
          splashTextColor: isLightBg ? 'rgba(0,0,0,0.75)' : 'rgba(255,255,255,0.85)',
          headerLogoUrl: headerLogo ? toFull(i18n.pick(headerLogo, 'imageUrl')) : '',
          heroBgUrl: singleBg,
          heroBgList,
        });
        if (typeof done === 'function') done();
      })
      .catch(() => {
        if (typeof done === 'function') done();
      });
  },

  onHeroBgError(e) {
    const idx = e.currentTarget.dataset.idx;
    const list = this.data.heroBgList || [];
    if (!list[idx] || !list[idx].url) return;
    if (list[idx].displayUrl === list[idx].url) return;
    const next = [...list];
    next[idx] = { ...next[idx], displayUrl: next[idx].url };
    this.setData({ heroBgList: next });
  },
});

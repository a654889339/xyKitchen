const STORAGE_KEY = 'xykitchen_lang';
const DEFAULT_BASE_URL = 'http://106.54.50.88:5402/api';

let _lang = '';
try { _lang = wx.getStorageSync(STORAGE_KEY) || ''; } catch (e) {}

const _texts = {};
let _loaded = false;
let _loading = false;
const _pendingCallbacks = [];

function getLang() { return _lang || 'zh'; }
function isEn() { return _lang === 'en'; }

function setLang(lang) {
  _lang = lang;
  try { wx.setStorageSync(STORAGE_KEY, lang); } catch (e) {}
}

function _getBaseUrl() {
  try {
    const app = getApp();
    if (app && app.globalData && app.globalData.baseUrl) return app.globalData.baseUrl;
  } catch (e) {}
  return DEFAULT_BASE_URL;
}

function detectLangByIp(cb) {
  if (_lang) { if (cb) cb(_lang); return; }
  wx.request({
    url: 'https://ipapi.co/json/',
    timeout: 3000,
    success(res) {
      const d = res.data || {};
      const cn = d.country_code === 'CN' || d.country_code === 'HK' || d.country_code === 'MO' || d.country_code === 'TW';
      setLang(cn ? 'zh' : 'en');
      if (cb) cb(getLang());
    },
    fail() {
      setLang('zh');
      if (cb) cb('zh');
    },
  });
}

function loadI18nTexts(cb) {
  if (_loaded) { if (cb) cb(); return; }
  if (cb) _pendingCallbacks.push(cb);
  if (_loading) return;
  _loading = true;
  wx.request({
    url: _getBaseUrl() + '/i18n',
    success(res) {
      const data = (res.data && res.data.data) || [];
      for (const item of data) {
        _texts[item.key] = { zh: item.zh, en: item.en };
      }
      _loaded = true;
    },
    complete() {
      _loading = false;
      const cbs = _pendingCallbacks.splice(0);
      cbs.forEach(function(fn) { fn(); });
    },
  });
}

function isLoaded() { return _loaded; }

function t(keyOrZh, enFallback) {
  if (_texts[keyOrZh]) {
    const entry = _texts[keyOrZh];
    return isEn() ? (entry.en || entry.zh || '') : (entry.zh || '');
  }
  return isEn() ? (enFallback || keyOrZh || '') : (keyOrZh || '');
}

function pick(obj, field) {
  if (!obj) return '';
  if (isEn()) {
    const enVal = obj[field + 'En'] || obj[field + '_en'];
    if (enVal && String(enVal).trim()) return enVal;
  }
  return obj[field] || '';
}

module.exports = { getLang, isEn, setLang, detectLangByIp, loadI18nTexts, isLoaded, t, pick };

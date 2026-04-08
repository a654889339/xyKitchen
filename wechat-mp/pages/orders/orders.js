const { ensureWxLogin } = require('../../utils/auth.js');

Page({
  data: {
    list: [],
    empty: false,
  },

  onShow() {
    ensureWxLogin()
      .then(() => this.loadList())
      .catch(() => {
        wx.showToast({ title: '请先登录', icon: 'none' });
        setTimeout(() => wx.navigateBack(), 500);
      });
  },

  loadList() {
    const app = getApp();
    app
      .request({ url: '/orders/mine?pageSize=50' })
      .then((res) => {
        const raw = (res.data && res.data.list) || [];
        const list = raw.map((o) => ({
          ...o,
          bookingAtText: fmtIso(o.bookingAt),
          priceText: o.price != null ? String(o.price) : '-',
        }));
        this.setData({ list, empty: list.length === 0 });
      })
      .catch(() => {
        this.setData({ list: [], empty: true });
      });
  },

});

function fmtIso(iso) {
  if (!iso) return '-';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '-';
  return (
    d.getFullYear() +
    '-' +
    pad(d.getMonth() + 1) +
    '-' +
    pad(d.getDate()) +
    ' ' +
    pad(d.getHours()) +
    ':' +
    pad(d.getMinutes())
  );
}

function pad(n) {
  return n < 10 ? '0' + n : '' + n;
}

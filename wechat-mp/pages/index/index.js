const { ensureWxLogin } = require('../../utils/auth.js');

Page({
  data: {
    homepageStyle: '',
    loading: true,
  },

  onShow() {
    this.loadBookingHome();
  },

  loadBookingHome() {
    const app = getApp();
    const base = (app.globalData.baseUrl || '').replace(/\/api\/?$/, '') || '';
    app
      .request({ url: '/booking-config' })
      .then((res) => {
        const d = res.data || {};
        let u = d.homepageBgUrl || '';
        if (u && !/^https?:\/\//i.test(u)) {
          u = base + (u.startsWith('/') ? u : '/' + u);
        }
        this.setData({
          homepageStyle: u ? 'background-image:url(' + u + ')' : '',
          loading: false,
        });
      })
      .catch(() => {
        this.setData({ loading: false });
      });
  },

  onStartBooking() {
    wx.navigateTo({ url: '/pages/booking-notice/booking-notice' });
  },

  onMyOrders() {
    ensureWxLogin()
      .then(() => {
        wx.navigateTo({ url: '/pages/orders/orders' });
      })
      .catch((e) => {
        wx.showToast({ title: (e && e.message) || '请先登录', icon: 'none' });
      });
  },
});

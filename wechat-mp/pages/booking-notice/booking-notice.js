Page({
  data: {
    noticeTitle: '預定須知',
    noticeBody: '',
  },

  onLoad() {
    const app = getApp();
    app
      .request({ url: '/booking-config' })
      .then((res) => {
        const d = res.data || {};
        this.setData({
          noticeTitle: d.noticeTitle || '預定須知',
          noticeBody: d.noticeBody || '',
        });
      })
      .catch(() => {});
  },

  onNext() {
    wx.navigateTo({ url: '/pages/booking/booking' });
  },

  onGoHome() {
    wx.navigateBack({ delta: 99 }).catch(() => {
      wx.reLaunch({ url: '/pages/index/index' });
    });
  },
});

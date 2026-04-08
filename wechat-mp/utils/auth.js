function ensureWxLogin() {
  const app = getApp();
  if (app.globalData.token) {
    return Promise.resolve();
  }
  try {
    const t = wx.getStorageSync('xykitchen_token');
    if (t) {
      app.globalData.token = t;
      return Promise.resolve();
    }
  } catch (e) {}
  return new Promise((resolve, reject) => {
    wx.login({
      success: (r) => {
        if (!r.code) {
          reject(new Error('无法获取登录凭证'));
          return;
        }
        app
          .request({ url: '/auth/wx-login', method: 'POST', data: { code: r.code } })
          .then((res) => {
            const tok = res.data && res.data.token;
            if (tok) {
              app.setToken(tok);
              resolve();
            } else {
              reject(new Error('登录失败'));
            }
          })
          .catch(reject);
      },
      fail: () => reject(new Error('wx.login 失败')),
    });
  });
}

module.exports = { ensureWxLogin };

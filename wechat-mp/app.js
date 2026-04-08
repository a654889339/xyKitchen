App({
  globalData: {
    baseUrl: 'http://106.54.50.88:5402/api',
    userInfo: null,
    token: '',
  },

  // 勿在 onLaunch 里发起网络请求，避免部分基础库出现 appLaunch / 页面栈异常与启动超时
  onLaunch() {},

  request(options) {
    const app = this;
    const { baseUrl, token } = app.globalData;
    return new Promise((resolve, reject) => {
      wx.request({
        url: baseUrl + options.url,
        method: options.method || 'GET',
        data: options.data,
        timeout: options.timeout || 20000,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: 'Bearer ' + token } : {}),
        },
        success: (res) => {
          if (res.statusCode === 401) {
            app.clearToken();
            reject(new Error('请先登录'));
            return;
          }
          if (res.data && res.data.code === 0) {
            resolve(res.data);
          } else {
            reject(new Error((res.data && res.data.message) || '请求失败'));
          }
        },
        fail: (err) => {
          reject(err.errMsg || new Error('网络错误'));
        },
      });
    });
  },

  setToken(token) {
    this.globalData.token = token;
    wx.setStorageSync('xykitchen_token', token);
  },

  clearToken() {
    this.globalData.token = '';
    this.globalData.userInfo = null;
    wx.removeStorageSync('xykitchen_token');
  },
});

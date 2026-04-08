const { ensureWxLogin } = require('../../utils/auth.js');

function pad2(n) {
  return n < 10 ? '0' + n : '' + n;
}

Page({
  data: {
    viewYear: 2026,
    viewMonth: 4,
    monthLabel: '',
    weekdays: ['日', '一', '二', '三', '四', '五', '六'],
    cells: [],
    availableDays: [],
    metaLoaded: false,
    timeSlots: ['17:00', '19:00', '21:00'],
    guestOptions: [2, 3, 4, 6],
    perPersonDeposit: 50,
    selectedDate: '',
    selectedDay: 0,
    selectedSlot: '',
    guestCount: 0,
    contactPhone: '',
    depositText: '0',
    submitting: false,
  },

  onLoad() {
    const now = new Date();
    const y = now.getFullYear();
    const m = now.getMonth() + 1;
    this.setData({ viewYear: y, viewMonth: m, monthLabel: m + '月' });
    this.loadMeta();
    this.loadCalendar(y, m);
  },

  loadMeta() {
    const app = getApp();
    app
      .request({ url: '/booking/meta' })
      .then((res) => {
        const d = res.data || {};
        this.setData({
          metaLoaded: true,
          timeSlots: d.timeSlots && d.timeSlots.length ? d.timeSlots : ['17:00', '19:00', '21:00'],
          guestOptions: d.guestOptions && d.guestOptions.length ? d.guestOptions : [2, 3, 4, 6],
          perPersonDeposit: d.perPersonDeposit > 0 ? d.perPersonDeposit : 50,
        });
        this.updateDeposit();
      })
      .catch(() => {});
  },

  loadCalendar(y, m) {
    const app = getApp();
    app
      .request({ url: '/booking/calendar?year=' + y + '&month=' + m })
      .then((res) => {
        const d = res.data || {};
        const days = d.days || [];
        this.setData({
          availableDays: days,
          viewYear: d.year || y,
          viewMonth: d.month || m,
          monthLabel: (d.month || m) + '月',
        });
        this.buildGrid(d.year || y, d.month || m, days);
      })
      .catch(() => {
        this.buildGrid(y, m, []);
      });
  },

  buildGrid(year, month, availableDays) {
    const first = new Date(year, month - 1, 1);
    const lastDate = new Date(year, month, 0).getDate();
    const startWd = first.getDay();
    const cells = [];
    for (let i = 0; i < startWd; i++) {
      cells.push({ day: null, label: '', available: false, muted: true });
    }
    for (let day = 1; day <= lastDate; day++) {
      const ok = availableDays.indexOf(day) >= 0;
      cells.push({ day, label: String(day), available: ok, muted: !ok });
    }
    while (cells.length % 7 !== 0) {
      cells.push({ day: null, label: '', available: false, muted: true });
    }
    this.setData({ cells });
  },

  onPrevMonth() {
    let { viewYear, viewMonth } = this.data;
    viewMonth--;
    if (viewMonth < 1) {
      viewMonth = 12;
      viewYear--;
    }
    this.setData({ selectedDate: '', selectedDay: 0, selectedSlot: '', guestCount: 0 });
    this.loadCalendar(viewYear, viewMonth);
  },

  onNextMonth() {
    let { viewYear, viewMonth } = this.data;
    viewMonth++;
    if (viewMonth > 12) {
      viewMonth = 1;
      viewYear++;
    }
    this.setData({ selectedDate: '', selectedDay: 0, selectedSlot: '', guestCount: 0 });
    this.loadCalendar(viewYear, viewMonth);
  },

  onPickDay(e) {
    const day = e.currentTarget.dataset.day;
    const avail = e.currentTarget.dataset.avail;
    if (day == null || !avail) return;
    const { viewYear, viewMonth } = this.data;
    const ds = viewYear + '-' + pad2(viewMonth) + '-' + pad2(day);
    this.setData({ selectedDate: ds, selectedDay: day, selectedSlot: '', guestCount: 0 });
    this.updateDeposit();
  },

  onPickSlot(e) {
    const slot = e.currentTarget.dataset.slot;
    this.setData({ selectedSlot: slot, guestCount: 0 });
    this.updateDeposit();
  },

  onPickGuest(e) {
    const n = Number(e.currentTarget.dataset.n);
    this.setData({ guestCount: n });
    this.updateDeposit();
  },

  onPhoneInput(e) {
    this.setData({ contactPhone: (e.detail.value || '').replace(/\D/g, '').slice(0, 11) });
  },

  updateDeposit() {
    const { perPersonDeposit, guestCount } = this.data;
    const total = perPersonDeposit * (guestCount || 0);
    const t = Number.isInteger(total) ? String(total) : total.toFixed(2);
    this.setData({ depositText: t });
  },

  onConfirm() {
    const { selectedDate, selectedSlot, guestCount, contactPhone, depositText, perPersonDeposit } = this.data;
    if (!selectedDate || !selectedSlot || !guestCount) {
      wx.showToast({ title: '请选择日期、时段与人数', icon: 'none' });
      return;
    }
    if (!/^1[3-9]\d{9}$/.test(contactPhone)) {
      wx.showToast({ title: '请输入11位手机号', icon: 'none' });
      return;
    }
    const expect = perPersonDeposit * guestCount;
    const price = parseFloat(depositText);
    if (Math.abs(expect - price) > 0.02) {
      wx.showToast({ title: '金额异常，请重试', icon: 'none' });
      return;
    }
    this.setData({ submitting: true });
    ensureWxLogin()
      .then(() => {
        const app = getApp();
        return app.request({
          url: '/orders',
          method: 'POST',
          data: {
            bookingDate: selectedDate,
            timeSlot: selectedSlot,
            guestCount,
            contactPhone,
            price: expect,
          },
        });
      })
      .then(() => {
        wx.showToast({ title: '预订已提交', icon: 'success' });
        setTimeout(() => {
          wx.navigateBack({ delta: 2 });
        }, 800);
      })
      .catch((err) => {
        wx.showToast({ title: (err && err.message) || '提交失败', icon: 'none' });
      })
      .finally(() => {
        this.setData({ submitting: false });
      });
  },
});

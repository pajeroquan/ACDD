const { request, fenToYuan } = require('../../utils/request')

Page({
  data: {
    sid: '',
    partnerId: 0,
    partner: null,
    scheduleDate: '',
    startTime: '14:00',
    durationHours: 2,
    contactPhone: '',
    quote: null,
    quoting: false,
    submitting: false,
  },

  onLoad(query) {
    const app = getApp()
    const sid = query.sid || app.globalData.sid || ''
    const partnerId = Number(query.partner_id || 0)
    const partner = app.globalData.selectedPartner
    const today = new Date()
    const scheduleDate = [
      today.getFullYear(),
      String(today.getMonth() + 1).padStart(2, '0'),
      String(today.getDate()).padStart(2, '0'),
    ].join('-')

    this.setData({
      sid,
      partnerId,
      partner,
      scheduleDate,
      durationHours: partner?.min_hours || 2,
    })

    if (!app.globalData.token) {
      wx.showToast({ title: '请先从推荐页进入', icon: 'none' })
    }
  },

  onDateChange(e) {
    this.setData({ scheduleDate: e.detail.value, quote: null })
  },

  onTimeChange(e) {
    this.setData({ startTime: e.detail.value, quote: null })
  },

  onDurationInput(e) {
    this.setData({ durationHours: Number(e.detail.value) || 0, quote: null })
  },

  onPhoneInput(e) {
    this.setData({ contactPhone: e.detail.value })
  },

  async onQuote() {
    const { partnerId, scheduleDate, startTime, durationHours } = this.data
    if (!partnerId || !scheduleDate || !startTime || !durationHours) {
      wx.showToast({ title: '请完善预约信息', icon: 'none' })
      return
    }
    this.setData({ quoting: true })
    try {
      const quote = await request('/api/orders/quote', {
        method: 'POST',
        data: {
          partner_id: partnerId,
          schedule_date: scheduleDate,
          start_time: startTime,
          duration_hours: durationHours,
        },
      })
      this.setData({
        quote: {
          ...quote,
          totalText: fenToYuan(quote.total_amount_fen),
        },
      })
    } catch (e) {
      wx.showToast({ title: e.message || '询价失败', icon: 'none' })
    } finally {
      this.setData({ quoting: false })
    }
  },

  async onSubmit() {
    const { sid, partnerId, scheduleDate, startTime, durationHours, contactPhone, quote } = this.data
    if (!contactPhone) {
      wx.showToast({ title: '请填写联系电话', icon: 'none' })
      return
    }
    if (!quote) {
      wx.showToast({ title: '请先询价', icon: 'none' })
      return
    }
    this.setData({ submitting: true })
    try {
      const order = await request('/api/orders', {
        method: 'POST',
        data: {
          sid,
          partner_id: partnerId,
          schedule_date: scheduleDate,
          start_time: startTime,
          duration_hours: durationHours,
          contact_phone: contactPhone,
        },
      })
      const pay = await request('/api/orders/' + order.id + '/pay', { method: 'POST' })
      if (pay.mock) {
        await request('/api/pay/mock-notify', {
          method: 'POST',
          data: { out_trade_no: pay.out_trade_no },
        })
        wx.redirectTo({
          url: '/pages/order/success?order_no=' + (pay.order_no || order.order_no),
        })
        return
      }
      // Non-mock: attempt wx.requestPayment with returned params
      await new Promise((resolve, reject) => {
        wx.requestPayment({
          timeStamp: pay.timeStamp,
          nonceStr: pay.nonceStr,
          package: pay.package,
          signType: pay.signType || 'RSA',
          paySign: pay.paySign,
          success: resolve,
          fail: reject,
        })
      })
      wx.redirectTo({
        url: '/pages/order/success?order_no=' + (pay.order_no || order.order_no),
      })
    } catch (e) {
      wx.showToast({ title: e.message || '下单失败', icon: 'none' })
    } finally {
      this.setData({ submitting: false })
    }
  },
})

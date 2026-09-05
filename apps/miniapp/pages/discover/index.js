const { request, fenToYuan } = require('../../utils/request')

Page({
  data: {
    sid: '',
    cards: [],
    current: 0,
    loading: true,
    empty: false,
  },

  onLoad(query) {
    const sid = query.sid || ''
    this.setData({ sid })
    getApp().globalData.sid = sid
    this.bootstrap(sid)
  },

  async bootstrap(sid) {
    if (!sid) {
      this.setData({ loading: false, empty: true })
      wx.showToast({ title: '缺少 sid', icon: 'none' })
      return
    }
    try {
      await this.ensureLogin()
      const data = await request('/api/browse/' + sid)
      const cards = (data.cards || []).map((c) => ({
        ...c,
        priceText: fenToYuan(c.hourly_price_fen),
        cover: c.avatar_url || (c.gallery && c.gallery[0]) || '',
        tagsText: (c.tags || []).join(' · '),
      }))
      this.setData({
        cards,
        loading: false,
        empty: cards.length === 0,
        current: 0,
      })
    } catch (e) {
      this.setData({ loading: false, empty: true })
      wx.showToast({ title: e.message || '加载失败', icon: 'none' })
    }
  },

  ensureLogin() {
    const app = getApp()
    if (app.globalData.token) {
      return Promise.resolve(app.globalData.token)
    }
    return new Promise((resolve, reject) => {
      // Mock-friendly: prefer wx.login code; fall back to mock code for tools/dev
      wx.login({
        success: async (res) => {
          try {
            const code = res.code || 'mock_code'
            const data = await request('/api/wx/login', {
              method: 'POST',
              data: { code, nickname: '微信用户' },
            })
            app.globalData.token = data.token
            wx.setStorageSync('token', data.token)
            resolve(data.token)
          } catch (err) {
            reject(err)
          }
        },
        fail: async () => {
          try {
            const data = await request('/api/wx/login', {
              method: 'POST',
              data: { code: 'mock_code', nickname: '微信用户' },
            })
            app.globalData.token = data.token
            wx.setStorageSync('token', data.token)
            resolve(data.token)
          } catch (err) {
            reject(err)
          }
        },
      })
    })
  },

  onSwiperChange(e) {
    this.setData({ current: e.detail.current })
  },

  onSkip() {
    const { current, cards } = this.data
    if (current < cards.length - 1) {
      this.setData({ current: current + 1 })
    } else {
      wx.showToast({ title: '没有更多了', icon: 'none' })
    }
  },

  onLike() {
    const card = this.data.cards[this.data.current]
    if (!card) return
    getApp().globalData.selectedPartner = card
    wx.navigateTo({
      url: '/pages/order/create?partner_id=' + card.partner_id + '&sid=' + this.data.sid,
    })
  },
})

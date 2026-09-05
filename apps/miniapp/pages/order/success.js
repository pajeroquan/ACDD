Page({
  data: {
    orderNo: '',
  },
  onLoad(query) {
    this.setData({ orderNo: query.order_no || '' })
  },
  onBack() {
    wx.reLaunch({ url: '/pages/discover/index?sid=' + (getApp().globalData.sid || '') })
  },
})

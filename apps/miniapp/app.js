App({
  globalData: {
    token: '',
    sid: '',
    selectedPartner: null,
  },
  onLaunch() {
    this.globalData.token = wx.getStorageSync('token') || ''
  },
})

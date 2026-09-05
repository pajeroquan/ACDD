const API_BASE = 'http://localhost:8080'

function request(path, options = {}) {
  const app = getApp()
  const token = options.token || app.globalData.token || wx.getStorageSync('token') || ''
  return new Promise((resolve, reject) => {
    wx.request({
      url: API_BASE + path,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: 'Bearer ' + token } : {}),
        ...(options.header || {}),
      },
      success(res) {
        const body = res.data || {}
        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
          return
        }
        const msg = body.message || '请求失败(' + res.statusCode + ')'
        reject(new Error(msg))
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

function fenToYuan(fen) {
  return ((fen || 0) / 100).toFixed(2)
}

module.exports = {
  API_BASE,
  request,
  fenToYuan,
}

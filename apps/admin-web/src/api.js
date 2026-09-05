import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from './router'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || 'http://localhost:8080',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && typeof body.code === 'number' && body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(body)
    }
    return body?.data !== undefined ? body.data : body
  },
  (err) => {
    const status = err.response?.status
    const msg = err.response?.data?.message || err.message || '网络错误'
    if (status === 401) {
      localStorage.removeItem('admin_token')
      if (router.currentRoute.value.path !== '/login') {
        router.push('/login')
      }
    }
    ElMessage.error(msg)
    return Promise.reject(err)
  }
)

export const login = (username, password) =>
  api.post('/admin/login', { username, password })

export const matchChat = (text, matchRequestId) =>
  api.post('/admin/match/chat', {
    text,
    match_request_id: matchRequestId || undefined,
  })

export const matchConfirm = (id, partnerIds) =>
  api.post(`/admin/match/${id}/confirm`, { partner_ids: partnerIds || [] })

export const listUnions = () => api.get('/admin/unions')
export const createUnion = (data) => api.post('/admin/unions', data)
export const updateUnion = (id, data) => api.patch(`/admin/unions/${id}`, data)

export const listPartners = (params) => api.get('/admin/partners', { params })
export const createPartner = (data) => api.post('/admin/partners', data)
export const updatePartner = (id, data) => api.put(`/admin/partners/${id}`, data)

export const listOrders = (params) => api.get('/admin/orders', { params })
export const refundOrder = (id) => api.post(`/admin/orders/${id}/refund`)
export const commissionReport = () => api.get('/admin/commission/report')

export default api

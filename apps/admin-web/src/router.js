import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('./views/Login.vue'),
  },
  {
    path: '/',
    component: () => import('./views/Layout.vue'),
    redirect: '/match',
    children: [
      { path: 'match', name: 'match', component: () => import('./views/Match.vue'), meta: { title: 'Agent匹配台' } },
      { path: 'unions', name: 'unions', component: () => import('./views/Unions.vue'), meta: { title: '工会管理' } },
      { path: 'partners', name: 'partners', component: () => import('./views/Partners.vue'), meta: { title: '搭子管理' } },
      { path: 'orders', name: 'orders', component: () => import('./views/Orders.vue'), meta: { title: '订单中心' } },
      { path: 'commission', name: 'commission', component: () => import('./views/Commission.vue'), meta: { title: '分成报表' } },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  if (to.path === '/login') return next()
  if (!localStorage.getItem('admin_token')) return next('/login')
  next()
})

export default router

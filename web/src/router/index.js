import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/families' },
  { path: '/families', name: 'Families', component: () => import('../views/Families.vue') },
  { path: '/families/:id', name: 'FamilyDetail', component: () => import('../views/FamilyDetail.vue') },
  { path: '/persons/:id', name: 'PersonDetail', component: () => import('../views/PersonDetail.vue') },
  { path: '/birthdays', name: 'Birthdays', component: () => import('../views/Birthdays.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router

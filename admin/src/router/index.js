import Vue from 'vue'
import Router from 'vue-router'
import Bot from '@/components/Bot'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      redirect: '/admin/bots'
    },
    {
      path: '/admin/bots',
      name: 'Bot',
      component: Bot
    }
  ]
})

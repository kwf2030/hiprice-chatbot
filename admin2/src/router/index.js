import Vue from 'vue'
import Router from 'vue-router'
import Bot from '@/components/Bot'

Vue.use(Router)

export default new Router({
  // mode: 'history',
  routes: [
    {
      path: '/bot',
      name: 'Bot',
      component: Bot
    },
    {
      path: '/',
      redirect: '/bot'
    }
  ]
})

import Vue from 'vue'
import Router from 'vue-router'
import WatchList from '@/components/WatchList'
import Remind from '@/components/Remind'
import Settings from '@/components/Settings'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      redirect: '/watchlist'
    },
    {
      path: '/watchlist',
      name: 'WatchList',
      component: WatchList
    },
    {
      path: '/remind',
      name: 'Remind',
      component: Remind
    },
    {
      path: '/settings',
      name: 'Settings',
      component: Settings
    }
  ]
})

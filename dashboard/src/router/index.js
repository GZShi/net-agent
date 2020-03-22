import Vue from 'vue'
import Router from 'vue-router'
import Main from '@/components/dash/Main'
import DashOverview from '@/components/dash/overview/Main'
import DashPortproxy from '@/components/dash/Portproxy'
import DashSocks5 from '@/components/dash/Socks5'
import DashConns from '@/components/dash/Conns'
import Login from '@/components/Login'
import HomePage from '@/components/home/HomePage'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      component: HomePage
    },
    {
      path: '/login',
      component: Login
    },
    {
      path: '/dash',
      component: Main,
      children: [{
        path: 'overview',
        component: DashOverview
      }, {
        path: 'portproxy',
        component: DashPortproxy
      }, {
        path: 'socks5',
        component: DashSocks5
      }, {
        path: 'connections',
        component: DashConns
      }]
    },
    {
      path: '/*',
      redirect: '/'
    }
  ]
})

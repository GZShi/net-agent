import Vue from 'vue'
import Router from 'vue-router'
import HelloWorld from '@/components/HelloWorld'
import ActiveConns from '@/components/ActiveConns'
import HistoryConns from '@/components/HistoryConns'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/active-conns',
      component: ActiveConns
    },
    {
      path: '/history-conns',
      component: HistoryConns
    },
    {
      path: '/*',
      redirect: '/active-conns'
    }
  ]
})

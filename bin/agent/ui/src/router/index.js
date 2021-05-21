import Vue from 'vue'
import Router from 'vue-router'
import Login from '@/components/Login'
import ChatPanel from '@/components/ChatPanel/ChatPanelMain.vue'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/chat-panel',
      name: 'ChatPanel',
      component: ChatPanel
    },
    {
      path: '/',
      name: 'Login',
      component: Login
    }
  ]
})

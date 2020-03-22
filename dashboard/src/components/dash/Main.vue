<template>
  <div id="app">
    <div class="nav">
      <span style="margin-right: var(--gap-size)">Dashboard</span>
      <ul class="menu-container">
        <router-link
          tag="li"
          class="menu-item"
          v-for="(m, mi) in menus"
          :key="mi"
          :to="m.to"
        >{{m.label}}</router-link>
      </ul>
      <span style="float: right;" @click="logined=!logined">logout</span>
    </div>
    <!-- <div class="base-info">
      <base-info></base-info>
    </div> -->
    <div class="router-view">
      <router-view class="router-view"/>
    </div>
  </div>
</template>

<script>
import BaseInfo from './BaseInfo.vue'
import Login from '@/components/Login'
export default {
  name: 'App',
  components: {BaseInfo, Login},
  data() {
    return {
      logined: false,
      menus: [
        { label: '概览', to: '/dash/overview'},
        { label: '端口代理', to: '/dash/portproxy'},
        { label: 'Socks5代理', to: '/dash/socks5'},
        { label: '连接信息', to: '/dash/connections'},
      ],
      activeIndex: 0
    }
  }
}
</script>

<style lang="less">
body {
  padding: 0;
  margin: 0;
}
</style>


<style lang="less" scoped>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  width: 100vw;
  height: 100vh;
  overflow-x: hidden;
  overflow-y: auto;
  padding-top: 50px;
  box-sizing: border-box;

  display: flex;
  flex-direction: column;

  .menu-container {
    display: inline-block;
    margin-left: 3em;
    list-style: none;
    padding: 0;
    margin: 0;

    .menu-item {
      margin: 0;
      padding: 0 1em;
      min-width: 6em;
      text-align: center;
      display: inline-block;
      cursor: pointer;
      background-color: rgba(255,255,255,.03);
      color: #aaa;
      border-bottom:  2px solid rgba(0,0,0,0);
    }
    .menu-item:hover {
      color: #ddd;
    }
    .menu-item.router-link-active {
      color: white;
      background-color: rgba(0,0,0,.28);
      border-color: white;
      cursor: default;
    }
  }

  .base-info {
    height: 160px;
    flex-shrink: 0;
  }
  .router-view {
    flex: 1;
    background-color: #f2f2f2;
  }
}

.nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 9999;
  line-height: 3;
  background-color: #282c34;
  text-shadow: 1px 1px 2px rgba(55,55,55,.3);
  color: #efefef;
  padding: 0 var(--gap-size);

  box-shadow: 0 0 2px black, 0 0 8px rgba(33,24,80,.8);
}

</style>

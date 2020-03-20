<template>
  <div id="app">
    <div class="nav">
      <span>Dashboard of NetAgent</span>
      <div class="menu-container">
        <router-link
          tag="span"
          class="menu-item"
          v-for="(m, mi) in menus"
          :key="mi"
          :to="m.to"
        >{{m.label}}</router-link>
      </div>
    </div>
    <div class="base-info">
      <base-info></base-info>
    </div>
    <div class="router-view">
      <router-view class="router-view"/>
    </div>
  </div>
</template>

<script>
import BaseInfo from '@/components/BaseInfo.vue'
export default {
  name: 'App',
  components: {BaseInfo},
  data() {
    return {
      menus: [
        { label: 'actived', to: '/active-conns'},
        { label: 'history', to: '/history-conns'},
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

    .menu-item {
      cursor: pointer;
      padding: 1em;
      background-color: rgba(0,0,0,.28);
      color: #aaa;
    }
    .menu-item.router-link-active {
      color: white;
      background-color: rgba(255,255,255,.12);
    }
  }

  .base-info {
    height: 160px;
    flex-shrink: 0;
  }
  .router-view {
    flex: 1;
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
  padding: 0 30px;

  box-shadow: 0 0 8px rgba(33,24,193,.8);
}

</style>

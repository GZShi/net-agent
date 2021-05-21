<template>
  <div class="bubble-board-container">
    <div class="header-bar">沟通大厅（212）</div>
    <div class="content-bar">
      <Bubble
        ref="bubbles"
        v-for="(msg, mi) in messages" :key="mi"
        :is-self="msg.sender==myVhost"
        :info="msg.sender"
        :content="msg.content"
      />
    </div>
  </div>
</template>

<script>
import Bubble from './Bubble.vue'
export default {
  components: {Bubble},
  props: ['messages', 'myVhost'],
  watch: {
    messages() {
      if (this.messages && this.messages.length > 0) {
        setTimeout(() => {
          let el = this.$refs.bubbles[this.$refs.bubbles.length-1].$el
          el.scrollIntoView(/*{behavior: 'smooth'}*/)
        }, 5)
      }
    }
  }
}
</script>

<style lang="less" scoped>
.bubble-board-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #fafafa;
  font-size: 16px;

  .header-bar {
    height: 63px;
    padding: 20px 25px;
    border-bottom: 1px solid #ddd;
  }
  .content-bar {
    flex: 1;
    flex-shrink: 0;
    padding: 20px 25px;
    overflow: auto;
    scroll-behavior: smooth;
  }
}
</style>
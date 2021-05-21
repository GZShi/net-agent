<template>
  <div class="bubble-container" :class="{
    right: isSelf,
    min: true,
  }">
    <div class="meta-info">{{info}}</div>
    <div class="content">
      <div class="bubble">{{content}}</div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['isSelf', 'info', 'content']
}
</script>

<style lang="less" scoped>
.bubble-container {
  padding: .4em 0;
  font-size: 14px;
  text-align: left;

  .meta-info {
    display: inline-block;
    text-align: left;
    font-size: 12px;
    color: gray;
    line-height: 20px;
  }

  .bubble {
    display: inline-block;
    text-align: left;
    background-color: white;
    border: 1px solid #ddd;
    border-radius: 5px;
    padding: 5px 10px;
  }
}

.bubble-container.min.right {
  text-align: right;
}

.bubble-container.right {
  .bubble {
    background-color: greenyellow;
    border-color: greenyellow;
  }
}
</style>
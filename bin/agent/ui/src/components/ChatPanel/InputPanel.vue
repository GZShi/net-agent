<template>
  <div class="input-panel-container">
    <div class="tools-bar">Text message input area: </div>
    <div class="input-area">
      <textarea
        v-model="textareaValue"
        @keyup.ctrl.enter="sendGroupMessage"
      ></textarea>
    </div>
    <div class="bottom-actions">
      <div class="left">
        <span>使用Ctrl+Enter发送</span>
      </div>
      <div class="right">
        <span class="error-info">{{sendErrorInfo}}</span>
        <button @click="sendGroupMessage">发送</button>
      </div>
    </div>
  </div>
</template>

<script>
import * as api from '@/lib/api.js'

export default {
  data() {
    return {
      textareaValue: '',
      sendErrorInfo: ''
    }
  },
  methods: {
    async sendGroupMessage() {
      if (/^[\s ]*$/.test(this.textareaValue)) {
        this.setErrInfo('内容不能为空')
        return
      }
      try {
        await api.sendGroupMessage(0, 0, this.textareaValue)
        this.textareaValue = ''
      } catch (ex) {
        this.setErrInfo(`发送失败：${ex}`)
        return
      }
    },
    setErrInfo(err) {
      this.sendErrorInfo = err
      setTimeout(() => this.sendErrorInfo = '', 2000)
    }
  }
}
</script>

<style lang="less" scoped>
.input-panel-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  // background-color: #eee;

  .tools-bar {
    padding: 8px 20px;
  }

  .input-area {
    background-color: #fff;
    flex: 1;
    flex-shrink: 0;
    overflow: hidden;
    padding-left: 20px;

    textarea {
      border: none;
      width: 100%;
      height: 100%;
      resize: none;
      overflow: auto;
      outline: none;
      font-size: 14px;
    }
  }

  .bottom-actions {
    padding: 5px 20px;
    // text-align: right;
    font-size: 12px;
    display: flex;
    flex-direction: row;
    .left {
      flex: 1;
    }

    .error-info {
      margin-right: 0.5em;
      color: #dc6872;
    }
  }
}
</style>
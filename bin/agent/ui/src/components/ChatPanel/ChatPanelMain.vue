<template>
  <div class="chat-panel-container">
    <div class="left-group-list">
      <GroupList :group-list="groupList"/>
    </div>
    <div class="right-chat-board">
      <div class="top-bubble">
        <TalkContentBoard :messages="group.messages" :my-vhost="ctx.vhost"/>
      </div>
      <div class="bottom-input">
        <InputPanel/>
      </div>
    </div>
  </div>
</template>

<script>
import * as api from '@/lib/api.js'
import GroupList from './GroupList.vue'
import TalkContentBoard from './TalkContentBoard.vue'
import InputPanel from './InputPanel.vue'

export default {
  components: {
    GroupList, 
    TalkContentBoard, 
    InputPanel
  },
  data() {
    return {
      groups: {
        '0': {
          id: 0,
          info: {
            label: '沟通大厅',
          },
          messages: []
        }
      },
      activeGroupID: 0,

      ctx: {},
      chatws: null
    }
  },
  computed: {
    groupList() {
      return Object.values(this.groups)
    },
    group() {
      return this.groups[`${this.activeGroupID}`]
    },
    messages() {
      let messages = this.group.messages
      if (messages.length < 300) return messages
      return messages.slice(messages.length - 300)
    }
  },
  mounted() {
    this.loadCtxInfo()
    this.connectChatServer()
    this.loadRecentMessages()
  },
  methods: {
    async loadCtxInfo() {
      this.ctx = await api.loadCtxInfo()
    },
    async loadRecentMessages() {
      let groupMsgMap = await api.loadRecentMessages()
      console.log('group msg map', groupMsgMap)
      // merge messages
      Object.keys(groupMsgMap).map(key => {
        this.groups[key].messages.push(...groupMsgMap[key].msgs)
      })
    },
    async connectChatServer() {
      if (this.chatws !== null) {
        return
      }
      let ws = new WebSocket(`ws://localhost:2021/ws-conn`)
      this.chatws = ws
      ws.onopen = () => {
        console.log('ws connected')
      }
      ws.onerror = (err) => {
        console.log('ws onerror', err)
      }
      ws.onmessage = ({type, data}) => {
        console.log('ws onmessage', type, data)
        let {senderType, sender, groupID, message, msgType} = JSON.parse(data)
        let group = this.groups[`${groupID}`]
        if (!group) {
          console.log('group not found', groupID)
          return
        }
        group.messages.push({
          content: message,
          date: '',
          groupID,
          id: 0,
          sender,
          senderType,
          type: msgType
        })
      }
    }
  }
}
</script>


<style lang="less" scoped>
.chat-panel-container {
  display: flex;
  flex-direction: row;
  width: 100%;
  height: 100%;
  font-size: 14px;

  .left-group-list {
    width: 220px;
    background-color: #eee;
  }
  .right-chat-board {
    flex: 1;
    background-color: #fdfdfd;

    display: flex;
    flex-direction: column;
    height: 100%;

    .top-bubble {
      flex: 1;
      overflow: auto;
    }
    .bottom-input {
      height: 180px;
      border-top: 1px solid #ddd;
    }
  }
}
</style>
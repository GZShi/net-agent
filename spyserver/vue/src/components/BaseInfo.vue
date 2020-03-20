<template>
  <div class="base-info-container">
    <div class="infos" v-if="detail">
      <!-- <div>Upload: {{info.upSize}}bytes / {{info.upPack}}</div>
      <div>Download: {{info.downSize}}bytes / {{info.downPack}}</div> -->
      <div class="item" v-for="(info, i) in infos" :key="i">
        <div class="big-key">{{info.mainTitle}}</div>
        <div class="big-value">{{info.mainValue.value()}}<span class="big-unit">{{info.mainValue.unit()}}</span></div>
        <div class="small">
          <div class="small-container" v-for="(sub, si) in info.subs" :key="si">
            <span class="key">{{sub.key}}</span>
            <span class="value">{{sub.value.str()}}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import * as api from './api.js'
import * as utils from '@/lib/utils.js'
import {Count, Byte, Duration} from './units.js'

export default {
  data() {
    return {
      loop: false,
      infos: [
        {
          mainTitle: '活跃连接数',
          mainValue: new Count(123),
          subs: [
            { key: '已完成', value: new Count(19393) },
            { key: '连接失败', value: new Count(30) }
          ]
        }, {
          mainTitle: '下发速度（Server to Agent）',
          mainValue: new Byte(0),
          subs: [
            { key: '总大小', value: new Byte(0) },
            { key: '数据包个数', value: new Count(19393) }
          ]
        }, {
          mainTitle: '接收速度（Agent to Server）',
          mainValue: new Byte(0),
          subs: [
            { key: '总大小', value: new Byte(0) },
            { key: '数据包个数', value: new Count(19393) }
          ]
        }, {
          mainTitle: '活跃隧道',
          mainValue: new Count(123),
          subs: [
            { key: '服务时间', value: new Count(0) },
          ]
        }
      ],
      detail: null
    }
  },
  watch: {
    detail() {
      if (!this.detail) return
      
      // 活跃连接数
      // 已完成连接数
      // 连接失败的连接数
      this.infos[0].mainValue = new Count(this.detail.activePortCount)
      this.infos[0].subs[0].value = new Count(this.detail.finishedPortCount)
      this.infos[0].subs[1].value = new Count(this.detail.failedPortCount)

      // 下发数据
      // 下发数据包个数
      // 数据包平均大小
      this.infos[1].mainValue = new Byte(this.detail.upSpeed, true)
      this.infos[1].subs[0].value = new Byte(this.detail.upSize)
      this.infos[1].subs[1].value = new Count(this.detail.upPack)

      // 接收数据
      // 接收数据包个数
      // 数据包平均大小
      this.infos[2].mainValue = new Byte(this.detail.downSpeed, true)
      this.infos[2].subs[0].value = new Byte(this.detail.downSize)
      this.infos[2].subs[1].value = new Count(this.detail.downPack)

      // 活跃隧道数量
      this.infos[3].mainValue = new Count(this.detail.tunnelsCount)
      this.infos[3].subs[0].value = new Duration(this.detail.duration)
    }
  },
  async mounted() {
    this.loop = true
    while (this.loop) {
      try {
        this.detail = await api.getBaseInfo()
        await utils.sleep(1000)
      } catch (ex) {
        this.detail = null
        await utils.sleep(1000*5)
      }
    }
  },
  destroyed() {
    this.loop = false
  }
}
</script>


<style lang="less" scoped>
.base-info-container {
  box-sizing: border-box;
  padding: 25px 30px 20px 30px;
  height: 100%;
  overflow: hidden;
  // background-color: #ddd;

  .infos {
    display: flex;
    flex-direction: row;

    .item {
      flex: 1;
      padding: 0 30px;

      .big-key {
        font-size: 14px;
        font-weight: bold;
      }

      .big-value {
        font-size: 60px;
        .big-unit {
          font-size: 20px;
          margin-left: .5em;
          font-style: italic;
        }
      }

      .small {
        margin-top: 10px;

        .small-container {
          display: inline-block;
          font-size: 12px;

          .key {
            font-weight: bold;
          }
          .value {
            margin-right: 1.5em;
          }
        }
      }
    }
  }
}
</style>

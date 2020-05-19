<template>
  <div class="overview-container">
    <div class="actions">
      <button>关闭通道</button>
      <button>限时禁用</button>
      <button>流量限制</button>
      <button>带宽限制</button>
    </div>
    <div class="main-board">
      <div class="left-info">
        <div class="box">
          <div class="box-title">Cluster</div>
          <div class="box-body">
            <table class="multi-line">
              <tr v-for="(info, index) in infos" :key="index">
                <td class="key">{{info.key}}</td>
                <td class="value">
                  <ul v-if="Array.isArray(info.value)">
                    <li v-for="(str, i) in info.value" :key="i">{{str}}</li>
                  </ul>
                  <span v-else>{{info.value}}</span>
                </td>
              </tr>
            </table>
          </div>
        </div>

        <div class="box">
          <div class="box-title">Agents</div>
          <div class="box-body">
            <table class="multi-line">
              <tr v-for="(info, index) in agents" :key="index">
                <td class="key">{{info.key}}</td>
                <td class="value">
                  <ul v-if="Array.isArray(info.value)">
                    <li v-for="(str, i) in info.value" :key="i">{{str}}</li>
                  </ul>
                  <span v-else>{{info.value}}</span>
                </td>
              </tr>
            </table>
          </div>
        </div>
      </div>
      <div class="right-chart">
        <div class="box">
          <div class="box-title">实时带宽</div>
          <line-chart />
        </div>
        <div class="box">
          <div class="box-title">连接数</div>
          <line-chart />
        </div>
        <div class="box">
          <div class="box-title">终端数</div>
          <line-chart />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import LineChart from './Linechart'
import * as api from '@/components/api.js'

export default {
  components: {
    LineChart
  },
  data() {
    return {
      infos: [
        { key: '名称', value: 'garden_city'},
        { key: '状态', value: 'running'},
        { key: '代理数量', value: '5'},
        { key: '上行流量', value: '200 KB'},
        { key: '下行流量', value: '102 MB'},
      ],
      agents: [
        { key: '1', value: 'shiguozhong(172.254.41.23:1234)' },
        { key: '2', value: 'client_2(172.254.41.23:2234)' },
        { key: '3', value: 'testclient(172.254.41.23:2332)' },
        { key: '4', value: 'viva(172.254.41.23:9293)' },
      ],
      rank: {
        conn: [
          { key: '1', value: '172.254.23.1' },
          { key: '2', value: '172.254.23.1' },
          { key: '3', value: '172.254.23.1' },
          { key: '4', value: '172.254.23.1' },
          { key: '5', value: '172.254.23.1' },
        ],
        flow: [
          { key: '1', value: '172.254.23.1' },
          { key: '2', value: '172.254.23.1' },
          { key: '3', value: '172.254.23.1' },
          { key: '4', value: '172.254.23.1' },
          { key: '5', value: '172.254.23.1' },
        ]
      }

    }
  },
  methods: {
    async fetchData() {
      api.getBaseInfo()
    }
  }
}
</script>

<style lang="less" scoped>
.overview-container {
  padding: var(--gap-size);

  .actions {
    margin-bottom: var(--gap-size);
  }

  .main-board {
    display: flex;
    flex-direction: row;

    .left-info {
      flex: 1;
      margin-right: var(--gap-size);
    }
    .right-chart {
      flex: 4;
    }
  }

  .multi-line {
    border-spacing: 1px;
    font-size: 12px;
    td {
      padding: 0;
      padding-bottom: var(--gap-size);
    }

    .key {
      color: gray;
      padding-right: var(--gap-size);
      vertical-align: top;
    }
    .value {
      ul {
        margin: 0;
        padding: 0;
        list-style-type: none;
      }
    }
  }
}
</style>
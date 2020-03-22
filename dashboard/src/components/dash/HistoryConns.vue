<template>
  <div class="active-connections-container">
    <ag-grid-vue
      style="width: 100%; height: 100%"
      class="ag-theme-balham"
      :columnDefs="columnDefs"
      :rowData="rowData"
      :suppressScrollOnNewData="true"
      :suppressColumnVirtualisation="true"
    ></ag-grid-vue>
  </div>
</template>

<script>
import { AgGridVue } from 'ag-grid-vue'
import * as api from '@/components/api.js'
import * as utils from '@/lib/utils.js'

export default {
  name: 'App',
  data() {
    return {
      loop: false,
      columnDefs: [],
      rowData: [],
      existsConnID: {},
    };
  },
  components: {
    AgGridVue
  },
  beforeMount() {
    this.columnDefs = [
      { headerName: "Conn ID", field: 'cid', width: 90 },
      { headerName: 'Source', field: 'sourceAddr', width: 150 },
      { headerName: 'SrcInfo', field: '_ipinfo', width: 130 },
      { headerName: 'Target', field: 'targetAddr', width: 300 },
      { headerName: 'State', field: 'state', width: 100 },
      { headerName: 'During', field: '_alive', width: 90 },
      { headerName: 'Create', field: 'created', width: 160 },
      { headerName: 'Closed', field: 'closed', width: 160 },
      { headerName: 'Tunnel', field: '_tunnelName', width: 100 },
      { headerName: 'User', field: 'uname', width: 80 },
      { headerName: 'Sent', field: '_c2t', width: 140 },
      { headerName: 'Received', field: '_t2c', width: 140 },
    ]

    this.rowData = []
  },
  async mounted() {
    this.loop = true
    while (this.loop) {
      try {
        let conns = await api.getHistoryConns(this)
        conns = conns.filter(conn => {
          let exist = this.existsConnID[conn.cid]
          if (exist) {
            return false
          } else {
            this.existsConnID[conn.cid] = true
            return true
          }
        })
        if (conns.length > 0) {
          console.log('new conns data count', conns.length)
          this.rowData = [...conns, ...this.rowData]
        }
        await utils.sleep(1000*5)
      } catch (ex) {
        this.info = null
        await utils.sleep(1000*10)
      }
    }
  },
  destroyed() {
    this.loop = false
  },
  methods: {

  }
};
</script>


<style lang="less" scoped>
.active-connections-container {
  box-sizing: border-box;
  padding: 30px;
  height: 100%;
}
</style>

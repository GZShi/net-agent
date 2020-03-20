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
import * as api from './api.js'
import * as utils from '@/lib/utils.js'

export default {
  name: 'App',
  data() {
    return {
      loop: false,
      columnDefs: [],
      rowData: []
    };
  },
  components: {
    AgGridVue
  },
  beforeMount() {
    this.columnDefs = [
      { headerName: "Conn ID", field: 'cid', width: 90 },
      { headerName: 'Source', field: 'sourceAddr', width: 190 },
      { headerName: 'Target', field: 'targetAddr', width: 330 },
      { headerName: 'State', field: 'state', width: 120 },
      { headerName: 'Created', field: '_alive', width: 90 },
      { headerName: 'Tunnel', field: '_tunnelName', width: 130 },
      { headerName: 'Sent', field: '_c2t', width: 140 },
      { headerName: 'Received', field: '_t2c', width: 140 },
    ]

    this.rowData = []
  },
  async mounted() {
    this.loop = true
    while (this.loop) {
      try {
        this.rowData = await api.getActiveConns()
        await utils.sleep(1000)
      } catch (ex) {
        this.info = null
        await utils.sleep(1000*5)
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

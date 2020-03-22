<template>
  <div class="portproxy-container">
    <div class="box">
      <div class="box-title">规则列表</div>
      <table>
        <tr>
          <th>监听状态</th>
          <th>监听端口</th>
          <th>目标地址</th>
          <th>目标端口</th>
          <th>允许连接数</th>
          <th>监听时长</th>
          <th>备注</th>
          <th>操作</th>
        </tr>
        <tr v-for="(r, ri) in rules" :key="ri">
          <td>{{r.state}}</td>
          <td>{{r.listen}}</td>
          <td>{{r.target.host}}</td>
          <td>{{r.target.port}}</td>
          <td>{{r.maxConns}}</td>
          <td>{{r.listenDuration}}</td>
          <td>{{r.remark}}</td>
          <td>
            <button>开始监听端口</button>
            <button>设置</button>
            <button>删除</button>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      rules: [
        { state: '关闭', listen: 20001, 
          target: { host: '172.254.41.73', port: 8080},
          maxConns: 5,
          listenDuration: 40,
          remark: '测试机器' },
        { state: '关闭', listen: 20001, 
          target: { host: '172.254.41.73', port: 8080},
          maxConns: 5,
          listenDuration: 40,
          remark: '测试机器' },
        { state: '关闭', listen: 20001, 
          target: { host: '172.254.41.73', port: 8080},
          maxConns: 5,
          listenDuration: 40,
          remark: '测试机器' },
        { state: '关闭', listen: 20001, 
          target: { host: '172.254.41.73', port: 8080},
          maxConns: 5,
          listenDuration: 40,
          remark: '测试机器' },
      ]
    }
  }
}
</script>

<style lang="less" scoped>
.portproxy-container {
  padding: var(--gap-size);

  table {
    width: 100%;
    th {
      text-align: left;
    }
    th, td {
      padding: .3em 0;
    }
  }
}
</style>
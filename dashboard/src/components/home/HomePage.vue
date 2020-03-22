<template>
  <div class="home-container">
    <div class="nav">
      <div class="logo">
        <span class="icon"><span class="circle">RW</span></span>
        <span class="remote label">Remote</span><span class="work label">Worker</span>
      </div>
      <ul class="menu">
        <li>产品</li>
        <li>解决方案</li>
        <li>资源</li>
        <li>集成</li>
        <li>合作伙伴</li>
        <li>公司简介</li>
      </ul>
    </div>

    <div class="content big-image">
      <div class="title">远程控制和远程支持领域的新标准</div>
      <div class="subtitle">我们为您提供领先的一体化解决方案。今天就来购买或升级吧！个人使用免费。</div>
      <div class="mode-list">
        <div class="mode server">
          <div class="label">服务器模式</div>
          <div class="desc">数据流动的枢纽</div>
        </div>
        <div class="mode agent">
          <div class="label">代理端模式</div>
          <div class="desc">等待控制的终端</div>
        </div>
        <div class="mode visitor">
          <div class="label">访问者模式</div>
          <div class="desc">控制器终端</div>
        </div>
      </div>
      <div class="learn-more">
        <a href="https://github.com/GZShi/net-agent" target="_blank" style="color: white !important;">在GitHub上了解更多信息</a>
      </div>
    </div>

    <div class="content use-step">
      <div>第一步：下载指定版本程序</div>
      <div class="download">
        <button @click="downloadBin('windows', 'x64', '.exe')">下载Windows(x64)版本程序</button>
      </div>
      <div class="download">
        <div class="build-with-src">
          <span @click="showBuildTips=!showBuildTips">Linux and Mac downloads?</span>
          <div class="build-tips" :class="showBuildTips?'show':'hide'">
            <div>使用Golang工具进行构建跨平台程序</div>
            <div><code>go get github.com/GZShi/net-agent/exec</code></div>
          </div>
        </div>
      </div>
    </div>

    <div class="content use-step dark">
      <div>第二步：设置终端配置，然后运行程序</div>

      <div class="flex-row-container" style="text-align: center; margin: 1em 0;">
        <div class="flex-1">服务器配置</div>
        <div class="flex-1">代理端配置</div>
        <div class="flex-1">访问者配置</div>
      </div>

      <div class="flex-row-container">
        <div class="flex-1">
          <div class="form-line">
            <label class="label">服务器域名或地址</label>
            <input type="text" placeholder="示例：mysite.com" v-model="config.serverHost">
          </div>
          <div class="form-line">
            <label class="label">监听端口</label>
            <input type="text" placeholder="示例：1080" v-model="config.serverPort">
          </div>
          <div class="form-line">
            <label class="label">连接密码 <span class="text-btn" @click="randPrivateKey">随机生成</span></label>
            <input type="text" placeholder="" v-model="config.privateKey" style="font-family: monospace">
            <div class="alertinfo">密码请妥善保管，勿轻易泄露</div>
          </div>
        </div>
        <div class="flex-1">
          <div class="form-line">
            <label class="label">你的代号</label>
            <input type="text" v-model="config.clientName">
          </div>
          <div class="form-line">
            <label class="label">局域网代号</label>
            <input type="text" v-model="config.channelName">
            <div class="alertinfo">代号信息请勿轻易泄露</div>
          </div>
        </div>
        <div class="flex-1">
          <div class="form-line">
            <label class="label">本地端口</label>
            <input type="text" placeholder="示例：13389" v-model="config.portproxy.port">
          </div>
          <div class="form-line">
            <label class="label">目标地址</label>
            <input type="text" placeholder="示例：10.254.1.1:3389" v-model="config.portproxy.targetAddr">
          </div>
        </div>
      </div>

      <div class="flex-row-container">
        <div class="flex-1 editor-container">
          <div class="editor">
            <pre>{{buildConfig('server')}}</pre>
          </div>
        </div>
        <div class="flex-1 editor-container">
          <div class="editor">
            <pre>{{buildConfig('agent')}}</pre>
          </div>
        </div>
        <div class="flex-1 editor-container">
          <div class="editor">
            <pre>{{buildConfig('visitor')}}</pre>
          </div>
        </div>
      </div>

      <div class="flex-row-container">
        <div class="flex-1">
          <div class="download">
            <button @click="downloadConfigFile('server')">下载服务器配置</button>
          </div>
        </div>
        <div class="flex-1">
          <div class="download">
            <button @click="downloadConfigFile('agent')">下载代理端配置</button>
          </div>
        </div>
        <div class="flex-1">
          <div class="download">
            <button @click="downloadConfigFile('visitor')">下载访问者配置</button>
          </div>
        </div>
      </div>
    </div>

    <div class="content use-step">
      <div>第三步：访问者打开真正的远程控制工具</div>
      <div class="mstsc">
        <img src="../../assets/mstsc.png" alt="">
        <div class="address-cover">localhost:{{config.portproxy.port}}</div>
      </div>
    </div>

    <div class="content bottom-info">
      <b>Copyrighter @ 2020 RemoteWorker</b>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      showBuildTips: false,
      config: {
        serverHost: 'localhost',
        serverPort: '1080',
        privateKey: 'randrand',
        clientName: 'default-client',
        channelName: 'remotework',
        portproxy: {
          port: '13389',
          targetAddr: '10.254.1.1:3389'
        }
      }
    }
  },
  methods: {
    buildConfig(mode) {
      switch(mode) {
        case 'server':
          return JSON.stringify({
            mode: 'server',
            addr: `0.0.0.0:${this.config.serverPort}`,
            privateKey: this.config.privateKey
          }, null, 2)
        case 'agent':
          return JSON.stringify({
            mode: 'agent',
            addr: `${this.config.serverHost}:${this.config.serverPort}`,
            privateKey: this.config.privateKey,
            clientName: this.config.clientName,
            channelName: this.config.channelName
          }, null, 2)
        case 'visitor':
          return JSON.stringify({
            mode: 'visitor',
            addr: `${this.config.serverHost}:${this.config.serverPort}`,
            privateKey: this.config.privateKey,
            clientName: this.config.clientName,
            channelName: this.config.channelName,
            portproxy: [{
              listen: `localhost:${this.config.portproxy.port}`,
              targetAddr: this.config.portproxy.targetAddr
            }]
          }, null, 2)
      }
    },
    downloadConfigFile(mode) {
      let content = this.buildConfig(mode)
      let btn = document.createElement('a')
      btn.href = 'data:application/json;charset=utf-8,' + encodeURI(content)
      btn.target = '_blank'
      btn.download = 'config.json'
      btn.click()
    },
    downloadBin(platform, arch, ext) {
      let fileName = `remotework-${platform}-${arch}${ext}`
      let btn = document.createElement('a')
      btn.href = `/download/${fileName}`
      btn.target = '_blank'
      btn.download = fileName
      btn.click()
    },
    randPrivateKey() {
      let tokens = [
        'abcdefghijklmnopqrstuvwxyz'.split(''),
        'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split(''),
        '01234567890'.split(''),
        '!@#$%^&()+='.split(''),
      ]
      let pswdstruct = '0001112233'.split('') // 3个小写字母，3个大写字母，2个数字，2个字符
      pswdstruct = shuffle(pswdstruct)

      let pswd = pswdstruct.map(tokenType => {
        tokens[tokenType] = shuffle(tokens[tokenType])
        return tokens[tokenType][0]
      })
      
      this.config.privateKey = pswd.join('')

      function shuffle(array) {
        var m = array.length, t, i

        // While there remain elements to shuffle…
        while (m) {

          // Pick a remaining element…
          i = Math.floor(Math.random() * m--)

          // And swap it with the current element.
          t = array[m]
          array[m] = array[i]
          array[i] = t
        }

        return array
      }
    }
  }
}
</script>

<style lang="less" scoped>
.home-container {
  position: relative;
  font-family: 'Microsoft Yahei Light', Roboto, Helvetica, Arial, sans-serif;
  .nav {
    position: sticky;
    top: 0;
    background-color: white;
    color: #4a4a4a;
    padding: 15px;
    box-shadow: 0 3px 6px 0 rgba(0,0,0,.15);

    div.logo {
      display: inline-block;
      margin: 0 20px;
      .icon {
        display: inline-block;
        width: 35px;
        height: 35px;
        border-radius: 3px;
        background-color: #038FF4;
        vertical-align: middle;
        text-align: center;

        .circle {
          display: inline-block;
          background-color: white;
          margin-top: 3px;
          width: 29px;
          height: 29px;
          border-radius: 50%;
          line-height: 29px;
          font-size: 11px;
          color: #0B77D8;
          font-weight: bold;
        }
      }
      .label {
        font-size: 18px;
        font-family: Arial, Helvetica, sans-serif;
        color: #0B77D8;
        vertical-align: middle;
      }
      .remote {
        font-weight: bold;
      }
    }

    ul.menu {
      display: inline-block;
      list-style-type: none;
      padding: 0;
      margin: 0;
      li {
        display: inline-block;
        margin: .5em;
      }
    }
  }
  .content {
    padding-left: 8%;
    padding-right: 8%;
  }
  .use-step {
    font-size: 20px;
    padding-top: 50px;
    padding-bottom: 50px;
  }
  .use-step.dark {
    background-color: #F7F7F8;
  }
  .big-image {
    padding-top: 180px;
    padding-bottom: 180px;
    background-color: #2C77B0;
    color: white;
    font-weight: normal;
    font-size: 20px;

    .title {
      font-size: 48px;
    }
    .subtitle {
      font-size: 20px;
      margin: 1em 0;
    }
    .mode-list {
      padding: 20px 0;
      .mode {
        font-size: 24px;
        display: inline-block;
        background-color: rgba(255,255,255,.2);
        border: 2px solid white;
        padding: .8em 1.2em;
        margin-right: 2em;
        border-radius: 30px;

        .label {
          font-size: 24px;
        }
        .desc {
          font-size: 14px;
          line-height: 18px;
        }
      }
      .mode.server {
        background-color: rgba(150, 195, 34, 0.3);
        border-color: #96C322;
        color: #afec11;
      }
      .mode.visitor {
        background-color: rgba(254, 149, 57, 0.4);
        border-color: #ffb472;
        color: #ffb472;
      }
    }
  }

  .download {
    text-align: center;
    margin-top: 1em;
    
    button {
      font-size: 18px;
      padding: 1em 1.5em;
      border: none;
      background-color: #96C322;
      color: white;
      border-radius: 100px;
    }

    .build-with-src {
      font-size: 14px;
      span {
        text-decoration: underline;
        cursor: pointer;
      }

      .build-tips {
        margin-top: 20px;
        transition: .3s;
        overflow: hidden;
      }
      .build-tips.hide {
        height: 0;
      }
      .build-tips.show {
        height: 40px;
      }
    }
  }

  .flex-row-container {
    display: flex;
    flex-direction: row;
    .flex-1 {
      flex: 1
    }
    .editor-container {
      padding: 0 5px;
      text-align: center;
      .editor {
        height: 100%;;
        text-align: left;
        margin: 0 auto;
        max-width: 450px;
        background-color: #fff;
        padding: 1em;
        font-size: 14px;
      }
    }
  }

  .form-line {
    font-size: 14px;
    margin-bottom: 10px;
    margin: 0 auto;
    margin-bottom: 1em;
    width: 60%;
    font-family: serif;

    label {
      display: block;
      min-width: 5em;
      line-height: 1.8;
      padding-left: 10px;

      .text-btn {
        color: #0B77D8;
        cursor: pointer;
      }
    }
    input {
      display: block;
      width: 100%;
      line-height: 22px;
      padding: 12px 10px;
      background-color: #fff;
      border: none;
      border-bottom: 1px solid #ddd;
      font-family: 'Microsoft Yahei', monospace;
    }
    .alertinfo {
      padding-left: 10px;
      padding-top: 5px;
      color: #b91815;
    }
  }

  .mstsc {
    position: relative;
    .address-cover {
      position: absolute;
      left: 133px;
      top: 167px;
      line-height: 22px;
      width: 255px;
      padding: 0 5px;
      background-color: #ffffcc;
      font-size: 12px;
      font-family: 'Microsoft Yahei';
    }
  }

  .bottom-info {
    background-color: #253B4D;
    color: #b7b7b7;
    padding-top: 40px;
    padding-bottom: 40px;
    text-align: center;
    font-size: 12px;
  }
}
</style>
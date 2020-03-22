<template>
  <div class="login-container">
    <div class="panel">
      <div class="header">
        <div>🤐接入网络管道🤐</div>
      </div>
      <div class="body">
        <div class="input-form">
          <label for="plhd-1">账号</label>
          <input type="text" v-model="username" name="plhd-1" placeholder="zhangsan">
        </div>
        <div class="input-form">
          <label for="plhd-2">网络代号</label>
          <input type="text" v-model="netname" name="plhd-2" placeholder="dev_network">
        </div>
        <div class="input-form">
          <label for="plhd-3">验证码</label>
          <input type="text" v-model="token" name="plhd-3" placeholder="123456">
        </div>
      </div>
      <div class="footer">
        <button @click="login">验证身份</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      username: '',
      netname: '',
      token: '',
    }
  },
  methods: {
    login() {
      alert('登录成功')
      let redirect = this.$route.query.redirect || ''
      if (redirect) {
        this.$router.push(redirect)
      } else {
        this.$router.push('/dash')
      }
    }
  }
}
</script>

<style lang="less" scoped>

.login-container {
  padding: var(--gap-size);
  text-align: center;
  .panel {
    max-width: 500px;
    margin: 40px auto;
    text-align: left;
    border-radius: 3px;
    box-shadow: 0 0 6px 0 rgba(163, 163, 163, 0.5);
    padding: var(--gap-size);

    .header {
      padding: var(--gap-size) 0;
      font-size: 18px;
    }

    .body {
      padding: var(--gap-size) 0;
      border-top: .5px solid var(--main-color);
      // border-bottom: .5px solid var(--main-color);
      .input-form {
        margin-bottom: var(--gap-size);
        label {
          line-height: 28px;
        }

        input {
          width: 100%;
          font-size: 24px;
          padding: .4em;

          border: var(--input-border);
          border-radius: 6px;
        }
      }
    }
    .footer {
      button {
        display: block;
        width: 100%;
        font-size: 24px;
        padding: .4em;
        border-radius: 6px;
        border: none;
        background-image: var(--button-bgimg);
        color: #eee;
      }
    }
  }
}
</style>
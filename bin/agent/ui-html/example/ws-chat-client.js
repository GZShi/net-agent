;(function () {
  let ws = new WebSocket(`ws://${location.host}/ws-conn`)
  ws.onopen = function () {
    console.log('ws connected')
  }
  ws.onerror = function (err) {
    console.log('ws onerror', err)
  }
  ws.onmessage = function ({type, data}) {
    console.log('new message', type, data)
    let msg = JSON.parse(data)
    appendMessage(msg)
  }
})()
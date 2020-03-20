import {get, post} from '@/lib/http.js'
import {Count, Byte} from './units.js'
import { getIpInfo } from './ipinfo.js'

let lastCall = {
  tick: new Date(0),
  upSize: 0,
  downSize: 0
}
export async function getBaseInfo(calcSum=true) {
  let {now, tunnels} = await get('/gss/admin/base-info')

  if (!calcSum) return tunnels

  let sum = {
    tick: new Date(now),
    upSize: 0,
    downSize: 0,
    upPack: 0,
    downPack: 0,
    tunnelsCount: 0,
    activePortCount: 0,
    finishedPortCount: 0,
    failedPortCount: 0,
    duration: '--'
  }
  tunnels.forEach(t => {
    sum.upSize += t.upSize
    sum.downSize += t.downSize
    sum.upPack += t.upPack
    sum.downPack += t.downPack

    sum.tunnelsCount += 1
    sum.activePortCount += t.activePortCount
    sum.finishedPortCount += t.finishedPortCount
    sum.failedPortCount += t.failedPortCount

    sum.duration = t.created.split('.')[0]
  })

  let f = 1000 / (sum.tick.getTime() - lastCall.tick.getTime())
  sum.upSpeed = (sum.upSize - lastCall.upSize) *f
  sum.downSpeed = (sum.downSize - lastCall.downSize) *f
  lastCall = sum

  return sum
}


function duration(before, after) {
  let n = after.getTime() - before.getTime()
  let date = new Date(n - 8*60*60*1000)
  let day = date.getDate() - 1
  let hour = String(date.getHours()).padStart(2, '0')
  let minute = String(date.getMinutes()).padStart(2, '0')
  let second = String(date.getSeconds()).padStart(2, '0')

  if (day > 0) {
    return `${day}:${hour}:${minute}:${second}`
  }
  return `${hour}:${minute}:${second}`
}

export async function getActiveConns(flat=true) {
  let {now, tunnels} = await get('/gss/admin/active-conns')

  if (!flat) return tunnels

  let allConns = []
  tunnels.forEach(({name, conns}) => {
    conns.forEach(c => {
      c._tunnelName = name
      c._c2t = (new Byte(c.c2t)).str()
      c._t2c = (new Byte(c.t2c)).str()
      c._alive = duration(new Date(c.created), new Date(now))
      allConns.push(c)
    })
  })
  return allConns.sort((a, b) => a.cid - b.cid)
}

export async function queryIpInfo(ip) {
  try {
    let resp = await get('ip-api/service/getIpInfo.php', {ip})
    let json = await resp.json()
    if (json.code == 0) {
      return `${json.data.country}${json.data.city}(${json.data.isp})`
    }
    return '--'
  } catch(ex) {
    return 'err'
  }
}

export async function getHistoryConns(vm, flat=true) {
  let {now, tunnels} = await get('/gss/admin/history-conns')

  if (!flat) return tunnels

  let allConns = []
  tunnels.forEach(({name, conns}) => {
    conns.forEach(c => {
      c._tunnelName = name
      c._c2t = (new Byte(c.c2t)).str()
      c._t2c = (new Byte(c.t2c)).str()
      c._closed = new Date(c.closed)
      c._alive = duration(new Date(c.created), c._closed)
      c._ipinfo = '查询中'
      allConns.push(c)

      getIpInfo((c.sourceAddr || '').split(':')[0]).then(info => {
        // c._ipinfo = info
        vm.$set(c, '_ipinfo', info)
      }, reason => {
        vm.$set(c, '_ipinfo', reason)
      })
    })
  })

  return allConns.sort((a, b) => b._closed - a._closed)
}
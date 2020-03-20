import * as utils from './utils.js'

const headers = {}

function parseResponse(res, path) {
  return res
    .then(res => {
      if (res.status != 200) {
        throw `网络错误：${res.statusText}\n${path}`
      }
      return res
    })
    .then(res => res.json())
    .then(({ ErrCode, ErrMsg, Data }) => {
      if (ErrCode != 0) throw ErrMsg || `ErrCode is ${ErrCode}`
      return Data
    })
}

export function get(path, query = {}) {
  let resp = fetch(`${path}?${utils.querify(query)}`, {
    credentials: 'same-origin',
    headers
  })

  return parseResponse(resp, path)
}

export function post(path, json, query = {}) {
  let resp = fetch(`${path}?${utils.querify(query)}`, {
    method: 'POST',
    body: JSON.stringify(json),
    credentials: 'same-origin',
    headers
  })

  return parseResponse(resp, path)
}

import * as utils from './utils.js'

const headers = {}

export async function parseResponse(res, path) {
  if (res.status != 200) {
    throw `http_error::${res.statusText}(${res.status})\n${path}`
  }

  let json
  try {
    json = await res.json()
  } catch (ex) {
    throw `parse body as json failed`
  }

  if ('ErrCode' in json) {
    let {ErrCode, ErrMsg, Data} = json
    if (ErrCode != 0) throw ErrMsg || `normal_error::${ErrCode}`
    return Data
  } else if ('errcode' in json) {
    // 兼容小写errcode的情况
    // 语音助手的API有这个情况
    let {errcode, errmsg, data} = json
    if (errcode != 0) throw errmsg || `normal_error::${errcode}`
    return data
  }
  
  throw `unknown error code`
}

export async function get(path, query = {}) {
  let resp = await fetch(`${path}?${utils.querify(query)}`, {
    credentials: 'same-origin',
    headers
  })

  return parseResponse(resp, path)
}

export async function post(path, json, query = {}) {
  let resp = await fetch(`${path}?${utils.querify(query)}`, {
    method: 'POST',
    body: JSON.stringify(json),
    credentials: 'same-origin',
    headers
  })

  return parseResponse(resp, path)
}

export async function postform(path, json, query={}) {
  let querystr = utils.querify(query)
  if (querystr) {
    path = path + '?' + querystr
  }

  let form = new FormData()
  Object.keys(json).forEach(key => {
    form.append(key, json[key])
  })

  let resp = await fetch(path, {
    method: 'POST',
    body: form
  })

  return parseResponse(resp, path)
}
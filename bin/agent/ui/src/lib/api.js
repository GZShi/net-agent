import {get, post} from './http.js'

const prefix = '/agent-api'

export async function loadCtxInfo() {
  return get(`${prefix}/ctx-info`)
}

export async function sendGroupMessage(groupID, msgType, message) {
  return post(`${prefix}/new-message`, {
    groupID, msgType, message
  })
}

export async  function loadRecentMessages() {
  let list = await post(`${prefix}/recent-message`, {})
  let groupMap = {}
  list.forEach(msg => {
    if (!groupMap[msg.groupID]) {
      groupMap[msg.groupID] = {id: msg.groupID, msgs: []}
    }
    groupMap[msg.groupID].msgs.push(msg)
  })

  return groupMap
}
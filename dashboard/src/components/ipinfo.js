let cacheStorage = {}

export function getIpInfo(ip) {
  return new Promise(async (resolve, reject) => {

    // 首先，查询缓存是否存在
    let cache = cacheStorage[ip]
    if (cache) {
      // 如果缓存存在，且数据已经获取完毕，则直接回调数据
      // 如果缓存数据显示加载失败，则进行失败回调
      if (cache.state == 'success') {
        return resolve(cache.data)
      } else if (cache.state == 'failed') {
        return reject(cache.data)
      }

      // 如果缓存还未加载完毕，则将回到函数进行缓存
      cache.callbacks.push([resolve, reject])
      return
    }

    // 如果缓存不存在，则创建缓存。存储缓存并将缓存状态置为pending
    cache = { state: 'pending', callbacks: [[resolve, reject]], data: 'null' }
    cacheStorage[ip] = cache
  

    try {
      // 调用淘宝的api获取信息
      let res = await fetch(`/taobaoip/service/getIpInfo.php?ip=${encodeURIComponent(ip)}`)
      if (res.status != 200) throw res.statusText
    
      let json = await res.json()
      if (json.code != 0) throw json
    
      let { country, region, city, isp } = json.data
    
      cache.data = `${country}-${region + city}-${isp}`
      cache.state = 'success'
    } catch (ex) {
      cache.state = 'failed'
      cache.data = String(ex)
    }
    
    // 统一对缓存的回调函数进行调用
    if (cache.state == 'success') {
      cache.callbacks.forEach(cb => setTimeout(cb[0], 0, cache.data))
    } else {
      cache.callbacks.forEach(cb => setTimeout(cb[1], 0, cache.data))
    }
  })
}

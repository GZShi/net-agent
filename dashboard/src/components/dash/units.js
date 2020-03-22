export class Byte {
  constructor(n, isSpeed=false) {
    this.isSpeed = isSpeed
    this.count = n
    this.calc()
  }
  value() { return this.v }
  unit() { return this.u }
  str() { return `${this.v} ${this.u}` }

  calc() {
    let units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    let n = this.count
    let index = 0

    while (n >= 1024) {
      n = n / 1024
      index++
    }

    if (this.isSpeed) {
      this.v = n.toFixed(2)
    } else {
      this.v = index > 0 ? n.toFixed(2) : String(n)
    }

    if (this.isSpeed) {
      this.u = `${units[index]}/s`
    } else {
      this.u = units[index]
    }
  }
}

// Count 单位：个
export class Count {
  constructor(n) {
    this.count = n
    this.calc()
  }

  value() { return this.v }
  unit() { return this.u }
  str() { return `${this.v}${this.u}` }

  calc() {
    if (this.count > 10000) {
      this.v = (this.count / 10000).toFixed(2)
      this.u = 'w'
      return
    }
    // if (this.count > 1000) {
    //   this.v = (this.count / 1000).toFixed(3)
    //   this.u = 'k'
    //   return
    // }
    this.v = this.count
    this.u = ''
  }
}

export class Duration {
  constructor(n) {
    this.count = n
    this.calc()
  }
  calc() {
    this.days = ''
    this.hours = '00'
    this.mins = '00'
    this.seconds = '00'
  }

  str() {
    return this.count
    if (this.days) {
      return `${this.days}:${this.hours}:${this.mins}:${this.seconds}`
    }
    return `${this.hours}:${this.mins}:${this.seconds}`
  }
}
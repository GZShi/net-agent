<template>
  <div class="line-chart-container">
    <svg ref="svg">
      <text text="haha" x="100px" y="50px">line chart container</text>
    </svg>
  </div>
</template>

<script>
import * as d3 from 'd3'

function randArr(length, min, max) {
  if (length < 0) length = 0

  let arr = []
  for (let i = 0; i < length; i++) {
    arr.push(Math.random() * (max-min) + min)
  }
  
  return arr
}

export default {
  data() {
    return {
      svg: null,
      root: null,
      width: 0,
      height: 0,
      margin: {
        top: 0, right: 0, bottom: 0, left: 0
      },

      ticks: []
    }
  },
  computed: {
    w() { return this.width - this.margin.left - this.margin.right },
    h() { return this.height - this.margin.top - this.margin.bottom }
  },
  mounted() {
    this.svg = this.$refs.svg

    // calc size
    let style = getComputedStyle(this.svg)
    this.width = parseInt(style.width, 10)
    this.height = parseInt(style.height, 10)

    this.root = d3.select(this.svg).append('g')
      .attr('transform', `translate(${this.margin.top}, ${this.margin.left})`)

    this.initDraw()
    this.updateAllTicks(randArr(60, 0, 100))

    setInterval(() => {
      this.appendTicks(randArr(1, 0, 100))
    }, 2200)
  },
  methods: {
    // 在svg上绘制必要的元素
    initDraw() {
      this.root.append('path').attr('class', 'line-a')
        .attr('stroke', '#666')
        .attr('fill', 'none')
    },
    // 根据数据重绘所有数据
    updateAllTicks(ticks) {
      this.ticks = ticks
      let x = d3.scaleLinear().domain([0, ticks.length-1]).range([0, this.w])
      let y = d3.scaleLinear().domain([0, 100]).range([this.h, 0])
      let l = d3.line()
        .x((d, i) => x(i))
        .y(d => y(d))

      this.root.select('path.line-a').datum(ticks)
        .attr('d', l)
    },
    // 根据追加数据重绘
    appendTicks(ticks) {
      // ticks = this.ticks = [...this.ticks, ...ticks].slice(ticks.length)
      // this.updateAllTicks(ticks)
      let all = [...this.ticks, ...ticks]
      let x = d3.scaleLinear().domain([0, all.length-1]).range([0, this.w * all.length/60])
      let y = d3.scaleLinear().domain([0, 100]).range([this.h, 0])
      let l = d3.line()
        .x((d, i) => x(i))
        .y(d => y(d))
      
      this.root.select('path.line-a').datum(all)
        .attr('d', l)
        .attr('transform', 'translate(0,0)')
        .transition().duration(2000)
        .attr('transform', `translate(${-ticks.length*this.w/60}, 0)`)

      this.ticks = all.slice(ticks.length)
    }
  }
}
</script>

<style lang="less">
.line-chart-container {
  height: 150px;
  svg {
    width: 100%;
    height: 100%;
    background-color: #f2f2f2;
  }
}
</style>
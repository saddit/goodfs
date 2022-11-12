<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ $t('overview') }}</div>
  <div
      class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 justify-items-center placeholder:py-2">
    <ServerCard v-for="info in infos" :info="info"></ServerCard>
  </div>
  <div class="mt-8 text-2xl text-gray-900 font-bold mb-4">{{ $t('monitor') }}</div>
  <div id="cpu-usage" ref="cpuChart"
       class="bg-white rounded-2xl border border-gray-200 outline outline-3 outline-offset-4 outline-indigo-600 h-56">
  </div>
  <div class="inline-flex justify-center text-sm text-indigo-600 font-medium pt-2.5 w-full">{{ $t('cpu-usage') }}</div>
  <div id="mem-chart" ref="memChart"
       class="mt-8 bg-white rounded-2xl border border-gray-200 outline outline-3 outline-offset-4 outline-indigo-600 w-full h-56">
  </div>
  <div class="inline-flex justify-center text-sm text-indigo-600 font-medium pt-2.5 w-full">{{ $t('mem-usage') }}</div>
</template>

<script setup lang="ts">
import * as echarts from 'echarts';
import {rand} from "@vueuse/core";

type EChartsOption = echarts.EChartsOption;

const cpuChart = ref()
const memChart = ref()
const infos = ref<ServerInfo[]>([])

onBeforeMount(() => {
  let stats =  useStore().serverStat.apiServer
  for (let k in stats) {
    infos.value.push(stats[k])
  }
})

//TODO: monitor sample
onMounted(() => {
  let cpuEchart = echarts.init(cpuChart.value)
  let memEchart = echarts.init(memChart.value)
  window.addEventListener("resize", () => {
    cpuEchart.resize()
    memEchart.resize()
  })
  let option: EChartsOption

  const data = [["2000-06-05", 20], ["2000-06-06", 18], ["2000-06-07", 30], ["2000-06-08", 40], ["2000-06-09", 15], ["2000-06-10", 18], ["2000-06-11", 100], ["2000-06-12", 100], ["2000-06-13", 99], ["2000-06-14", 80], ["2000-06-15", 70], ["2000-06-16", 63], ["2000-06-17", 64], ["2000-06-18", 70], ["2000-06-19", 100], ["2000-06-20", 70], ["2000-06-21", 40], ["2000-06-22", 60], ["2000-06-23", 63], ["2000-06-24", 52], ["2000-06-25", 71], ["2000-06-26", 80], ["2000-06-27", 48], ["2000-06-28", 43], ["2000-06-29", 32], ["2000-06-30", 22], ["2000-07-01", 12], ["2000-07-02", 8], ["2000-07-03", 5], ["2000-07-04", 6], ["2000-07-05", 19], ["2000-07-06", 20], ["2000-07-07", 14], ["2000-07-08", 15], ["2000-07-09", 17], ["2000-07-10", 18], ["2000-07-11", 32], ["2000-07-12", 21], ["2000-07-13", 26], ["2000-07-14", 17], ["2000-07-15", 28], ["2000-07-16", 30], ["2000-07-17", 32], ["2000-07-18", 88], ["2000-07-19", 77], ["2000-07-20", 83], ["2000-07-21", 100], ["2000-07-22", 57], ["2000-07-23", 55], ["2000-07-24", 60]];
  var dateList = data.map(function (item) {
    return item[0]
  })
  var valueList = data.map(function (item) {
    return item[1]
  })

  option = {
    // Make gradient line here
    visualMap: [
      {
        show: false,
        type: 'continuous',
        seriesIndex: 0,
        min: 0,
        max: 100,
        inRange: {
          color: ['#c7d2fe', '#4f46e5']
        },
      },
    ],
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dateList,
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: 'percent',
        data: valueList,
        type: 'line',
        areaStyle: {}
      }
    ],
    grid: {
      left: 46,
      top: 40,
      right: 46,
      bottom: 40
    }
  }

  option && cpuEchart.setOption(option)
  option && memEchart.setOption(option)

  let cnt = 0

  let maxLen = dateList.length

  setInterval(() => {
    let now = useNow().value
    now.setDate(now.getDate() + cnt++)
    let s = [now.getFullYear(), now.getMonth(), now.getDate()].join('-');
    if (dateList.length >= maxLen) {
      let start = dateList.length - maxLen + 1
      dateList = dateList.slice(start)
      valueList = valueList.slice(start)
    }
    dateList.push(s)
    valueList.push(rand(0, 100))

    cpuEchart.setOption({
      xAxis: {
        data: dateList,
      },
      series: [
        {
          data: valueList,
          name: 'percent'
        }
      ],
    })
  }, 30000)
})
</script>

<style scoped>

</style>

<route lang="json">
{
  "meta": {
    "title": "api-server"
  }
}
</route>
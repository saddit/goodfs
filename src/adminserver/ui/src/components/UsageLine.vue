<template>
  <div class="bg-white shadow-lg rounded-2xl">
    <div class="inline-flex justify-center text-sm text-indigo-600 font-bold pt-2 w-full">{{ $t(`${type}-usage`) }}
    </div>
    <div :class="h" ref="usageChart"></div>
  </div>
</template>

<script setup lang="ts">
import * as echarts from 'echarts'

const prop = defineProps<{
    type: 'cpu' | 'mem'
    tl: Record<string, TimeStat[]>
    h: string
}>()

function fetchTl(): [string[], number[]] {
    let dateList: string[] = []
    let valueList: number[] = []
    for (let k in prop.tl) {
        let v = prop.tl[k]
        for (let i = 0; i < v.length; i++) {
            let date = new Date(v[i].time)
            let hh = date.getHours().toString()
            if (date.getHours() < 10) {
                hh = '0' + hh
            }
            let mm = date.getMinutes().toString()
            if (date.getMinutes() < 10) {
                mm = '0' + mm
            }
            if (valueList.length <= i) {
                dateList.push(`${hh}:${mm}`)
                valueList.push(v[i].percent * 100)
            } else {
                valueList[i] = (valueList[i] + v[i].percent * 100) / 2
            }
        }
    }
    return [dateList, valueList]
}

const usageChart = ref()

onMounted(() => {
    let chart = echarts.init(usageChart.value)
    watch(() => prop.tl, () => {
        let res = fetchTl()
        chart.setOption({
            xAxis: {
                data: res[0]
            },
            series: [{
                name: 'percent',
                data: res[1]
            }]
        })
    })
    window.addEventListener("resize", () => {
        chart.resize()
    })
    chart.setOption({
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
        tooltip: {
            trigger: 'axis',
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: [],
        },
        yAxis: {
            type: 'value',
            axisLabel: {
                formatter: '{value} %'
            }
        },
        series: [
            {
                name: 'percent',
                data: [],
                type: 'line',
                areaStyle: {},
                tooltip: {
                    valueFormatter: (value: any) => value + ' %'
                },
                showSymbol: false
            }
        ],
        grid: {
            left: 50,
            top: 20,
            right: 46,
            bottom: 40
        }
    })
})
</script>
<template>
	<div>
		<div ref="usageChart"
			class="bg-white rounded-2xl border border-gray-200 outline outline-3 outline-offset-4 outline-indigo-600 h-56">
		</div>
		<div class="inline-flex justify-center text-sm text-indigo-600 font-medium pt-2.5 w-full">{{ $t(`${type}-usage`) }}
		</div>
	</div>
</template>

<script setup lang="ts">
import * as echarts from 'echarts'

const prop = defineProps<{
	type: 'cpu' | 'mem'
	serverNo: 0 | 1 | 2
}>()

async function fetchTl(): Promise<[string[], number[]]> {
	let res = await api.serverStat.timeline(prop.serverNo, prop.type)
	let dateList: string[] = []
	let valueList: number[] = []
	for (let k in res) {
		let v = res[k]
		for (let i = 0; i < v.length; i++) {
			let date = new Date(v[i].time)
			if (valueList.length <= i) {
				dateList.push(`${date.getHours()}:${date.getMinutes}`)
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
		xAxis: {
			type: 'category',
			boundaryGap: false,
			data: [],
		},
		yAxis: {
			type: 'value'
		},
		series: [
			{
				name: 'percent',
				data: [],
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
	})
	pkg.utils.invokeInterval(() => {
		fetchTl().then(res => {
			chart.setOption({
				xAxis: {
					data: res[0]
				},
				series: [{
					name: 'percent',
					data: res[1]
				}]
			})
		}).catch((err: Error) => useToast().error(err.message))
	}, 1000 * 60 * 60)
})
</script>
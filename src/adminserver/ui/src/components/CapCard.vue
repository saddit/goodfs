<template>
	<div class="bg-white rounded-lg shadow-lg inline-flex items-center relative justify-center">
		<span class="absolute top-2 left-2 text-indigo-600 font-bold">{{ $t('capacity') }}</span>
		<div id="cap-chart" ref="capChart" class="w-1/2 h-44"></div>
		<div class="w-fit">
			<div class="text-2xl md:text-3xl text-gray-800 lg:text-4xl 2xl:text-[3rem] font-extrabold text-center py-1">{{
					$utils.formatBytes(capInfo.used)
			}}</div>
			<div class="text-gray-500 text-xs md:text-sm font-mono font-light text-right">
				<div>Free: {{ $utils.formatBytes(capInfo.free) }}</div>
				<div>Total: {{ $utils.formatBytes(capInfo.total) }}</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import * as echarts from 'echarts'

const props = defineProps<{
	capInfo: DiskInfo
}>()

const capChart = ref()

onMounted(()=>{
	let capEcahrt = echarts.init(capChart.value)
  window.addEventListener("resize", () => {
    capEcahrt.resize()
  })
  capEcahrt.setOption({
    color: ['#4f46e5', '#6ee7b7'],
    series: [
      {
        name: 'Capcity',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        data: [
          { value: props.capInfo.used, name: 'Used' },
          { value: props.capInfo.free, name: 'Free' },
        ]
      }
    ]
  })

	watch(props.capInfo, res=>{
		capEcahrt.setOption({
			serise: [{
				name: 'Capacity',
				data: [
          { value: res.used, name: 'Used' },
          { value: res.free, name: 'Free' },
        ]
			}]
		})
	}, {deep: true})
})
</script>
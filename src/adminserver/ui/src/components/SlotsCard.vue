<template>
  <div ref="slotsCardDom" class="p-3 bg-white shadow-md rounded-md">
    <div class="font-bold text-xl text-indigo-600">Slots</div>
    <!-- legend -->
    <div class="inline-flex flex-wrap space-x-3 mt-2">
      <div v-for="v in badgers" class="inline-flex items-center">
        <div class="w-6 h-4 rounded mr-1" :class="[getBgColor(v)]"></div>
        <div class="text-sm">{{ v }}</div>
      </div>
    </div>
    <!-- lines -->
    <div class="inline-flex items-center pt-2 mt-2">
      <div v-for="v in value" :style="{width: getWid(v)}" class="h-2 group relative" :class="[getBgColor(v.identify)]">
        <span
            class="transition-opacity font-light text-xs opacity-0 group-hover:opacity-100 fixed -top-4 left-0 text-gray-600 select-none">
          {{ `${v.start}-${v.end}` }}
        </span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
const prop = defineProps<{
    value: SlotRange[]
}>()

const slotsCardDom = ref()
const allColors = ['bg-orange-500', 'bg-indigo-500', 'bg-red-500', 'bg-green-500', 'bg-blue-500', 'bg-yellow-500']
const colorDict: { [key: string]: string } = {}

const badgers = computed(() => {
    let s: string[] = []
    for (let v of prop.value) {
        if (s.includes(v.identify)) {
            continue
        }
        s.push(v.identify)
    }
    console.log(s)
    return s
})

function getWid(v: SlotRange): string {
    let len = v.end - v.start
    return `${unitWidth.value * len * 0.85}px`
}

function getBgColor(v: string): string {
    if (!(v in colorDict)) {
        colorDict[v] = allColors.pop() || "bg-gray-500"
    }
    return colorDict[v]
}

const unitWidth = ref(0)

onMounted(() => {
    unitWidth.value = slotsCardDom.value.clientWidth / 16384
    watch(slotsCardDom.value.clientWidth, (v: number) => {
        unitWidth.value = v / 16384
    })
})

</script>

<style scoped>

</style>
<template>
  <div class="group relative cursor-pointer">
    <slot></slot>
    <div v-if="enable"
         :class="[pos]"
         class="group-hover:opacity-100 whitespace-nowrap transition-opacity absolute w-fit opacity-0 bg-gray-900 text-white px-1 text-sm py-0.5 rounded">
      {{ tip }}
    </div>
  </div>
</template>

<script setup lang="ts">
const prop = withDefaults(defineProps<{
    tip: string
    pos: string
    enable: boolean
}>(), {
    pos: '-top-5',
    enable: true
})
</script>

<style scoped>

</style>
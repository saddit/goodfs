<template>
  <div class="group relative cursor-pointer">
    <slot></slot>
    <div class="group-hover:opacity-100 whitespace-nowrap transition-opacity absolute w-fit opacity-0 bg-gray-900 text-white px-1 text-sm py-0.5 -top-5 rounded">
      {{ tip }}
    </div>
  </div>
</template>

<script setup lang="ts">
const prop = defineProps<{
    tip: string
}>()
</script>

<style scoped>

</style>
<script setup>
</script>

<template>
  <div>
    <header class="bg-white shadow" v-if="$route.meta.title">
      <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8 flex items-baseline space-x-2">
        <h1 class="text-4xl font-bold leading-tight text-gray-900">
          DFS Management
        </h1>
        <h1 class="text-2xl leading-tight text-gray-900 text-indigo-600">
          {{ $route.meta.title }}
        </h1>
      </div>
    </header>
    <main>
      <router-view />
    </main>
  </div>
</template>

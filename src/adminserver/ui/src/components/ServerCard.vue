<template>
    <div class="min-w-40 flex flex-col justify-between py-3 px-2 bg-white h-44 shadow-lg rounded">
        <div class="inline-flex justify-between items-center px-2">
            <ServerIcon class="text-indigo-600 w-10 h-10" />
            <span class="font-bold text-gray-900 text-lg">{{ info.serverId }}</span>
        </div>
        <div class="space-y-1">
            <div class="text-gray-500 text-xs px-2">HTTP: {{ info.httpAddr }}</div>
            <div class="text-gray-500 text-xs px-2">GRPC: {{ info.rpcAddr }}</div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ServerIcon } from "@heroicons/vue/20/solid"
const props = defineProps<{
    info: ServerInfo
}>()

</script>
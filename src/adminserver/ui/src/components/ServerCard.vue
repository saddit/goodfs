<template>
  <div :class="{'outline outline-indigo-400 outline-offset-4': info.isMaster}"
       class="relative min-w-40 py-3 px-2 bg-white h-48 w-40 md:h-44 md:w-44 shadow-lg rounded">
    <div class="inline-flex justify-between items-center px-2 w-full">
      <ServerIcon class="text-indigo-600 w-10 h-10"/>
      <span class="font-bold text-gray-900 text-lg">{{ info.serverId }}</span>
    </div>
    <div class="inline-flex items-center w-fit py-2 text-gray-500 text-xs">
      <CpuChipIcon class="w-8 h-8 mx-1 text-indigo-600"/>
      <div>
        <div>{{ $t('core') }}: {{
            info.sysInfo.cpuStatus.logicalCount
          }}
        </div>
        <div>{{ $t('ram') }}: {{ $utils.formatBytes(info.sysInfo.memStatus.all) }}</div>
      </div>
    </div>
    <div class="space-y-1 absolute bottom-2">
      <div class="text-gray-500 text-xs px-2">HTTP: {{ info.httpAddr }}</div>
      <div v-if="info.rpcAddr" class="text-gray-500 text-xs px-2">GRPC: {{ info.rpcAddr }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ServerIcon} from "@heroicons/vue/20/solid"
import {CpuChipIcon} from "@heroicons/vue/24/outline"

const props = defineProps<{
  info: ServerInfo
}>()

</script>
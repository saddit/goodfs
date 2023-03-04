<template>
  <div class="min-w-40 py-3 px-2 bg-white h-48 w-40 md:h-44 md:w-44 shadow-lg rounded flex flex-col justify-between">
    <div>
      <div class="inline-flex justify-between items-center px-2 w-full">
        <ServerIcon class="text-indigo-600 w-10 h-10"/>
        <div>
          <div class="font-bold text-gray-900 text-lg">{{ info.serverId }}</div>
          <div v-if="info.isMaster" class="font-medium text-red-500 text-xs text-right">master</div>
        </div>
      </div>
      <div class="inline-flex items-center w-fit py-2 text-gray-500 text-xs">
        <CpuChipIcon class="w-8 h-8 mx-1 text-indigo-600"/>
        <div>
          <div>{{ $t('core') }}: {{ info.sysInfo.cpuStatus.logicalCount }}</div>
          <div>{{ $t('ram') }}: {{ $utils.formatBytes(info.sysInfo.memStatus.all) }}</div>
        </div>
      </div>
    </div>
    <div class="text-left ml-2 text-gray-500 text-sm">
      IP: {{ info.httpAddr }}
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
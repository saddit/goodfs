<template>
  <div class="w-full h-full overflow-y-auto bg-gray-100 p-6">
    <div class="flex">
      <div class="display-overview w-1/4 min-w-[24rem]">
        <!-- alive counts -->
        <div class="mr-2">
          <div class="font-medium text-indigo-500">{{ t('online') }}</div>
          <div class="flex space-x-3 items-center">
            <span class="font-extrabold text-6xl"
                  :class="[totalAlive > 0 ? 'text-green-500' : 'text-red-600' ]">{{ totalAlive }}</span>
            <div class="space-y-0.5 text-sm font-mono hidden md:block">
              <div>API Server: {{ overview.aliveCounts.apiserver }}</div>
              <div>META Server: {{ overview.aliveCounts.metaserver }}</div>
              <div>DATA Server: {{ overview.aliveCounts.objectserver }}</div>
            </div>
          </div>
        </div>
        <!-- total buckets -->
        <div class="flex flex-col justify-center mr-2">
          <div class="font-bold text-3xl text-gray-900 text-center">{{ overview.totalBuckets }}</div>
          <div class="font-medium text-center text-indigo-500">Buckets</div>
        </div>
        <!-- total objects -->
        <div class="flex flex-col justify-center mr-2">
          <div class="font-bold text-3xl text-gray-900 text-center">{{ overview.totalObjects }}</div>
          <div class="font-medium text-center text-indigo-500">Objects</div>
        </div>
        <!-- avg cpu -->
        <span class="mb-2">
          <UsageLine h="w-96 h-32" :type="$cst.statTypeCpu" :tl="overview.avgCpu"></UsageLine>
        </span>
        <!-- avg mem -->
        <span class="mb-2">
          <UsageLine h="w-96 h-32" :type="$cst.statTypeMem" :tl="overview.avgMem"></UsageLine>
        </span>
      </div>
      <!-- etcd -->
      <div class="w-1/2 ml-5 bg-white rounded-md shadow-md px-4">
        <div class="mt-3 font-medium text-indigo-600 text-xl" @click="filterLogs('all')">{{ t('etcd-cluster') }}</div>
        <!-- member list -->
        <div class="flex flex-wrap space-x-2 mt-4 mb-5">
          <div v-for="v in etcdStats" class="p-2 mb-2 border border-gray-500 rounded-md" @click="filterLogs(v.endpoint)">
            <div class="text-center text-4xl text-indigo-600 my-1">
              <font-awesome-icon icon="database"/>
            </div>
            <div class="my-3">
              <span class="text-sm">{{ t('db-size') + ' ' }}</span>
              <span class="font-bold">{{ $utils.formatBytes(v.dbSizeInUse) }}/{{ $utils.formatBytes(v.dbSize) }}</span>
            </div>
            <div class="text-center text-sm text-gray-500 my-1">{{ v.endpoint }}</div>
          </div>
        </div>
        <!-- logs -->
        <div class="text-gray-900">
          {{ t('alarm-logs') }}
          <span class="text-sm text-gray-400">- {{ currentLog }}</span>
        </div>
        <div class="alarm-logs">
          <li v-for="s in alarmMessages" class="break-words">{{ s }}</li>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {useI18n} from "vue-i18n";

const route = useRoute()
const router = useRouter()
const store = useStore()
const {t} = useI18n({inheritLocale: true})

const overview = ref({
    "aliveCounts": {
        "apiserver": 0,
        "metaserver": 0,
        "objectserver": 0
    },
    "totalBuckets": 0,
    "totalObjects": 0,
    "avgCpu": {},
    "avgMem": {},
})

const etcdStats = ref<EtcdStatus[]>([])
const alarmMessages = ref<string[]>([])
const currentLog = ref("all")

const totalAlive = computed(() => {
    let ac = overview.value.aliveCounts
    return ac.apiserver + ac.metaserver + ac.objectserver
})

onBeforeMount(() => {
    let redirect = route.query['redirect'] as string
    if (redirect) {
        delete route.query.redirect
        router.push({path: redirect, query: route.query})
        return
    }
})

onMounted(() => {
    getOverview()
})

function filterLogs(t: string) {
    if (t == currentLog.value) {
        return
    }
    currentLog.value = t
    let msg = []
    for (let e of etcdStats.value) {
        if (t == "all" || e.endpoint == t) {
            msg.push(...e.alarmMessage)
        }
    }
    alarmMessages.value = msg
}

function getOverview() {
    api.serverStat.overview().then(r => {
        overview.value = r
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
    api.serverStat.stat().then((res) => {
        useStore().setServerInfo(res)
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
    api.serverStat.etcdStat().then(r => {
        etcdStats.value = r
        filterLogs("all")
    })
}
</script>

<style scoped>
.display-overview {
    @apply flex flex-wrap
}

.display-overview > div {
    @apply bg-white rounded-md p-3 shadow-md mb-2
}

.display-overview > hr {
    @apply w-full invisible
}

.alarm-logs {
    @apply bg-gray-200 rounded-lg px-4 py-3 my-3 font-mono text-sm overflow-y-auto h-44
}

.alarm-logs > li::marker {
    @apply text-gray-400;
    content: '> ';
}
</style>

<i18n lang="yaml">
zh:
  etcd-cluster: "ETCD 集群"
  db-size: "数据大小"
  online: "在线数量"
  alarm-logs: "警告日志"
en:
  etcd-cluster: "ETCD Cluster"
  db-size: "DB Size"
  online: "Online"
  alarm-logs: "Alarm Logs"
</i18n>

<route lang="json">
{
  "meta": {
    "title": "home",
    "icon": "house"
  }
}
</route>
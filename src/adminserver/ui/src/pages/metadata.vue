<template>
  <div class="w-full h-full overflow-y-auto">
    <div class="p-2 w-fit mx-auto mt-10">
      <div class="flex items-center">
        <MagnifyingGlassIcon class="w-6 h-6 mr-2 text-indigo-600"/>
        <input type="text" class="text-input-pri" :placeholder="t('search-by-name')"/>
      </div>
      <!-- metadata table -->
      <table class="mt-4">
        <thead>
        <tr
            v-for="headerGroup in dataTable.getHeaderGroups()"
            :key="headerGroup.id"
        >
          <th
              v-for="header in headerGroup.headers"
              :key="header.id"
              :colSpan="header.colSpan"
          >
            <FlexRender
                v-if="!header.isPlaceholder"
                :render="header.column.columnDef.header"
                :props="header.getContext()"
            />
          </th>
        </tr>
        </thead>
        <tbody>
        <template v-if="dataList.length > 0">
          <tr v-for="row in dataTable.getRowModel().rows" :key="row.id">
            <td v-for="cell in row.getVisibleCells()" :key="cell.id">
              <FlexRender
                  :render="cell.column.columnDef.cell"
                  :props="cell.getContext()"
              />
            </td>
          </tr>
        </template>
        <tr v-else><td colspan="4" class="text-center">{{t('no-data')}}</td></tr>
        </tbody>
      </table>
      <Pagination class="my-4" :max-num="5" :total="dataReq.total" :page-size="dataReq.pageSize"
                  :model-value="dataReq.page"/>
    </div>
  </div>
</template>

<script setup lang="ts">
import {createColumnHelper, FlexRender, getCoreRowModel, useVueTable} from '@tanstack/vue-table'
import {MagnifyingGlassIcon} from '@heroicons/vue/20/solid'

const dataList = ref<Metadata[]>([])
const versionList = ref<Version[]>([])
const dataReq = reactive<MetadataReq>({name: '', page: 1, total: 0, pageSize: 10})
const versionReq = reactive<MetadataReq>({name: '', page: 1, total: 0, pageSize: 10})

const {t} = useI18n({inheritLocale: true})

function queryMetadata() {
    let req = unref(dataReq)
    api.metadata.metadataPage(req)
        .then(res => {
            dataList.value = res.list
            req.total = res.total
        })
        .catch((err: Error) => {
            useToast().error(err.message)
        })
}

function queryVersion(name: string) {
    let req = unref(versionReq)
    req.name = name
    api.metadata.versionPage(req)
        .then(res => {
            versionList.value = res.list
            versionReq.total = res.total
        })
        .catch((err: Error) => {
            useToast().error(err.message)
        })
}

watch(dataReq, () => {
    queryMetadata()
})

watch(versionReq, v => {
    queryVersion(v.name)
})

onBeforeMount(() => {
    queryMetadata()
})

const dataColumnHelper = createColumnHelper<Metadata>()
const versionColumnHelper = createColumnHelper<Version>()

const dataColumns = [
    dataColumnHelper.accessor('name', {
        id: 'metadata-id',
        header: () => 'Name',
        cell: props => props.getValue()
    }),
    dataColumnHelper.accessor('createTime', {
        header: () => 'Created At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    dataColumnHelper.accessor('updateTime', {
        header: () => 'Updated At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    dataColumnHelper.display({
        id: 'action',
        header: 'Actions',
        cell: ({row}) => h('button', {
            class: 'btn-action',
            onClick: () => queryVersion(row.getValue('metadata-id'))
        }, t('see-version'))
    }),
]

const versionColumns = [
    versionColumnHelper.accessor('sequence', {
        header: () => 'Number',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('size', {
        header: () => 'Size',
        cell: props => pkg.utils.formatBytes(props.getValue(), 1)
    }),
    versionColumnHelper.accessor('ts', {
        header: () => 'Timestamp',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('hash', {
        header: () => 'Digest',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('storeStrategy', {
        header: () => 'Strategy',
        cell: props => pkg.cst.storeStrategy[props.getValue()]
    }),
    versionColumnHelper.accessor('dataShards', {
        header: () => 'Data Shards',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('parityShards', {
        header: () => 'Parity Shards',
        cell: props => props.getValue()
    })
]

const dataTable = useVueTable({
    get data() {
        return dataList.value
    },
    columns: dataColumns,
    getCoreRowModel: getCoreRowModel(),
})
</script>

<style scoped>
table {
    @apply border border-gray-300 rounded-md
}

thead tr {
    @apply border-b border-gray-300 bg-indigo-400 bg-opacity-10 text-indigo-600
}

thead th {
    @apply py-2 px-4
}

tbody td {
    @apply px-6 py-4 text-sm text-gray-900
}

/*noinspection CssUnusedSymbol*/
.btn-action {
    @apply underline text-indigo-500 hover:text-indigo-400 text-sm
}
</style>

<route lang="json">
{
  "meta": {
    "title": "metadata",
    "icon": "table"
  }
}
</route>

<i18n lang="yaml">
en:
  no-data: 'Empty Data Table'
  search-by-name: 'Search By Name Prefix'
  see-version: 'See Versions'
zh:
  no-data: '暂无数据'
  search-by-name: '根据文件名前缀查询'
  see-version: '查询版本'
</i18n>
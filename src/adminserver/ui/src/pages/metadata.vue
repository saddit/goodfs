<template>
  <div class="w-full flex flex-col items-center">
    <div class="flex w-full mt-1 items-center justify-start">
      <MagnifyingGlassIcon class="w-6 h-6 mx-2 text-indigo-600"/>
      <input type="text"
             @change="e => searchData(e.target)"
             class="text-input-pri"
             :placeholder="t('search-by-name')"/>
      <!-- TODO: beautify -->
      <input v-if="dataReq.bucket" type="file" class="ml-8" @change="uploadObject"/>
    </div>
    <!-- metadata table -->
    <table class="mt-4 w-full">
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
      <tr v-else>
        <td colspan="5" class="text-center">{{ t('no-data') }}</td>
      </tr>
      </tbody>
    </table>
    <div class="inline-flex items-center space-x-3 my-4">
      <span class="text-gray-900 text-sm">{{ t('total-num') }}: {{ dataReq.total }}</span>
      <select class="select-pri-sm" v-model="dataReq.pageSize">
        <option v-for="v in [10,20,50]" :value="v">{{ v }}/page</option>
      </select>
      <Pagination :max-num="10" :total="dataReq.total" :page-size="dataReq.pageSize"
                  v-model="dataReq.page"/>
    </div>
    <!-- version table -->
    <ModalTemplate v-model="showVersionTb" title="Versions">
      <template #panel>
        <div class="flex flex-col items-center">
          <table class="m-1">
            <thead>
            <tr
                v-for="headerGroup in versionTable.getHeaderGroups()"
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
            <template v-if="versionList.length > 0">
              <tr v-for="row in versionTable.getRowModel().rows" :key="row.id">
                <td v-for="cell in row.getVisibleCells()" :key="cell.id">
                  <FlexRender
                      :render="cell.column.columnDef.cell"
                      :props="cell.getContext()"
                  />
                </td>
              </tr>
            </template>
            <tr v-else>
              <td :colspan="7" class="text-center">{{ t('no-data') }}</td>
            </tr>
            </tbody>
          </table>
          <div class="inline-flex items-center space-x-3 my-4 mx-auto">
            <span class="text-gray-900 text-sm">{{ t('total-num') }}: {{ versionReq.total }}</span>
            <select class="select-pri-sm">
              <option v-for="v in [10,20,50]" :value="v">{{ v }}/page</option>
            </select>
            <Pagination :max-num="5" :total="versionReq.total" :page-size="versionReq.pageSize"
                        v-model="versionReq.page"/>
          </div>
        </div>
      </template>
    </ModalTemplate>
  </div>
</template>

<script setup lang="ts">
import {createColumnHelper, FlexRender, getCoreRowModel, useVueTable} from '@tanstack/vue-table'
import {MagnifyingGlassIcon} from '@heroicons/vue/20/solid'

const defPage: Pageable = {page: 1, total: 0, pageSize: 10}

const dataList = ref<Metadata[]>([])
const versionList = ref<Version[]>([])
const dataReq = reactive<MetadataReq>({name: '', bucket: '', ...defPage})
const versionReq = reactive<MetadataReq>({name: '', bucket: '', ...defPage})
const showVersionTb = ref(false)

const {t} = useI18n({inheritLocale: true})

function searchData(elem: any) {
    dataReq.name = (elem as HTMLTextAreaElement).value
    queryMetadata()
}

function queryMetadata() {
    api.metadata.metadataPage(dataReq)
        .then(res => {
            dataList.value = res.list
            dataReq.total = res.total
        })
        .catch((err: Error) => {
            useToast().error(err.message)
        })
}

function queryVersion(name: string, bucket: string) {
    versionReq.name = name
    versionReq.bucket = bucket
    api.metadata.versionPage(versionReq)
        .then(res => {
            versionList.value = res.list
            versionReq.total = res.total
            showVersionTb.value = true
        })
        .catch((err: Error) => {
            useToast().error(err.message)
        })
}

watch(() => dataReq.page, () => {
    queryMetadata()
})
watch(() => dataReq.pageSize, () => {
    queryMetadata()
})

watch(() => versionReq.page, () => {
    queryVersion(versionReq.name, versionReq.bucket)
})

watch(() => versionReq.pageSize, () => {
    queryVersion(versionReq.name, versionReq.bucket)
})

onBeforeMount(() => {
    dataReq.bucket = useRoute().query['bucket'] as string
    queryMetadata()
})

const dataColumnHelper = createColumnHelper<Metadata>()
const versionColumnHelper = createColumnHelper<Version>()

function downloadObject(name: string, bucket: string, version: number) {
    api.objects.download(name, bucket, version).catch((err: Error) => {
        useToast().error(err.message)
    })
}

function uploadObject(event: any) {
    if (event.target?.files?.length == 0) {
        return
    }
    let file: File = event.target.files[0]
    api.objects.upload(file, dataReq.bucket).then(() => {
        useToast().success("Upload Success!")
        queryMetadata()
    }).catch((err: Error) => {
        useToast().error("Upload Fail: " + err.message)
    })
}

const dataColumns = [
    dataColumnHelper.accessor('name', {
        header: 'Name',
        cell: props => pkg.utils.cutStr(props.getValue(), 30),
    }),
    dataColumnHelper.accessor('bucket', {
        header: 'Bucket',
        cell: props => props.getValue()
    }),
    dataColumnHelper.accessor('createTime', {
        header: 'Created At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    dataColumnHelper.accessor('updateTime', {
        header: 'Updated At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    dataColumnHelper.display({
        id: 'action',
        header: 'Actions',
        cell: ({row}) => h('button', {
            class: 'btn-action',
            onClick: () => queryVersion(row.original.name, row.original.bucket)
        }, t('see-version'))
    }),
]

const versionColumns = [
    versionColumnHelper.accessor('sequence', {
        header: 'Version',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('size', {
        header: 'Size',
        cell: props => pkg.utils.formatBytes(props.getValue(), 1)
    }),
    versionColumnHelper.accessor('ts', {
        header: 'Timestamp',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    versionColumnHelper.accessor('storeStrategy', {
        header: 'Strategy',
        cell: props => t(pkg.cst.storeStrategy[props.getValue()])
    }),
    versionColumnHelper.accessor('dataShards', {
        header: 'Data Shards',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('parityShards', {
        header: 'Parity Shards',
        cell: props => props.getValue()
    }),
    versionColumnHelper.accessor('compress', {
        header: 'Compress',
        cell: ({row}) => h('input', {
            type: "checkbox",
            disabled: true,
            class: "checkbox-pri",
            checked: row.original.compress
        }, '')
    }),
    versionColumnHelper.display({
        id: 'action',
        header: 'Actions',
        cell: ({row}) => h('p', [
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 text-sm',
                onClick: () => useToast().info(`Digest: ${row.original.hash}`)
            }, 'Checksum'),
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 text-sm ml-1',
                onClick: () => downloadObject(versionReq.name, versionReq.bucket, row.original.sequence)
            }, 'Download')
        ])
    }),
]

const dataTable = useVueTable({
    get data() {
        return dataList.value
    },
    columns: dataColumns,
    getCoreRowModel: getCoreRowModel(),
})

const versionTable = useVueTable({
    get data() {
        return versionList.value
    },
    columns: versionColumns,
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
    @apply py-2 px-6
}

tbody td {
    @apply px-4 py-6 text-sm text-gray-900 text-center
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
  search-by-name: '根据文件名前缀查找'
  see-version: '查询版本'
</i18n>
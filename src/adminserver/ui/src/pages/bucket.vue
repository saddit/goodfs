<template>
  <div class="flex items-center mx-4">
    <MagnifyingGlassIcon class="w-6 h-6 mr-2 text-indigo-600"/>
    <input type="text"
           class="text-input-pri"
           :placeholder="t('search-by-name')"/>
  </div>
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
    <tr v-else>
      <td :colspan="columns.length" class="text-center">{{ t('no-data') }}</td>
    </tr>
    </tbody>
  </table>
  <!-- FIXME: bug when click next page -->
  <Pagination class="my-4" :max-num="5" :total="dataReq.total" :page-size="dataReq.pageSize"
              v-model="dataReq.page"/>
</template>

<script setup lang="ts">
import {createColumnHelper, FlexRender, getCoreRowModel, useVueTable} from "@tanstack/vue-table";
import {MagnifyingGlassIcon} from "@heroicons/vue/20/solid";

const defPage: Pageable = {page: 1, total: 0, pageSize: 10, orderBy: 'create_time', desc: false}
const dataList = ref<Bucket[]>([])
const dataReq = ref<BucketReq>({name: '', ...defPage})

function routeToMetadata(name: string) {
    useRouter().push({
        path: '/metadata',
        query: {
            'bucket': name
        }
    })
}

const {t} = useI18n({inheritLocale: true})
const columnHelper = createColumnHelper<Bucket>()

const columns = [
    columnHelper.accessor('name', {
        header: 'Name',
        cell: props => props.getValue()
    }),
    columnHelper.accessor('versioning', {
        header: 'Versioning',
        cell: props => props.getValue()
    }),
    columnHelper.accessor('storeStrategy', {
        header: 'Strategy',
        cell: props => pkg.cst.storeStrategy[props.getValue()]
    }),
    columnHelper.accessor('compress', {
        header: 'Compress',
        cell: props => props.getValue()
    }),
    columnHelper.accessor('createTime', {
        header: 'Created At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    columnHelper.accessor('updateTime', {
        header: 'Updated At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    columnHelper.display({
        id: 'action',
        header: 'Actions',
        cell: ({row}) => h('p', [
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 text-sm ml-1'
            }, t('upload')),
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 text-sm ml-1'
            }, t('objects')),
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 text-sm'
            }, t('detail')),
        ])
    }),
]

const dataTable = useVueTable({
    get data() {
        return dataList.value
    },
    columns,
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
    @apply px-3 py-6 text-sm text-gray-900 text-center
}

/*noinspection CssUnusedSymbol*/
.btn-action {
    @apply underline text-indigo-500 hover:text-indigo-400 text-sm
}
</style>

<route lang="json">
{
  "meta": {
    "title": "bucket",
    "icon": "bucket"
  }
}
</route>

<i18n lang="yaml">
en:
  no-data: 'Empty Data Table'
  upload: 'Upload'
  objects: 'See Objects'
  detail: 'Detail'
  search-by-name: 'Search By Name Prefix'
zh:
  no-data: '暂无数据'
  upload: '上传'
  objects: '查看对象'
  detail: '详情'
  search-by-name: '根据名称前缀查找'
</i18n>
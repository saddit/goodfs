<template>
  <div class="flex items-center mx-4">
    <MagnifyingGlassIcon class="w-6 h-6 mr-2 text-indigo-600"/>
    <input type="text"
           class="text-input-pri"
           :placeholder="t('search-by-name')"/>
    <button class="btn-pri-sm ml-4" @click="operateType = opAdd">{{ t('add-bucket') }}</button>
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
  <Pagination class="my-4" :max-num="5" :total="dataReq.total" :page-size="dataReq.pageSize"
              v-model="dataReq.page"/>
  <!-- add or update dialog -->
  <ModalTemplate v-model="isOperating" title="Bucket" v-if="operateType > 0">
    <template #panel>
      <div class="mt-6 grid grid-cols-3 items-center gap-y-4 gap-x-1 max-w-lg sm:max-w-md">
        <!-- row: name -->
        <span>{{ t('field-name') }}</span>
        <input type="text" class="col-span-2 input-text-pri" v-model="operatingBucket.name"/>

        <!-- row: store strategy -->
        <span>{{ t('field-store-strategy') }}</span>
        <select class="col-span-2 select-pri" v-model="operatingBucket.storeStrategy">
          <option v-for="(v,idx) in $cst.storeStrategy" :value="idx">{{ t(v) }}</option>
        </select>

        <!-- optional area: strategy params -->
        <template v-if="operatingBucket.storeStrategy > 0">
          <!-- row: data shards -->
          <span>{{ t('field-data-shards') }}</span>
          <input type="number" class="input-text-pri" v-model="operatingBucket.dataShards"/>
          <span></span>

          <!-- row: parity shards -->
          <template v-if="operatingBucket.storeStrategy == $cst.ssRS">
            <span>{{ t('field-parity-shards') }}</span>
            <input type="number" class="input-text-pri"
                   v-model="operatingBucket.parityShards"/>
            <span></span>
          </template>
        </template>

        <!-- row: enable versioning -->
        <span></span>
        <div class="col-span-2">
          <input type="checkbox" class="checkbox-pri"
                 v-model="operatingBucket.versioning"/>
          <span class="ml-2">{{ t('field-versioning') }}</span>
        </div>

        <!-- optional area: versioning-->
        <template v-if="operatingBucket.versioning">
          <span>{{ t('field-version-remains') }}</span>
          <input type="number" class="input-text-pri"
                 v-model="operatingBucket.versionRemains"/>
          <span></span>
        </template>

        <!-- row: enable compress -->
        <span></span>
        <div class="col-span-2">
          <input type="checkbox" class="checkbox-pri"
                 v-model="operatingBucket.compress"/>
          <span class="ml-2">{{ t('field-compress') }}</span>
        </div>

        <!-- row: is readonly -->
        <span></span>
        <div class="col-span-2">
          <input type="checkbox" class="checkbox-pri"
                 v-model="operatingBucket.readonly"/>
          <span class="ml-2">{{ t('field-readonly') }}<span
              class="text-xs text-red-500 ml-1">({{ t('update-hint') }})</span></span>
        </div>

        <!-- row: buttons -->
        <span></span>
        <button class="btn-pri-sm" @click="operateType == 1 ? addBucket() : updateBucket()">{{ t('btn-ok') }}</button>
        <button class="btn-normal-sm" @click="isOperating = false">{{ t('btn-cancel') }}</button>
      </div>
    </template>
  </ModalTemplate>
  <!-- delete confirm dialog -->
  <ModalTemplate v-model="isDeleting" :title="t('is-confirm-to-delete')">
    <template #panel>
      <div>
        <div class="break-words max-w-md min-h-28 p-4 text-sm text-orange-500">{{ t('danger-notification') }}</div>
        <div class="inline-flex justify-end w-full space-x-6 sm:space-x-4">
          <button class="btn-revert sm:btn-revert-sm" @click="removeBucket()">{{ t('confirm-delete') }}</button>
          <button class="btn-normal sm:btn-normal-sm" @click="isDeleting = false">{{ t('btn-cancel') }}</button>
        </div>
      </div>
    </template>
  </ModalTemplate>
</template>

<script setup lang="ts">
import {createColumnHelper, FlexRender, getCoreRowModel, useVueTable} from "@tanstack/vue-table";
import {MagnifyingGlassIcon} from "@heroicons/vue/20/solid";

const opNone = 0
const opAdd = 1
const opUpdate = 2
const opDel = 3

const defPage: Pageable = {page: 1, total: 0, pageSize: 10}
const dataList = ref<Bucket[]>([])
const dataReq = reactive<BucketReq>({name: '', ...defPage})
const operatingBucket = ref<Bucket>({} as Bucket)
const operateType = ref(0)

const isDeleting = computed({
    get: () => {
        return operateType.value == opDel
    },
    set: (v) => {
        if (v) {
            operateType.value = opDel
            return
        }
        operateType.value = opNone
    }
})

const isOperating = computed({
    get: () => {
        return operateType.value == opAdd || operateType.value == opUpdate
    },
    set: (v) => {
        if (v) {
            return
        }
        operateType.value = opNone
        operatingBucket.value = {} as Bucket
    }
})

const router = useRouter()

function routeToMetadata(name: string) {
    router.push({
        path: '/metadata',
        query: {
            'bucket': name
        }
    })
}

function queryBuckets() {
    api.metadata.bucketPage(dataReq).then(res => {
        dataReq.total = res.total
        dataList.value = res.list
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
}

async function addBucket() {
    if (!operatingBucket.value.name) {
        useToast().error("bucket name required")
        return
    }
    try {
        await api.metadata.addBucket(operatingBucket.value)
        useToast().success(t('req-success'))
        operateType.value = opNone
        operatingBucket.value = {} as Bucket
        queryBuckets()
    } catch (e: any) {
        useToast().error(e.message)
    }
}

async function updateBucket() {
    if (!operatingBucket.value.name) {
        useToast().error("bucket name required")
        return
    }
    try {
        await api.metadata.updateBucket(operatingBucket.value)
        useToast().success(t('req-success'))
        operateType.value = opNone
        operatingBucket.value = {} as Bucket
        queryBuckets()
    } catch (e: any) {
        useToast().error(e.message)
    }
}

async function removeBucket() {
    if (!operatingBucket.value.name) {
        useToast().error("bucket name required")
        return
    }
    try {
        await api.metadata.removeBucket(operatingBucket.value.name)
        useToast().success(t('req-success'))
        operateType.value = opNone
        operatingBucket.value = {} as Bucket
        queryBuckets()
    } catch (e: any) {
        useToast().error(e.message)
    }
}

watch(() => dataReq.page, () => {
    queryBuckets()
})
watch(() => dataReq.pageSize, () => {
    queryBuckets()
})

onBeforeMount(() => {
    queryBuckets()
})

const {t} = useI18n({inheritLocale: true})
const columnHelper = createColumnHelper<Bucket>()

const columns = [
    columnHelper.accessor('name', {
        header: 'Name',
        cell: props => props.getValue()
    }),
    columnHelper.accessor('versioning', {
        header: 'Versioning',
        cell: ({row}) => h('p', {
            class: 'inline-flex items-center'
        }, [
            h('input', {
                type: "checkbox",
                disabled: true,
                class: "checkbox-pri",
                checked: row.original.versioning
            }, ''),
            row.original.versioning ? h('span', {
                class: 'text-sm ml-1 text-indigo-500'
            }, `${row.original.versionRemains}`) : h('span')
        ])
    }),
    columnHelper.accessor('storeStrategy', {
        header: 'Strategy',
        cell: props => `${t(pkg.cst.storeStrategy[props.getValue()])} (${props.row.original.dataShards}+${props.row.original.parityShards})`
    }),
    columnHelper.accessor('compress', {
        header: 'Compress',
        cell: ({row}) => h('input', {
            type: "checkbox",
            disabled: true,
            class: "checkbox-pri",
            checked: row.original.compress
        }, '')
    }),
    columnHelper.accessor('createTime', {
        header: 'Created At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    columnHelper.accessor('updateTime', {
        header: 'Updated At',
        cell: props => new Date(props.getValue()).toLocaleString()
    }),
    columnHelper.accessor('readonly', {
        header: 'Readonly',
        cell: ({row}) => h('input', {
            type: "checkbox",
            disabled: true,
            class: "checkbox-pri",
            checked: row.original.readonly
        }, '')
    }),
    columnHelper.display({
        id: 'action',
        header: 'Actions',
        cell: ({row}) => h('div', {
            class: 'overflow-x-auto inline-flex justify-center w-full space-x-4 md:space-x-2 sm:space-x-1'
        }, [
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 sm:text-sm',
                onClick: () => routeToMetadata(row.original.name)
            }, t('objects')),
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 sm:text-sm',
                onClick: () => {
                    operatingBucket.value = row.original
                    operateType.value = opUpdate
                }
            }, t('detail')),
            h('button', {
                class: 'underline text-indigo-500 hover:text-indigo-400 text-sm',
                onClick: () => {
                    operatingBucket.value = {name: row.original.name} as Bucket
                    operateType.value = opDel
                }
            }, t('delete')),
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
  objects: 'Objects'
  detail: 'Detail'
  delete: 'Delete'
  search-by-name: 'Search By Name Prefix'
  danger-notification: 'This is a dangerous operation that will cause all objects under the bucket inaccessible, recreating it with the same name will restore it, are you sure continue?'
  is-confirm-to-delete: 'Do you confirm to delete?'
  confirm-delete: 'Continue'
  add-bucket: 'Add New Bucket'
  update-hint: "Affect existed objects"
  field-compress: 'Enable data compression'
  field-versioning: 'Enable multi-version for objects'
  field-readonly: 'Set to readonly'
  field-name: 'Bucket Name'
  field-version-remains: 'Maximum Remains'
  field-store-strategy: 'Store Strategy'
  field-data-shards: 'Data Shards'
  field-parity-shards: 'Parity Shards'
zh:
  no-data: '暂无数据'
  upload: '上传'
  objects: '查看对象'
  detail: '详情'
  delete: '移除'
  search-by-name: '根据名称前缀查找'
  danger-notification: '这是一个危险操作，会导致分区下的所有对象无法正常访问，重新创建同名分区可恢复，你确定要这么做吗？'
  is-confirm-to-delete: '确认删除吗？'
  confirm-delete: '确认删除'
  add-bucket: '新建分区'
  update-hint: "影响已上传的对象"
  field-compress: '启用数据压缩'
  field-versioning: '开启对象多版本机制'
  field-readonly: '设为只读'
  field-name: '分区名称'
  field-version-remains: '最多保留版本数'
  field-store-strategy: '存储策略'
  field-data-shards: '数据副本数'
  field-parity-shards: '校验副本数'
</i18n>
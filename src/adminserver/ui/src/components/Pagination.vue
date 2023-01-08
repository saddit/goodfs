<template>
  <div class="flex justify-center items-center min-w-max">
    <button
        @click="switchPage(-1)"
        :class="curPage <= 1 ? 'opacity-0 cursor-default' : 'opacity-100'"
        class="switch-btn"
    >
      <font-awesome-icon icon="chevron-left"/>
    </button>
    <div class="grid grid-flow-col grid-rows-1 gap-x-3 mx-4">
      <transition-group name="list-complete">
        <template v-for="i in totalPage" :key="i">
          <div
              class="page-item text-sm"
              :class="[ curPage === i ? 'page-item-selected' : 'page-item-normal']"
              v-if="i >= start && i < start + maxNum"
              @click="setCurrentPage(i)"
          >
            {{ i }}
          </div>
        </template>
      </transition-group>
    </div>
    <button
        @click="switchPage(1)"
        :class="curPage >= totalPage ? 'opacity-0 cursor-default' : 'opacity-100'"
    >
      <font-awesome-icon icon="chevron-right" class="switch-btn"/>
    </button>
  </div>
</template>
<script>
export default {
    emits: ["onPageChange", "update:modelValue"],
    props: {
        total: Number,
        pageSize: Number,
        maxNum: {
            type: Number,
            default: () => 3,
        },
        modelValue: Number,
    },
    watch: {
        modelValue(nval) {
            this.curPage = nval;
        },
    },
    computed: {
        totalPage() {
            return Math.ceil(this.total / this.pageSize)
        }
    },
    data() {
        return {
            start: 1,
            curPage: this.modelValue || 1,
        };
    },
    methods: {
        setCurrentPage(i) {
            if (i <= this.totalPage && i > 0) {
                this.$emit("update:modelValue", i, this.curPage);
                this.curPage = i;
                this.$emit("onPageChange", this.curPage);
            }
        },
        switchPage(i) {
            this.setCurrentPage(this.curPage + i);
            this.changeStart();
        },
        changeStart() {
            if (this.curPage >= this.start + this.maxNum) {
                this.start += this.maxNum;
            } else if (this.curPage < this.start) {
                this.start -= this.maxNum;
            }
        },
    },
};
</script>
<style scoped>
.switch-btn {
    @apply text-lg text-indigo-600 transition-opacity duration-150;
}

.page-item {
    @apply rounded-full h-fit px-2 py-1 transition-all duration-200 cursor-pointer select-none block;
}

.page-item-normal {
    @apply text-indigo-600;
}

.page-item-selected {
    @apply text-white bg-indigo-600 font-bold;
}

.list-complete-item {
    transition: all 0.8s ease;
    display: inline-block;
    /* margin-right: 10px; */
}

.list-complete-enter-from,
.list-complete-leave-to {
    opacity: 0;
    transform: translateY(30px);
}

.list-complete-leave-active {
    position: absolute;
}
</style>
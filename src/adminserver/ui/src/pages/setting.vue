<template>
  <div class="w-1/4 divide-y divide-gray-400 mx-10">
    <!-- group1 -->
    <div class="panel">
      <div class="title">{{ t('general-setting') }}</div>
      <div class="form-body">
        <label>{{ t('locale-label') }}</label>
        <PriListBox @change="setLocale" v-model="$i18n.locale" :options="$i18n.availableLocales" :format="t" class="col-span-2">
        </PriListBox>
      </div>
    </div>
    <!-- group2 -->
    <div class="panel">
      <div class="title">{{ t('ui-setting') }}</div>
      <div class="form-body">
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const {t} = useI18n({
  inheritLocale: true
})

const store = useStore()

function setLocale(lang: string) {
  store.setLocale(lang)
}
</script>

<style scoped>
.panel {
  @apply pt-6 pb-10 px-2
}

.panel .title {
  @apply font-bold text-xl text-gray-900
}

.panel .form-body {
  @apply grid grid-cols-3 items-center mt-4 gap-y-2
}

.panel .form-body label {
  @apply text-gray-900 text-center
}
</style>

<route lang="json">
{
  "meta": {
    "title": "setting",
    "icon": "tools"
  }
}
</route>

<i18n lang="yaml">
en:
  general-setting: 'General'
  ui-setting: 'User Interface'
  locale-label: 'Locale'
  en: 'English(US)'
  zh: 'Chinese-Simple'
zh:
  general-setting: '常规'
  ui-setting: '界面'
  locale-label: '语言'
  en: '英语(美国)'
  zh: '中文简体'
</i18n>
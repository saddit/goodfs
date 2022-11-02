<template>
  <div class="bg-gray-50">
    <div class="mx-auto max-w-screen-xl px-4 py-12 sm:px-6 lg:flex lg:items-center lg:justify-between lg:py-16 lg:px-8">
      <h2 class="text-3xl font-extrabold leading-9 tracking-tight text-gray-900 sm:text-4xl sm:leading-10">
        {{ t('hello') }}? http-url: {{ getBaseUrl() }}
        <span class="text-indigo-600">Vite + Vue 3 + Tailwind CSS</span>
      </h2>
      <div class="mt-8 flex lg:mt-0 lg:flex-shrink-0">
        <div class="inline-flex rounded-md shadow space-x-2">
          <router-link to="/about" class="btn-pri">{{ t('about') }}</router-link>
          <router-link to="/setting" class="btn-pri">{{ t('setting') }}</router-link>
        </div>
        <ButtonRepo />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getBaseUrl } from "@/api/base";
import { useI18n } from "vue-i18n";

const { t } = useI18n({
  inheritLocale: true
})
</script>

<i18n lang="yaml">
en:
  hello: 'Ready to dive in'
zh:
  hello: '准备好了吗'
</i18n>

<route lang="json">
{
  "meta": {
    "title": "home"
  }
}
</route>
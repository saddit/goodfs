<route lang="json">
{
  "meta": {
    "title": "about",
    "icon": "info-circle"
  }
}
</route>

<script setup lang="ts">
import {EnvelopeIcon} from '@heroicons/vue/24/solid'

const {t} = useI18n({inheritLocale: true})
</script>

<template>
  <div class="bg-gray-50 px-4 sm:px-6 lg:px-8">
    <!-- page header -->
    <div class="mx-auto max-w-screen-xl py-12 lg:flex lg:items-center lg:justify-between lg:py-16">
      <h2 class="text-3xl font-extrabold leading-9 tracking-tight text-gray-900 sm:text-4xl sm:leading-10">
        {{ t('about-page') }}
        <br/>
        <span class="text-indigo-600">GoodFS Web Console</span>
      </h2>
      <div class="mt-8 flex lg:mt-0 lg:flex-shrink-0">
        <div class="inline-flex rounded-md shadow">
          <router-link to="/" class="rt-link bg-indigo-600 text-white">{{ t('back-home') }}</router-link>
        </div>
        <div class="ml-3 inline-flex rounded-md shadow">
          <a target="_blank"
             href="https://github.com/oss-go/goodfs"
             class="bg-white text-indigo-600 rt-link">Github Repo</a>
        </div>
      </div>
    </div>
    <article class="prose lg:prose-xl max-w-screen-xl mx-auto">
      <p>
        GoodFS is a highly focused distributed system for reading and writing file objects, which has excellent high
        availability, ultimate consistency, and powerful horizontal scalability.
      </p>
      <p>
        It is particularly suitable for scenarios where there are more reads than writes. In terms of object storage, it
        adopts multi-replica strategy and erasure code file repair technology, supports Bucket classification
        management, and flexible fine-grained object configuration, making it easy for users to manage and maintain
        stored objects. The data migration function of GoodFS includes both metadata and object data.
      </p>
    </article>
    <!-- page tail-->
    <div class="mx-auto max-w-screen-xl py-12 lg:py-16">
      <h3 class="text-xl font-extrabold leading-9 tracking-tight text-gray-900 sm:text-2xl sm:leading-10">
        {{ t('contact-us') }}
        <br/>
        <span class="text-lg font-bold text-gray-700"><EnvelopeIcon class="w-6 text-indigo-600 h-6 mr-1 inline"/>a838239178@gmail.com</span>
      </h3>
    </div>
  </div>
</template>

<style scoped>
.rt-link {
    @apply inline-flex items-center justify-center rounded-md border border-transparent px-5 py-3 text-base font-medium leading-6  transition duration-150 ease-in-out hover:bg-indigo-500 focus:outline-none
}
</style>

<i18n>
en:
  about-page: 'About Page'
  back-home: 'Back Home'
  contact-us: 'Contact Us'
zh:
  about-page: '关于我们'
  back-home: '返回首页'
  contact-us: '联系我们'
</i18n>
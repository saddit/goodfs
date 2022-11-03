<template>
<div>Home</div>
</template>

<script setup lang="ts">
import { getBaseUrl } from "@/api/base";
import { useI18n } from "vue-i18n";



const { t } = useI18n({
  inheritLocale: true
})
</script>

<i18n lang="yaml">
</i18n>

<route lang="json">
{
  "meta": {
    "title": "home",
    "icon": "house"
  }
}
</route>
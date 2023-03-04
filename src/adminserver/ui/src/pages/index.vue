<template>
<div>Home</div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

const route = useRoute()
const router = useRouter()

onBeforeMount(()=>{
    let redirect = route.query['redirect'] as string
    if (redirect) {
        delete route.query.redirect
        router.push({path: redirect, query: route.query})
    }
})

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
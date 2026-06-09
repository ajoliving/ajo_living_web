<!--
 * 服務式住宅列表頁。
 * 1. 承接服務式住宅主路由入口。
 * 2. 復用物業頻道列表呈現，固定服務式住宅頻道。
-->
<script setup lang="ts">
import PropertyChannelListPage from '@/pages/properties/Page.vue';
</script>

<template>
  <PropertyChannelListPage channel="serviced" />
</template>

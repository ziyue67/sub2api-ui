<template>
  <div class='scheme3-model-square-hint mb-8 mx-1 inline-flex items-center gap-3 rounded-2xl px-5 py-3.5 text-xs font-semibold'>
    <Icon name='infoCircle' size='xs' />
    <span>提示：下方价格为渠道基础价，实际扣费按所选分组倍率计算；专属倍率和高峰倍率会直接显示在分组标签中。</span>
  </div>
</template>

<script setup lang='ts'>
import Icon from '@/components/icons/Icon.vue'
</script>

<style scoped>
.scheme3-model-square-hint { border: 1px solid rgba(183,121,31,.28); border-radius: 6px; background: rgba(183,121,31,.08); color: #765213; box-shadow: none; }
:global(html.dark .scheme3-model-square-hint) { border-color: rgba(214,166,93,.35); background: rgba(183,121,31,.12); color: #e8c878; }
</style>

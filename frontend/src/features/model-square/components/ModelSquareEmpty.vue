<template>
  <div class='scheme3-model-square-empty py-32 text-center'>
    <div class='scheme3-model-square-empty-icon mx-auto mb-4 flex h-16 w-16 items-center justify-center'>
      <Icon name='search' size='md' />
    </div>
    <p class='text-sm font-bold text-gray-400 uppercase tracking-widest'>没有可展示的模型</p>
    <p class='mt-2 text-xs text-gray-500'>尝试调整搜索或平台筛选</p>
  </div>
</template>

<script setup lang='ts'>
import Icon from '@/components/icons/Icon.vue'
</script>

<style scoped>
.scheme3-model-square-empty { color: var(--scheme3-muted, #6b695f); }
.scheme3-model-square-empty-icon { border: 1px solid var(--scheme3-line, #dad5c8); border-radius: 7px; background: var(--scheme3-subtle, #f1eee6); color: var(--scheme3-muted, #6b695f); }
.scheme3-model-square-empty p { letter-spacing: 0; }
:global(html.dark .scheme3-model-square-empty-icon) { border-color: #47443a; background: #2b2924; color: #aaa69a; }
</style>

<template>
  <div class='scheme3-model-square-loading py-32 text-center'>
    <div class='inline-block h-10 w-10 animate-spin rounded-full border-4 border-indigo-500/20 border-t-indigo-400'></div>
  </div>
</template>

<style scoped>
.scheme3-model-square-loading { color: var(--scheme3-muted, #6b695f); }
.scheme3-model-square-loading > div { border-color: rgba(30,92,66,.18); border-top-color: #1e5c42; }
:global(html.dark .scheme3-model-square-loading > div) { border-color: rgba(143,194,165,.18); border-top-color: #8fc2a5; }
</style>

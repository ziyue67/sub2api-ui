<template>
  <div
    class='scheme3-model-square-background fixed inset-0 pointer-events-none z-0 overflow-hidden transition-colors duration-500'
    :class='[isDark ? "dark" : ""]'
  >
    <!-- Quiet paper texture; no decorative gradients or floating blobs. -->
    <div class='scheme3-model-square-background-grid absolute inset-0'></div>
  </div>
</template>

<script setup lang='ts'>
interface Props {
  isDark?: boolean
}

defineProps<Props>()
</script>

<style scoped>
.scheme3-model-square-background { background: #f4f2ec; }
.scheme3-model-square-background-grid {
  opacity: .32;
  background-image: linear-gradient(rgba(30, 92, 66, .055) 1px, transparent 1px), linear-gradient(90deg, rgba(30, 92, 66, .055) 1px, transparent 1px);
  background-size: 48px 48px;
}
:global(html.dark .scheme3-model-square-background) { background: #1b1b18; }
:global(html.dark .scheme3-model-square-background-grid) {
  opacity: .22;
  background-image: linear-gradient(rgba(143, 194, 165, .06) 1px, transparent 1px), linear-gradient(90deg, rgba(143, 194, 165, .06) 1px, transparent 1px);
}
</style>

<template>
  <Teleport to="body">
    <div
      class="pointer-events-none fixed right-4 top-4 z-[9999] space-y-3"
      aria-live="polite"
      aria-atomic="true"
    >
      <TransitionGroup
        enter-active-class="transition ease-out duration-300"
        enter-from-class="opacity-0 translate-x-full"
        enter-to-class="opacity-100 translate-x-0"
        leave-active-class="transition ease-in duration-200"
        leave-from-class="opacity-100 translate-x-0"
        leave-to-class="opacity-0 translate-x-full"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="[
            'scheme3-toast',
            `scheme3-toast--${toast.type}`
          ]"
        >
          <div class="scheme3-toast-content">
            <div class="scheme3-toast-row">
              <!-- Icon -->
              <div class="scheme3-toast-icon">
                <Icon
                  :name="getToastIconName(toast.type)"
                  size="md"
                  aria-hidden="true"
                />
              </div>

              <!-- Content -->
              <div class="scheme3-toast-copy">
                <p v-if="toast.title" class="scheme3-toast-title">
                  {{ toast.title }}
                </p>
                <p
                  :class="['scheme3-toast-message', { 'scheme3-toast-message-with-title': toast.title }]"
                >
                  {{ toast.message }}
                </p>
              </div>

              <!-- Close button -->
              <button
                @click="removeToast(toast.id)"
                class="scheme3-toast-close"
                aria-label="关闭通知"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </div>

          <!-- Progress bar -->
          <div v-if="toast.duration" class="scheme3-toast-track">
            <div
              class="scheme3-toast-progress"
              :style="{ animationDuration: `${toast.duration}ms` }"
            ></div>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const toasts = computed(() => appStore.toasts)

const getToastIconName = (type: string): 'checkCircle' | 'xCircle' | 'exclamationTriangle' | 'infoCircle' => {
  switch (type) {
    case 'success':
      return 'checkCircle'
    case 'error':
      return 'xCircle'
    case 'warning':
      return 'exclamationTriangle'
    case 'info':
    default:
      return 'infoCircle'
  }
}

const removeToast = (id: string) => {
  appStore.hideToast(id)
}
</script>

<style scoped>
.scheme3-toast {
  --toast-paper: #fffefa;
  --toast-ink: #24231f;
  --toast-muted: #777266;
  --toast-line: #d8d2c3;
  --toast-accent: #39706c;
  pointer-events: auto;
  min-width: min(20rem, calc(100vw - 2rem));
  max-width: 26rem;
  overflow: hidden;
  border: 1px solid var(--toast-line);
  border-left: 4px solid var(--toast-accent);
  border-radius: 7px;
  background: var(--toast-paper);
  box-shadow: 0 14px 30px rgba(54, 48, 34, .14);
}

.scheme3-toast--success { --toast-accent: #1e5c42; }
.scheme3-toast--error { --toast-accent: #a14234; }
.scheme3-toast--warning { --toast-accent: #a36a1d; }
.scheme3-toast--info { --toast-accent: #39706c; }

.scheme3-toast-content { padding: .82rem .9rem .78rem; }
.scheme3-toast-row { display: flex; align-items: flex-start; gap: .68rem; }
.scheme3-toast-icon { display: inline-flex; flex: 0 0 auto; margin-top: .08rem; color: var(--toast-accent); }
.scheme3-toast-copy { min-width: 0; flex: 1; }
.scheme3-toast-title { margin: 0; color: var(--toast-ink); font-size: .82rem; font-weight: 800; line-height: 1.45; }
.scheme3-toast-message { margin: 0; color: var(--toast-ink); font-size: .78rem; line-height: 1.55; overflow-wrap: anywhere; }
.scheme3-toast-message-with-title { margin-top: .16rem; color: var(--toast-muted); }
.scheme3-toast-close { display: inline-flex; width: 1.65rem; height: 1.65rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 0; border-radius: 5px; background: transparent; color: var(--toast-muted); cursor: pointer; transition: background-color 150ms ease, color 150ms ease, transform 150ms ease; }
.scheme3-toast-close:hover { background: #f1eee6; color: var(--toast-ink); }
.scheme3-toast-close:active { transform: scale(.95); }
.scheme3-toast-track { height: 3px; background: #ece8de; }

.scheme3-toast-progress {
  width: 100%;
  height: 100%;
  background: var(--toast-accent);
  animation-name: toast-progress-shrink;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}

:global(html.dark .scheme3-toast) { --toast-paper: #24231f; --toast-ink: #f4f2ec; --toast-muted: #aaa69a; --toast-line: #47443a; box-shadow: 0 16px 34px rgba(0, 0, 0, .3); }
:global(html.dark .scheme3-toast--success) { --toast-accent: #8fc2a5; }
:global(html.dark .scheme3-toast--error) { --toast-accent: #e7988b; }
:global(html.dark .scheme3-toast--warning) { --toast-accent: #e7bf78; }
:global(html.dark .scheme3-toast--info) { --toast-accent: #82b8ae; }
:global(html.dark .scheme3-toast-close:hover) { background: #2e2c27; }
:global(html.dark .scheme3-toast-track) { background: #34322c; }

@media (max-width: 480px) {
  .scheme3-toast { min-width: min(20rem, calc(100vw - 2rem)); max-width: calc(100vw - 2rem); }
}

@keyframes toast-progress-shrink {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}
</style>

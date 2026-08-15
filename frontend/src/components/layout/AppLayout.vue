<template>
  <Scheme3ConsoleLayout v-if="useScheme3Layout" :admin-mode="useScheme3AdminLayout">
    <slot />
  </Scheme3ConsoleLayout>

  <div v-else class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <!-- Background Decoration -->
    <div class="pointer-events-none fixed inset-0 bg-mesh-gradient"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main class="p-4 md:p-6 lg:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import Scheme3ConsoleLayout from './Scheme3ConsoleLayout.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')
// The third-version shell is the single authenticated application frame. Keep
// admin pages in the same shell so the upstream sidebar/header cannot leak
// back in when an operator leaves the channel monitor.
const useScheme3AdminLayout = computed(() =>
  authStore.isAuthenticated && isAdmin.value,
)
const useScheme3MonitorLayout = computed(() =>
  authStore.isAuthenticated && route.path === '/monitor',
)
const useScheme3Layout = computed(() =>
  authStore.isAuthenticated && (
    useScheme3AdminLayout.value || useScheme3MonitorLayout.value || (!isAdmin.value && route.meta.requiresAdmin !== true)
  ),
)

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: !useScheme3AdminLayout.value && !useScheme3MonitorLayout.value,
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>

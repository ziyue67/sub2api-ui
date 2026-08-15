import { computed, ref } from 'vue'
import { keysAPI } from '@/api/keys'
import { useAuthStore } from '@/stores/auth'
import type { ApiKey } from '@/types'

const loaded = ref(false)
const loading = ref(false)
const hasAllowedBatchImageKey = ref(false)
let pendingLoad: Promise<boolean> | null = null
let pendingUserId: number | null = null
let loadedUserId: number | null = null
const pageSize = 100

function keyAllowsBatchImage(key: ApiKey): boolean {
  return (
    key.status === 'active' &&
    key.group?.platform === 'gemini' &&
    key.group?.allow_batch_image_generation === true
  )
}

async function loadBatchImageAccess(force = false): Promise<boolean> {
  const authStore = useAuthStore()
  const userId = authStore.user?.id ?? null
  if (!authStore.isAuthenticated) {
    loaded.value = true
    hasAllowedBatchImageKey.value = false
    loadedUserId = null
    return false
  }

  // The composable is a module singleton, but API-key capabilities belong to
  // one authenticated user. Reset the snapshot whenever the identity changes.
  if (loadedUserId !== userId) {
    loaded.value = false
    hasAllowedBatchImageKey.value = false
    loadedUserId = userId
  }

  if (loaded.value && !force) {
    return hasAllowedBatchImageKey.value
  }

  if (pendingLoad && pendingUserId === userId && !force) {
    return pendingLoad
  }

  loading.value = true
  const requestUserId = userId
  const request = (async () => {
    let page = 1
    while (true) {
      const response = await keysAPI.list(page, pageSize, {
        status: 'active',
        sort_by: 'created_at',
        sort_order: 'desc'
      })

      if ((response.items || []).some(keyAllowsBatchImage)) {
        if (useAuthStore().user?.id === requestUserId) {
          hasAllowedBatchImageKey.value = true
          loaded.value = true
          loadedUserId = requestUserId
        }
        return true
      }

      if (page >= response.pages || (response.items || []).length === 0) {
        if (useAuthStore().user?.id === requestUserId) {
          hasAllowedBatchImageKey.value = false
          loaded.value = true
          loadedUserId = requestUserId
        }
        return false
      }

      page += 1
    }
  })()
    .catch(() => {
      if (useAuthStore().user?.id === requestUserId) {
        hasAllowedBatchImageKey.value = false
        loaded.value = true
        loadedUserId = requestUserId
      }
      return false
    })
    .finally(() => {
      if (pendingLoad === request) {
        loading.value = false
        pendingLoad = null
        pendingUserId = null
      }
    })

  pendingLoad = request
  pendingUserId = requestUserId
  return request
}

export function useBatchImageAccess() {
  const canUseBatchImage = computed(() => hasAllowedBatchImageKey.value)

  return {
    canUseBatchImage,
    batchImageAccessLoaded: computed(() => loaded.value),
    batchImageAccessLoading: computed(() => loading.value),
    refreshBatchImageAccess: loadBatchImageAccess,
  }
}

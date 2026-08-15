<template>
  <BaseDialog
    :show="show"
    :title="t('admin.channelMonitor.form.selectKeyTitle')"
    width="wide"
    content-class="scheme3-monitor-dialog"
    @close="$emit('close')"
  >
    <div class="space-y-3">
      <p class="scheme3-monitor-copy text-xs">
        {{ t('admin.channelMonitor.form.selectKeyHint') }}
      </p>

      <div class="relative">
        <input
          v-model="search"
          type="text"
          class="scheme3-monitor-input pl-9"
          :placeholder="t('keys.searchPlaceholder')"
        />
        <Icon name="search" size="sm" class="scheme3-monitor-input-icon absolute left-3 top-1/2 -translate-y-1/2" aria-hidden="true" />
      </div>

      <div v-if="loading" class="scheme3-monitor-empty py-6 text-center text-sm">
        {{ t('common.loading') }}
      </div>
      <div v-else-if="filteredKeys.length === 0" class="scheme3-monitor-empty py-6 text-center text-sm">
        {{ t('admin.channelMonitor.form.noActiveKey') }}
      </div>
      <div v-else class="scheme3-monitor-table-wrap max-h-96 overflow-auto">
        <table class="scheme3-monitor-table w-full text-sm">
          <thead class="scheme3-monitor-table-head sticky top-0 z-10">
            <tr class="text-left text-xs font-medium uppercase tracking-wider">
              <th class="px-3 py-2">{{ t('common.name') }}</th>
              <th class="px-3 py-2">{{ t('keys.apiKey') }}</th>
              <th class="px-3 py-2">{{ t('keys.group') }}</th>
            </tr>
          </thead>
          <tbody class="scheme3-monitor-table-body">
            <tr
              v-for="k in filteredKeys"
              :key="k.id"
              class="scheme3-monitor-table-row cursor-pointer"
              @click="$emit('pick', k)"
            >
              <td class="scheme3-monitor-table-primary px-3 py-2 font-medium">{{ k.name }}</td>
              <td class="scheme3-monitor-table-muted px-3 py-2 font-mono text-xs">{{ maskApiKey(k.key) }}</td>
              <td class="px-3 py-2">
                <GroupBadge
                  v-if="k.group"
                  scheme3
                  :name="k.group.name"
                  :platform="k.group.platform"
                  :subscription-type="k.group.subscription_type"
                  :rate-multiplier="k.group.rate_multiplier"
                  :user-rate-multiplier="userGroupRates[k.group.id]"
                />
                <span v-else class="scheme3-monitor-table-muted text-xs">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="scheme3-monitor-button">
          {{ t('common.cancel') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import type { Provider } from '@/api/admin/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import { maskApiKey } from '@/utils/maskApiKey'

const props = withDefaults(defineProps<{
  show: boolean
  loading: boolean
  keys: ApiKey[]
  provider: Provider
  userGroupRates?: Record<number, number>
}>(), {
  userGroupRates: () => ({}),
})

defineEmits<{
  (e: 'close'): void
  (e: 'pick', key: ApiKey): void
}>()

const { t } = useI18n()

const search = ref('')

watch(() => props.show, (shown) => {
  if (!shown) search.value = ''
})

const filteredKeys = computed<ApiKey[]>(() => {
  const q = search.value.trim().toLowerCase()
  return props.keys.filter((k) => {
    if (k.group?.platform !== props.provider) return false
    if (!q) return true
    return (
      k.name.toLowerCase().includes(q) ||
      k.key.toLowerCase().includes(q) ||
      (k.group?.name || '').toLowerCase().includes(q)
    )
  })
})
</script>

<template>
  <div
    v-if="currentAnnouncement"
    class="announcement-ticker"
    @mouseenter="pauseOnHover = true"
    @mouseleave="pauseOnHover = false"
  >
    <button
      type="button"
      class="announcement-ticker__main"
      :aria-label="`${t('announcements.title')}: ${currentAnnouncement.title}`"
      @click="openAnnouncement"
    >
      <Icon name="infoCircle" size="sm" class="shrink-0 text-primary-600 dark:text-primary-400" />
      <span class="announcement-ticker__title">{{ currentAnnouncement.title }}</span>
      <span v-if="unreadCount > 0" class="announcement-ticker__dot" aria-hidden="true"></span>
    </button>

    <div v-if="tickerAnnouncements.length > 1" class="announcement-ticker__controls">
      <button
        type="button"
        class="announcement-ticker__control"
        :aria-label="t('announcements.previous')"
        @click="previous"
      >
        <Icon name="chevronLeft" size="xs" />
      </button>
      <span class="announcement-ticker__count" aria-live="polite">
        {{ currentIndex + 1 }}/{{ tickerAnnouncements.length }}
      </span>
      <button
        type="button"
        class="announcement-ticker__control"
        :aria-label="t('announcements.next')"
        @click="next"
      >
        <Icon name="chevronRight" size="xs" />
      </button>
      <button
        type="button"
        class="announcement-ticker__control announcement-ticker__pause"
        :aria-label="isPaused ? t('announcements.resume') : t('announcements.pause')"
        @click="isPaused = !isPaused"
      >
        <Icon :name="isPaused ? 'play' : 'pause'" size="xs" />
      </button>
    </div>

    <AnnouncementPopup
      v-if="previewAnnouncement"
      :announcement="previewAnnouncement"
      preview
      @close="previewAnnouncement = null"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useAnnouncementStore } from '@/stores/announcements'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import AnnouncementPopup from './AnnouncementPopup.vue'

const ROTATION_MS = 6000
const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const { announcements } = storeToRefs(announcementStore)

const currentIndex = ref(0)
const isPaused = ref(false)
const pauseOnHover = ref(false)
const previewAnnouncement = ref<UserAnnouncement | null>(null)
let rotationTimer: ReturnType<typeof setInterval> | undefined

const tickerAnnouncements = computed(() => announcements.value
  .filter((item) => item.ticker_enabled !== false)
  .sort((a, b) => b.priority - a.priority || b.id - a.id))
const currentAnnouncement = computed(() => tickerAnnouncements.value[currentIndex.value] ?? null)
const unreadCount = computed(() => announcementStore.unreadCount)

function next() {
  if (tickerAnnouncements.value.length < 2) return
  currentIndex.value = (currentIndex.value + 1) % tickerAnnouncements.value.length
}

function previous() {
  if (tickerAnnouncements.value.length < 2) return
  currentIndex.value = (currentIndex.value - 1 + tickerAnnouncements.value.length) % tickerAnnouncements.value.length
}

function openAnnouncement() {
  const announcement = currentAnnouncement.value
  if (!announcement) return
  previewAnnouncement.value = announcement
  if (!announcement.read_at) {
    void announcementStore.markAsRead(announcement.id)
  }
}

function resetIndex() {
  if (currentIndex.value >= tickerAnnouncements.value.length) currentIndex.value = 0
}

function startRotation() {
  if (rotationTimer || tickerAnnouncements.value.length < 2) return
  rotationTimer = setInterval(() => {
    if (!isPaused.value && !pauseOnHover.value && !previewAnnouncement.value) next()
  }, ROTATION_MS)
}

function stopRotation() {
  if (rotationTimer) clearInterval(rotationTimer)
  rotationTimer = undefined
}

watch(tickerAnnouncements, () => {
  resetIndex()
  stopRotation()
  startRotation()
}, { deep: true })

onMounted(startRotation)
onBeforeUnmount(stopRotation)
</script>

<style scoped>
.announcement-ticker {
  display: flex;
  min-width: 0;
  max-width: min(42vw, 360px);
  align-items: center;
  overflow: hidden;
  border: 1px solid rgb(226 232 240 / 0.8);
  border-radius: 8px;
  background: rgb(248 250 252 / 0.8);
}

.dark .announcement-ticker {
  border-color: rgb(71 85 105 / 0.7);
  background: rgb(30 41 59 / 0.65);
}

.announcement-ticker__main {
  display: flex;
  position: relative;
  min-width: 0;
  flex: 1;
  align-items: center;
  gap: 0.45rem;
  overflow: hidden;
  padding: 0.4rem 0.55rem;
  text-align: left;
}

.announcement-ticker__main:hover,
.announcement-ticker__control:hover {
  background: rgb(226 232 240 / 0.7);
}

.dark .announcement-ticker__main:hover,
.dark .announcement-ticker__control:hover {
  background: rgb(51 65 85 / 0.75);
}

.announcement-ticker__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(51 65 85);
}

.dark .announcement-ticker__title {
  color: rgb(226 232 240);
}

.announcement-ticker__dot {
  width: 0.4rem;
  height: 0.4rem;
  flex: 0 0 auto;
  border-radius: 9999px;
  background: rgb(239 68 68);
}

.announcement-ticker__controls {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  border-left: 1px solid rgb(226 232 240 / 0.8);
}

.dark .announcement-ticker__controls {
  border-left-color: rgb(71 85 105 / 0.7);
}

.announcement-ticker__control {
  display: flex;
  width: 1.65rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  color: rgb(100 116 139);
}

.announcement-ticker__count {
  min-width: 2.2rem;
  text-align: center;
  font-size: 0.65rem;
  color: rgb(100 116 139);
}

@media (max-width: 767px) {
  .announcement-ticker {
    max-width: 2.25rem;
    border-color: transparent;
    background: transparent;
  }

  .announcement-ticker__main {
    width: 2.25rem;
    flex: 0 0 2.25rem;
    justify-content: center;
    padding: 0;
  }

  .announcement-ticker__title,
  .announcement-ticker__controls {
    display: none;
  }

  .announcement-ticker__dot {
    position: absolute;
    top: 0.25rem;
    right: 0.25rem;
  }
}
</style>

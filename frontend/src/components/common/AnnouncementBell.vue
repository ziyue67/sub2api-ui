<template>
  <div class="scheme3-announcement-bell">
    <button type="button" class="scheme3-announcement-trigger" :class="{ 'scheme3-announcement-trigger-unread': unreadCount > 0 }" :aria-label="t('announcements.title')" :title="t('announcements.title')" @click="openModal">
      <Icon name="bell" size="md" />
      <span v-if="unreadCount > 0" class="scheme3-announcement-trigger-dot"></span>
    </button>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="isModalOpen" class="scheme3-announcement-backdrop" @click="closeModal">
          <section class="scheme3-announcement-panel" role="dialog" aria-modal="true" aria-labelledby="announcements-list-title" @click.stop>
            <header class="scheme3-announcement-panel-header">
              <div class="scheme3-announcement-heading">
                <span class="scheme3-announcement-heading-mark"><Icon name="bell" size="sm" /></span>
                <div>
                  <p class="scheme3-announcement-kicker">SHOUR OR TOKEN / 消息</p>
                  <h2 id="announcements-list-title">{{ t('announcements.title') }}</h2>
                </div>
              </div>
              <div class="scheme3-announcement-header-actions">
                <button v-if="unreadCount > 0" type="button" class="scheme3-announcement-quiet-button" :disabled="loading" @click="markAllAsRead">{{ t('announcements.markAllRead') }}</button>
                <button type="button" class="scheme3-announcement-icon-button" :aria-label="t('common.close')" :title="t('common.close')" @click="closeModal"><Icon name="x" size="sm" /></button>
              </div>
            </header>

            <div class="scheme3-announcement-panel-note">
              <span class="scheme3-announcement-status-dot" :class="{ 'scheme3-announcement-status-dot-muted': unreadCount === 0 }"></span>
              <span v-if="unreadCount > 0">{{ unreadCount }} {{ t('announcements.unread') }}</span>
              <span v-else>{{ t('announcements.readStatus') }}</span>
            </div>

            <div class="scheme3-announcement-list">
              <div v-if="loading" class="scheme3-announcement-loading"><span></span></div>
              <div v-else-if="announcements.length > 0" class="scheme3-announcement-items">
                <button v-for="item in announcements" :key="item.id" type="button" class="scheme3-announcement-row" :class="{ 'scheme3-announcement-row-unread': !item.read_at }" @click="openDetail(item)">
                  <span class="scheme3-announcement-row-icon"><Icon :name="item.read_at ? 'check' : 'bell'" size="sm" /></span>
                  <span class="min-w-0 flex-1 text-left">
                    <span class="scheme3-announcement-row-title">{{ item.title }}</span>
                    <span class="scheme3-announcement-row-meta"><time>{{ formatRelativeTime(item.created_at) }}</time><span v-if="!item.read_at" class="scheme3-announcement-unread-label">{{ t('announcements.unread') }}</span></span>
                  </span>
                  <Icon name="chevronRight" size="sm" class="scheme3-announcement-row-arrow" />
                </button>
              </div>
              <div v-else class="scheme3-announcement-empty"><span class="scheme3-announcement-empty-mark"><Icon name="inbox" size="lg" /></span><p>{{ t('announcements.empty') }}</p><span>{{ t('announcements.emptyDescription') }}</span></div>
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="detailModalOpen && selectedAnnouncement" class="scheme3-announcement-backdrop scheme3-announcement-detail-backdrop" @click="closeDetail">
          <section class="scheme3-announcement-panel scheme3-announcement-detail-panel" role="dialog" aria-modal="true" aria-labelledby="announcement-detail-title" @click.stop>
            <header class="scheme3-announcement-panel-header">
              <div class="scheme3-announcement-heading min-w-0">
                <span class="scheme3-announcement-heading-mark"><Icon name="document" size="sm" /></span>
                <div class="min-w-0">
                  <p class="scheme3-announcement-kicker">{{ t('announcements.title') }}</p>
                  <h2 id="announcement-detail-title" class="scheme3-announcement-detail-title">{{ selectedAnnouncement.title }}</h2>
                  <div class="scheme3-announcement-detail-meta"><Icon name="clock" size="xs" /><time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time></div>
                </div>
              </div>
              <button type="button" class="scheme3-announcement-icon-button" :aria-label="t('common.close')" :title="t('common.close')" @click="closeDetail"><Icon name="x" size="sm" /></button>
            </header>

            <div class="scheme3-announcement-detail-body"><div class="scheme3-announcement-popup-rule"></div><div class="scheme3-announcement-popup-content markdown-body prose prose-sm max-w-none dark:prose-invert" v-html="renderMarkdown(selectedAnnouncement.content)"></div></div>

            <footer class="scheme3-announcement-detail-footer">
              <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
              <div class="flex items-center gap-2">
                <button type="button" class="scheme3-announcement-quiet-button" @click="closeDetail">{{ t('common.close') }}</button>
                <button v-if="!selectedAnnouncement.read_at" type="button" class="scheme3-announcement-primary-button" @click="markAsReadAndClose(selectedAnnouncement.id)"><Icon name="check" size="sm" />{{ t('announcements.markRead') }}</button>
              </div>
            </footer>
          </section>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import '@/styles/announcement-markdown.css'

const { t } = useI18n()
const appStore = useAppStore()
const announcementStore = useAnnouncementStore()
const { announcements, loading } = storeToRefs(announcementStore)
const unreadCount = computed(() => announcementStore.unreadCount)
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)

marked.setOptions({ breaks: true, gfm: true })

function renderMarkdown(content: string): string {
  return content ? DOMPurify.sanitize(marked.parse(content) as string) : ''
}
function openModal() { isModalOpen.value = true }
function closeModal() { isModalOpen.value = false }
function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) void markAsRead(announcement.id)
}
function closeDetail() { detailModalOpen.value = false; selectedAnnouncement.value = null }
async function markAsRead(id: number) {
  try { await announcementStore.markAsRead(id) } catch (err: any) { appStore.showError(err?.message || t('common.unknownError')) }
}
async function markAsReadAndClose(id: number) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}
async function markAllAsRead() {
  try { await announcementStore.markAllAsRead(); appStore.showSuccess(t('announcements.allMarkedAsRead')) } catch (err: any) { appStore.showError(err?.message || t('common.unknownError')) }
}
function handleEscape(event: KeyboardEvent) {
  if (event.key !== 'Escape') return
  if (detailModalOpen.value) closeDetail()
  else if (isModalOpen.value) closeModal()
}

onMounted(() => document.addEventListener('keydown', handleEscape))
onBeforeUnmount(() => { document.removeEventListener('keydown', handleEscape); document.body.style.overflow = '' })
watch([isModalOpen, detailModalOpen, () => announcementStore.currentPopup], ([modal, detail, popup]) => { document.body.style.overflow = modal || detail || popup ? 'hidden' : '' })
</script>

<style scoped>
.scheme3-announcement-bell { display: inline-flex; }
.scheme3-announcement-trigger { position: relative; display: inline-flex; width: 2.2rem; height: 2.2rem; align-items: center; justify-content: center; border: 1px solid transparent; border-radius: 7px; background: transparent; color: #6b695f; transition: color 150ms ease,background-color 150ms ease,transform 150ms ease; }
.scheme3-announcement-trigger:hover { background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-announcement-trigger:active { transform: scale(.95); }
.scheme3-announcement-trigger:focus-visible,.scheme3-announcement-icon-button:focus-visible,.scheme3-announcement-quiet-button:focus-visible,.scheme3-announcement-primary-button:focus-visible,.scheme3-announcement-row:focus-visible { outline: 3px solid rgba(30,92,66,.24); outline-offset: 2px; }
.scheme3-announcement-trigger-unread { color: #1e5c42; }
.scheme3-announcement-trigger-dot { position: absolute; top: .42rem; right: .42rem; width: .43rem; height: .43rem; border: 2px solid #f4f2ec; border-radius: 50%; background: #b7791f; box-shadow: 0 0 0 .15rem rgba(183,121,31,.12); }
.scheme3-announcement-backdrop { position: fixed; z-index: 100; inset: 0; display: flex; align-items: flex-start; justify-content: center; overflow-y: auto; padding: clamp(1rem, 7vh, 4.5rem) 1rem 1rem; background: rgba(22,21,15,.58); backdrop-filter: blur(5px); }
.scheme3-announcement-detail-backdrop { z-index: 110; }
.scheme3-announcement-panel { width: min(100%, 39rem); overflow: hidden; border: 1px solid #dad5c8; border-radius: 8px; background: #fbfaf6; box-shadow: 0 24px 70px rgba(22,21,15,.26); color: #16150f; }
.scheme3-announcement-detail-panel { width: min(100%, 47rem); }
.scheme3-announcement-panel-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; border-bottom: 1px solid #dad5c8; padding: 1.1rem 1.25rem; background: #f4f2ec; }
.scheme3-announcement-heading { display: flex; align-items: center; gap: .7rem; }
.scheme3-announcement-heading-mark { display: inline-flex; width: 2.15rem; height: 2.15rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 1px solid rgba(30,92,66,.34); border-radius: 7px; background: #1e5c42; color: #f4f2ec; }
.scheme3-announcement-kicker { margin: 0 0 .22rem; color: #6b695f; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .54rem; font-weight: 800; letter-spacing: .12em; }
.scheme3-announcement-heading h2 { margin: 0; color: #16150f; font-family: Georgia,'Times New Roman',serif; font-size: 1.22rem; font-weight: 600; letter-spacing: 0; }
.scheme3-announcement-header-actions { display: flex; align-items: center; gap: .35rem; }
.scheme3-announcement-icon-button { display: inline-flex; width: 2rem; height: 2rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 1px solid transparent; border-radius: 6px; background: transparent; color: #6b695f; transition: background-color 150ms ease,color 150ms ease,transform 150ms ease; }
.scheme3-announcement-icon-button:hover { background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-announcement-icon-button:active { transform: scale(.95); }
.scheme3-announcement-panel-note { display: flex; align-items: center; gap: .45rem; border-bottom: 1px solid #dad5c8; padding: .58rem 1.25rem; background: #fbfaf6; color: #6b695f; font-size: .68rem; }
.scheme3-announcement-status-dot { width: .38rem; height: .38rem; flex: 0 0 auto; border-radius: 50%; background: #b7791f; box-shadow: 0 0 0 .18rem rgba(183,121,31,.12); }
.scheme3-announcement-status-dot-muted { background: #aaa69a; box-shadow: none; }
.scheme3-announcement-list { max-height: min(60vh, 32rem); overflow-y: auto; }
.scheme3-announcement-items { display: flex; flex-direction: column; }
.scheme3-announcement-row { position: relative; display: flex; min-height: 4.65rem; align-items: center; gap: .8rem; border: 0; border-bottom: 1px solid #e4e0d6; padding: .8rem 1.25rem; background: #fbfaf6; color: #16150f; transition: background-color 150ms ease,transform 150ms ease; }
.scheme3-announcement-row:hover { background: #f1eee6; }
.scheme3-announcement-row:active { transform: scale(.995); }
.scheme3-announcement-row-unread { background: rgba(30,92,66,.045); }
.scheme3-announcement-row-unread::before { position: absolute; top: .9rem; bottom: .9rem; left: 0; width: 3px; border-radius: 0 2px 2px 0; background: #1e5c42; content: ''; }
.scheme3-announcement-row-icon { display: inline-flex; width: 2.25rem; height: 2.25rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 1px solid #dad5c8; border-radius: 7px; background: #f4f2ec; color: #6b695f; }
.scheme3-announcement-row-unread .scheme3-announcement-row-icon { border-color: rgba(30,92,66,.32); background: #1e5c42; color: #f4f2ec; }
.scheme3-announcement-row-title { display: block; overflow: hidden; color: #16150f; font-size: .8rem; font-weight: 800; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-announcement-row-meta { display: flex; align-items: center; gap: .45rem; margin-top: .26rem; color: #6b695f; font-size: .65rem; }
.scheme3-announcement-unread-label { border-radius: 3px; padding: .1rem .28rem; background: rgba(30,92,66,.1); color: #1e5c42; font-size: .56rem; font-weight: 800; }
.scheme3-announcement-row-arrow { flex: 0 0 auto; color: #aaa69a; transition: transform 150ms ease,color 150ms ease; }
.scheme3-announcement-row:hover .scheme3-announcement-row-arrow { transform: translateX(2px); color: #1e5c42; }
.scheme3-announcement-empty { display: flex; min-height: 16rem; flex-direction: column; align-items: center; justify-content: center; padding: 2rem; color: #6b695f; text-align: center; }
.scheme3-announcement-empty-mark { display: inline-flex; width: 3.5rem; height: 3.5rem; align-items: center; justify-content: center; border: 1px solid #dad5c8; border-radius: 7px; background: #f4f2ec; color: #aaa69a; }
.scheme3-announcement-empty p { margin: .85rem 0 .2rem; color: #16150f; font-size: .8rem; font-weight: 800; }
.scheme3-announcement-empty > span:last-child { font-size: .68rem; }
.scheme3-announcement-detail-title { overflow: hidden; max-width: 34rem; text-overflow: ellipsis; white-space: nowrap; }
.scheme3-announcement-detail-meta { display: flex; align-items: center; gap: .3rem; margin-top: .36rem; color: #6b695f; font-size: .67rem; }
.scheme3-announcement-detail-body { position: relative; max-height: min(60vh, 38rem); overflow-y: auto; padding: 1.45rem 1.5rem 1.6rem 1.8rem; background: #fbfaf6; }
.scheme3-announcement-popup-rule { position: absolute; top: 1.5rem; bottom: 1.65rem; left: 1.15rem; width: 3px; border-radius: 1px; background: #b7791f; }
.scheme3-announcement-popup-content { color: #36342d; }
.scheme3-announcement-detail-footer { display: flex; align-items: center; justify-content: space-between; gap: 1rem; border-top: 1px solid #dad5c8; padding: .9rem 1.25rem; background: #f4f2ec; color: #6b695f; font-size: .68rem; }
.scheme3-announcement-quiet-button { display: inline-flex; min-height: 2rem; align-items: center; justify-content: center; border: 1px solid #dad5c8; border-radius: 6px; padding: .35rem .62rem; background: #fbfaf6; color: #36342d; font-size: .64rem; font-weight: 800; transition: border-color 150ms ease,background-color 150ms ease,color 150ms ease; }
.scheme3-announcement-quiet-button:hover { border-color: rgba(30,92,66,.38); background: rgba(30,92,66,.07); color: #1e5c42; }
.scheme3-announcement-primary-button { display: inline-flex; min-height: 2rem; align-items: center; justify-content: center; gap: .35rem; border: 1px solid #1e5c42; border-radius: 6px; padding: .35rem .65rem; background: #1e5c42; color: #f4f2ec; font-size: .64rem; font-weight: 800; transition: background-color 150ms ease,transform 150ms ease,box-shadow 150ms ease; }
.scheme3-announcement-primary-button:hover { background: #174a35; box-shadow: 0 7px 16px rgba(30,92,66,.18); }
.scheme3-announcement-primary-button:active { transform: scale(.98); }
.modal-fade-enter-active { transition: opacity 180ms ease; }
.modal-fade-leave-active { transition: opacity 140ms ease; }
.modal-fade-enter-from,.modal-fade-leave-to { opacity: 0; }
.modal-fade-enter-from .scheme3-announcement-panel { transform: translateY(-8px) scale(.985); }
.modal-fade-leave-to .scheme3-announcement-panel { transform: translateY(-4px) scale(.99); }
.scheme3-announcement-panel { transition: transform 180ms ease; }
.scheme3-announcement-list::-webkit-scrollbar,.scheme3-announcement-detail-body::-webkit-scrollbar { width: 7px; }
.scheme3-announcement-list::-webkit-scrollbar-thumb,.scheme3-announcement-detail-body::-webkit-scrollbar-thumb { border: 2px solid #fbfaf6; border-radius: 999px; background: #bcb6a8; }
@media (max-width: 640px) { .scheme3-announcement-backdrop { padding: .75rem; } .scheme3-announcement-panel-header { padding: 1rem; } .scheme3-announcement-panel-note { padding-right: 1rem; padding-left: 1rem; } .scheme3-announcement-row { padding-right: 1rem; padding-left: 1rem; } .scheme3-announcement-detail-body { max-height: calc(100vh - 13rem); padding: 1.15rem 1rem 1.25rem 1.45rem; } .scheme3-announcement-popup-rule { top: 1.18rem; bottom: 1.28rem; left: .78rem; } .scheme3-announcement-detail-footer { align-items: flex-end; padding: .8rem 1rem; } .scheme3-announcement-detail-footer > span { display: none; } .scheme3-announcement-detail-title { max-width: 13rem; } }
</style>

<style>
/* The list and detail panels are teleported to body, outside this component tree. */
html.dark .scheme3-announcement-trigger { color: #aaa69a; }
html.dark .scheme3-announcement-trigger:hover,
html.dark .scheme3-announcement-trigger-unread { color: #8fc2a5; }
html.dark .scheme3-announcement-trigger-dot { border-color: #1b1b18; }
html.dark .scheme3-announcement-panel { border-color: #47443a; background: #24231f; color: #f4f2ec; }
html.dark .scheme3-announcement-panel-header,
html.dark .scheme3-announcement-detail-footer { border-color: #47443a; background: #1b1b18; }
html.dark .scheme3-announcement-panel-note,
html.dark .scheme3-announcement-row,
html.dark .scheme3-announcement-detail-body { border-color: #47443a; background: #24231f; }
html.dark .scheme3-announcement-heading h2,
html.dark .scheme3-announcement-row-title,
html.dark .scheme3-announcement-empty p { color: #f4f2ec; }
html.dark .scheme3-announcement-kicker,
html.dark .scheme3-announcement-panel-note,
html.dark .scheme3-announcement-row-meta,
html.dark .scheme3-announcement-detail-meta,
html.dark .scheme3-announcement-detail-footer,
html.dark .scheme3-announcement-empty { color: #aaa69a; }
html.dark .scheme3-announcement-row:hover { background: #2b2924; }
html.dark .scheme3-announcement-row-unread { background: rgba(143,194,165,.08); }
html.dark .scheme3-announcement-row-icon,
html.dark .scheme3-announcement-empty-mark { border-color: #47443a; background: #2b2924; color: #aaa69a; }
html.dark .scheme3-announcement-quiet-button { border-color: #47443a; background: #24231f; color: #dedbd1; }
html.dark .scheme3-announcement-list::-webkit-scrollbar-thumb,
html.dark .scheme3-announcement-detail-body::-webkit-scrollbar-thumb { border-color: #24231f; background: #5e5a4f; }
</style>

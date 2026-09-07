<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="displayedAnnouncement"
        class="scheme3-announcement-backdrop"
        @click="handleDismiss"
      >
        <section class="scheme3-announcement-popup" role="dialog" aria-modal="true" aria-labelledby="announcement-popup-title" @click.stop>
          <header class="scheme3-announcement-popup-header">
            <div class="scheme3-announcement-popup-mark"><Icon name="bell" size="md" /></div>
            <div class="min-w-0 flex-1">
              <div class="scheme3-announcement-popup-eyebrow">
                <span class="scheme3-announcement-status-dot"></span>
                {{ t('announcements.unread') }}
              </div>
              <h2 id="announcement-popup-title" class="scheme3-announcement-popup-title">{{ displayedAnnouncement.title }}</h2>
              <div class="scheme3-announcement-popup-meta"><Icon name="clock" size="xs" /><time>{{ formatRelativeWithDateTime(displayedAnnouncement.created_at) }}</time></div>
            </div>
            <button type="button" class="scheme3-announcement-icon-button" :aria-label="t('common.close')" :title="t('common.close')" @click="handleDismiss">
              <Icon name="x" size="sm" />
            </button>
          </header>

          <div class="scheme3-announcement-popup-body">
            <div class="scheme3-announcement-popup-rule"></div>
            <div class="scheme3-announcement-popup-content markdown-body prose prose-sm max-w-none dark:prose-invert" v-html="renderedContent"></div>
          </div>

          <footer class="scheme3-announcement-popup-footer">
            <span>{{ preview ? t('announcements.title') : t('announcements.markReadHint') }}</span>
            <button type="button" class="scheme3-announcement-primary-button" data-testid="announcement-popup-dismiss" @click="handleDismiss">
              <Icon :name="preview ? 'x' : 'check'" size="sm" />
              {{ preview ? t('common.close') : t('announcements.markRead') }}
            </button>
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import type { Announcement, UserAnnouncement } from '@/types'
import '@/styles/announcement-markdown.css'

type PreviewAnnouncement = Pick<Announcement | UserAnnouncement, 'title' | 'content' | 'created_at'>

const props = withDefaults(defineProps<{
  announcement?: PreviewAnnouncement | null
  preview?: boolean
}>(), {
  announcement: null,
  preview: false,
})

const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const displayedAnnouncement = computed(() => (props.preview ? props.announcement : announcementStore.currentPopup))

marked.setOptions({ breaks: true, gfm: true })

const renderedContent = computed(() => {
  const content = displayedAnnouncement.value?.content
  return content ? DOMPurify.sanitize(marked.parse(content) as string) : ''
})

function handleDismiss() {
  if (props.preview) {
    emit('close')
    return
  }
  announcementStore.dismissPopup()
}

watch(
  displayedAnnouncement,
  (popup) => {
    if (popup) document.body.style.overflow = 'hidden'
    else if (props.preview) document.body.style.overflow = ''
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (props.preview) document.body.style.overflow = ''
})
</script>

<style scoped>
.scheme3-announcement-backdrop { position: fixed; z-index: 120; inset: 0; display: flex; align-items: flex-start; justify-content: center; overflow-y: auto; padding: clamp(1rem, 7vh, 4.5rem) 1rem 1rem; background: rgba(22,21,15,.58); backdrop-filter: blur(5px); }
.scheme3-announcement-popup { width: min(100%, 42.5rem); overflow: hidden; border: 1px solid #dad5c8; border-radius: 8px; background: #fbfaf6; box-shadow: 0 24px 70px rgba(22,21,15,.26); color: #16150f; }
.scheme3-announcement-popup-header { display: flex; align-items: flex-start; gap: .8rem; border-bottom: 1px solid #dad5c8; padding: 1.25rem 1.4rem 1.15rem; background: #f4f2ec; }
.scheme3-announcement-popup-mark { display: inline-flex; width: 2.4rem; height: 2.4rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 1px solid rgba(30,92,66,.34); border-radius: 7px; background: #1e5c42; color: #f4f2ec; box-shadow: 0 7px 16px rgba(30,92,66,.15); }
.scheme3-announcement-popup-eyebrow { display: flex; align-items: center; gap: .38rem; color: #1e5c42; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size: .61rem; font-weight: 800; letter-spacing: .08em; text-transform: uppercase; }
.scheme3-announcement-status-dot { width: .38rem; height: .38rem; border-radius: 50%; background: #b7791f; box-shadow: 0 0 0 .18rem rgba(183,121,31,.13); }
.scheme3-announcement-popup-title { margin: .45rem 0 .35rem; color: #16150f; font-family: Georgia,'Times New Roman',serif; font-size: clamp(1.28rem, 2.5vw, 1.7rem); font-weight: 600; line-height: 1.22; letter-spacing: 0; }
.scheme3-announcement-popup-meta { display: flex; align-items: center; gap: .32rem; color: #6b695f; font-size: .72rem; }
.scheme3-announcement-icon-button { display: inline-flex; width: 2rem; height: 2rem; flex: 0 0 auto; align-items: center; justify-content: center; border: 1px solid transparent; border-radius: 6px; background: transparent; color: #6b695f; transition: background-color 150ms ease,color 150ms ease,transform 150ms ease; }
.scheme3-announcement-icon-button:hover { background: rgba(30,92,66,.08); color: #1e5c42; }
.scheme3-announcement-icon-button:active { transform: scale(.95); }
.scheme3-announcement-icon-button:focus-visible,.scheme3-announcement-primary-button:focus-visible { outline: 3px solid rgba(30,92,66,.24); outline-offset: 2px; }
.scheme3-announcement-popup-body { position: relative; max-height: min(50vh, 32rem); overflow-y: auto; padding: 1.45rem 1.5rem 1.6rem 1.8rem; background: #fbfaf6; }
.scheme3-announcement-popup-rule { position: absolute; top: 1.5rem; bottom: 1.65rem; left: 1.15rem; width: 3px; border-radius: 1px; background: #b7791f; }
.scheme3-announcement-popup-content { color: #36342d; }
.scheme3-announcement-popup-footer { display: flex; align-items: center; justify-content: space-between; gap: 1rem; border-top: 1px solid #dad5c8; padding: .95rem 1.4rem; background: #f4f2ec; color: #6b695f; font-size: .68rem; }
.scheme3-announcement-primary-button { display: inline-flex; min-height: 2.35rem; align-items: center; justify-content: center; gap: .42rem; border: 1px solid #1e5c42; border-radius: 6px; padding: .48rem .8rem; background: #1e5c42; color: #f4f2ec; font-size: .7rem; font-weight: 800; transition: background-color 150ms ease,transform 150ms ease,box-shadow 150ms ease; }
.scheme3-announcement-primary-button:hover { background: #174a35; box-shadow: 0 8px 17px rgba(30,92,66,.18); }
.scheme3-announcement-primary-button:active { transform: scale(.98); }
.popup-fade-enter-active { transition: opacity 180ms ease; }
.popup-fade-leave-active { transition: opacity 140ms ease; }
.popup-fade-enter-from,.popup-fade-leave-to { opacity: 0; }
.popup-fade-enter-from .scheme3-announcement-popup { transform: translateY(-8px) scale(.985); }
.popup-fade-leave-to .scheme3-announcement-popup { transform: translateY(-4px) scale(.99); }
.scheme3-announcement-popup { transition: transform 180ms ease; }
.scheme3-announcement-popup-body::-webkit-scrollbar { width: 7px; }
.scheme3-announcement-popup-body::-webkit-scrollbar-thumb { border: 2px solid #fbfaf6; border-radius: 999px; background: #bcb6a8; }
@media (max-width: 640px) { .scheme3-announcement-backdrop { align-items: flex-start; padding: .75rem; } .scheme3-announcement-popup-header { padding: 1rem; } .scheme3-announcement-popup-body { max-height: calc(100vh - 14rem); padding: 1.15rem 1rem 1.25rem 1.45rem; } .scheme3-announcement-popup-rule { top: 1.18rem; bottom: 1.28rem; left: .78rem; } .scheme3-announcement-popup-footer { align-items: flex-end; padding: .8rem 1rem; } }
</style>

<style>
/* Teleported dialogs are direct body children, so their dark palette must remain global. */
html.dark .scheme3-announcement-popup { border-color: #47443a; background: #24231f; color: #f4f2ec; }
html.dark .scheme3-announcement-popup-header,
html.dark .scheme3-announcement-popup-footer { border-color: #47443a; background: #1b1b18; }
html.dark .scheme3-announcement-popup-title { color: #f4f2ec; }
html.dark .scheme3-announcement-popup-meta,
html.dark .scheme3-announcement-popup-footer { color: #aaa69a; }
html.dark .scheme3-announcement-popup-body { background: #24231f; }
html.dark .scheme3-announcement-popup-content { color: #dedbd1; }
html.dark .scheme3-announcement-popup-body::-webkit-scrollbar-thumb { border-color: #24231f; background: #5e5a4f; }
</style>

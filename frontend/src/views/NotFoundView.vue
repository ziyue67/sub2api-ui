<template>
  <main class="scheme3-not-found">
    <div class="scheme3-not-found-rule" aria-hidden="true"></div>
    <header class="scheme3-not-found-brand">
      <span class="scheme3-not-found-mark">
        <img v-if="siteLogo" :src="siteLogo" alt="" />
        <span v-else>ST</span>
      </span>
      <span>
        <small>SHOUR OR TOKEN / 页面记录</small>
        <strong>{{ siteName }}</strong>
      </span>
    </header>

    <section class="scheme3-not-found-panel" aria-labelledby="not-found-title">
      <div class="scheme3-not-found-index" aria-hidden="true">404</div>
      <div class="scheme3-not-found-copy">
        <p class="scheme3-not-found-kicker">访问记录 / 暂无此页</p>
        <h1 id="not-found-title">页面未找到</h1>
        <p>这个地址暂时没有对应内容，可能已经移动，或者链接填写有误。</p>
      </div>
      <div class="scheme3-not-found-actions">
        <button type="button" class="scheme3-not-found-button scheme3-not-found-button-secondary" @click="goBack">
          <Icon name="arrowLeft" size="sm" />
          <span>返回上一页</span>
        </button>
        <router-link to="/dashboard" class="scheme3-not-found-button scheme3-not-found-button-primary">
          <Icon name="home" size="sm" />
          <span>回到工作台</span>
        </router-link>
      </div>
      <div class="scheme3-not-found-help">
        <Icon name="chat" size="sm" />
        <span>需要帮助？联系客服 QQ 776523718</span>
      </div>
    </section>

    <footer class="scheme3-not-found-footer">&copy; {{ currentYear }} {{ siteName }} · 页面访问记录</footer>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { resolveDisplaySiteName } from '@/utils/branding'

const router = useRouter()
const appStore = useAppStore()
const siteName = computed(() => resolveDisplaySiteName(appStore.siteName))
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const currentYear = new Date().getFullYear()

function goBack(): void {
  if (window.history.length > 1) router.back()
  else void router.push('/dashboard')
}

onMounted(() => {
  void appStore.fetchPublicSettings()
})
</script>

<style scoped>
.scheme3-not-found {
  --not-found-paper: #f4f2ec;
  --not-found-card: #fbfaf6;
  --not-found-ink: #16150f;
  --not-found-muted: #6b695f;
  --not-found-line: #dad5c8;
  --not-found-green: #1e5c42;
  position: relative;
  display: flex;
  min-height: 100vh;
  flex-direction: column;
  align-items: center;
  overflow: hidden;
  background: var(--not-found-paper);
  color: var(--not-found-ink);
  padding: 1.4rem 1rem 1.2rem;
}

.scheme3-not-found::before {
  position: absolute;
  inset: 1rem;
  border: 1px solid color-mix(in srgb, var(--not-found-line) 72%, transparent);
  content: '';
  pointer-events: none;
}

.scheme3-not-found-rule {
  position: absolute;
  top: 4.8rem;
  right: 0;
  left: 0;
  height: 1px;
  background: var(--not-found-line);
  opacity: .8;
}

.scheme3-not-found-brand,
.scheme3-not-found-panel,
.scheme3-not-found-footer { position: relative; z-index: 1; width: min(100%, 48rem); }

.scheme3-not-found-brand { display: flex; align-items: center; gap: .75rem; padding: .4rem .2rem 1.4rem; }
.scheme3-not-found-mark { display: inline-flex; width: 2.8rem; height: 2.8rem; align-items: center; justify-content: center; overflow: hidden; border: 1px solid rgba(30,92,66,.35); border-radius: 8px; background: var(--not-found-green); color: #f4f2ec; font-size: .7rem; font-weight: 800; letter-spacing: .08em; box-shadow: 0 8px 18px rgba(30,92,66,.16); }
.scheme3-not-found-mark img { width: 100%; height: 100%; object-fit: contain; }
.scheme3-not-found-brand small { display: block; color: var(--not-found-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .57rem; font-weight: 800; letter-spacing: .12em; }
.scheme3-not-found-brand strong { display: block; margin-top: .18rem; font-family: Georgia, 'Times New Roman', serif; font-size: 1.1rem; font-weight: 500; }

.scheme3-not-found-panel { display: grid; grid-template-columns: minmax(9rem, .7fr) minmax(0, 1.3fr); align-items: center; gap: 1.4rem; margin: auto 0; border: 1px solid var(--not-found-line); border-radius: 8px; background: var(--not-found-card); padding: clamp(1.35rem, 4vw, 2.2rem); box-shadow: 0 18px 38px rgba(54,48,34,.08); }
.scheme3-not-found-index { color: var(--not-found-green); font-family: Georgia, 'Times New Roman', serif; font-size: clamp(5.2rem, 17vw, 9.5rem); font-weight: 500; letter-spacing: .02em; line-height: .85; opacity: .16; }
.scheme3-not-found-copy { min-width: 0; }
.scheme3-not-found-kicker { margin: 0; color: var(--not-found-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .62rem; font-weight: 800; letter-spacing: .1em; }
.scheme3-not-found-copy h1 { margin: .4rem 0 0; font-family: Georgia, 'Times New Roman', serif; font-size: clamp(1.65rem, 4vw, 2.35rem); font-weight: 500; letter-spacing: 0; }
.scheme3-not-found-copy > p:last-child { max-width: 28rem; margin: .65rem 0 0; color: var(--not-found-muted); font-size: .8rem; line-height: 1.7; }
.scheme3-not-found-actions { grid-column: 2; display: flex; flex-wrap: wrap; gap: .6rem; }
.scheme3-not-found-button { display: inline-flex; min-height: 2.55rem; align-items: center; justify-content: center; gap: .45rem; border: 1px solid var(--not-found-line); border-radius: 7px; padding: .55rem .85rem; font-size: .72rem; font-weight: 700; text-decoration: none; transition: transform 150ms ease, background-color 150ms ease, border-color 150ms ease; }
.scheme3-not-found-button:active { transform: scale(.98); }
.scheme3-not-found-button-secondary { background: transparent; color: var(--not-found-ink); }
.scheme3-not-found-button-secondary:hover { border-color: rgba(30,92,66,.42); background: #f1eee6; }
.scheme3-not-found-button-primary { border-color: var(--not-found-green); background: var(--not-found-green); color: #f4f2ec; }
.scheme3-not-found-button-primary:hover { background: #2b7655; }
.scheme3-not-found-help { grid-column: 2; display: flex; align-items: center; gap: .45rem; border-top: 1px solid var(--not-found-line); padding-top: .85rem; color: var(--not-found-muted); font-size: .68rem; }
.scheme3-not-found-footer { padding: 1.1rem .2rem .2rem; color: var(--not-found-muted); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .58rem; letter-spacing: .04em; text-align: center; }

:global(html.dark .scheme3-not-found) { --not-found-paper: #1b1b18; --not-found-card: #24231f; --not-found-ink: #f4f2ec; --not-found-muted: #aaa69a; --not-found-line: #47443a; --not-found-green: #8fc2a5; }
:global(html.dark .scheme3-not-found-mark) { border-color: rgba(143,194,165,.36); background: var(--not-found-green); color: #1b1b18; }
:global(html.dark .scheme3-not-found-panel) { box-shadow: 0 20px 44px rgba(0,0,0,.28); }
:global(html.dark .scheme3-not-found-button-secondary) { color: var(--not-found-ink); }
:global(html.dark .scheme3-not-found-button-secondary:hover) { border-color: rgba(143,194,165,.4); background: #2b2924; }
:global(html.dark .scheme3-not-found-button-primary) { border-color: var(--not-found-green); background: var(--not-found-green); color: #1b1b18; }
:global(html.dark .scheme3-not-found-button-primary:hover) { background: #a7d2b7; }

@media (max-width: 620px) {
  .scheme3-not-found { padding: 1rem .8rem; }
  .scheme3-not-found::before { inset: .65rem; }
  .scheme3-not-found-brand { padding-bottom: 1.05rem; }
  .scheme3-not-found-panel { grid-template-columns: 1fr; gap: .9rem; margin: auto 0; padding: 1.2rem; }
  .scheme3-not-found-index { font-size: 5.5rem; }
  .scheme3-not-found-actions,
  .scheme3-not-found-help { grid-column: 1; }
  .scheme3-not-found-actions { flex-direction: column; }
  .scheme3-not-found-button { width: 100%; }
}
</style>

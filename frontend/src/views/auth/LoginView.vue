<template>
  <AuthLayout>
    <div class="auth-panel">
      <header class="auth-panel-heading">
        <p class="auth-eyebrow">{{ t('auth.scheme3.loginEyebrow') }}</p>
        <h2 class="auth-panel-title">{{ t('auth.welcomeBack') }}</h2>
        <p class="auth-panel-subtitle">{{ t('auth.signInToAccount') }}</p>
      </header>

      <div v-if="errorMessage" class="auth-alert auth-alert-error" role="alert">
        <Icon name="exclamationCircle" size="sm" />
        <span>{{ errorMessage }}</span>
      </div>

      <form class="auth-form" @submit.prevent="handleLogin">
        <div class="auth-field">
          <label for="email" class="auth-label">{{ t('auth.emailLabel') }}</label>
          <div class="auth-input-shell">
            <Icon name="mail" size="sm" class="auth-input-icon" />
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="authActionDisabled"
              class="auth-input"
              :class="{ 'auth-input-invalid': errors.email }"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
          <p v-if="errors.email" class="auth-field-error">{{ errors.email }}</p>
        </div>

        <div class="auth-field">
          <label for="password" class="auth-label">{{ t('auth.passwordLabel') }}</label>
          <div class="auth-input-shell">
            <Icon name="lock" size="sm" class="auth-input-icon" />
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :disabled="authActionDisabled"
              class="auth-input auth-input-with-action"
              :class="{ 'auth-input-invalid': errors.password }"
              :placeholder="t('auth.passwordPlaceholder')"
            />
            <button
              type="button"
              class="auth-input-action"
              :disabled="authActionDisabled"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              @click="showPassword = !showPassword"
            >
              <Icon v-if="showPassword" name="eyeOff" size="sm" />
              <Icon v-else name="eye" size="sm" />
            </button>
          </div>
          <div class="auth-field-meta">
            <span></span>
            <router-link
              v-if="passwordResetEnabled && !backendModeEnabled"
              to="/forgot-password"
              class="auth-link"
            >
              {{ t('auth.forgotPassword') }}
            </router-link>
          </div>
          <p v-if="errors.password" class="auth-field-error">{{ errors.password }}</p>
        </div>

        <div v-if="captchaEnabled" class="auth-captcha">
          <TurnstileWidget
            ref="turnstileRef"
            :turnstile-enabled="turnstileEnabled"
            :turnstile-site-key="turnstileSiteKey"
            :tencent-enabled="tencentCaptchaEnabled"
            :tencent-app-id="tencentCaptchaAppId"
            :tencent-region="tencentCaptchaRegion"
            :aliyun-enabled="aliyunCaptchaEnabled"
            :aliyun-scene-id="aliyunCaptchaSceneId"
            :aliyun-prefix="aliyunCaptchaPrefix"
            :aliyun-region="aliyunCaptchaRegion"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
          <p v-if="errors.turnstile" class="auth-field-error">{{ errors.turnstile }}</p>
        </div>

        <LoginAgreementPrompt
          v-if="loginAgreementEnabled"
          :accepted="agreementAccepted"
          :documents="loginAgreementDocuments"
          :mode="loginAgreementMode"
          :updated-at="loginAgreementUpdatedAt"
          :visible="showAgreementModal"
          @accept="acceptLoginAgreement"
          @reject="rejectLoginAgreement"
          @open="showAgreementModal = true"
        />

        <button
          type="submit"
          class="auth-submit"
          :disabled="authActionDisabled || (turnstileEnabled && !turnstileToken)"
        >
          <svg v-if="isLoading" class="auth-spinner" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <Icon v-else name="login" size="sm" />
          <span>{{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}</span>
        </button>
      </form>

      <div v-if="showPasskeyLogin || showOAuthLogin" class="auth-alternatives">
        <div class="auth-divider"><span>{{ t('auth.scheme3.otherEntry') }}</span></div>

        <button
          v-if="showPasskeyLogin"
          type="button"
          class="auth-secondary"
          :disabled="authActionDisabled"
          @click="handlePasskeyLogin"
        >
          <Icon name="key" size="sm" />
          <span>{{ passkeyLoading ? t('auth.passkeySigningIn') : t('auth.passkeySignIn') }}</span>
        </button>

        <EmailOAuthButtons
          :disabled="authActionDisabled"
          :github-enabled="githubOAuthEnabled"
          :google-enabled="googleOAuthEnabled"
          :show-divider="false"
          @start="handleOAuthStart"
        />
        <LinuxDoOAuthSection
          v-if="linuxdoOAuthEnabled"
          :disabled="authActionDisabled"
          :show-divider="false"
          @start="handleOAuthStart"
        />
        <DingTalkOAuthSection
          v-if="dingtalkOAuthEnabled"
          :disabled="authActionDisabled"
          :show-divider="false"
          @start="handleOAuthStart"
        />
        <WechatOAuthSection
          v-if="wechatOAuthEnabled"
          :disabled="authActionDisabled"
          :show-divider="false"
          @start="handleOAuthStart"
        />
        <OidcOAuthSection
          v-if="oidcOAuthEnabled"
          :disabled="authActionDisabled"
          :provider-name="oidcOAuthProviderName"
          :show-divider="false"
          @start="handleOAuthStart"
        />
      </div>
    </div>

    <template v-if="!backendModeEnabled" #footer>
      <p class="auth-footer-copy">
        {{ t('auth.dontHaveAccount') }}
        <router-link to="/register" class="auth-link">{{ t('auth.signUp') }}</router-link>
      </p>
    </template>
  </AuthLayout>

  <TotpLoginModal
    v-if="show2FAModal"
    ref="totpModalRef"
    :temp-token="totpTempToken"
    :user-email-masked="totpUserEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import DingTalkOAuthSection from '@/components/auth/DingTalkOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import EmailOAuthButtons from '@/components/auth/EmailOAuthButtons.vue'
import LoginAgreementPrompt from '@/components/auth/LoginAgreementPrompt.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/CaptchaChallenge.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  buildOAuthLoginStartURL,
  getPublicSettings,
  isTotp2FARequired,
  isWeChatWebOAuthEnabled,
  startOAuthLogin,
  type OAuthLoginStart
} from '@/api/auth'
import type {
  ActionCaptchaRequestProof,
  LoginAgreementDocument,
  TotpLoginResponse
} from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { clearAllAffiliateReferralCodes } from '@/utils/oauthAffiliate'

const { t } = useI18n()
const LOGIN_AGREEMENT_STORAGE_KEY = 'sub2api_login_agreement_consent'

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const passkeyLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const publicSettingsLoaded = ref<boolean>(false)

// Public settings
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const tencentCaptchaEnabled = ref<boolean>(false)
const tencentCaptchaAppId = ref<string>('')
const tencentCaptchaRegion = ref<string>('cn')
const aliyunCaptchaEnabled = ref<boolean>(false)
const aliyunCaptchaSceneId = ref<string>('')
const aliyunCaptchaPrefix = ref<string>('')
const aliyunCaptchaRegion = ref<string>('cn')
const linuxdoOAuthEnabled = ref<boolean>(false)
const dingtalkOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const backendModeEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const githubOAuthEnabled = ref<boolean>(false)
const googleOAuthEnabled = ref<boolean>(false)
const passwordResetEnabled = ref<boolean>(false)
const passkeyEnabled = ref<boolean>(false)
const loginAgreementEnabled = ref<boolean>(false)
const loginAgreementMode = ref<'modal' | 'checkbox' | string>('modal')
const loginAgreementUpdatedAt = ref<string>('')
const loginAgreementRevision = ref<string>('')
const loginAgreementDocuments = ref<LoginAgreementDocument[]>([])
const agreementAccepted = ref<boolean>(false)
const showAgreementModal = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')
const tencentCaptchaRandstr = ref<string>('')
const aliyunCaptchaReady = computed(
  () =>
    aliyunCaptchaEnabled.value &&
    Boolean(aliyunCaptchaSceneId.value) &&
    Boolean(aliyunCaptchaPrefix.value)
)
// 动作触发式验证码（腾讯/阿里云）：提交、OAuth 启动、passkey 时弹窗验证
const actionCaptchaEnabled = computed(
  () =>
    (tencentCaptchaEnabled.value && Boolean(tencentCaptchaAppId.value)) ||
    aliyunCaptchaReady.value
)
const captchaEnabled = computed(
  () =>
    (turnstileEnabled.value && Boolean(turnstileSiteKey.value)) || actionCaptchaEnabled.value
)

// 2FA state
const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

const validationToastMessage = computed(
  () => errors.email || errors.password || errors.turnstile || ''
)

const agreementGateActive = computed(
  () => loginAgreementEnabled.value && !agreementAccepted.value
)

const authActionDisabled = computed(
  () => isLoading.value || passkeyLoading.value || !publicSettingsLoaded.value || agreementGateActive.value
)

const showPasskeyLogin = computed(
  () => passkeyEnabled.value && typeof window.PublicKeyCredential !== 'undefined'
)

const showOAuthLogin = computed(
  () =>
    !backendModeEnabled.value &&
    (linuxdoOAuthEnabled.value ||
      dingtalkOAuthEnabled.value ||
      wechatOAuthEnabled.value ||
      oidcOAuthEnabled.value ||
      githubOAuthEnabled.value ||
      googleOAuthEnabled.value)
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// ==================== Lifecycle ====================

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    tencentCaptchaEnabled.value = settings.tencent_captcha_enabled === true
    tencentCaptchaAppId.value = settings.tencent_captcha_app_id || ''
    tencentCaptchaRegion.value = settings.tencent_captcha_region || 'cn'
    aliyunCaptchaEnabled.value = settings.aliyun_captcha_enabled === true
    aliyunCaptchaSceneId.value = settings.aliyun_captcha_scene_id || ''
    aliyunCaptchaPrefix.value = settings.aliyun_captcha_prefix || ''
    aliyunCaptchaRegion.value = settings.aliyun_captcha_region || 'cn'
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
    wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
    backendModeEnabled.value = settings.backend_mode_enabled
    oidcOAuthEnabled.value = settings.oidc_oauth_enabled
    oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    githubOAuthEnabled.value = settings.github_oauth_enabled
    googleOAuthEnabled.value = settings.google_oauth_enabled
    backendModeEnabled.value = settings.backend_mode_enabled
    passwordResetEnabled.value = settings.password_reset_enabled
    passkeyEnabled.value = settings.passkey_enabled === true
    applyLoginAgreementSettings(settings)
  } catch (error) {
    console.error('Failed to load public settings:', error)
    loginAgreementEnabled.value = false
    agreementAccepted.value = true
  } finally {
    publicSettingsLoaded.value = true
  }
})

// ==================== Login Agreement ====================

function applyLoginAgreementSettings(settings: {
  login_agreement_enabled?: boolean
  login_agreement_mode?: string
  login_agreement_updated_at?: string
  login_agreement_revision?: string
  login_agreement_documents?: LoginAgreementDocument[]
}): void {
  const documents = Array.isArray(settings.login_agreement_documents)
    ? settings.login_agreement_documents.filter((doc) => doc.title?.trim())
    : []
  loginAgreementDocuments.value = documents
  loginAgreementEnabled.value = settings.login_agreement_enabled === true && documents.length > 0
  loginAgreementMode.value = settings.login_agreement_mode === 'checkbox' ? 'checkbox' : 'modal'
  loginAgreementUpdatedAt.value = settings.login_agreement_updated_at || ''
  loginAgreementRevision.value =
    settings.login_agreement_revision ||
    `${loginAgreementUpdatedAt.value}:${documents.map((doc) => `${doc.id}:${doc.title}`).join('|')}`

  agreementAccepted.value = !loginAgreementEnabled.value || hasAcceptedLoginAgreement(loginAgreementRevision.value)
  showAgreementModal.value =
    loginAgreementEnabled.value && !agreementAccepted.value && loginAgreementMode.value !== 'checkbox'
}

function hasAcceptedLoginAgreement(revision: string): boolean {
  if (!revision) {
    return false
  }
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return false
    }
    const parsed = JSON.parse(raw) as { revision?: string }
    return parsed.revision === revision
  } catch {
    return false
  }
}

function acceptLoginAgreement(): void {
  if (loginAgreementRevision.value) {
    localStorage.setItem(
      LOGIN_AGREEMENT_STORAGE_KEY,
      JSON.stringify({
        revision: loginAgreementRevision.value,
        accepted_at: new Date().toISOString()
      })
    )
  }
  agreementAccepted.value = true
  showAgreementModal.value = false
}

function rejectLoginAgreement(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
  agreementAccepted.value = false
  showAgreementModal.value = false
  appStore.showWarning(t('legal.loginAgreementPrompt.loginRejectedWarning'))
}

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string, randstr = ''): void {
  turnstileToken.value = token
  tencentCaptchaRandstr.value = randstr
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  tencentCaptchaRandstr.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  tencentCaptchaRandstr.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

function resetCaptchaProof(): void {
  turnstileRef.value?.reset()
  turnstileToken.value = ''
  tencentCaptchaRandstr.value = ''
  errors.turnstile = ''
}

async function acquireActionProof(): Promise<boolean> {
  if (!actionCaptchaEnabled.value) return true

  const proof = await turnstileRef.value?.verifyAction()
  if (!proof) return false

  turnstileToken.value = proof.token
  tencentCaptchaRandstr.value = proof.randstr
  return true
}

// ==================== Validation ====================

function validateForm(): boolean {
  // Reset errors
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (agreementGateActive.value) {
    appStore.showWarning(t('legal.loginAgreementPrompt.loginRequiredWarning'))
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return false
  }

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Turnstile validation
  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleLogin(): Promise<void> {
  // Clear previous error
  errorMessage.value = ''

  // Validate form
  if (!validateForm()) {
    return
  }

  if (!(await acquireActionProof())) {
    return
  }

  isLoading.value = true

  try {
    // Call auth store login（阿里云 captchaVerifyParam 复用 turnstile_token 字段）
    const response = await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token:
        turnstileEnabled.value || aliyunCaptchaEnabled.value ? turnstileToken.value : undefined,
      tencent_captcha_ticket: tencentCaptchaEnabled.value ? turnstileToken.value : undefined,
      tencent_captcha_randstr: tencentCaptchaEnabled.value
        ? tencentCaptchaRandstr.value
        : undefined
    })

    // Check if 2FA is required
    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    // Show success toast
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    errorMessage.value = extractI18nErrorMessage(error, t, 'auth.errors', t('auth.loginFailed'))

    // Also show error toast
    appStore.showError(errorMessage.value)
  } finally {
    if (captchaEnabled.value) {
      resetCaptchaProof()
    }
    isLoading.value = false
  }
}

async function handlePasskeyLogin(): Promise<void> {
  if (agreementGateActive.value) {
    appStore.showWarning(t('legal.loginAgreementPrompt.loginRequiredWarning'))
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return
  }

  passkeyLoading.value = true
  try {
    let proof: ActionCaptchaRequestProof | undefined
    if (actionCaptchaEnabled.value) {
      const result = await turnstileRef.value?.verifyAction()
      if (!result) return
      proof = tencentCaptchaEnabled.value
        ? {
            tencent_captcha_ticket: result.token,
            tencent_captcha_randstr: result.randstr
          }
        : { turnstile_token: result.token }
    }

    await authStore.loginWithPasskey(proof)
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const fallback = error instanceof DOMException && error.name === 'NotAllowedError'
      ? t('auth.passkeyCancelled')
      : t('auth.passkeyFailed')
    errorMessage.value = extractI18nErrorMessage(error, t, 'auth.errors', fallback)
    appStore.showError(errorMessage.value)
  } finally {
    if (actionCaptchaEnabled.value) {
      resetCaptchaProof()
    }
    passkeyLoading.value = false
  }
}

async function handleOAuthStart(request: OAuthLoginStart): Promise<void> {
  if (authActionDisabled.value) return

  if (!actionCaptchaEnabled.value) {
    window.location.href = buildOAuthLoginStartURL(request)
    return
  }

  isLoading.value = true
  try {
    const proof = await turnstileRef.value?.verifyAction()
    if (!proof) return

    const result = await startOAuthLogin(
      request,
      tencentCaptchaEnabled.value
        ? {
            tencent_captcha_ticket: proof.token,
            tencent_captcha_randstr: proof.randstr
          }
        : { turnstile_token: proof.token }
    )
    window.location.href = result.authorize_url
  } catch (error: unknown) {
    errorMessage.value = extractI18nErrorMessage(
      error,
      t,
      'auth.errors',
      t('auth.turnstileFailed')
    )
    appStore.showError(errorMessage.value)
  } finally {
    resetCaptchaProof()
    isLoading.value = false
  }
}

// ==================== 2FA Handlers ====================

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)

    // Close modal and show success
    show2FAModal.value = false
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('profile.totp.loginFailed')

    if (totpModalRef.value) {
      totpModalRef.value.setError(message)
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  show2FAModal.value = false
  totpTempToken.value = ''
  totpUserEmailMasked.value = ''
}
</script>

<style scoped>
@keyframes auth-input-breathe {
  0%, 100% { box-shadow: 0 0 0 2px color-mix(in srgb, var(--auth-green) 12%, transparent); }
  50% { box-shadow: 0 0 0 4px color-mix(in srgb, var(--auth-green) 20%, transparent); }
}

.auth-panel {
  --auth-line: #dad5c8;
  --auth-card: #fbfaf6;
  --auth-ink: #16150f;
  --auth-muted: #6b695f;
  --auth-green: #1e5c42;
  --auth-red: #a63d32;
}

:global(html.dark .auth-panel) {
  --auth-line: #47443a;
  --auth-card: #24231f;
  --auth-ink: #f4f2ec;
  --auth-muted: #aaa69a;
  --auth-green: #8fc2a5;
  --auth-red: #e28b80;
}

.auth-panel-heading {
  border-bottom: 1px solid var(--auth-line);
  margin-bottom: 1.35rem;
  padding-bottom: 1.1rem;
}

.auth-eyebrow,
.auth-label,
.auth-divider span {
  color: var(--auth-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.68rem;
  letter-spacing: 0.12em;
  line-height: 1.4;
  text-transform: uppercase;
}

.auth-panel-title {
  margin: 0.5rem 0 0;
  color: var(--auth-ink);
  font-family: Georgia, "Times New Roman", serif;
  font-size: clamp(1.65rem, 7vw, 2.2rem);
  font-weight: 400;
  letter-spacing: 0;
  line-height: 1.05;
}

.auth-panel-subtitle {
  margin: 0.5rem 0 0;
  color: var(--auth-muted);
  font-size: 0.84rem;
  line-height: 1.55;
}

.auth-form,
.auth-alternatives {
  display: grid;
  gap: 1rem;
}

.auth-field {
  min-width: 0;
}

.auth-label {
  display: block;
  margin-bottom: 0.4rem;
  color: var(--auth-ink);
  font-size: 0.7rem;
  letter-spacing: 0.08em;
}

.auth-input-shell {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 2.85rem;
  border: 1px solid var(--auth-line);
  background: var(--auth-card);
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.auth-input-shell:focus-within {
  border-color: var(--auth-green);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--auth-green) 16%, transparent);
  animation: auth-input-breathe 1.8s ease-in-out infinite;
}

.auth-input-icon {
  position: absolute;
  left: 0.75rem;
  color: var(--auth-muted);
  pointer-events: none;
}

.auth-input {
  width: 100%;
  min-width: 0;
  min-height: 2.8rem;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--auth-ink);
  padding: 0.65rem 0.8rem 0.65rem 2.55rem;
  font-size: 0.9rem;
}

.auth-input::placeholder {
  color: color-mix(in srgb, var(--auth-muted) 72%, transparent);
}

.auth-input-with-action {
  padding-right: 2.8rem;
}

.auth-input-invalid {
  color: var(--auth-red);
}

.auth-input-action {
  position: absolute;
  right: 0.25rem;
  display: inline-flex;
  width: 2.25rem;
  height: 2.25rem;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: var(--auth-muted);
}

.auth-input-action:hover:not(:disabled) {
  color: var(--auth-green);
}

.auth-input-action:disabled,
.auth-submit:disabled,
.auth-secondary:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.auth-field-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 1.2rem;
  margin-top: 0.25rem;
}

.auth-field-error {
  margin: 0.35rem 0 0;
  color: var(--auth-red);
  font-size: 0.75rem;
  line-height: 1.4;
}

.auth-captcha {
  display: grid;
  justify-items: center;
  min-width: 0;
  border-top: 1px solid var(--auth-line);
  border-bottom: 1px solid var(--auth-line);
  padding: 0.85rem 0;
}

.auth-alert {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  margin-bottom: 1rem;
  border: 1px solid;
  padding: 0.75rem 0.85rem;
  font-size: 0.8rem;
  line-height: 1.45;
}

.auth-alert-error {
  border-color: color-mix(in srgb, var(--auth-red) 35%, var(--auth-line));
  background: #fbefec;
  color: var(--auth-red);
}

.auth-submit,
.auth-secondary {
  position: relative;
  overflow: hidden;
  display: inline-flex;
  min-height: 2.85rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  border: 1px solid var(--auth-green);
  padding: 0.7rem 1rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.76rem;
  letter-spacing: 0.06em;
  transition: background 160ms ease, color 160ms ease, border-color 160ms ease;
}

.auth-submit::after,
.auth-secondary::after {
  position: absolute;
  top: -35%;
  left: 0;
  width: 1.5rem;
  height: 170%;
  background: color-mix(in srgb, white 24%, transparent);
  content: "";
  opacity: 0;
  pointer-events: none;
  transform: translateX(-150%) skewX(-18deg);
}

.auth-submit:hover:not(:disabled)::after,
.auth-secondary:hover:not(:disabled)::after {
  opacity: 1;
  animation: auth-button-sheen 680ms ease-out both;
}

.auth-submit:active:not(:disabled),
.auth-secondary:active:not(:disabled),
.auth-input-action:active:not(:disabled) {
  transform: translateY(1px) scale(0.98);
}

@keyframes auth-button-sheen {
  from { transform: translateX(-150%) skewX(-18deg); }
  to { transform: translateX(900%) skewX(-18deg); }
}

.auth-submit {
  background: var(--auth-green);
  color: #fbfaf6;
}

.auth-submit:hover:not(:disabled) {
  background: #174a35;
}

.auth-secondary {
  border-color: var(--auth-line);
  background: var(--auth-card);
  color: var(--auth-ink);
}

.auth-secondary:hover:not(:disabled) {
  border-color: var(--auth-green);
  color: var(--auth-green);
}

.auth-spinner {
  width: 1rem;
  height: 1rem;
  animation: auth-spin 0.8s linear infinite;
}

@keyframes auth-spin {
  to { transform: rotate(360deg); }
}

.auth-alternatives {
  margin-top: 1.3rem;
  padding-top: 1.2rem;
  border-top: 1px solid var(--auth-line);
}

.auth-divider {
  display: flex;
  align-items: center;
  gap: 0.7rem;
}

.auth-divider::before,
.auth-divider::after {
  height: 1px;
  flex: 1;
  background: var(--auth-line);
  content: "";
}

.auth-link {
  color: var(--auth-green);
  font-weight: 600;
  text-decoration: underline;
  text-decoration-color: color-mix(in srgb, var(--auth-green) 35%, transparent);
  text-underline-offset: 0.2em;
}

.auth-link:hover {
  color: #174a35;
}

:global(html.dark .auth-panel .auth-link) {
  color: var(--auth-green);
}

:global(html.dark .auth-panel .auth-link:hover) {
  color: #b4dfc5;
}

.auth-footer-copy {
  margin: 0;
  color: var(--auth-muted);
}

.auth-panel :deep(.btn) {
  min-height: 2.8rem;
  border-radius: 0;
  border: 1px solid var(--auth-line);
  background: var(--auth-card);
  color: var(--auth-ink);
  box-shadow: none;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.76rem;
  letter-spacing: 0.04em;
}

.auth-panel :deep(.btn:hover:not(:disabled)) {
  border-color: var(--auth-green);
  background: var(--auth-card);
  color: var(--auth-green);
}

.auth-panel :deep(.auth-agreement) {
  font-size: 0.78rem;
}

.auth-panel :deep(.rounded-xl),
.auth-panel :deep(.rounded-2xl),
.auth-panel :deep(.rounded-lg) {
  border-radius: 0;
}

.auth-panel :deep(.bg-primary-600) {
  background: var(--auth-green);
}

@media (max-width: 480px) {
  .auth-panel-title {
    font-size: 1.75rem;
  }

  .auth-panel :deep(.grid-cols-2) {
    grid-template-columns: 1fr;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>

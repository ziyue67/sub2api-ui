import { sanitizeUrl } from '@/utils/url'

export const DEFAULT_DISPLAY_SITE_NAME = 'Shour or ToKen'

export function resolveDisplaySiteName(siteName?: string | null): string {
  const normalized = siteName?.trim()
  return !normalized || normalized === 'Sub2API' ? DEFAULT_DISPLAY_SITE_NAME : normalized
}

export function updateFavicon(logoUrl: string): void {
  const sanitizedLogoUrl = sanitizeUrl(logoUrl, {
    allowRelative: true,
    allowDataUrl: true,
  })
  if (!sanitizedLogoUrl) {
    return
  }

  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }

  link.type = sanitizedLogoUrl.endsWith('.svg') ? 'image/svg+xml' : 'image/x-icon'
  link.href = sanitizedLogoUrl
}

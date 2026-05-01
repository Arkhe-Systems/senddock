export const CHECKOUT_PRO = 'https://senddock.lemonsqueezy.com/checkout/buy/08756076-6890-4a78-ad53-2bcd15032360'
export const CHECKOUT_TEAM = 'https://senddock.lemonsqueezy.com/checkout/buy/2f43c2fa-faa7-4ec9-b8cd-1faf44932219'

export const LAUNCH_DISCOUNT_CODE = 'LAUNCH3FREE'
export const LAUNCH_DISCOUNT_EXPIRES = '2026-05-10'

export type Tier = 'pro' | 'team'

export function checkoutUrl(tier: Tier, options?: { discount?: string }): string {
    const base = tier === 'team' ? CHECKOUT_TEAM : CHECKOUT_PRO
    const params = new URLSearchParams()
    if (options?.discount) params.set('discount', options.discount)
    const qs = params.toString()
    return qs ? `${base}?${qs}` : base
}

export function isLaunchPromoActive(now: Date = new Date()): boolean {
    return now < new Date(`${LAUNCH_DISCOUNT_EXPIRES}T23:59:59Z`)
}

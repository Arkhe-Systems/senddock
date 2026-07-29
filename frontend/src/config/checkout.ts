export const CHECKOUT_PRO = 'https://senddock.lemonsqueezy.com/checkout/buy/b5b0d1f3-ca5c-4b47-9004-a044f693441d'
export const CHECKOUT_TEAM = 'https://senddock.lemonsqueezy.com/checkout/buy/c265cb7e-f0a4-4b19-8ac7-dda9938d030f'

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

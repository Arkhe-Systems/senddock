import { defineStore } from 'pinia'
import { api, getApiBase } from '@/api/client'

export interface Bucket { bucket: string }
export interface OpenBucket extends Bucket { opens: number }
export interface ClickBucket extends Bucket { clicks: number }
export interface StatusBucket { status: string; count: number }
export interface TemplateStat { template_id: string; name: string; sends: number }
export interface LinkStat { url: string; clicks: number }
export interface BroadcastInFlight {
    id: string; subject: string; total: number; sent: number
    failed: number; suppressed: number; remaining: number; started_at: string
}
export interface PeriodMetrics {
    total_sent: number; total_failed: number; total_bounced: number
    total_opened: number; total_clicked: number; total_complained: number
    acceptance_pct: number; bounce_rate_pct: number; complaint_rate_pct: number
    open_rate_pct: number; click_rate_pct: number
}
export interface Overview {
    from: string; to: string; granularity: string; range_days: number; segment_id?: string
    total_sent: number; total_failed: number; total_bounced: number
    total_opened: number; total_clicked: number; total_complained: number
    acceptance_pct: number; bounce_rate_pct: number; complaint_rate_pct: number
    open_rate_pct: number; click_rate_pct: number; click_to_open_pct: number
    opens_series: OpenBucket[]; clicks_series: ClickBucket[]
    top_templates: TemplateStat[]; top_clicked_links: LinkStat[]
    active_subscribers: number; sends_by_status: StatusBucket[]
    previous: PeriodMetrics; broadcasts_in_flight: BroadcastInFlight[]
}

export interface CampaignStat {
    broadcast_id: string; subject: string; status: string
    started_at: string; finished_at?: string
    total_recipients: number; sent: number; failed: number; bounced: number
    opened: number; clicked: number
    acceptance_pct: number; bounce_rate_pct: number
    open_rate_pct: number; click_rate_pct: number; click_to_open_pct: number
}
export interface CampaignDetail extends CampaignStat { top_clicked_links: LinkStat[] }

export interface AudienceBucket {
    bucket: string; added: number; unsubscribed: number
    net_growth: number; cumulative_net: number
}
export interface Audience {
    from: string; to: string; granularity: string
    active_total: number; unsubscribed_total: number
    added_in_range: number; unsubscribed_in_range: number
    series: AudienceBucket[]
}

export interface Breakdown { label: string; count: number }
export interface Funnel { sent: number; opened: number; clicked: number }
export interface EngagementBucket { bucket: string; opens: number; clicks: number }
export interface HeatCell { weekday: number; hour: number; count: number }
export interface Engagement {
    devices: Breakdown[]; clients: Breakdown[]
    funnel: Funnel; series: EngagementBucket[]; heatmap: HeatCell[]
}

export type CheckStatus = 'pass' | 'warn' | 'fail'
export interface DomainCheck {
    name: string; status: CheckStatus; detail: string; value?: string; fix?: string
}
export interface DomainHealth { domain: string; checks: DomainCheck[] }

export interface ProviderStats {
    provider: string
    sent: number; failed: number; bounced: number; opened: number; clicked: number; complained: number
    hard_bounces: number; soft_bounces: number
    acceptance_pct: number; bounce_rate_pct: number; complaint_rate_pct: number
    open_rate_pct: number; click_rate_pct: number
}
export interface ProviderBreakdown {
    from: string; to: string; total_bounced: number; providers: ProviderStats[]
}

function windowParams(from: string, to: string, segmentID?: string): string {
    const p = new URLSearchParams({ from, to })
    if (segmentID) p.set('segment_id', segmentID)
    return p.toString()
}

export const useAnalyticsStore = defineStore('analytics', () => {
    const base = (projectID: string) => `/projects/${projectID}/analytics`

    function overview(projectID: string, from: string, to: string, segmentID?: string) {
        return api<Overview>(`${base(projectID)}/overview?${windowParams(from, to, segmentID)}`)
    }
    function campaigns(projectID: string, from: string, to: string) {
        return api<{ campaigns: CampaignStat[] }>(`${base(projectID)}/campaigns?${windowParams(from, to)}`)
    }
    function campaign(projectID: string, broadcastID: string) {
        return api<CampaignDetail>(`${base(projectID)}/campaigns/${broadcastID}`)
    }
    function audience(projectID: string, from: string, to: string) {
        return api<Audience>(`${base(projectID)}/audience?${windowParams(from, to)}`)
    }
    function engagement(projectID: string, from: string, to: string) {
        return api<Engagement>(`${base(projectID)}/engagement?${windowParams(from, to)}`)
    }
    // Export is a file download, so it goes straight to the URL rather than through api().
    function exportUrl(projectID: string): string {
        return `${getApiBase()}${base(projectID)}/export`
    }

    // Deliverability lives under its own (Pro-gated) base, not /analytics.
    const deliverabilityBase = (projectID: string) => `/projects/${projectID}/deliverability`
    function domainHealth(projectID: string) {
        return api<DomainHealth>(`${deliverabilityBase(projectID)}/domain-health`)
    }
    function providers(projectID: string, from: string, to: string) {
        return api<ProviderBreakdown>(`${deliverabilityBase(projectID)}/providers?${windowParams(from, to)}`)
    }

    return { overview, campaigns, campaign, audience, engagement, exportUrl, domainHealth, providers }
})

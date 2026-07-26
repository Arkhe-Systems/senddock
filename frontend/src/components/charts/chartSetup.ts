import {
    Chart,
    LineElement,
    PointElement,
    BarElement,
    ArcElement,
    LineController,
    BarController,
    DoughnutController,
    CategoryScale,
    LinearScale,
    Filler,
    Tooltip,
    Legend,
} from 'chart.js'

// Registered once, tree-shaken to only the controllers the dashboard uses.
Chart.register(
    LineElement, PointElement, BarElement, ArcElement,
    LineController, BarController, DoughnutController,
    CategoryScale, LinearScale,
    Filler, Tooltip, Legend,
)

// Palette aligned with the app's zinc + accent scheme, tuned for the dark UI.
export const chartColors = {
    grid: 'rgba(63, 63, 70, 0.4)',      // zinc-700 @ 40%
    tick: '#a1a1aa',                    // zinc-400
    indigo: '#818cf8',                  // indigo-400
    indigoFill: 'rgba(129, 140, 248, 0.15)',
    violet: '#c084fc',                  // purple-400
    emerald: '#34d399',                 // emerald-400
    amber: '#fbbf24',                   // amber-400
    red: '#f87171',                     // red-400
    zinc: '#71717a',                    // zinc-500
}

// The categorical palette for donuts / multi-series, in order.
export const seriesPalette = [
    chartColors.indigo, chartColors.emerald, chartColors.amber,
    chartColors.violet, chartColors.red, chartColors.zinc,
]

// Shared options every chart starts from: dark tooltips, muted grid, no title.
export function baseOptions() {
    return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { labels: { color: chartColors.tick, boxWidth: 12, boxHeight: 12 } },
            tooltip: {
                backgroundColor: '#18181b',   // zinc-900
                borderColor: '#3f3f46',       // zinc-700
                borderWidth: 1,
                titleColor: '#fafafa',
                bodyColor: '#d4d4d8',
                padding: 10,
            },
        },
        scales: {
            x: { grid: { color: chartColors.grid }, ticks: { color: chartColors.tick } },
            y: { grid: { color: chartColors.grid }, ticks: { color: chartColors.tick }, beginAtZero: true },
        },
    }
}

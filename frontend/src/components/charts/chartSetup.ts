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

Chart.register(
    LineElement, PointElement, BarElement, ArcElement,
    LineController, BarController, DoughnutController,
    CategoryScale, LinearScale,
    Filler, Tooltip, Legend,
)

export const chartColors = {
    grid: 'rgba(63, 63, 70, 0.4)',
    tick: '#a1a1aa',
    indigo: '#818cf8',
    indigoFill: 'rgba(129, 140, 248, 0.15)',
    violet: '#c084fc',
    emerald: '#34d399',
    amber: '#fbbf24',
    red: '#f87171',
    zinc: '#71717a',
}

export const seriesPalette = [
    chartColors.indigo, chartColors.emerald, chartColors.amber,
    chartColors.violet, chartColors.red, chartColors.zinc,
]

export function baseOptions() {
    return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { labels: { color: chartColors.tick, boxWidth: 12, boxHeight: 12 } },
            tooltip: {
                backgroundColor: '#18181b',
                borderColor: '#3f3f46',
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

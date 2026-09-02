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
    grid: 'rgba(57, 58, 64, 0.45)',
    tick: '#b2b3bd',
    indigo: '#4e91ff',
    indigoFill: 'rgba(78, 145, 255, 0.16)',
    violet: '#d6b1ff',
    emerald: '#4ec695',
    amber: '#ffd099',
    red: '#ef5853',
    zinc: '#eceef9',
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
                backgroundColor: '#19191b',
                borderColor: '#46484f',
                borderWidth: 1,
                titleColor: '#eeeef0',
                bodyColor: '#b2b3bd',
                padding: 10,
            },
        },
        scales: {
            x: { grid: { color: chartColors.grid }, ticks: { color: chartColors.tick } },
            y: { grid: { color: chartColors.grid }, ticks: { color: chartColors.tick }, beginAtZero: true },
        },
    }
}

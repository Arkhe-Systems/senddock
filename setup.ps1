param(
    [switch]$Reset,
    [switch]$Logs
)

$ComposeFile = "docker-compose.prod.yml"

function Test-DockerReady {
    docker info 2>&1 | Out-Null
    return $LASTEXITCODE -eq 0
}

function Get-SenddockVolumes {
    $project = (Split-Path -Leaf (Get-Location)).ToLower() -replace '[^a-z0-9]', ''
    $output = docker volume ls --format "{{.Name}}" 2>&1
    if ($LASTEXITCODE -ne 0) { return @() }
    return $output | Where-Object { $_ -like "${project}_*" }
}

if (-not (Test-DockerReady)) {
    Write-Host "Docker is not running or not reachable." -ForegroundColor Red
    Write-Host "Start Docker Desktop and wait until the whale icon shows 'Docker Desktop is running', then re-run this script."
    exit 1
}

if ($Reset) {
    Write-Host "Reset requested. Tearing down containers and volumes..." -ForegroundColor Yellow
    docker compose -f $ComposeFile down -v 2>&1 | Out-Null
    if (Test-Path .env) { Remove-Item .env -Force }
    Write-Host "Reset complete. Continuing with a fresh install." -ForegroundColor Green
    Write-Host ""
}

$EnvExists = Test-Path .env
$Volumes = Get-SenddockVolumes
$VolumesExist = $Volumes -and ($Volumes.Count -gt 0)

if ($EnvExists -and -not $VolumesExist) {
    Write-Host "Found a .env file but no Postgres volume." -ForegroundColor Yellow
    Write-Host "This usually means a previous install was wiped at the Docker level but the .env was left behind."
    Write-Host "Continuing will reuse the existing .env, including its Postgres password."
    Write-Host "If that password no longer matches anything, the database will fail to start."
    Write-Host ""
    $confirm = Read-Host "Continue with existing .env? [y/N]"
    if ($confirm -notmatch '^[yY]') {
        Write-Host "Aborted. Re-run with -Reset to wipe everything and start clean." -ForegroundColor Cyan
        exit 1
    }
}

if (-not $EnvExists) {
    Write-Host "Fresh install detected. Generating secrets..." -ForegroundColor Cyan
    $JWT_SECRET = -join ((1..64) | ForEach-Object { '{0:x}' -f (Get-Random -Max 16) })
    $POSTGRES_PASSWORD = -join ((1..32) | ForEach-Object { '{0:x}' -f (Get-Random -Max 16) })
    @"
JWT_SECRET=$JWT_SECRET
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
FRONTEND_URL=http://localhost:8080
PUBLIC_URL=http://localhost:8080
DEPLOYMENT_MODE=self-hosted
"@ | Set-Content .env -NoNewline
    Write-Host ".env created." -ForegroundColor Green
    $Mode = "install"
} else {
    Write-Host "Existing install detected. Keeping current .env." -ForegroundColor Cyan
    Write-Host "Run with -Reset to wipe data and start fresh." -ForegroundColor Gray
    $Mode = "update"
}

Write-Host ""
Write-Host "Building image (this picks up any code changes)..."
docker compose -f $ComposeFile build --pull
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed. See output above for details." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Starting services..."
docker compose -f $ComposeFile up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "docker compose up failed. Check 'docker compose logs' for details." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host -NoNewline "Waiting for SendDock to be ready"
$ready = $false
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 2
    Write-Host -NoNewline "."
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            $ready = $true
            break
        }
    } catch {
    }
}
Write-Host ""

if (-not $ready) {
    Write-Host ""
    Write-Host "SendDock did not become ready within 60 seconds." -ForegroundColor Red
    Write-Host "Check what the app container is doing:"
    Write-Host "  docker compose -f $ComposeFile ps"
    Write-Host "  docker compose -f $ComposeFile logs app --tail 100"
    Write-Host ""
    Write-Host "Common causes are documented at https://docs.senddock.dev/self-hosting/troubleshooting"
    exit 1
}

Write-Host ""
if ($Mode -eq "install") {
    Write-Host "SendDock is running at http://localhost:8080" -ForegroundColor Green
    Write-Host "Open it in your browser to create your admin account."
} else {
    Write-Host "SendDock updated and running at http://localhost:8080" -ForegroundColor Green
}

if ($Logs) {
    Write-Host ""
    Write-Host "Following app logs (Ctrl+C to stop)..."
    docker compose -f $ComposeFile logs -f app
}

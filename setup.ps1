if (Test-Path .env) {
    Write-Host "Error: .env already exists. Delete it first if you want to re-run setup." -ForegroundColor Red
    exit 1
}

$JWT_SECRET = -join ((1..64) | ForEach-Object { '{0:x}' -f (Get-Random -Max 16) })
$POSTGRES_PASSWORD = -join ((1..32) | ForEach-Object { '{0:x}' -f (Get-Random -Max 16) })

@"
JWT_SECRET=$JWT_SECRET
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
FRONTEND_URL=http://localhost:8080
DEPLOYMENT_MODE=self-hosted
"@ | Set-Content .env -NoNewline

Write-Host "SendDock setup complete." -ForegroundColor Green
Write-Host ""
Write-Host "Starting services..."

docker compose -f docker-compose.prod.yml up -d

Write-Host ""
Write-Host "SendDock is running at http://localhost:8080" -ForegroundColor Green
Write-Host "Open it in your browser to create your admin account."

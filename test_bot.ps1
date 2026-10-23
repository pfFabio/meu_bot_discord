# Script de Diagnóstico e Verificação de Saúde do Kenobi Bot
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host " 🛡️  Iniciando Verificação de Saúde do Kenobi D&D Bot..." -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan

# 1. Executa a suíte de diagnósticos via Go / bot.exe
if (Test-Path ".\bot.exe") {
    .\bot.exe --doctor
} else {
    Write-Host "Compilando bot para diagnóstico..." -ForegroundColor Yellow
    go run .\cmd\bot --doctor
}

# 2. Executa a suíte de testes unitários automatizados
Write-Host "`n🧪 Executando Testes Automatizados Unitários (Go Test)..." -ForegroundColor Cyan
go test -v ./internal/health ./internal/audio ./internal/dashboard

Write-Host "`n✅ Verificação concluída!" -ForegroundColor Green

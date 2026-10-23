@echo off
setlocal
cd /d "%~dp0"
title Kenobi Bot - Painel de Controle

echo ======================================================================
echo   Kenobi Bot - Painel de Controle do Mestre
echo ======================================================================
echo.

REM 1. Verifica se a porta 8080 ou 8081 ja esta respondendo
netstat -ano | findstr :8080 >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] O Painel ja esta em execucao na porta 8080!
    echo [INFO] Abrindo janela dedicada do aplicativo...
    goto open_window_8080
)

netstat -ano | findstr :8081 >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] O Painel ja esta em execucao na porta 8081!
    echo [INFO] Abrindo janela dedicada do aplicativo...
    goto open_window_8081
)

REM 2. Se o bot nao estiver rodando, inicia o bot
echo [INFO] Iniciando Kenobi Bot e servicos de audio...
if exist "bot.exe" (
    start "Kenobi Bot Service" "bot.exe"
) else (
    start "Kenobi Bot Service" go run .\cmd\bot
)

echo [INFO] Aguardando inicializacao do servidor local...
timeout /t 3 /nobreak >nul 2>&1

:open_window_8080
REM 3. Abre a janela do aplicativo (Edge App Mode, Chrome App Mode ou Navegador padrao)
if exist "%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe" (
    start "" "%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe" --app=http://127.0.0.1:8080/ --window-size=1380,880
    exit /b 0
)
if exist "%ProgramFiles%\Microsoft\Edge\Application\msedge.exe" (
    start "" "%ProgramFiles%\Microsoft\Edge\Application\msedge.exe" --app=http://127.0.0.1:8080/ --window-size=1380,880
    exit /b 0
)
if exist "%ProgramFiles%\Google\Chrome\Application\chrome.exe" (
    start "" "%ProgramFiles%\Google\Chrome\Application\chrome.exe" --app=http://127.0.0.1:8080/ --window-size=1380,880
    exit /b 0
)

start http://127.0.0.1:8080/
exit /b 0

:open_window_8081
if exist "%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe" (
    start "" "%ProgramFiles(x86)%\Microsoft\Edge\Application\msedge.exe" --app=http://127.0.0.1:8081/ --window-size=1380,880
    exit /b 0
)
if exist "%ProgramFiles%\Microsoft\Edge\Application\msedge.exe" (
    start "" "%ProgramFiles%\Microsoft\Edge\Application\msedge.exe" --app=http://127.0.0.1:8081/ --window-size=1380,880
    exit /b 0
)
if exist "%ProgramFiles%\Google\Chrome\Application\chrome.exe" (
    start "" "%ProgramFiles%\Google\Chrome\Application\chrome.exe" --app=http://127.0.0.1:8081/ --window-size=1380,880
    exit /b 0
)

start http://127.0.0.1:8081/
exit /b 0

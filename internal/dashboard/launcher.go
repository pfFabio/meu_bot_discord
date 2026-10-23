package dashboard

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

// LaunchStandaloneWindow tenta abrir uma janela de aplicativo dedicada (sem barras de navegação)
// utilizando o Microsoft Edge ou Google Chrome em modo --app.
// Caso nenhum esteja disponível, recorre ao navegador padrão do sistema operacional.
func LaunchStandaloneWindow(url string) {
	if runtime.GOOS != "windows" {
		openDefaultBrowser(url)
		return
	}

	// 1. Caminhos conhecidos do Microsoft Edge
	edgeCandidates := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("LocalAppData"), "Microsoft", "Edge", "Application", "msedge.exe"),
	}

	for _, path := range edgeCandidates {
		if path != "" && fileExists(path) {
			slog.Info("Abrindo janela dedicada do Painel via Microsoft Edge App Mode...", "path", path)
			if err := spawnAppWindow(path, url); err == nil {
				return
			}
		}
	}

	// 2. Caminhos conhecidos do Google Chrome
	chromeCandidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
		filepath.Join(os.Getenv("LocalAppData"), "Google", "Chrome", "Application", "chrome.exe"),
	}

	for _, path := range chromeCandidates {
		if path != "" && fileExists(path) {
			slog.Info("Abrindo janela dedicada do Painel via Google Chrome App Mode...", "path", path)
			if err := spawnAppWindow(path, url); err == nil {
				return
			}
		}
	}

	// 3. Fallback: abre no navegador padrão
	slog.Info("Recorrendo ao navegador padrão do sistema...", "url", url)
	openDefaultBrowser(url)
}

func spawnAppWindow(browserExe, targetURL string) error {
	appArg := fmt.Sprintf("--app=%s", targetURL)
	sizeArg := "--window-size=1380,880"
	nameArg := "--app-id=kenobi_bot_dashboard"

	cmd := exec.Command(browserExe, appArg, sizeArg, nameArg)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008, // DETACHED_PROCESS no Windows
	}

	return cmd.Start()
}

func openDefaultBrowser(targetURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	if err := cmd.Start(); err != nil {
		slog.Warn("Não foi possível abrir o navegador automaticamente", "url", targetURL, "error", err)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

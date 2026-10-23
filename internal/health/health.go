// Package health provê rotinas de diagnóstico e verificação de saúde pré-execução para o Kenobi Bot.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/audio"
	"kenobi-bot/internal/config"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/dnd/repository"
	"kenobi-bot/internal/handler"
	"kenobi-bot/internal/userpref"
)

// Status indica a severidade de uma checagem
type Status string

const (
	StatusPass Status = "PASS"
	StatusWarn Status = "WARN"
	StatusFail Status = "FAIL"
)

// CheckItem detalha o resultado de uma verificação
type CheckItem struct {
	Category    string `json:"category"`
	Name        string `json:"name"`
	Status      Status `json:"status"`
	Message     string `json:"message"`
	Remediation string `json:"remediation,omitempty"`
}

// Report agrega todos os resultados do diagnóstico
type Report struct {
	Timestamp  time.Time   `json:"timestamp"`
	Checks     []CheckItem `json:"checks"`
	TotalPass  int         `json:"total_pass"`
	TotalWarn  int         `json:"total_warn"`
	TotalFail  int         `json:"total_fail"`
}

// AddCheck adiciona um resultado e atualiza os contadores
func (r *Report) AddCheck(c CheckItem) {
	r.Checks = append(r.Checks, c)
	switch c.Status {
	case StatusPass:
		r.TotalPass++
	case StatusWarn:
		r.TotalWarn++
	case StatusFail:
		r.TotalFail++
	}
}

// HasCriticalFailures indica se há falhas impeditivas para iniciar o bot
func (r *Report) HasCriticalFailures() bool {
	return r.TotalFail > 0
}

// FormatConsole formata o relatório com cores e visual amigável
func (r *Report) FormatConsole() string {
	var sb strings.Builder

	sb.WriteString("\n======================================================================\n")
	sb.WriteString(" 🛡️  KENOBI BOT — DIAGNÓSTICO E SAÚDE PRÉ-EXECUÇÃO\n")
	sb.WriteString(fmt.Sprintf(" Data/Hora: %s\n", r.Timestamp.Format("02/01/2006 15:04:05")))
	sb.WriteString("======================================================================\n\n")

	currentCategory := ""
	for _, c := range r.Checks {
		if c.Category != currentCategory {
			currentCategory = c.Category
			sb.WriteString(fmt.Sprintf("📁 [%s]\n", currentCategory))
		}

		icon := "✅"
		tag := "[OK]"
		colorCode := "\033[32m" // Verde
		if c.Status == StatusWarn {
			icon = "⚠️"
			tag = "[AVISO]"
			colorCode = "\033[33m" // Amarelo
		} else if c.Status == StatusFail {
			icon = "❌"
			tag = "[ERRO]"
			colorCode = "\033[31m" // Vermelho
		}
		resetCode := "\033[0m"

		sb.WriteString(fmt.Sprintf("  %s %s%-8s%s %-32s: %s\n", icon, colorCode, tag, resetCode, c.Name, c.Message))
		if c.Remediation != "" {
			sb.WriteString(fmt.Sprintf("     💡 Como consertar: %s\n", c.Remediation))
		}
	}

	sb.WriteString("\n----------------------------------------------------------------------\n")
	statusSummary := "\033[32mSISTEMA SAUDÁVEL E PRONTO!\033[0m"
	if r.TotalFail > 0 {
		statusSummary = "\033[31mFALHAS CRÍTICAS ENCONTRADAS (Corrija antes de rodar o bot!)\033[0m"
	} else if r.TotalWarn > 0 {
		statusSummary = "\033[33mPRONTO COM AVISOS (Verifique as recomendações acima)\033[0m"
	}

	sb.WriteString(fmt.Sprintf(" RESUMO: %d Passou | %d Avisos | %d Falhas\n", r.TotalPass, r.TotalWarn, r.TotalFail))
	sb.WriteString(fmt.Sprintf(" STATUS: %s\n", statusSummary))
	sb.WriteString("======================================================================\n\n")

	return sb.String()
}

// RunDiagnostics executa todas as suítes de teste de saúde
func RunDiagnostics(rootPath string) *Report {
	report := &Report{
		Timestamp: time.Now(),
		Checks:    make([]CheckItem, 0),
	}

	checkConfig(report, rootPath)
	checkExternalTools(report)
	checkAudioBridge(report, rootPath)
	checkMusicAndAmbience(report, rootPath)
	checkDNDDatabase(report, rootPath)
	checkSlashCommands(report)

	return report
}

// 1. Configuração e Variáveis de Ambiente
func checkConfig(r *Report, rootPath string) {
	category := "Configuração (.env)"

	envPath := filepath.Join(rootPath, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Arquivo .env",
			Status:      StatusWarn,
			Message:     "Arquivo .env não encontrado no diretório raiz.",
			Remediation: "Crie um arquivo .env copiando o modelo .env.example e configure o DISCORD_TOKEN.",
		})
	} else {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Arquivo .env",
			Status:   StatusPass,
			Message:  "Arquivo .env encontrado no diretório do bot.",
		})
	}

	cfg, err := config.LoadFrom(envPath)
	if err != nil || strings.TrimSpace(cfg.BotToken) == "" {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "DISCORD_TOKEN",
			Status:      StatusFail,
			Message:     "Token do Discord não configurado ou vazio.",
			Remediation: "Adicione DISCORD_TOKEN=seu_token_aqui no arquivo .env ou nas variáveis de ambiente.",
		})
		return
	}

	// Validar formato básico do token do Discord (2 pontos separando partes)
	tokenParts := strings.Split(cfg.BotToken, ".")
	if len(tokenParts) < 3 {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Formato do DISCORD_TOKEN",
			Status:      StatusWarn,
			Message:     "Token não parece seguir o padrão clássico do Discord (3 seções separadas por pontos).",
			Remediation: "Certifique-se de que copiou o Token do Bot no Discord Developer Portal e não o Client Secret.",
		})
	} else {
		masked := tokenParts[0] + ".***." + tokenParts[2][:4] + "..."
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "DISCORD_TOKEN",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Token presente e formatado corretamente (%s)", masked),
		})
	}
}

// 2. Ferramentas do Sistema (FFmpeg, Node, yt-dlp)
func checkExternalTools(r *Report) {
	category := "Ferramentas do Sistema"

	// Node.js
	if nodeVer, err := execCommand("node", "-v"); err != nil {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Node.js Runtime",
			Status:      StatusFail,
			Message:     "Node.js não foi encontrado no PATH do sistema.",
			Remediation: "Instale o Node.js v18+ a partir de https://nodejs.org ou via 'winget install OpenJS.NodeJS'.",
		})
	} else {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Node.js Runtime",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Instalado (%s)", strings.TrimSpace(nodeVer)),
		})
	}

	// FFmpeg
	if ffmpegVer, err := execCommand("ffmpeg", "-version"); err != nil {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "FFmpeg Audio Engine",
			Status:      StatusFail,
			Message:     "FFmpeg não foi encontrado no PATH.",
			Remediation: "Instale o FFmpeg e garanta que sua pasta bin/ está no PATH do Windows.",
		})
	} else {
		firstLine := strings.Split(ffmpegVer, "\n")[0]
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "FFmpeg Audio Engine",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Instalado (%s)", strings.TrimSpace(firstLine)),
		})
	}

	// FFprobe
	if ffprobeVer, err := execCommand("ffprobe", "-version"); err != nil {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "FFprobe Inspector",
			Status:      StatusWarn,
			Message:     "FFprobe não encontrado. Leitura de duração exata das faixas locais pode usar fallback.",
			Remediation: "Geralmente vem no mesmo pacote do FFmpeg. Certifique-se de que ffprobe.exe está no PATH.",
		})
	} else {
		firstLine := strings.Split(ffprobeVer, "\n")[0]
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "FFprobe Inspector",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Instalado (%s)", strings.TrimSpace(firstLine)),
		})
	}

	// yt-dlp
	if ytVer, err := execCommand("yt-dlp", "--version"); err != nil {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "yt-dlp Downloader/Streamer",
			Status:      StatusWarn,
			Message:     "yt-dlp não encontrado. Reprodução de links externos do YouTube usará fallback.",
			Remediation: "Instale o yt-dlp com 'pip install yt-dlp' ou baixe o binário para o PATH.",
		})
	} else {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "yt-dlp Downloader/Streamer",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Instalado (versão %s)", strings.TrimSpace(ytVer)),
		})
	}
}

// 3. Microserviço Audio Bridge (DAVE E2EE)
func checkAudioBridge(r *Report, rootPath string) {
	category := "Microserviço de Áudio (DAVE Sidecar)"

	indexPath := filepath.Join(rootPath, "audio-bridge", "index.js")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Código Audio Bridge",
			Status:      StatusFail,
			Message:     "audio-bridge/index.js não encontrado!",
			Remediation: "Verifique se a pasta audio-bridge está presente no diretório do projeto.",
		})
		return
	}
	r.AddCheck(CheckItem{
		Category: category,
		Name:     "Código Audio Bridge",
		Status:   StatusPass,
		Message:  "audio-bridge/index.js encontrado e pronto.",
	})

	// Verificar node_modules
	nodeModules := filepath.Join(rootPath, "audio-bridge", "node_modules")
	voiceModule := filepath.Join(nodeModules, "@discordjs", "voice")
	if _, err := os.Stat(voiceModule); os.IsNotExist(err) {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Dependências NPM (@discordjs/voice)",
			Status:      StatusFail,
			Message:     "Dependências do Node.js não instaladas na pasta audio-bridge.",
			Remediation: "Navegue até a pasta audio-bridge e execute 'npm install'.",
		})
	} else {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Dependências NPM (@discordjs/voice)",
			Status:   StatusPass,
			Message:  "Módulos do Discord Voice e DAVE instalados.",
		})
	}

	// Verificar se o Bridge já está respondendo na porta 2389
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://127.0.0.1:2389/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		var healthData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&healthData)
		resp.Body.Close()
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Porta 2389 (HTTP API)",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Microserviço já está online e ativo (DAVE: %v, Uptime: %.0fs)", healthData["dave"], healthData["uptime"]),
		})
	} else {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Porta 2389 (HTTP API)",
			Status:   StatusPass,
			Message:  "Microserviço inativo no momento (o Supervisor iniciará automaticamente com o bot).",
		})
	}
}

// 4. Músicas Locais e Ambientes de Curse of Strahd
func checkMusicAndAmbience(r *Report, rootPath string) {
	category := "Músicas & Ambientes (Baróvia)"

	// 1. Diretório de Músicas
	musicDir := filepath.Join(rootPath, "musicas")
	if info, err := os.Stat(musicDir); os.IsNotExist(err) || !info.IsDir() {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Diretório /musicas",
			Status:      StatusWarn,
			Message:     "Diretório musicas/ não encontrado.",
			Remediation: "Crie a pasta musicas/ no diretório raiz do bot.",
		})
	} else {
		tracks := audio.ScanLocalMusic()
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Pasta musicas/ (Músicas)",
			Status:   StatusPass,
			Message:  fmt.Sprintf("%d faixa(s) de música pronta(s) na pasta musicas/", len(tracks)),
		})
	}

	// 2. Diretório de Ambientes
	ambienceDir := filepath.Join(rootPath, "ambiente")
	if info, err := os.Stat(ambienceDir); os.IsNotExist(err) || !info.IsDir() {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Diretório /ambiente",
			Status:      StatusWarn,
			Message:     "Diretório ambiente/ não encontrado.",
			Remediation: "Crie a pasta ambiente/ no diretório raiz do bot para as faixas de Curse of Strahd.",
		})
	} else {
		cosList := audio.GetCurseOfStrahdAmbienceList()
		availableCoS := audio.GetAvailableAmbienceTracks()
		pctReady := (len(availableCoS) * 100) / len(cosList)

		if len(availableCoS) == len(cosList) {
			r.AddCheck(CheckItem{
				Category: category,
				Name:     "Pasta ambiente/ (Curse of Strahd)",
				Status:   StatusPass,
				Message:  fmt.Sprintf("100%% das faixas prontas localmente (%d/%d em MP3).", len(availableCoS), len(cosList)),
			})
		} else {
			r.AddCheck(CheckItem{
				Category: category,
				Name:     "Pasta ambiente/ (Curse of Strahd)",
				Status:   StatusPass,
				Message:  fmt.Sprintf("%d/%d faixas em MP3 (%d%%). Faixas restantes tocarão via streaming do YouTube.", len(availableCoS), len(cosList), pctReady),
			})
		}
	}

	// 3. Verificar se há arquivos residuais .part ou .webm ainda sendo convertidos
	checkDirs := []string{musicDir, ambienceDir}
	var convertingFiles []string
	for _, d := range checkDirs {
		entries, _ := os.ReadDir(d)
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".webm") || strings.HasSuffix(entry.Name(), ".part") {
				convertingFiles = append(convertingFiles, entry.Name())
			}
		}
	}

	if len(convertingFiles) > 0 {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Status de Downloads/Conversões",
			Status:   StatusPass,
			Message:  fmt.Sprintf("%d arquivo(s) sendo baixado(s)/convertido(s) em segundo plano.", len(convertingFiles)),
		})
	}
}

// 5. Acervo de Dados do D&D 2024 (5.5e)
func checkDNDDatabase(r *Report, rootPath string) {
	category := "Acervo D&D 2024 (5.5e)"

	dndRepo, err := repository.LoadDNDData(rootPath)
	if err != nil {
		r.AddCheck(CheckItem{
			Category:    category,
			Name:        "Carregamento D&D 2024",
			Status:      StatusFail,
			Message:     fmt.Sprintf("Falha ao inicializar banco de dados D&D: %v", err),
			Remediation: "Verifique a integridade dos dados e do parser do repositório D&D.",
		})
		return
	}

	classes := dndRepo.GetClasses()
	species := dndRepo.GetSpeciesList("")
	spells := dndRepo.SearchSpells("", -1, "", "", 0)
	monsters := dndRepo.SearchMonsters("", "", "", 0)
	items := dndRepo.SearchItems("", "", "", "", 0)
	rules := dndRepo.SearchRules("", 0)

	r.AddCheck(CheckItem{
		Category: category,
		Name:     "Dados do D&D 2024",
		Status:   StatusPass,
		Message: fmt.Sprintf("Classes: %d | Espécies: %d | Magias: %d | Monstros: %d | Itens: %d | Regras: %d",
			len(classes), len(species), len(spells),
			len(monsters), len(items), len(rules)),
	})
}

// 6. Integridade dos Comandos Slash
func checkSlashCommands(r *Report) {
	category := "Comandos Slash do Discord"

	// Mock do cliente para instanciar os handlers
	client := audio.NewBridgeClient("http://127.0.0.1:2389")
	audioH := handler.NewAudioHandler(client)
	recH := handler.NewRecordingHandler(client)
	diceRoller := dice.NewRoller()
	prefService := userpref.NewMemoryService()
	_ = handler.NewDiceHandler(diceRoller, prefService)

	var allCommands []*discordgo.ApplicationCommand
	allCommands = append(allCommands, audioH.GetSlashCommands()...)
	allCommands = append(allCommands, recH.GetSlashCommands()...)

	nameRegex := regexp.MustCompile(`^[-_\p{L}\p{N}]+$`)
	seenNames := make(map[string]bool)
	hasErrors := false

	for _, cmd := range allCommands {
		if seenNames[cmd.Name] {
			r.AddCheck(CheckItem{
				Category:    category,
				Name:        "Colisão de Comando",
				Status:      StatusFail,
				Message:     fmt.Sprintf("Nome de comando duplicado: /%s", cmd.Name),
				Remediation: "Renomeie um dos comandos conflitantes.",
			})
			hasErrors = true
		}
		seenNames[cmd.Name] = true

		if len(cmd.Name) == 0 || len(cmd.Name) > 32 || !nameRegex.MatchString(cmd.Name) {
			r.AddCheck(CheckItem{
				Category:    category,
				Name:        "Nome Inválido",
				Status:      StatusFail,
				Message:     fmt.Sprintf("Nome do comando /%s viola as regras do Discord.", cmd.Name),
				Remediation: "Nomes de slash command devem ter entre 1 e 32 caracteres alfanuméricos sem espaços.",
			})
			hasErrors = true
		}

		if len(cmd.Description) == 0 || len(cmd.Description) > 100 {
			r.AddCheck(CheckItem{
				Category:    category,
				Name:        "Descrição Inválida",
				Status:      StatusFail,
				Message:     fmt.Sprintf("Descrição do comando /%s excede 100 caracteres ou está vazia.", cmd.Name),
				Remediation: "Ajuste a descrição para ter no máximo 100 caracteres.",
			})
			hasErrors = true
		}
	}

	if !hasErrors {
		r.AddCheck(CheckItem{
			Category: category,
			Name:     "Validação dos Slash Commands",
			Status:   StatusPass,
			Message:  fmt.Sprintf("Todos os %d comandos foram validados conforme as diretrizes do Discord.", len(allCommands)),
		})
	}
}

func execCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmdPath := name
	if p, err := exec.LookPath(name); err == nil {
		cmdPath = p
	} else if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		candidates := []string{
			filepath.Join(userProfile, "AppData", "Local", "Programs", "Python", "Python313", "Scripts", name+".exe"),
			filepath.Join(userProfile, "AppData", "Local", "Programs", "Python", "Python312", "Scripts", name+".exe"),
			filepath.Join(userProfile, "AppData", "Local", "Programs", "Python", "Python311", "Scripts", name+".exe"),
			filepath.Join(userProfile, "AppData", "Roaming", "Python", "Scripts", name+".exe"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				cmdPath = c
				break
			}
		}
	}

	cmd := exec.CommandContext(ctx, cmdPath, args...)
	out, err := cmd.Output()
	if err != nil {
		// Fallback para python -m yt_dlp caso seja yt-dlp instalado no ambiente Python
		if name == "yt-dlp" {
			pyCmd := exec.CommandContext(ctx, "python", append([]string{"-m", "yt_dlp"}, args...)...)
			if pyOut, pyErr := pyCmd.Output(); pyErr == nil {
				return string(pyOut), nil
			}
		}
		return "", err
	}
	return string(out), nil
}

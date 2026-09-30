package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Resolver é responsável por converter URLs ou termos de busca em faixas reproduzíveis
type Resolver struct{}

// NewResolver instancia um novo resolvedor de áudio
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve analisa uma query (URL ou busca) e extrai o título e a URL direta de áudio via yt-dlp ou preset
func (r *Resolver) Resolve(query, requester string) (*Track, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("termo de busca ou URL não pode estar vazio")
	}

	// 0. Verificar se é uma Música Local da pasta /musicas
	if local, ok := FindLocalTrack(query); ok {
		return &Track{
			Title:     local.Emoji + " " + local.Title,
			URL:       local.Path,
			Duration:  local.Duration,
			Requester: requester,
			IsLocal:   true,
			Filename:  local.Filename,
		}, nil
	}

	// 1. Verificar se é um Preset de D&D
	if preset, ok := FindPreset(strings.ToLower(query)); ok {
		return &Track{
			Title:     preset.Name,
			URL:       preset.URL,
			Duration:  "Stream Contínuo 24/7",
			Requester: requester,
			IsPreset:  true,
			PresetKey: preset.Key,
		}, nil
	}

	// 2. Se for link direto de áudio (ex: .mp3, .ogg, .aac, streams)
	if strings.HasSuffix(strings.ToLower(query), ".mp3") ||
		strings.HasSuffix(strings.ToLower(query), ".ogg") ||
		strings.HasSuffix(strings.ToLower(query), ".wav") {
		return &Track{
			Title:     "Áudio Direto",
			URL:       query,
			Duration:  "Arquivo de Áudio",
			Requester: requester,
		}, nil
	}

	// 3. Usar yt-dlp para extrair stream direto do YouTube, SoundCloud ou Bandcamp
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var target string
	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") {
		target = query
	} else {
		target = fmt.Sprintf("ytsearch1:%s", query)
	}

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-f", "bestaudio/best",
		"--print", "%(title)s",
		"-g",
		"--print", "%(duration_string)s",
		"--no-playlist",
		"--no-warnings",
		target,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Se o yt-dlp falhar (ex: bloqueio de IP ou link inválido)
		return nil, fmt.Errorf("não foi possível extrair o áudio: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 2 {
		return nil, errors.New("nenhum resultado de áudio encontrado para a busca")
	}

	title := strings.TrimSpace(lines[0])
	streamURL := strings.TrimSpace(lines[1])
	duration := "Desconhecida"
	if len(lines) >= 3 && strings.TrimSpace(lines[2]) != "NA" {
		duration = strings.TrimSpace(lines[2])
	}

	return &Track{
		Title:     title,
		URL:       streamURL,
		Duration:  duration,
		Requester: requester,
	}, nil
}

package audio

import (
	"os"
	"path/filepath"
	"strings"
)

var validAudioExtensions = map[string]bool{
	".mp3":  true,
	".ogg":  true,
	".wav":  true,
	".flac": true,
	".m4a":  true,
	".aac":  true,
}

// getMusicDir retorna o caminho da pasta musicas independente do diretório de execução
func getMusicDir() string {
	if envDir := os.Getenv("MUSIC_DIR"); envDir != "" {
		if _, err := os.Stat(envDir); err == nil {
			return envDir
		}
	}
	candidates := []string{
		"musicas",
		"../../musicas",
		"../musicas",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "musicas"
}

// ScanLocalMusic varre a pasta "musicas" e retorna as faixas disponíveis
func ScanLocalMusic() []LocalTrack {
	dir := getMusicDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []LocalTrack{}
	}

	var tracks []LocalTrack
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !validAudioExtensions[ext] {
			continue
		}

		baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		title, emoji := formatTrackTitle(baseName)

		tracks = append(tracks, LocalTrack{
			Filename: entry.Name(),
			Title:    title,
			Emoji:    emoji,
			Duration: "Local",
			Path:     filepath.Join(dir, entry.Name()),
		})
	}

	return tracks
}

// FindLocalTrack procura uma faixa local por nome de arquivo, prefixo "local:" ou correspondência de texto
func FindLocalTrack(query string) (*LocalTrack, bool) {
	tracks := ScanLocalMusic()
	cleanQuery := strings.TrimSpace(query)
	lowerQuery := strings.ToLower(cleanQuery)

	if strings.HasPrefix(lowerQuery, "local:") {
		target := strings.TrimPrefix(lowerQuery, "local:")
		for _, t := range tracks {
			if strings.ToLower(t.Filename) == target ||
				strings.ToLower(strings.TrimSuffix(t.Filename, filepath.Ext(t.Filename))) == target {
				return &t, true
			}
		}
	}

	// 1. Match exato por nome de arquivo ou base sem extensão
	for _, t := range tracks {
		baseWithoutExt := strings.ToLower(strings.TrimSuffix(t.Filename, filepath.Ext(t.Filename)))
		if strings.ToLower(t.Filename) == lowerQuery || baseWithoutExt == lowerQuery {
			return &t, true
		}
	}

	// 2. Match parcial / substring
	for _, t := range tracks {
		baseWithoutExt := strings.ToLower(strings.TrimSuffix(t.Filename, filepath.Ext(t.Filename)))
		if strings.Contains(baseWithoutExt, lowerQuery) ||
			strings.Contains(strings.ToLower(t.Title), lowerQuery) ||
			strings.Contains(lowerQuery, baseWithoutExt) {
			return &t, true
		}
	}

	return nil, false
}

// formatTrackTitle gera um título amigável e escolhe o emoji temático apropriado
func formatTrackTitle(baseName string) (string, string) {
	lower := strings.ToLower(baseName)

	if strings.Contains(lower, "mago") && strings.Contains(lower, "strahd") {
		return "Mago e Strahd (História de Baróvia)", "🧙‍♂️"
	}
	if strings.Contains(lower, "vistani") && (strings.Contains(lower, "introdução") || strings.Contains(lower, "introducao")) {
		return "Vistani (Canção de Introdução)", "🎪"
	}
	if strings.Contains(lower, "vistani") && strings.Contains(lower, "festa") {
		return "Vistani (Festa de Abertura)", "🎪"
	}
	if strings.Contains(lower, "vistani") || strings.Contains(lower, "gypsy") || strings.Contains(lower, "balkan") || strings.Contains(lower, "cigano") {
		return capitalizeFirst(baseName), "🎪"
	}
	if strings.Contains(lower, "strahd") || strings.Contains(lower, "vampiro") {
		return capitalizeFirst(baseName), "🩸"
	}
	if strings.Contains(lower, "combate") || strings.Contains(lower, "batalha") {
		return capitalizeFirst(baseName), "⚔️"
	}
	if strings.Contains(lower, "taverna") {
		return capitalizeFirst(baseName), "🍻"
	}

	return capitalizeFirst(baseName), "🎵"
}

func capitalizeFirst(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

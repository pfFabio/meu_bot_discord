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

// GetMusicDir retorna o caminho da pasta musicas independente do diretório de execução
func GetMusicDir() string {
	return getMusicDir()
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

	// 2. Match parcial no nome do arquivo (prioridade sobre título gerado)
	for _, t := range tracks {
		baseWithoutExt := strings.ToLower(strings.TrimSuffix(t.Filename, filepath.Ext(t.Filename)))
		if strings.Contains(baseWithoutExt, lowerQuery) || strings.Contains(lowerQuery, baseWithoutExt) {
			return &t, true
		}
	}

	// 3. Match parcial no título amigável formatado
	for _, t := range tracks {
		if strings.Contains(strings.ToLower(t.Title), lowerQuery) {
			return &t, true
		}
	}

	// 4. Fallback: procurar nas faixas de ambiente da pasta ambiente/
	for _, amb := range GetAvailableAmbienceTracks() {
		baseWithoutExt := strings.ToLower(strings.TrimSuffix(amb.Filename, filepath.Ext(amb.Filename)))
		if strings.ToLower(amb.Filename) == lowerQuery || baseWithoutExt == lowerQuery ||
			strings.Contains(baseWithoutExt, lowerQuery) || strings.Contains(strings.ToLower(amb.Title), lowerQuery) {
			return &LocalTrack{
				Filename: amb.Filename,
				Title:    amb.Title,
				Emoji:    amb.Emoji,
				Duration: amb.Duration,
				Path:     amb.Path,
			}, true
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

	// Faixas de Curse of Strahd (CoS)
	if strings.Contains(lower, "svalich") {
		return "Floresta Svalich (Svalich Woods)", "🌲"
	}
	if strings.Contains(lower, "death house dungeon") || (strings.Contains(lower, "death house") && strings.Contains(lower, "dungeon")) {
		return "Masmorra da Casa da Morte (Death House Dungeon)", "💀"
	}
	if strings.Contains(lower, "death house") {
		return "Casa da Morte (Death House)", "🏚️"
	}
	if strings.Contains(lower, "village of barovia") || (strings.Contains(lower, "barovia") && strings.Contains(lower, "village")) || (strings.Contains(lower, "vila") && strings.Contains(lower, "barovia")) {
		return "Vila de Baróvia (Village of Barovia)", "🏚️"
	}
	if strings.Contains(lower, "vallaki") {
		return "Cidade de Vallaki (Town of Vallaki)", "🏘️"
	}
	if strings.Contains(lower, "bonegrinder") || strings.Contains(lower, "moedor") {
		return "Velho Moedor de Ossos (Old Bonegrinder)", "🦴"
	}
	if strings.Contains(lower, "argynvostholt") {
		return "Argynvostholt (Mansão do Dragão)", "🛡️"
	}
	if strings.Contains(lower, "catacombs") || strings.Contains(lower, "catacumbas") {
		return "Catacumbas de Ravenloft (Catacombs)", "⚰️"
	}
	if strings.Contains(lower, "castle ravenloft") || (strings.Contains(lower, "castelo") && strings.Contains(lower, "ravenloft")) || (strings.Contains(lower, "ravenloft") && !strings.Contains(lower, "mists")) {
		return "Castelo Ravenloft (Castle Ravenloft)", "🏰"
	}
	if strings.Contains(lower, "krezk") {
		return "Vila de Krezk (Village of Krezk)", "🏡"
	}
	if strings.Contains(lower, "markovia") || strings.Contains(lower, "abbey") || strings.Contains(lower, "abadia") {
		return "Abadia de Santa Markovia (Abbey of St. Markovia)", "⛪"
	}
	if strings.Contains(lower, "madhouse") || strings.Contains(lower, "asilo") {
		return "Asilo dos Loucos (Madhouse)", "🤪"
	}
	if strings.Contains(lower, "tsolenka") {
		return "Passo de Tsolenka (Tsolenka Pass)", "❄️"
	}
	if strings.Contains(lower, "van richten") || strings.Contains(lower, "richten") {
		return "Torre de Van Richten (Van Richten's Tower)", "🗼"
	}
	if strings.Contains(lower, "berez") {
		return "Ruínas de Berez (Ruins of Berez)", "🌊"
	}
	if strings.Contains(lower, "wines") || strings.Contains(lower, "vinhedo") || strings.Contains(lower, "vinhos") {
		return "Mago dos Vinhos (Wizard of the Wines)", "🍇"
	}
	if strings.Contains(lower, "amber temple") || strings.Contains(lower, "templo de âmbar") || strings.Contains(lower, "templo de ambar") {
		return "Templo de Âmbar (Amber Temple)", "🏛️"
	}
	if strings.Contains(lower, "mists") || strings.Contains(lower, "névoas") || strings.Contains(lower, "nevoas") {
		return "Névoas de Ravenloft (Mists of Ravenloft)", "🌫️"
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

	clean := strings.TrimPrefix(baseName, "CoS - ")
	clean = strings.TrimPrefix(clean, "D&D Ambience - [CoS] - ")
	return capitalizeFirst(clean), "🎵"
}

func capitalizeFirst(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

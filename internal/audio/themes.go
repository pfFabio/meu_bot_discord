package audio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ThemeTrack define uma trilha musical temática de Curse of Strahd (personagem, combate, facção, relíquia)
type ThemeTrack struct {
	Index       int    `json:"index"`
	ID          string `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Filename    string `json:"filename"`
	Category    string `json:"category"`
	Emoji       string `json:"emoji"`
	Duration    string `json:"duration"`
	YouTubeURL  string `json:"youtube_url"`
	Path        string `json:"path,omitempty"`
	IsAvailable bool   `json:"is_available"`
}

var (
	themesCatalogCache []ThemeTrack
	themesCatalogMutex sync.RWMutex
)

// GetTemasDir retorna o diretório da pasta temas
func GetTemasDir() string {
	return getTemasDir()
}

// getTemasDir localiza a pasta temas/ no diretório raiz do bot
func getTemasDir() string {
	if envDir := os.Getenv("TEMAS_DIR"); envDir != "" {
		if _, err := os.Stat(envDir); err == nil {
			return envDir
		}
	}
	candidates := []string{
		"temas",
		"../../temas",
		"../temas",
		"musicas/temas",
		"../../musicas/temas",
		"../musicas/temas",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "temas"
}

// GetThemeTracks retorna todas as faixas temáticas do catálogo com status de disponibilidade local
func GetThemeTracks() []ThemeTrack {
	tracks := loadThemeCatalog()
	temasDir := getTemasDir()

	result := make([]ThemeTrack, len(tracks))
	copy(result, tracks)

	for i := range result {
		// Checa se arquivo .mp3 ou .m4a já existe localmente
		fullPath := filepath.Join(temasDir, result[i].Filename)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() && info.Size() > 0 {
			result[i].Path = fullPath
			result[i].IsAvailable = true
			continue
		}

		// Fallback para extensão .m4a caso tenha sido baixado em m4a
		m4aFilename := strings.TrimSuffix(result[i].Filename, filepath.Ext(result[i].Filename)) + ".m4a"
		m4aPath := filepath.Join(temasDir, m4aFilename)
		if info, err := os.Stat(m4aPath); err == nil && !info.IsDir() && info.Size() > 0 {
			result[i].Path = m4aPath
			result[i].Filename = m4aFilename
			result[i].IsAvailable = true
			continue
		}

		result[i].IsAvailable = false
	}

	return result
}

// GetThemeCategories retorna as categorias ordenadas de temas
func GetThemeCategories() []string {
	return []string{
		"👑 Personagens & NPCs",
		"⚔️ Combates & Encontros",
		"✨ Relíquias & Destino",
		"🦅 Facções & Ordens",
		"🕯️ Cenários & Mistério",
	}
}

// GetThemeTracksByCategory retorna as faixas temáticas de uma categoria específica
func GetThemeTracksByCategory(category string) []ThemeTrack {
	all := GetThemeTracks()
	var filtered []ThemeTrack
	cleanCat := strings.ToLower(strings.TrimSpace(category))

	for _, t := range all {
		tCat := strings.ToLower(t.Category)
		if strings.Contains(tCat, cleanCat) || strings.Contains(cleanCat, tCat) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// FindThemeTrack procura uma faixa de tema por chave, nome de arquivo, título ou consulta
func FindThemeTrack(query string) (*ThemeTrack, bool) {
	tracks := GetThemeTracks()
	cleanQuery := strings.TrimSpace(query)
	lowerQuery := strings.ToLower(cleanQuery)

	if strings.HasPrefix(lowerQuery, "theme:") {
		target := strings.TrimPrefix(lowerQuery, "theme:")
		for _, t := range tracks {
			if strings.ToLower(t.Key) == target || strings.ToLower(t.Filename) == target || strings.ToLower(t.Title) == target {
				return &t, true
			}
		}
	}

	if strings.HasPrefix(lowerQuery, "local:") {
		target := strings.TrimPrefix(lowerQuery, "local:")
		for _, t := range tracks {
			if strings.ToLower(t.Filename) == target || strings.ToLower(strings.TrimSuffix(t.Filename, filepath.Ext(t.Filename))) == target {
				return &t, true
			}
		}
	}

	// 1. Match exato por chave, nome de arquivo ou título
	for _, t := range tracks {
		if strings.ToLower(t.Key) == lowerQuery ||
			strings.ToLower(t.Filename) == lowerQuery ||
			strings.ToLower(strings.TrimSuffix(t.Filename, filepath.Ext(t.Filename))) == lowerQuery ||
			strings.ToLower(t.Title) == lowerQuery {
			return &t, true
		}
	}

	// 2. Match parcial no título ou nome de arquivo
	for _, t := range tracks {
		lowerTitle := strings.ToLower(t.Title)
		lowerFile := strings.ToLower(t.Filename)
		if strings.Contains(lowerTitle, lowerQuery) || strings.Contains(lowerFile, lowerQuery) || strings.Contains(lowerQuery, lowerTitle) {
			return &t, true
		}
	}

	return nil, false
}

// loadThemeCatalog carrega o catálogo do arquivo JSON ou inicializa fallback
func loadThemeCatalog() []ThemeTrack {
	themesCatalogMutex.RLock()
	if len(themesCatalogCache) > 0 {
		defer themesCatalogMutex.RUnlock()
		return themesCatalogCache
	}
	themesCatalogMutex.RUnlock()

	themesCatalogMutex.Lock()
	defer themesCatalogMutex.Unlock()

	if len(themesCatalogCache) > 0 {
		return themesCatalogCache
	}

	catalogFile := filepath.Join(getTemasDir(), "temas_catalog.json")
	if data, err := os.ReadFile(catalogFile); err == nil {
		var list []ThemeTrack
		if err := json.Unmarshal(data, &list); err == nil && len(list) > 0 {
			themesCatalogCache = list
			return themesCatalogCache
		}
	}

	// Fallback embutido com os principais temas
	themesCatalogCache = getDefaultThemes()
	return themesCatalogCache
}

// getDefaultThemes fornece os temas essenciais como fallback
func getDefaultThemes() []ThemeTrack {
	return []ThemeTrack{
		{
			Index:      15,
			ID:         "WXYz0mkt8s8",
			Key:        "ismark_kolyanovich",
			Title:      "Ismark Kolyanovich (Heroic Deeds)",
			Filename:   "CoS - Ismark Kolyanovich Theme.mp3",
			Category:   "👑 Personagens & NPCs",
			Emoji:      "👑",
			Duration:   "59:53",
			YouTubeURL: "https://www.youtube.com/watch?v=WXYz0mkt8s8",
		},
		{
			Index:      16,
			ID:         "lTjkTCtYBu8",
			Key:        "ireena_kolyana",
			Title:      "Ireena Kolyana (Theme)",
			Filename:   "CoS - Ireena Kolyana Theme.mp3",
			Category:   "👑 Personagens & NPCs",
			Emoji:      "👑",
			Duration:   "60:29",
			YouTubeURL: "https://www.youtube.com/watch?v=lTjkTCtYBu8",
		},
		{
			Index:      23,
			ID:         "6qGawsBZIT8",
			Key:        "strahd_von_zarovich",
			Title:      "Strahd von Zarovich (Cello Theme)",
			Filename:   "CoS - Strahd von Zarovich.mp3",
			Category:   "👑 Personagens & NPCs",
			Emoji:      "🩸",
			Duration:   "62:00",
			YouTubeURL: "https://www.youtube.com/watch?v=6qGawsBZIT8",
		},
		{
			Index:      25,
			ID:         "Gbut7lFzEgA",
			Key:        "strahd_battle_theme",
			Title:      "Strahd Battle Theme (BBEG)",
			Filename:   "CoS - Strahd Battle Theme.mp3",
			Category:   "⚔️ Combates & Encontros",
			Emoji:      "⚔️",
			Duration:   "60:55",
			YouTubeURL: "https://www.youtube.com/watch?v=Gbut7lFzEgA",
		},
		{
			Index:      27,
			ID:         "mtQ9hizvZzM",
			Key:        "rahadin_theme",
			Title:      "Rahadin Theme (Chamberlain)",
			Filename:   "CoS - Rahadin Theme.mp3",
			Category:   "👑 Personagens & NPCs",
			Emoji:      "👑",
			Duration:   "60:24",
			YouTubeURL: "https://www.youtube.com/watch?v=mtQ9hizvZzM",
		},
		{
			Index:      31,
			ID:         "c-IJ0Hrn2_E",
			Key:        "strahd_brides_theme",
			Title:      "Strahd's Brides Theme",
			Filename:   "CoS - Strahd's Brides Theme.mp3",
			Category:   "👑 Personagens & NPCs",
			Emoji:      "👑",
			Duration:   "61:14",
			YouTubeURL: "https://www.youtube.com/watch?v=c-IJ0Hrn2_E",
		},
		{
			Index:      61,
			ID:         "4qAtcKE1I5k",
			Key:        "ezmerelda",
			Title:      "Ezmerelda d'Avenir",
			Filename:   "CoS - Ezmerelda.mp3",
			Category:   "👑 Personagens & NPCs",
			Emoji:      "👑",
			Duration:   "61:00",
			YouTubeURL: "https://www.youtube.com/watch?v=4qAtcKE1I5k",
		},
		{
			Index:      80,
			ID:         "gsow6nhq03I",
			Key:        "sunsword",
			Title:      "Sunsword (Espada Solar)",
			Filename:   "CoS - Sunsword.mp3",
			Category:   "✨ Relíquias & Destino",
			Emoji:      "✨",
			Duration:   "00:48",
			YouTubeURL: "https://www.youtube.com/watch?v=gsow6nhq03I",
		},
	}
}

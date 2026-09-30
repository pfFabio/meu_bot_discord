package audio

// Track representa uma faixa de áudio na fila de reprodução
type Track struct {
	Title     string // Título da música / efeito
	URL       string // URL de stream ou arquivo
	Duration  string // Duração formatada
	Requester string // Nome do usuário que solicitou
	IsPreset  bool   // Se é uma trilha de D&D pré-configurada
	PresetKey string // Chave do preset (ex: taverna, combate)
	IsLocal   bool   // Se é uma música local da pasta /musicas
	Filename  string // Nome do arquivo local (se aplicável)
}

// LocalTrack representa um arquivo de áudio presente na pasta /musicas
type LocalTrack struct {
	Filename string `json:"filename"` // Nome do arquivo (ex: mago e strahd.mp3)
	Title    string `json:"title"`    // Título amigável formatado
	Emoji    string `json:"emoji"`    // Emoji temático para o Discord
	Duration string `json:"duration"` // Duração formatada (ex: 02:41)
	Path     string `json:"path"`     // Caminho completo ou relativo
}

// AmbientPreset representa uma trilha temática de D&D pronta para tocar
type AmbientPreset struct {
	Key         string // Identificador único (ex: taverna, combate)
	Name        string // Nome amigável (ex: 🍻 Taverna Medieval)
	Description string // Descrição temática
	URL         string // Stream direto de alta qualidade
	Emoji       string // Emoji do botão
}

// GetAmbientPresets retorna a lista de trilhas prontas para mesa de RPG
func GetAmbientPresets() []AmbientPreset {
	return []AmbientPreset{
		{
			Key:         "taverna",
			Name:        "🍻 Taverna Medieval",
			Description: "Música de alaúde e ambiente animado de taberna medieval.",
			URL:         "https://stream.zeno.fm/f3wvbbqmdg8uv", // Stream Medieval Fantasy Radio 24/7
			Emoji:       "🍻",
		},
		{
			Key:         "combate",
			Name:        "⚔️ Combate Épico",
			Description: "Trilha orquestral tensa e rápida para combates e encontros com chefes.",
			URL:         "https://stream.zeno.fm/87v2x8yvh7zuv", // Epic Battle Fantasy Stream
			Emoji:       "⚔️",
		},
		{
			Key:         "masmorra",
			Name:        "🕯️ Masmorra & Caverna",
			Description: "Sons atmosféricos sombrios, gotejamento e mistério para exploração.",
			URL:         "https://stream.zeno.fm/0r0xa792kwzuv", // Dark Ambient Dungeon Stream
			Emoji:       "🕯️",
		},
		{
			Key:         "floresta",
			Name:        "🌿 Floresta Feérica",
			Description: "Pássaros, brisa suave e harpa relaxante para viagens e descansos.",
			URL:         "https://stream.zeno.fm/6wz252u1x1zuv", // Nature & Fantasy Lofi Stream
			Emoji:       "🌿",
		},
	}
}

// FindPreset busca um preset por chave ou nome
func FindPreset(keyOrName string) (*AmbientPreset, bool) {
	presets := GetAmbientPresets()
	for _, p := range presets {
		if p.Key == keyOrName || p.Name == keyOrName {
			return &p, true
		}
	}
	return nil, false
}

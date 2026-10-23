package audio

import (
	"os"
	"path/filepath"
)

// AmbienceTrack define uma trilha de áudio ambiente para D&D e Baróvia
type AmbienceTrack struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Filename    string `json:"filename"`
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Duration    string `json:"duration"`
	Path        string `json:"path"`
}

// GetCurseOfStrahdAmbienceList retorna a lista completa de cenários de Curse of Strahd
func GetCurseOfStrahdAmbienceList() []AmbienceTrack {
	return []AmbienceTrack{
		{
			Key:         "svalich",
			Title:       "Floresta Svalich (Svalich Woods)",
			Filename:    "CoS - Svalich Woods.mp3",
			Emoji:       "🌲",
			Description: "Neblina densa, uivos distantes e árvores retorcidas de Baróvia",
			Category:    "Exploração & Regiões",
		},
		{
			Key:         "barovia_village",
			Title:       "Vila de Baróvia (Village of Barovia)",
			Filename:    "CoS - Village of Barovia.mp3",
			Emoji:       "🏚️",
			Description: "Ruas desertas, casas lacradas e corvos à sombra do castelo",
			Category:    "Povoados & Civilização",
		},
		{
			Key:         "vallaki",
			Title:       "Cidade de Vallaki (Town of Vallaki)",
			Filename:    "CoS - Town Of Vallaki.mp3",
			Emoji:       "🏘️",
			Description: "Festival do Sol Supremo, tensão velada e falsa alegria",
			Category:    "Povoados & Civilização",
		},
		{
			Key:         "krezk",
			Title:       "Vila de Krezk (Village of Krezk)",
			Filename:    "CoS - Village of Krezk.mp3",
			Emoji:       "🏡",
			Description: "Comunidade murada nas montanhas, isolada e cautelosa",
			Category:    "Povoados & Civilização",
		},
		{
			Key:         "castle_ravenloft",
			Title:       "Castelo Ravenloft (Castle Ravenloft)",
			Filename:    "CoS - Castle Ravenloft.mp3",
			Emoji:       "🏰",
			Description: "A morada sombria de Strahd von Zarovich, órgão e trovões",
			Category:    "Masmorras & Covis",
		},
		{
			Key:         "catacombs",
			Title:       "Catacumbas de Ravenloft (Catacombs)",
			Filename:    "CoS - Catacombs.mp3",
			Emoji:       "⚰️",
			Description: "Tumbas ancestrais, poeira milenar e perigo nas profundezas",
			Category:    "Masmorras & Covis",
		},
		{
			Key:         "death_house",
			Title:       "Casa da Morte (Death House)",
			Filename:    "CoS - Death House.mp3",
			Emoji:       "🏚️",
			Description: "Mansão vitoriana abandonada, tábuas rangendo e assombração",
			Category:    "Masmorras & Covis",
		},
		{
			Key:         "death_house_dungeon",
			Title:       "Masmorra da Casa da Morte (Dungeon)",
			Filename:    "CoS - Death House Dungeon.mp3",
			Emoji:       "💀",
			Description: "Culto dos Durst, cânticos no escuro e horrores subterrâneos",
			Category:    "Masmorras & Covis",
		},
		{
			Key:         "old_bonegrinder",
			Title:       "Velho Moedor de Ossos (Old Bonegrinder)",
			Filename:    "CoS - Old Bonegrinder.mp3",
			Emoji:       "🦴",
			Description: "Moinho de vento decadente, tortas de sonho e bruxas da noite",
			Category:    "Marcos Notáveis",
		},
		{
			Key:         "argynvostholt",
			Title:       "Argynvostholt (Mansão do Dragão)",
			Filename:    "CoS - Argynvostholt.mp3",
			Emoji:       "🛡️",
			Description: "Fortaleza da Ordem do Dragão de Prata e fantasmas vingativos",
			Category:    "Marcos Notáveis",
		},
		{
			Key:         "abbey",
			Title:       "Abadia de Santa Markovia (Abbey)",
			Filename:    "CoS - Abbey of St Markovia.mp3",
			Emoji:       "⛪",
			Description: "Gargalhadas maníacas, mutantes Belview e loucura sagrada",
			Category:    "Marcos Notáveis",
		},
		{
			Key:         "madhouse",
			Title:       "Asilo dos Loucos (Madhouse)",
			Filename:    "CoS - Madhouse.mp3",
			Emoji:       "🤪",
			Description: "Sussurros inquietantes, histeria e ecos perturbadores",
			Category:    "Marcos Notáveis",
		},
		{
			Key:         "tsolenka_pass",
			Title:       "Passo de Tsolenka (Tsolenka Pass)",
			Filename:    "CoS - Tsolenka Pass.mp3",
			Emoji:       "❄️",
			Description: "Ventos gélidos do Monte Baratok, ponte levadiça e abismo",
			Category:    "Exploração & Regiões",
		},
		{
			Key:         "van_richtens_tower",
			Title:       "Torre de Van Richten (Lake Baratok)",
			Filename:    "CoS - Van Richten's Tower.mp3",
			Emoji:       "🗼",
			Description: "Refúgio do caçador de monstros, relâmpagos e isolamento",
			Category:    "Marcos Notáveis",
		},
		{
			Key:         "ruins_of_berez",
			Title:       "Ruínas de Berez (Ruins of Berez)",
			Filename:    "CoS - Ruins of Berez.mp3",
			Emoji:       "🌊",
			Description: "Pântano alagado, névoa densa e a cabana da Baba Lysaga",
			Category:    "Exploração & Regiões",
		},
		{
			Key:         "wizard_of_wines",
			Title:       "Mago dos Vinhos (Wizard of the Wines)",
			Filename:    "CoS - Wizard Of The Wines.mp3",
			Emoji:       "🍇",
			Description: "Vinhedo Martikov, pragas de gravetos e druidas corrompidos",
			Category:    "Marcos Notáveis",
		},
		{
			Key:         "amber_temple",
			Title:       "Templo de Âmbar (Amber Temple)",
			Filename:    "CoS - Amber Temple.mp3",
			Emoji:       "🏛️",
			Description: "Vestígios sombrios, sarcófagos de âmbar e poder proibido",
			Category:    "Masmorras & Covis",
		},
		{
			Key:         "mists_of_ravenloft",
			Title:       "Névoas de Ravenloft (Mists of Ravenloft)",
			Filename:    "CoS - Mists of Ravenloft.mp3",
			Emoji:       "🌫️",
			Description: "As brumas sufocantes que cercam os domínios do pavor",
			Category:    "Exploração & Regiões",
		},
	}
}

// GetAmbienceDir retorna o caminho da pasta ambiente independente do diretório de execução
func GetAmbienceDir() string {
	return getAmbienceDir()
}

// getAmbienceDir localiza a pasta ambiente/ no diretório raiz do bot
func getAmbienceDir() string {
	if envDir := os.Getenv("AMBIENCE_DIR"); envDir != "" {
		if _, err := os.Stat(envDir); err == nil {
			return envDir
		}
	}
	candidates := []string{
		"ambiente",
		"../../ambiente",
		"../ambiente",
		"ambientes",
		"../../ambientes",
		"../ambientes",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "ambiente"
}

// GetAvailableAmbienceTracks retorna as faixas de ambiente presentes na pasta ambiente (ou musicas como fallback)
func GetAvailableAmbienceTracks() []AmbienceTrack {
	ambienceDir := getAmbienceDir()
	musicDir := getMusicDir()
	all := GetCurseOfStrahdAmbienceList()
	var available []AmbienceTrack

	for _, item := range all {
		fullPath := filepath.Join(ambienceDir, item.Filename)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() && info.Size() > 0 {
			item.Path = fullPath
			item.Duration = "Local"
			available = append(available, item)
			continue
		}

		// Fallback para musicas/ caso o arquivo ainda esteja lá
		fallbackPath := filepath.Join(musicDir, item.Filename)
		if info, err := os.Stat(fallbackPath); err == nil && !info.IsDir() && info.Size() > 0 {
			item.Path = fallbackPath
			item.Duration = "Local"
			available = append(available, item)
		}
	}

	return available
}


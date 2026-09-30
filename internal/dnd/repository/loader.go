package repository

import (
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"kenobi-bot/internal/dnd/model"
)

// LoadDNDData carrega e processa todo o acervo do D&D 2024 (5.5e)
func LoadDNDData(basePath string) (*MemoryDNDRepository, error) {
	slog.Info("Iniciando carregamento do acervo D&D 2024 (5.5e)...")

	classes := loadClasses()
	species := loadSpecies()
	spells := loadSpells(basePath)
	monsters := loadMonsters(basePath)
	items := loadItems(basePath)
	rules := loadRules(basePath)

	slog.Info("Dados do D&D 2024 carregados com sucesso!",
		"classes", len(classes),
		"species", len(species),
		"spells", len(spells),
		"monsters", len(monsters),
		"items", len(items),
		"rules", len(rules),
	)

	return NewMemoryRepository(classes, species, spells, monsters, items, rules), nil
}

func loadClasses() []model.Class {
	return []model.Class{
		{
			ID:     "artificer",
			Name:   "Artificer",
			NamePT: "Artífice",
			Emoji:  "🔧",
			Color:  0x4a6fa5,
			Overview: "Mestres da invenção e magia arcana aplicada, os Artífices infundem itens mundanos com poder mágico extraordinário, criam autômatos de combate e armaduras tecnológicas.",
			Role:     "Suporte / Utilidade / Conjurador Especialista / Criação de Itens Mágicos",
			Traits: model.ClassTraits{
				PrimaryAbility: "Inteligência (Intelligence)",
				HitDie:         "d8 por nível de Artífice",
				SavingThrows:   "Constituição e Inteligência",
				SkillProf:      "Escolha 2: Arcanismo, História, Investigação, Medicina, Natureza, Percepção, Prestidigitação",
				WeaponProf:     "Armas Simples e Armas de Fogo (se permitidas)",
				ArmorTraining:  "Armaduras Leves, Médias e Escudos",
				WeaponMastery:  "Nenhuma (substituída por Infusões)",
				StartingEquip:  "(A) 2 Adagas, Armadura de Couro Batido, Ferramentas de Ladrão, Pacote de Masmorra; ou (B) 100 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Magia com Ferramentas", Description: "Conjura magias usando ferramentas de artesão ou ladrão como foco mágico."},
				{Level: 2, Name: "Infundir Item (Infuse Item)", Description: "Infunde itens mundanos com propriedades mágicas (ex: Arma Aprimorada +1/+2, Defesa Aprimorada, Canhão Homúnculo)."},
				{Level: 3, Name: "A Espoleta Certa (Right Tool for the Job)", Description: "Cria qualquer conjunto de ferramentas de artesão magicamente em 1 hora."},
				{Level: 7, Name: "Flash de Genialidade", Description: "Usa reação para somar mod Inteligência em qualquer teste de atributo ou salvaguarda de aliado."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "alchemist", Name: "Alchemist [TCE / Tasha]", ClassName: "Artificer", UnlockLevel: 3,
					Synopsis: "Especialista em elixires experimentais que concedem cura, velocidade, voo e transformação.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Elixir Experimental", Description: "Cria elixires mágicos aleatórios ou escolhidos no descanso com efeitos de cura, voo, alter self e bônus de CA."},
					},
				},
				{
					ID: "armorer", Name: "Armorer [TCE / Tasha]", ClassName: "Artificer", UnlockLevel: 3,
					Synopsis: "Transforma sua armadura pesada em um exoesqueleto de combate místico (Guardião ou Infiltrador).",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Armadura de Poder (Arcane Armor)", Description: "A armadura cobre todo o corpo, substitui membros perdidos e permite alternar entre Guardião (manoplas de trovão e PV temporário) ou Infiltrador (raio elétrico e vantagem em furtividade)."},
					},
				},
				{
					ID: "artillerist", Name: "Artillerist [TCE / Tasha]", ClassName: "Artificer", UnlockLevel: 3,
					Synopsis: "Mestre da balística mágica e canhões de energia ambulantes que disparam fogo, força ou escudos de cura.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Canhão Ancestral (Eldritch Cannon)", Description: "Cria um canhão mágico portátil ou com pernas que dispara Balistas de Força, Lança-chamas ou concede campo de PVs temporários a cada rodada."},
					},
				},
				{
					ID: "battle_smith", Name: "Battle Smith [TCE / Tasha]", ClassName: "Artificer", UnlockLevel: 3,
					Synopsis: "Combatente marcial que ataca usando Inteligência e luta ao lado de um Defensor de Aço leal.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Prontidão para Batalha", Description: "Ganha proficiência com armas marciais e ataca/causa dano usando Inteligência em vez de Força/Destreza."},
						{Level: 3, Name: "Defensor de Aço (Steel Defender)", Description: "Um autômato mecânico quadrúpede ou bípede que ataca e impõe desvantagem em ataques contra você."},
					},
				},
				{
					ID: "reanimator", Name: "Reanimator [Ravenloft / The Horrors Within]", ClassName: "Artificer", UnlockLevel: 3,
					Synopsis: "Médico sombrio e cientista gótico de Lamordia (estilo Dr. Frankenstein) que costura servos de cadáveres e revive aliados com choques elétricos.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Companheiro Reanimado (Reanimated Companion)", Description: "Cria um monstro costurado com PVs (5 + 5x nível), CA 10+mod INT, resistências a Necrótico/Veneno e Explosão Cadavérica."},
						{Level: 3, Name: "Choque Vital (Jolt to Life)", Description: "Ao conjurar Poupar os Moribundos (Spare the Dying), cura PVs iguais ao nível de Artífice e causa 2d4 a 4d4 de dano elétrico em inimigos ao redor."},
						{Level: 3, Name: "Magias de Reanimador", Description: "Prepara Falsa Vida, Raio da Bruxa, Cegueira/Surdez, Animar Mortos, Raio Relâmpago, Barreira Antivida e Reviver os Mortos."},
					},
				},
			},
		},
		{
			ID:     "barbarian",
			Name:   "Barbarian",
			NamePT: "Bárbaro",
			Emoji:  "🪓",
			Color:  0xe76f51,
			Overview: "Um guerreiro feroz impulsionado pela Fúria primal, combinando grande vigor físico, resistência colossal e ataques devastadores em combate corpo a corpo.",
			Role:     "Tanque / Dano Físico Corpo a Corpo / Controle Bruto",
			Traits: model.ClassTraits{
				PrimaryAbility: "Força (Strength)",
				HitDie:         "d12 por nível de Bárbaro",
				SavingThrows:   "Força e Constituição",
				SkillProf:      "Escolha 2: Adestrar Animais, Atletismo, Intimidação, Natureza, Percepção, Sobrevivência",
				WeaponProf:     "Armas Simples e Marciais",
				ArmorTraining:  "Armaduras Leves, Médias e Escudos",
				WeaponMastery:  "2 tipos de armas (aumenta para 3 no nv 4 e 4 no nv 10)",
				StartingEquip:  "(A) Machado Grande, 4 Machadinhas, Pacote de Explorador e 15 PO; ou (B) 75 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Fúria (Rage)", Description: "Entra em estado primal ganhando dano extra de fúria, vantagem em testes/saves de Força, resistência a dano Concussão/Perfurante/Cortante e manutenção de fúria com Ação Bônus."},
				{Level: 1, Name: "Defesa Sem Armadura", Description: "Quando não estiver usando armadura, sua CA é 10 + mod DES + mod CON + bônus de escudo."},
				{Level: 1, Name: "Maestria em Armas (Weapon Mastery)", Description: "Desbloqueia as propriedades de maestria de armas escolhidas (ex: Cleave, Topple, Graze)."},
				{Level: 2, Name: "Sentido de Perigo (Danger Sense)", Description: "Vantagem em salvaguardas de Destreza contra efeitos que você possa ver."},
				{Level: 2, Name: "Ataque Descuidado (Reckless Attack)", Description: "Ganha vantagem nos ataques de Força no turno, mas concede vantagem aos ataques inimigos contra você."},
				{Level: 3, Name: "Subclasse de Bárbaro", Description: "Escolha uma trilha de subclasse."},
				{Level: 5, Name: "Ataque Extra & Movimento Rápido", Description: "Ataca 2 vezes por Ação de Ataque e ganha +3m de deslocamento sem armadura pesada."},
				{Level: 9, Name: "Golpe Brutal (Brutal Strike)", Description: "Pode abrir mão de vantagem para aplicar efeitos debilitantes (Forceful Blow ou Hamstring Blow) e causar +1d10 de dano extra."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "berserker", Name: "Path of the Berserker [PHB'24]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Mergulha na violência desenfreada do combate, desferindo golpes extras e ignorando o medo e o charme.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Frenesi (Frenzy)", Description: "Ao usar Ataque Descuidado em Fúria, desfere dano extra igual a dados de d6 adicionais."},
						{Level: 6, Name: "Fúria Descuidada (Mindless Rage)", Description: "Fica imune às condições Enfeitiçado (Charmed) e Amedrontado (Frightened) durante a Fúria."},
					},
				},
				{
					ID: "wild_heart", Name: "Path of the Wild Heart [PHB'24]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Canaliza a essência totêmica e os espíritos dos animais (Urso, Águia, Lobo) para conceder resistências e suporte tático.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Fúria Selvagem (Rage of the Wilds)", Description: "Escolhe um aspecto animal na fúria: Urso (resistência a quase todos danos), Águia (Disparada bônus) ou Lobo (vantagem para aliados)."},
					},
				},
				{
					ID: "world_tree", Name: "Path of the World Tree [PHB'24]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Nova subclasse 2024! Sintoniza com a mística Árvore do Mundo (Yggdrasil), canalizando vitalidade regenerativa e raízes dimensionais.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Vitalidade da Árvore (Vitality of the Tree)", Description: "Concede Pontos de Vida Temporários contínuos a si mesmo e a um aliado próximo a cada rodada."},
						{Level: 6, Name: "Galhos da Árvore (Branches of the Tree)", Description: "Puxa inimigos distantes até 9 metros para perto de si com galhos espectrais."},
					},
				},
				{
					ID: "zealot", Name: "Path of the Zealot [PHB'24]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Um guerreiro sagrado alimentado pelo poder divino de sua divindade, causando dano radiante/necrótico e desafiando a própria morte.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Fúria Divina (Divine Fury)", Description: "O primeiro ataque do turno causa dano radiante ou necrótico adicional (1d6 + metade do nível)."},
						{Level: 6, Name: "Foco Fanático (Fanatical Focus)", Description: "Rola novamente um teste de resistência falho durante a Fúria."},
					},
				},
				{
					ID: "beast", Name: "Path of the Beast [TCE / Tasha]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Manifesta garras, presas com mordida regenerativa ou cauda espinhosa defensiva ao entrar em fúria.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Forma da Fera", Description: "Escolhe Mordida (cura PVs), Garras (ataque extra gratuito) ou Cauda (+1d8 na CA com reação)."},
					},
				},
				{
					ID: "giant", Name: "Path of the Giant [BGG / Glória dos Gigantes]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Cresce para tamanho Grande/Enorme na fúria, arremessa armas com retorno e lança inimigos pelo ar.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Poder dos Gigantes", Description: "Torna-se Grande (Large), ganha alcance estendido de +1,5m e arremessa qualquer arma infundida de energia."},
					},
				},
				{
					ID: "ancestral_guardian", Name: "Path of the Ancestral Guardian [XGE / Xanathar]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Invoca espíritos de ancestrais para proteger aliados e punir qualquer inimigo que ataque seus companheiros.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Protetores Ancestrais", Description: "O primeiro inimigo atingido tem desvantagem ao atacar aliados e causa metade do dano."},
					},
				},
				{
					ID: "wild_magic_barb", Name: "Path of Wild Magic [TCE / Tasha]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Canaliza a magia caótica da Agrestia das Fadas em surtos mágicos aleatórios toda vez que entra em fúria.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Surto de Magia Selvagem", Description: "Rola na tabela de magia selvagem para teletransporte, rajadas de luz, tentáculos e auras protetoras."},
					},
				},
				{
					ID: "storm_herald", Name: "Path of the Storm Herald [XGE / Xanathar]", ClassName: "Barbarian", UnlockLevel: 3,
					Synopsis: "Envolto em tempestades elementais da Tundra, Deserto ou Mar que eletrocutam ou queimam inimigos adjacentes.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Aura da Tempestade", Description: "Emite uma aura de 3m de fogo contínuo, choque elétrico ou PV temporário gélido a cada turno."},
					},
				},
			},
		},
		{
			ID:     "bard",
			Name:   "Bard",
			NamePT: "Bardo",
			Emoji:  "🎭",
			Color:  0xf4a261,
			Overview: "Mestre da oratória, música e feitiçaria, o Bardo tece a canção da criação para inspirar aliados, desmoralizar adversários e conjurar magias versáteis.",
			Role:     "Suporte / Controle / Especialista em Perícias / Conjurador Versátil",
			Traits: model.ClassTraits{
				PrimaryAbility: "Carisma (Charisma)",
				HitDie:         "d8 por nível de Bardo",
				SavingThrows:   "Destreza e Carisma",
				SkillProf:      "Escolha quaisquer 3 perícias",
				WeaponProf:     "Armas Simples",
				ArmorTraining:  "Armaduras Leves",
				WeaponMastery:  "Nenhuma (a menos que escolha College of Valor)",
				StartingEquip:  "(A) Adaga, Pacote de Artista, Alaúde e 19 PO; ou (B) 90 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Inspiração de Bardo (Bardic Inspiration)", Description: "Usa Ação Bônus para conceder um dado de inspiração (d6 até d12) para aliados somarem em testes, ataques ou saves."},
				{Level: 1, Name: "Conjuração de Magias", Description: "Conjura magias preparadas da lista de Bardo usando Carisma."},
				{Level: 2, Name: "Pau pra Toda Obra (Jack of All Trades)", Description: "Soma metade do bônus de proficiência em qualquer teste de atributo sem proficiência."},
				{Level: 2, Name: "Especialização (Expertise)", Description: "Dobra o bônus de proficiência em duas perícias à sua escolha."},
				{Level: 10, Name: "Segredos Mágicos (Magical Secrets)", Description: "No D&D 2024, pode preparar magias diretamente das listas de Clérigo, Druida ou Mago!"},
			},
			Subclasses: []model.Subclass{
				{
					ID: "dance", Name: "College of Dance [PHB'24]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Nova subclasse 2024! Movimento corporal em arte letal desarmada, esquivando de golpes e compartilhando iniciativa.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Passos Deslumbrantes (Dazzling Footwork)", Description: "CA baseada em Carisma sem armadura e ataques desarmados usando Carisma e dados de Inspiração."},
					},
				},
				{
					ID: "glamour", Name: "College of Glamour [PHB'24]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Envolto na beleza hipnótica de Faerie, comanda a atenção de multidões e concede vigor feérico a aliados.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Manto da Inspiração (Mantle of Inspiration)", Description: "Concede PVs temporários a múltiplos aliados e permite reposicionamento tático imediato."},
					},
				},
				{
					ID: "lore", Name: "College of Lore [PHB'24]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Erudito e místico perspicaz que descobre segredos arcanos mais cedo e confunde inimigos com palavras cortantes.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Palavras Cortantes (Cutting Words)", Description: "Usa reação para subtrair o dado de inspiração do ataque, dano ou perícia do inimigo."},
					},
				},
				{
					ID: "valor", Name: "College of Valor [PHB'24]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Menestrel de guerra treinado em armas marciais e armaduras médias que inspira danos massivos na linha de frente.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Inspiração de Combate", Description: "Aliados somam o dado de inspiração no dano das armas ou na própria CA contra ataques."},
					},
				},
				{
					ID: "swords", Name: "College of Swords [XGE / Xanathar]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Duelista acrobático e mestre da lâmina que realiza floreios de combate para aumentar dano, CA e mobilidade.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Floreios da Lâmina (Blade Flourish)", Description: "Gasta dados de inspiração para aplicar Floreio Defensivo (+CA), Cortante (dano em área) ou Móvel (empurrão e teletransporte)."},
					},
				},
				{
					ID: "eloquence", Name: "College of Eloquence [TCE / Tasha]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "O maior mestre da persuasão e oratória do multiverso, impossível de falhar em testes de lábia e que reduz saves de inimigos.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Língua Prateada", Description: "Qualquer resultado menor que 10 no d20 em Persuasão e Enganação vira automaticamente um 10."},
						{Level: 3, Name: "Palavras Desestabilizadoras", Description: "Subtrai o dado de inspiração da próxima salvaguarda do inimigo."},
					},
				},
				{
					ID: "creation", Name: "College of Creation [TCE / Tasha]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Canaliza a Canção da Criação cósmica para animar objetos inanimados e materializar itens do nada.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Performance da Criação", Description: "Cria qualquer item não-mágico sólido do nada e infunde notas mágicas dançantes."},
					},
				},
				{
					ID: "whispers", Name: "College of Whispers [XGE / Xanathar]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Espião e mestre dos segredos sombrios, roubando sombras de pessoas mortas para personificá-las e desferindo veneno psíquico.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Lâminas Psíquicas", Description: "Causa até 8d6 de dano psíquico extra com arma usando dados de inspiração."},
					},
				},
				{
					ID: "spirits", Name: "College of Spirits [VRGtR / Ravenloft]", ClassName: "Bard", UnlockLevel: 3,
					Synopsis: "Médium e oráculo de Ravenloft que canaliza espíritos de além-túmulo através de cartas de tarô e relatos fúnebres para liberar bênçãos ou calamidades.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Canalizador & Foco Espiritual (Channeler)", Description: "Aprende o truque Orientação (Guidance) com alcance aumentado de 18 metros. Usa cartas de tarô/jogos, velas ou penas de tinta como foco de conjuração."},
						{Level: 3, Name: "Espíritos do Além (Spirits from Beyond)", Description: "Ao conceder Inspiração de Bardo ou com Ação Bônus, canaliza espíritos (Beloved cura, Sharpshooter dano de força, Avenger retribuição, Renegade teletransporte, Fortune Teller vantagem em d20, Wayfarer velocidade, Trickster charme psíquico, Shade invisibilidade necrótica, Arsonist fogo, Coward pavor, Brute trovão, Priest cura e remoção de condições) e libera com Ação Mágica em até 9m."},
						{Level: 6, Name: "Canalização Aprimorada (Empowered Channeling)", Description: "Soma +1d6 no dano ou cura de magias de bardo. Sempre tem Guardiões Espirituais (Spirit Guardians) preparado e pode conjurar 1x grátis por descanso longo; a aura concede Meia Cobertura aos aliados."},
						{Level: 14, Name: "Conexão Mística (Mystical Connection)", Description: "Ao rolar na tabela Espíritos do Além, rola 2 dados e escolhe o efeito; se der números iguais, escolhe qualquer espírito da tabela!"},
					},
				},
			},
		},
		{
			ID:     "cleric",
			Name:   "Cleric",
			NamePT: "Clérigo",
			Emoji:  "✨",
			Color:  0xf3c68f,
			Overview: "Um condutor do poder divino dos deuses, o Clérigo canaliza magia sagrada tanto para proteger, curar e reviver aliados quanto para banir mortos-vivos e punir o mal.",
			Role:     "Cura / Suporte Divino / Tanque / Dano Radiante e Necrótico",
			Traits: model.ClassTraits{
				PrimaryAbility: "Sabedoria (Wisdom)",
				HitDie:         "d8 por nível de Clérigo",
				SavingThrows:   "Sabedoria e Carisma",
				SkillProf:      "Escolha 2: História, Intuição, Medicina, Persuasão, Religião",
				WeaponProf:     "Armas Simples (e Marciais se Proteção)",
				ArmorTraining:  "Armaduras Leves, Médias e Escudos (Pesadas se Proteção)",
				WeaponMastery:  "Nenhuma",
				StartingEquip:  "(A) Maça, Cota de Escamas, Escudo, Símbolo Sagrado, Pacote de Sacerdote e 7 PO; ou (B) 110 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Ordem Sagrada (Divine Order)", Description: "Escolhe Proteção (Armaduras Pesadas e Armas Marciais) ou Taumaturgo (Truque extra e bônus em Religião/Arcanismo)."},
				{Level: 1, Name: "Conjuração Divina", Description: "Conjura e prepara magias diárias da lista de Clérigo com Sabedoria."},
				{Level: 2, Name: "Canalizar Divindade (Channel Divinity)", Description: "Canaliza poder divino para Expulsar Mortos-Vivos (Turn Undead) ou efeitos do Domínio."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "life", Name: "Life Domain [PHB'24]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "O ápice dos curandeiros, multiplicando a eficácia de curas e sustentando a vida do grupo sob qualquer adversidade.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Discípulo da Vida", Description: "Curas aumentadas em 2 + nível da magia."},
						{Level: 3, Name: "Preservar a Vida", Description: "Cura em área até 5x nível de clérigo para aliados feridos."},
					},
				},
				{
					ID: "light", Name: "Light Domain [PHB'24]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Serve às divindades da luz e da verdade, projetando labaredas ofuscantes e bolas de fogo devoradoras.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Labareda Protetora", Description: "Impõe desvantagem em ataques inimigos contra você ou aliados."},
						{Level: 3, Name: "Radiância do Alvorecer", Description: "Explosão radiante de 9m causando dano massivo e dissipando escuridão mágica."},
					},
				},
				{
					ID: "trickery", Name: "Trickery Domain [PHB'24]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Mestre do engano e ilusões divinas, confundindo adversários com duplicatas ilusórias perfeitas.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Invocar Duplicata (Invoke Duplicity)", Description: "Cria cópia ilusória móvel que permite conjurar através dela e concede vantagem em ataques."},
					},
				},
				{
					ID: "war", Name: "War Domain [PHB'24]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Campeão militar das divindades da batalha, empunhando martelos e espadas com fúria sagrada.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Sacerdote da Guerra", Description: "Ataques adicionais com Ação Bônus."},
						{Level: 3, Name: "Golpe Guiado", Description: "+10 de bônus no ataque com Canalizar Divindade."},
					},
				},
				{
					ID: "twilight", Name: "Twilight Domain [TCE / Tasha]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Guardião do crepúsculo com visão no escuro de 90m, voo noturno e uma aura de santuário que concede PVs temporários massivos a cada rodada.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Santuário do Crepúsculo (Twilight Sanctuary)", Description: "Cria uma esfera de 9m que a cada final de turno concede 1d6 + nível de clérigo em PVs temporários a todos os aliados ou encerra efeitos de medo/charme!"},
					},
				},
				{
					ID: "peace", Name: "Peace Domain [TCE / Tasha]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Canaliza a harmonia e o vínculo da paz, conectando a alma dos aliados para somarem +1d4 em todos os ataques e saves.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Vínculo Encorajador (Emboldening Bond)", Description: "Vincula aliados para somarem 1d4 adicional em jogadas de ataque, testes e saves, acumulável com a magia Bênção!"},
					},
				},
				{
					ID: "forge", Name: "Forge Domain [XGE / Xanathar]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Devoto de deuses ferreiros com armadura impenetrável, resistência a fogo e criação de armas mágicas na bigorna divina.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Bênção da Forja", Description: "Transforma armadura ou arma comum em item mágico +1 permanente a cada descanso."},
					},
				},
				{
					ID: "grave", Name: "Grave Domain [XGE / Xanathar]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Vigia a fronteira entre vida e morte, anulando acertos críticos de inimigos e impondo vulnerabilidade a danos para combos devastadores.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Círculo da Mortalidade", Description: "Cura máxima em criaturas caídas com 0 PVs."},
						{Level: 3, Name: "Caminho para o Túmulo", Description: "Torna o inimigo VULNERÁVEL a todo o dano do próximo ataque aliado!"},
					},
				},
				{
					ID: "tempest", Name: "Tempest Domain [PHB]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Comanda o trovão e os relâmpagos dos céus, causando dano elétrico MÁXIMO sem rolar dados.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Ira Destrutiva (Destructive Wrath)", Description: "Ao rolar dano elétrico ou de trovão, causa automaticamente o dano MÁXIMO dos dados."},
					},
				},
				{
					ID: "death", Name: "Death Domain [DMG]", ClassName: "Cleric", UnlockLevel: 3,
					Synopsis: "Conjurador de magia necromântica letal que duplica truques necróticos e canaliza toque da morte destruidor.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Toque da Morte", Description: "Causa 5 + 2x nível de clérigo em dano necrótico extra ao acertar ataque corpo a corpo."},
					},
				},
			},
		},
		{
			ID:     "druid",
			Name:   "Druid",
			NamePT: "Druida",
			Emoji:  "🌿",
			Color:  0x2a9d8f,
			Overview: "Guardião da natureza e do equilíbrio elemental, o Druida assume formas animais selvagens (Wild Shape) e manipula as forças da terra, mar e céu.",
			Role:     "Transformação em Feras / Controle de Campo / Cura / Dano Elemental",
			Traits: model.ClassTraits{
				PrimaryAbility: "Sabedoria (Wisdom)",
				HitDie:         "d8 por nível de Druida",
				SavingThrows:   "Inteligência e Sabedoria",
				SkillProf:      "Escolha 2: Arcana, Adestrar Animais, Intuição, Medicina, Natureza, Percepção, Religião, Sobrevivência",
				WeaponProf:     "Armas Simples",
				ArmorTraining:  "Armaduras Leves, Médias e Escudos (sem restrição de metal no 2024!)",
				WeaponMastery:  "Nenhuma",
				StartingEquip:  "(A) Foice, Escudo, Foco Druídico, Pacote de Explorador e 9 PO; ou (B) 50 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Ordem Primal (Primal Order)", Description: "Escolhe Magician (truque extra) ou Warden (armaduras médias e armas marciais)."},
				{Level: 2, Name: "Forma Selvagem (Wild Shape)", Description: "Assume a forma de feras para combate, infiltração e exploração."},
				{Level: 2, Name: "Companheiro Selvagem", Description: "Gasta Forma Selvagem para conjurar Encontrar Familiar sem componentes."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "circle_land", Name: "Circle of the Land [PHB'24]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Sintonizado com biomas (Árido, Polar, Temperado, Tropical), recuperando magias no descanso curto.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Ajuda da Terra (Land's Aid)", Description: "Invoca raízes que curam aliados e causam dano necrótico a inimigos."},
					},
				},
				{
					ID: "circle_moon", Name: "Circle of the Moon [PHB'24]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Especialista supremo em combate com Forma Selvagem, ganhando formas de feras com armadura e ataques radiantes.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Forma Selvagem de Combate", Description: "Ganha PVs temporários massivos e CA 13 + mod SAB na forma de fera."},
					},
				},
				{
					ID: "circle_sea", Name: "Circle of the Sea [PHB'24]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Nova subclasse 2024! Empunha a tempestade marítima, ondas furiosas e ventos oceânicos protetores.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Fúria do Mar (Wrath of the Sea)", Description: "Aura de 3m que empurra inimigos e causa dano de trovão/frio a cada rodada."},
					},
				},
				{
					ID: "circle_stars", Name: "Circle of the Stars [PHB'24]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Canaliza corpos celestes através de um mapa estelar, assumindo formas de Arqueiro, Cálice ou Dragão.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Forma Estelar (Starry Form)", Description: "Assume constelações de Arqueiro (ataques radiantes), Cálice (cura extra) ou Dragão (concentração infalível)."},
					},
				},
				{
					ID: "circle_spores", Name: "Circle of Spores [TCE / Tasha]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Druida fúngico que manipula esporos necromânticos para criar zumbis fúngicos, ganhar PV temporário e causar dano por contato.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Entidade Simbiótica (Symbiotic Entity)", Description: "Desperta esporos ganhando 4x nível em PVs temporários e dano necrótico adicional em cada ataque."},
					},
				},
				{
					ID: "circle_wildfire", Name: "Circle of Wildfire [TCE / Tasha]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Invoca um Espírito do Fogo Primal leal que incendeia inimigos e teletransporta aliados pelo campo de batalha.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Invocar Espírito Selvagem (Wildfire Spirit)", Description: "Um espírito flamejante que ataca à distância e teletransporta você e aliados até 4,5m com explosão de fogo."},
					},
				},
				{
					ID: "circle_shepherd", Name: "Circle of the Shepherd [XGE / Xanathar]", ClassName: "Druid", UnlockLevel: 3,
					Synopsis: "Comandante supremo de invocações e protetor de animais, com totens espirituais que curam todo o grupo.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Totem Espiritual", Description: "Invoca espírito de Urso (PVs temporários massivos), Falcão (vantagem em ataques) ou Unicórnio (cura em grupo toda vez que conjurar)."},
					},
				},
			},
		},
		{
			ID:     "fighter",
			Name:   "Fighter",
			NamePT: "Guerreiro",
			Emoji:  "🛡️",
			Color:  0x457b9d,
			Overview: "O ápice do combate com armas e tática militar. O Guerreiro domina todas as armas e armaduras, realiza múltiplos ataques por turno e adapta seu estilo com maestria incomparável.",
			Role:     "Dano Físico Máximo / Tanque Marcial / Controle de Campo de Batalha",
			Traits: model.ClassTraits{
				PrimaryAbility: "Força ou Destreza",
				HitDie:         "d10 por nível de Guerreiro",
				SavingThrows:   "Força e Constituição",
				SkillProf:      "Escolha 2: Acrobacia, Adestrar Animais, Atletismo, História, Intuição, Intimidação, Percepção, Sobrevivência",
				WeaponProf:     "Todas as Armas Simples e Marciais",
				ArmorTraining:  "Todas as Armaduras (Leves, Médias, Pesadas) e Escudos",
				WeaponMastery:  "3 tipos de armas no 1º nível (até 6 no nv 16)",
				StartingEquip:  "(A) Cota de Malha, Espada Grande ou Arco Longo, 2 Machadinhas, Pacote de Masmorra e 9 PO; ou (B) 155 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Estilo de Luta (Fighting Style)", Description: "Ganha um talento de Estilo de Luta (Arquearia, Defesa, Duelo, Grandes Armas, Duas Armas)."},
				{Level: 1, Name: "Retomar o Fôlego (Second Wind)", Description: "Recupera 1d10 + nível de PVs como Ação Bônus."},
				{Level: 1, Name: "Maestria em Armas (3 armas)", Description: "Aplica efeitos especiais de armas (Topple, Vex, Cleave, Graze, Push, Slow, Sap)."},
				{Level: 2, Name: "Surto de Ação (Action Surge)", Description: "Executa uma Ação adicional completa no seu turno."},
				{Level: 2, Name: "Mente Tática (Tactical Mind)", Description: "Gasta uso de Second Wind para somar +1d10 em teste de perícia falho."},
				{Level: 5, Name: "Ataque Extra (até 4x no nv 20)", Description: "Ataca múltiplas vezes por turno."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "battle_master", Name: "Battle Master [PHB'24]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Estrategista de elite que emprega Manobras e Dados de Superioridade para desarmar, derrubar e liderar.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Superioridade em Combate", Description: "Ganha dados de d8 a d12 para manobras como Trip Attack, Riposte, Menacing Attack."},
					},
				},
				{
					ID: "champion", Name: "Champion [PHB'24]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Máquina de combate atlética com margem de acerto crítico expandida (19-20 e 18-20) e sobrevivência lendária.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Crítico Aprimorado", Description: "Acerto Crítico em 19 e 20 nos dados (18-20 no nv 15)."},
					},
				},
				{
					ID: "eldritch_knight", Name: "Eldritch Knight [PHB'24]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Funde a esgrima com a magia de Mago, vinculando armas para retorno instantâneo e mesclando feitiços com ataques.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Vínculo com Arma", Description: "Invoca a arma para a mão instantaneamente."},
						{Level: 7, Name: "Magia de Guerra", Description: "Substitui um ataque por qualquer truque ou magia de nível 1-2!"},
					},
				},
				{
					ID: "psi_warrior", Name: "Psi Warrior [PHB'24]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Desperta poderes psíquicos da mente para criar campos de força de proteção, saltos aumentados e golpes telecinéticos.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Poder Psiônico", Description: "Dados de energia psiônica para dano de força extra, escudos protetores ou mover objetos."},
					},
				},
				{
					ID: "echo_knight", Name: "Echo Knight [EGW / Wildemount]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Manifesta ecos temporais de si mesmo em outras linhas do tempo para atacar à distância, trocar de lugar e realizar emboscadas.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Manifestar Eco (Manifest Echo)", Description: "Cria um clone espectral que se move, ataca pelo seu alcance e permite teletransporte instantâneo de troca de posição!"},
					},
				},
				{
					ID: "rune_knight", Name: "Rune Knight [TCE / Tasha]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Aprende a forja rúnica ancestral dos gigantes, entalhando runas em armaduras para crescer para tamanho Gigante e prender inimigos em correntes de fogo.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Estatura dos Gigantes", Description: "Cresce para tamanho Grande (Large), causando dano extra em cada ataque."},
						{Level: 3, Name: "Magia das Runas", Description: "Ativa Runas da Nuvem (redireciona ataques), Fogo (correntes flamejantes) ou Pedra."},
					},
				},
				{
					ID: "samurai", Name: "Samurai [XGE / Xanathar]", ClassName: "Fighter", UnlockLevel: 3,
					Synopsis: "Guerreiro com foco inabalável que concede Vantagem em todos os ataques e recusa a morte temporariamente.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Espírito de Luta (Fighting Spirit)", Description: "Ganha Vantagem em todas as jogadas de ataque no turno e PVs temporários instantâneos."},
					},
				},
			},
		},
		{
			ID:     "monk",
			Name:   "Monk",
			NamePT: "Monge",
			Emoji:  "🥋",
			Color:  0xe9c46a,
			Overview: "Mestre da disciplina física e espiritual, o Monge canaliza o Foco (Ki) para desferir rajadas rápidas de golpes desarmados, correr pelas paredes e desviar de projéteis.",
			Role:     "Combatente Veloz / Atacante Desarmado / Controle e Mobilidade Extrema",
			Traits: model.ClassTraits{
				PrimaryAbility: "Destreza e Sabedoria",
				HitDie:         "d8 por nível de Monge",
				SavingThrows:   "Força e Destreza (todas no nv 14)",
				SkillProf:      "Escolha 2: Acrobacia, Atletismo, História, Intuição, Religião, Furtividade",
				WeaponProf:     "Armas Simples e Armas Marciais com Finesse/Leve",
				ArmorTraining:  "Nenhuma",
				WeaponMastery:  "Nenhuma (substituída por Artes Marciais)",
				StartingEquip:  "(A) Lança, 5 Dardos, Pacote de Explorador e 11 PO; ou (B) 65 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Artes Marciais (Martial Arts)", Description: "Dano desarmado d6 até d12, uso de Destreza e Ataque Desarmado com Ação Bônus grátis."},
				{Level: 1, Name: "Defesa Sem Armadura", Description: "CA é 10 + mod DES + mod SAB sem armadura nem escudo."},
				{Level: 2, Name: "Pontos de Foco (Monk's Focus / Ki)", Description: "Gasta Foco em Rajada de Golpes (Flurry of Blows), Defesa Paciente ou Passo do Vento."},
				{Level: 2, Name: "Desviar Ataques (Deflect Attacks)", Description: "Desvia de projéteis e ataques corpo a corpo e redireciona dano!"},
				{Level: 5, Name: "Golpe Atordoante (Stunning Strike)", Description: "Atinge pontos vitais para atordoar a criatura."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "open_hand", Name: "Warrior of the Open Hand [PHB'24]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Mestre supremo das artes marciais puras que empurra, derruba e destrói inimigos com a lendária Palma Trêmula.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Técnica da Palma Aberta", Description: "Seus golpes de Rajada derrubam inimigos propensos, empurram 4,5m ou anulam reações."},
					},
				},
				{
					ID: "shadow", Name: "Warrior of Shadow [PHB'24]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Ninja místico que se funde com as trevas, conjura escuridão mágica que ele próprio enxerga e salta entre sombras.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Passo das Sombras", Description: "Teletransporta até 18m de uma sombra para outra como Ação Bônus."},
					},
				},
				{
					ID: "elements", Name: "Warrior of the Elements [PHB'24]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Reformulado em 2024! Envolve seus membros em fogo, gelo, trovão e ácido com alcance de combate ampliado.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Sintonização Elemental", Description: "Ganha +3m de alcance nos socos e causa dano de Fogo, Frio, Ácido ou Trovão."},
					},
				},
				{
					ID: "mercy", Name: "Warrior of Mercy [PHB'24]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Curandeiro e flagelo místico que manipula a energia vital para curar ferimentos ou aplicar toques necróticos letais.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Mãos da Cura & Mãos do Dano", Description: "Cura aliados gastando Foco ou infunde dano necrótico e envenenamento nos socos."},
					},
				},
				{
					ID: "astral_self", Name: "Way of the Astral Self [TCE / Tasha]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Projeta braços espectrais e uma forma astral translúcida para golpear com Sabedoria e alcance aumentado.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Braços do Eu Astral", Description: "Invoca braços astrais atacando a 3m com mod Sabedoria e dano de Força."},
					},
				},
				{
					ID: "ascendant_dragon", Name: "Way of the Ascendant Dragon [FTD / Fizban]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Canaliza a presença dracônica, soprando labaredas elementais e brotando asas de dragão temporárias.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Sopro do Dragão", Description: "Substitui ataque desarmado por sopro de Fogo, Gelo, Ácido, Veneno ou Eletricidade."},
					},
				},
				{
					ID: "kensei", Name: "Way of the Kensei [XGE / Xanathar]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Mestre espadachim e arqueiro zen que transforma armas marciais em armas de monge com dano aprimorado.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Armas do Kensei", Description: "Domina armas com bônus de CA defensiva e dano de ki."},
					},
				},
				{
					ID: "drunken_master", Name: "Way of the Drunken Master [XGE / Xanathar]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Lutador imprevisível e acrobático que confunde inimigos com tropeços enganosos e contra-ataques.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Ginga Embriagada", Description: "Redireciona ataques inimigos errados para outros alvos adjacentes!"},
					},
				},
				{
					ID: "sun_soul", Name: "Way of the Sun Soul [XGE / Xanathar]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Canaliza a energia luminosa da alma em rajadas solares de pura luz radiante à distância.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Raio Solar Radiante", Description: "Dispara projéteis de luz radiante a até 9m usando dados de artes marciais."},
					},
				},
				{
					ID: "long_death", Name: "Way of the Long Death [SCAG]", ClassName: "Monk", UnlockLevel: 3,
					Synopsis: "Monge necromântico que rouba a força vital de criaturas que morrem ao seu redor.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Toque da Morte", Description: "Ganha PVs temporários ao reduzir criaturas a 0 PVs."},
					},
				},
			},
		},
		{
			ID:     "paladin",
			Name:   "Paladin",
			NamePT: "Paladino",
			Emoji:  "⚔️",
			Color:  0x9b5de5,
			Overview: "Cavaleiro sagrado vinculado a um Juramento solene. O Paladino combina combate pesado, cura pelas mãos (Lay on Hands), destruição de demônios e auras protetoras para o grupo.",
			Role:     "Tanque Sagrado / Dano Radiante Explosivo (Smite) / Auras de Suporte",
			Traits: model.ClassTraits{
				PrimaryAbility: "Força e Carisma",
				HitDie:         "d10 por nível de Paladino",
				SavingThrows:   "Sabedoria e Carisma",
				SkillProf:      "Escolha 2: Atletismo, Intuição, Intimidação, Medicina, Persuasão, Religião",
				WeaponProf:     "Todas as Armas Simples e Marciais",
				ArmorTraining:  "Todas as Armaduras (Leves, Médias, Pesadas) e Escudos",
				WeaponMastery:  "2 tipos de armas marciais",
				StartingEquip:  "(A) Cota de Malha, Espada Longa, Escudo, Símbolo Sagrado, Pacote de Sacerdote e 9 PO; ou (B) 150 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Imposição das Mãos (Lay on Hands)", Description: "Reserva de cura igual a 5x nível de Paladino como Ação Bônus."},
				{Level: 1, Name: "Conjuração de Paladino", Description: "Prepara e conjura magias divinas usando Carisma desde o nível 1!"},
				{Level: 2, Name: "Destruição Divina (Divine Smite)", Description: "Conjura Smite como magia bônus ao acertar um ataque para causar 2d8+ de dano radiante extra."},
				{Level: 6, Name: "Aura de Proteção (Aura of Protection)", Description: "Concede seu modificador de Carisma em TODAS as salvaguardas suas e de aliados em até 3m (9m no nv 18)!"},
			},
			Subclasses: []model.Subclass{
				{
					ID: "devotion", Name: "Oath of Devotion [PHB'24]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "O arquétipo do cavaleiro heroico em armadura reluzente, imune a encantos e portador de armas radiantes.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Arma Sagrada (Sacred Weapon)", Description: "Soma mod Carisma no ataque com arma e emite luz solar brilhante."},
					},
				},
				{
					ID: "glory", Name: "Oath of Glory [PHB'24]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Campeão das lendas atléticas, com velocidade sobre-humana, saltos incríveis e defesa gloriosa.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Atleta Inigualável", Description: "Vantagem em testes de Atletismo/Acrobacia e saltos aumentados."},
					},
				},
				{
					ID: "ancients", Name: "Oath of the Ancients [PHB'24]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Cavaleiro Verde que ama a luz e a vida, envolto em vinhas místicas e com aura de resistência a dano mágico.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Aura da Guarda", Description: "Você e aliados na aura ganham resistência a dano de feitiços e magias."},
					},
				},
				{
					ID: "vengeance", Name: "Oath of Vengeance [PHB'24]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Vingador implacável que pune criminosos e vilões com Voto de Inimizade para vantagem implacável.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Voto de Inimizade (Vow of Enmity)", Description: "Ação Bônus para ter Vantagem em TODAS as jogadas de ataque contra o alvo jurado."},
					},
				},
				{
					ID: "conquest", Name: "Oath of Conquest [XGE / Xanathar]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Conquistador tirânico que subjuga inimigos com auras de pavor avassalador que reduzem a velocidade a zero.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Aura de Conquista", Description: "Criaturas amedrontadas dentro de 3m têm velocidade reduzida a 0 e sofrem dano psíquico a cada rodada."},
					},
				},
				{
					ID: "watchers", Name: "Oath of the Watchers [TCE / Tasha]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Sentinela contra invasões extraplanares (aberrações, celestiais, demônios, fadas), concedendo iniciativa relâmpago ao grupo.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Aura do Sentinela", Description: "Concede bônus de proficiência na Iniciativa de todos os aliados próximos!"},
					},
				},
				{
					ID: "redemption", Name: "Oath of Redemption [XGE / Xanathar]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Pacificador sagrado que absorve dano direcionado aos seus aliados e reflete violência de volta aos agressores.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Emissário da Paz", Description: "+5 em testes de Persuasão e reação para absorver o dano recebido por um aliado próximo."},
					},
				},
				{
					ID: "oathbreaker", Name: "Oathbreaker [DMG]", ClassName: "Paladin", UnlockLevel: 3,
					Synopsis: "Paladino corrompido que quebrou seus votos sagrados, comandando mortos-vivos e adicionando dano necrótico.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Aura de Ódio", Description: "Soma mod Carisma no dano corpo a corpo seu e de todos os mortos-vivos/fiends ao redor."},
					},
				},
			},
		},
		{
			ID:     "ranger",
			Name:   "Ranger",
			NamePT: "Patrulheiro / Ranger",
			Emoji:  "🏹",
			Color:  0x52b788,
			Overview: "Mestre da sobrevivência nos ermos, rastreador supremo e arqueiro ou combatente de duas armas sintonizado com a magia da natureza e sua Marca do Caçador.",
			Role:     "Dano à Distância / Duas Armas / Exploração e Rastreamento / Suporte Primal",
			Traits: model.ClassTraits{
				PrimaryAbility: "Destreza e Sabedoria",
				HitDie:         "d10 por nível de Ranger",
				SavingThrows:   "Força e Destreza",
				SkillProf:      "Escolha 3: Adestrar Animais, Atletismo, Intuição, Investigação, Natureza, Percepção, Furtividade, Sobrevivência",
				WeaponProf:     "Todas as Armas Simples e Marciais",
				ArmorTraining:  "Armaduras Leves, Médias e Escudos",
				WeaponMastery:  "2 tipos de armas",
				StartingEquip:  "(A) Arco Longo com 20 flechas, 2 Espadas Curtas, Pacote de Explorador e 7 PO; ou (B) 150 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Marca do Caçador (Hunter's Mark)", Description: "Conjura Marca do Caçador sem gastar espaço de magia várias vezes por dia."},
				{Level: 1, Name: "Conjuração Primal", Description: "Prepara e conjura magias de Ranger usando Sabedoria desde o 1º nível."},
				{Level: 1, Name: "Maestria em Armas", Description: "Aplica propriedades como Slow, Vex e Nick em suas armas."},
				{Level: 2, Name: "Estilo de Luta & Deft Explorer", Description: "Ganha Estilo de Luta (Arquearia ou Duas Armas) e perícias com Especialização."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "gloom_stalker", Name: "Gloom Stalker [PHB'24]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Emboscador supremo do Subterrâneo, invisível para criaturas que dependem de Visão no Escuro e letal na 1ª rodada.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Emboscador Sombrio", Description: "Iniciativa bônus, deslocamento extra no 1º turno e dano extra com medo."},
						{Level: 3, Name: "Visão Umbral", Description: "Visão no escuro de 18m e invisibilidade na escuridão total."},
					},
				},
				{
					ID: "fey_wanderer", Name: "Fey Wanderer [PHB'24]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Abençoado com o charme das fadas, soma Sabedoria em testes de Carisma e desfere dano psíquico adicional.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Golpes Pavorosos", Description: "Causa 1d4 a 1d6 de dano psíquico extra por turno em cada criatura atingida."},
					},
				},
				{
					ID: "hunter", Name: "Hunter [PHB'24]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Caçador tático adaptável especializado em dizimar hordas de inimigos ou derrubar feras colossais.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Presa do Caçador", Description: "Escolhe Colossus Slayer (1d8 em alvo ferido) ou Horde Breaker (ataque extra adjacente)."},
					},
				},
				{
					ID: "beast_master", Name: "Beast Master [PHB'24]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Luta ao lado de um companheiro animal leal (Besta da Terra, do Mar ou do Céu) que ataca coordenado com o Ranger.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Companheiro Primal", Description: "Invoca uma fera leal que atua no combate sob seus comandos de Ação Bônus."},
					},
				},
				{
					ID: "drakewarden", Name: "Drakewarden [FTD / Fizban]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Cria e cavalga um dragão leal (Drake Companion) que cospe elementos e cresce até se tornar montaria alada.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Companheiro Dracônico", Description: "Invoca um draco que ataca com dano elemental e defende o ranger com reações."},
					},
				},
				{
					ID: "swarmkeeper", Name: "Swarmkeeper [TCE / Tasha]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Comanda um enxame de espíritos naturais (insetos, fadas, galhos) que empurram inimigos e concedem teletransporte e voo.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Enxame Coletado", Description: "Causa 1d6 de dano de força extra, empurra o alvo 4,5m ou move o próprio ranger sem provocar ataques."},
					},
				},
				{
					ID: "horizon_walker", Name: "Horizon Walker [XGE / Xanathar]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Guardião dos portais dimensionais que se teletransporta antes de cada ataque e converte todo o dano em dano de Força pura.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Guerreiro Planar", Description: "Converte todo o dano do ataque em Dano de Força + 2d8 extras."},
						{Level: 11, Name: "Golpe Distante", Description: "Teletransporta 3m antes de CADA ataque, atacando até 3 criaturas diferentes por turno!"},
					},
				},
				{
					ID: "hollow_warden", Name: "Hollow Warden [Ravenloft / The Horrors Within]", ClassName: "Ranger", UnlockLevel: 3,
					Synopsis: "Guardião macabro sintonizado com os terrores e espíritos das florestas amaldiçoadas de Ravenloft, assumindo forma bestial com aura de medo e retribuição.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Ira dos Ermos (Wrath of the Wild)", Description: "Gasta Inimigo Favorito com Ação Bônus para assumir forma grotesca por 1 min: +1 CA (+2 no nv 11), Ataque de Oportunidade imediato contra quem machuca você ou aliados, e Aura Aterrorizante de 3m que impõe a condição Amedrontado."},
						{Level: 3, Name: "Magias do Guardião Oco", Description: "Prepara Destruição Furiosa (Wrathful Smite), Alterar-se (Alter Self), Montaria Fantasmagórica, Dominar Fera e Ataque do Vento de Aço."},
						{Level: 7, Name: "Poder Faminto (Hungering Might)", Description: "Soma Sabedoria nas salvaguardas de Constituição e cura 1d10 + mod Sabedoria ao acertar ataques quando estiver machucado (Bloodied)."},
						{Level: 11, Name: "Podridão e Violência (Rot and Violence)", Description: "Criaturas amedrontadas na sua aura não recuperam PV nem usam reações; aplica propriedades de Maestria Sap ou Slow em armas."},
						{Level: 15, Name: "Poder Ancestral (Ancient Might)", Description: "Dano extra igual a mod Sabedoria em alvos amedrontados, imunidade a Exaustão e revive com 2x nível ao cair a 0 PV."},
					},
				},
			},
		},
		{
			ID:     "rogue",
			Name:   "Rogue",
			NamePT: "Ladino",
			Emoji:  "🗡️",
			Color:  0x6c757d,
			Overview: "Mestre da furtividade, perícias críticas e precisão cirúrgica. O Ladino explora fraquezas para causar Ataque Furtivo (Sneak Attack) massivo e escapa com Ação Astuta.",
			Role:     "Dano de Precisão / Infiltração / Desarme de Armadilhas / Especialista em Perícias",
			Traits: model.ClassTraits{
				PrimaryAbility: "Destreza (Dexterity)",
				HitDie:         "d8 por nível de Ladino",
				SavingThrows:   "Destreza e Inteligência",
				SkillProf:      "Escolha 4 perícias",
				WeaponProf:     "Armas Simples e Armas Marciais com Finesse",
				ArmorTraining:  "Armaduras Leves",
				WeaponMastery:  "2 tipos de armas com Finesse/Leve",
				StartingEquip:  "(A) Rapieira, Arco Curto com 20 flechas, 2 Adagas, Ferramentas de Ladrão, Pacote de Ladrão e 8 PO; ou (B) 110 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Especialização (Expertise)", Description: "Dobra seu bônus de proficiência em duas perícias ou Ferramentas de Ladrão."},
				{Level: 1, Name: "Ataque Furtivo (Sneak Attack)", Description: "Causa 1d6 a 10d6 de dano extra uma vez por turno com Vantagem ou aliado a 1,5m."},
				{Level: 2, Name: "Ação Astuta (Cunning Action)", Description: "Dash, Disengage ou Hide como Ação Bônus a cada turno."},
				{Level: 5, Name: "Golpe Astuto (Cunning Strike)", Description: "Troca dados de Sneak Attack por efeitos (Derrubar, Envenenar, Desarmar ou Retirar)."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "thief", Name: "Thief [PHB'24]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "O mestre dos arrombamentos e velocidade, capaz de usar itens mágicos como Ação Bônus e escalar telhados.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Mãos Rápidas", Description: "Ação Bônus para usar ferramentas de ladrão ou usar itens."},
					},
				},
				{
					ID: "assassin", Name: "Assassin [PHB'24]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "Especialista em emboscadas mortais e disfarces perfeitos, aniquilando alvos antes que possam reagir.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Assassinar", Description: "Vantagem em ataques e iniciativa no primeiro turno com dano crítico massivo."},
					},
				},
				{
					ID: "arcane_trickster", Name: "Arcane Trickster [PHB'24]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "Ladino que aprimora seus truques com magias de Ilusão e Encantamento de Mago e uma Mão Mágica invisível.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Mãos Mágicas de Prestidigitador", Description: "Controla Mão Mágica invisível que furta e desarma fechaduras à distância."},
					},
				},
				{
					ID: "soulknife", Name: "Soulknife [PHB'24]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "Gera lâminas de pura energia psíquica das mãos para arremesso letal, telepatia de grupo e saltos mentais.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Lâminas Psíquicas", Description: "Invoca facas de energia psíquica (1d6 de dano) que não deixam marcas físicas nem ruído."},
					},
				},
				{
					ID: "swashbuckler", Name: "Swashbuckler [XGE / Xanathar]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "Duelista carismático e audacioso que aplica Sneak Attack em duelos 1v1 sem precisar de vantagem e soma Carisma na iniciativa.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Audácia Espalhafatosa (Rakish Audacity)", Description: "Soma mod Carisma na Iniciativa e aplica Ataque Furtivo em combates individuais sem aliados por perto!"},
					},
				},
				{
					ID: "phantom", Name: "Phantom [TCE / Tasha]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "Caminha no limiar da morte, colhendo fragmentos de alma de inimigos mortos e causando dano necrótico duplo com sussurros fúnebres.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Sussurros dos Mortos", Description: "Ganha proficiência temporária em qualquer perícia ao consultar fantasmas."},
						{Level: 3, Name: "Goles da Morte", Description: "Causa dano necrótico a um segundo alvo adjacente ao usar Ataque Furtivo."},
					},
				},
				{
					ID: "scout", Name: "Scout [XGE / Xanathar]", ClassName: "Rogue", UnlockLevel: 3,
					Synopsis: "Batedor furtivo e sobrevivente que se esquiva de inimigos adjacentes com reação e ataca duas vezes na rodada com Emboscada Súbita.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Esparramador", Description: "Usa reação para se mover metade do deslocamento sem provocar ataques quando um inimigo termina o turno ao seu lado."},
					},
				},
			},
		},
		{
			ID:     "sorcerer",
			Name:   "Sorcerer",
			NamePT: "Feiticeiro",
			Emoji:  "🔮",
			Color:  0xd90429,
			Overview: "A magia corre em seu próprio sangue através de heranças ancestrais de dragões, caos selvagem ou anomalias cósmicas. Manipula a Metamagia para dobrar as regras dos feitiços.",
			Role:     "Dano Arcano Massivo / Metamagia / Conjurador Espontâneo",
			Traits: model.ClassTraits{
				PrimaryAbility: "Carisma (Charisma)",
				HitDie:         "d6 por nível de Feiticeiro",
				SavingThrows:   "Constituição e Carisma",
				SkillProf:      "Escolha 2: Arcanismo, Enganação, Intuição, Intimidação, Persuasão, Religião",
				WeaponProf:     "Armas Simples",
				ArmorTraining:  "Nenhuma",
				WeaponMastery:  "Nenhuma",
				StartingEquip:  "(A) 2 Adagas, Foco Arcano, Pacote de Masmorra e 28 PO; ou (B) 50 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Feitiçaria Inata (Innate Sorcery)", Description: "Vantagem nos ataques mágicos e +1 na CD das magias por 1 minuto."},
				{Level: 2, Name: "Pontos de Feitiçaria & Metamagia", Description: "Acelerar (Quickened), Duplicada (Twinned), Sutil, Cuidadosa, etc."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "draconic", Name: "Draconic Sorcery [PHB'24]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Sangue de dragão antigo que concede escamas protetoras (CA 13+DES), dano elemental avassalador e asas para voo livre.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Resiliência Draconiana", Description: "CA 13 + mod DES sem armadura e +1 PV por nível."},
						{Level: 14, Name: "Asas de Dragão", Description: "Brota asas de dragão para voar livremente."},
					},
				},
				{
					ID: "wild_magic", Name: "Wild Magic Sorcery [PHB'24]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Controla o caos puro da magia, desencadeando surtos aleatórios imprevisíveis e manipulando as marés da sorte.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Marés do Caos", Description: "Ganha Vantagem imediata em qualquer rolagem à vontade."},
					},
				},
				{
					ID: "aberrant", Name: "Aberrant Sorcery [PHB'24]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Tocado por horrores do Reino Distante, empunha telepatia, magias psíquicas sem componentes e tentáculos etéreos.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Magias Psiônicas", Description: "Conjura magias psíquicas silenciosamente com Pontos de Feitiçaria sem gastar componentes."},
					},
				},
				{
					ID: "clockwork", Name: "Clockwork Sorcery [PHB'24]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Sintonizado com a harmonia mecânica de Mechanus, anula desvantagens/vantagens caóticas e cria bastiões de proteção.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Restaurar o Equilíbrio", Description: "Anula vantagem ou desvantagem de qualquer rolagem com reação."},
					},
				},
				{
					ID: "divine_soul", Name: "Divine Soul [XGE / Xanathar]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Descendente de seres celestiais divinos, capaz de conjurar qualquer magia da lista de Clérigo E de Feiticeiro ao mesmo tempo!",
					Features: []model.SubclassFeature{
						{Level: 1, Name: "Magia Divina", Description: "Aprende qualquer magia da lista de Clérigo e ganha asas divinas de anjo no nv 14."},
					},
				},
				{
					ID: "shadow_sorc", Name: "Shadow Magic [XGE / Xanathar]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Extrai poder da Umbra (Shadowfell), invocando um Cão da Morte sombrio que persegue alvos e sobrevivendo a golpes mortais.",
					Features: []model.SubclassFeature{
						{Level: 1, Name: "Olhos da Escuridão", Description: "Visão no escuro de 36m e conjura Escuridão mágica que ele próprio enxerga."},
						{Level: 6, Name: "Cão da Ilusão (Hound of Ill Omen)", Description: "Invoca um lobo de sombras que impõe desvantagem em todas as salvaguardas do alvo contra suas magias!"},
					},
				},
				{
					ID: "storm_sorc", Name: "Storm Sorcery [XGE / Xanathar]", ClassName: "Sorcerer", UnlockLevel: 3,
					Synopsis: "Mestre dos ventos e trovões tempestuosos, flutuando após cada feitiço e comandando a eletricidade.",
					Features: []model.SubclassFeature{
						{Level: 1, Name: "Magia Tempestuosa", Description: "Voa até 3m sem provocar ataques de oportunidade sempre que conjurar uma magia."},
					},
				},
			},
		},
		{
			ID:     "warlock",
			Name:   "Warlock",
			NamePT: "Bruxo / Warlock",
			Emoji:  "👁️",
			Color:  0x7209b7,
			Overview: "Pactuou com uma entidade transcendental (Demônio, Feérico, Ser Celestial ou Ancião das Estrelas) em troca de Magia de Pacto de alto poder que recarrega no Descanso Curto e Invocações Ocultas.",
			Role:     "Disparador Mágico à Distância (Eldritch Blast) / Maldições / Magia de Pacto Recarregável",
			Traits: model.ClassTraits{
				PrimaryAbility: "Carisma (Charisma)",
				HitDie:         "d8 por nível de Bruxo",
				SavingThrows:   "Sabedoria e Carisma",
				SkillProf:      "Escolha 2: Arcanismo, Enganação, História, Intimidação, Investigação, Natureza, Religião",
				WeaponProf:     "Armas Simples",
				ArmorTraining:  "Armaduras Leves",
				WeaponMastery:  "Nenhuma (a menos que Pact of the Blade)",
				StartingEquip:  "(A) Foco Arcano, Adaga, Pacote de Erudito e 15 PO; ou (B) 100 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Magia de Pacto (Pact Magic)", Description: "Espaços de magia no nível MÁXIMO que recarregam no Descanso Curto (Short Rest)."},
				{Level: 1, Name: "Invocações Ocultas (Eldritch Invocations)", Description: "Customizações místicas desbloqueadas desde o nível 1 no 2024!"},
				{Level: 2, Name: "Rajada Mística (Eldritch Blast)", Description: "1d10 de dano de força por raio, somando mod Carisma com Agonizing Blast."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "fiend", Name: "Fiend Patron [PHB'24]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto infernal com lordes demoníacos, ganhando vida temporária ao abater inimigos e arremessando alvos ao inferno.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Bênção do Tenebroso", Description: "Ganha PVs temporários ao reduzir inimigo a 0 PVs."},
						{Level: 14, Name: "Arremessar ao Inferno", Description: "Bane o alvo aos nove infernos causando 10d10 de dano psíquico."},
					},
				},
				{
					ID: "archfey", Name: "Archfey Patron [PHB'24]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com nobres feéricos da Agrestia das Fadas, concedendo passos nebulosos constantes e truques ilusionistas.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Passos da Agrestia", Description: "Conjura Passo Nebuloso (Misty Step) sem gastar slots várias vezes ao dia com efeitos mágicos."},
					},
				},
				{
					ID: "celestial", Name: "Celestial Patron [PHB'24]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com um ser celestial dos céus, garantindo piscina de cura de reserva e dano radiante sagrado.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Luz Curativa (Healing Light)", Description: "Piscina de dados de d6 para curar aliados como Ação Bônus à distância."},
					},
				},
				{
					ID: "great_old_one", Name: "Great Old One Patron [PHB'24]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com divindades cósmicas como Cthulhu, ganhando telepatia, feitiços psíquicos e proteção mental.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Mente Desperta", Description: "Conexão telepática psíquica e dano psíquico extra com encantamento."},
					},
				},
				{
					ID: "hexblade", Name: "The Hexblade [XGE / Xanathar]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com uma lâmina senciente sombria de Shadowfell, atacando usando Carisma em vez de Força e amaldiçoando inimigos para críticos em 19-20.",
					Features: []model.SubclassFeature{
						{Level: 1, Name: "Guerreiro do Hex", Description: "Usa mod Carisma no ataque e dano com armas e ganha proficiência em Armaduras Médias e Escudos!"},
						{Level: 1, Name: "Maldição do Hexblade", Description: "Crítico em 19-20 e dano extra igual ao bônus de proficiência no alvo marcado."},
					},
				},
				{
					ID: "genie", Name: "The Genie [TCE / Tasha]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com um Gênio elemental (Dao, Djiin, Efreeti, Marid), abrigando-se dentro de uma lâmpada mágica e voando livremente.",
					Features: []model.SubclassFeature{
						{Level: 1, Name: "Recipiente do Gênio (Genie's Vessel)", Description: "Entra dentro de um anel/lâmpada mágica para descansar e causa dano elemental extra em ataques."},
					},
				},
				{
					ID: "fathomless", Name: "The Fathomless [TCE / Tasha]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com horrores das profundezas oceânicas como Kraken, invocando tentáculos de água que atacam e protegem aliados.",
					Features: []model.SubclassFeature{
						{Level: 1, Name: "Tentáculo das Profundezas", Description: "Invoca tentáculo espectral de 3m como Ação Bônus que ataca com dano de frio e reduz velocidade do inimigo."},
					},
				},
				{
					ID: "undead", Name: "The Undead [VRGtR / Ravenloft]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com uma entidade que desafia a morte (como Strahd von Zarovich ou Azalin Rex), manifestando a temida Forma do Pavor e imunidade biológica necrótica.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Forma de Pavor (Form of Dread)", Description: "Com Ação Bônus, transforma-se por 1 minuto: ganha 1d10 + nível em PVs temporários, Imunidade à condição Amedrontado, e 1x por turno ao acertar um ataque pode amedrontar o alvo (save Sabedoria). Usos = mod Carisma por Long Rest."},
						{Level: 3, Name: "Magias do Morto-Vivo (Undead Spells)", Description: "Prepara Perdição (Bane), Cegueira/Surdez, Força Fantasmagórica, Raio do Adoecimento, Falar com Mortos, Invocar Morto-Vivo, Invisibilidade Maior, Assassino Fantasmagórico, Barreira Antivida e Névoa Mortal."},
						{Level: 6, Name: "Toque da Cova (Grave Touched)", Description: "Dano necrótico ignora resistência e pode trocar dano de magias para necrótico. Em Forma de Pavor, rola +1 dado extra de dano necrótico ao acertar ataques! Além disso, não sofre exaustão por sede/fome/asfixia e não precisa dormir."},
						{Level: 10, Name: "Casca Necrótica (Necrotic Husk)", Description: "Resistência a dano Necrótico (Imunidade durante a Forma de Pavor). Ao cair a 0 PV, explode em necrose de 9m causando 2d10+CAR e revive com 2x nível de PVs (1x por descanso)."},
						{Level: 14, Name: "Pavor Superior (Superior Dread)", Description: "Na Forma de Pavor ganha Resistência a dano Cortante, Perfurante e Concussão, deslocamento de Voo com flutuar atravessando paredes/criaturas, e conjura magias de Necromancia e Conjuração sem componentes verbais/somáticos/materiais."},
					},
				},
				{
					ID: "vestige", Name: "Vestige Patron [Ravenloft / Amber Temple]", ClassName: "Warlock", UnlockLevel: 3,
					Synopsis: "Pacto com os vestígios sombrios de deuses mortos e entidades proibidas aprisionadas nos blocos de âmbar de Barovia (Templo de Âmbar).",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Companheiro Vestígio (Vestige Companion)", Description: "Manifesta um avatar do deus caído (Celestial, Fiend ou Morto-Vivo) que flutua, ataca com dano elemental/necrótico e concede habilidades amaldiçoadas."},
						{Level: 6, Name: "Eco Divino", Description: "Canaliza relíquias do poder cósmico esquecido para impor maldições e trocar de lugar com o vestígio instantaneamente."},
					},
				},
			},
		},
		{
			ID:     "wizard",
			Name:   "Wizard",
			NamePT: "Mago / Bruxo Arcano",
			Emoji:  "📜",
			Color:  0x3a86ff,
			Overview: "O estudioso supremo dos mistérios arcanos. O Mago possui o mais vasto grimório de feitiços de todo o D&D, alterando a realidade com rituais, abjurações e invocações de alta complexidade.",
			Role:     "Maior Versatilidade Mágica do Jogo / Controle em Massa / Dano em Área / Rituais",
			Traits: model.ClassTraits{
				PrimaryAbility: "Inteligência (Intelligence)",
				HitDie:         "d6 por nível de Mago",
				SavingThrows:   "Inteligência e Sabedoria",
				SkillProf:      "Escolha 2: Arcanismo, História, Intuição, Investigação, Medicina, Religião",
				WeaponProf:     "Armas Simples",
				ArmorTraining:  "Nenhuma",
				WeaponMastery:  "Nenhuma",
				StartingEquip:  "(A) Grimório de Magias, Foco Arcano, 2 Adagas, Pacote de Erudito e 5 PO; ou (B) 55 PO",
			},
			Features: []model.ClassFeature{
				{Level: 1, Name: "Grimório de Magias (Spellbook)", Description: "Aprende 6 magias no nível 1 e +2 a cada novo nível, podendo copiar pergaminhos."},
				{Level: 1, Name: "Conjuração de Rituais", Description: "Conjura qualquer magia com tag Ritual do grimório sem gastar espaços de magia."},
				{Level: 1, Name: "Recuperação Arcana", Description: "Recupera espaços de magia no Descanso Curto uma vez por dia."},
			},
			Subclasses: []model.Subclass{
				{
					ID: "abjurer", Name: "Abjurer [PHB'24]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Mestre da proteção e anulação de magias inimigas, criando um Escudo Arcano que absorve dano massivo.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Guarda Arcana", Description: "Barreira mágica com PVs próprios que absorve dano para você ou aliados."},
					},
				},
				{
					ID: "diviner", Name: "Diviner [PHB'24]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Mestre das profecias e do destino, rola dados de Portento a cada amanhecer para substituir rolagens em momentos cruciais.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Portento (Portent)", Description: "Rola 2 dados de d20 após cada descanso e substitui QUALQUER rolagem de ataque, teste ou save no dia!"},
					},
				},
				{
					ID: "evoker", Name: "Evoker [PHB'24]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Mestre da destruição elementar, moldando bolas de fogo para não ferir aliados e potencializando truques.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Esculpir Feitiços (Sculpt Spells)", Description: "Cria bolsões seguros em magias de área (ex: Bola de Fogo), fazendo aliados passarem no save e não sofrerem dano."},
					},
				},
				{
					ID: "illusionist", Name: "Illusionist [PHB'24]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Mestre da decepção sensorial e miragens, alterando ilusões ativas e tornando-as quase reais.",
					Features: []model.SubclassFeature{
						{Level: 3, Name: "Ilusão Menor Aprimorada", Description: "Conjura Ilusão Menor como Ação Bônus combinando som e imagem."},
					},
				},
				{
					ID: "bladesinging", Name: "Bladesinging [TCE / Tasha]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Esgrimista arcano élfico que entra na Dança das Lâminas, somando Inteligência na CA e na concentração e atacando com armas leves.",
					Features: []model.SubclassFeature{
						{Level: 2, Name: "Canção das Lâminas (Bladesong)", Description: "Ganha bônus de Inteligência na CA, +3m de deslocamento e bônus na concentração de feitiços!"},
						{Level: 6, Name: "Ataque Extra com Truque", Description: "Ataca duas vezes e substitui um ataque por qualquer truque mágico (Cantrip)!"},
					},
				},
				{
					ID: "chronurgy", Name: "Chronurgy Magic [EGW / Wildemount]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Manipulador do tempo e linhas temporais, forçando re-rolagens de dados e paralisando criaturas em estase temporal.",
					Features: []model.SubclassFeature{
						{Level: 2, Name: "Mudança Temporal (Chronal Shift)", Description: "Usa reação para forçar um atacante ou aliado a rolar novamente o d20!"},
						{Level: 10, Name: "Esfera do Momento", Description: "Condensa uma magia de até 4º nível em uma conta temporal que qualquer aliado pode detonar."},
					},
				},
				{
					ID: "scribes", Name: "Order of Scribes [TCE / Tasha]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Grimório consciente que altera o tipo elemental das magias (ex: Bola de Fogo de Dano Radiante) e conjura através de um avatar espectral.",
					Features: []model.SubclassFeature{
						{Level: 2, Name: "Mente do Grimório", Description: "Pode trocar o tipo de dano de qualquer feitiço pelo dano de outra magia do grimório!"},
					},
				},
				{
					ID: "necromancy", Name: "School of Necromancy [PHB]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Comanda legiões de mortos-vivos, aumentando os PVs e o dano de esqueletos e zumbis reanimados.",
					Features: []model.SubclassFeature{
						{Level: 6, Name: "Servos Mortos-Vivos", Description: "Cria mortos-vivos adicionais com PVs aumentados e bônus de dano de proficiência."},
					},
				},
				{
					ID: "war_magic", Name: "War Magic [XGE / Xanathar]", ClassName: "Wizard", UnlockLevel: 3,
					Synopsis: "Mago tático de campo de batalha com reflexos arcanos que somam +4 na salvaguarda e iniciativa aumentada.",
					Features: []model.SubclassFeature{
						{Level: 2, Name: "Deflexão Arcana", Description: "Reação para somar +2 na CA ou +4 em salvaguardas contra qualquer perigo."},
					},
				},
			},
		},
	}
}

func loadSpells(basePath string) []model.Spell {
	filePath := basePath + "/references/dnd-5e-srd-markdown/spells.md"
	b, err := os.ReadFile(filePath)
	if err != nil {
		slog.Warn("Não foi possível carregar spells.md, usando lista base", "error", err)
		return nil
	}

	content := string(b)
	descIdx := strings.Index(content, "## Spell Descriptions")
	if descIdx == -1 {
		return nil
	}
	spellContent := content[descIdx:]

	reHeader := regexp.MustCompile(`(?m)^#### ([^\n\r]+)[\r\n]+_([^\n\r]+)_`)
	matches := reHeader.FindAllStringSubmatchIndex(spellContent, -1)

	var spells []model.Spell

	for i, m := range matches {
		name := strings.TrimSpace(spellContent[m[2]:m[3]])
		headerType := strings.TrimSpace(spellContent[m[4]:m[5]])
		
		bodyEnd := len(spellContent)
		if i < len(matches)-1 {
			bodyEnd = matches[i+1][0]
		}
		body := spellContent[m[1]:bodyEnd]

		s := model.Spell{
			ID:   strings.ToLower(strings.ReplaceAll(name, " ", "_")),
			Name: name,
		}

		// Extrair Nível, Escola e Classes do headerType
		parseSpellHeader(&s, headerType)

		// Extrair propriedades do body
		s.CastingTime = extractProperty(body, "Casting Time:")
		s.Range = extractProperty(body, "Range:")
		s.Components = extractProperty(body, "Components:")
		s.Duration = extractProperty(body, "Duration:")

		if strings.Contains(strings.ToLower(s.Duration), "concentration") {
			s.Concentration = true
		}
		if strings.Contains(strings.ToLower(headerType), "ritual") || strings.Contains(strings.ToLower(body), "ritual") {
			s.Ritual = true
		}

		// Extrair Higher Levels e Descrição limpa
		if hlIdx := strings.Index(body, "_**Using a Higher-Level Spell Slot.**_"); hlIdx != -1 {
			s.HigherLevels = strings.TrimSpace(body[hlIdx+len("_**Using a Higher-Level Spell Slot.**_"):])
			s.Description = cleanSpellBody(body[:hlIdx])
		} else if hlIdx2 := strings.Index(body, "**Using a Higher-Level Spell Slot.**"); hlIdx2 != -1 {
			s.HigherLevels = strings.TrimSpace(body[hlIdx2+len("**Using a Higher-Level Spell Slot.**"):])
			s.Description = cleanSpellBody(body[:hlIdx2])
		} else {
			s.Description = cleanSpellBody(body)
		}

		spells = append(spells, s)
	}

	return spells
}

func parseSpellHeader(s *model.Spell, header string) {
	headerLow := strings.ToLower(header)

	if strings.Contains(headerLow, "cantrip") {
		s.Level = 0
	} else if idx := strings.Index(headerLow, "level "); idx != -1 {
		numStr := headerLow[idx+6 : idx+7]
		if n, err := strconv.Atoi(numStr); err == nil {
			s.Level = n
		}
	}

	schools := []string{"abjuration", "conjuration", "divination", "enchantment", "evocation", "illusion", "necromancy", "transmutation"}
	for _, sc := range schools {
		if strings.Contains(headerLow, sc) {
			s.School = strings.Title(sc)
			break
		}
	}

	// Classes entre parênteses
	if pStart := strings.Index(header, "("); pStart != -1 {
		if pEnd := strings.Index(header[pStart:], ")"); pEnd != -1 {
			classesStr := header[pStart+1 : pStart+pEnd]
			parts := strings.Split(classesStr, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" && !strings.EqualFold(p, "ritual") {
					s.Classes = append(s.Classes, p)
				}
			}
		}
	}
}

func extractProperty(body, prefix string) string {
	idx := strings.Index(body, prefix)
	if idx == -1 {
		return ""
	}
	sub := body[idx+len(prefix):]
	endIdx := strings.IndexAny(sub, "\r\n")
	if endIdx == -1 {
		endIdx = len(sub)
	}
	return strings.TrimSpace(sub[:endIdx])
}

func cleanSpellBody(body string) string {
	lines := strings.Split(body, "\n")
	var clean []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "Casting Time:") || strings.HasPrefix(l, "Range:") ||
			strings.HasPrefix(l, "Components:") || strings.HasPrefix(l, "Duration:") {
			continue
		}
		clean = append(clean, l)
	}
	result := strings.Join(clean, "\n\n")
	if len(result) > 1000 {
		return result[:997] + "..."
	}
	return result
}

func loadMonsters(basePath string) []model.Monster {
	filePath := basePath + "/references/dnd-5e-srd-markdown/monsters-A-Z.md"
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	content := string(b)

	reHeader := regexp.MustCompile(`(?m)^###? ([^\n\r]+)[\r\n]+_([^\n\r]+)_`)
	matches := reHeader.FindAllStringSubmatchIndex(content, -1)

	var monsters []model.Monster

	for i, m := range matches {
		name := strings.TrimSpace(content[m[2]:m[3]])
		typeLine := strings.TrimSpace(content[m[4]:m[5]])
		
		bodyEnd := len(content)
		if i < len(matches)-1 {
			bodyEnd = matches[i+1][0]
		}
		body := content[m[1]:bodyEnd]

		mon := model.Monster{
			ID:   strings.ToLower(strings.ReplaceAll(name, " ", "_")),
			Name: name,
		}

		// Tipo e Tendência
		parts := strings.Split(typeLine, ",")
		if len(parts) > 0 {
			mon.Type = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			mon.Alignment = strings.TrimSpace(parts[1])
		}

		mon.ArmorClass = cleanMonsterProp(extractProperty(body, "**AC**"))
		if mon.ArmorClass == "" {
			mon.ArmorClass = cleanMonsterProp(extractProperty(body, "**Armor Class**"))
		}
		mon.HitPoints = cleanMonsterProp(extractProperty(body, "**HP**"))
		if mon.HitPoints == "" {
			mon.HitPoints = cleanMonsterProp(extractProperty(body, "**Hit Points**"))
		}
		mon.Speed = cleanMonsterProp(extractProperty(body, "**Speed**"))
		mon.SavingThrows = cleanMonsterProp(extractProperty(body, "**Saving Throws**"))
		mon.Skills = cleanMonsterProp(extractProperty(body, "**Skills**"))
		mon.DamageResist = cleanMonsterProp(extractProperty(body, "**Damage Resistances**"))
		mon.DamageImmune = cleanMonsterProp(extractProperty(body, "**Damage Immunities**"))
		mon.ConditionImmune = cleanMonsterProp(extractProperty(body, "**Condition Immunities**"))
		mon.Senses = cleanMonsterProp(extractProperty(body, "**Senses**"))
		mon.Languages = cleanMonsterProp(extractProperty(body, "**Languages**"))
		mon.ChallengeRating = cleanMonsterProp(extractProperty(body, "**CR**"))
		if mon.ChallengeRating == "" {
			mon.ChallengeRating = cleanMonsterProp(extractProperty(body, "**Challenge**"))
		}

		// Extrair Ações e Traits
		mon.Actions = extractMonsterActions(body, "### Actions")
		mon.BonusActions = extractMonsterActions(body, "### Bonus Actions")
		mon.Reactions = extractMonsterActions(body, "### Reactions")

		monsters = append(monsters, mon)
	}

	return monsters
}

func extractMonsterActions(body, sectionTitle string) []model.MonsterAction {
	idx := strings.Index(body, sectionTitle)
	if idx == -1 {
		return nil
	}
	secBody := body[idx+len(sectionTitle):]
	nextSec := strings.Index(secBody, "### ")
	if nextSec != -1 {
		secBody = secBody[:nextSec]
	}

	reAct := regexp.MustCompile(`(?m)^\*\*\*([^*]+)\.\*\*\*`)
	matches := reAct.FindAllStringSubmatchIndex(secBody, -1)

	var actions []model.MonsterAction
	for i, m := range matches {
		actName := strings.TrimSpace(secBody[m[2]:m[3]])
		actEnd := len(secBody)
		if i < len(matches)-1 {
			actEnd = matches[i+1][0]
		}
		actDesc := strings.TrimSpace(secBody[m[1]:actEnd])
		actions = append(actions, model.MonsterAction{
			Name:        actName,
			Description: actDesc,
		})
	}
	return actions
}

func loadItems(basePath string) []model.Item {
	var items []model.Item

	// 1. Armas com Maestria de 2024 (Equipment.md)
	weapons := []model.Item{
		{ID: "greataxe", Name: "Greataxe (Machado Grande)", Category: "Arma Marcial Corpo a Corpo", Cost: "30 PO", Weight: "3,1 kg", Damage: "1d12 cortante", Mastery: "Cleave (Gera ataque extra em criatura adjacente)", Properties: "Pesada, Duas Mãos"},
		{ID: "greatsword", Name: "Greatsword (Espada Grande)", Category: "Arma Marcial Corpo a Corpo", Cost: "50 PO", Weight: "2,7 kg", Damage: "2d6 cortante", Mastery: "Graze (Causa mod de FOR mesmo se errar o ataque!)", Properties: "Pesada, Duas Mãos"},
		{ID: "longsword", Name: "Longsword (Espada Longa)", Category: "Arma Marcial Corpo a Corpo", Cost: "15 PO", Weight: "1,3 kg", Damage: "1d8 (1d10) cortante", Mastery: "Sap (Impõe desvantagem no próximo ataque do inimigo)", Properties: "Versátil"},
		{ID: "rapier", Name: "Rapier (Rapieira)", Category: "Arma Marcial Corpo a Corpo", Cost: "25 PO", Weight: "0,9 kg", Damage: "1d8 perfurante", Mastery: "Vex (Concede Vantagem no seu próximo ataque contra o alvo!)", Properties: "Finesse"},
		{ID: "shortsword", Name: "Shortsword (Espada Curta)", Category: "Arma Marcial Corpo a Corpo", Cost: "10 PO", Weight: "0,9 kg", Damage: "1d6 perfurante", Mastery: "Vex (Concede Vantagem no próximo ataque)", Properties: "Finesse, Leve"},
		{ID: "dagger", Name: "Dagger (Adaga)", Category: "Arma Simples Corpo a Corpo", Cost: "2 PO", Weight: "0,4 kg", Damage: "1d4 perfurante", Mastery: "Nick (Permite o ataque da mão inábil como parte da Ação de Ataque sem gastar Ação Bônus!)", Properties: "Finesse, Leve, Arremesso (6/18m)"},
		{ID: "halberd", Name: "Halberd (Alabarda)", Category: "Arma Marcial Corpo a Corpo", Cost: "20 PO", Weight: "2,7 kg", Damage: "1d10 cortante", Mastery: "Cleave (Ataque extra adjacente)", Properties: "Pesada, Alcance, Duas Mãos"},
		{ID: "maul", Name: "Maul (Malho)", Category: "Arma Marcial Corpo a Corpo", Cost: "10 PO", Weight: "4,5 kg", Damage: "2d6 concussão", Mastery: "Topple (Força salvaguarda de CON para não cair Caído/Prone!)", Properties: "Pesada, Duas Mãos"},
		{ID: "warhammer", Name: "Warhammer (Martelo de Guerra)", Category: "Arma Marcial Corpo a Corpo", Cost: "15 PO", Weight: "0,9 kg", Damage: "1d8 (1d10) concussão", Mastery: "Push (Empurra o alvo 3 metros para longe sem save!)", Properties: "Versátil"},
		{ID: "longbow", Name: "Longbow (Arco Longo)", Category: "Arma Marcial à Distância", Cost: "50 PO", Weight: "0,9 kg", Damage: "1d8 perfurante", Mastery: "Slow (Reduz o deslocamento do inimigo em 3 metros)", Properties: "Munição (45/180m), Pesada, Duas Mãos"},
		{ID: "shortbow", Name: "Shortbow (Arco Curto)", Category: "Arma Simples à Distância", Cost: "25 PO", Weight: "0,9 kg", Damage: "1d6 perfurante", Mastery: "Vex (Concede Vantagem no próximo ataque)", Properties: "Munição (24/96m), Duas Mãos"},
		{ID: "heavy_crossbow", Name: "Heavy Crossbow (Besta Pesada)", Category: "Arma Marcial à Distância", Cost: "50 PO", Weight: "8,1 kg", Damage: "1d10 perfurante", Mastery: "Push (Empurra o alvo 3 metros para trás)", Properties: "Munição (30/120m), Pesada, Recarga, Duas Mãos"},
		{ID: "hand_crossbow", Name: "Hand Crossbow (Besta de Mão)", Category: "Arma Marcial à Distância", Cost: "75 PO", Weight: "1,3 kg", Damage: "1d6 perfurante", Mastery: "Vex (Vantagem no próximo ataque)", Properties: "Munição (9/36m), Leve, Recarga"},
	}
	items = append(items, weapons...)

	// 2. Itens Mágicos Principais de D&D 2024
	magicItems := []model.Item{
		{ID: "potion_healing", Name: "Potion of Healing (Poção de Cura)", Category: "Poção", Rarity: "Comum", Cost: "50 PO", Description: "No D&D 2024, beber uma poção em si mesmo requer apenas uma **Ação Bônus**! Restaura 2d4 + 2 Pontos de Vida."},
		{ID: "potion_greater_healing", Name: "Potion of Greater Healing (Poção de Cura Maior)", Category: "Poção", Rarity: "Incomum", Cost: "250 PO", Description: "Restaura 4d4 + 4 Pontos de Vida como Ação Bônus."},
		{ID: "bag_of_holding", Name: "Bag of Holding (Bolsa Espaçosa)", Category: "Item Maravilhoso", Rarity: "Incomum", Weight: "6,8 kg", Description: "Armazena até 226 kg (500 lbs) em um espaço dimensional de 1,8m cúbicos sem alterar seu peso exterior."},
		{ID: "cloak_of_protection", Name: "Cloak of Protection (Manto de Proteção)", Category: "Item Maravilhoso", Rarity: "Incomum", Attunement: "Requer Sintonização", Description: "Ganha +1 na Classe de Armadura e +1 em TODAS as salvaguardas enquanto estiver vestindo o manto."},
		{ID: "flame_tongue", Name: "Flame Tongue (Língua Flamejante)", Category: "Arma Mágica (Qualquer Espada)", Rarity: "Rara", Attunement: "Requer Sintonização", Description: "Pronuncia a palavra de comando com Ação Bônus para inflamar a lâmina, causando +2d6 de dano de fogo em cada golpe e emitindo luz."},
		{ID: "sun_blade", Name: "Sun Blade (Lâmina Solar)", Category: "Arma Mágica (Espada Longa)", Rarity: "Rara", Attunement: "Requer Sintonização", Description: "Lâmina de pura luz solar radiante. Ganha +2 em ataques e dano, causando 1d8 (1d10) de dano radiante (+1d8 extra contra mortos-vivos)."},
		{ID: "vorpal_sword", Name: "Vorpal Sword (Espada Vorpal)", Category: "Arma Mágica", Rarity: "Lendária", Attunement: "Requer Sintonização", Description: "Arma +3 que decepa instantaneamente a cabeça de uma criatura ao rolar um 20 natural no dado de ataque."},
		{ID: "deck_of_many_things", Name: "Deck of Many Things (Baralho das Surpresas)", Category: "Item Maravilhoso", Rarity: "Lendário / Artefato", Description: "Baralho místico de cartas mágicas capaz de conceder desejos infinitos, castelos, demônios devoradores ou apagar a alma do jogador."},
	}
	items = append(items, magicItems...)

	return items
}

func loadRules(basePath string) []model.Rule {
	filePath := basePath + "/references/dnd-5e-srd-markdown/rules-glossary.md"
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	content := string(b)

	reHeader := regexp.MustCompile(`(?m)^#### ([^\n\r]+)`)
	matches := reHeader.FindAllStringSubmatchIndex(content, -1)

	var rules []model.Rule
	for i, m := range matches {
		name := strings.TrimSpace(content[m[2]:m[3]])
		
		bodyEnd := len(content)
		if i < len(matches)-1 {
			bodyEnd = matches[i+1][0]
		}
		body := strings.TrimSpace(content[m[1]:bodyEnd])

		// Resumo simples com primeiras 2 frases
		summary := body
		if dotIdx := strings.Index(body, "."); dotIdx != -1 {
			if dotIdx2 := strings.Index(body[dotIdx+1:], "."); dotIdx2 != -1 {
				summary = body[:dotIdx+1+dotIdx2+1]
			} else {
				summary = body[:dotIdx+1]
			}
		}

		if len(body) > 1200 {
			body = body[:1197] + "..."
		}

		rules = append(rules, model.Rule{
			ID:          strings.ToLower(strings.ReplaceAll(name, " ", "_")),
			Name:        name,
			Summary:     summary,
			Description: body,
		})
	}

	return rules
}

func cleanMonsterProp(val string) string {
	val = strings.ReplaceAll(val, "<br>", "")
	val = strings.ReplaceAll(val, "<br/>", "")
	val = strings.ReplaceAll(val, "**Initiative**", "• Inic:")
	return strings.TrimSpace(val)
}

func loadSpecies() []model.Species {
	return []model.Species{
		// === RAVENLOFT SPECIES / LINHAGENS ===
		{
			ID:           "dhampir",
			Name:         "Dhampir",
			NamePT:       "Dhampiro",
			Category:     "Ravenloft",
			Source:       "Ravenloft - The Horrors Within / VRGtR",
			CreatureType: "Humanoide (Humanoid)",
			Size:         "Médio (Medium) ou Pequeno (Small)",
			Speed:        "10,5 metros (35 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Criaturas que habitam o limiar entre os vivos e os mortos-vivos. Amaldiçoados por uma fome macabra (por sangue, sonhos ou energia vital), possuem agilidade sobrenatural e presas mortais herdadas de vampiros.",
			Traits: []model.SpeciesTrait{
				{
					Name:        "Escalar como Aranha (Spider Climb)",
					Description: "Você possui deslocamento de Escalada igual ao seu deslocamento terrestre (10,5m). A partir do nível 3 de personagem, pode se mover por superfícies verticais e tetos mantendo as mãos totalmente livres.",
				},
				{
					Name:        "Vestígio dos Mortos-Vivos (Trace of Undeath)",
					Description: "Você tem Resistência a dano Necrótico.",
				},
				{
					Name:        "Mordida Vampírica (Vampiric Bite)",
					Description: "Ao desferir um Ataque Desarmado e causar dano, pode morder com suas presas, causando 1d4 + mod Constituição de dano Perfurante em vez do dano normal.\n\nAlém disso, ao morder uma criatura que não seja Construto nem Morto-Vivo, pode escolher um dos efeitos (usos = Bônus de Proficiência por Descanso Longo):\n• **Drenar (Drain):** Recupera Pontos de Vida iguais ao dano perfurante causado.\n• **Fortalecer (Strengthen):** Ganha um bônus igual ao dano perfurante causado no próximo teste de atributo ou jogada de ataque que fizer em até 1 minuto.",
				},
			},
		},
		{
			ID:           "hexblood",
			Name:         "Hexblood",
			NamePT:       "Sangue-de-Bruxa",
			Category:     "Ravenloft",
			Source:       "Ravenloft - The Horrors Within / VRGtR",
			CreatureType: "Fada (Fey)",
			Size:         "Médio (Medium) ou Pequeno (Small)",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Indivíduos infundidos com a magia feérica misteriosa de bruxas (Hags) ou rituais sombrios. Manifestam coroas de espinhos ou marcas grotescas e possuem clarividência assustadora e feitiçaria fascinante.",
			Traits: []model.SpeciesTrait{
				{
					Name:        "Fetiche Sobrenatural (Eerie Token)",
					Description: "Com uma Ação Bônus, você arranca sem dor uma mecha de cabelo, unha ou dente para criar um fetiche mágico que dura até o próximo Descanso Longo (1 uso por descanso longo):\n• **Mensagem Distante:** Com uma Ação Mágica, envia uma mensagem telepática de até 25 palavras para a criatura carregando o fetiche em até 16 km (10 milhas).\n• **Visão Remota:** Com uma Ação Mágica, você estende seus sentidos pelo fetiche por 1 minuto (se estiver a até 16 km), enxergando e ouvindo como se estivesse no local do fetiche. Ao terminar, o fetiche se destrói.",
				},
				{
					Name:        "Magia Hex (Hex Magic)",
					Description: "Você sempre tem preparadas as magias *Disfarçar-se (Disguise Self)* e *Bruxaria (Hex)*. Pode conjurar cada uma 1 vez sem gastar espaço de magia por Descanso Longo, ou conjurá-las normalmente com espaços de magia. O atributo de conjuração é Inteligência, Sabedoria ou Carisma (à sua escolha).",
				},
			},
		},
		{
			ID:           "lupin",
			Name:         "Lupin",
			NamePT:       "Lupino",
			Category:     "Ravenloft",
			Source:       "Ravenloft - The Horrors Within",
			CreatureType: "Humanoide (Humanoid)",
			Size:         "Médio (Medium) ou Pequeno (Small)",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Criaturas com pelagem espessa e focinho canino que emergem das Brumas de Ravenloft. Nascem do contato com lobisomens onde a maldição se estabilizou em forma híbrida permanente, concedendo fúria bestial e um uivo apavorante.",
			Traits: []model.SpeciesTrait{
				{
					Name:        "Bote Feral (Feral Pounce)",
					Description: "Seus Ataques Desarmados causam dano Cortante em vez de Concussão. Além disso, ao acertar uma criatura com Ataque Desarmado na sua Ação de Ataque, pode aplicar SIMULTANEAMENTE as opções de Dano E de Empurrão (Shove) — 1 vez por turno.",
				},
				{
					Name:        "Uivo Aterrorizante (Howl)",
					Description: "Com uma Ação Bônus, você solta um uivo aterrador. Cada criatura à sua escolha em até 4,5 metros (15 ft) deve passar numa salvaguarda de Sabedoria (CD 8 + mod Constituição + Bônus de Proficiência) ou sofre Desvantagem em todas as jogadas de ataque e salvaguardas até o início do seu próximo turno! Usos iguais ao Bônus de Proficiência por Descanso Longo.",
				},
				{
					Name:        "Instintos de Lobisomem (Werewolf Instincts)",
					Description: "Você tem proficiência na perícia Percepção, Furtividade ou Sobrevivência (à sua escolha).",
				},
			},
		},
		{
			ID:           "reborn",
			Name:         "Reborn",
			NamePT:       "Renascido",
			Category:     "Ravenloft",
			Source:       "Ravenloft - The Horrors Within / VRGtR",
			CreatureType: "Humanoide (Humanoid)",
			Size:         "Médio (Medium) ou Pequeno (Small)",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Pessoas que morreram mas inexplicavelmente retornaram à vida. Alguns trazem cicatrizes fatais e veias secas cinzentas; outros são construções de retalhos de corpos ou mentes aprisionadas em invólucros manufaturados.",
			Traits: []model.SpeciesTrait{
				{
					Name:        "Escapou da Morte (Escaped Death)",
					Description: "Você tem Vantagem em todas as Salvaguardas contra a Morte (Death Saving Throws).",
				},
				{
					Name:        "Imortal / Natureza Eterna (Everlasting)",
					Description: "Você não sofre níveis de Exaustão por desidratação, fome ou sufocamento. Não precisa dormir e magia não pode colocá-lo para dormir. Completa um Descanso Longo em apenas 4 horas em estado inativo e imóvel, permanecendo totalmente consciente.",
				},
				{
					Name:        "Conhecimento de Vidas Passadas (Knowledge from a Past Life)",
					Description: "Ganha proficiência em uma perícia à sua escolha. Além disso, quando falhar em um teste de atributo, pode rolar 1d6 e somar ao d20, potencialmente transformando a falha em sucesso! Usos iguais ao Bônus de Proficiência por Descanso Longo.",
				},
				{
					Name:        "Resistência Estranha (Strange Endurance)",
					Description: "Você tem Resistência a um dos seguintes tipos de dano (escolhido ao criar o personagem): Frio, Necrótico ou Veneno.",
				},
			},
		},

		// === D&D 2024 / PHB 2024 CORE SPECIES ===
		{
			ID:           "human",
			Name:         "Human",
			NamePT:       "Humano",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio ou Pequeno",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "Nenhuma",
			Overview:     "Inovadores, adaptáveis e ambiciosos, humanos são a espécie mais comum nos mundos de D&D, capazes de se destacar em qualquer vocação graças à sua versatilidade e tenacidade inata.",
			Traits: []model.SpeciesTrait{
				{Name: "Talento Versátil (Resourceful)", Description: "Você ganha Inspiração Heroica sempre que termina um Descanso Longo."},
				{Name: "Habilidoso (Skillful)", Description: "Ganha proficiência em uma perícia à sua escolha."},
				{Name: "Versátil (Versatile)", Description: "Ganha um Talento de Origem (Origin Feat) adicional à sua escolha."},
			},
		},
		{
			ID:           "elf",
			Name:         "Elf",
			NamePT:       "Elfo",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft) ou 36m (Drow)",
			Overview:     "Seres graciosos de longevidade mágica e conexão com o Reino Feérico, divididos em três linhagens principais: Alto Elfo (High Elf), Elfo da Floresta (Wood Elf) e Drow.",
			Traits: []model.SpeciesTrait{
				{Name: "Ancestralidade Feérica", Description: "Vantagem em salvaguardas para evitar ou encerrar a condição Enfeitiçado."},
				{Name: "Transe (Trance)", Description: "Não precisa dormir e magia não pode colocá-lo para dormir; completa descanso longo meditando por 4 horas."},
				{Name: "Sentidos Aguçados", Description: "Proficiência na perícia Percepção."},
				{Name: "Linhagem Élfica", Description: "Escolha Alto Elfo (truques de Mago e trocar truque no descanso), Elfo da Floresta (deslocamento 10,5m e magias de druida) ou Drow (Visão no escuro 36m e magias drow)."},
			},
		},
		{
			ID:           "dwarf",
			Name:         "Dwarf",
			NamePT:       "Anão",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "36 metros (120 ft)",
			Overview:     "Guerreiros e artesãos tenazes esculpidos na rocha das montanhas, com resistência sobre-humana a toxinas e vigor inabalável.",
			Traits: []model.SpeciesTrait{
				{Name: "Visão no Escuro Superior", Description: "Enxerga na escuridão até 36 metros (120 ft)."},
				{Name: "Resiliência dos Anões", Description: "Resistência a dano de Veneno e vantagem em salvaguardas contra a condição Envenenado."},
				{Name: "Robustez Anã (Dwarven Toughness)", Description: "Seu máximo de Pontos de Vida aumenta em 1 por nível de personagem."},
				{Name: "Afinidade com Rochas (Stonecunning)", Description: "Com Ação Bônus, ganha Sentido Sísmico (Tremorsense) de 18 metros na pedra por 10 minutos."},
			},
		},
		{
			ID:           "aasimar",
			Name:         "Aasimar",
			NamePT:       "Aasimar",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio ou Pequeno",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Mortais tocados pela centelha divina dos Planos Superiores, capazes de manifestar asas celestiais de luz ou auras radiantes de julgamento.",
			Traits: []model.SpeciesTrait{
				{Name: "Resistência Celestial", Description: "Resistência a dano Necrótico e Radiante."},
				{Name: "Mãos Que Curam (Healing Hands)", Description: "Como Ação Mágica, toca uma criatura para curar dados de d4 iguais ao seu Bônus de Proficiência (1x por Long Rest)."},
				{Name: "Revelação Celestial (Level 3)", Description: "Com Ação Bônus, assume forma celestial por 1 minuto: Asas Celestiais (voo), Manto Radiante (dano radiante ao redor) ou Sudário Necrótico (pavor e dano necrótico)."},
			},
		},
		{
			ID:           "dragonborn",
			Name:         "Dragonborn",
			NamePT:       "Draconato",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Descendentes nobres de dragões cromáticos e metálicos que herdam a respiração destrutiva e a resistência elemental de seus ancestrais.",
			Traits: []model.SpeciesTrait{
				{Name: "Arma de Sopro (Breath Weapon)", Description: "Substitui um dos ataques para exalar um cone de 4,5m ou linha de 9m de energia elemental (dano d10s aumentando por nível)."},
				{Name: "Resistência a Dano", Description: "Resistência ao tipo de dano associado à sua ancestralidade de dragão (Fogo, Frio, Relâmpago, Ácido ou Veneno)."},
				{Name: "Voo Dracônico (Level 5)", Description: "Manifesta asas espectrais temporárias para voar com velocidade igual ao deslocamento por 10 minutos (1x por Long Rest)."},
			},
		},
		{
			ID:           "gnome",
			Name:         "Gnome",
			NamePT:       "Gnomo",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Pequeno (Small)",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Inventores e místicos brilhantes com curiosidade insaciável e astúcia mágica capaz de resistir aos mais poderosos feitiços mentais.",
			Traits: []model.SpeciesTrait{
				{Name: "Esperteza Gnômica (Gnomish Cunning)", Description: "Vantagem em TODAS as salvaguardas de Inteligência, Sabedoria e Carisma contra magias!"},
				{Name: "Linhagem Gnômica", Description: "Escolha Gnomo da Floresta (truque Ilusão Menor e falar com animais) ou Gnomo das Rochas (truques Prestidigitação/Remendar e cria pequenos dispositivos mecânicos)."},
			},
		},
		{
			ID:           "goliath",
			Name:         "Goliath",
			NamePT:       "Golias",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio",
			Speed:        "10,5 metros (35 ft)",
			Darkvision:   "Nenhuma",
			Overview:     "Descendentes gigantescos de povos das montanhas que carregam o sangue dos Gigantes (Fogo, Gelo, Colina, Pedra, Tempestade ou Nuvem).",
			Traits: []model.SpeciesTrait{
				{Name: "Ancestralidade dos Gigantes", Description: "Escolha uma herança: Gigante da Nuvem (teletransporte Misty Step), Fogo (dano de fogo extra), Gelo (dano de gelo e lentidão), Colina (derruba deitado), Pedra (reduz dano com reação) ou Tempestade (relâmpago retaliatório)."},
				{Name: "Forma Grande (Large Form - Level 5)", Description: "Com Ação Bônus, cresce para tamanho Grande por 10 minutos, aumentando alcance e vantagem em testes de Força."},
			},
		},
		{
			ID:           "halfling",
			Name:         "Halfling",
			NamePT:       "Halfling",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Pequeno (Small)",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "Nenhuma",
			Overview:     "Pequenos aventureiros amantes do conforto mas repletos de coragem e uma sorte quase sobrenatural que desafia as probabilidades.",
			Traits: []model.SpeciesTrait{
				{Name: "Sortudo (Lucky)", Description: "Ao tirar um 1 natural em um d20 em jogada de ataque, teste de atributo ou salvaguarda, pode rolar novamente e deve usar o novo resultado!"},
				{Name: "Corajoso (Brave)", Description: "Vantagem em salvaguardas para evitar ou encerrar a condição Amedrontado."},
				{Name: "Agilidade Halfling", Description: "Pode se mover através do espaço de qualquer criatura que seja de tamanho maior que o seu."},
				{Name: "Furtividade Natural", Description: "Pode tentar se esconder mesmo quando estiver obscurecido apenas por uma criatura maior."},
			},
		},
		{
			ID:           "orc",
			Name:         "Orc",
			NamePT:       "Orc",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "36 metros (120 ft)",
			Overview:     "Guerreiros vigorosos e honrados dotados de resistência lendária e adrenalina explosiva que lhes permite avançar destemidamente no combate.",
			Traits: []model.SpeciesTrait{
				{Name: "Investida Adrenérgica (Adrenaline Rush)", Description: "Pode usar a Ação de Disparada (Dash) como Ação Bônus e ganha Pontos de Vida Temporários iguais ao Bônus de Proficiência (usos = PB por Long Rest)."},
				{Name: "Resistência Implacável (Relentless Endurance)", Description: "Quando for reduzido a 0 Pontos de Vida sem ser morto imediatamente, cai para 1 PV em vez disso (1x por Long Rest)."},
			},
		},
		{
			ID:           "tiefling",
			Name:         "Tiefling",
			NamePT:       "Tiefling",
			Category:     "Common",
			Source:       "PHB 2024",
			CreatureType: "Humanoide",
			Size:         "Médio ou Pequeno",
			Speed:        "9 metros (30 ft)",
			Darkvision:   "18 metros (60 ft)",
			Overview:     "Indivíduos marcados pela herança dos Planos Inferiores (Infernal, Abissal ou Ctônico), ostentando chifres, caudas e magia profana inata.",
			Traits: []model.SpeciesTrait{
				{Name: "Legado Infernal", Description: "Escolha Abissal (resistência a Veneno e magias venenosas), Ctônico (resistência a Necrótico e magias de morte) ou Infernal (resistência a Fogo e magias de fogo como Repreensão Infernal)."},
				{Name: "Taumaturgia", Description: "Aprende o truque Taumaturgia com carisma/int/sabedoria."},
			},
		},
	}
}



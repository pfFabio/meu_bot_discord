package character

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WizardStep enumera os passos da criação de ficha
type WizardStep int

const (
	StepName WizardStep = iota
	StepRace
	StepClass
	StepLevel
	StepAttributes
	StepCombatStats
	StepSaves
	StepSkills
	StepFeatures
	StepSpellcasting
	StepEquipment
	StepCompleted
)

// CreationSession representa uma sessão ativa de criação de ficha de um usuário
type CreationSession struct {
	UserID     string
	ChannelID  string
	Step       WizardStep
	Draft      *CharacterSheet
	LastActive time.Time
}

// WizardManager gerencia as sessões de criação passo a passo
type WizardManager struct {
	mu       sync.Mutex
	sessions map[string]*CreationSession // userID -> session
	service  Service
}

// NewWizardManager instancia o gerenciador do assistente interativo de ficha
func NewWizardManager(service Service) *WizardManager {
	return &WizardManager{
		sessions: make(map[string]*CreationSession),
		service:  service,
	}
}

// StartSession inicia uma nova sessão de criação para o usuário
func (m *WizardManager) StartSession(userID, channelID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &CreationSession{
		UserID:    userID,
		ChannelID: channelID,
		Step:      StepName,
		Draft: &CharacterSheet{
			ID:           fmt.Sprintf("sheet_%d", time.Now().UnixNano()),
			UserID:       userID,
			Level:        1,
			ArmorClass:   10,
			Speed:        "9m / 30ft",
			MaxHP:        10,
			CurrentHP:    10,
			SavingThrows: make(map[string]SavingThrow),
			Skills:       make(map[string]Skill),
			Spellcasting: Spellcasting{
				Slots: make(map[string]SpellSlotInfo),
			},
			CreatedAt: time.Now(),
		},
		LastActive: time.Now(),
	}

	m.sessions[userID] = session

	return "🧙‍♂️ **Bem-vindo ao Modo de Criação de Ficha!**\n" +
		"Vou montar sua ficha de RPG passo a passo junto com você, criando todos os atributos, perícias, magias, PV e recursos no seu objeto de ficha.\n" +
		"*(A qualquer momento digite `cancelar` para abortar ou `voltar` para corrigir o passo anterior.)*\n\n" +
		"▶ **Passo 1/11: Qual é o nome do seu personagem?**", nil
}

// CancelSession cancela a criação de ficha
func (m *WizardManager) CancelSession(userID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[userID]; ok {
		delete(m.sessions, userID)
		return true
	}
	return false
}

// HasActiveSession verifica se o usuário possui sessão de criação em andamento
func (m *WizardManager) HasActiveSession(userID, channelID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[userID]
	if !ok {
		return false
	}
	// Expira após 20 minutos de inatividade
	if time.Since(sess.LastActive) > 20*time.Minute {
		delete(m.sessions, userID)
		return false
	}
	return sess.ChannelID == channelID
}

// Roll4d6DropLowest rola 4d6 e descarta o menor (método canônico de criação de atributos do D&D)
func Roll4d6DropLowest() (int, string) {
	rolls := make([]int, 4)
	for i := 0; i < 4; i++ {
		rolls[i] = rand.Intn(6) + 1
	}
	sort.Ints(rolls)
	sum := rolls[1] + rolls[2] + rolls[3]
	detail := fmt.Sprintf("[%d, %d, %d, ~~%d~~] = **%d**", rolls[3], rolls[2], rolls[1], rolls[0], sum)
	return sum, detail
}

// ProcessInput processa a mensagem de texto do usuário e avança o assistente
func (m *WizardManager) ProcessInput(userID, input string) (response string, finishedSheet *CharacterSheet, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[userID]
	if !ok {
		return "", nil, fmt.Errorf("nenhuma sessão ativa")
	}

	trimmed := strings.TrimSpace(input)
	sess.LastActive = time.Now()

	// Comandos de controle
	if strings.EqualFold(trimmed, "cancelar") || strings.EqualFold(trimmed, "sair") {
		delete(m.sessions, userID)
		return "❌ Criação de ficha cancelada. Nenhum dado foi salvo.", nil, nil
	}

	if strings.EqualFold(trimmed, "voltar") {
		if sess.Step > StepName {
			sess.Step--
			return fmt.Sprintf("⬅️ Voltando para o passo anterior.\n%s", m.promptForStep(sess)), nil, nil
		}
		return "Você já está no primeiro passo!", nil, nil
	}

	// Processa o passo atual
	switch sess.Step {
	case StepName:
		if len(trimmed) < 2 {
			return "Por favor, digite um nome válido para o seu personagem (mínimo 2 letras):", nil, nil
		}
		sess.Draft.Name = trimmed
		sess.Step = StepRace
		return fmt.Sprintf("✅ Nome definido: **%s**!\n\n▶ **Passo 2/11: Qual é a raça ou espécie de %s?**\n*(ex: Humano, Anão, Elfo, Owlin, Loxodon, Dhampir, Lupin, Draconato...)*", sess.Draft.Name, sess.Draft.Name), nil, nil

	case StepRace:
		sess.Draft.Race = trimmed
		sess.Step = StepClass
		return fmt.Sprintf("✅ Espécie: **%s**!\n\n▶ **Passo 3/11: Qual é a classe e subclasse de %s?**\n*(ex: Guerreiro Campeão, Clérigo da Luz, Ladino Assassino, Bardo dos Espíritos, Monge da Misericórdia...)*", sess.Draft.Race, sess.Draft.Name), nil, nil

	case StepClass:
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 {
			sess.Draft.Class = parts[0]
			sess.Draft.Subclass = strings.Join(parts[1:], " ")
		} else {
			sess.Draft.Class = trimmed
		}
		sess.Step = StepLevel
		return fmt.Sprintf("✅ Classe: **%s** %s!\n\n▶ **Passo 4/11: Qual é o nível de %s?**\n*(Digite um número de 1 a 20)*", sess.Draft.Class, sess.Draft.Subclass, sess.Draft.Name), nil, nil

	case StepLevel:
		lvl, err := strconv.Atoi(trimmed)
		if err != nil || lvl < 1 || lvl > 20 {
			return "⚠️ Por favor, digite um nível válido entre **1** e **20**:", nil, nil
		}
		sess.Draft.Level = lvl
		sess.Draft.ProficiencyBonus = ProficiencyBonusByLevel(lvl)
		sess.Step = StepAttributes
		return fmt.Sprintf("✅ Nível **%d** (Bônus de Proficiência: **+%d**)!\n\n"+
			"▶ **Passo 5/11: Quais são os 6 atributos principais de %s?**\n"+
			"Você pode digitar os 6 números em ordem: `FOR DES CON INT SAB CAR` (ex: `16 10 16 10 14 8`)\n"+
			"🎲 Ou digite **`rolar`** se quiser que eu role 4d6 (descartando o menor) para você!", sess.Draft.Level, sess.Draft.ProficiencyBonus, sess.Draft.Name), nil, nil

	case StepAttributes:
		if strings.EqualFold(trimmed, "rolar") || strings.EqualFold(trimmed, "roll") {
			s1, d1 := Roll4d6DropLowest()
			s2, d2 := Roll4d6DropLowest()
			s3, d3 := Roll4d6DropLowest()
			s4, d4 := Roll4d6DropLowest()
			s5, d5 := Roll4d6DropLowest()
			s6, d6 := Roll4d6DropLowest()

			sess.Draft.Attributes.STR.Score = s1
			sess.Draft.Attributes.DEX.Score = s2
			sess.Draft.Attributes.CON.Score = s3
			sess.Draft.Attributes.INT.Score = s4
			sess.Draft.Attributes.WIS.Score = s5
			sess.Draft.Attributes.CHA.Score = s6
			sess.Draft.CalculateDerived()

			rollReport := fmt.Sprintf("🎲 **Rolagens de 4d6k3 realizadas:**\n"+
				"• **FOR:** %d %s\n"+
				"• **DES:** %d %s\n"+
				"• **CON:** %d %s\n"+
				"• **INT:** %d %s\n"+
				"• **SAB:** %d %s\n"+
				"• **CAR:** %d %s\n\n",
				s1, d1, s2, d2, s3, d3, s4, d4, s5, d5, s6, d6)

			sess.Step = StepCombatStats
			return rollReport + fmt.Sprintf("▶ **Passo 6/11: Qual é o HP Máximo e a Classe de Armadura (CA) de %s?**\n*(ex: `24 HP, 15 CA` ou simplesmente `24 15`)*", sess.Draft.Name), nil, nil
		}

		// Tenta extrair 6 números
		reNums := regexp.MustCompile(`\b\d+\b`)
		matches := reNums.FindAllString(trimmed, -1)
		if len(matches) < 6 {
			return "⚠️ Preciso de 6 valores numéricos para `FOR DES CON INT SAB CAR` (ex: `14 16 14 10 12 8`) ou digite `rolar`:", nil, nil
		}

		nums := make([]int, 6)
		for i := 0; i < 6; i++ {
			n, _ := strconv.Atoi(matches[i])
			if n < 1 || n > 30 {
				return fmt.Sprintf("⚠️ Valor %d inválido. Atributos devem ser entre 1 e 30.", n), nil, nil
			}
			nums[i] = n
		}

		sess.Draft.Attributes.STR.Score = nums[0]
		sess.Draft.Attributes.DEX.Score = nums[1]
		sess.Draft.Attributes.CON.Score = nums[2]
		sess.Draft.Attributes.INT.Score = nums[3]
		sess.Draft.Attributes.WIS.Score = nums[4]
		sess.Draft.Attributes.CHA.Score = nums[5]
		sess.Draft.CalculateDerived()

		sess.Step = StepCombatStats
		return fmt.Sprintf("✅ Atributos definidos: FOR %d (%+d) | DES %d (%+d) | CON %d (%+d) | INT %d (%+d) | SAB %d (%+d) | CAR %d (%+d)\n\n"+
			"▶ **Passo 6/11: Qual é o HP Máximo e a Classe de Armadura (CA) de %s?**\n*(ex: `24 HP, 15 CA` ou simplesmente `24 15`)*",
			nums[0], sess.Draft.Attributes.STR.Modifier,
			nums[1], sess.Draft.Attributes.DEX.Modifier,
			nums[2], sess.Draft.Attributes.CON.Modifier,
			nums[3], sess.Draft.Attributes.INT.Modifier,
			nums[4], sess.Draft.Attributes.WIS.Modifier,
			nums[5], sess.Draft.Attributes.CHA.Modifier,
			sess.Draft.Name), nil, nil

	case StepCombatStats:
		reNums := regexp.MustCompile(`\b\d+\b`)
		matches := reNums.FindAllString(trimmed, -1)
		hp := 10
		ca := 10
		if len(matches) >= 2 {
			hp, _ = strconv.Atoi(matches[0])
			ca, _ = strconv.Atoi(matches[1])
		} else if len(matches) == 1 {
			hp, _ = strconv.Atoi(matches[0])
		}

		sess.Draft.MaxHP = hp
		sess.Draft.CurrentHP = hp
		sess.Draft.ArmorClass = ca
		sess.Step = StepSaves

		return fmt.Sprintf("✅ Pontos de Vida: **%d PV** | Classe de Armadura: **%d CA**!\n\n"+
			"▶ **Passo 7/11: Em quais Salvaguardas (Testes de Resistência) %s tem proficiência?**\n"+
			"*(ex: `Constituição e Sabedoria` ou `FOR, CON` ou `nenhuma`)*", sess.Draft.MaxHP, sess.Draft.ArmorClass, sess.Draft.Name), nil, nil

	case StepSaves:
		low := strings.ToLower(trimmed)
		if low != "nenhuma" && low != "nenhum" && low != "-" {
			for ak, name := range map[string]string{
				"str": "força", "dex": "destreza", "con": "constituição",
				"int": "inteligência", "wis": "sabedoria", "cha": "carisma",
			} {
				short := ak
				if strings.Contains(low, name) || strings.Contains(low, short) || strings.Contains(low, name[:3]) {
					mod := sess.Draft.GetAttributeModifier(ak) + sess.Draft.ProficiencyBonus
					sess.Draft.SavingThrows[ak] = SavingThrow{
						Modifier:   mod,
						Proficient: true,
					}
				}
			}
		}
		sess.Step = StepSkills
		return fmt.Sprintf("✅ Salvaguardas registradas!\n\n"+
			"▶ **Passo 8/11: Quais Perícias %s é treinado/proficiente?**\n"+
			"*(ex: `Atletismo, Percepção, Intuição, Furtividade` ou `nenhuma`)*\n"+
			"*(Dica: Se tiver Expertise em alguma, coloque `(Expertise)` ao lado)*", sess.Draft.Name), nil, nil

	case StepSkills:
		low := strings.ToLower(trimmed)
		if low != "nenhuma" && low != "nenhum" && low != "-" {
			allSkills := []struct {
				Name string
				Attr string
			}{
				{"Acrobacia", "DEX"}, {"Adestrar Animais", "WIS"}, {"Arcanismo", "INT"},
				{"Atletismo", "FOR"}, {"Enganação", "CHA"}, {"História", "INT"},
				{"Intuição", "WIS"}, {"Intimidação", "CHA"}, {"Investigação", "INT"},
				{"Medicina", "WIS"}, {"Natureza", "INT"}, {"Percepção", "WIS"},
				{"Atuação", "CHA"}, {"Persuasão", "CHA"}, {"Religião", "INT"},
				{"Prestidigitação", "DEX"}, {"Furtividade", "DEX"}, {"Sobrevivência", "WIS"},
			}

			for _, sk := range allSkills {
				skLow := strings.ToLower(sk.Name)
				if strings.Contains(low, skLow) {
					isExp := strings.Contains(low, skLow+" (expertise)") || strings.Contains(low, skLow+" expertise")
					attrMod := sess.Draft.GetAttributeModifier(sk.Attr)
					bonus := sess.Draft.ProficiencyBonus
					if isExp {
						bonus *= 2
					}
					sess.Draft.Skills[sk.Name] = Skill{
						Modifier:   attrMod + bonus,
						Proficient: true,
						Expertise:  isExp,
						Attribute:  sk.Attr,
					}
				}
			}
		}
		sess.Draft.CalculateDerived()
		sess.Step = StepFeatures
		return fmt.Sprintf("✅ Perícias registradas! Percepção Passiva: **%d**.\n\n"+
			"▶ **Passo 9/11: Quais Habilidades ou Características de classe e raça %s possui?**\n"+
			"*(ex: `Visão no Escuro, Ataque Furtivo, Canalizar Divindade (2 usos), Pontos de Sorte (2 usos)...`)*\n"+
			"*(Separe por vírgula. Se tiver quantidade de usos, coloque entre parênteses)*", sess.Draft.PassivePerception, sess.Draft.Name), nil, nil

	case StepFeatures:
		parts := strings.Split(trimmed, ",")
		for _, p := range parts {
			fName := strings.TrimSpace(p)
			if fName == "" || strings.EqualFold(fName, "nenhuma") {
				continue
			}

			// Procura se tem número de usos (ex: "Canalizar Divindade (2 usos)")
			reUsos := regexp.MustCompile(`\((\d+)\s*usos?\)`)
			match := reUsos.FindStringSubmatch(fName)
			if len(match) == 2 {
				uses, _ := strconv.Atoi(match[1])
				cleanName := strings.TrimSpace(reUsos.ReplaceAllString(fName, ""))
				sess.Draft.LimitedUses = append(sess.Draft.LimitedUses, LimitedResource{
					Name:        cleanName,
					MaxUses:     uses,
					CurrentUses: uses,
					ResetOn:     "Descanso Longo",
				})
				sess.Draft.Features = append(sess.Draft.Features, Feature{
					Name:   cleanName,
					Source: "Classe/Espécie",
				})
			} else {
				sess.Draft.Features = append(sess.Draft.Features, Feature{
					Name:   fName,
					Source: "Classe/Espécie",
				})
			}
		}
		sess.Step = StepSpellcasting
		return fmt.Sprintf("✅ Características registradas (%d habilidades / recursos adicionados)!\n\n"+
			"▶ **Passo 10/11: %s é um conjurador de magias?**\n"+
			"Se sim, informe o atributo de conjuração, seus truques, magias e espaços diários de magia.\n"+
			"*(ex: `Sabedoria. Truques: Sacred Flame, Light. Nível 1 (4 slots): Bless, Cure Wounds. Nível 2 (2 slots): Spiritual Weapon.`)*\n"+
			"*(Se não conjura magias, digite apenas `nenhuma`)*", len(sess.Draft.Features), sess.Draft.Name), nil, nil

	case StepSpellcasting:
		low := strings.ToLower(trimmed)
		if low != "nenhuma" && low != "nenhum" && low != "-" && low != "não" && low != "nao" {
			// Atributo de conjuração
			if strings.Contains(low, "sabedoria") || strings.Contains(low, "wis") {
				sess.Draft.Spellcasting.Ability = "Sabedoria"
			} else if strings.Contains(low, "inteligência") || strings.Contains(low, "inteligencia") || strings.Contains(low, "int") {
				sess.Draft.Spellcasting.Ability = "Inteligência"
			} else if strings.Contains(low, "carisma") || strings.Contains(low, "cha") {
				sess.Draft.Spellcasting.Ability = "Carisma"
			}

			// Procura slots (ex: Nível 1 (4 slots), Nível 2 (2 slots))
			reSlot := regexp.MustCompile(`(?i)(?:nível|nivel|lvl)\s*(\d+)\s*\((\d+)\s*slots?\)`)
			matchesSlot := reSlot.FindAllStringSubmatch(trimmed, -1)
			for _, m := range matchesSlot {
				lvl, _ := strconv.Atoi(m[1])
				sMax, _ := strconv.Atoi(m[2])
				sess.Draft.Spellcasting.Slots[fmt.Sprintf("%d", lvl)] = SpellSlotInfo{
					Level: lvl,
					Max:   sMax,
					Used:  0,
				}
			}

			// Procura Truques e Magias
			if strings.Contains(low, "truques:") {
				parts := strings.Split(trimmed, "Truques:")
				if len(parts) > 1 {
					sub := strings.Split(parts[1], "Nível")[0]
					for _, sp := range strings.Split(sub, ",") {
						spName := strings.TrimSpace(sp)
						if spName != "" {
							sess.Draft.Spellcasting.Spells = append(sess.Draft.Spellcasting.Spells, CharacterSpell{
								Name:     spName,
								Level:    0,
								Prepared: true,
							})
						}
					}
				}
			}

			sess.Draft.CalculateDerived()
		}
		sess.Step = StepEquipment
		return fmt.Sprintf("✅ Grimório e magia configurados!\n\n"+
			"▶ **Passo 11/11: Quais armas, ataques e equipamentos principais %s carrega?**\n"+
			"*(ex: `Espada Longa (+5, 1d8+3 Cortante), Arco Curto (+5, 1d6+3 Perfurante), Cota de Malha, Mochila, 25 PO`)*\n"+
			"*(Se quiser usar equipamento padrão, digite `padrao`)*", sess.Draft.Name), nil, nil

	case StepEquipment:
		low := strings.ToLower(trimmed)
		if low != "padrao" && low != "padrão" && low != "nenhum" && low != "-" {
			// Procura ouro (ex: 25 PO, 50 GP)
			reGold := regexp.MustCompile(`(?i)(\d+)\s*(po|gp|ouro)`)
			if m := reGold.FindStringSubmatch(trimmed); len(m) == 3 {
				gp, _ := strconv.Atoi(m[1])
				sess.Draft.Coins.GP = gp
			}

			// Separa itens e armas por vírgula
			items := strings.Split(trimmed, ",")
			for _, it := range items {
				itTrim := strings.TrimSpace(it)
				if itTrim == "" || reGold.MatchString(itTrim) {
					continue
				}

				// Verifica se parece um ataque com bônus e dano: ex: "Espada Curta (+5, 1d6+3 Piercing)"
				if strings.Contains(itTrim, "(") && strings.Contains(itTrim, ")") {
					parts := strings.Split(itTrim, "(")
					wName := strings.TrimSpace(parts[0])
					inner := strings.TrimSuffix(parts[1], ")")
					innerParts := strings.Split(inner, ",")
					atkBonus := "+0"
					dmg := ""
					if len(innerParts) >= 1 {
						atkBonus = strings.TrimSpace(innerParts[0])
					}
					if len(innerParts) >= 2 {
						dmg = strings.TrimSpace(innerParts[1])
					}
					sess.Draft.Attacks = append(sess.Draft.Attacks, CharacterAttack{
						Name:        wName,
						AttackBonus: atkBonus,
						Damage:      dmg,
					})
					sess.Draft.Inventory = append(sess.Draft.Inventory, InventoryItem{
						Name:     wName,
						Quantity: 1,
					})
				} else {
					sess.Draft.Inventory = append(sess.Draft.Inventory, InventoryItem{
						Name:     itTrim,
						Quantity: 1,
					})
				}
			}
		}

		// Conclui a criação da ficha
		sess.Draft.CalculateDerived()
		if err := m.service.Save(sess.Draft); err != nil {
			return "", nil, fmt.Errorf("falha ao salvar ficha: %w", err)
		}
		_ = m.service.SetActive(userID, sess.Draft.ID)

		finishedSheet = sess.Draft
		delete(m.sessions, userID)

		return fmt.Sprintf("🎉 **Ficha de %s criada com sucesso no objeto 'ficha'!**\n"+
			"Todos os atributos, proficiências, HP, magias e equipamentos foram calculados e armazenados com segurança no Discord.", finishedSheet.Name), finishedSheet, nil
	}

	return "", nil, nil
}

func (m *WizardManager) promptForStep(sess *CreationSession) string {
	switch sess.Step {
	case StepName:
		return "▶ **Passo 1/11: Qual é o nome do seu personagem?**"
	case StepRace:
		return fmt.Sprintf("▶ **Passo 2/11: Qual é a raça ou espécie de %s?**", sess.Draft.Name)
	case StepClass:
		return fmt.Sprintf("▶ **Passo 3/11: Qual é a classe e subclasse de %s?**", sess.Draft.Name)
	case StepLevel:
		return fmt.Sprintf("▶ **Passo 4/11: Qual é o nível de %s?** (1 a 20)", sess.Draft.Name)
	case StepAttributes:
		return "▶ **Passo 5/11: Digite os 6 atributos (`FOR DES CON INT SAB CAR`) ou `rolar`:**"
	case StepCombatStats:
		return fmt.Sprintf("▶ **Passo 6/11: Qual é o HP Máximo e a CA de %s?** (ex: `24 HP, 15 CA`)", sess.Draft.Name)
	case StepSaves:
		return fmt.Sprintf("▶ **Passo 7/11: Em quais Salvaguardas %s tem proficiência?**", sess.Draft.Name)
	case StepSkills:
		return fmt.Sprintf("▶ **Passo 8/11: Quais Perícias %s é treinado?**", sess.Draft.Name)
	case StepFeatures:
		return fmt.Sprintf("▶ **Passo 9/11: Quais características e habilidades %s possui?**", sess.Draft.Name)
	case StepSpellcasting:
		return fmt.Sprintf("▶ **Passo 10/11: %s conjura magias? Informe truques, magias e slots diários ou `nenhuma`:**", sess.Draft.Name)
	case StepEquipment:
		return fmt.Sprintf("▶ **Passo 11/11: Quais armas e equipamentos %s carrega?**", sess.Draft.Name)
	}
	return ""
}

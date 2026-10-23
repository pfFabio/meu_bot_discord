package character

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// AttributeScore representa a pontuação e modificador de um atributo
type AttributeScore struct {
	Score    int `json:"score"`
	Modifier int `json:"modifier"`
}

// AbilityScores representa os 6 atributos principais do D&D
type AbilityScores struct {
	STR AttributeScore `json:"str"` // Força
	DEX AttributeScore `json:"dex"` // Destreza
	CON AttributeScore `json:"con"` // Constituição
	INT AttributeScore `json:"int"` // Inteligência
	WIS AttributeScore `json:"wis"` // Sabedoria
	CHA AttributeScore `json:"cha"` // Carisma
}

// SavingThrow representa uma salvaguarda (teste de resistência)
type SavingThrow struct {
	Modifier   int  `json:"modifier"`
	Proficient bool `json:"proficient"`
}

// Skill representa uma perícia treinada ou não
type Skill struct {
	Modifier   int    `json:"modifier"`
	Proficient bool   `json:"proficient"`
	Expertise  bool   `json:"expertise"`
	Attribute  string `json:"attribute"` // FOR, DES, CON, INT, SAB, CAR
}

// Proficiencies agrupa proficiências em armas, armaduras, ferramentas e idiomas
type Proficiencies struct {
	Armor     []string `json:"armor"`
	Weapons   []string `json:"weapons"`
	Tools     []string `json:"tools"`
	Languages []string `json:"languages"`
}

// Feature representa uma característica ou traço de raça/classe
type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"` // Raça, Classe, Antecedente, Talento
}

// LimitedResource representa um recurso com usos limitados diários ou por descanso
type LimitedResource struct {
	Name        string `json:"name"`
	MaxUses     int    `json:"max_uses"`
	CurrentUses int    `json:"current_uses"`
	ResetOn     string `json:"reset_on"` // "Descanso Curto", "Descanso Longo"
}

// SpellSlotInfo representa os espaços de magia de determinado nível
type SpellSlotInfo struct {
	Level int `json:"level"`
	Max   int `json:"max"`
	Used  int `json:"used"`
}

// CharacterSpell representa uma magia conhecida ou preparada na ficha
type CharacterSpell struct {
	Name        string `json:"name"`
	Level       int    `json:"level"` // 0 = Truque
	School      string `json:"school,omitempty"`
	Prepared    bool   `json:"prepared"`
	Description string `json:"description,omitempty"`
}

// Spellcasting agrupa as informações de conjuração de magias do personagem
type Spellcasting struct {
	Ability     string                   `json:"ability"`      // Sabedoria, Inteligência, Carisma
	SaveDC      int                      `json:"save_dc"`      // CD para resistir
	AttackBonus int                      `json:"attack_bonus"` // Bônus de ataque mágico
	Slots       map[string]SpellSlotInfo `json:"slots"`        // Chaves "1" a "9"
	Spells      []CharacterSpell         `json:"spells"`       // Lista de magias
}

// CharacterAttack representa uma arma ou ataque físico/mágico
type CharacterAttack struct {
	Name        string `json:"name"`
	AttackBonus string `json:"attack_bonus"` // ex: "+5"
	Damage      string `json:"damage"`       // ex: "1d8+3 Cortante"
	DamageType  string `json:"damage_type,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// Coins armazena as moedas do personagem
type Coins struct {
	CP int `json:"cp"` // Cobre
	SP int `json:"sp"` // Prata
	EP int `json:"ep"` // Electro
	GP int `json:"gp"` // Ouro
	PP int `json:"pp"` // Platina
}

// InventoryItem representa um item no inventário
type InventoryItem struct {
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	Weight      string `json:"weight,omitempty"`
	Description string `json:"description,omitempty"`
}

// DeathSaves representa os testes contra a morte
type DeathSaves struct {
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

// CharacterSheet é o objeto completo e adaptativo que armazena a ficha de RPG
type CharacterSheet struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id"`
	GuildID          string                 `json:"guild_id,omitempty"`
	Name             string                 `json:"name"`
	PlayerName       string                 `json:"player_name,omitempty"`
	Class            string                 `json:"class"`
	Subclass         string                 `json:"subclass,omitempty"`
	Level            int                    `json:"level"`
	Race             string                 `json:"race"`
	Background       string                 `json:"background,omitempty"`
	Alignment        string                 `json:"alignment,omitempty"`
	XP               int                    `json:"xp"`
	Inspiration      bool                   `json:"inspiration"`
	ProficiencyBonus int                    `json:"proficiency_bonus"`
	Attributes       AbilityScores          `json:"attributes"`
	ArmorClass       int                    `json:"armor_class"`
	Initiative       int                    `json:"initiative"`
	Speed            string                 `json:"speed"`
	MaxHP            int                    `json:"max_hp"`
	CurrentHP        int                    `json:"current_hp"`
	TempHP           int                    `json:"temp_hp"`
	HitDice          string                 `json:"hit_dice"`
	HitDiceCurrent   int                    `json:"hit_dice_current"`
	DeathSaves       DeathSaves             `json:"death_saves"`
	SavingThrows     map[string]SavingThrow `json:"saving_throws"`
	Skills           map[string]Skill       `json:"skills"`
	PassivePerception int                   `json:"passive_perception"`
	PassiveInsight    int                   `json:"passive_insight"`
	PassiveInvestigation int               `json:"passive_investigation"`
	Proficiencies    Proficiencies          `json:"proficiencies"`
	Features         []Feature              `json:"features"`
	LimitedUses      []LimitedResource      `json:"limited_uses"`
	Spellcasting     Spellcasting           `json:"spellcasting"`
	Attacks          []CharacterAttack      `json:"attacks"`
	Coins            Coins                  `json:"coins"`
	Inventory        []InventoryItem        `json:"inventory"`
	AdvantageMode    string                 `json:"advantage_mode,omitempty"` // "", "normal", "advantage", "disadvantage"
	AvatarURL        string                 `json:"avatar_url,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CalculateModifier calcula o modificador a partir de um valor de atributo (ex: 10 -> 0, 14 -> +2)
func CalculateModifier(score int) int {
	return int(math.Floor(float64(score-10) / 2.0))
}

// ProficiencyBonusByLevel calcula o bônus de proficiência base do D&D para o nível
func ProficiencyBonusByLevel(level int) int {
	if level <= 4 {
		return 2
	} else if level <= 8 {
		return 3
	} else if level <= 12 {
		return 4
	} else if level <= 16 {
		return 5
	}
	return 6
}

// CalculateDerived recalcula modificadores e estatísticas derivadas da ficha
func (c *CharacterSheet) CalculateDerived() {
	if c.Level < 1 {
		c.Level = 1
	}
	if c.ProficiencyBonus <= 0 {
		c.ProficiencyBonus = ProficiencyBonusByLevel(c.Level)
	}

	// Recalcula modificadores dos atributos
	c.Attributes.STR.Modifier = CalculateModifier(c.Attributes.STR.Score)
	c.Attributes.DEX.Modifier = CalculateModifier(c.Attributes.DEX.Score)
	c.Attributes.CON.Modifier = CalculateModifier(c.Attributes.CON.Score)
	c.Attributes.INT.Modifier = CalculateModifier(c.Attributes.INT.Score)
	c.Attributes.WIS.Modifier = CalculateModifier(c.Attributes.WIS.Score)
	c.Attributes.CHA.Modifier = CalculateModifier(c.Attributes.CHA.Score)

	// Iniciativa padrão = Modificador de Destreza
	if c.Initiative == 0 {
		c.Initiative = c.Attributes.DEX.Modifier
	}

	// Percepção passiva padrão = 10 + Modificador de Sabedoria (+ proficiência se tiver)
	if c.PassivePerception <= 0 {
		base := 10 + c.Attributes.WIS.Modifier
		if s, ok := c.Skills["Percepção"]; ok && s.Proficient {
			base += c.ProficiencyBonus
			if s.Expertise {
				base += c.ProficiencyBonus
			}
		}
		c.PassivePerception = base
	}

	// CD e Ataque de Magia
	if c.Spellcasting.Ability != "" {
		spellMod := 0
		switch strings.ToLower(c.Spellcasting.Ability) {
		case "sabedoria", "wis":
			spellMod = c.Attributes.WIS.Modifier
		case "inteligência", "inteligencia", "int":
			spellMod = c.Attributes.INT.Modifier
		case "carisma", "cha":
			spellMod = c.Attributes.CHA.Modifier
		}
		if c.Spellcasting.SaveDC <= 0 {
			c.Spellcasting.SaveDC = 8 + c.ProficiencyBonus + spellMod
		}
		if c.Spellcasting.AttackBonus == 0 {
			c.Spellcasting.AttackBonus = c.ProficiencyBonus + spellMod
		}
	}

	if c.HitDice == "" {
		c.HitDice = fmt.Sprintf("%dd8", c.Level)
		c.HitDiceCurrent = c.Level
	}

	c.UpdatedAt = time.Now()
}

// TakeDamage aplica dano ao personagem, consumindo primeiro TempHP e depois CurrentHP
func (c *CharacterSheet) TakeDamage(amount int) (hpLost int, isDown bool) {
	if amount <= 0 {
		return 0, c.CurrentHP <= 0
	}

	remaining := amount

	// 1. Desconta de PV temporários primeiro
	if c.TempHP > 0 {
		if c.TempHP >= remaining {
			c.TempHP -= remaining
			remaining = 0
		} else {
			remaining -= c.TempHP
			c.TempHP = 0
		}
	}

	// 2. Desconta de PV atuais
	if remaining > 0 {
		hpLost = remaining
		if hpLost > c.CurrentHP {
			hpLost = c.CurrentHP
		}
		c.CurrentHP -= remaining
		if c.CurrentHP < 0 {
			c.CurrentHP = 0
		}
	}

	c.UpdatedAt = time.Now()
	return hpLost, c.CurrentHP == 0
}

// Heal recupera pontos de vida respeitando o HP Máximo
func (c *CharacterSheet) Heal(amount int) int {
	if amount <= 0 || c.CurrentHP >= c.MaxHP {
		return 0
	}

	prev := c.CurrentHP
	c.CurrentHP += amount
	if c.CurrentHP > c.MaxHP {
		c.CurrentHP = c.MaxHP
	}

	// Se recuperou vida e estava com 0 HP, zera falhas e sucessos de morte
	if prev == 0 && c.CurrentHP > 0 {
		c.DeathSaves.Successes = 0
		c.DeathSaves.Failures = 0
	}

	c.UpdatedAt = time.Now()
	return c.CurrentHP - prev
}

// SetTemporaryHP define pontos de vida temporários (substitui apenas se o novo valor for maior)
func (c *CharacterSheet) SetTemporaryHP(amount int) {
	if amount > c.TempHP {
		c.TempHP = amount
		c.UpdatedAt = time.Now()
	}
}

// UseSpellSlot consome um espaço de magia de determinado nível (1 a 9)
func (c *CharacterSheet) UseSpellSlot(level int) error {
	if level < 1 || level > 9 {
		return fmt.Errorf("nível de magia inválido: %d (deve ser entre 1 e 9)", level)
	}
	key := fmt.Sprintf("%d", level)
	slot, exists := c.Spellcasting.Slots[key]
	if !exists || slot.Max <= 0 {
		return fmt.Errorf("o personagem não possui espaços de magia de nível %d", level)
	}
	if slot.Used >= slot.Max {
		return fmt.Errorf("todos os espaços de magia de nível %d já foram gastos (%d/%d)", level, slot.Used, slot.Max)
	}

	slot.Used++
	c.Spellcasting.Slots[key] = slot
	c.UpdatedAt = time.Now()
	return nil
}

// RestoreSpellSlot restaura um espaço de magia de determinado nível
func (c *CharacterSheet) RestoreSpellSlot(level int) error {
	if level < 1 || level > 9 {
		return fmt.Errorf("nível de magia inválido: %d", level)
	}
	key := fmt.Sprintf("%d", level)
	slot, exists := c.Spellcasting.Slots[key]
	if !exists {
		return fmt.Errorf("o personagem não possui espaços de nível %d", level)
	}
	if slot.Used <= 0 {
		return fmt.Errorf("os espaços de nível %d já estão totalmente disponíveis", level)
	}

	slot.Used--
	c.Spellcasting.Slots[key] = slot
	c.UpdatedAt = time.Now()
	return nil
}

// UseResource gasta 1 uso de um recurso limitado diário (ex: Canalizar Divindade, Pontos de Sorte)
func (c *CharacterSheet) UseResource(resourceName string) error {
	for i, res := range c.LimitedUses {
		if strings.EqualFold(res.Name, resourceName) {
			if res.CurrentUses <= 0 {
				return fmt.Errorf("recurso '%s' esgotado! (%d/%d)", res.Name, res.CurrentUses, res.MaxUses)
			}
			c.LimitedUses[i].CurrentUses--
			c.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("recurso limitado '%s' não encontrado na ficha", resourceName)
}

// RestoreResource recupera 1 uso de um recurso limitado
func (c *CharacterSheet) RestoreResource(resourceName string) error {
	for i, res := range c.LimitedUses {
		if strings.EqualFold(res.Name, resourceName) {
			if res.CurrentUses >= res.MaxUses {
				return fmt.Errorf("recurso '%s' já está com usos máximos (%d/%d)", res.Name, res.CurrentUses, res.MaxUses)
			}
			c.LimitedUses[i].CurrentUses++
			c.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("recurso limitado '%s' não encontrado na ficha", resourceName)
}

// LongRest executa um Descanso Longo (recupera todo o HP, todos os slots de magia e recursos diários)
func (c *CharacterSheet) LongRest() {
	c.CurrentHP = c.MaxHP
	c.TempHP = 0
	c.DeathSaves.Successes = 0
	c.DeathSaves.Failures = 0

	// Restaura dados de vida (metade do total, mínimo de 1)
	diceRecovery := int(math.Max(1, math.Floor(float64(c.Level)/2.0)))
	c.HitDiceCurrent += diceRecovery
	if c.HitDiceCurrent > c.Level {
		c.HitDiceCurrent = c.Level
	}

	// Restaura todos os slots de magia
	for k, slot := range c.Spellcasting.Slots {
		slot.Used = 0
		c.Spellcasting.Slots[k] = slot
	}

	// Restaura todos os recursos limitados
	for i := range c.LimitedUses {
		c.LimitedUses[i].CurrentUses = c.LimitedUses[i].MaxUses
	}

	c.UpdatedAt = time.Now()
}

// ShortRest executa um Descanso Curto (restaura recursos que recarregam em descanso curto)
func (c *CharacterSheet) ShortRest() {
	for i := range c.LimitedUses {
		if strings.EqualFold(c.LimitedUses[i].ResetOn, "Descanso Curto") {
			c.LimitedUses[i].CurrentUses = c.LimitedUses[i].MaxUses
		}
	}
	c.UpdatedAt = time.Now()
}

// GetAttributeModifier retorna o modificador do atributo
func (c *CharacterSheet) GetAttributeModifier(attr string) int {
	switch strings.ToLower(attr) {
	case "for", "str", "força", "forca":
		return c.Attributes.STR.Modifier
	case "des", "dex", "destreza":
		return c.Attributes.DEX.Modifier
	case "con", "constituição", "constituicao":
		return c.Attributes.CON.Modifier
	case "int", "inteligência", "inteligencia":
		return c.Attributes.INT.Modifier
	case "sab", "wis", "sabedoria":
		return c.Attributes.WIS.Modifier
	case "car", "cha", "carisma":
		return c.Attributes.CHA.Modifier
	}
	return 0
}

// GetAdvantageMode retorna o modo de rolagem atual ("normal", "advantage" ou "disadvantage")
func (c *CharacterSheet) GetAdvantageMode() string {
	if c.AdvantageMode == "advantage" || c.AdvantageMode == "disadvantage" {
		return c.AdvantageMode
	}
	return "normal"
}

// ToggleAdvantageMode alterna ciclicamente entre normal -> advantage -> disadvantage -> normal
func (c *CharacterSheet) ToggleAdvantageMode() string {
	switch c.AdvantageMode {
	case "advantage":
		c.AdvantageMode = "disadvantage"
	case "disadvantage":
		c.AdvantageMode = "normal"
	default:
		c.AdvantageMode = "advantage"
	}
	return c.AdvantageMode
}

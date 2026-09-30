package model

// ClassFeature representa uma habilidade de classe obtida em determinado nível
type ClassFeature struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SubclassFeature representa uma habilidade de subclasse
type SubclassFeature struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Subclass representa uma subclasse de D&D 2024 / 5.5e
type Subclass struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ClassName   string            `json:"class_name"`
	UnlockLevel int               `json:"unlock_level"`
	Synopsis    string            `json:"synopsis"`
	Features    []SubclassFeature `json:"features"`
}

// ClassTraits resume os traços essenciais da classe
type ClassTraits struct {
	PrimaryAbility   string `json:"primary_ability"`
	HitDie           string `json:"hit_die"`
	SavingThrows     string `json:"saving_throws"`
	SkillProf        string `json:"skill_prof"`
	WeaponProf       string `json:"weapon_prof"`
	ArmorTraining    string `json:"armor_training"`
	WeaponMastery    string `json:"weapon_mastery"`
	StartingEquip    string `json:"starting_equip"`
}

// Class representa uma das 12 classes principais de D&D 2024 / 5.5e
type Class struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	NamePT      string         `json:"name_pt"`
	Emoji       string         `json:"emoji"`
	Color       int            `json:"color"`
	Overview    string         `json:"overview"`
	Role        string         `json:"role"`
	Traits      ClassTraits    `json:"traits"`
	Features    []ClassFeature `json:"features"`
	Subclasses  []Subclass     `json:"subclasses"`
}

// Spell representa uma magia de D&D 2024 / 5.5e
type Spell struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Level         int      `json:"level"` // 0 = Cantrip (Truque)
	School        string   `json:"school"`
	Classes       []string `json:"classes"`
	CastingTime   string   `json:"casting_time"`
	Range         string   `json:"range"`
	Components    string   `json:"components"`
	Duration      string   `json:"duration"`
	Concentration bool     `json:"concentration"`
	Ritual        bool     `json:"ritual"`
	Description   string   `json:"description"`
	HigherLevels  string   `json:"higher_levels"`
}

// MonsterStats atributos básicos
type MonsterStats struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
}

// MonsterAction ação ou habilidade de monstro
type MonsterAction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Monster representa um monstro do Bestiário D&D 2024
type Monster struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Size             string          `json:"size"`
	Type             string          `json:"type"`
	Alignment        string          `json:"alignment"`
	ArmorClass       string          `json:"armor_class"`
	HitPoints        string          `json:"hit_points"`
	Speed            string          `json:"speed"`
	Stats            MonsterStats    `json:"stats"`
	SavingThrows     string          `json:"saving_throws"`
	Skills           string          `json:"skills"`
	DamageResist     string          `json:"damage_resist"`
	DamageImmune     string          `json:"damage_immune"`
	ConditionImmune  string          `json:"condition_immune"`
	Senses           string          `json:"senses"`
	Languages        string          `json:"languages"`
	ChallengeRating  string          `json:"challenge_rating"`
	XP               string          `json:"xp"`
	Traits           []MonsterAction `json:"traits"`
	Actions          []MonsterAction `json:"actions"`
	BonusActions     []MonsterAction `json:"bonus_actions"`
	Reactions        []MonsterAction `json:"reactions"`
	LegendaryActions []MonsterAction `json:"legendary_actions"`
}

// Item representa equipamento ou item mágico de D&D 2024
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"` // Weapon, Armor, Wondrous Item, Potion, Ring, Scroll, Wand, etc.
	Rarity      string `json:"rarity"`   // Mundane, Common, Uncommon, Rare, Very Rare, Legendary, Artifact
	Attunement  string `json:"attunement"`
	Damage      string `json:"damage"`
	Mastery     string `json:"mastery"` // Cleave, Graze, Nick, Push, Sap, Slow, Topple, Vex (D&D 2024)
	Cost        string `json:"cost"`
	Weight      string `json:"weight"`
	Properties  string `json:"properties"`
	Description string `json:"description"`
}

// Rule representa um termo do Glossário de Regras do D&D 2024
type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"` // Action, Condition, Hazard, General Rule
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

// SpeciesTrait representa um traço racial ou habilidade especial de espécie
type SpeciesTrait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Species representa uma espécie / raça jogável de D&D (ex: Ravenloft, PHB 2024)
type Species struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	NamePT       string         `json:"name_pt"`
	Category     string         `json:"category"` // Ravenloft, Common, etc.
	Source       string         `json:"source"`
	CreatureType string         `json:"creature_type"` // Humanoid, Fey, etc.
	Size         string         `json:"size"`          // Medium or Small
	Speed        string         `json:"speed"`         // 30 ft (9m)
	Darkvision   string         `json:"darkvision"`    // 60 ft (18m)
	Overview     string         `json:"overview"`
	Traits       []SpeciesTrait `json:"traits"`
}


package repository

import (
	"fmt"
	"strings"
	"sync"

	"kenobi-bot/internal/dnd/model"
)

// DNDRepository define o contrato de consulta para dados do D&D 2024 / 5.5e
type DNDRepository interface {
	GetClasses() []model.Class
	GetClass(idOrName string) (*model.Class, bool)
	GetSubclasses(className string) []model.Subclass
	GetSubclass(className, subclassName string) (*model.Subclass, bool)

	GetSpeciesList(category string) []model.Species
	GetSpecies(idOrName string) (*model.Species, bool)
	GetSpeciesCategories() []string

	SearchSpells(name string, level int, school string, className string, limit int) []model.Spell
	GetSpell(idOrName string) (*model.Spell, bool)
	GetSpellSchools() []string

	SearchMonsters(name string, cr string, monType string, limit int) []model.Monster
	GetMonster(idOrName string) (*model.Monster, bool)
	GetMonsterTypes() []string

	SearchItems(name string, category string, rarity string, mastery string, limit int) []model.Item
	GetItem(idOrName string) (*model.Item, bool)
	GetItemCategories() []string

	SearchRules(query string, limit int) []model.Rule
	GetRule(idOrName string) (*model.Rule, bool)
}

// MemoryDNDRepository armazena todos os dados em memória para acesso ultrarrápido
type MemoryDNDRepository struct {
	mu        sync.RWMutex
	classes   []model.Class
	species   []model.Species
	spells    []model.Spell
	monsters  []model.Monster
	items     []model.Item
	rules     []model.Rule

	schools     []string
	monTypes    []string
	itemCats    []string
	speciesCats []string
}

// NewMemoryRepository inicializa o repositório em memória
func NewMemoryRepository(
	classes []model.Class,
	species []model.Species,
	spells []model.Spell,
	monsters []model.Monster,
	items []model.Item,
	rules []model.Rule,
) *MemoryDNDRepository {
	repo := &MemoryDNDRepository{
		classes:  classes,
		species:  species,
		spells:   spells,
		monsters: monsters,
		items:    items,
		rules:    rules,
	}
	repo.buildIndices()
	return repo
}

func (r *MemoryDNDRepository) buildIndices() {
	schoolSet := make(map[string]struct{})
	for _, s := range r.spells {
		if s.School != "" {
			schoolSet[s.School] = struct{}{}
		}
	}
	for sc := range schoolSet {
		r.schools = append(r.schools, sc)
	}

	monTypeSet := make(map[string]struct{})
	for _, m := range r.monsters {
		if m.Type != "" {
			monTypeSet[m.Type] = struct{}{}
		}
	}
	for mt := range monTypeSet {
		r.monTypes = append(r.monTypes, mt)
	}

	itemCatSet := make(map[string]struct{})
	for _, it := range r.items {
		if it.Category != "" {
			itemCatSet[it.Category] = struct{}{}
		}
	}
	for ic := range itemCatSet {
		r.itemCats = append(r.itemCats, ic)
	}

	specCatSet := make(map[string]struct{})
	for _, sp := range r.species {
		if sp.Category != "" {
			specCatSet[sp.Category] = struct{}{}
		}
	}
	for sc := range specCatSet {
		r.speciesCats = append(r.speciesCats, sc)
	}
}

func (r *MemoryDNDRepository) GetClasses() []model.Class {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.classes
}

func (r *MemoryDNDRepository) GetClass(idOrName string) (*model.Class, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(idOrName))
	for _, c := range r.classes {
		if strings.ToLower(c.ID) == query || strings.ToLower(c.Name) == query || strings.ToLower(c.NamePT) == query {
			return &c, true
		}
	}
	return nil, false
}

func (r *MemoryDNDRepository) GetSubclasses(className string) []model.Subclass {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, found := r.GetClass(className)
	if !found {
		return nil
	}
	return c.Subclasses
}

func (r *MemoryDNDRepository) GetSubclass(className, subclassName string) (*model.Subclass, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, found := r.GetClass(className)
	if !found {
		// Tentar buscar em todas as subclasses caso className seja genérico
		subQuery := strings.ToLower(strings.TrimSpace(subclassName))
		for _, cls := range r.classes {
			for _, sub := range cls.Subclasses {
				if strings.ToLower(sub.ID) == subQuery || strings.ToLower(sub.Name) == subQuery || strings.Contains(strings.ToLower(sub.Name), subQuery) {
					return &sub, true
				}
			}
		}
		return nil, false
	}

	subQuery := strings.ToLower(strings.TrimSpace(subclassName))
	for _, sub := range c.Subclasses {
		if strings.ToLower(sub.ID) == subQuery || strings.ToLower(sub.Name) == subQuery || strings.Contains(strings.ToLower(sub.Name), subQuery) {
			return &sub, true
		}
	}
	return nil, false
}

func (r *MemoryDNDRepository) GetSpeciesList(category string) []model.Species {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if category == "" || strings.EqualFold(category, "all") {
		return r.species
	}

	var res []model.Species
	catLower := strings.ToLower(category)
	for _, sp := range r.species {
		if strings.ToLower(sp.Category) == catLower {
			res = append(res, sp)
		}
	}
	return res
}

func (r *MemoryDNDRepository) GetSpecies(idOrName string) (*model.Species, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(idOrName))
	for _, sp := range r.species {
		if strings.ToLower(sp.ID) == query || strings.ToLower(sp.Name) == query || strings.ToLower(sp.NamePT) == query || strings.Contains(strings.ToLower(sp.Name), query) || strings.Contains(strings.ToLower(sp.NamePT), query) {
			return &sp, true
		}
	}
	return nil, false
}

func (r *MemoryDNDRepository) GetSpeciesCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.speciesCats
}

func (r *MemoryDNDRepository) SearchSpells(name string, level int, school string, className string, limit int) []model.Spell {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []model.Spell
	name = strings.ToLower(strings.TrimSpace(name))
	school = strings.ToLower(strings.TrimSpace(school))
	className = strings.ToLower(strings.TrimSpace(className))

	for _, s := range r.spells {
		if name != "" && !strings.Contains(strings.ToLower(s.Name), name) {
			continue
		}
		if level >= 0 && s.Level != level {
			continue
		}
		if school != "" && !strings.EqualFold(s.School, school) {
			continue
		}
		if className != "" {
			matchedClass := false
			for _, c := range s.Classes {
				if strings.EqualFold(c, className) {
					matchedClass = true
					break
				}
			}
			if !matchedClass {
				continue
			}
		}

		results = append(results, s)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func (r *MemoryDNDRepository) GetSpell(idOrName string) (*model.Spell, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(idOrName))
	for _, s := range r.spells {
		if strings.ToLower(s.ID) == query || strings.ToLower(s.Name) == query {
			return &s, true
		}
	}
	// Busca por aproximação
	for _, s := range r.spells {
		if strings.Contains(strings.ToLower(s.Name), query) {
			return &s, true
		}
	}
	return nil, false
}

func (r *MemoryDNDRepository) GetSpellSchools() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.schools
}

func (r *MemoryDNDRepository) SearchMonsters(name string, cr string, monType string, limit int) []model.Monster {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []model.Monster
	name = strings.ToLower(strings.TrimSpace(name))
	monType = strings.ToLower(strings.TrimSpace(monType))

	for _, m := range r.monsters {
		if name != "" && !strings.Contains(strings.ToLower(m.Name), name) {
			continue
		}
		if cr != "" && !matchCR(m.ChallengeRating, cr) {
			continue
		}
		if monType != "" && !strings.Contains(strings.ToLower(m.Type), monType) {
			continue
		}

		results = append(results, m)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func matchCR(monsterCR, filterCR string) bool {
	if filterCR == "" {
		return true
	}
	monsterCR = strings.TrimSpace(strings.ToLower(monsterCR))
	filterCR = strings.TrimSpace(strings.ToLower(filterCR))

	if strings.EqualFold(monsterCR, filterCR) {
		return true
	}

	var val float64
	if strings.HasPrefix(monsterCR, "1/8") {
		val = 0.125
	} else if strings.HasPrefix(monsterCR, "1/4") {
		val = 0.25
	} else if strings.HasPrefix(monsterCR, "1/2") {
		val = 0.5
	} else if strings.HasPrefix(monsterCR, "0") {
		val = 0.0
	} else {
		_, _ = fmt.Sscanf(monsterCR, "%f", &val)
	}

	switch filterCR {
	case "0-1/2":
		return val <= 0.5
	case "1-4":
		return val >= 1.0 && val <= 4.0
	case "5-9":
		return val >= 5.0 && val <= 9.0
	case "10-15":
		return val >= 10.0 && val <= 15.0
	case "16+":
		return val >= 16.0
	}

	return strings.Contains(monsterCR, filterCR)
}

func (r *MemoryDNDRepository) GetMonster(idOrName string) (*model.Monster, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(idOrName))
	for _, m := range r.monsters {
		if strings.ToLower(m.ID) == query || strings.ToLower(m.Name) == query {
			return &m, true
		}
	}
	for _, m := range r.monsters {
		if strings.Contains(strings.ToLower(m.Name), query) {
			return &m, true
		}
	}
	return nil, false
}

func (r *MemoryDNDRepository) GetMonsterTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.monTypes
}

func (r *MemoryDNDRepository) SearchItems(name string, category string, rarity string, mastery string, limit int) []model.Item {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []model.Item
	name = strings.ToLower(strings.TrimSpace(name))
	category = strings.ToLower(strings.TrimSpace(category))
	rarity = strings.ToLower(strings.TrimSpace(rarity))
	mastery = strings.ToLower(strings.TrimSpace(mastery))

	for _, it := range r.items {
		if name != "" && !strings.Contains(strings.ToLower(it.Name), name) {
			continue
		}
		if category != "" && !strings.Contains(strings.ToLower(it.Category), category) {
			continue
		}
		if rarity != "" && !strings.EqualFold(it.Rarity, rarity) {
			continue
		}
		if mastery != "" && !strings.Contains(strings.ToLower(it.Mastery), mastery) {
			continue
		}

		results = append(results, it)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func (r *MemoryDNDRepository) GetItem(idOrName string) (*model.Item, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(idOrName))
	for _, it := range r.items {
		if strings.ToLower(it.ID) == query || strings.ToLower(it.Name) == query {
			return &it, true
		}
	}
	for _, it := range r.items {
		if strings.Contains(strings.ToLower(it.Name), query) {
			return &it, true
		}
	}
	return nil, false
}

func (r *MemoryDNDRepository) GetItemCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.itemCats
}

func (r *MemoryDNDRepository) SearchRules(query string, limit int) []model.Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []model.Rule
	q := strings.ToLower(strings.TrimSpace(query))

	for _, ru := range r.rules {
		if q != "" && !strings.Contains(strings.ToLower(ru.Name), q) && !strings.Contains(strings.ToLower(ru.Summary), q) {
			continue
		}
		results = append(results, ru)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func (r *MemoryDNDRepository) GetRule(idOrName string) (*model.Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(idOrName))
	for _, ru := range r.rules {
		if strings.ToLower(ru.ID) == query || strings.ToLower(ru.Name) == query {
			return &ru, true
		}
	}
	for _, ru := range r.rules {
		if strings.Contains(strings.ToLower(ru.Name), query) {
			return &ru, true
		}
	}
	return nil, false
}

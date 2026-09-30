package dice

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// diceRegex captura:
	// Grupo 1: Quantidade de dados (opcional, padrão 1)
	// Grupo 2: Lados do dado (1 a 100)
	// Grupo 3: Flag de keep (k ou kh) (opcional)
	// Grupo 4: Quantidade de dados a manter (opcional)
	// Grupo 5: Sinal do modificador (+ ou -) (opcional)
	// Grupo 6: Valor do modificador (opcional)
	// Grupo 7: Flag explícita de reroll (ex: r1, ro1) (opcional)
	diceRegex = regexp.MustCompile(`^(?i)\s*(\d*)d(\d+)(?:\s*(k(?:h)?)\s*(\d*))?(?:\s*([+-])\s*(\d+))?(?:\s*(r1|ro1))?\s*$`)

	ErrInvalidFormat = errors.New("formato de dado inválido (use ex: d20, 2d10, 3d20k2, 1d8+2, 4d6-1)")
	ErrMaxSides      = errors.New("o dado máximo permitido é d100")
	ErrMaxCount      = errors.New("a quantidade máxima permitida é de 100 dados por rolagem")
	ErrZeroDice      = errors.New("a quantidade de dados e lados deve ser no mínimo 1")
	ErrKeepTooHigh   = errors.New("a quantidade de dados a manter (k) não pode ser maior que o total de dados rolados")
)

// DieDetail detalha a rolagem individual de cada dado
type DieDetail struct {
	InitialRoll int  // Primeiro resultado rolado
	FinalRoll   int  // Resultado final considerado
	Rerolled    bool // Se foi rerolado por tirar 1
	Kept        bool // Se este dado foi mantido (para regras de keep k/kh)
}

// RollOptions define opções avançadas de rolagem (ex: Sorte Halfling / Great Weapon Fighting)
type RollOptions struct {
	RerollOn1 bool // Se true, rerrola 1s automaticamente uma vez
}

// RollResult armazena todos os detalhes da rolagem
type RollResult struct {
	Notation string      // ex: "3d20k2+3"
	Count    int         // ex: 3
	Sides    int         // ex: 20
	Keep     int         // ex: 2 (0 significa todos mantidos)
	Modifier int         // ex: +3
	Rolls    []int       // ex: [18, 5, 12]
	Details  []DieDetail // ex: [{18, 18, false, true}, ...]
	DiceSum  int         // ex: 30 (soma apenas dos dados mantidos)
	Total    int         // ex: 33
}

// DiceRoller define o contrato para rolagem de dados
type DiceRoller interface {
	Roll(notation string) (*RollResult, error)
	RollWithOptions(notation string, opts RollOptions) (*RollResult, error)
	IsDiceNotation(text string) bool
}

// Roller implementa DiceRoller com gerador criptograficamente seguro
type Roller struct{}

// NewRoller cria um novo rolador de dados
func NewRoller() *Roller {
	return &Roller{}
}

// IsDiceNotation verifica rapidamente se uma string é uma notação de dados válida
func (r *Roller) IsDiceNotation(text string) bool {
	return diceRegex.MatchString(text)
}

// Roll interpreta a notação e executa a rolagem padrão
func (r *Roller) Roll(notation string) (*RollResult, error) {
	return r.RollWithOptions(notation, RollOptions{})
}

// RollWithOptions interpreta a notação e executa a rolagem com opções (como rerrolar 1s e manter N maiores)
func (r *Roller) RollWithOptions(notation string, opts RollOptions) (*RollResult, error) {
	matches := diceRegex.FindStringSubmatch(notation)
	if matches == nil {
		return nil, ErrInvalidFormat
	}

	countStr := matches[1]
	sidesStr := matches[2]
	keepFlag := matches[3]
	keepStr := matches[4]
	signStr := matches[5]
	modStr := matches[6]
	rerollFlag := matches[7]

	// Se a notação tiver tag explícita "r1" ou "ro1", ativa reroll
	if rerollFlag != "" {
		opts.RerollOn1 = true
	}

	// 1. Quantidade de dados (padrão 1 se omitido, ex: "d20" -> 1)
	count := 1
	if countStr != "" {
		c, err := strconv.Atoi(countStr)
		if err != nil || c < 1 {
			return nil, ErrZeroDice
		}
		if c > 100 {
			return nil, ErrMaxCount
		}
		count = c
	}

	// 2. Número de lados do dado (1 a 100)
	sides, err := strconv.Atoi(sidesStr)
	if err != nil || sides < 1 {
		return nil, ErrZeroDice
	}
	if sides > 100 {
		return nil, ErrMaxSides
	}

	// 3. Quantidade de dados a manter (k / kh)
	keep := 0
	if keepFlag != "" {
		if keepStr != "" {
			k, err := strconv.Atoi(keepStr)
			if err != nil || k < 1 {
				return nil, errors.New("a quantidade de dados a manter (k) deve ser no mínimo 1")
			}
			if k > count {
				return nil, ErrKeepTooHigh
			}
			keep = k
		} else {
			// Se digitou apenas 'k' ou 'kh' sem número (ex: 2d20k), mantém 1
			keep = 1
		}
	}

	// 4. Modificador opcional (+ ou -)
	modifier := 0
	if modStr != "" {
		m, err := strconv.Atoi(modStr)
		if err == nil {
			if signStr == "-" {
				modifier = -m
			} else {
				modifier = m
			}
		}
	}

	// 5. Rolar cada dado de forma aleatória e uniforme no intervalo [1, sides]
	var rolls []int
	var details []DieDetail

	for i := 0; i < count; i++ {
		initial := rollSingleDie(sides)
		final := initial
		rerolled := false

		// Se a opção de rerrolar 1 estiver ativa e o dado tirou 1
		if opts.RerollOn1 && initial == 1 && sides > 1 {
			final = rollSingleDie(sides)
			rerolled = true
		}

		rolls = append(rolls, final)
		details = append(details, DieDetail{
			InitialRoll: initial,
			FinalRoll:   final,
			Rerolled:    rerolled,
			Kept:        true,
		})
	}

	// 6. Aplicar regra de Keep (manter os N maiores valores)
	diceSum := 0
	if keep > 0 && keep < count {
		type indexedRoll struct {
			index int
			val   int
		}
		indexed := make([]indexedRoll, count)
		for i, d := range details {
			indexed[i] = indexedRoll{index: i, val: d.FinalRoll}
		}

		// Ordena decrescente pelo valor rolado
		sort.SliceStable(indexed, func(i, j int) bool {
			return indexed[i].val > indexed[j].val
		})

		// Marca todos como descartados inicialmente
		for i := range details {
			details[i].Kept = false
		}
		// Marca apenas os 'keep' maiores como mantidos
		for i := 0; i < keep; i++ {
			idx := indexed[i].index
			details[idx].Kept = true
		}
	}

	// Somar apenas os dados mantidos
	for _, d := range details {
		if d.Kept {
			diceSum += d.FinalRoll
		}
	}

	total := diceSum + modifier

	return &RollResult{
		Notation: strings.TrimSpace(notation),
		Count:    count,
		Sides:    sides,
		Keep:     keep,
		Modifier: modifier,
		Rolls:    rolls,
		Details:  details,
		DiceSum:  diceSum,
		Total:    total,
	}, nil
}

// rollSingleDie gera um número natural aleatório entre 1 e sides usando crypto/rand
func rollSingleDie(sides int) int {
	if sides <= 1 {
		return 1
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 1
	}
	return int(nBig.Int64()) + 1
}

// Format gera a mensagem estilizada para o Discord
func (res *RollResult) Format() string {
	var rollsStr []string
	hasReroll := false
	isKeep := res.Keep > 0 && res.Keep < res.Count

	for _, d := range res.Details {
		isDropped := isKeep && !d.Kept
		if d.Rerolled {
			hasReroll = true
			if isDropped {
				rollsStr = append(rollsStr, fmt.Sprintf("~~%d ➔ %d~~", d.InitialRoll, d.FinalRoll))
			} else {
				rollsStr = append(rollsStr, fmt.Sprintf("~~%d~~ ➔ **%d**", d.InitialRoll, d.FinalRoll))
			}
		} else {
			if isDropped {
				rollsStr = append(rollsStr, fmt.Sprintf("~~%d~~", d.FinalRoll))
			} else {
				rollsStr = append(rollsStr, fmt.Sprintf("**%d**", d.FinalRoll))
			}
		}
	}
	rollsFormatted := fmt.Sprintf("[%s]", strings.Join(rollsStr, ", "))

	rerollTag := ""
	if hasReroll {
		rerollTag = " *(🍀 1 Rerrolado!)*"
	}

	// Caso 1: Apenas 1 dado sem modificador (ex: d20 -> [15] = 15)
	if res.Count == 1 && res.Modifier == 0 {
		return fmt.Sprintf("🎲 Rolando `%s`%s:\nResultado: %s ➔ **Total: %d**",
			res.Notation, rerollTag, rollsFormatted, res.Total)
	}

	// Caso 2: Apenas 1 dado com modificador (ex: 1d8+2 -> [6] + 2 = 8)
	if res.Count == 1 && res.Modifier != 0 {
		modSign := "+"
		modVal := res.Modifier
		if res.Modifier < 0 {
			modSign = "-"
			modVal = -res.Modifier
		}
		return fmt.Sprintf("🎲 Rolando `%s`%s:\nResultado: %s %s %d ➔ **Total: %d**",
			res.Notation, rerollTag, rollsFormatted, modSign, modVal, res.Total)
	}

	// Rótulo da soma dos dados
	sumLabel := fmt.Sprintf("Soma: %d", res.DiceSum)
	if isKeep {
		sumLabel = fmt.Sprintf("%d maiores: %d", res.Keep, res.DiceSum)
	}

	// Caso 3: Múltiplos dados sem modificador (ex: 3d20k2 -> [**18**, ~~5~~, **12**] (2 maiores: 30) = 30)
	if res.Count > 1 && res.Modifier == 0 {
		return fmt.Sprintf("🎲 Rolando `%s`%s:\nDados: %s (%s) ➔ **Total: %d**",
			res.Notation, rerollTag, rollsFormatted, sumLabel, res.Total)
	}

	// Caso 4: Múltiplos dados com modificador (ex: 3d20k2+5 -> [**18**, ~~5~~, **12**] (2 maiores: 30) + 5 = 35)
	modSign := "+"
	modVal := res.Modifier
	if res.Modifier < 0 {
		modSign = "-"
		modVal = -res.Modifier
	}
	return fmt.Sprintf("🎲 Rolando `%s`%s:\nDados: %s (%s) %s %d ➔ **Total: %d**",
		res.Notation, rerollTag, rollsFormatted, sumLabel, modSign, modVal, res.Total)
}


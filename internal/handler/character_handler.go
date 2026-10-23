package handler

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"kenobi-bot/internal/character"
	"kenobi-bot/internal/dice"
	"kenobi-bot/internal/ui"
)

// CharacterHandler gerencia comandos, botões, modais e criação de fichas
type CharacterHandler struct {
	service character.Service
	wizard  *character.WizardManager
}

// NewCharacterHandler instancia o handler de fichas de personagens
func NewCharacterHandler(service character.Service) *CharacterHandler {
	return &CharacterHandler{
		service: service,
		wizard:  character.NewWizardManager(service),
	}
}

// GetSlashCommands retorna as definições dos comandos slash do sistema de fichas
func (h *CharacterHandler) GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "ficha",
			Description: "📜 Sistema de Fichas de RPG D&D (criação, visualização, importação e testes)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ver",
					Description: "Exibe sua ficha ativa com botões interativos e abas de combate",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "usuario",
							Description: "Usuário cuja ficha você deseja visualizar (opcional)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "criar",
					Description: "🧙 Inicia o modo de criação interativa de ficha passo a passo",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "cancelar",
					Description: "Cancela a criação de ficha em andamento",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "importar",
					Description: "📄 Importa sua ficha enviando o arquivo PDF oficial do D&D",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionAttachment,
							Name:        "pdf",
							Description: "Arquivo PDF da sua ficha de personagem",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "listar",
					Description: "Lista todas as fichas de RPG cadastradas para o seu usuário",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "selecionar",
					Description: "Define qual das suas fichas será a ativa no Discord",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "nome",
							Description: "Nome ou ID da ficha que você quer ativar",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "deletar",
					Description: "Exclui uma das suas fichas cadastradas",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "nome",
							Description: "Nome ou ID da ficha a ser excluída",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "rolar",
					Description: "🎲 Rola um teste somando o modificador exato da sua ficha ativa",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "teste",
							Description: "Nome do atributo ou perícia (ex: percepcao, furtividade, forca, iniciativa)",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "atribuir",
					Description: "👥 Atribui uma ficha de personagem a um jogador do servidor",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "ficha",
							Description: "Nome ou ID da ficha (ex: Jaci)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "jogador",
							Description: "O jogador no Discord que controlará esta ficha",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "hp",
			Description: "❤️ Gerencia os Pontos de Vida (HP) da sua ficha ativa no Discord",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "dano",
					Description: "Aplica dano à sua ficha (desconta de tempHP e depois HP atual)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "quantidade",
							Description: "Valor do dano a receber",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "cura",
					Description: "Cura pontos de vida da sua ficha até o máximo",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "quantidade",
							Description: "Valor da cura a receber",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "temp",
					Description: "Define pontos de vida temporários",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "quantidade",
							Description: "Valor de HP Temporário",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "magia",
			Description: "✨ Gerencia espaços e slots diários de magia da sua ficha",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "usar",
					Description: "Gasta um espaço de magia diário de determinado nível",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "nivel",
							Description: "Nível da magia (1 a 9)",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "restaurar",
					Description: "Recupera um espaço de magia de determinado nível",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "nivel",
							Description: "Nível da magia (1 a 9)",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "descanso",
			Description: "⛺ Realiza Descanso Curto ou Longo recuperando vida, slots e recursos",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "curto",
					Description: "Descanso Curto: recupera recursos que recarregam em descanso curto",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "longo",
					Description: "Descanso Longo: restaura todo o HP, todos os slots e recursos diários",
				},
			},
		},
	}
}

// HandleSlashCommand processa comandos de barra
func (h *CharacterHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	switch cmdName {
	case "ficha":
		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			h.showActiveSheet(s, i, userID, "overview")
			return
		}

		subCmd := options[0].Name
		switch subCmd {
		case "ver":
			targetUser := userID
			if len(options[0].Options) > 0 {
				targetUser = options[0].Options[0].UserValue(s).ID
			}
			h.showActiveSheet(s, i, targetUser, "overview")

		case "criar":
			prompt, err := h.wizard.StartSession(userID, i.ChannelID)
			if err != nil {
				h.respondEphemeral(s, i, fmt.Sprintf("Erro ao iniciar criação: %v", err))
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: prompt,
				},
			})

		case "cancelar":
			if h.wizard.CancelSession(userID) {
				h.respondText(s, i, "❌ Sessão de criação de ficha cancelada com sucesso.")
			} else {
				h.respondEphemeral(s, i, "Você não possui nenhuma criação de ficha em andamento.")
			}

		case "importar":
			h.handlePDFImport(s, i, userID, options[0].Options)

		case "listar":
			h.handleListSheets(s, i, userID)

		case "selecionar":
			if len(options[0].Options) > 0 {
				query := options[0].Options[0].StringValue()
				h.handleSelectSheet(s, i, userID, query)
			}

		case "deletar":
			if len(options[0].Options) > 0 {
				query := options[0].Options[0].StringValue()
				h.handleDeleteSheet(s, i, userID, query)
			}

		case "rolar":
			if len(options[0].Options) > 0 {
				testName := options[0].Options[0].StringValue()
				h.handleSheetRoll(s, i, userID, testName)
			}

		case "atribuir":
			fichaName := ""
			var targetUser *discordgo.User
			for _, opt := range options[0].Options {
				switch opt.Name {
				case "ficha":
					fichaName = opt.StringValue()
				case "jogador":
					targetUser = opt.UserValue(s)
				}
			}
			if targetUser == nil || fichaName == "" {
				h.respondEphemeral(s, i, "Especifique o nome da ficha e mencione o jogador para atribuir.")
				return
			}
			sheet, err := h.service.AssignSheet(fichaName, targetUser.ID)
			if err != nil {
				h.respondEphemeral(s, i, fmt.Sprintf("⚠️ Erro ao atribuir ficha: %v", err))
				return
			}
			msg := fmt.Sprintf("✅ **Ficha Atribuída com Sucesso!**\nA ficha **%s** agora pertence a <@%s> e foi definida como sua ficha ativa!\nO jogador já pode usar `/ficha ver`, rolar ataques e usar todos os recursos com 1 clique.", sheet.Name, targetUser.ID)
			h.respondText(s, i, msg)
		}

	case "hp":
		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			return
		}
		subCmd := options[0].Name
		amount := 0
		if len(options[0].Options) > 0 {
			amount = int(options[0].Options[0].IntValue())
		}

		switch subCmd {
		case "dano":
			sheet, lost, isDown, err := h.service.ApplyDamage(userID, amount)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			msg := fmt.Sprintf("💥 **%s** sofreu **%d de dano**! (HP: %d/%d)", sheet.Name, lost, sheet.CurrentHP, sheet.MaxHP)
			if isDown {
				msg += "\n💀 **ATENÇÃO:** O personagem caiu a **0 Pontos de Vida**!"
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "overview"),
				},
			})

		case "cura":
			sheet, healed, err := h.service.ApplyHealing(userID, amount)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			msg := fmt.Sprintf("💚 **%s** recuperou **%d Pontos de Vida**! (HP: %d/%d)", sheet.Name, healed, sheet.CurrentHP, sheet.MaxHP)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "overview"),
				},
			})

		case "temp":
			sheet, err := h.service.SetTempHP(userID, amount)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			msg := fmt.Sprintf("🛡️ **%s** agora tem **%d Pontos de Vida Temporários**!", sheet.Name, sheet.TempHP)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "overview"),
				},
			})
		}

	case "magia":
		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			return
		}
		subCmd := options[0].Name
		lvl := 1
		if len(options[0].Options) > 0 {
			lvl = int(options[0].Options[0].IntValue())
		}

		switch subCmd {
		case "usar":
			sheet, err := h.service.UseSpellSlot(userID, lvl)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			slot := sheet.Spellcasting.Slots[fmt.Sprintf("%d", lvl)]
			msg := fmt.Sprintf("✨ **%s** conjurou uma magia gastando 1 espaço de **%dº Nível**! (Restam: %d/%d)",
				sheet.Name, lvl, slot.Max-slot.Used, slot.Max)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterSpellsEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "spells"),
				},
			})

		case "restaurar":
			sheet, err := h.service.RestoreSpellSlot(userID, lvl)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			slot := sheet.Spellcasting.Slots[fmt.Sprintf("%d", lvl)]
			msg := fmt.Sprintf("✨ Espaço de **%dº Nível** restaurado para **%s**! (Disponíveis: %d/%d)",
				lvl, sheet.Name, slot.Max-slot.Used, slot.Max)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterSpellsEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "spells"),
				},
			})
		}

	case "descanso":
		options := i.ApplicationCommandData().Options
		if len(options) == 0 {
			return
		}
		subCmd := options[0].Name

		switch subCmd {
		case "curto":
			sheet, err := h.service.PerformShortRest(userID)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			msg := fmt.Sprintf("⛺ **%s** concluiu um **Descanso Curto**!\nRecursos de recarga curta foram totalmente restaurados.", sheet.Name)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "overview"),
				},
			})

		case "longo":
			sheet, err := h.service.PerformLongRest(userID)
			if err != nil {
				h.respondEphemeral(s, i, err.Error())
				return
			}
			msg := fmt.Sprintf("🛌 **%s** concluiu um **Descanso Longo**!\n❤️ Pontos de Vida restaurados ao máximo (%d/%d PV).\n✨ Todos os espaços de magia e recursos diários recuperados.",
				sheet.Name, sheet.MaxHP, sheet.MaxHP)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msg,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "overview"),
				},
			})
		}
	}
}

// HandleComponentInteraction processa cliques em abas e botões de HP/descanso/magias
func (h *CharacterHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	sheet, ok := h.service.GetActive(userID)
	if !ok {
		// Fallback: se o usuário não tem ficha ativa selecionada, identifica a ficha a partir da mensagem ou embed clicado
		if i.Message != nil && len(i.Message.Embeds) > 0 {
			embedTitle := i.Message.Embeds[0].Title
			for _, sCandidate := range h.service.GetAll() {
				if strings.Contains(embedTitle, sCandidate.Name) {
					sheet = sCandidate
					ok = true
					break
				}
			}
		}
	}
	if !ok {
		h.respondEphemeral(s, i, "⚠️ Você não possui uma ficha de personagem ativa. Peça ao Mestre para usar `/ficha atribuir` ou crie uma com `/ficha criar`.")
		return
	}

	// 1. Abas de navegação (Geral, Ataques, Perícias, Habilidades, Inventário)
	if strings.HasPrefix(customID, "sheet:tab:") {
		tab := strings.TrimPrefix(customID, "sheet:tab:")
		var embed *discordgo.MessageEmbed
		switch tab {
		case "attacks":
			embed = ui.BuildCharacterAttacksEmbed(sheet)
		case "skills", "attributes":
			embed = ui.BuildCharacterSkillsEmbed(sheet)
		case "spells":
			embed = ui.BuildCharacterSpellsEmbed(sheet)
		case "features":
			embed = ui.BuildCharacterFeaturesEmbed(sheet, 0)
		case "inventory":
			embed = ui.BuildCharacterInventoryEmbed(sheet)
		default:
			tab = "overview"
			embed = ui.BuildCharacterOverviewEmbed(sheet)
		}
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: ui.BuildCharacterActionComponents(sheet, tab, 0),
			},
		}); err != nil {
			slog.Error("Falha ao atualizar aba da ficha", "error", err, "tab", tab, "char", sheet.Name)
		}
		return
	}

	// 1.1 Alternar Vantagem / Desvantagem / Normal
	if customID == "sheet:adv:toggle" {
		_ = sheet.ToggleAdvantageMode()
		_ = h.service.Save(sheet)

		currentTab := "overview"
		if i.Message != nil && len(i.Message.Embeds) > 0 {
			title := i.Message.Embeds[0].Title
			if strings.Contains(title, "Ataques") {
				currentTab = "attacks"
			} else if strings.Contains(title, "Perícias") {
				currentTab = "skills"
			} else if strings.Contains(title, "Habilidades") {
				currentTab = "features"
			} else if strings.Contains(title, "Inventário") {
				currentTab = "inventory"
			} else if strings.Contains(title, "Grimório") || strings.Contains(title, "Magias") {
				currentTab = "spells"
			}
		}

		h.updateSheetMessage(s, i, sheet, currentTab)
		return
	}

	// 1.2 Paginação da aba de habilidades
	if strings.HasPrefix(customID, "sheet:feat:page:") {
		pageStr := strings.TrimPrefix(customID, "sheet:feat:page:")
		pageNum, err := strconv.Atoi(pageStr)
		if err != nil || pageNum < 0 {
			pageNum = 0
		}
		embed := ui.BuildCharacterFeaturesEmbed(sheet, pageNum)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: ui.BuildCharacterActionComponents(sheet, "features", pageNum),
			},
		}); err != nil {
			slog.Error("Falha ao paginar habilidades", "error", err, "page", pageNum)
		}
		return
	}

	// 1.3 Gastar recursos limitados com 1 clique (Foco, Uivos, Metabolismo, etc.)
	if strings.HasPrefix(customID, "sheet:res:use:") {
		idxStr := strings.TrimPrefix(customID, "sheet:res:use:")
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 || idx >= len(sheet.LimitedUses) {
			h.respondEphemeral(s, i, "⚠️ Recurso não encontrado na ficha.")
			return
		}
		res := &sheet.LimitedUses[idx]
		if res.CurrentUses <= 0 {
			resetMsg := "Descanso Curto ou Longo"
			if strings.EqualFold(res.ResetOn, "Descanso Longo") {
				resetMsg = "Descanso Longo"
			}
			h.respondEphemeral(s, i, fmt.Sprintf("⚠️ Você já esgotou todos os usos de **%s**! Faça um %s para recuperar.", res.Name, resetMsg))
			return
		}

		resNameLow := strings.ToLower(res.Name)

		// Regra Especial: Metabolismo (Rola 1d6 + nível de cura e recupera todos os pontos de Foco)
		if strings.Contains(resNameLow, "metabol") {
			res.CurrentUses--

			lvl := sheet.Level
			if lvl <= 0 {
				lvl = 1
			}
			healFormula := fmt.Sprintf("1d6+%d", lvl)
			roller := dice.NewRoller()
			rollResult, errRoll := roller.Roll(healFormula)

			diceVal := 0
			totalHeal := 0
			if errRoll == nil {
				totalHeal = rollResult.Total
				if len(rollResult.Details) > 0 {
					diceVal = rollResult.Details[0].FinalRoll
				}
			} else {
				diceVal = rand.Intn(6) + 1
				totalHeal = diceVal + lvl
			}

			// Cura PVs da ficha
			actualHealed := sheet.Heal(totalHeal)

			// Recupera TODOS os pontos de Foco
			focoRestored := 0
			focoMax := 0
			for fIdx := range sheet.LimitedUses {
				if strings.Contains(strings.ToLower(sheet.LimitedUses[fIdx].Name), "foco") {
					focoRestored = sheet.LimitedUses[fIdx].MaxUses - sheet.LimitedUses[fIdx].CurrentUses
					sheet.LimitedUses[fIdx].CurrentUses = sheet.LimitedUses[fIdx].MaxUses
					focoMax = sheet.LimitedUses[fIdx].MaxUses
				}
			}

			_ = h.service.Save(sheet)

			msgContent := fmt.Sprintf(
				"🧬 **%s ativou o Metabolismo!**\n"+
					"🎲 **Rolagem de Cura:** `1d6+%d` = **[%d] + %d** = **+%d PV recuperados**! (Vida Atual: `%d/%d PV`)\n"+
					"⚡ **Foco Totalmente Restaurado:** Recuperou **%d Ponto(s) de Foco** (`%d/%d Foco`)\n"+
					"⏳ *Metabolismo esgotado até o término de um Descanso Longo.*",
				sheet.Name, lvl, diceVal, lvl, actualHealed, sheet.CurrentHP, sheet.MaxHP,
				focoRestored, focoMax, focoMax,
			)

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content:    msgContent,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
					Components: ui.BuildCharacterActionComponents(sheet, "overview"),
				},
			})
			return
		}

		// Recursos Padrão
		res.CurrentUses--
		_ = h.service.Save(sheet)

		h.updateSheetMessage(s, i, sheet, "overview")
		return
	}

	// 1.4 Rolagem rápida de ataque automatizada
	if strings.HasPrefix(customID, "sheet:atk:roll:") {
		idxStr := strings.TrimPrefix(customID, "sheet:atk:roll:")
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 || idx >= len(sheet.Attacks) {
			h.respondEphemeral(s, i, "⚠️ Ataque não encontrado na ficha.")
			return
		}
		atk := sheet.Attacks[idx]

		// Extrai bônus numérico de ataque
		cleanBonus := strings.TrimSpace(atk.AttackBonus)
		bonusVal := 0
		if cleanBonus != "" {
			bStr := strings.TrimPrefix(cleanBonus, "+")
			bonusVal, _ = strconv.Atoi(bStr)
		}

		advMode := sheet.GetAdvantageMode()
		var natD20 int
		var atkLine string

		switch advMode {
		case "advantage":
			d1 := rand.Intn(20) + 1
			d2 := rand.Intn(20) + 1
			natD20 = d1
			discarded := d2
			if d2 > d1 {
				natD20 = d2
				discarded = d1
			}
			totalAtk := natD20 + bonusVal

			critMsg := ""
			switch natD20 {
			case 20:
				critMsg = " 🔥 **ACERTO CRÍTICO!**"
			case 1:
				critMsg = " 💀 **FALHA CRÍTICA!**"
			}

			bonusDisplay := fmt.Sprintf("%+d", bonusVal)
			atkLine = fmt.Sprintf("🎯 **Teste de Acerto (✨ Vantagem):** dados `[%d, %d]` ➔ Maior: **[%d]** %s = **`%d`**%s *(descartou %d)*",
				d1, d2, natD20, bonusDisplay, totalAtk, critMsg, discarded)

		case "disadvantage":
			d1 := rand.Intn(20) + 1
			d2 := rand.Intn(20) + 1
			natD20 = d1
			discarded := d2
			if d2 < d1 {
				natD20 = d2
				discarded = d1
			}
			totalAtk := natD20 + bonusVal

			critMsg := ""
			switch natD20 {
			case 20:
				critMsg = " 🔥 **ACERTO CRÍTICO!**"
			case 1:
				critMsg = " 💀 **FALHA CRÍTICA!**"
			}

			bonusDisplay := fmt.Sprintf("%+d", bonusVal)
			atkLine = fmt.Sprintf("🎯 **Teste de Acerto (💀 Desvantagem):** dados `[%d, %d]` ➔ Menor: **[%d]** %s = **`%d`**%s *(descartou %d)*",
				d1, d2, natD20, bonusDisplay, totalAtk, critMsg, discarded)

		default:
			natD20 = rand.Intn(20) + 1
			totalAtk := natD20 + bonusVal

			critMsg := ""
			switch natD20 {
			case 20:
				critMsg = " 🔥 **ACERTO CRÍTICO!**"
			case 1:
				critMsg = " 💀 **FALHA CRÍTICA!**"
			}

			bonusDisplay := fmt.Sprintf("%+d", bonusVal)
			atkLine = fmt.Sprintf("🎯 **Teste de Acerto:** **[%d]** %s = **`%d`**%s",
				natD20, bonusDisplay, totalAtk, critMsg)
		}

		// Rola dano se houver fórmula de dados
		var dmgLine string
		dmgParts := strings.Split(atk.Damage, " ")
		dmgFormula := dmgParts[0]
		dmgType := atk.DamageType
		if dmgType == "" && len(dmgParts) > 1 {
			dmgType = strings.Join(dmgParts[1:], " ")
		}

		roller := dice.NewRoller()
		dmgRoll, errDmg := roller.Roll(dmgFormula)
		if errDmg == nil {
			dmgLine = fmt.Sprintf("💥 **Dano:** %s", formatDamageBreakdown(dmgRoll, dmgType))
		} else {
			dmgLine = fmt.Sprintf("💥 **Dano:** `%s`", atk.Damage)
		}

		var respComponents []discordgo.MessageComponent
		if hasSavageAttacker(sheet) && errDmg == nil {
			savageBtn := discordgo.Button{
				Label:    "🎲 Rerolar Dano (Atacante Selvagem)",
				Style:    discordgo.SecondaryButton,
				Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				CustomID: fmt.Sprintf("sheet:atk:savage:%s:%d:%d", sheet.ID, idx, dmgRoll.Total),
			}
			respComponents = append(respComponents, discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{savageBtn},
			})
		}

		respMsg := fmt.Sprintf("⚔️ **%s** desferiu um ataque com **%s**!\n%s\n%s", sheet.Name, atk.Name, atkLine, dmgLine)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:    respMsg,
				Components: respComponents,
			},
		})
		return
	}

	// 1.5 Rerolar Dano com Talento Atacante Selvagem (Savage Attacker)
	if strings.HasPrefix(customID, "sheet:atk:savage:") {
		parts := strings.Split(customID, ":")
		if len(parts) < 6 {
			h.respondEphemeral(s, i, "⚠️ Informações do ataque inválidas.")
			return
		}
		sheetID := parts[3]
		atkIdx, errIdx := strconv.Atoi(parts[4])
		prevTotal, errTot := strconv.Atoi(parts[5])
		if errIdx != nil || errTot != nil {
			h.respondEphemeral(s, i, "⚠️ Erro ao processar dados de rerolagem.")
			return
		}

		targetSheet, found := h.service.GetByID(sheetID)
		if !found {
			targetSheet = sheet
		}

		if atkIdx < 0 || atkIdx >= len(targetSheet.Attacks) {
			h.respondEphemeral(s, i, "⚠️ Ataque não encontrado na ficha.")
			return
		}

		atk := targetSheet.Attacks[atkIdx]
		dmgParts := strings.Split(atk.Damage, " ")
		dmgFormula := dmgParts[0]
		dmgType := atk.DamageType
		if dmgType == "" && len(dmgParts) > 1 {
			dmgType = strings.Join(dmgParts[1:], " ")
		}

		roller := dice.NewRoller()
		newRoll, errRoll := roller.Roll(dmgFormula)
		if errRoll != nil {
			h.respondEphemeral(s, i, "⚠️ Falha ao rerrolar dano.")
			return
		}

		newBreakdown := formatDamageBreakdown(newRoll, dmgType)
		newTotal := newRoll.Total

		outcomeDesc := ""
		bestTotal := newTotal
		if newTotal > prevTotal {
			bestTotal = newTotal
			outcomeDesc = fmt.Sprintf("🎉 **A nova rolagem foi maior!** (de `%d` para `%d`)", prevTotal, newTotal)
		} else if prevTotal > newTotal {
			bestTotal = prevTotal
			outcomeDesc = fmt.Sprintf("🛡️ **A rolagem anterior era maior.** Mantido o valor original (`%d`)!", prevTotal)
		} else {
			outcomeDesc = fmt.Sprintf("⚖️ Ambas as rolagens empataram em `%d`!", newTotal)
		}

		origContent := ""
		if i.Message != nil {
			origContent = i.Message.Content
		}

		typeLabel := ""
		if dmgType != "" {
			typeLabel = " " + dmgType
		}

		updatedContent := fmt.Sprintf("%s\n\n🔄 **Atacante Selvagem (2ª Rolagem):** %s\n✨ **Dano Final:** **`%d`%s** — %s",
			origContent, newBreakdown, bestTotal, typeLabel, outcomeDesc)

		disabledBtn := discordgo.Button{
			Label:    "✅ Atacante Selvagem Usado",
			Style:    discordgo.SecondaryButton,
			Emoji:    &discordgo.ComponentEmoji{Name: "🎲"},
			Disabled: true,
			CustomID: "sheet:atk:savage_done",
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: updatedContent,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{Components: []discordgo.MessageComponent{disabledBtn}},
				},
			},
		})
		return
	}

	// 2. Ações rápidas de HP
	if strings.HasPrefix(customID, "sheet:hp:dmg:") {
		valStr := strings.TrimPrefix(customID, "sheet:hp:dmg:")
		val, _ := strconv.Atoi(valStr)
		sheet.TakeDamage(val)
		_ = h.service.Save(sheet)
		h.updateSheetMessage(s, i, sheet, "overview")
		return
	}

	if strings.HasPrefix(customID, "sheet:hp:heal:") {
		valStr := strings.TrimPrefix(customID, "sheet:hp:heal:")
		val, _ := strconv.Atoi(valStr)
		sheet.Heal(val)
		_ = h.service.Save(sheet)
		h.updateSheetMessage(s, i, sheet, "overview")
		return
	}

	// 3. Modais para dano e cura customizados
	if customID == "sheet:hp:custom_dmg" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "sheet:modal:damage",
				Title:    fmt.Sprintf("Aplicar Dano a %s", sheet.Name),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "damage_amount",
								Label:       "Quantidade de Dano",
								Style:       discordgo.TextInputShort,
								Placeholder: "Ex: 12",
								Required:    true,
								MaxLength:   4,
							},
						},
					},
				},
			},
		})
		return
	}

	if customID == "sheet:hp:custom_heal" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "sheet:modal:heal",
				Title:    fmt.Sprintf("Curar %s", sheet.Name),
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "heal_amount",
								Label:       "Quantidade de Pontos de Vida Curados",
								Style:       discordgo.TextInputShort,
								Placeholder: "Ex: 8",
								Required:    true,
								MaxLength:   4,
							},
						},
					},
				},
			},
		})
		return
	}

	// 4. Descanso curto e longo via botão
	if customID == "sheet:rest:short" {
		sheet.ShortRest()
		_ = h.service.Save(sheet)
		h.updateSheetMessage(s, i, sheet, "overview")
		return
	}

	if customID == "sheet:rest:long" {
		sheet.LongRest()
		_ = h.service.Save(sheet)
		h.updateSheetMessage(s, i, sheet, "overview")
		return
	}

	// 5. Gastar slot de magia rápido
	if customID == "sheet:spell:use_menu" {
		// Tenta gastar o menor nível disponível (1 a 9)
		usedLevel := 0
		for l := 1; l <= 9; l++ {
			if slot, ok := sheet.Spellcasting.Slots[fmt.Sprintf("%d", l)]; ok && slot.Used < slot.Max {
				_ = sheet.UseSpellSlot(l)
				_ = h.service.Save(sheet)
				usedLevel = l
				break
			}
		}

		if usedLevel > 0 {
			h.updateSheetMessage(s, i, sheet, "spells")
		} else {
			h.respondEphemeral(s, i, "⚠️ Nenhum espaço de magia disponível para gastar!")
		}
		return
	}
}

// HandleModalSubmit processa envios de formulários modais do Discord
func (h *CharacterHandler) HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.ModalSubmitData().CustomID
	userID := i.Member.User.ID
	if i.Member == nil && i.User != nil {
		userID = i.User.ID
	}

	switch customID {
	case "sheet:modal:damage":
		valStr := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		val, err := strconv.Atoi(strings.TrimSpace(valStr))
		if err != nil || val <= 0 {
			h.respondEphemeral(s, i, "Por favor, digite um número válido maior que zero.")
			return
		}
		sheet, lost, isDown, _ := h.service.ApplyDamage(userID, val)
		msg := fmt.Sprintf("💥 **%s** sofreu **%d de dano**! (HP: %d/%d)", sheet.Name, lost, sheet.CurrentHP, sheet.MaxHP)
		if isDown {
			msg += "\n💀 **ATENÇÃO:** O personagem caiu a **0 PV**!"
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:    msg,
				Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
				Components: ui.BuildCharacterActionComponents(sheet, "overview"),
			},
		})

	case "sheet:modal:heal":
		valStr := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		val, err := strconv.Atoi(strings.TrimSpace(valStr))
		if err != nil || val <= 0 {
			h.respondEphemeral(s, i, "Por favor, digite um número válido maior que zero.")
			return
		}
		sheet, healed, _ := h.service.ApplyHealing(userID, val)
		msg := fmt.Sprintf("💚 **%s** foi curado em **%d PV**! (HP: %d/%d)", sheet.Name, healed, sheet.CurrentHP, sheet.MaxHP)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:    msg,
				Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
				Components: ui.BuildCharacterActionComponents(sheet, "overview"),
			},
		})
	}
}

// HandleMessageCreate processa mensagens de texto para criação interativa e upload direto de PDF
func (h *CharacterHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if m.Author.Bot {
		return false
	}

	userID := m.Author.ID

	// 1. Intercepta mensagens se o usuário estiver na criação interativa (Wizard)
	if h.wizard.HasActiveSession(userID, m.ChannelID) {
		resp, finishedSheet, err := h.wizard.ProcessInput(userID, m.Content)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Erro: %v", err))
			return true
		}

		if s != nil {
			if finishedSheet != nil {
				// Criação concluída! Envia o Embed rico com botões
				s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
					Content:    resp,
					Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(finishedSheet)},
					Components: ui.BuildCharacterActionComponents(finishedSheet, "overview"),
				})
			} else if resp != "" {
				s.ChannelMessageSend(m.ChannelID, resp)
			}
		}
		return true
	}

	// 2. Detecção automática de envio de PDF de ficha no chat
	if len(m.Attachments) > 0 {
		for _, att := range m.Attachments {
			if strings.HasSuffix(strings.ToLower(att.Filename), ".pdf") {
				// Verifica se a mensagem menciona ficha
				textLow := strings.ToLower(m.Content)
				if strings.Contains(textLow, "ficha") || strings.Contains(textLow, "personagem") || strings.Contains(textLow, "importar") {
					h.importPDFFromAttachment(s, m.ChannelID, userID, att)
					return true
				}
			}
		}
	}

	return false
}

func (h *CharacterHandler) showActiveSheet(s *discordgo.Session, i *discordgo.InteractionCreate, userID, tab string) {
	sheet, ok := h.service.GetActive(userID)
	if !ok {
		h.respondEphemeral(s, i, "⚠️ Nenhuma ficha ativa encontrada para este usuário.\nCrie uma agora com `/ficha criar` ou importe seu PDF com `/ficha importar`!")
		return
	}

	var embed *discordgo.MessageEmbed
	switch tab {
	case "attacks":
		embed = ui.BuildCharacterAttacksEmbed(sheet)
	case "skills", "attributes":
		embed = ui.BuildCharacterSkillsEmbed(sheet)
	case "spells":
		embed = ui.BuildCharacterSpellsEmbed(sheet)
	case "features":
		embed = ui.BuildCharacterFeaturesEmbed(sheet, 0)
	case "inventory":
		embed = ui.BuildCharacterInventoryEmbed(sheet)
	default:
		tab = "overview"
		embed = ui.BuildCharacterOverviewEmbed(sheet)
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: ui.BuildCharacterActionComponents(sheet, tab),
		},
	})
}

func (h *CharacterHandler) updateSheetMessage(s *discordgo.Session, i *discordgo.InteractionCreate, sheet *character.CharacterSheet, tab string) {
	var embed *discordgo.MessageEmbed
	switch tab {
	case "attacks":
		embed = ui.BuildCharacterAttacksEmbed(sheet)
	case "skills", "attributes":
		embed = ui.BuildCharacterSkillsEmbed(sheet)
	case "spells":
		embed = ui.BuildCharacterSpellsEmbed(sheet)
	case "features":
		embed = ui.BuildCharacterFeaturesEmbed(sheet, 0)
	case "inventory":
		embed = ui.BuildCharacterInventoryEmbed(sheet)
	default:
		tab = "overview"
		embed = ui.BuildCharacterOverviewEmbed(sheet)
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: ui.BuildCharacterActionComponents(sheet, tab),
		},
	})
}

func (h *CharacterHandler) handlePDFImport(s *discordgo.Session, i *discordgo.InteractionCreate, userID string, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) == 0 {
		h.respondEphemeral(s, i, "Nenhum arquivo PDF anexado.")
		return
	}

	attID := options[0].Value.(string)
	resolvedAtt := i.ApplicationCommandData().Resolved.Attachments[attID]
	if resolvedAtt == nil {
		h.respondEphemeral(s, i, "Falha ao carregar o anexo do Discord.")
		return
	}

	// Defer imediato para processamento do PDF
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Baixar PDF para pasta temporária
	tempDir := filepath.Join(os.TempDir(), "kenobi_sheets")
	_ = os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%s", resolvedAtt.ID, resolvedAtt.Filename))

	resp, err := http.Get(resolvedAtt.URL)
	if err != nil {
		msg := fmt.Sprintf("Falha ao baixar PDF do Discord: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(tempPath)
	if err != nil {
		msg := fmt.Sprintf("Falha ao criar arquivo temporário: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}
	_, _ = io.Copy(out, resp.Body)
	_ = out.Close()
	defer os.Remove(tempPath)

	// Importa via parser adaptativo
	sheet, err := h.service.ImportFromPDF(userID, tempPath)
	if err != nil {
		msg := fmt.Sprintf("⚠️ Erro ao processar ficha do PDF: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	content := fmt.Sprintf("🎉 **Ficha importada com sucesso do PDF!**\nO objeto 'ficha' de **%s** foi configurado e ativado para você.", sheet.Name)
	embed := ui.BuildCharacterOverviewEmbed(sheet)
	comps := ui.BuildCharacterActionComponents(sheet, "overview")

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &comps,
	})
}

func (h *CharacterHandler) importPDFFromAttachment(s *discordgo.Session, channelID, userID string, att *discordgo.MessageAttachment) {
	tempDir := filepath.Join(os.TempDir(), "kenobi_sheets")
	_ = os.MkdirAll(tempDir, 0755)
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%s_%s", att.ID, att.Filename))

	resp, err := http.Get(att.URL)
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("Falha ao baixar PDF: %v", err))
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(tempPath)
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("Falha ao criar arquivo temporário: %v", err))
		return
	}
	_, _ = io.Copy(out, resp.Body)
	_ = out.Close()
	defer os.Remove(tempPath)

	sheet, err := h.service.ImportFromPDF(userID, tempPath)
	if err != nil {
		s.ChannelMessageSend(channelID, fmt.Sprintf("⚠️ Erro ao processar ficha do PDF: %v", err))
		return
	}

	s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    fmt.Sprintf("🎉 **Ficha de %s importada automaticamente via PDF!**", sheet.Name),
		Embeds:     []*discordgo.MessageEmbed{ui.BuildCharacterOverviewEmbed(sheet)},
		Components: ui.BuildCharacterActionComponents(sheet, "overview"),
	})
}

func (h *CharacterHandler) handleListSheets(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	sheets := h.service.ListByUser(userID)
	if len(sheets) == 0 {
		h.respondEphemeral(s, i, "Você ainda não possui nenhuma ficha cadastrada. Crie com `/ficha criar` ou envie um PDF com `/ficha importar`!")
		return
	}

	active, _ := h.service.GetActive(userID)

	var lines []string
	for idx, sh := range sheets {
		tag := "▫️"
		if active != nil && active.ID == sh.ID {
			tag = "⭐ **[ATIVA]**"
		}
		lines = append(lines, fmt.Sprintf("%d. %s **%s** — Nível %d %s (`%d/%d PV`)",
			idx+1, tag, sh.Name, sh.Level, sh.Class, sh.CurrentHP, sh.MaxHP))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Suas Fichas de RPG",
		Description: strings.Join(lines, "\n\n") + "\n\n*Para trocar a ficha ativa, use `/ficha selecionar <nome>`*",
		Color:       ui.ColorDNDBlue,
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}); err != nil {
		slog.Error("Erro ao responder /ficha listar", "error", err)
	}
}

func (h *CharacterHandler) handleSelectSheet(s *discordgo.Session, i *discordgo.InteractionCreate, userID, query string) {
	sheets := h.service.ListByUser(userID)
	qLow := strings.ToLower(query)

	var target *character.CharacterSheet
	for _, sh := range sheets {
		if strings.EqualFold(sh.ID, query) || strings.Contains(strings.ToLower(sh.Name), qLow) {
			target = sh
			break
		}
	}

	if target == nil {
		h.respondEphemeral(s, i, fmt.Sprintf("Nenhuma ficha encontrada correspondente a '%s'.", query))
		return
	}

	_ = h.service.SetActive(userID, target.ID)
	h.respondText(s, i, fmt.Sprintf("⭐ Ficha ativa definida com sucesso: **%s** (Nível %d %s)!", target.Name, target.Level, target.Class))
}

func (h *CharacterHandler) handleDeleteSheet(s *discordgo.Session, i *discordgo.InteractionCreate, userID, query string) {
	sheets := h.service.ListByUser(userID)
	qLow := strings.ToLower(query)

	var target *character.CharacterSheet
	for _, sh := range sheets {
		if strings.EqualFold(sh.ID, query) || strings.Contains(strings.ToLower(sh.Name), qLow) {
			target = sh
			break
		}
	}

	if target == nil {
		h.respondEphemeral(s, i, fmt.Sprintf("Nenhuma ficha encontrada correspondente a '%s'.", query))
		return
	}

	_ = h.service.Delete(target.ID)
	h.respondText(s, i, fmt.Sprintf("🗑️ Ficha **%s** excluída com sucesso.", target.Name))
}

func (h *CharacterHandler) handleSheetRoll(s *discordgo.Session, i *discordgo.InteractionCreate, userID, testName string) {
	sheet, ok := h.service.GetActive(userID)
	if !ok {
		h.respondEphemeral(s, i, "⚠️ Você não possui uma ficha ativa. Use `/ficha criar` ou `/ficha importar`.")
		return
	}

	tLow := strings.ToLower(testName)
	mod := 0
	label := testName

	// 1. Perícia
	foundSkill := false
	for k, sk := range sheet.Skills {
		if strings.Contains(strings.ToLower(k), tLow) {
			mod = sk.Modifier
			label = k
			foundSkill = true
			break
		}
	}

	// 2. Atributo ou Salvaguarda
	if !foundSkill {
		switch {
		case strings.Contains(tLow, "for"):
			mod = sheet.Attributes.STR.Modifier
			label = "Força"
		case strings.Contains(tLow, "des"):
			mod = sheet.Attributes.DEX.Modifier
			label = "Destreza"
		case strings.Contains(tLow, "con"):
			mod = sheet.Attributes.CON.Modifier
			label = "Constituição"
		case strings.Contains(tLow, "int"):
			mod = sheet.Attributes.INT.Modifier
			label = "Inteligência"
		case strings.Contains(tLow, "sab"):
			mod = sheet.Attributes.WIS.Modifier
			label = "Sabedoria"
		case strings.Contains(tLow, "car"):
			mod = sheet.Attributes.CHA.Modifier
			label = "Carisma"
		case strings.Contains(tLow, "inic"):
			mod = sheet.Initiative
			label = "Iniciativa"
		}
	}

	advMode := sheet.GetAdvantageMode()
	var d20 int
	rollBreakdown := ""

	switch advMode {
	case "advantage":
		d1 := rand.Intn(20) + 1
		d2 := rand.Intn(20) + 1
		d20 = d1
		discarded := d2
		if d2 > d1 {
			d20 = d2
			discarded = d1
		}
		rollBreakdown = fmt.Sprintf("`dados [%d, %d] ➔ Maior: %d %+d` *(descartou %d)*", d1, d2, d20, mod, discarded)
	case "disadvantage":
		d1 := rand.Intn(20) + 1
		d2 := rand.Intn(20) + 1
		d20 = d1
		discarded := d2
		if d2 < d1 {
			d20 = d2
			discarded = d1
		}
		rollBreakdown = fmt.Sprintf("`dados [%d, %d] ➔ Menor: %d %+d` *(descartou %d)*", d1, d2, d20, mod, discarded)
	default:
		d20 = rand.Intn(20) + 1
		rollBreakdown = fmt.Sprintf("`d20 (%d) %+d`", d20, mod)
	}

	total := d20 + mod

	resultEmoji := "🎲"
	switch d20 {
	case 20:
		resultEmoji = "🔥 **SUCESSO DECISIVO (NAT 20)!**"
	case 1:
		resultEmoji = "💀 **FALHA CRÍTICA (NAT 1)!**"
	}

	advTag := ""
	switch advMode {
	case "advantage":
		advTag = " *(com ✨ Vantagem)*"
	case "disadvantage":
		advTag = " *(com 💀 Desvantagem)*"
	}

	msg := fmt.Sprintf("%s **%s** rolou um teste de **%s**%s:\n%s = **%d**",
		resultEmoji, sheet.Name, label, advTag, rollBreakdown, total)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}

func (h *CharacterHandler) respondText(s *discordgo.Session, i *discordgo.InteractionCreate, text string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func (h *CharacterHandler) respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, text string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// formatDamageBreakdown formata a discriminação dos dados de dano rolados e modificadores
func formatDamageBreakdown(roll *dice.RollResult, dmgType string) string {
	if roll == nil {
		return ""
	}
	var diceStr []string
	for _, r := range roll.Rolls {
		diceStr = append(diceStr, strconv.Itoa(r))
	}
	diceFormatted := "[" + strings.Join(diceStr, ", ") + "]"

	modStr := ""
	if roll.Modifier > 0 {
		modStr = fmt.Sprintf(" + %d", roll.Modifier)
	} else if roll.Modifier < 0 {
		modStr = fmt.Sprintf(" - %d", -roll.Modifier)
	}

	typeStr := ""
	if dmgType != "" {
		typeStr = " " + dmgType
	}

	if modStr == "" && len(roll.Rolls) == 1 {
		return fmt.Sprintf("Dado **%s** = **`%d`**%s", diceFormatted, roll.Total, typeStr)
	}
	return fmt.Sprintf("Dados **%s**%s = **`%d`**%s", diceFormatted, modStr, roll.Total, typeStr)
}

// hasSavageAttacker verifica se a ficha possui o talento/habilidade Atacante Selvagem
func hasSavageAttacker(sheet *character.CharacterSheet) bool {
	if sheet == nil {
		return false
	}
	for _, f := range sheet.Features {
		name := strings.ToLower(f.Name)
		if strings.Contains(name, "atacante selvagem") || strings.Contains(name, "savage attacker") {
			return true
		}
	}
	return false
}

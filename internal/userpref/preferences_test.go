package userpref_test

import (
	"testing"

	"kenobi-bot/internal/userpref"
)

func TestUserPreferences(t *testing.T) {
	service := userpref.NewMemoryService()

	t.Run("Preferências padrão de novo usuário (tudo desativado)", func(t *testing.T) {
		pref := service.GetPreferences("user_100")
		if pref.RerollOn1All != false {
			t.Errorf("RerollOn1All padrão deveria ser falso")
		}
		if pref.HalflingLuck != false {
			t.Errorf("HalflingLuck padrão deveria ser falso")
		}
	})

	t.Run("ToggleRerollOn1All ativa e desativa corretamente", func(t *testing.T) {
		newState := service.ToggleRerollOn1All("user_100")
		if newState != true {
			t.Errorf("Esperado true após primeiro toggle, obteve %v", newState)
		}
		pref := service.GetPreferences("user_100")
		if pref.RerollOn1All != true {
			t.Errorf("Preferência persistida deveria ser true")
		}

		newState2 := service.ToggleRerollOn1All("user_100")
		if newState2 != false {
			t.Errorf("Esperado false após segundo toggle, obteve %v", newState2)
		}
	})

	t.Run("ToggleHalflingLuck ativa e desativa corretamente", func(t *testing.T) {
		newState := service.ToggleHalflingLuck("user_200")
		if newState != true {
			t.Errorf("Esperado true após toggle de Halfling Luck, obteve %v", newState)
		}
		pref := service.GetPreferences("user_200")
		if pref.HalflingLuck != true {
			t.Errorf("Preferência persistida deveria ser true")
		}
	})

	t.Run("SuccessThreshold padrão e customizado", func(t *testing.T) {
		pref := service.GetPreferences("user_300")
		if pref.SuccessThreshold != userpref.DefaultSuccessThreshold {
			t.Errorf("SuccessThreshold padrão deveria ser %d, obteve %d", userpref.DefaultSuccessThreshold, pref.SuccessThreshold)
		}

		updated := service.SetSuccessThreshold("user_300", 6)
		if updated != 6 {
			t.Errorf("Esperado 6 após SetSuccessThreshold, obteve %d", updated)
		}

		prefUpdated := service.GetPreferences("user_300")
		if prefUpdated.SuccessThreshold != 6 {
			t.Errorf("SuccessThreshold persistido deveria ser 6, obteve %d", prefUpdated.SuccessThreshold)
		}
	})
}

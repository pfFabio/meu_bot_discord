package repository_test

import (
	"path/filepath"
	"testing"

	"kenobi-bot/internal/dnd/repository"
)

func TestLoadDNDData(t *testing.T) {
	// Caminho relativo a partir de internal/dnd/repository
	basePath, err := filepath.Abs("../../../")
	if err != nil {
		t.Fatalf("Erro ao obter caminho base: %v", err)
	}

	repo, err := repository.LoadDNDData(basePath)
	if err != nil {
		t.Fatalf("LoadDNDData() retornou erro inesperado: %v", err)
	}
	if repo == nil {
		t.Fatalf("LoadDNDData() retornou repo nil")
	}

	t.Run("Valida carregamento de Classes (mínimo 13 classes)", func(t *testing.T) {
		classes := repo.GetClasses()
		if len(classes) < 13 {
			t.Errorf("Total de classes carregadas = %d, esperado pelo menos 13", len(classes))
		}
	})

	t.Run("Valida carregamento de Magias do SRD (mínimo 300 magias)", func(t *testing.T) {
		spells := repo.SearchSpells("", -1, "", "", 0)
		if len(spells) < 300 {
			t.Errorf("Total de magias carregadas = %d, esperado pelo menos 300", len(spells))
		}
	})

	t.Run("Valida carregamento de Monstros do SRD (mínimo 150 monstros)", func(t *testing.T) {
		monsters := repo.SearchMonsters("", "", "", 0)
		if len(monsters) < 150 {
			t.Errorf("Total de monstros carregados = %d, esperado pelo menos 150", len(monsters))
		}
	})

	t.Run("Valida carregamento de Regras do SRD (mínimo 100 regras)", func(t *testing.T) {
		rules := repo.SearchRules("", 0)
		if len(rules) < 100 {
			t.Errorf("Total de regras carregadas = %d, esperado pelo menos 100", len(rules))
		}
	})

	t.Run("Valida carregamento de Espécies de Ravenloft e D&D 2024", func(t *testing.T) {
		species := repo.GetSpeciesList("all")
		if len(species) < 14 {
			t.Errorf("Total de espécies carregadas = %d, esperado pelo menos 14", len(species))
		}

		ravenloft := repo.GetSpeciesList("Ravenloft")
		if len(ravenloft) < 4 {
			t.Errorf("Espécies de Ravenloft carregadas = %d, esperado pelo menos 4", len(ravenloft))
		}

		dhampir, found := repo.GetSpecies("dhampir")
		if !found || dhampir.Name != "Dhampir" {
			t.Errorf("Dhampir não encontrado no repositório")
		}

		lupin, foundLupin := repo.GetSpecies("lupin")
		if !foundLupin || lupin.Name != "Lupin" {
			t.Errorf("Lupin não encontrado no repositório")
		}
	})
}


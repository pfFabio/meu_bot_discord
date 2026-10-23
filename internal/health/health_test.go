package health_test

import (
	"testing"

	"kenobi-bot/internal/health"
)

func TestDiagnostics(t *testing.T) {
	report := health.RunDiagnostics("../..")

	t.Log(report.FormatConsole())

	if report.TotalPass == 0 {
		t.Errorf("Esperava que pelo menos algumas checagens passassem, total_pass: 0")
	}

	if report.HasCriticalFailures() {
		t.Errorf("Diagnóstico encontrou %d falhas críticas impeditivas para a execução do bot", report.TotalFail)
	}
}

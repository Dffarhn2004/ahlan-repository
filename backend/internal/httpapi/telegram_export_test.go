package httpapi

import (
	"testing"
	"time"
)

func TestTelegramExportWindow(t *testing.T) {
	start, end, err := telegramExportWindow("", "")
	if err != nil || start != nil || end != nil {
		t.Fatalf("empty range = %v %v %v", start, end, err)
	}
	start, end, err = telegramExportWindow("2026-11-11", "2026-11-11")
	if err != nil || start == nil || end == nil {
		t.Fatal(err)
	}
	if start.Format("2006-01-02 15:04 -0700") != "2026-11-11 00:00 +0700" || end.Format("2006-01-02 15:04 -0700") != "2026-11-12 00:00 +0700" {
		t.Fatalf("same-day window = %s .. %s", start, end)
	}
	if _, _, err = telegramExportWindow("2026-11-12", "2026-11-11"); err == nil {
		t.Fatal("reversed range should fail")
	}
	if _, _, err = telegramExportWindow("11-11-2026", ""); err == nil {
		t.Fatal("bad date should fail")
	}
}

func TestFilterExportTurnsKeepsTheQuestionDay(t *testing.T) {
	start, end, err := telegramExportWindow("2026-11-11", "2026-11-11")
	if err != nil {
		t.Fatal(err)
	}
	turns := []telegramExportTurn{
		{Question: "kemarin", CreatedAt: start.AddDate(0, 0, -1)},
		{Question: "hari ini", CreatedAt: start.Add(3 * time.Hour)},
		{Question: "besok", CreatedAt: *end},
	}
	got := filterExportTurns(turns, start, end)
	if len(got) != 1 || got[0].Question != "hari ini" {
		t.Fatalf("filtered = %+v", got)
	}
}

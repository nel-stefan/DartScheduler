package excel_test

import (
	"bytes"
	"testing"

	"DartScheduler/infra/excel"

	"github.com/xuri/excelize/v2"
)

// buildLedenlijst creates an in-memory Excel that mimics the Dutch ledenlijst format.
func buildLedenlijst(t *testing.T, rows [][]string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()

	header := []string{"nr", "Naam", "Adres", "Pc", "Woonpl.", "Telefoon", "Mobiel", "E-mail adres", "Lid Sinds", "Klasse"}
	all := append([][]string{header}, rows...)
	for i, row := range all {
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue("Sheet1", cell, val)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	return &buf
}

func TestImportPlayers_DutchFormat(t *testing.T) {
	buf := buildLedenlijst(t, [][]string{
		{"1", "Jansen, Jan", "Hoofdstraat 1", "1234AB", "Amsterdam", "020-1234567", "06-12345678", "jan@example.com", "2021", "1"},
		{"2", "Pietersen, Piet", "Kerkstraat 2", "5678CD", "Rotterdam", "", "06-98765432", "piet@example.com", "2019", "2"},
		{"3", "De Vries, Klaas", "Molenweg 3", "9012EF", "Utrecht", "", "", "", "2020", ""},
	})

	players, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 3 {
		t.Fatalf("want 3 players, got %d", len(players))
	}

	p := players[0]
	if p.Nr != "1" {
		t.Errorf("players[0].Nr = %q, want %q", p.Nr, "1")
	}
	if p.Name != "Jansen, Jan" {
		t.Errorf("players[0].Name = %q, want %q", p.Name, "Jansen, Jan")
	}
	if p.Class != "1" {
		t.Errorf("players[0].Class = %q, want %q", p.Class, "1")
	}
	if p.Email != "jan@example.com" {
		t.Errorf("players[0].Email = %q, want %q", p.Email, "jan@example.com")
	}
	if p.City != "Amsterdam" {
		t.Errorf("players[0].City = %q, want %q", p.City, "Amsterdam")
	}
	if p.MemberSince != "2021" {
		t.Errorf("players[0].MemberSince = %q, want %q", p.MemberSince, "2021")
	}

	if players[1].Class != "2" {
		t.Errorf("players[1].Class = %q, want %q", players[1].Class, "2")
	}
	if players[2].Class != "" {
		t.Errorf("players[2].Class = %q, want empty", players[2].Class)
	}
}

func TestImportPlayers_SponsorsSkipped(t *testing.T) {
	buf := buildLedenlijst(t, [][]string{
		{"1", "Lid Een", "", "", "", "", "", "", "2020", ""},
		{"2-s", "Sponsor A", "", "", "", "", "", "", "2010", ""},
		{"3", "Lid Drie", "", "", "", "", "", "", "2022", ""},
		{"10-s", "Sponsor B", "", "", "", "", "", "", "2005", ""},
		{"4", "Lid Vier", "", "", "", "", "", "", "2023", ""},
	})

	players, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 3 {
		t.Fatalf("want 3 players (sponsors excluded), got %d", len(players))
	}
	for _, p := range players {
		if p.Nr == "2-s" || p.Nr == "10-s" {
			t.Errorf("sponsor %q should have been skipped", p.Nr)
		}
	}
}

func TestImportPlayers_EmptyRowsSkipped(t *testing.T) {
	buf := buildLedenlijst(t, [][]string{
		{"1", "Lid Een", "", "", "", "", "", "", "", ""},
		{"", "", "", "", "", "", "", "", "", ""}, // empty row
		{"2", "Lid Twee", "", "", "", "", "", "", "", ""},
	})

	players, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("want 2 players (empty row skipped), got %d", len(players))
	}
}

func TestImportPlayers_MissingNaamColumn(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()
	_ = f.SetCellValue("Sheet1", "A1", "nr")
	_ = f.SetCellValue("Sheet1", "A2", "1")
	var buf bytes.Buffer
	_ = f.Write(&buf)

	_, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("expected error for missing Naam column, got nil")
	}
}

func TestImportPlayers_NoDataRows(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()
	_ = f.SetCellValue("Sheet1", "A1", "nr")
	_ = f.SetCellValue("Sheet1", "B1", "Naam")
	var buf bytes.Buffer
	_ = f.Write(&buf)

	_, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("expected error for file with only a header row, got nil")
	}
}

func TestImportPlayers_EnglishFormat(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()
	for j, h := range []string{"Name", "Email", "Sponsor"} {
		cell, _ := excelize.CoordinatesToCellName(j+1, 1)
		_ = f.SetCellValue("Sheet1", cell, h)
	}
	for j, v := range []string{"Smith, John", "john@example.com", "ACME"} {
		cell, _ := excelize.CoordinatesToCellName(j+1, 2)
		_ = f.SetCellValue("Sheet1", cell, v)
	}
	var buf bytes.Buffer
	_ = f.Write(&buf)

	players, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("want 1 player, got %d", len(players))
	}
	if players[0].Name != "Smith, John" {
		t.Errorf("Name = %q, want %q", players[0].Name, "Smith, John")
	}
	if players[0].Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", players[0].Email, "john@example.com")
	}
}

func TestImportPlayers_TrimWhitespace(t *testing.T) {
	buf := buildLedenlijst(t, [][]string{
		{"  5  ", "  Van Eck, Rob  ", "  Dorpstraat 10  ", "", "", "", "", "", "", "1"},
	})

	players, err := excel.ImportPlayers(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("want 1 player, got %d", len(players))
	}
	if players[0].Nr != "5" {
		t.Errorf("Nr = %q, want %q", players[0].Nr, "5")
	}
	if players[0].Name != "Van Eck, Rob" {
		t.Errorf("Name = %q, want %q", players[0].Name, "Van Eck, Rob")
	}
	if players[0].Address != "Dorpstraat 10" {
		t.Errorf("Address = %q, want %q", players[0].Address, "Dorpstraat 10")
	}
}

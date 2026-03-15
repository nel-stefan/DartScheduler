package excel

import (
	"fmt"
	"io"
	"strings"

	"DartScheduler/usecase"

	"github.com/xuri/excelize/v2"
)

// ImportPlayers reads an Excel file from r.
// Supports the Dutch ledenlijst format with columns:
//
//	nr, Naam, Adres, Pc, Woonpl., Telefoon, Mobiel, E-mail adres, Lid Sinds, Samen, Klasse
//
// "Samen" (col J) contains the member nr of the player's buddy.
// "Klasse" (col K) contains the player's competition class.
// Returns PlayerInput slice ready for PlayerUseCase.ImportPlayers.
func ImportPlayers(r io.Reader) ([]usecase.PlayerInput, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("excel: open: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel: no sheets found")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("excel: read rows: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("%w: Excel file has no data rows", usecase.ErrImport)
	}

	// Detect column indices from header row.
	header := rows[0]
	colNr, colName, colEmail, colSponsor := -1, -1, -1, -1
	colAddress, colPostalCode, colCity := -1, -1, -1
	colPhone, colMobile, colMemberSince, colClass, colSamen := -1, -1, -1, -1, -1

	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "nr":
			colNr = i
		case "naam", "name":
			colName = i
		case "e-mail adres", "e-mailadres", "email adres", "email":
			colEmail = i
		case "sponsor":
			colSponsor = i
		case "adres":
			colAddress = i
		case "pc", "postcode":
			colPostalCode = i
		case "woonpl.", "woonpl", "woonplaats":
			colCity = i
		case "telefoon":
			colPhone = i
		case "mobiel":
			colMobile = i
		case "lid sinds":
			colMemberSince = i
		case "klasse", "class":
			colClass = i
		case "samen":
			colSamen = i
		}
	}
	if colName < 0 {
		return nil, fmt.Errorf("%w: missing 'Naam' or 'Name' column", usecase.ErrImport)
	}

	cell := func(row []string, col int) string {
		if col < 0 || col >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[col])
	}

	var out []usecase.PlayerInput
	for _, row := range rows[1:] {
		name := cell(row, colName)
		if name == "" {
			continue
		}
		nr := cell(row, colNr)
		if strings.Contains(strings.ToLower(nr), "-s") {
			continue // skip sponsor members
		}
		out = append(out, usecase.PlayerInput{
			Nr:          nr,
			Name:        name,
			Email:       cell(row, colEmail),
			Sponsor:     cell(row, colSponsor),
			Address:     cell(row, colAddress),
			PostalCode:  cell(row, colPostalCode),
			City:        cell(row, colCity),
			Phone:       cell(row, colPhone),
			Mobile:      cell(row, colMobile),
			MemberSince: cell(row, colMemberSince),
			Class:       cell(row, colClass),
			SamenNr:     cell(row, colSamen),
		})
	}
	return out, nil
}

package excel

import (
	"fmt"
	"io"
	"strings"

	"DartScheduler/usecase"

	"github.com/xuri/excelize/v2"
)

// ImportPlayers reads an Excel file from r.
// Expected columns (row 1 = header): Name, Email, Sponsor
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
	colName, colEmail, colSponsor := -1, -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "name":
			colName = i
		case "email":
			colEmail = i
		case "sponsor":
			colSponsor = i
		}
	}
	if colName < 0 {
		return nil, fmt.Errorf("%w: missing 'Name' column", usecase.ErrImport)
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
		out = append(out, usecase.PlayerInput{
			Name:    name,
			Email:   cell(row, colEmail),
			Sponsor: cell(row, colSponsor),
		})
	}
	return out, nil
}

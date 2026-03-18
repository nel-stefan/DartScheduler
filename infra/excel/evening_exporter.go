package excel

import (
	"context"
	"fmt"
	"io"
	"strings"

	"DartScheduler/domain"

	"github.com/xuri/excelize/v2"
)

const (
	fontCalibri = "Calibri"
	colorBlack  = "000000"
	styleThin   = 1 // thin continuous border   (excelize index 1)
	styleMedium = 2 // medium continuous border  (excelize index 2)
	styleThick  = 5 // thick continuous border   (excelize index 5)
)

// EveningExporter implements usecase.EveningExporter for the wedstrijdformulier Excel format.
type EveningExporter struct{}

func (e EveningExporter) ExportEvening(ctx context.Context, sched domain.Schedule, ev domain.Evening, players []domain.Player, w io.Writer) error {
	return ExportEvening(ctx, sched, ev, players, w)
}

// ExportEvening writes a single evening to w in the Dartclub Grolzicht
// wedstrijdformulier format (matches the "Blanco wedstrijformulier.xlsx" layout).
//
// Build order per section to ensure correct Excel rendering:
//  1. Row heights
//  2. Column widths
//  3. Merges  (MergeCell clears styles, so must come before SetCellStyle)
//  4. Cell values
//  5. Cell styles / borders  (always last, after all merges)
func ExportEvening(ctx context.Context, sched domain.Schedule, ev domain.Evening, players []domain.Player, w io.Writer) error {
	playerMap := make(map[string]domain.Player, len(players))
	for _, p := range players {
		playerMap[p.ID.String()] = p
	}
	firstNameNr := func(p domain.Player) string {
		// Names are stored as "Achternaam, Voornaam"; first name is after the comma.
		firstName := p.Name
		if parts := strings.SplitN(p.Name, ", ", 2); len(parts) == 2 {
			firstName = strings.SplitN(parts[1], " ", 2)[0]
		}
		return firstName + " - " + p.Nr
	}

	playerLabel := func(id string) string {
		p, ok := playerMap[id]
		if !ok || id == "" {
			return ""
		}
		return firstNameNr(p)
	}

	// reportedByLabel converts a stored "nr ..." string to "Voornaam - nr"
	// by matching on player number. Falls back to the raw value for free-text entries.
	reportedByLabel := func(raw string) string {
		if raw == "" {
			return ""
		}
		for _, p := range players {
			if p.Nr == "" {
				continue
			}
			if strings.HasPrefix(raw, p.Nr+" ") || strings.HasPrefix(raw, p.Nr+"\t") {
				return firstNameNr(p)
			}
		}
		return raw
	}

	f := excelize.NewFile()
	defer f.Close()
	ws := "Blad1"
	f.SetSheetName("Sheet1", ws)

	brd := func(l, r, t, b int) []excelize.Border {
		var borders []excelize.Border
		if l > 0 {
			borders = append(borders, excelize.Border{Type: "left", Color: colorBlack, Style: l})
		}
		if r > 0 {
			borders = append(borders, excelize.Border{Type: "right", Color: colorBlack, Style: r})
		}
		if t > 0 {
			borders = append(borders, excelize.Border{Type: "top", Color: colorBlack, Style: t})
		}
		if b > 0 {
			borders = append(borders, excelize.Border{Type: "bottom", Color: colorBlack, Style: b})
		}
		return borders
	}

	ns := func(s *excelize.Style) int {
		id, _ := f.NewStyle(s)
		return id
	}

	hdrCenter := &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}

	// hdrStyle returns a style ID for a header cell: bold Calibri 9 with the given borders.
	hdrStyle := func(l, r, t, b int) int {
		return ns(&excelize.Style{
			Font:      &excelize.Font{Family: fontCalibri, Bold: true, Size: 9},
			Alignment: hdrCenter,
			Border:    brd(l, r, t, b),
		})
	}

	// ================================================================ 1. ROW HEIGHTS
	f.SetRowHeight(ws, 1, 21)
	f.SetRowHeight(ws, 2, 15)
	f.SetRowHeight(ws, 3, 13.5)
	f.SetRowHeight(ws, 4, 15)
	f.SetRowHeight(ws, 5, 13.5)
	f.SetRowHeight(ws, 6, 13.5)

	// ================================================================ 2. COLUMN WIDTHS
	maxNameLen := 10
	for _, p := range players {
		if l := len(domain.FormatDisplayName(p.Name)); l > maxNameLen {
			maxNameLen = l
		}
	}
	nameColWidth := float64(maxNameLen)*1.0 + 1.5 // ≈ autofit: 1 unit/char + padding

	for col, width := range map[string]float64{
		"A": 4.0, "B": nameColWidth, "C": 1.7109, "D": 4.0, "E": nameColWidth,
		"F": 14.4258, "G": 6.5, "H": 14.4258, "I": 6.5,
		"J": 14.4258, "K": 6.5, "L": 13.8555, "M": 5.5703,
		"N": 12.1406, "O": 7.8555, "P": 6.1406, "Q": 6.1406,
	} {
		f.SetColWidth(ws, col, col, width)
	}

	// ================================================================ 3. MERGES
	f.MergeCell(ws, "A1", "Q1")
	f.MergeCell(ws, "A2", "Q2")

	for _, m := range [][2]string{
		{"A4", "A6"}, {"B4", "B6"}, {"C4", "C6"}, {"D4", "D6"}, {"E4", "E6"},
		{"F4", "G4"}, {"F5", "F6"}, {"G5", "G6"},
		{"H4", "I4"}, {"H5", "H6"}, {"I5", "I6"},
		{"J4", "K4"}, {"J5", "J6"}, {"K5", "K6"},
		{"L4", "L6"}, {"M4", "M6"}, {"N4", "N6"}, {"O4", "O6"}, {"P4", "P6"}, {"Q4", "Q6"},
	} {
		f.MergeCell(ws, m[0], m[1])
	}

	// ================================================================ 4. CELL VALUES
	f.SetCellValue(ws, "A1", "DARTCLUB GROLZICHT")

	dateLabel := ev.Date.Format("2-1-2006")
	if ev.IsCatchUpEvening {
		dateLabel = "Inhaal"
	}
	f.SetCellValue(ws, "A2", fmt.Sprintf(
		"Wedstrijdformulier     Spelsoort: 501 dubbel uit best of 3     Speeldatum: %s",
		dateLabel))

	// Rows 4-6: header labels
	f.SetCellValue(ws, "A4", "nr")
	f.SetCellValue(ws, "B4", "naam")
	f.SetCellValue(ws, "C4", "/")
	f.SetCellValue(ws, "D4", "nr")
	f.SetCellValue(ws, "E4", "naam")
	f.SetCellValue(ws, "F4", "Leg 1")
	f.SetCellValue(ws, "F5", "voornaam+nr.")
	f.SetCellValue(ws, "G5", "aantal beurten")
	f.SetCellValue(ws, "H4", "Leg 2")
	f.SetCellValue(ws, "H5", "voornaam+nr.")
	f.SetCellValue(ws, "I5", "aantal beurten")
	f.SetCellValue(ws, "J4", "Leg 3")
	f.SetCellValue(ws, "J5", "voornaam+nr.")
	f.SetCellValue(ws, "K5", "aantal beurten")
	f.SetCellValue(ws, "L4", "totaal\nwinnaar")
	f.SetCellValue(ws, "M4", "eind-\nstand")
	f.SetCellValue(ws, "N4", "afgemeld\ndoor")
	f.SetCellValue(ws, "O4", "vooruit-\ngooi\ndatum")
	f.SetCellValue(ws, "P4", "nr.\nschrij-\nver")
	f.SetCellValue(ws, "Q4", "nr.\ntel-\nler")

	// ================================================================ 5. CELL STYLES / BORDERS
	// Row 1
	f.SetCellStyle(ws, "A1", "Q1", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 16, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	}))

	// Row 2
	f.SetCellStyle(ws, "A2", "Q2", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	}))

	// Row 3: thick bottom rule on specific columns only
	row3Style := ns(&excelize.Style{
		Font:   &excelize.Font{Family: fontCalibri, Size: 11},
		Border: brd(0, 0, 0, styleThick),
	})
	for _, cell := range []string{"B3", "D3", "H3", "L3", "M3", "N3"} {
		f.SetCellStyle(ws, cell, cell, row3Style)
	}

	// Rows 4-6: column headers
	f.SetCellStyle(ws, "A4", "A6", hdrStyle(styleThick, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "B4", "B6", hdrStyle(0, styleThick, styleThick, styleThin))
	f.SetCellStyle(ws, "C4", "C6", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 10},
		Alignment: hdrCenter,
		Border:    brd(styleThick, styleThick, styleThick, styleThin),
	}))
	f.SetCellStyle(ws, "D4", "D6", hdrStyle(styleThick, 0, styleThick, styleThick))
	f.SetCellStyle(ws, "E4", "E6", hdrStyle(styleThick, styleThick, styleThick, styleThin))
	f.SetCellStyle(ws, "F4", "G4", hdrStyle(styleThick, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "F5", "F6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "G5", "G6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "H4", "I4", hdrStyle(styleThick, styleMedium, styleThick, 0))
	f.SetCellStyle(ws, "H5", "H6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "I5", "I6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "J4", "K4", hdrStyle(0, styleMedium, styleThick, 0))
	f.SetCellStyle(ws, "J5", "J6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "K5", "K6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "L4", "L6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "M4", "M6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellStyle(ws, "N4", "N6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.SetCellStyle(ws, "O4", "O6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.SetCellStyle(ws, "P4", "P6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.SetCellStyle(ws, "Q4", "Q6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))

	// ================================================================ DATA ROWS
	// colSpec defines per-column style: font size, horizontal align, shrink-to-fit, and border sides.
	type colSpec struct {
		sz          float64
		ha          string
		shrinkToFit bool
		l, r, t, b  int
	}

	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q"}

	// Row 7: first data row — thick top on ALL columns.
	row7 := []colSpec{
		{12, "right", false, styleThick, styleThin, styleThick, styleThin},     // A: nr
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},         // B: naam
		{10, "center", false, styleMedium, styleMedium, styleThick, styleThin}, // C: /
		{12, "right", false, styleMedium, styleThin, styleThick, styleThin},    // D: nr
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},         // E: naam
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},          // F: leg1 winner
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},         // G: leg1 turns
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},          // H: leg2 winner
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},         // I: leg2 turns
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},          // J: leg3 winner
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},         // K: leg3 turns
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},          // L: totaal winnaar
		{11, "center", false, styleThin, styleMedium, styleThick, styleThin},   // M: eindstand
		{11, "", true, styleMedium, styleMedium, styleThick, styleThin},        // N: afgemeld door
		{11, "", false, styleMedium, styleMedium, styleThick, styleThin},       // O: vooruitgooi datum
		{11, "center", false, styleMedium, styleMedium, styleThick, styleThin}, // P: nr. schrijver
		{11, "center", false, styleMedium, styleThick, styleThick, styleThin},  // Q: nr. teller
	}

	// Rows 8+: subsequent data rows.
	rowN := []colSpec{
		{12, "right", false, styleThick, styleThin, 0, styleThin},             // A
		{11, "", false, styleThin, styleMedium, styleThin, styleThin},         // B
		{10, "center", false, styleMedium, styleMedium, styleThin, styleThin}, // C
		{12, "right", false, styleMedium, styleThin, 0, styleThin},            // D
		{11, "", false, styleThin, styleMedium, styleThin, styleThin},         // E
		{11, "", true, styleMedium, styleThin, 0, styleThin},                  // F
		{11, "", false, styleThin, styleMedium, 0, styleThin},                 // G
		{11, "", true, styleMedium, styleThin, 0, styleThin},                  // H
		{11, "", false, styleThin, styleMedium, 0, styleThin},                 // I
		{11, "", true, styleMedium, styleThin, 0, styleThin},                  // J
		{11, "", false, styleThin, styleMedium, 0, styleThin},                 // K
		{11, "", true, styleMedium, styleThin, 0, styleThin},                  // L
		{11, "center", false, styleThin, styleMedium, 0, styleThin},           // M
		{11, "", true, styleMedium, styleMedium, styleThin, styleThin},        // N
		{11, "", false, styleMedium, styleMedium, styleThin, styleThin},       // O
		{11, "center", false, styleMedium, styleMedium, styleThin, styleThin}, // P
		{11, "center", false, styleMedium, styleThick, styleThin, styleThin},  // Q
	}

	buildStyles := func(specs []colSpec) []int {
		ids := make([]int, len(specs))
		for i, sp := range specs {
			ids[i], _ = f.NewStyle(&excelize.Style{
				Font: &excelize.Font{Family: fontCalibri, Size: sp.sz},
				Alignment: &excelize.Alignment{
					Horizontal:  sp.ha,
					ShrinkToFit: sp.shrinkToFit,
				},
				Border: brd(sp.l, sp.r, sp.t, sp.b),
			})
		}
		return ids
	}

	// rowLast is identical to rowN but with a thick bottom border on every column.
	rowLast := make([]colSpec, len(rowN))
	copy(rowLast, rowN)
	for i := range rowLast {
		rowLast[i].b = styleThick
	}

	row7Styles := buildStyles(row7)
	rowNStyles := buildStyles(rowN)
	rowLastStyles := buildStyles(rowLast)

	// Fill to the bottom of the last page so blank rows are always available
	// for manual entries. Approximate page capacity for A4 landscape with
	// narrow margins (0.75") and data rows at 17.25pt each:
	//   page 1: 6-row header (91.5pt) leaves ~20 data rows
	//   page 2+: header repeated via print titles → same ~20 rows per page
	const rowsPerPage = 20
	matchCount := len(ev.Matches)
	pages := (matchCount + rowsPerPage - 1) / rowsPerPage
	if pages < 1 {
		pages = 1
	}
	totalRows := pages * rowsPerPage

	emptyValues := []interface{}{"", "", "/", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}

	// Pass A: row heights + cell values (no styles yet — styles come after in pass B).
	for i := 0; i < totalRows; i++ {
		row := 7 + i
		f.SetRowHeight(ws, row, 17.25)

		var values []interface{}
		if i < len(ev.Matches) {
			m := ev.Matches[i]
			pA := playerMap[m.PlayerA.String()]
			pB := playerMap[m.PlayerB.String()]

			totalWinner, eindstand := "", ""
			if m.Played && m.ScoreA != nil && m.ScoreB != nil {
				eindstand = fmt.Sprintf("%d-%d", *m.ScoreA, *m.ScoreB)
				if *m.ScoreA > *m.ScoreB {
					totalWinner = playerLabel(m.PlayerA.String())
				} else {
					totalWinner = playerLabel(m.PlayerB.String())
				}
			}

			leg1Turns, leg2Turns, leg3Turns := "", "", ""
			if m.Leg1Turns > 0 {
				leg1Turns = fmt.Sprintf("%d", m.Leg1Turns)
			}
			if m.Leg2Turns > 0 {
				leg2Turns = fmt.Sprintf("%d", m.Leg2Turns)
			}
			if m.Leg3Turns > 0 {
				leg3Turns = fmt.Sprintf("%d", m.Leg3Turns)
			}

			values = []interface{}{
				pA.Nr, domain.FormatDisplayName(pA.Name), "/", pB.Nr, domain.FormatDisplayName(pB.Name),
				playerLabel(m.Leg1Winner), leg1Turns,
				playerLabel(m.Leg2Winner), leg2Turns,
				playerLabel(m.Leg3Winner), leg3Turns,
				totalWinner, eindstand,
				reportedByLabel(m.ReportedBy), m.RescheduleDate, m.SecretaryNr, m.CounterNr,
			}
		} else {
			values = emptyValues
		}

		for j, col := range cols {
			f.SetCellValue(ws, fmt.Sprintf("%s%d", col, row), values[j])
		}
	}

	// Pass B: cell styles / borders (after all heights and values are set).
	for i := 0; i < totalRows; i++ {
		row := 7 + i
		isFirst := i == 0
		isLast := i == totalRows-1

		var styleIDs []int
		switch {
		case isFirst && isLast:
			merged := make([]colSpec, len(row7))
			copy(merged, row7)
			for k := range merged {
				merged[k].b = styleThick
			}
			styleIDs = buildStyles(merged)
		case isFirst:
			styleIDs = row7Styles
		case isLast:
			styleIDs = rowLastStyles
		default:
			styleIDs = rowNStyles
		}

		for j, col := range cols {
			f.SetCellStyle(ws, fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), styleIDs[j])
		}
	}

	// ================================================================ PAGE SETUP
	// Landscape orientation, A4 paper.
	// FitToWidth=1 scales columns to fit one page wide; FitToHeight=0 allows
	// the sheet to span multiple pages vertically when there are many matches.
	orientation := "landscape"
	pageSize := 9 // A4
	fitToWidth := 1
	fitToHeight := 0
	f.SetPageLayout(ws, &excelize.PageLayoutOptions{
		Orientation: &orientation,
		Size:        &pageSize,
		FitToWidth:  &fitToWidth,
		FitToHeight: &fitToHeight,
	})

	// FitToPage is a sheet property separate from the page layout options.
	fitToPage := true
	f.SetSheetProps(ws, &excelize.SheetPropsOptions{FitToPage: &fitToPage})

	// Repeat rows 1–6 as print titles (afdruk titels) on every printed page.
	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Titles",
		RefersTo: ws + "!$1:$6",
		Scope:    ws,
	})

	// Freeze rows 1-6 so the header stays visible when scrolling in Excel.
	f.SetPanes(ws, &excelize.Panes{
		Freeze:      true,
		YSplit:      6,
		TopLeftCell: "A7",
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: "A7", ActiveCell: "A7", Pane: "bottomLeft"},
		},
	})

	// Narrow margins (in inches) — matches Excel's built-in "Narrow" preset.
	left, right, top, bottom, header, footer := 0.25, 0.25, 0.75, 0.75, 0.3, 0.3
	f.SetPageMargins(ws, &excelize.PageLayoutMarginsOptions{
		Left:   &left,
		Right:  &right,
		Top:    &top,
		Bottom: &bottom,
		Header: &header,
		Footer: &footer,
	})

	return f.Write(w)
}

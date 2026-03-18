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
	f.SetRowHeight(ws, 4, 45)

	// ================================================================ 2. COLUMN WIDTHS
	maxNameLen := 10
	for _, p := range players {
		if l := len(domain.FormatDisplayName(p.Name)); l > maxNameLen {
			maxNameLen = l
		}
	}
	nameColWidth := float64(maxNameLen)*1.0 + 1.5 // ≈ autofit: 1 unit/char + padding
	if nameColWidth < 18.0 {
		nameColWidth = 18.0
	}

	for col, width := range map[string]float64{
		"A": 5.5, "B": nameColWidth, "C": 2.5, "D": 5.5, "E": nameColWidth,
		"F": 11.5, "G": 6.5, "H": 11.5, "I": 6.5,
		"J": 11.5, "K": 6.5, "L": 11.5, "M": 6.5,
		"N": 12.5, "O": 7.5, "P": 5.5, "Q": 5.5,
	} {
		f.SetColWidth(ws, col, col, width)
	}

	// ================================================================ 3. MERGES
	f.MergeCell(ws, "A1", "Q1")
	f.MergeCell(ws, "A2", "Q2")
	// No header merges in row 4 — single header row, one cell per column.

	// ================================================================ 4. CELL VALUES
	f.SetCellValue(ws, "A1", "DARTCLUB GROLZICHT")

	dateLabel := ev.Date.Format("2-1-2006")
	if ev.IsCatchUpEvening {
		dateLabel = "Inhaal"
	}
	f.SetCellValue(ws, "A2", fmt.Sprintf(
		"Wedstrijdformulier   Spelsoort: 501 dubbel uit best of 3   Speeldatum: %s",
		dateLabel))

	// Row 4: single header row labels
	f.SetCellValue(ws, "A4", "nr.")
	f.SetCellValue(ws, "B4", "naam")
	f.SetCellValue(ws, "C4", "/")
	f.SetCellValue(ws, "D4", "nr.")
	f.SetCellValue(ws, "E4", "naam")
	f.SetCellValue(ws, "F4", "winnaar\n(naam + nr.)\nleg 1")
	f.SetCellValue(ws, "G4", "aantal\nbeurten")
	f.SetCellValue(ws, "H4", "winnaar\n(naam + nr.)\nleg 2")
	f.SetCellValue(ws, "I4", "aantal\nbeurten")
	f.SetCellValue(ws, "J4", "winnaar\n(naam + nr.)\nleg 3")
	f.SetCellValue(ws, "K4", "aantal\nbeurten")
	f.SetCellValue(ws, "L4", "winnaar\n(naam + nr.)")
	f.SetCellValue(ws, "M4", "Eind-\nstand")
	f.SetCellValue(ws, "N4", "afgemeld\ndoor\n(naam + nr.)")
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

	// Row 4: single header row — thick top + thick bottom on all columns.
	// Border spec per column: L, R, T, B
	f.SetCellStyle(ws, "A4", "A4", hdrStyle(styleThick, styleThin, styleThick, styleThick))
	f.SetCellStyle(ws, "B4", "B4", hdrStyle(styleThin, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "C4", "C4", hdrStyle(styleMedium, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "D4", "D4", hdrStyle(styleMedium, styleThin, styleThick, styleThick))
	f.SetCellStyle(ws, "E4", "E4", hdrStyle(styleThin, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "F4", "F4", hdrStyle(styleMedium, styleThin, styleThick, styleThick))
	f.SetCellStyle(ws, "G4", "G4", hdrStyle(styleThin, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "H4", "H4", hdrStyle(styleMedium, styleThin, styleThick, styleThick))
	f.SetCellStyle(ws, "I4", "I4", hdrStyle(styleThin, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "J4", "J4", hdrStyle(styleMedium, styleThin, styleThick, styleThick))
	f.SetCellStyle(ws, "K4", "K4", hdrStyle(styleThin, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "L4", "L4", hdrStyle(styleMedium, styleThin, styleThick, styleThick))
	f.SetCellStyle(ws, "M4", "M4", hdrStyle(styleThin, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "N4", "N4", hdrStyle(styleMedium, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "O4", "O4", hdrStyle(styleMedium, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "P4", "P4", hdrStyle(styleMedium, styleMedium, styleThick, styleThick))
	f.SetCellStyle(ws, "Q4", "Q4", hdrStyle(styleMedium, styleThick, styleThick, styleThick))

	// ================================================================ DATA ROWS
	// colSpec defines per-column style: font size, horizontal align, shrink-to-fit, and border sides.
	type colSpec struct {
		sz          float64
		ha          string
		shrinkToFit bool
		l, r, t, b  int
	}

	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q"}

	// Row 5 (first data row): thick top on ALL columns.
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

	// Subsequent data rows.
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
	// narrow margins and data rows at 17.25pt each:
	//   header rows 1-4: 21 + 15 + 13.5 + 45 = 94.5pt
	//   A4 landscape printable height ≈ 444pt
	//   data rows: floor((444 - 94.5) / 17.25) = 20
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
		row := 5 + i
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
		row := 5 + i
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

	// Repeat rows 1–4 as print titles (afdruk titels) on every printed page.
	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Titles",
		RefersTo: ws + "!$1:$4",
		Scope:    ws,
	})

	// Freeze rows 1-4 so the header stays visible when scrolling in Excel.
	f.SetPanes(ws, &excelize.Panes{
		Freeze:      true,
		YSplit:      4,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: "A5", ActiveCell: "A5", Pane: "bottomLeft"},
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

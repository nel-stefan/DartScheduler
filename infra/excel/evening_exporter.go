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

	// reportedByLabel converts a stored "nr Achternaam, Voornaam" string to
	// "Voornaam - nr" by looking up the matching player. Falls back to the
	// raw value for free-text entries.
	reportedByLabel := func(raw string) string {
		if raw == "" {
			return ""
		}
		for _, p := range players {
			if p.Nr+" "+p.Name == raw {
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

	// ------------------------------------------------------------------ Row 1
	f.MergeCell(ws, "A1", "Q1")
	f.SetCellValue(ws, "A1", "DARTCLUB GROLZICHT")
	f.SetCellStyle(ws, "A1", "Q1", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 16, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	}))
	f.SetRowHeight(ws, 1, 21)

	// ------------------------------------------------------------------ Row 2
	dateLabel := ev.Date.Format("2-1-2006")
	if ev.IsCatchUpEvening {
		dateLabel = "Inhaal"
	}
	f.MergeCell(ws, "A2", "Q2")
	f.SetCellValue(ws, "A2", fmt.Sprintf(
		"Wedstrijdformulier     Spelsoort: 501 dubbel uit best of 3     Speeldatum: %s",
		dateLabel))
	f.SetCellStyle(ws, "A2", "Q2", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	}))
	f.SetRowHeight(ws, 2, 15)

	// ------------------------------------------------------------------ Row 3 (spacer with thick bottom rule above the header)
	f.SetRowHeight(ws, 3, 13.5)
	row3Style := ns(&excelize.Style{
		Font:   &excelize.Font{Family: fontCalibri, Size: 11},
		Border: brd(0, 0, 0, styleThick),
	})
	for _, cell := range []string{"B3", "D3", "H3", "L3", "M3", "N3"} {
		f.SetCellStyle(ws, cell, cell, row3Style)
	}

	// ------------------------------------------------------------------ Rows 4-6: column headers
	f.SetRowHeight(ws, 4, 15)
	f.SetRowHeight(ws, 5, 13.5)
	f.SetRowHeight(ws, 6, 13.5)

	// Phase 1: merge all header regions before applying any styles.
	// (excelize clears cell styles when merging, so merges must come first.)
	for _, m := range [][2]string{
		{"A4", "A6"}, {"B4", "B6"}, {"C4", "C6"}, {"D4", "D6"}, {"E4", "E6"},
		{"F4", "G4"}, {"F5", "F6"}, {"G5", "G6"},
		{"H4", "I4"}, {"H5", "H6"}, {"I5", "I6"},
		{"J4", "K4"}, {"J5", "J6"}, {"K5", "K6"},
		{"L4", "L6"}, {"M4", "M6"}, {"N4", "N6"}, {"O4", "O6"}, {"P4", "P6"}, {"Q4", "Q6"},
	} {
		f.MergeCell(ws, m[0], m[1])
	}

	// Phase 2: set values, then apply styles to the full merged range so that
	// borders appear correctly on every edge cell of the merged region.

	// Player A: nr and naam
	f.SetCellValue(ws, "A4", "nr")
	f.SetCellStyle(ws, "A4", "A6", hdrStyle(styleThick, styleMedium, styleThick, styleThick))
	f.SetCellValue(ws, "B4", "naam")
	f.SetCellStyle(ws, "B4", "B6", hdrStyle(0, styleThick, styleThick, styleThin))

	// Separator /
	f.SetCellValue(ws, "C4", "/")
	f.SetCellStyle(ws, "C4", "C6", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 10},
		Alignment: hdrCenter,
		Border:    brd(styleThick, styleThick, styleThick, styleThin),
	}))

	// Player B: nr and naam
	f.SetCellValue(ws, "D4", "nr")
	f.SetCellStyle(ws, "D4", "D6", hdrStyle(styleThick, 0, styleThick, styleThick))
	f.SetCellValue(ws, "E4", "naam")
	f.SetCellStyle(ws, "E4", "E6", hdrStyle(styleThick, styleThick, styleThick, styleThin))

	// Partij 1: group header (thick bottom), sub-headers for winner name and turns.
	f.SetCellValue(ws, "F4", "Leg 1")
	f.SetCellStyle(ws, "F4", "G4", hdrStyle(styleThick, styleMedium, styleThick, styleThick))
	f.SetCellValue(ws, "F5", "voornaam+nr.")
	f.SetCellStyle(ws, "F5", "F6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellValue(ws, "G5", "aantal beurten")
	f.SetCellStyle(ws, "G5", "G6", hdrStyle(styleThick, styleThick, styleThick, styleThick))

	// Leg 2: group header (no bottom — sub-headers form the visual separator).
	f.SetCellValue(ws, "H4", "Leg 2")
	f.SetCellStyle(ws, "H4", "I4", hdrStyle(styleThick, styleMedium, styleThick, 0))
	f.SetCellValue(ws, "H5", "voornaam+nr.")
	f.SetCellStyle(ws, "H5", "H6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellValue(ws, "I5", "aantal beurten")
	f.SetCellStyle(ws, "I5", "I6", hdrStyle(styleThick, styleThick, styleThick, styleThick))

	// Leg 3: group header (no left, no bottom).
	f.SetCellValue(ws, "J4", "Leg 3")
	f.SetCellStyle(ws, "J4", "K4", hdrStyle(0, styleMedium, styleThick, 0))
	f.SetCellValue(ws, "J5", "voornaam+nr.")
	f.SetCellStyle(ws, "J5", "J6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellValue(ws, "K5", "aantal beurten")
	f.SetCellStyle(ws, "K5", "K6", hdrStyle(styleThick, styleThick, styleThick, styleThick))

	// Summary and admin columns.
	f.SetCellValue(ws, "L4", "totaal\nwinnaar")
	f.SetCellStyle(ws, "L4", "L6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellValue(ws, "M4", "eind-\nstand")
	f.SetCellStyle(ws, "M4", "M6", hdrStyle(styleThick, styleThick, styleThick, styleThick))
	f.SetCellValue(ws, "N4", "afgemeld\ndoor")
	f.SetCellStyle(ws, "N4", "N6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.SetCellValue(ws, "O4", "vooruit-\ngooi\ndatum")
	f.SetCellStyle(ws, "O4", "O6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.SetCellValue(ws, "P4", "nr.\nschrij-\nver")
	f.SetCellStyle(ws, "P4", "P6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.SetCellValue(ws, "Q4", "nr.\ntel-\nler")
	f.SetCellStyle(ws, "Q4", "Q6", hdrStyle(styleThick, styleThick, styleThick, styleMedium))

	// ------------------------------------------------------------------ Print-title border fix
	// excelize only stores merged-range styles on the top-left cell.  When
	// rows 1-6 are repeated as print titles Excel renders each cell
	// individually, so interior cells (e.g. A5, A6, F6 …) show no borders.
	// Stamp an explicit style on every cell in rows 4-6 so that borders are
	// correct on both the original page and every repeated header page.
	type hdrCellSpec struct{ l, r, t, b int }
	hdrCellSpecs := map[string]hdrCellSpec{
		// Row 4 — only single-row merges (F4:G4, H4:I4, J4:K4).
		// Multi-row merge top-left cells (A4-E4, L4-Q4) are intentionally omitted:
		// they already have the correct bottom=thick style from the SetCellStyle range
		// calls above.  Overriding them here would set bottom=0, which Excel uses when
		// rendering the repeated print-title rows — hiding the thick border on page 2+.
		"F4": {styleThick, 0, styleThick, styleThick},  // F4:G4 left cell — bottom thick
		"G4": {0, styleMedium, styleThick, styleThick},  // F4:G4 right cell
		"H4": {styleThick, 0, styleThick, 0},            // H4:I4 left cell — no bottom
		"I4": {0, styleMedium, styleThick, 0},           // H4:I4 right cell
		"J4": {0, 0, styleThick, 0},                     // J4:K4 left cell
		"K4": {0, styleMedium, styleThick, 0},           // J4:K4 right cell
		// Row 5 — interior rows of tall merges (A-E, L-Q).
		// F5-K5 are top-left cells of F5:F6 … K5:K6 merges — omitted for the same
		// reason as above (their original bottom=thick must not be overridden).
		"A5": {styleThick, styleMedium, 0, 0},
		"B5": {0, styleThick, 0, 0},
		"C5": {styleThick, styleThick, 0, 0},
		"D5": {styleThick, 0, 0, 0},
		"E5": {styleThick, styleThick, 0, 0},
		"L5": {styleThick, styleThick, 0, 0},
		"M5": {styleThick, styleThick, 0, 0},
		"N5": {styleThick, styleThick, 0, 0},
		"O5": {styleThick, styleThick, 0, 0},
		"P5": {styleThick, styleThick, 0, 0},
		"Q5": {styleThick, styleThick, 0, 0},
		// Row 6 — bottom edge of every merge (carries the thick-bottom borders)
		"A6": {styleThick, styleMedium, 0, styleThick},
		"B6": {0, styleThick, 0, styleThin},
		"C6": {styleThick, styleThick, 0, styleThin},
		"D6": {styleThick, 0, 0, styleThick},
		"E6": {styleThick, styleThick, 0, styleThin},
		"F6": {styleThick, styleThick, 0, styleThick},
		"G6": {styleThick, styleThick, 0, styleThick},
		"H6": {styleThick, styleThick, 0, styleThick},
		"I6": {styleThick, styleThick, 0, styleThick},
		"J6": {styleThick, styleThick, 0, styleThick},
		"K6": {styleThick, styleThick, 0, styleThick},
		"L6": {styleThick, styleThick, 0, styleThick},
		"M6": {styleThick, styleThick, 0, styleThick},
		"N6": {styleThick, styleThick, 0, styleMedium},
		"O6": {styleThick, styleThick, 0, styleMedium},
		"P6": {styleThick, styleThick, 0, styleMedium},
		"Q6": {styleThick, styleThick, 0, styleMedium},
	}
	for cell, sp := range hdrCellSpecs {
		f.SetCellStyle(ws, cell, cell, hdrStyle(sp.l, sp.r, sp.t, sp.b))
	}

	// ------------------------------------------------------------------ Column widths
	// Name columns B and E are sized to fit the longest player name.
	maxNameLen := 10
	for _, p := range players {
		if l := len(p.Name); l > maxNameLen {
			maxNameLen = l
		}
	}
	nameColWidth := float64(maxNameLen)*0.9 + 2.0 // approximate character-unit width

	for col, width := range map[string]float64{
		"A": 4.0, "B": nameColWidth, "C": 1.7109, "D": 4.0, "E": nameColWidth,
		"F": 14.4258, "G": 6.5, "H": 14.4258, "I": 6.5,
		"J": 14.4258, "K": 6.5, "L": 13.8555, "M": 5.5703,
		"N": 12.1406, "O": 7.8555, "P": 6.1406, "Q": 6.1406,
	} {
		f.SetColWidth(ws, col, col, width)
	}

	// ------------------------------------------------------------------ Data rows
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q"}

	// colSpec defines per-column style: font size, horizontal align, shrink-to-fit, and border sides.
	type colSpec struct {
		sz          float64
		ha          string
		shrinkToFit bool
		l, r, t, b  int
	}

	// Row 7: first data row — thick top on B/C/E/N–Q to mark the start of data;
	// no top on A/D/F–M (visually bounded by the header bottom borders above).
	row7 := []colSpec{
		{12, "right", false, styleThick, styleThin, 0, styleThin},              // A: nr
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},          // B: naam
		{10, "center", false, styleMedium, styleMedium, styleThick, styleThin},  // C: /
		{12, "right", false, styleMedium, styleThin, 0, styleThin},              // D: nr
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},          // E: naam
		{11, "", true, styleMedium, styleThin, 0, styleThin},                    // F: leg1 winner
		{11, "", false, styleThin, styleMedium, 0, styleThin},                   // G: leg1 turns
		{11, "", true, styleMedium, styleThin, 0, styleThin},                    // H: leg2 winner
		{11, "", false, styleThin, styleMedium, 0, styleThin},                   // I: leg2 turns
		{11, "", true, styleMedium, styleThin, 0, styleThin},                    // J: leg3 winner
		{11, "", false, styleThin, styleMedium, 0, styleThin},                   // K: leg3 turns
		{11, "", true, styleMedium, styleThin, 0, styleThin},                    // L: totaal winnaar
		{11, "center", false, styleThin, styleMedium, 0, styleThin},             // M: eindstand
		{11, "", true, styleMedium, styleMedium, styleThick, styleThin},         // N: afgemeld door
		{11, "", false, styleMedium, styleMedium, styleThick, styleThin},        // O: vooruitgooi datum
		{11, "center", false, styleMedium, styleMedium, styleThick, styleThin},  // P: nr. schrijver
		{11, "center", false, styleMedium, styleThick, styleThick, styleThin},   // Q: nr. teller
	}

	// Rows 8+: subsequent data rows — thin top on B/C/E/N–Q; no top on A/D/F–M.
	rowN := []colSpec{
		{12, "right", false, styleThick, styleThin, 0, styleThin},              // A
		{11, "", false, styleThin, styleMedium, styleThin, styleThin},          // B
		{10, "center", false, styleMedium, styleMedium, styleThin, styleThin},  // C
		{12, "right", false, styleMedium, styleThin, 0, styleThin},             // D
		{11, "", false, styleThin, styleMedium, styleThin, styleThin},          // E
		{11, "", true, styleMedium, styleThin, 0, styleThin},                   // F
		{11, "", false, styleThin, styleMedium, 0, styleThin},                  // G
		{11, "", true, styleMedium, styleThin, 0, styleThin},                   // H
		{11, "", false, styleThin, styleMedium, 0, styleThin},                  // I
		{11, "", true, styleMedium, styleThin, 0, styleThin},                   // J
		{11, "", false, styleThin, styleMedium, 0, styleThin},                  // K
		{11, "", true, styleMedium, styleThin, 0, styleThin},                   // L
		{11, "center", false, styleThin, styleMedium, 0, styleThin},            // M
		{11, "", true, styleMedium, styleMedium, styleThin, styleThin},         // N
		{11, "", false, styleMedium, styleMedium, styleThin, styleThin},        // O
		{11, "center", false, styleMedium, styleMedium, styleThin, styleThin},  // P
		{11, "center", false, styleMedium, styleThick, styleThin, styleThin},   // Q
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

	// A4 landscape printable height minus the header rows leaves room for ~22
	// data rows at natural size; use 33 so there are always ample blank lines
	// for manual entries (FitToPage scales the sheet to fit on one page).
	const minDataRows = 33
	totalRows := len(ev.Matches)
	if totalRows < minDataRows {
		totalRows = minDataRows
	}

	emptyValues := []interface{}{"", "", "/", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}

	for i := 0; i < totalRows; i++ {
		row := 7 + i
		styleIDs := rowNStyles
		if i == 0 {
			styleIDs = row7Styles
		} else if i == totalRows-1 {
			styleIDs = rowLastStyles
		}

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
				pA.Nr, pA.Name, "/", pB.Nr, pB.Name,
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
			cell := fmt.Sprintf("%s%d", col, row)
			f.SetCellValue(ws, cell, values[j])
			f.SetCellStyle(ws, cell, cell, styleIDs[j])
		}
		f.SetRowHeight(ws, row, 17.25)
	}

	// ------------------------------------------------------------------ Page setup
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

	// Repeat rows 1–6 as print titles on every page.
	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Titles",
		RefersTo: ws + "!$1:$6",
		Scope:    ws,
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

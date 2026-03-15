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
	playerLabel := func(id string) string {
		p, ok := playerMap[id]
		if !ok || id == "" {
			return ""
		}
		firstName := strings.SplitN(p.Name, " ", 2)[0]
		return firstName + "+" + p.Nr
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
	f.SetCellValue(ws, "A1", "       ")
	f.SetCellStyle(ws, "A1", "A1", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	}))
	f.SetCellValue(ws, "B1", "   ")
	f.SetCellStyle(ws, "B1", "B1", ns(&excelize.Style{Font: &excelize.Font{Family: fontCalibri, Size: 16}}))
	f.SetCellValue(ws, "C1", "                                       DARTCLUB GROLZICHT")
	f.SetCellStyle(ws, "C1", "C1", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 16, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	}))
	f.SetRowHeight(ws, 1, 21)

	// ------------------------------------------------------------------ Row 2
	f.SetCellValue(ws, "A2", fmt.Sprintf(
		"                                          Wedstrijdformulier        Spelsoort: 501 dubbel uit best of 3       Speeldatum: %s",
		ev.Date.Format("2-1-2006")))
	f.SetCellStyle(ws, "A2", "A2", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "left"},
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

	// mergedHdr merges cells, sets the value and a bold-9pt header style.
	mergedHdr := func(topLeft, bottomRight, value string, l, r, t, b int) {
		f.MergeCell(ws, topLeft, bottomRight)
		f.SetCellValue(ws, topLeft, value)
		f.SetCellStyle(ws, topLeft, topLeft, hdrStyle(l, r, t, b))
	}

	// Player A: nr (A4:A6) and naam (B4:B6)
	mergedHdr("A4", "A6", "nr", styleThick, styleMedium, styleMedium, styleThin)
	mergedHdr("B4", "B6", "naam", 0, styleThick, styleThick, styleThin)

	// Separator /  (C4:C6)
	f.MergeCell(ws, "C4", "C6")
	f.SetCellValue(ws, "C4", "/")
	f.SetCellStyle(ws, "C4", "C4", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 10},
		Alignment: hdrCenter,
		Border:    brd(styleThick, styleThick, styleThick, styleThin),
	}))

	// Player B: nr (D4:D6) and naam (E4:E6)
	mergedHdr("D4", "D6", "nr", styleThick, 0, styleThick, styleThin)
	mergedHdr("E4", "E6", "naam", styleThick, styleThick, styleThick, styleThin)

	// Partij 1: group header F4:G4 (thick bottom to separate from sub-headers),
	// then sub-headers F5:F6 (winner name) and G5:G6 (number of turns).
	mergedHdr("F4", "G4", "Partij 1", styleThick, styleMedium, styleThick, styleThick)
	f.MergeCell(ws, "F5", "F6")
	f.SetCellValue(ws, "F5", "voornaam+nr.")
	f.SetCellStyle(ws, "F5", "F5", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.MergeCell(ws, "G5", "G6")
	f.SetCellValue(ws, "G5", "aantal beurten")
	f.SetCellStyle(ws, "G5", "G5", hdrStyle(styleThick, styleThick, styleThick, styleMedium))

	// Partij 2: group header H4:I4 (no bottom — sub-headers form the visual bottom),
	// then sub-headers H5:H6 and I5:I6.
	mergedHdr("H4", "I4", "Partij 2", styleThick, styleMedium, styleThick, 0)
	f.MergeCell(ws, "H5", "H6")
	f.SetCellValue(ws, "H5", "voornaam+nr.")
	f.SetCellStyle(ws, "H5", "H5", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.MergeCell(ws, "I5", "I6")
	f.SetCellValue(ws, "I5", "aantal beurten")
	f.SetCellStyle(ws, "I5", "I5", hdrStyle(styleThick, styleThick, styleThick, styleMedium))

	// Partij 3: group header J4:K4 (no left, no bottom), then sub-headers J5:J6 and K5:K6.
	mergedHdr("J4", "K4", "Partij 3", 0, styleMedium, styleThick, 0)
	f.MergeCell(ws, "J5", "J6")
	f.SetCellValue(ws, "J5", "voornaam+nr.")
	f.SetCellStyle(ws, "J5", "J5", hdrStyle(styleThick, styleThick, styleThick, styleMedium))
	f.MergeCell(ws, "K5", "K6")
	f.SetCellValue(ws, "K5", "aantal beurten")
	f.SetCellStyle(ws, "K5", "K5", hdrStyle(styleThick, styleThick, styleThick, styleMedium))

	// Summary and admin columns (each spans rows 4-6, medium bottom separates header from data).
	mergedHdr("L4", "L6", "totaal\nwinnaar", styleThick, styleThick, styleThick, styleMedium)
	mergedHdr("M4", "M6", "eind-\nstand", styleThick, styleThick, styleThick, styleMedium)
	mergedHdr("N4", "N6", "afgemeld\ndoor", styleThick, styleThick, styleThick, styleMedium)
	mergedHdr("O4", "O6", "vooruit-\ngooi\ndatum", styleThick, styleThick, styleThick, styleMedium)
	mergedHdr("P4", "P6", "nr.\nschrij-\nver", styleThick, styleThick, styleThick, styleMedium)
	mergedHdr("Q4", "Q6", "nr.\ntel-\nler", styleThick, styleThick, styleThick, styleMedium)

	// ------------------------------------------------------------------ Column widths
	for col, width := range map[string]float64{
		"A": 2.5, "B": 14.0, "C": 1.7109, "D": 2.5, "E": 14.0,
		"F": 14.4258, "G": 8.5, "H": 14.4258, "I": 8.5,
		"J": 14.4258, "K": 8.5, "L": 13.8555, "M": 5.5703,
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
		{11, "", true, styleThin, styleMedium, styleThick, styleThin},           // B: naam
		{10, "center", false, styleMedium, styleMedium, styleThick, styleThin},  // C: /
		{12, "right", false, styleMedium, styleThin, 0, styleThin},              // D: nr
		{11, "", true, styleThin, styleMedium, styleThick, styleThin},           // E: naam
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
		{11, "", true, styleThin, styleMedium, styleThin, styleThin},           // B
		{10, "center", false, styleMedium, styleMedium, styleThin, styleThin},  // C
		{12, "right", false, styleMedium, styleThin, 0, styleThin},             // D
		{11, "", true, styleThin, styleMedium, styleThin, styleThin},           // E
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

	row7Styles := buildStyles(row7)
	rowNStyles := buildStyles(rowN)

	for i, m := range ev.Matches {
		row := 7 + i
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

		values := []interface{}{
			pA.Nr, pA.Name, "/", pB.Nr, pB.Name,
			playerLabel(m.Leg1Winner), leg1Turns,
			playerLabel(m.Leg2Winner), leg2Turns,
			playerLabel(m.Leg3Winner), leg3Turns,
			totalWinner, eindstand,
			m.ReportedBy, m.RescheduleDate, m.SecretaryNr, m.CounterNr,
		}

		styleIDs := row7Styles
		if i > 0 {
			styleIDs = rowNStyles
		}

		for j, col := range cols {
			cell := fmt.Sprintf("%s%d", col, row)
			f.SetCellValue(ws, cell, values[j])
			f.SetCellStyle(ws, cell, cell, styleIDs[j])
		}
		f.SetRowHeight(ws, row, 17.25)
	}

	// ------------------------------------------------------------------ Page setup
	// Landscape orientation, A4 paper, fit all content to one page.
	orientation := "landscape"
	pageSize := 9 // A4
	fitToWidth := 1
	fitToHeight := 1
	f.SetPageLayout(ws, &excelize.PageLayoutOptions{
		Orientation: &orientation,
		Size:        &pageSize,
		FitToWidth:  &fitToWidth,
		FitToHeight: &fitToHeight,
	})

	// FitToPage is a sheet property separate from the page layout options.
	fitToPage := true
	f.SetSheetProps(ws, &excelize.SheetPropsOptions{FitToPage: &fitToPage})

	// Reduced margins (in inches).
	left, right, top, bottom, header, footer := 0.7, 0.7, 0.75, 0.75, 0.3, 0.3
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

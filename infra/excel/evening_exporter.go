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
// wedstrijdformulier format.
//
// For a regular evening the workbook contains:
//   - Tab 1 "Blad1": the evening's own matches
//   - Tab 2+ "Inhaal <datum>": one tab per catch-up evening in sched.Evenings (if any)
//
// Build order per sheet: heights → widths → merges → values → styles.
func ExportEvening(_ context.Context, sched domain.Schedule, ev domain.Evening, players []domain.Player, w io.Writer) error {
	// ---- helpers ----
	playerMap := make(map[string]domain.Player, len(players))
	for _, p := range players {
		playerMap[p.ID.String()] = p
	}

	firstNameNr := func(p domain.Player) string {
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

	// ---- workbook ----
	f := excelize.NewFile()
	defer f.Close()

	// Primary sheet: the requested evening.
	ws := "Blad1"
	f.SetSheetName("Sheet1", ws)
	if err := writeEveningSheet(f, ws, ev, players, playerMap, firstNameNr, playerLabel, reportedByLabel); err != nil {
		return err
	}

	// Extra tab for each catch-up evening supplied via sched.Evenings.
	for _, inhaalEv := range sched.Evenings {
		if !inhaalEv.IsCatchUpEvening {
			continue
		}
		wsTab := fmt.Sprintf("Inhaal %s", inhaalEv.Date.Format("2-1-2006"))
		f.NewSheet(wsTab)
		if err := writeEveningSheet(f, wsTab, inhaalEv, players, playerMap, firstNameNr, playerLabel, reportedByLabel); err != nil {
			return err
		}
	}

	return f.Write(w)
}

// writeEveningSheet renders one wedstrijdformulier sheet into workbook f.
// Build order: heights → widths → merges → values → styles.
func writeEveningSheet(
	f *excelize.File,
	ws string,
	ev domain.Evening,
	players []domain.Player,
	playerMap map[string]domain.Player,
	firstNameNr func(domain.Player) string,
	playerLabel func(string) string,
	reportedByLabel func(string) string,
) error {
	// ---- style helpers ----
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
	for col, width := range map[string]float64{
		"A": 3.5, "B": 18.0, "C": 2.0, "D": 3.5, "E": 18.0,
		"F": 14.0, "G": 6.5, "H": 14.0, "I": 6.5,
		"J": 14.0, "K": 6.5, "L": 14.0, "M": 6.5,
		"N": 12.5, "O": 7.5, "P": 5.5, "Q": 5.5,
	} {
		f.SetColWidth(ws, col, col, width)
	}

	// ================================================================ 3. MERGES
	f.MergeCell(ws, "A1", "Q1")
	f.MergeCell(ws, "A2", "Q2")

	// ================================================================ 4. CELL VALUES
	f.SetCellValue(ws, "A1", "DARTCLUB GROLZICHT")

	dateLabel := ev.Date.Format("2-1-2006")
	if ev.IsCatchUpEvening {
		dateLabel = "Inhaal"
	}
	f.SetCellValue(ws, "A2", fmt.Sprintf(
		"Wedstrijdformulier   Spelsoort: 501 dubbel uit best of 3   Speeldatum: %s", dateLabel))

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
	f.SetCellStyle(ws, "A1", "Q1", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 16, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	}))
	f.SetCellStyle(ws, "A2", "Q2", ns(&excelize.Style{
		Font:      &excelize.Font{Family: fontCalibri, Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	}))
	row3Style := ns(&excelize.Style{
		Font:   &excelize.Font{Family: fontCalibri, Size: 11},
		Border: brd(0, 0, 0, styleThick),
	})
	for _, cell := range []string{"B3", "D3", "H3", "L3", "M3", "N3"} {
		f.SetCellStyle(ws, cell, cell, row3Style)
	}
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
	type colSpec struct {
		sz          float64
		ha          string
		shrinkToFit bool
		l, r, t, b  int
	}
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q"}

	row7 := []colSpec{ // first data row: thick top
		{12, "right", false, styleThick, styleThin, styleThick, styleThin},
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},
		{10, "center", false, styleMedium, styleMedium, styleThick, styleThin},
		{12, "right", false, styleMedium, styleThin, styleThick, styleThin},
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},
		{11, "", false, styleThin, styleMedium, styleThick, styleThin},
		{11, "", true, styleMedium, styleThin, styleThick, styleThin},
		{11, "center", false, styleThin, styleMedium, styleThick, styleThin},
		{11, "", true, styleMedium, styleMedium, styleThick, styleThin},
		{11, "", false, styleMedium, styleMedium, styleThick, styleThin},
		{11, "center", false, styleMedium, styleMedium, styleThick, styleThin},
		{11, "center", false, styleMedium, styleThick, styleThick, styleThin},
	}
	rowN := []colSpec{ // middle data rows
		{12, "right", false, styleThick, styleThin, 0, styleThin},
		{11, "", false, styleThin, styleMedium, styleThin, styleThin},
		{10, "center", false, styleMedium, styleMedium, styleThin, styleThin},
		{12, "right", false, styleMedium, styleThin, 0, styleThin},
		{11, "", false, styleThin, styleMedium, styleThin, styleThin},
		{11, "", true, styleMedium, styleThin, 0, styleThin},
		{11, "", false, styleThin, styleMedium, 0, styleThin},
		{11, "", true, styleMedium, styleThin, 0, styleThin},
		{11, "", false, styleThin, styleMedium, 0, styleThin},
		{11, "", true, styleMedium, styleThin, 0, styleThin},
		{11, "", false, styleThin, styleMedium, 0, styleThin},
		{11, "", true, styleMedium, styleThin, 0, styleThin},
		{11, "center", false, styleThin, styleMedium, 0, styleThin},
		{11, "", true, styleMedium, styleMedium, styleThin, styleThin},
		{11, "", false, styleMedium, styleMedium, styleThin, styleThin},
		{11, "center", false, styleMedium, styleMedium, styleThin, styleThin},
		{11, "center", false, styleMedium, styleThick, styleThin, styleThin},
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

	rowLast := make([]colSpec, len(rowN))
	copy(rowLast, rowN)
	for i := range rowLast {
		rowLast[i].b = styleThick
	}

	row7Styles := buildStyles(row7)
	rowNStyles := buildStyles(rowN)
	rowLastStyles := buildStyles(rowLast)

	const rowsPerPage = 25
	matchCount := len(ev.Matches)
	pages := (matchCount + rowsPerPage - 1) / rowsPerPage
	if pages < 1 {
		pages = 1
	}
	totalRows := pages * rowsPerPage

	emptyValues := []interface{}{"", "", "/", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}

	// Pass A: heights + values.
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

	// Pass B: styles / borders.
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
	fitToPage := true
	f.SetSheetProps(ws, &excelize.SheetPropsOptions{FitToPage: &fitToPage})
	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Titles",
		RefersTo: ws + "!$1:$4",
		Scope:    ws,
	})
	f.SetPanes(ws, &excelize.Panes{
		Freeze:      true,
		YSplit:      4,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
		Selection:   []excelize.Selection{{SQRef: "A5", ActiveCell: "A5", Pane: "bottomLeft"}},
	})
	left, right, top, bottom, header, footer := 0.25, 0.25, 0.75, 0.75, 0.3, 0.3
	f.SetPageMargins(ws, &excelize.PageLayoutMarginsOptions{
		Left: &left, Right: &right, Top: &top, Bottom: &bottom,
		Header: &header, Footer: &footer,
	})
	return nil
}

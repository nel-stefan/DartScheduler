package pdf

import (
	"context"
	"fmt"
	"io"
	"strings"

	"DartScheduler/domain"

	"github.com/jung-kurt/gofpdf"
)

// EveningExporter implements usecase.EveningExporter for the wedstrijdformulier
// PDF format.  The output is a 1-on-1 copy of the Excel evening export:
// A4 landscape, 25 data rows per page, same column widths and header layout.
type EveningExporter struct {
	ClubName string
	LogoPath string
}

func (e EveningExporter) ExportEvening(_ context.Context, sched domain.Schedule, ev domain.Evening, players []domain.Player, w io.Writer) error {
	return ExportEvening(sched, ev, players, e.ClubName, e.LogoPath, w)
}

// ─── column widths in mm (A4 landscape, ~284 mm usable) ───

// Excel column widths (char-units): A 3.5 · B 18 · C 2 · D 3.5 · E 18 ·
// F 14 · G 6.5 · H 14 · I 6.5 · J 14 · K 6.5 · L 14 · M 6.5 · N 12.5 ·
// O 7.5 · P 5.5 · Q 5.5  →  total 157.5.
// Scaled to 284 mm usable width (6.35 mm margins each side on 297 mm).

var colW = [17]float64{
	6.3, 32.5, 3.6, 6.3, 32.5, // A-E
	25.3, 11.7, 25.3, 11.7, // F-I
	25.3, 11.7, 25.3, 11.7, // J-M
	22.5, 13.5, 9.9, 9.9, // N-Q
}

const (
	rowsPerPage  = 25
	marginLR     = 6.35 // left/right margin mm
	marginTop    = 8.0
	marginBottom = 6.0
	dataRowH     = 5.8  // data row height mm
	hdrRowH      = 14.0 // column header row height mm
	titleH       = 8.0
	subtitleH    = 6.0
	gapH         = 2.0
	fontFamily   = "Arial"
	borderThin   = "0.25"
	borderMedium = "0.5"
	borderThick  = "0.75"
)

// ExportEvening writes the wedstrijdformulier for one evening to w as PDF.
func ExportEvening(sched domain.Schedule, ev domain.Evening, players []domain.Player, clubName, logoPath string, w io.Writer) error {
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
			if p.Nr != "" && (strings.HasPrefix(raw, p.Nr+" ") || strings.HasPrefix(raw, p.Nr+"\t")) {
				return firstNameNr(p)
			}
		}
		return raw
	}

	dateLabel := ev.Date.Format("2-1-2006")
	if ev.IsCatchUpEvening {
		dateLabel = "Inhaal"
	}

	// Pad match count to full pages of rowsPerPage.
	matchCount := len(ev.Matches)
	pages := (matchCount + rowsPerPage - 1) / rowsPerPage
	if pages < 1 {
		pages = 1
	}
	totalRows := pages * rowsPerPage

	// Build row data.
	type rowData struct {
		nrA, nameA              string
		nrB, nameB              string
		leg1Winner, leg1Turns   string
		leg2Winner, leg2Turns   string
		leg3Winner, leg3Turns   string
		totalWinner, eindstand  string
		reportedBy, reschedDate string
		secretaryNr, counterNr  string
	}

	rows := make([]rowData, totalRows)
	for i := range rows {
		if i >= matchCount {
			break
		}
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
		turns := func(n int) string {
			if n > 0 {
				return fmt.Sprintf("%d", n)
			}
			return ""
		}
		rows[i] = rowData{
			nrA: pA.Nr, nameA: domain.FormatDisplayName(pA.Name),
			nrB: pB.Nr, nameB: domain.FormatDisplayName(pB.Name),
			leg1Winner: playerLabel(m.Leg1Winner), leg1Turns: turns(m.Leg1Turns),
			leg2Winner: playerLabel(m.Leg2Winner), leg2Turns: turns(m.Leg2Turns),
			leg3Winner: playerLabel(m.Leg3Winner), leg3Turns: turns(m.Leg3Turns),
			totalWinner: totalWinner, eindstand: eindstand,
			reportedBy: reportedByLabel(m.ReportedBy), reschedDate: m.RescheduleDate,
			secretaryNr: m.SecretaryNr, counterNr: m.CounterNr,
		}
	}

	// ── PDF setup ──
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(marginLR, marginTop, marginLR)
	pdf.SetAutoPageBreak(false, marginBottom)

	tableW := totalWidth()

	// ── draw helpers ──

	// drawTitle renders the two header lines: club name + subtitle.
	drawTitle := func() {
		if logoPath != "" {
			imgOpts := gofpdf.ImageOptions{ImageType: "", ReadDpi: true}
			pdf.ImageOptions(logoPath, pdf.GetX(), pdf.GetY(), 0, titleH, false, imgOpts, 0, "")
			pdf.Ln(titleH)
		}
		pdf.SetFont(fontFamily, "B", 16)
		pdf.CellFormat(tableW, titleH, clubName, "", 1, "C", false, 0, "")
		pdf.SetFont(fontFamily, "B", 9)
		sub := fmt.Sprintf("Wedstrijdformulier   Spelsoort: 501 dubbel uit best of 3   Speeldatum: %s", dateLabel)
		pdf.CellFormat(tableW, subtitleH, sub, "", 1, "C", false, 0, "")
		pdf.Ln(gapH)
	}

	// drawColumnHeaders renders the header row with column names.
	drawColumnHeaders := func() {
		pdf.SetFont(fontFamily, "B", 7.5)
		headers := [17]string{
			"nr.", "naam", "/", "nr.", "naam",
			"winnaar\n(naam + nr.)\nleg 1", "aantal\nbeurten",
			"winnaar\n(naam + nr.)\nleg 2", "aantal\nbeurten",
			"winnaar\n(naam + nr.)\nleg 3", "aantal\nbeurten",
			"winnaar\n(naam + nr.)", "Eind-\nstand",
			"afgemeld\ndoor\n(naam + nr.)", "vooruit-\ngooi\ndatum",
			"nr.\nschr.", "nr.\ntel.",
		}
		x0 := pdf.GetX()
		y0 := pdf.GetY()
		for i, h := range headers {
			x := x0 + colOffset(i)
			pdf.Rect(x, y0, colW[i], hdrRowH, "D")
			pdf.SetXY(x, y0)
			pdf.MultiCell(colW[i], hdrRowH/float64(lineCount(h)), h, "", "C", false)
		}
		pdf.SetXY(x0, y0+hdrRowH)
	}

	// drawDataRow renders one data row at the current Y position.
	drawDataRow := func(r rowData, isFirst, isLast bool) {
		x0 := pdf.GetX()
		y0 := pdf.GetY()
		cells := [17]string{
			r.nrA, r.nameA, "/", r.nrB, r.nameB,
			r.leg1Winner, r.leg1Turns,
			r.leg2Winner, r.leg2Turns,
			r.leg3Winner, r.leg3Turns,
			r.totalWinner, r.eindstand,
			r.reportedBy, r.reschedDate,
			r.secretaryNr, r.counterNr,
		}
		aligns := [17]string{
			"R", "L", "C", "R", "L",
			"L", "C", "L", "C",
			"L", "C", "L", "C",
			"L", "L", "C", "C",
		}

		// Use thin borders for all cell edges; draw thick outer border around
		// the whole table separately.
		pdf.SetFont(fontFamily, "", 8)
		pdf.SetLineWidth(0.15)
		for i := range cells {
			x := x0 + colOffset(i)
			pdf.Rect(x, y0, colW[i], dataRowH, "D")

			// Clip text to cell width (shrink-to-fit simulation).
			pdf.SetXY(x+0.5, y0+0.3)
			pdf.CellFormat(colW[i]-1.0, dataRowH-0.6, cells[i], "", 0, aligns[i], false, 0, "")
		}

		// Heavy outer table borders: thick left, right edges; thick top on
		// first row of each page; thick bottom on last row of each page.
		pdf.SetLineWidth(0.55)
		// Left edge
		pdf.Line(x0, y0, x0, y0+dataRowH)
		// Right edge
		pdf.Line(x0+tableW, y0, x0+tableW, y0+dataRowH)
		if isFirst {
			pdf.Line(x0, y0, x0+tableW, y0)
		}
		if isLast {
			pdf.Line(x0, y0+dataRowH, x0+tableW, y0+dataRowH)
		}

		// Medium column-group separators (between the main column groups).
		pdf.SetLineWidth(0.35)
		groupEnds := []int{1, 4, 6, 8, 10, 12, 13, 14, 15} // after col B, E, G, I, K, M, N, O, P
		for _, gi := range groupEnds {
			xSep := x0 + colOffset(gi+1)
			pdf.Line(xSep, y0, xSep, y0+dataRowH)
		}

		pdf.SetLineWidth(0.15)
		pdf.SetXY(x0, y0+dataRowH)
	}

	// ── render pages ──
	rowIdx := 0
	for p := 0; p < pages; p++ {
		pdf.AddPage()
		drawTitle()

		// Thick border around the header row.
		hdrX, hdrY := pdf.GetX(), pdf.GetY()
		drawColumnHeaders()
		pdf.SetLineWidth(0.55)
		pdf.Rect(hdrX, hdrY, tableW, hdrRowH, "D")
		pdf.SetLineWidth(0.15)

		for r := 0; r < rowsPerPage && rowIdx < totalRows; r++ {
			drawDataRow(rows[rowIdx], r == 0, r == rowsPerPage-1 || rowIdx == totalRows-1)
			rowIdx++
		}
	}

	// ── "Afgemeld" extra pages (cancelled matches) ──
	for _, inhaalEv := range sched.Evenings {
		if !inhaalEv.IsCatchUpEvening || len(inhaalEv.Matches) == 0 {
			continue
		}
		// Recursively export the catch-up evening on additional pages.
		// We build a temporary sched without extra evenings to avoid recursion.
		emptyExport(pdf, inhaalEv, players, playerMap, firstNameNr, playerLabel, reportedByLabel, clubName, logoPath)
	}

	return pdf.Output(w)
}

// emptyExport renders additional pages for a catch-up evening into an existing
// PDF document (used for the "Afgemeld" tab equivalent).
func emptyExport(
	pdf *gofpdf.Fpdf,
	ev domain.Evening,
	players []domain.Player,
	playerMap map[string]domain.Player,
	firstNameNr func(domain.Player) string,
	playerLabel func(string) string,
	reportedByLabel func(string) string,
	clubName, logoPath string,
) {
	matchCount := len(ev.Matches)
	pages := (matchCount + rowsPerPage - 1) / rowsPerPage
	if pages < 1 {
		pages = 1
	}
	totalRows := pages * rowsPerPage
	tableW := totalWidth()

	type rowData struct {
		nrA, nameA              string
		nrB, nameB              string
		leg1Winner, leg1Turns   string
		leg2Winner, leg2Turns   string
		leg3Winner, leg3Turns   string
		totalWinner, eindstand  string
		reportedBy, reschedDate string
		secretaryNr, counterNr  string
	}

	rows := make([]rowData, totalRows)
	for i := range rows {
		if i >= matchCount {
			break
		}
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
		turns := func(n int) string {
			if n > 0 {
				return fmt.Sprintf("%d", n)
			}
			return ""
		}
		rows[i] = rowData{
			nrA: pA.Nr, nameA: domain.FormatDisplayName(pA.Name),
			nrB: pB.Nr, nameB: domain.FormatDisplayName(pB.Name),
			leg1Winner: playerLabel(m.Leg1Winner), leg1Turns: turns(m.Leg1Turns),
			leg2Winner: playerLabel(m.Leg2Winner), leg2Turns: turns(m.Leg2Turns),
			leg3Winner: playerLabel(m.Leg3Winner), leg3Turns: turns(m.Leg3Turns),
			totalWinner: totalWinner, eindstand: eindstand,
			reportedBy: reportedByLabel(m.ReportedBy), reschedDate: m.RescheduleDate,
			secretaryNr: m.SecretaryNr, counterNr: m.CounterNr,
		}
	}

	drawTitle := func() {
		if logoPath != "" {
			imgOpts := gofpdf.ImageOptions{ImageType: "", ReadDpi: true}
			pdf.ImageOptions(logoPath, pdf.GetX(), pdf.GetY(), 0, titleH, false, imgOpts, 0, "")
			pdf.Ln(titleH)
		}
		pdf.SetFont(fontFamily, "B", 16)
		pdf.CellFormat(tableW, titleH, clubName, "", 1, "C", false, 0, "")
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(tableW, subtitleH,
			"Wedstrijdformulier   Spelsoort: 501 dubbel uit best of 3   Speeldatum: Inhaal",
			"", 1, "C", false, 0, "")
		pdf.Ln(gapH)
	}

	drawColumnHeaders := func() {
		pdf.SetFont(fontFamily, "B", 7.5)
		headers := [17]string{
			"nr.", "naam", "/", "nr.", "naam",
			"winnaar\n(naam + nr.)\nleg 1", "aantal\nbeurten",
			"winnaar\n(naam + nr.)\nleg 2", "aantal\nbeurten",
			"winnaar\n(naam + nr.)\nleg 3", "aantal\nbeurten",
			"winnaar\n(naam + nr.)", "Eind-\nstand",
			"afgemeld\ndoor\n(naam + nr.)", "vooruit-\ngooi\ndatum",
			"nr.\nschr.", "nr.\ntel.",
		}
		x0 := pdf.GetX()
		y0 := pdf.GetY()
		for i, h := range headers {
			x := x0 + colOffset(i)
			pdf.Rect(x, y0, colW[i], hdrRowH, "D")
			pdf.SetXY(x, y0)
			pdf.MultiCell(colW[i], hdrRowH/float64(lineCount(h)), h, "", "C", false)
		}
		pdf.SetXY(x0, y0+hdrRowH)
	}

	drawDataRow := func(r rowData, isFirst, isLast bool) {
		x0 := pdf.GetX()
		y0 := pdf.GetY()
		cells := [17]string{
			r.nrA, r.nameA, "/", r.nrB, r.nameB,
			r.leg1Winner, r.leg1Turns,
			r.leg2Winner, r.leg2Turns,
			r.leg3Winner, r.leg3Turns,
			r.totalWinner, r.eindstand,
			r.reportedBy, r.reschedDate,
			r.secretaryNr, r.counterNr,
		}
		aligns := [17]string{
			"R", "L", "C", "R", "L",
			"L", "C", "L", "C",
			"L", "C", "L", "C",
			"L", "L", "C", "C",
		}
		pdf.SetFont(fontFamily, "", 8)
		pdf.SetLineWidth(0.15)
		for i := range cells {
			x := x0 + colOffset(i)
			pdf.Rect(x, y0, colW[i], dataRowH, "D")
			pdf.SetXY(x+0.5, y0+0.3)
			pdf.CellFormat(colW[i]-1.0, dataRowH-0.6, cells[i], "", 0, aligns[i], false, 0, "")
		}
		pdf.SetLineWidth(0.55)
		pdf.Line(x0, y0, x0, y0+dataRowH)
		pdf.Line(x0+tableW, y0, x0+tableW, y0+dataRowH)
		if isFirst {
			pdf.Line(x0, y0, x0+tableW, y0)
		}
		if isLast {
			pdf.Line(x0, y0+dataRowH, x0+tableW, y0+dataRowH)
		}
		pdf.SetLineWidth(0.35)
		groupEnds := []int{1, 4, 6, 8, 10, 12, 13, 14, 15}
		for _, gi := range groupEnds {
			xSep := x0 + colOffset(gi+1)
			pdf.Line(xSep, y0, xSep, y0+dataRowH)
		}
		pdf.SetLineWidth(0.15)
		pdf.SetXY(x0, y0+dataRowH)
	}

	rowIdx := 0
	for p := 0; p < pages; p++ {
		pdf.AddPage()
		drawTitle()

		hdrX, hdrY := pdf.GetX(), pdf.GetY()
		drawColumnHeaders()
		pdf.SetLineWidth(0.55)
		pdf.Rect(hdrX, hdrY, tableW, hdrRowH, "D")
		pdf.SetLineWidth(0.15)

		for r := 0; r < rowsPerPage && rowIdx < totalRows; r++ {
			drawDataRow(rows[rowIdx], r == 0, r == rowsPerPage-1 || rowIdx == totalRows-1)
			rowIdx++
		}
	}
}

// ── helpers ──

func totalWidth() float64 {
	var w float64
	for _, c := range colW {
		w += c
	}
	return w
}

func colOffset(i int) float64 {
	var x float64
	for j := 0; j < i; j++ {
		x += colW[j]
	}
	return x
}

func lineCount(s string) int {
	n := strings.Count(s, "\n") + 1
	if n < 1 {
		return 1
	}
	return n
}

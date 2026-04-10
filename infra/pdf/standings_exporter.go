package pdf

import (
	"fmt"
	"io"
	"sort"

	"DartScheduler/usecase"

	"github.com/jung-kurt/gofpdf"
)

type standingsClassGroup struct {
	label string
	key   string
	stats []usecase.PlayerStats
}

// ExportStandings writes a klassement PDF (landscape A4) to w.
// Klasse 1 and Klasse 2 are rendered side-by-side on the first page;
// any additional classes each get their own page. Duty stats follow on a final page.
func ExportStandings(competitionName, season string, stats []usecase.PlayerStats, dutyStats []usecase.DutyStats, w io.Writer) error {
	f := gofpdf.New("L", "mm", "A4", "")
	f.SetMargins(14, 14, 14)
	f.SetAutoPageBreak(false, 0)
	tr := f.UnicodeTranslatorFromDescriptor("")

	// A4 landscape usable area: 297 - 28 = 269mm wide, 210 - 28 = 182mm tall
	const (
		pageW       = 297.0
		marginL     = 14.0
		usableW     = pageW - 2*marginL // 269mm
		twoColW     = (usableW - 10) / 2 // each column ~129.5mm (gap = 10mm)
		twoColGap   = 10.0
		col2X       = marginL + twoColW + twoColGap
		rowH        = 6.5
		headerRowH  = 7.5
	)

	// Table column widths (must sum to twoColW ≈ 129.5)
	cw := []float64{8, 13, 65, 17, 17, 10} // 130mm
	ch := []string{"#", "Nr", "Naam", "+ pnt", "- pnt", "180"}
	ca := []string{"C", "C", "L", "C", "C", "C"}

	// Colors
	setHeaderFill := func() { f.SetFillColor(52, 73, 94) }  // dark blue-grey
	setAltFill := func() { f.SetFillColor(236, 240, 241) }   // very light grey
	setClearFill := func() { f.SetFillColor(255, 255, 255) }

	// Draw full-page header (title + subtitle). Returns Y after header.
	drawPageHeader := func(subtitle string) float64 {
		f.AddPage()
		startY := marginL

		// Title bar background
		f.SetFillColor(52, 73, 94)
		f.SetXY(marginL, startY)
		f.CellFormat(usableW, 11, "", "", 0, "L", true, 0, "")

		// Title text
		f.SetFont("Arial", "B", 16)
		f.SetTextColor(255, 255, 255)
		f.SetXY(marginL, startY)
		f.CellFormat(usableW, 11, tr(competitionName), "", 0, "L", false, 0, "")

		// Subtitle / season (right-aligned, italic)
		f.SetFont("Arial", "I", 11)
		f.SetXY(marginL, startY)
		f.CellFormat(usableW, 11, tr(season), "", 1, "R", false, 0, "")

		// Section label below header bar
		f.SetTextColor(80, 80, 80)
		f.SetFont("Arial", "I", 9)
		f.SetXY(marginL, startY+12)
		f.CellFormat(usableW, 6, tr(subtitle), "", 1, "L", false, 0, "")

		f.SetTextColor(0, 0, 0)
		return startY + 12 + 6 + 3 // Y below header + small gap
	}

	// Draw one class table starting at (x, y). Returns Y after last row.
	drawClassTable := func(cg *standingsClassGroup, x, y float64) float64 {
		// Class label
		f.SetFont("Arial", "B", 12)
		f.SetTextColor(52, 73, 94)
		f.SetXY(x, y)
		f.CellFormat(twoColW, 8, tr(cg.label), "", 0, "L", false, 0, "")
		// Underline by drawing a line
		f.SetDrawColor(52, 73, 94)
		f.SetLineWidth(0.4)
		f.Line(x, y+8, x+twoColW, y+8)
		y += 9

		// Column headers
		setHeaderFill()
		f.SetFont("Arial", "B", 9)
		f.SetTextColor(255, 255, 255)
		f.SetDrawColor(200, 200, 200)
		f.SetLineWidth(0.2)
		xc := x
		for i, h := range ch {
			f.SetXY(xc, y)
			f.CellFormat(cw[i], headerRowH, h, "1", 0, ca[i], true, 0, "")
			xc += cw[i]
		}
		y += headerRowH

		// Data rows
		f.SetTextColor(0, 0, 0)
		for rank, s := range cg.stats {
			alt := rank%2 == 1
			if alt {
				setAltFill()
			} else {
				setClearFill()
			}
			// Bold + slightly larger for rank 1
			if rank == 0 {
				f.SetFont("Arial", "B", 9)
			} else {
				f.SetFont("Arial", "", 9)
			}

			e180 := ""
			if s.OneEighties > 0 {
				e180 = fmt.Sprintf("%d", s.OneEighties)
			}
			cells := []string{
				fmt.Sprintf("%d", rank+1),
				tr(s.Player.Nr),
				tr(s.Player.Name),
				fmt.Sprintf("%d", s.PointsFor),
				fmt.Sprintf("%d", s.PointsAgainst),
				e180,
			}
			xc := x
			for i, cell := range cells {
				f.SetXY(xc, y)
				border := "1"
				f.CellFormat(cw[i], rowH, cell, border, 0, ca[i], true, 0, "")
				xc += cw[i]
			}
			y += rowH
		}
		f.SetDrawColor(0, 0, 0)
		f.SetLineWidth(0.2)
		return y
	}

	// --- Group and sort stats by class ---
	classOrder := []string{}
	classMap := map[string]*standingsClassGroup{}
	for _, s := range stats {
		cls := s.Player.Class
		label := "Klasse " + cls
		if cls == "" {
			label = "Overig"
			cls = "~" // sort last
		}
		if _, ok := classMap[cls]; !ok {
			classOrder = append(classOrder, cls)
			classMap[cls] = &standingsClassGroup{label: label, key: cls}
		}
		classMap[cls].stats = append(classMap[cls].stats, s)
	}
	sort.Strings(classOrder) // "1" < "2" < "~"
	for _, cg := range classMap {
		sort.Slice(cg.stats, func(i, j int) bool {
			a, b := cg.stats[i], cg.stats[j]
			if a.PointsFor != b.PointsFor {
				return a.PointsFor > b.PointsFor
			}
			return (a.PointsFor - a.PointsAgainst) > (b.PointsFor - b.PointsAgainst)
		})
	}

	// --- Render pages ---
	// First two classes side-by-side on one landscape page
	for i := 0; i < len(classOrder); i += 2 {
		headerY := drawPageHeader("Klassement")

		leftCls := classMap[classOrder[i]]
		drawClassTable(leftCls, marginL, headerY)

		if i+1 < len(classOrder) {
			rightCls := classMap[classOrder[i+1]]
			drawClassTable(rightCls, col2X, headerY)
		}
	}

	// --- Duty stats page ---
	// Wider single table for duty stats
	dutyW := []float64{9, 13, 115, 19} // 156mm — fits nicely on half-width landscape
	dutyH := []string{"#", "Nr", "Naam", "Keer"}
	dutyA := []string{"C", "C", "L", "C"}
	dutyTableW := 0.0
	for _, w := range dutyW {
		dutyTableW += w
	}

	headerY := drawPageHeader("Schrijver / Teller")

	// Subtitle note
	f.SetFont("Arial", "I", 9)
	f.SetTextColor(100, 100, 100)
	f.SetXY(marginL, headerY)
	f.CellFormat(dutyTableW, 5, "Totaal aantal keer als schrijver of teller ingezet (gecombineerd).", "", 1, "L", false, 0, "")
	headerY += 7

	// Header row
	setHeaderFill()
	f.SetFont("Arial", "B", 9)
	f.SetTextColor(255, 255, 255)
	xc := marginL
	for i, h := range dutyH {
		f.SetXY(xc, headerY)
		f.CellFormat(dutyW[i], headerRowH, h, "1", 0, dutyA[i], true, 0, "")
		xc += dutyW[i]
	}
	headerY += headerRowH

	f.SetTextColor(0, 0, 0)
	rank := 1
	for ri, d := range dutyStats {
		if d.Count == 0 {
			continue
		}
		alt := ri%2 == 1
		if alt {
			setAltFill()
		} else {
			setClearFill()
		}
		if rank == 1 {
			f.SetFont("Arial", "B", 9)
		} else {
			f.SetFont("Arial", "", 9)
		}
		cells := []string{
			fmt.Sprintf("%d", rank),
			tr(d.Player.Nr),
			tr(d.Player.Name),
			fmt.Sprintf("%d", d.Count),
		}
		xc := marginL
		for i, cell := range cells {
			f.SetXY(xc, headerY)
			f.CellFormat(dutyW[i], rowH, cell, "1", 0, dutyA[i], true, 0, "")
			xc += dutyW[i]
		}
		headerY += rowH
		rank++
	}

	return f.Output(w)
}

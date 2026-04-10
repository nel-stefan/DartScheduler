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
// any additional classes each get their own page. Duty stats follow on a portrait page.
func ExportStandings(stats []usecase.PlayerStats, dutyStats []usecase.DutyStats, w io.Writer) error {
	f := gofpdf.New("L", "mm", "A4", "")
	f.SetMargins(14, 14, 14)
	f.SetAutoPageBreak(false, 0)

	// Register Verdana (embedded TTF) — no UnicodeTranslator needed for UTF-8 fonts
	f.AddUTF8FontFromBytes("Verdana", "", verdanaRegular)
	f.AddUTF8FontFromBytes("Verdana", "B", verdanaBold)
	f.AddUTF8FontFromBytes("Verdana", "I", verdanaItalic)
	f.AddUTF8FontFromBytes("Verdana", "BI", verdanaBoldItalic)

	// A4 landscape usable area: 297 - 28 = 269mm wide, 210 - 28 = 182mm tall
	const (
		pageW      = 297.0
		marginL    = 14.0
		usableW    = pageW - 2*marginL  // 269mm
		twoColW    = (usableW - 10) / 2 // each column ~129.5mm (gap = 10mm)
		twoColGap  = 10.0
		col2X      = marginL + twoColW + twoColGap
		rowH       = 7.0
		headerRowH = 8.0
	)

	// Table column widths (must sum to twoColW ≈ 129.5)
	cw := []float64{8, 13, 65, 17, 17, 10} // 130mm
	ch := []string{"#", "Nr", "Naam", "+ pnt", "- pnt", "180"}
	ca := []string{"C", "C", "L", "C", "C", "C"}

	// Colors
	setHeaderFill := func() { f.SetFillColor(52, 73, 94) } // dark blue-grey
	setAltFill := func() { f.SetFillColor(236, 240, 241) } // very light grey
	setClearFill := func() { f.SetFillColor(255, 255, 255) }

	// drawClassTable draws one standings table at (x, y). Returns Y after last row.
	drawClassTable := func(cg *standingsClassGroup, x, y float64) float64 {
		// Class label with underline
		f.SetFont("Verdana", "B", 11)
		f.SetTextColor(52, 73, 94)
		f.SetXY(x, y)
		f.CellFormat(twoColW, 7, cg.label, "", 0, "L", false, 0, "")
		f.SetDrawColor(52, 73, 94)
		f.SetLineWidth(0.4)
		f.Line(x, y+7, x+twoColW, y+7)
		y += 8

		// Column header row
		setHeaderFill()
		f.SetFont("Verdana", "B", 9)
		f.SetTextColor(255, 255, 255)
		f.SetDrawColor(180, 180, 180)
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
			if rank%2 == 1 {
				setAltFill()
			} else {
				setClearFill()
			}
			if rank == 0 {
				f.SetFont("Verdana", "B", 9) // winner in bold
			} else {
				f.SetFont("Verdana", "", 9)
			}

			e180 := ""
			if s.OneEighties > 0 {
				e180 = fmt.Sprintf("%d", s.OneEighties)
			}
			cells := []string{
				fmt.Sprintf("%d", rank+1),
				s.Player.Nr,
				s.Player.Name,
				fmt.Sprintf("%d", s.PointsFor),
				fmt.Sprintf("%d", s.PointsAgainst),
				e180,
			}
			xc := x
			for i, cell := range cells {
				f.SetXY(xc, y)
				f.CellFormat(cw[i], rowH, cell, "1", 0, ca[i], true, 0, "")
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

	// --- Render landscape pages (pairs of classes side-by-side) ---
	for i := 0; i < len(classOrder); i += 2 {
		f.AddPage()
		startY := float64(marginL)

		leftCls := classMap[classOrder[i]]
		drawClassTable(leftCls, marginL, startY)

		if i+1 < len(classOrder) {
			rightCls := classMap[classOrder[i+1]]
			drawClassTable(rightCls, col2X, startY)
		}
	}

	// --- Duty stats page (portrait A4) ---
	const portraitUsableW = 210.0 - 2*marginL // 182mm

	dutyW := []float64{9, 13, 140, 20} // 182mm total
	dutyH := []string{"#", "Nr", "Naam", "Keer"}
	dutyA := []string{"C", "C", "L", "C"}

	f.AddPageFormat("P", gofpdf.SizeType{Wd: 210, Ht: 297})
	headerY := float64(marginL)

	// Header row
	setHeaderFill()
	f.SetFont("Verdana", "B", 9)
	f.SetTextColor(255, 255, 255)
	xc := float64(marginL)
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
		if ri%2 == 1 {
			setAltFill()
		} else {
			setClearFill()
		}
		if rank == 1 {
			f.SetFont("Verdana", "B", 9)
		} else {
			f.SetFont("Verdana", "", 9)
		}
		cells := []string{
			fmt.Sprintf("%d", rank),
			d.Player.Nr,
			d.Player.Name,
			fmt.Sprintf("%d", d.Count),
		}
		xc := float64(marginL)
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

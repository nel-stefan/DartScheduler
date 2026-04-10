package pdf

import (
	"fmt"
	"io"
	"sort"

	"DartScheduler/usecase"

	"github.com/jung-kurt/gofpdf"
)

// ExportStandings writes a klassement PDF to w.
// stats must be pre-sorted (or will be sorted inside per class by points).
func ExportStandings(competitionName, season string, stats []usecase.PlayerStats, dutyStats []usecase.DutyStats, w io.Writer) error {
	f := gofpdf.New("P", "mm", "A4", "")
	tr := f.UnicodeTranslatorFromDescriptor("") // UTF-8 → Latin-1

	title := competitionName
	if season != "" {
		title += " — " + season
	}

	writePageHeader := func(subtitle string) {
		f.AddPage()
		f.SetFont("Arial", "B", 14)
		f.CellFormat(0, 8, tr(title), "", 1, "C", false, 0, "")
		f.SetFont("Arial", "", 9)
		f.CellFormat(0, 5, tr(subtitle), "", 1, "C", false, 0, "")
		f.Ln(4)
	}

	// Group by class, preserving insertion order of classes seen
	type classGroup struct {
		label string
		stats []usecase.PlayerStats
	}
	classOrder := []string{}
	classMap := map[string]*classGroup{}
	for _, s := range stats {
		cls := s.Player.Class
		label := "Klasse " + cls
		if cls == "" {
			label = "Overig"
		}
		if _, ok := classMap[cls]; !ok {
			classOrder = append(classOrder, cls)
			classMap[cls] = &classGroup{label: label}
		}
		classMap[cls].stats = append(classMap[cls].stats, s)
	}
	// Sort each class by points descending
	for _, cg := range classMap {
		sort.Slice(cg.stats, func(i, j int) bool {
			a, b := cg.stats[i], cg.stats[j]
			if a.PointsFor != b.PointsFor {
				return a.PointsFor > b.PointsFor
			}
			return (a.PointsFor - a.PointsAgainst) > (b.PointsFor - b.PointsAgainst)
		})
	}

	colW := []float64{10, 14, 70, 22, 22, 16}
	colH := []string{"#", "Nr", "Naam", "+ pnt", "- pnt", "180"}
	colA := []string{"C", "C", "L", "C", "C", "C"}

	for _, cls := range classOrder {
		cg := classMap[cls]
		writePageHeader("Klassement")

		f.SetFont("Arial", "B", 11)
		f.CellFormat(0, 7, tr(cg.label), "", 1, "L", false, 0, "")
		f.Ln(1)

		f.SetFont("Arial", "B", 9)
		for i, h := range colH {
			f.CellFormat(colW[i], 7, h, "1", 0, colA[i], false, 0, "")
		}
		f.Ln(-1)

		f.SetFont("Arial", "", 9)
		for rank, s := range cg.stats {
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
			for i, cell := range cells {
				f.CellFormat(colW[i], 6, cell, "1", 0, colA[i], false, 0, "")
			}
			f.Ln(-1)
		}
	}

	// Duty stats page
	writePageHeader("Schrijver / Teller")
	f.SetFont("Arial", "", 9)
	f.CellFormat(0, 5, "Totaal aantal keer als schrijver of teller ingezet (gecombineerd).", "", 1, "L", false, 0, "")
	f.Ln(2)

	dutyW := []float64{10, 14, 100, 20}
	dutyH := []string{"#", "Nr", "Naam", "Keer"}
	dutyA := []string{"C", "C", "L", "C"}

	f.SetFont("Arial", "B", 9)
	for i, h := range dutyH {
		f.CellFormat(dutyW[i], 7, h, "1", 0, dutyA[i], false, 0, "")
	}
	f.Ln(-1)

	f.SetFont("Arial", "", 9)
	rank := 1
	for _, d := range dutyStats {
		if d.Count == 0 {
			continue
		}
		cells := []string{
			fmt.Sprintf("%d", rank),
			tr(d.Player.Nr),
			tr(d.Player.Name),
			fmt.Sprintf("%d", d.Count),
		}
		for i, cell := range cells {
			f.CellFormat(dutyW[i], 6, cell, "1", 0, dutyA[i], false, 0, "")
		}
		f.Ln(-1)
		rank++
	}

	return f.Output(w)
}

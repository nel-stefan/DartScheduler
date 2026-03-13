package pdf

import (
	"context"
	"fmt"
	"io"

	"DartScheduler/domain"

	"github.com/jung-kurt/gofpdf"
)

// Exporter implements usecase.Exporter for PDF output.
type Exporter struct{}

func (e Exporter) Export(_ context.Context, sched domain.Schedule, players []domain.Player, w io.Writer) error {
	playerName := make(map[domain.PlayerID]string, len(players))
	for _, p := range players {
		playerName[p.ID] = p.Name
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetFont("Arial", "B", 14)

	for _, ev := range sched.Evenings {
		pdf.AddPage()
		pdf.CellFormat(0, 10,
			fmt.Sprintf("%s – Evening %d (%s)", sched.CompetitionName, ev.Number, ev.Date.Format("2006-01-02")),
			"", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "B", 10)
		cols := []string{"#", "Player A", "Player B", "Score A", "Score B"}
		widths := []float64{10, 70, 70, 25, 25}
		for i, c := range cols {
			pdf.CellFormat(widths[i], 8, c, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 9)
		for i, m := range ev.Matches {
			scoreA, scoreB := "-", "-"
			if m.ScoreA != nil {
				scoreA = fmt.Sprintf("%d", *m.ScoreA)
			}
			if m.ScoreB != nil {
				scoreB = fmt.Sprintf("%d", *m.ScoreB)
			}
			row := []string{
				fmt.Sprintf("%d", i+1),
				playerName[m.PlayerA],
				playerName[m.PlayerB],
				scoreA,
				scoreB,
			}
			for j, cell := range row {
				pdf.CellFormat(widths[j], 7, cell, "1", 0, "L", false, 0, "")
			}
			pdf.Ln(-1)
		}
	}

	return pdf.Output(w)
}

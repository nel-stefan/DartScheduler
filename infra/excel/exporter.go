package excel

import (
	"context"
	"fmt"
	"io"

	"DartScheduler/domain"

	"github.com/xuri/excelize/v2"
)

// Exporter implements usecase.Exporter for Excel output.
type Exporter struct{}

func (e Exporter) Export(_ context.Context, sched domain.Schedule, players []domain.Player, w io.Writer) error {
	// Build player lookup.
	playerName := make(map[domain.PlayerID]string, len(players))
	for _, p := range players {
		playerName[p.ID] = p.Name
	}

	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Schedule
	sheet := "Schedule"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Evening", "Date", "Player A", "Player B", "Score A", "Score B"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	row := 2
	for _, ev := range sched.Evenings {
		for _, m := range ev.Matches {
			scoreA, scoreB := "", ""
			if m.ScoreA != nil {
				scoreA = fmt.Sprintf("%d", *m.ScoreA)
			}
			if m.ScoreB != nil {
				scoreB = fmt.Sprintf("%d", *m.ScoreB)
			}
			vals := []interface{}{
				ev.Number,
				ev.Date.Format("2006-01-02"),
				playerName[m.PlayerA],
				playerName[m.PlayerB],
				scoreA,
				scoreB,
			}
			for col, v := range vals {
				cell, _ := excelize.CoordinatesToCellName(col+1, row)
				f.SetCellValue(sheet, cell, v)
			}
			row++
		}
	}

	return f.Write(w)
}

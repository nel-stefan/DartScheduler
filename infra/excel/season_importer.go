package excel

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ImportedSeason is the result of parsing a season Excel file.
type ImportedSeason struct {
	Matches        []SeasonMatchRow
	InhaalEvenings []InhaalEvening
}

// InhaalEvening represents a catch-up evening (inhaalavond) with no pre-assigned matches.
type InhaalEvening struct {
	EveningNr int
	Date      time.Time
}

// SeasonMatchRow represents one imported match row from a historical season Excel.
type SeasonMatchRow struct {
	EveningNr  int
	Date       time.Time
	NrA        string
	NameA      string
	NrB        string
	NameB      string
	Leg1Winner string // player nr of winner
	Leg1Turns  int
	Leg2Winner string
	Leg2Turns  int
	Leg3Winner string
	Leg3Turns  int
	ScoreA     int
	ScoreB     int
	Secretary  string
	Counter    string
}

// matchPairRe matches cells like "32 - 35" or "32 (Jeffrey Vrieze) - 35 (Daan Schurink)".
var matchPairRe = regexp.MustCompile(`^(\d+)\s*(?:\([^)]*\))?\s*-\s*(\d+)\s*(?:\([^)]*\))?$`)

// matchPairNameRe additionally captures names: "32 (Jeffrey Vrieze) - 35 (Daan Schurink)".
var matchPairNameRe = regexp.MustCompile(`^(\d+)\s*\(([^)]+)\)\s*-\s*(\d+)\s*\(([^)]+)\)$`)

// ImportSeason reads a historical season Excel file.
// Supports two formats:
//
// 1. Flat table (headers: avond/evening, datum/date, nr a/naam a, nr b/naam b, leg1, beurten1, …)
// 2. Schedule matrix (row 1 = dates, subsequent rows = "NR - NR" match pairs per column;
//    columns whose cells spell "INHAAL" are treated as inhaalavonden)
func ImportSeason(r io.Reader) (ImportedSeason, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return ImportedSeason{}, fmt.Errorf("season import: open: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return ImportedSeason{}, fmt.Errorf("season import: no sheets found")
	}

	// Try each sheet in order: prefer first sheet that yields results.
	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) < 2 {
			continue
		}
		if isFlatTableHeader(rows[0]) {
			matches, err := parseFlatTable(rows)
			return ImportedSeason{Matches: matches}, err
		}
		if result, ok := parseMatrixSchedule(rows); ok {
			return result, nil
		}
	}

	return ImportedSeason{}, fmt.Errorf("season import: could not recognise format — expected either a flat table (with avond/datum/nr columns) or a schedule matrix (dates in row 1, 'NR - NR' match pairs in cells)")
}

// isFlatTableHeader returns true when the header row contains flat-table keywords.
func isFlatTableHeader(header []string) bool {
	norm := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "_", "")
		return s
	}
	keywords := []string{"avond", "evening", "nra", "nrb", "naama", "naamb", "spelera", "spelerb"}
	for _, h := range header {
		n := norm(h)
		for _, kw := range keywords {
			if n == kw {
				return true
			}
		}
	}
	return false
}

// parseMatrixSchedule parses a schedule matrix where row 0 = dates and each
// subsequent row contains "NR - NR" match pairs per evening column.
// Columns whose non-empty cell values concatenate to contain "inhaal" (case-insensitive)
// are treated as inhaalavonden (catch-up evenings) with no pre-assigned matches.
// Returns (result, true) when the sheet looks like a schedule matrix, (zero, false) otherwise.
func parseMatrixSchedule(rows [][]string) (ImportedSeason, bool) {
	header := rows[0]

	type eveningCol struct {
		colIdx    int
		eveningNr int
		date      time.Time
	}
	var cols []eveningCol
	eveningNr := 0
	var prevMonth time.Month
	var inferredYear int

	for ci, cell := range header {
		d, ok := parseFlexDate(strings.TrimSpace(cell), &prevMonth, &inferredYear)
		if !ok {
			continue
		}
		eveningNr++
		cols = append(cols, eveningCol{colIdx: ci, eveningNr: eveningNr, date: d})
	}

	if len(cols) == 0 {
		return ImportedSeason{}, false
	}

	// Determine which columns are inhaalavond by concatenating their cell values.
	inhaalCols := make(map[int]bool, len(cols))
	for _, ec := range cols {
		var sb strings.Builder
		for _, row := range rows[1:] {
			if ec.colIdx < len(row) {
				sb.WriteString(strings.TrimSpace(row[ec.colIdx]))
			}
		}
		if strings.Contains(strings.ToLower(sb.String()), "inhaal") {
			inhaalCols[ec.colIdx] = true
		}
	}

	var result ImportedSeason
	for _, ec := range cols {
		if inhaalCols[ec.colIdx] {
			result.InhaalEvenings = append(result.InhaalEvenings, InhaalEvening{
				EveningNr: ec.eveningNr,
				Date:      ec.date,
			})
			continue
		}
		for _, row := range rows[1:] {
			if ec.colIdx >= len(row) {
				continue
			}
			cell := strings.TrimSpace(row[ec.colIdx])
			if cell == "" {
				continue
			}
			// Try "NR (Name) - NR (Name)" first.
			if m := matchPairNameRe.FindStringSubmatch(cell); m != nil {
				result.Matches = append(result.Matches, SeasonMatchRow{
					EveningNr: ec.eveningNr,
					Date:      ec.date,
					NrA:       m[1],
					NameA:     strings.TrimSpace(m[2]),
					NrB:       m[3],
					NameB:     strings.TrimSpace(m[4]),
				})
				continue
			}
			// Try plain "NR - NR".
			if m := matchPairRe.FindStringSubmatch(cell); m != nil {
				result.Matches = append(result.Matches, SeasonMatchRow{
					EveningNr: ec.eveningNr,
					Date:      ec.date,
					NrA:       m[1],
					NrB:       m[2],
				})
			}
			// Otherwise skip (e.g. single letters, club name acronyms, etc.)
		}
	}
	return result, true
}

// Dutch month abbreviations used in schedule files.
var dutchMonths = map[string]time.Month{
	"jan": time.January, "feb": time.February, "mrt": time.March,
	"apr": time.April, "mei": time.May, "jun": time.June,
	"jul": time.July, "aug": time.August, "sep": time.September,
	"okt": time.October, "nov": time.November, "dec": time.December,
}

// parseFlexDate parses a date string with multiple formats:
//   - "05-sep" (Dutch abbrev, year inferred by rolling)
//   - "2025-09-05" or " 2025-09-05" (ISO)
//   - "5-9-2025" / "05-09-2025" (day-month-year)
//   - "09/05/2025" (US)
//
// prevMonth and inferredYear are updated on each successful parse to enable year inference.
func parseFlexDate(s string, prevMonth *time.Month, inferredYear *int) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}

	// ISO: "2025-09-05"
	if t, err := time.Parse("2006-01-02", s); err == nil {
		*prevMonth = t.Month()
		*inferredYear = t.Year()
		return t, true
	}
	// Other standard formats.
	for _, layout := range []string{"2-1-2006", "02-01-2006", "01/02/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			*prevMonth = t.Month()
			*inferredYear = t.Year()
			return t, true
		}
	}

	// Dutch "DD-mon" e.g. "05-sep", "5-sep"
	parts := strings.SplitN(s, "-", 2)
	if len(parts) == 2 {
		dayStr := strings.TrimSpace(parts[0])
		monStr := strings.ToLower(strings.TrimSpace(parts[1]))
		if month, ok := dutchMonths[monStr]; ok {
			day, err := strconv.Atoi(dayStr)
			if err == nil {
				year := inferYear(month, prevMonth, inferredYear)
				*prevMonth = month
				return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), true
			}
		}
	}

	return time.Time{}, false
}

// inferYear determines the year for a date given only month, using rolling logic.
// If no previous month/year has been seen, it starts from the current year
// (or current year-1 if the month is in Jul-Dec, typical season start).
func inferYear(month time.Month, prevMonth *time.Month, inferredYear *int) int {
	if *inferredYear == 0 {
		now := time.Now()
		if month >= time.July {
			*inferredYear = now.Year() - 1
		} else {
			*inferredYear = now.Year()
		}
	}
	// If month wrapped back (e.g. we were in Dec and now see Jan), advance year.
	if *prevMonth != 0 && month < *prevMonth {
		*inferredYear++
	}
	return *inferredYear
}

// parseFlatTable parses the original flat-table format with explicit column headers.
func parseFlatTable(rows [][]string) ([]SeasonMatchRow, error) {
	if len(rows) < 2 {
		return nil, fmt.Errorf("season import: file has no data rows")
	}

	norm := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "_", "")
		return s
	}
	header := rows[0]
	col := func(names ...string) int {
		for _, name := range names {
			for i, h := range header {
				if norm(h) == name {
					return i
				}
			}
		}
		return -1
	}
	cell := func(row []string, c int) string {
		if c < 0 || c >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[c])
	}
	atoi := func(s string) int {
		n, _ := strconv.Atoi(strings.TrimSpace(s))
		return n
	}
	parseDate := func(s string) time.Time {
		var prevM time.Month
		var yr int
		t, _ := parseFlexDate(s, &prevM, &yr)
		return t
	}

	cEvening := col("avond", "evening", "avondnr", "nr")
	cDate := col("datum", "date")
	cNrA := col("nra", "nr1", "nrspelera", "nrspeler1")
	cNameA := col("naama", "spelera", "speler1", "naam1")
	if cNameA < 0 {
		cNameA = col("naam")
	}
	cNrB := col("nrb", "nr2", "nrspelerb", "nrspeler2")
	cNameB := col("naamb", "spelerb", "speler2", "naam2")
	cLeg1 := col("leg1", "partij1", "winnaar1", "w1")
	cBeurten1 := col("beurten1", "aantalbeurt1", "turns1")
	cLeg2 := col("leg2", "partij2", "winnaar2", "w2")
	cBeurten2 := col("beurten2", "aantalbeurt2", "turns2")
	cLeg3 := col("leg3", "partij3", "winnaar3", "w3")
	cBeurten3 := col("beurten3", "aantalbeurt3", "turns3")
	cScore := col("score", "eindstand", "uitslag")
	cSecr := col("schrijver", "secretary", "nrschrijver")
	cTeller := col("teller", "counter", "nrteller")

	var out []SeasonMatchRow
	for _, row := range rows[1:] {
		nrA := cell(row, cNrA)
		nrB := cell(row, cNrB)
		if nrA == "" && nrB == "" {
			continue
		}
		scoreA, scoreB := 0, 0
		if sc := cell(row, cScore); sc != "" {
			parts := strings.SplitN(sc, "-", 2)
			if len(parts) == 2 {
				scoreA = atoi(parts[0])
				scoreB = atoi(parts[1])
			}
		}
		out = append(out, SeasonMatchRow{
			EveningNr:  atoi(cell(row, cEvening)),
			Date:       parseDate(cell(row, cDate)),
			NrA:        nrA,
			NameA:      cell(row, cNameA),
			NrB:        nrB,
			NameB:      cell(row, cNameB),
			Leg1Winner: cell(row, cLeg1),
			Leg1Turns:  atoi(cell(row, cBeurten1)),
			Leg2Winner: cell(row, cLeg2),
			Leg2Turns:  atoi(cell(row, cBeurten2)),
			Leg3Winner: cell(row, cLeg3),
			Leg3Turns:  atoi(cell(row, cBeurten3)),
			ScoreA:     scoreA,
			ScoreB:     scoreB,
			Secretary:  cell(row, cSecr),
			Counter:    cell(row, cTeller),
		})
	}
	return out, nil
}

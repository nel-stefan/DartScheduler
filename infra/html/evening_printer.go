// Package html provides an HTML wedstrijdformulier renderer.
// The generated page reproduces the Blanco wedstrijdformulier.xlsx layout and
// is intended for in-browser printing. The <thead> element causes the column
// header to repeat automatically on every printed page.
package html

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

	"DartScheduler/domain"
)

// EveningPrinter implements usecase.EveningExporter and produces a printable HTML page.
type EveningPrinter struct {
	ClubName string
	LogoPath string
}

func (e EveningPrinter) ExportEvening(ctx context.Context, sched domain.Schedule, ev domain.Evening, players []domain.Player, w io.Writer) error {
	return PrintEvening(ctx, sched, ev, players, e.ClubName, e.LogoPath, w)
}

type printRow struct {
	NrA, NameA                 string
	NrB, NameB                 string
	Leg1Winner, Leg1Turns      string
	Leg2Winner, Leg2Turns      string
	Leg3Winner, Leg3Turns      string
	TotalWinner, Eindstand     string
	ReportedBy, RescheduleDate string
	SecretaryNr, CounterNr     string
}

type printData struct {
	DateLabel  string
	ClubName   string
	LogoImgTag string // non-empty when a logo file was provided
	Rows       []printRow
}

// PrintEvening renders the wedstrijdformulier as a self-contained HTML page.
func PrintEvening(_ context.Context, _ domain.Schedule, ev domain.Evening, players []domain.Player, clubName, logoPath string, w io.Writer) error {
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

	// Fill to the next multiple of rowsPerPage so blank rows are available.
	const rowsPerPage = 20
	matchCount := len(ev.Matches)
	pages := (matchCount + rowsPerPage - 1) / rowsPerPage
	if pages < 1 {
		pages = 1
	}
	totalRows := pages * rowsPerPage

	rows := make([]printRow, totalRows)
	for i := range rows {
		if i >= len(ev.Matches) {
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

		rows[i] = printRow{
			NrA:            pA.Nr,
			NameA:          domain.FormatDisplayName(pA.Name),
			NrB:            pB.Nr,
			NameB:          domain.FormatDisplayName(pB.Name),
			Leg1Winner:     playerLabel(m.Leg1Winner),
			Leg1Turns:      turns(m.Leg1Turns),
			Leg2Winner:     playerLabel(m.Leg2Winner),
			Leg2Turns:      turns(m.Leg2Turns),
			Leg3Winner:     playerLabel(m.Leg3Winner),
			Leg3Turns:      turns(m.Leg3Turns),
			TotalWinner:    totalWinner,
			Eindstand:      eindstand,
			ReportedBy:     reportedByLabel(m.ReportedBy),
			RescheduleDate: m.RescheduleDate,
			SecretaryNr:    m.SecretaryNr,
			CounterNr:      m.CounterNr,
		}
	}

	logoImgTag := ""
	if logoPath != "" {
		if data, err := os.ReadFile(logoPath); err == nil {
			mime := http.DetectContentType(data)
			logoImgTag = fmt.Sprintf(`<img src="data:%s;base64,%s" class="hdr-logo" alt="">`,
				mime, base64.StdEncoding.EncodeToString(data))
		}
	}

	tmpl, err := template.New("print").Parse(printTmpl)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, printData{DateLabel: dateLabel, ClubName: clubName, LogoImgTag: logoImgTag, Rows: rows})
}

const printTmpl = `<!DOCTYPE html>
<html lang="nl">
<head>
<meta charset="UTF-8">
<title>Wedstrijdformulier – {{.DateLabel}}</title>
<style>
@page { size: A4 landscape; margin: 19mm 6mm; }
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: Calibri, 'Segoe UI', Arial, sans-serif; background: white; padding: 6px; }

.no-print { margin-bottom: 8px; }
.btn-print {
  padding: 7px 18px; font-size: 13px; cursor: pointer;
  background: #1565c0; color: white; border: none; border-radius: 4px;
}
@media print { .no-print { display: none; } }

.hdr-logo { display: block; max-height: 32px; margin: 0 auto 4px; }
.hdr-title {
  font-size: 16pt; font-weight: bold; text-align: center; line-height: 1.3; padding: 2px 0;
}
.hdr-sub {
  font-size: 10pt; font-weight: bold; text-align: center; line-height: 1.3; padding: 0 0 4px;
}

table {
  border-collapse: collapse;
  width: 100%;
  table-layout: fixed;
  border: 1.5pt solid #000;
}
th, td {
  border: 0.5pt solid #000;
  padding: 1px 2px;
  font-family: Calibri, Arial, sans-serif;
  font-size: 7.5pt;
  vertical-align: middle;
  overflow: hidden;
}
th {
  font-weight: bold;
  text-align: center;
  white-space: normal;
  word-break: break-word;
  line-height: 1.15;
  min-height: 40pt;
}
tbody tr { height: 17.25pt; }

/* Column widths */
col.ca { width: 3.7%;  }
col.cb { width: 13.5%; }
col.cc { width: 1.6%;  }
col.cd { width: 3.7%;  }
col.ce { width: 13.5%; }
col.cf { width: 7.2%;  }
col.cg { width: 4.7%;  }
col.ch { width: 7.2%;  }
col.ci { width: 4.7%;  }
col.cj { width: 7.2%;  }
col.ck { width: 4.7%;  }
col.cl { width: 7.2%;  }
col.cm { width: 4.6%;  }
col.cn { width: 8.1%;  }
col.co { width: 5.1%;  }
col.cp { width: 4.1%;  }
col.cq { width: 4.1%;  }

/* Group separator borders */
.sep  { border-right: 1pt   solid #000; }
.sepl { border-left:  1pt   solid #000; }

/* Alignment */
.ar { text-align: right;  }
.ac { text-align: center; }
</style>
</head>
<body>
<div class="no-print">
  <button class="btn-print" onclick="window.print()">Afdrukken</button>
</div>

{{if .LogoImgTag}}{{.LogoImgTag}}{{end}}<div class="hdr-title">{{.ClubName}}</div>
<div class="hdr-sub">Wedstrijdformulier &nbsp;&nbsp;&nbsp; Spelsoort: 501 dubbel uit best of 3 &nbsp;&nbsp;&nbsp; Speeldatum: {{.DateLabel}}</div>

<table>
<colgroup>
  <col class="ca"><col class="cb"><col class="cc"><col class="cd"><col class="ce">
  <col class="cf"><col class="cg"><col class="ch"><col class="ci">
  <col class="cj"><col class="ck"><col class="cl"><col class="cm">
  <col class="cn"><col class="co"><col class="cp"><col class="cq">
</colgroup>
<thead>
  <tr>
    <th>nr.</th>
    <th class="sep">naam</th>
    <th class="sep">/</th>
    <th>nr.</th>
    <th class="sep">naam</th>
    <th>winnaar<br>(naam + nr.)<br>leg 1</th>
    <th class="sep">aantal<br>beurten</th>
    <th>winnaar<br>(naam + nr.)<br>leg 2</th>
    <th class="sep">aantal<br>beurten</th>
    <th>winnaar<br>(naam + nr.)<br>leg 3</th>
    <th class="sep">aantal<br>beurten</th>
    <th>winnaar<br>(naam + nr.)</th>
    <th class="sep">Eind-<br>stand</th>
    <th class="sep">afgemeld<br>door<br>(naam + nr.)</th>
    <th class="sep">vooruit-<br>gooi<br>datum</th>
    <th class="sep">nr.<br>schrij-<br>ver</th>
    <th>nr.<br>tel-<br>ler</th>
  </tr>
</thead>
<tbody>
{{range .Rows}}<tr>
  <td class="ar">{{.NrA}}</td>
  <td class="sep">{{.NameA}}</td>
  <td class="ac sep">/</td>
  <td class="ar">{{.NrB}}</td>
  <td class="sep">{{.NameB}}</td>
  <td>{{.Leg1Winner}}</td>
  <td class="ac sep">{{.Leg1Turns}}</td>
  <td>{{.Leg2Winner}}</td>
  <td class="ac sep">{{.Leg2Turns}}</td>
  <td>{{.Leg3Winner}}</td>
  <td class="ac sep">{{.Leg3Turns}}</td>
  <td>{{.TotalWinner}}</td>
  <td class="ac sep">{{.Eindstand}}</td>
  <td class="sep">{{.ReportedBy}}</td>
  <td class="sep">{{.RescheduleDate}}</td>
  <td class="ac sep">{{.SecretaryNr}}</td>
  <td class="ac">{{.CounterNr}}</td>
</tr>
{{end}}</tbody>
</table>
</body>
</html>`

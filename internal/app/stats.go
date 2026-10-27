package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"dactylo/internal/core/stats"
	"dactylo/internal/store"
)

// Fenêtres de calcul du tableau de bord.
const (
	ChartSessions   = 200 // points de la courbe
	HeatmapSessions = 50  // sessions analysées pour la carte du clavier
	TopBigrams      = 10
)

// Dashboard est le tableau de bord (F-33, F-34).
type Dashboard struct {
	Points  []store.SessionPoint `json:"points"`
	Totals  store.Totals         `json:"totals"`
	Streak  int                  `json:"streak"`
	Keys    []stats.KeyStat      `json:"keys"`
	Bigrams []stats.BigramStat   `json:"bigrams"`
}

// ExportResult indique où l'export a été écrit ("" si annulé).
type ExportResult struct {
	Path     string `json:"path"`
	Sessions int    `json:"sessions"`
}

// StatsService calcule les statistiques et l'export CSV.
type StatsService struct {
	store     *store.Store
	profileID int64
	picker    FilePicker
	now       func() time.Time
}

// NewStatsService crée le service.
func NewStatsService(s *store.Store, profileID int64, picker FilePicker) *StatsService {
	return &StatsService{store: s, profileID: profileID, picker: picker, now: time.Now}
}

// Dashboard rassemble courbes, totaux, série et carte du clavier.
func (s *StatsService) Dashboard() (Dashboard, error) {
	ctx := context.Background()
	points, err := s.store.SessionPoints(ctx, s.profileID, ChartSessions)
	if err != nil {
		return Dashboard{}, err
	}
	totals, err := s.store.SessionTotals(ctx, s.profileID)
	if err != nil {
		return Dashboard{}, err
	}
	days, err := s.store.SessionDays(ctx, s.profileID)
	if err != nil {
		return Dashboard{}, err
	}
	recent, err := s.store.RecentKeystrokes(ctx, s.profileID, HeatmapSessions)
	if err != nil {
		return Dashboard{}, err
	}
	keys, bigrams := stats.KeyStats(recent, TopBigrams)
	return Dashboard{
		Points: points, Totals: totals, Streak: stats.Streak(days, s.now()),
		Keys: keys, Bigrams: bigrams,
	}, nil
}

// ExportCSV fait choisir un fichier puis y écrit toutes les sessions (F-35).
func (s *StatsService) ExportCSV(dialogTitle string) (ExportResult, error) {
	points, err := s.store.SessionPoints(context.Background(), s.profileID, 0)
	if err != nil {
		return ExportResult{}, err
	}
	name := "dactylo-sessions-" + s.now().Format(time.DateOnly) + ".csv"
	path, err := s.picker.SaveCSV(dialogTitle, name)
	if err != nil || path == "" {
		return ExportResult{}, err
	}
	data, err := SessionsCSV(points)
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Path: path, Sessions: len(points)}, nil
}

// SessionsCSV produit un CSV lisible par un tableur français : séparateur
// « ; », virgule décimale, UTF-8 avec BOM.
func SessionsCSV(points []store.SessionPoint) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("\uFEFF")
	w := csv.NewWriter(&buf)
	w.Comma = ';'
	num := func(f float64) string {
		return strings.Replace(strconv.FormatFloat(f, 'f', 1, 64), ".", ",", 1)
	}
	rows := [][]string{{"date", "mode", "texte", "duree_s", "mpm_brut", "mpm_net", "cpm", "precision_pct", "mots_fautifs"}}
	for _, p := range points {
		date := p.EndedAt
		if t, err := time.Parse(time.RFC3339Nano, p.EndedAt); err == nil {
			date = t.Local().Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{
			date, p.Mode, p.TextTitle, strconv.FormatInt(p.DurationMs/1000, 10),
			num(p.WPMGross), num(p.WPMNet), num(p.CPM), num(p.Accuracy), strconv.Itoa(p.Errors),
		})
	}
	if err := w.WriteAll(rows); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

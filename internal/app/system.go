// Service d'informations système (version, plateforme, base).

package app

import (
	"context"
	"runtime"

	"dactylo/internal/store"
)

// SystemInfo décrit l'installation (affichée sur l'accueil et « À propos »).
type SystemInfo struct {
	Version       string `json:"version"`
	Platform      string `json:"platform"`
	DatabasePath  string `json:"databasePath"`
	SchemaVersion int64  `json:"schemaVersion"`
}

// SystemService expose des informations générales sur l'application.
type SystemService struct {
	store *store.Store
}

// NewSystemService crée le service.
func NewSystemService(s *store.Store) *SystemService {
	return &SystemService{store: s}
}

// Info renvoie la version, la plateforme et l'état de la base.
func (s *SystemService) Info() (SystemInfo, error) {
	v, err := s.store.SchemaVersion(context.Background())
	if err != nil {
		return SystemInfo{}, err
	}
	return SystemInfo{
		Version:       Version,
		Platform:      runtime.GOOS + "/" + runtime.GOARCH,
		DatabasePath:  s.store.Path,
		SchemaVersion: v,
	}, nil
}

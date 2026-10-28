package app

import (
	"context"
	"time"

	"dactylo/internal/store"
)

// BackupResult indique le fichier écrit ou restauré ("" si annulé).
type BackupResult struct {
	Path string `json:"path"`
}

// BackupService sauvegarde et restaure toutes les données dans un fichier
// unique, copie de la base SQLite (F-52). Les fichiers passent toujours par
// les boîtes de dialogue système (NF-11).
type BackupService struct {
	store   *store.Store
	profile *ActiveProfile
	picker  FilePicker
	now     func() time.Time
}

// NewBackupService crée le service.
func NewBackupService(s *store.Store, profile *ActiveProfile, picker FilePicker) *BackupService {
	return &BackupService{store: s, profile: profile, picker: picker, now: time.Now}
}

// Export écrit une copie complète de la base dans le fichier choisi.
func (b *BackupService) Export(dialogTitle string) (BackupResult, error) {
	name := "dactylo-sauvegarde-" + b.now().Format(time.DateOnly) + ".db"
	path, err := b.picker.SaveFile(dialogTitle, name, FilterBackup)
	if err != nil || path == "" {
		return BackupResult{}, err
	}
	if err := b.store.Backup(context.Background(), path); err != nil {
		return BackupResult{}, err
	}
	return BackupResult{Path: path}, nil
}

// Restore remplace toutes les données par celles du fichier choisi (une
// copie de sécurité de la base actuelle est gardée à côté, en .bak).
// L'interface doit ensuite se recharger.
func (b *BackupService) Restore(dialogTitle string) (BackupResult, error) {
	ctx := context.Background()
	path, err := b.picker.PickFile(dialogTitle, FilterBackup)
	if err != nil || path == "" {
		return BackupResult{}, err
	}
	if err := b.store.Restore(ctx, path); err != nil {
		return BackupResult{}, err
	}
	// le profil actif peut ne pas exister dans la sauvegarde
	restored, err := RestoreActiveProfile(ctx, b.store)
	if err != nil {
		return BackupResult{}, err
	}
	b.profile.Set(restored.ID())
	return BackupResult{Path: path}, nil
}

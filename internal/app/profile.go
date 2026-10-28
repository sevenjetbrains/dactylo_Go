package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"dactylo/internal/core/text"
	"dactylo/internal/store"
)

// ActiveProfile est le profil en cours d'utilisation, partagé par tous les
// services (F-50). Changer de profil change les données de tous les modules.
type ActiveProfile struct {
	id atomic.Int64
}

// NewActiveProfile crée le profil actif.
func NewActiveProfile(id int64) *ActiveProfile {
	p := &ActiveProfile{}
	p.id.Store(id)
	return p
}

// ID renvoie l'identifiant du profil actif.
func (p *ActiveProfile) ID() int64 { return p.id.Load() }

// Set change de profil.
func (p *ActiveProfile) Set(id int64) { p.id.Store(id) }

// settingActiveProfile mémorise le dernier profil utilisé.
const settingActiveProfile = "active_profile"

// Layouts sont les dispositions de clavier proposées (F-08).
var Layouts = []string{"azerty", "qwerty", "bepo"}

// MaxProfileName borne la longueur d'un nom de profil.
const MaxProfileName = 40

// RestoreActiveProfile choisit le profil au démarrage : le dernier utilisé
// s'il existe encore, sinon le premier (créé au besoin).
func RestoreActiveProfile(ctx context.Context, s *store.Store) (*ActiveProfile, error) {
	id, err := s.EnsureDefaultProfile(ctx)
	if err != nil {
		return nil, err
	}
	if v, ok, err := s.GetSetting(ctx, settingActiveProfile); err == nil && ok {
		if saved, err := strconv.ParseInt(v, 10, 64); err == nil {
			if _, err := s.GetProfile(ctx, saved); err == nil {
				id = saved
			}
		}
	}
	return NewActiveProfile(id), nil
}

// ProfileService gère les profils locaux (F-50) et leurs réglages (F-51).
type ProfileService struct {
	store   *store.Store
	profile *ActiveProfile
}

// NewProfileService crée le service.
func NewProfileService(s *store.Store, profile *ActiveProfile) *ProfileService {
	return &ProfileService{store: s, profile: profile}
}

// Get renvoie le profil actif.
func (p *ProfileService) Get() (store.Profile, error) {
	return p.store.GetProfile(context.Background(), p.profile.ID())
}

// List renvoie tous les profils.
func (p *ProfileService) List() ([]store.Profile, error) {
	return p.store.ListProfiles(context.Background())
}

// Create ajoute un profil (sans changer de profil actif).
func (p *ProfileService) Create(name, layout string) (store.Profile, error) {
	ctx := context.Background()
	name, layout, err := cleanProfile(name, layout)
	if err != nil {
		return store.Profile{}, err
	}
	id, err := p.store.CreateProfile(ctx, name, layout)
	if err != nil {
		return store.Profile{}, err
	}
	return p.store.GetProfile(ctx, id)
}

// Update change le nom et la disposition de clavier d'un profil.
func (p *ProfileService) Update(id int64, name, layout string) (store.Profile, error) {
	ctx := context.Background()
	name, layout, err := cleanProfile(name, layout)
	if err != nil {
		return store.Profile{}, err
	}
	if err := p.store.UpdateProfile(ctx, id, name, layout); err != nil {
		return store.Profile{}, err
	}
	return p.store.GetProfile(ctx, id)
}

// Rename change le prénom du profil actif (vide : « Moi »).
func (p *ProfileService) Rename(name string) (store.Profile, error) {
	cur, err := p.Get()
	if err != nil {
		return store.Profile{}, err
	}
	return p.Update(cur.ID, name, cur.KeyboardLayout)
}

// Switch change de profil actif et le mémorise pour le prochain démarrage.
func (p *ProfileService) Switch(id int64) (store.Profile, error) {
	ctx := context.Background()
	prof, err := p.store.GetProfile(ctx, id)
	if err != nil {
		return store.Profile{}, err
	}
	if err := p.store.SetSetting(ctx, settingActiveProfile, strconv.FormatInt(id, 10)); err != nil {
		return store.Profile{}, err
	}
	p.profile.Set(id)
	return prof, nil
}

// Delete supprime un profil et toutes ses données. Le dernier profil ne peut
// pas être supprimé ; supprimer le profil actif bascule sur un autre.
func (p *ProfileService) Delete(id int64) error {
	ctx := context.Background()
	all, err := p.store.ListProfiles(ctx)
	if err != nil {
		return err
	}
	if len(all) <= 1 {
		return errors.New("impossible de supprimer le dernier profil")
	}
	if err := p.store.DeleteProfile(ctx, id); err != nil {
		return err
	}
	if p.profile.ID() == id {
		for _, other := range all {
			if other.ID != id {
				_, err := p.Switch(other.ID)
				return err
			}
		}
	}
	return nil
}

// SaveSettings enregistre les réglages de l'interface du profil actif
// (thème, langue, frappe, dictée : F-51).
func (p *ProfileService) SaveSettings(settings string) error {
	return p.SaveSettingsFor(p.profile.ID(), settings)
}

// SaveSettingsFor enregistre les réglages d'un profil donné (réglages par
// défaut d'un profil qu'on vient de créer).
func (p *ProfileService) SaveSettingsFor(id int64, settings string) error {
	if !json.Valid([]byte(settings)) {
		return errors.New("réglages invalides")
	}
	return p.store.SaveProfileSettings(context.Background(), id, settings)
}

func cleanProfile(name, layout string) (string, string, error) {
	name = strings.TrimSpace(text.NFC(name))
	if name == "" {
		name = store.DefaultProfileName
	}
	if utf8.RuneCountInString(name) > MaxProfileName {
		name = string([]rune(name)[:MaxProfileName])
	}
	if layout == "" {
		layout = Layouts[0]
	}
	for _, l := range Layouts {
		if l == layout {
			return name, layout, nil
		}
	}
	return "", "", fmt.Errorf("disposition de clavier inconnue : %q", layout)
}

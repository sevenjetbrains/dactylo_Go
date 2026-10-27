// Package events définit les événements de domaine et un bus interne
// synchrone. Les modules s'y abonnent au lieu de s'importer entre eux.
package events

import "sync"

// Noms des événements de domaine (cahier des charges, section 3.1).
const (
	SessionCompleted = "SessionCompleted"
	WordMistyped     = "WordMistyped"
	TextImported     = "TextImported"
)

// Event est un événement de domaine : un nom et une charge utile
// sérialisable en JSON (relayée telle quelle à l'interface).
type Event struct {
	Name    string `json:"name"`
	Payload any    `json:"payload"`
}

// Handler traite un événement.
type Handler func(Event)

type subscription struct {
	id   uint64
	name string // "" = tous les événements
	fn   Handler
}

// Bus diffuse les événements aux abonnés, dans l'ordre d'abonnement.
// Il est sûr en accès concurrent.
type Bus struct {
	mu     sync.RWMutex
	nextID uint64
	subs   []subscription
}

// NewBus crée un bus vide.
func NewBus() *Bus {
	return &Bus{}
}

// Subscribe abonne fn à l'événement name et renvoie une fonction de
// désabonnement.
func (b *Bus) Subscribe(name string, fn Handler) (unsubscribe func()) {
	return b.add(name, fn)
}

// SubscribeAll abonne fn à tous les événements (ex. relais vers l'interface).
func (b *Bus) SubscribeAll(fn Handler) (unsubscribe func()) {
	return b.add("", fn)
}

func (b *Bus) add(name string, fn Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	b.subs = append(b.subs, subscription{id: id, name: name, fn: fn})
	return func() { b.remove(id) }
}

func (b *Bus) remove(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, s := range b.subs {
		if s.id == id {
			b.subs = append(b.subs[:i:i], b.subs[i+1:]...)
			return
		}
	}
}

// Publish diffuse e à tous les abonnés concernés. Les gestionnaires sont
// appelés hors verrou : ils peuvent eux-mêmes publier ou se désabonner.
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	targets := make([]Handler, 0, len(b.subs))
	for _, s := range b.subs {
		if s.name == "" || s.name == e.Name {
			targets = append(targets, s.fn)
		}
	}
	b.mu.RUnlock()

	for _, fn := range targets {
		fn(e)
	}
}

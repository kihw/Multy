package services

import (
	"log"
	"sync"

	hook "github.com/robotn/gohook"
)

type Shortcut struct {
	ID         int
	Key        string
	WindowName string
}

type ShortcutService struct {
	mu            sync.Mutex
	shortcut      Shortcut
	windowService *WindowService
}

func NewShortcutService() *ShortcutService {
	return &ShortcutService{}
}

func (ss *ShortcutService) GenerateHotkeyID() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.shortcut.ID + 1
}
func (ss *ShortcutService) RegisterShortcut(shortcut Shortcut) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Make sure to log the shortcut details
	log.Printf("Registering shortcut: %+v", shortcut)

	// Remplacer le raccourci existant
	ss.shortcut = shortcut

	// Configurer l'écoute sur la nouvelle touche
	go ss.listenForKey(shortcut)

	return nil
}

func (ss *ShortcutService) listenForKey(shortcut Shortcut) {
	evChan := hook.Start()
	defer hook.End()

	keyChar := int32(shortcut.Key[0])

	log.Printf("Listening for key '%s' with Keychar %d", shortcut.Key, keyChar)

	for ev := range evChan {
		if ev.Kind == hook.KeyDown && ev.Keychar == keyChar {
			log.Printf("Key '%s' pressed", shortcut.Key)

			// Log the current shortcut window name
			log.Printf("Current shortcut window name: '%s'", shortcut.WindowName)

			// Ensure WindowName is not empty
			if shortcut.WindowName == "" {
				log.Println("WindowName is empty. Cannot focus.")
				return
			}

			// Attempt to focus the window
			err := ss.windowService.FocusWindowWithTitle(shortcut.WindowName)
			if err != nil {
				log.Printf("Failed to focus window '%s': %v", shortcut.WindowName, err)
			}
		}
	}
}

func (ss *ShortcutService) UnregisterShortcut() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Réinitialiser le raccourci
	ss.shortcut = Shortcut{}
}

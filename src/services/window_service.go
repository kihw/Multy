package services

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
)

const (
	SW_RESTORE     = 9
	SW_SHOW        = 5
	HWND_TOP       = 0
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_SHOWWINDOW = 0x0040
)

// WindowService est le service qui interagit avec les fenêtres sur Windows.
type WindowService struct{}

// GetWindows retourne la liste des fenêtres ouvertes.
// @Summary Retourne la liste des fenêtres ouvertes
// @Description Obtient la liste des fenêtres actuellement ouvertes sur le système
// @Tags Windows
// @Produce json
// @Success 200 {array} string "Liste des fenêtres ouvertes"
// @Failure 500 {object} string "Erreur lors de la récupération des fenêtres"
// @Router /windows [get]
func (ws *WindowService) GetWindows() ([]string, error) {
	var windows []string

	// EnumWindowsCallback est appelé pour chaque fenêtre trouvée
	enumWindowsCallback := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		var buf [256]uint16
		procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))

		// Convertir le nom de la fenêtre en string et l'ajouter à la liste
		windowText := syscall.UTF16ToString(buf[:])
		if len(windowText) > 0 {
			windows = append(windows, windowText)
		}
		return 1 // Continuer l'énumération
	})

	// Enumérer les fenêtres
	ret, _, err := procEnumWindows.Call(enumWindowsCallback, 0)
	if ret == 0 {
		return nil, fmt.Errorf("erreur lors de l'énumération des fenêtres: %v", err)
	}

	return windows, nil
}

// FocusWindowWithTitle met en avant une fenêtre spécifique.
// @Summary Met en avant une fenêtre spécifique
// @Description Met en avant une fenêtre qui contient un mot-clé spécifique dans son titre
// @Tags Windows
// @Produce json
// @Param keyword path string true "Mot-clé pour identifier la fenêtre"
// @Success 200 {object} string "Fenêtre mise en avant avec succès"
// @Failure 500 {object} string "Erreur lors de la mise en avant de la fenêtre"
// @Router /focus/{keyword} [post]
func (ws *WindowService) FocusWindowWithTitle(keyword string) error {
	if keyword == "" {
		return fmt.Errorf("le mot-clé du titre de la fenêtre est vide")
	}

	found := false
	enumWindowsCallback := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		var buf [256]uint16
		procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))

		windowText := syscall.UTF16ToString(buf[:])
		log.Printf("Checking window title: '%s'", windowText) // Log the window title

		if strings.Contains(windowText, keyword) {
			log.Printf("Found matching window: '%s'", windowText) // Log when a match is found

			// Restore the window if minimized
			procShowWindow.Call(uintptr(hwnd), SW_RESTORE)

			// Get the thread ID of the foreground window
			fgWindow, _, _ := procGetForegroundWindow.Call()
			fgThreadID, _, _ := procGetWindowThreadProcessId.Call(fgWindow, 0)

			// Get the thread ID of the target window
			targetThreadID, _, _ := procGetWindowThreadProcessId.Call(uintptr(hwnd), 0)

			// Attach the foreground thread and the target window's thread
			procAttachThreadInput.Call(fgThreadID, targetThreadID, 1)

			// Bring the window to the top of the Z-order
			procSetWindowPos.Call(uintptr(hwnd), HWND_TOP, 0, 0, 0, 0, SWP_NOMOVE|SWP_NOSIZE|SWP_SHOWWINDOW)

			// Set the window as the foreground window
			procSetForegroundWindow.Call(uintptr(hwnd))

			// Detach the threads
			procAttachThreadInput.Call(fgThreadID, targetThreadID, 0)

			found = true
			return 0 // Stop enumeration
		}
		return 1 // Continue enumeration
	})

	// Enumerate the windows
	ret, _, err := procEnumWindows.Call(enumWindowsCallback, 0)
	if ret == 0 && !found {
		return fmt.Errorf("aucune fenêtre trouvée avec le mot-clé: %s", keyword)
	}
	if err != syscall.Errno(0) {
		return fmt.Errorf("erreur lors de l'énumération des fenêtres: %v", err)
	}

	if !found {
		return fmt.Errorf("aucune fenêtre trouvée avec le mot-clé: %s", keyword)
	}

	log.Println("Fenêtre mise en avant avec succès")
	return nil
}

// FindWindowByPartialTitle cherche une fenêtre qui contient un mot-clé dans son titre.
func (ws *WindowService) FindWindowByPartialTitle(partialTitle string) (windows.Handle, error) {
	var hwndFound windows.Handle
	enumWindowsCallback := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		var buf [256]uint16
		procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		windowText := syscall.UTF16ToString(buf[:])

		if strings.Contains(windowText, partialTitle) {
			hwndFound = windows.Handle(hwnd) // Conversion explicite
			return 0                         // Stop enumeration, car la fenêtre est trouvée
		}
		return 1 // Continue enumeration
	})

	ret, _, err := procEnumWindows.Call(enumWindowsCallback, 0)
	if ret == 0 && hwndFound == 0 { // L'appel système a échoué ou aucune fenêtre n'a été trouvée
		return 0, fmt.Errorf("erreur lors de l'énumération des fenêtres: %v", err)
	}

	if hwndFound == 0 { // Si aucune fenêtre n'est trouvée
		return 0, windows.ERROR_NOT_FOUND
	}

	return hwndFound, nil
}

// GetForegroundWindow retrieves the handle of the window currently in the foreground.
func (ws *WindowService) GetForegroundWindow() syscall.Handle {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return syscall.Handle(hwnd)
}

// GetWindowText retrieves the text of the specified window by its handle.
func (ws *WindowService) GetWindowText(hwnd syscall.Handle) string {
	var buf [256]uint16
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:])
}

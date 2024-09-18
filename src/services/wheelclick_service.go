package services

import (
	"log"
	"strings"
	"syscall"
	"time"
	"unsafe"

	hook "github.com/robotn/gohook"
)

const (
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	VK_MBUTTON     = 0x04 // Virtual key code for middle mouse button
)

type WheelClickService struct {
	stopChan chan struct{}
}

func GetCursorPos() (int, int) {
	var pt Point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}

func ConvertScreenToClient(hWnd syscall.Handle, x, y int) (int, int) {
	var point struct {
		X, Y int32
	}
	point.X = int32(x)
	point.Y = int32(y)

	ret, _, _ := procScreenToClient.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&point)))
	if ret == 0 {
		log.Printf("Erreur lors de la conversion des coordonnées écran en coordonnées client")
	}

	return int(point.X), int(point.Y)
}

func SimulateClick(hWnd syscall.Handle, x, y int) {
	// Mettre la fenêtre au premier plan
	procSetForegroundWindow.Call(uintptr(hWnd))

	// Convertir les coordonnées d'écran en coordonnées client
	clientX, clientY := ConvertScreenToClient(hWnd, x, y)

	// Journaliser les coordonnées utilisées pour définir la position du curseur
	log.Printf("Définir la position du curseur à : X=%d, Y=%d", x, y)

	// Définir la position du curseur
	ret, _, _ := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		log.Printf("Échec de la définition de la position du curseur à : X=%d, Y=%d", x, y)
	}

	// Vérifier la position actuelle du curseur
	actualX, actualY := GetCursorPos()
	log.Printf("Position actuelle du curseur après la définition : X=%d, Y=%d", actualX, actualY)

	// Introduire un léger délai pour s'assurer que le curseur a bougé
	time.Sleep(50 * time.Millisecond)

	// Utiliser SendMessage pour simuler le clic aux coordonnées client
	log.Println("Envoi du message WM_LBUTTONDOWN")
	procSendMessage.Call(uintptr(hWnd), WM_LBUTTONDOWN, 0, uintptr(clientY<<16|clientX))
	time.Sleep(15 * time.Millisecond)
	log.Println("Envoi du message WM_LBUTTONUP")
	procSendMessage.Call(uintptr(hWnd), WM_LBUTTONUP, 0, uintptr(clientY<<16|clientX))
}

// Start detects middle mouse clicks.
func (wcs *WheelClickService) Start() {
	wcs.stopChan = make(chan struct{})
	go wcs.DetectMiddleClick()
}

// Stop stops detecting middle mouse clicks.
func (wcs *WheelClickService) Stop() {
	close(wcs.stopChan) // Signal to stop
}

// DetectMiddleClick listens for mouse button presses.
func (wcs *WheelClickService) DetectMiddleClick() {
	evChan := hook.Start()
	defer hook.End()

	for {
		select {
		case ev := <-evChan:
			if ev.Kind == hook.MouseDown && ev.Button == 3 { // Check for middle mouse button
				x, y := ev.X, ev.Y
				log.Printf("Middle click detected at: (%d, %d)", x, y)
				wcs.SendClickToDofusWindows(int(x), int(y)) // Convert to int
			}
		case <-wcs.stopChan:
			log.Println("Stopping middle click detection")
			return
		}
	}
}

// SendClickToDofusWindows finds all windows containing "Dofus" and sends a click.
func (wcs *WheelClickService) SendClickToDofusWindows(x, y int) {
	windowService := &WindowService{}
	windows, err := windowService.GetWindows()
	if err != nil {
		log.Printf("Error getting windows: %v", err)
		return
	}

	for _, windowTitle := range windows {
		if strings.Contains(windowTitle, "Dofus") {
			log.Printf("Sending click to window: %s", windowTitle)
			hWnd := GetWindowHandle(windowTitle)
			log.Printf("Clicking at: X=%d, Y=%d", x, y)
			SimulateClick(hWnd, x, y)
		}
	}
}

// GetWindowHandle retrieves the handle of the window with the given title.
func GetWindowHandle(title string) syscall.Handle {
	titleUTF16, _ := syscall.UTF16PtrFromString(title)
	hWnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titleUTF16)))
	return syscall.Handle(hWnd)
}

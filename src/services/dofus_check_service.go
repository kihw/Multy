package services

import (
	"log"
	"strings"
	"time"
)

// DofusCheckService monitors the Dofus window and manages services accordingly.
type DofusCheckService struct {
	windowService      *WindowService
	isWheelClickActive bool
	isShortcutActive   bool
	stopChan           chan bool
}

// NewDofusCheckService creates a new instance of the DofusCheckService.
func NewDofusCheckService(windowService *WindowService) *DofusCheckService {
	return &DofusCheckService{
		windowService: windowService,
		stopChan:      make(chan bool),
	}
}

// StartMonitoring begins monitoring the Dofus window's state.
func (dcs *DofusCheckService) StartMonitoring() {
	go func() {
		for {
			select {
			case <-dcs.stopChan:
				log.Println("DofusCheckService stopped.")
				return
			default:
				dcs.checkDofusWindow()
				time.Sleep(1 * time.Second) // Poll every second
			}
		}
	}()
}

// StopMonitoring stops the window monitoring.
func (dcs *DofusCheckService) StopMonitoring() {
	dcs.stopChan <- true
}

// checkDofusWindow checks if Dofus is in the foreground and updates services accordingly.
func (dcs *DofusCheckService) checkDofusWindow() {
	hwnd := dcs.windowService.GetForegroundWindow()
	windowTitle := dcs.windowService.GetWindowText(hwnd)

	if strings.Contains(windowTitle, "Dofus") {
		// Dofus is in the foreground
		log.Println("Dofus window is active.")

		// Reactivate services if they were previously active
		if !dcs.isShortcutActive {
			dcs.activateShortcuts()
		}
		if dcs.isWheelClickActive {
			dcs.activateWheelClick()
		}
	} else {
		// Dofus is not in the foreground
		log.Println("Dofus window is not active.")

		// Deactivate services
		if dcs.isShortcutActive {
			dcs.deactivateShortcuts()
		}
		if dcs.isWheelClickActive {
			dcs.deactivateWheelClick()
		}
	}
}

// activateShortcuts re-enables shortcuts.
func (dcs *DofusCheckService) activateShortcuts() {
	log.Println("Activating shortcuts...")
	dcs.isShortcutActive = true
	// Call the existing shortcut service to register shortcuts
}

// deactivateShortcuts disables shortcuts.
func (dcs *DofusCheckService) deactivateShortcuts() {
	log.Println("Deactivating shortcuts...")
	dcs.isShortcutActive = false
	// Call the existing shortcut service to unregister shortcuts
}

// activateWheelClick restarts the wheel click monitoring if it was active.
func (dcs *DofusCheckService) activateWheelClick() {
	log.Println("Activating wheel click service...")
	dcs.isWheelClickActive = true
	// Call the existing wheel click service to start monitoring
}

// deactivateWheelClick stops the wheel click monitoring.
func (dcs *DofusCheckService) deactivateWheelClick() {
	log.Println("Deactivating wheel click service...")
	dcs.isWheelClickActive = false
	// Call the existing wheel click service to stop monitoring
}

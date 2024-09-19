package services

import (
	"fmt"
	"log"
	"strings"
	"syscall"
)

const (
	SW_RESTORE = 9
)

type WindowService struct{}

func (ws *WindowService) GetWindows() ([]string, error) {
	var windowsList []string

	enumFunc := func(hwnd syscall.Handle, lParam uintptr) uintptr {
		length := GetWindowTextLength(hwnd)
		if length > 0 {
			windowText := GetWindowText(hwnd)
			windowsList = append(windowsList, windowText)
		}
		return 1 // Continue enumeration
	}

	if err := EnumWindows(enumFunc, 0); err != nil && err.Error() != "The operation completed successfully." {
		return nil, fmt.Errorf("error enumerating windows: %v", err)
	}

	return windowsList, nil
}

func (ws *WindowService) FocusWindowWithTitle(keyword string) error {
	if keyword == "" {
		return fmt.Errorf("window title keyword is empty")
	}

	var found bool
	enumFunc := func(hwnd syscall.Handle, lParam uintptr) uintptr {
		length := GetWindowTextLength(hwnd)
		if length > 0 {
			windowText := GetWindowText(hwnd)

			if strings.Contains(windowText, keyword) {
				log.Printf("Found matching window: '%s'", windowText)

				// Restore the window if it is minimized
				if IsIconic(hwnd) {
					ShowWindow(hwnd, SW_RESTORE)
				}

				// Set the window to the foreground
				err := ws.setForegroundWindow(hwnd)
				if err != nil {
					log.Printf("Failed to set foreground window: %v", err)
				} else {
					log.Println("Window successfully brought to the foreground")
				}

				found = true
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue enumeration
	}

	if err := EnumWindows(enumFunc, 0); err != nil {
		return fmt.Errorf("error enumerating windows: %v", err)
	}

	if !found {
		return fmt.Errorf("no window found with keyword: %s", keyword)
	}

	return nil
}

func (ws *WindowService) setForegroundWindow(hwnd syscall.Handle) error {
	if SetForegroundWindow(hwnd) {
		return nil
	}

	fgWindow := GetForegroundWindow()
	if fgWindow == 0 {
		return fmt.Errorf("failed to retrieve the foreground window")
	}

	fgThreadID, _ := GetWindowThreadProcessId(fgWindow)
	targetThreadID, _ := GetWindowThreadProcessId(hwnd)

	if AttachThreadInput(fgThreadID, targetThreadID, true) {
		defer AttachThreadInput(fgThreadID, targetThreadID, false)
		if SetForegroundWindow(hwnd) {
			return nil
		}
	}

	return fmt.Errorf("failed to set window to the foreground")
}

func (ws *WindowService) FindWindowByPartialTitle(partialTitle string) (syscall.Handle, error) {
	var hwndFound syscall.Handle
	enumFunc := func(hwnd syscall.Handle, lParam uintptr) uintptr {
		length := GetWindowTextLength(hwnd)
		if length > 0 {
			windowText := GetWindowText(hwnd)

			if strings.Contains(windowText, partialTitle) {
				hwndFound = hwnd
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue enumeration
	}

	if err := EnumWindows(enumFunc, 0); err != nil {
		return 0, fmt.Errorf("error enumerating windows: %v", err)
	}

	if hwndFound == 0 {
		return 0, fmt.Errorf("no window found with title: %s", partialTitle)
	}

	return hwndFound, nil
}

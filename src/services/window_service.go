package services

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"strings"
	"syscall"
	"unsafe"
)

const (
	SW_RESTORE = 9
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
			procShowWindow.Call(uintptr(hwnd), SW_RESTORE)
			procSetForegroundWindow.Call(uintptr(hwnd))
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

	return nil
}

func (ws *WindowService) CaptureWindow(hwnd syscall.Handle) image.Image {
    user32 := syscall.NewLazyDLL("user32.dll")
    gdi32 := syscall.NewLazyDLL("gdi32.dll")

    getDC := user32.NewProc("GetDC")
    releaseDC := user32.NewProc("ReleaseDC")
    getClientRect := user32.NewProc("GetClientRect")

    createCompatibleDC := gdi32.NewProc("CreateCompatibleDC")
    createCompatibleBitmap := gdi32.NewProc("CreateCompatibleBitmap")
    selectObject := gdi32.NewProc("SelectObject")
    bitBlt := gdi32.NewProc("BitBlt")
    deleteObject := gdi32.NewProc("DeleteObject")

    // Obtenez le contexte de périphérique (DC) de la fenêtre
    hdcWindow, _, _ := getDC.Call(uintptr(hwnd))
    hdcMemDC, _, _ := createCompatibleDC.Call(hdcWindow)

    // Définissez les dimensions de l'image
    var rect RECT 
    getClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
    width, height := int(rect.Right-rect.Left), int(rect.Bottom-rect.Top)

    // Créez un bitmap compatible
    hBitmap, _, _ := createCompatibleBitmap.Call(hdcWindow, uintptr(width), uintptr(height))
    selectObject.Call(hdcMemDC, hBitmap)

    // Copiez l'écran dans le bitmap
    bitBlt.Call(
        hdcMemDC,
        0, 0,
        uintptr(width), uintptr(height),
        hdcWindow,
        0, 0,
        uintptr(0x00CC0020), // SRCCOPY
    )

    // Préparez une structure BITMAPINFO pour obtenir les pixels
    var bi BITMAPINFO
    bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
    bi.BmiHeader.BiWidth = int32(width)
    bi.BmiHeader.BiHeight = -int32(height) // top-down
    bi.BmiHeader.BiPlanes = 1
    bi.BmiHeader.BiBitCount = 32
    bi.BmiHeader.BiCompression = BI_RGB

    // Créez une image pour recevoir les pixels
    img := image.NewRGBA(image.Rect(0, 0, width, height))

    // Obtenez les pixels
    gdi32.NewProc("GetDIBits").Call(
        hdcMemDC,
        hBitmap,
        0, uintptr(height),
        uintptr(unsafe.Pointer(&img.Pix[0])),
        uintptr(unsafe.Pointer(&bi)),
        DIB_RGB_COLORS,
    )

    // Nettoyez les ressources
    deleteObject.Call(hBitmap)
    deleteObject.Call(hdcMemDC)
    releaseDC.Call(uintptr(hwnd), hdcWindow)

    return img
}

// BITMAPINFOHEADER and BITMAPINFO structures
type BITMAPINFOHEADER struct {
    BiSize          uint32
    BiWidth         int32
    BiHeight        int32
    BiPlanes        uint16
    BiBitCount      uint16
    BiCompression   uint32
    BiSizeImage     uint32
    BiXPelsPerMeter int32
    BiYPelsPerMeter int32
    BiClrUsed       uint32
    BiClrImportant  uint32
}

type BITMAPINFO struct {
    BmiHeader BITMAPINFOHEADER
    BmiColors [1]color.RGBA
}

const (
    BI_RGB         = 0
    DIB_RGB_COLORS = 0
)

type RECT struct {
    Left   int32
    Top    int32
    Right  int32
    Bottom int32
}
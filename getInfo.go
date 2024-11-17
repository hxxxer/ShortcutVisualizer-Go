package main

import (
	"fmt"
	"image"
	"image/color"
	"syscall"
	"unsafe"
)

var (
	shell32                = syscall.NewLazyDLL("shell32.dll")
	user32                 = syscall.NewLazyDLL("user32.dll")
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	procSHGetFileInfoW     = shell32.NewProc("SHGetFileInfoW")
	procGetIconInfo        = user32.NewProc("GetIconInfo")
	procGetDC              = user32.NewProc("GetDC")
	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procGetObject          = gdi32.NewProc("GetObjectW")
	procGetDIBits          = gdi32.NewProc("GetDIBits")
	procDeleteDC           = gdi32.NewProc("DeleteDC")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procReleaseDC          = user32.NewProc("ReleaseDC")
)

type SHFILEINFO struct {
	hIcon         uintptr
	iIcon         int32
	dwAttributes  uint32
	szDisplayName [260]uint16
	szTypeName    [80]uint16
}

type ICONINFO struct {
	fIcon    int32
	xHotspot int32
	yHotspot int32
	hbmMask  uintptr
	hbmColor uintptr
}

type BITMAP struct {
	bmType       int32
	bmWidth      int32
	bmHeight     int32
	bmWidthBytes int32
	bmPlanes     uint16
	bmBitsPixel  uint16
	bmBits       uintptr
}

type BITMAPINFOHEADER struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

const (
	SHGFI_ICON      = 0x000000100
	SHGFI_LARGEICON = 0x000000000
	DIB_RGB_COLORS  = 0
)

// 获取文件大图标句柄
func GetFileIcon(filePath string) (uintptr, error) {
	var shfi SHFILEINFO
	pathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return 0, err
	}

	ret, _, err := procSHGetFileInfoW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&shfi)),
		unsafe.Sizeof(shfi),
		SHGFI_ICON|SHGFI_LARGEICON,
	)
	if ret == 0 {
		return 0, fmt.Errorf("failed to get icon")
	}
	return shfi.hIcon, nil
}

// 将图标句柄转换为image.Image
func IconToImage(hIcon uintptr) (image.Image, error) {
	var iconInfo ICONINFO
	ret, _, _ := procGetIconInfo.Call(hIcon, uintptr(unsafe.Pointer(&iconInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("GetIconInfo failed")
	}
	defer procDeleteObject.Call(iconInfo.hbmMask)
	defer procDeleteObject.Call(iconInfo.hbmColor)

	var bm BITMAP
	ret, _, _ = procGetObject.Call(
		iconInfo.hbmColor,
		uintptr(unsafe.Sizeof(bm)),
		uintptr(unsafe.Pointer(&bm)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetObject failed")
	}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return nil, fmt.Errorf("GetDC failed")
	}
	defer procReleaseDC.Call(0, hdc)

	hdcMem, _, _ := procCreateCompatibleDC.Call(hdc)
	if hdcMem == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(hdcMem)

	ret, _, _ = procSelectObject.Call(hdcMem, iconInfo.hbmColor)
	if ret == 0 {
		return nil, fmt.Errorf("SelectObject failed")
	}

	bi := BITMAPINFOHEADER{
		biSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		biWidth:       bm.bmWidth,
		biHeight:      bm.bmHeight,
		biPlanes:      1,
		biBitCount:    32,
		biCompression: 0,
	}

	pixels := make([]byte, bm.bmWidth*bm.bmHeight*4)

	ret, _, _ = procGetDIBits.Call(
		hdcMem,
		iconInfo.hbmColor,
		0,
		uintptr(bm.bmHeight),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bi)),
		DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed")
	}

	img := image.NewRGBA(image.Rect(0, 0, int(bm.bmWidth), int(bm.bmHeight)))
	for y := 0; y < int(bm.bmHeight); y++ {
		for x := 0; x < int(bm.bmWidth); x++ {
			i := (int(bm.bmHeight)-y-1)*int(bm.bmWidth)*4 + x*4
			img.Set(x, y, color.RGBA{
				B: pixels[i],
				G: pixels[i+1],
				R: pixels[i+2],
				A: pixels[i+3],
			})
		}
	}

	return img, nil
}

// 销毁图标
func DestroyIcon(hIcon uintptr) {
	if hIcon != 0 {
		user32.NewProc("DestroyIcon").Call(hIcon)
	}
}

func GetFileIcon2Image(filePath string) (image.Image, error) {
	// 获取图标句柄
	hIcon, err := GetFileIcon(filePath)
	if err != nil {
		fmt.Printf("Error getting icon: %v\n", err)
		return nil, err
	}
	defer DestroyIcon(hIcon)

	// 转换为image.Image
	iconImage, err := IconToImage(hIcon)
	if err != nil {
		fmt.Printf("Error converting icon to image: %v\n", err)
		return nil, err
	}

	return iconImage, nil
}

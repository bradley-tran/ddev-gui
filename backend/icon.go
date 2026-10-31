package backend

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
)

// DefaultWindowsICOSizes lists the square icon dimensions (in pixels) packed
// into a Windows .ico container so that both small (SM_CXSMICON) and large
// (SM_CXICON) icon lookups match an exact pre-filtered entry across standard
// and custom Windows DPI scale factors (100% through 400%).
var DefaultWindowsICOSizes = []int{
	16, 18, 20, 22, 24, 28, 32, 36, 40, 44, 48, 56, 64, 72, 80, 96, 128, 256,
}

// BuildWindowsICO decodes a high-resolution source PNG and builds a multi-size
// Windows .ico file whose entries are PNG-encoded images downsampled with
// premultiplied-alpha area averaging.
//
// Passing this .ico to Wails on Windows avoids Win32 CreateIconFromResourceEx
// nearest-neighbor downscaling from 512x512 to tray/window icon sizes.
func BuildWindowsICO(pngData []byte, sizes ...int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("decode source PNG: %w", err)
	}

	if len(sizes) == 0 {
		sizes = DefaultWindowsICOSizes
	}

	type encodedEntry struct {
		size int
		data []byte
	}
	entries := make([]encodedEntry, 0, len(sizes))
	for _, s := range sizes {
		if s <= 0 || s > 256 {
			return nil, fmt.Errorf("invalid ICO entry size %d (must be 1..256)", s)
		}
		scaled := resizeImageArea(src, s, s)
		var buf bytes.Buffer
		if err := png.Encode(&buf, scaled); err != nil {
			return nil, fmt.Errorf("encode %dx%d PNG: %w", s, s, err)
		}
		entries = append(entries, encodedEntry{size: s, data: buf.Bytes()})
	}

	const headerSize = 6
	const dirEntrySize = 16
	dataOffset := uint32(headerSize + len(entries)*dirEntrySize)

	var out bytes.Buffer
	// ICONDIR header: idReserved=0, idType=1 (ICO), idCount=len(entries)
	_ = binary.Write(&out, binary.LittleEndian, uint16(0))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(len(entries)))

	offset := dataOffset
	for _, e := range entries {
		var dimByte byte
		if e.size < 256 {
			dimByte = byte(e.size)
		}
		out.WriteByte(dimByte)                                           // bWidth (0 means 256)
		out.WriteByte(dimByte)                                           // bHeight (0 means 256)
		out.WriteByte(0)                                                 // bColorCount
		out.WriteByte(0)                                                 // bReserved
		_ = binary.Write(&out, binary.LittleEndian, uint16(1))           // wPlanes
		_ = binary.Write(&out, binary.LittleEndian, uint16(32))          // wBitCount
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(e.data))) // dwBytesInRes
		_ = binary.Write(&out, binary.LittleEndian, offset)              // dwImageOffset
		offset += uint32(len(e.data))
	}

	for _, e := range entries {
		out.Write(e.data)
	}
	return out.Bytes(), nil
}

// resizeImageArea resizes src to dstW x dstH using exact box/area integration
// in premultiplied-alpha space so transparent background pixels never bleed dark
// halos into anti-aliased edges.
func resizeImageArea(src image.Image, dstW, dstH int) *image.NRGBA {
	srcB := src.Bounds()
	srcW := srcB.Dx()
	srcH := srcB.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	if srcW == 0 || srcH == 0 || dstW <= 0 || dstH <= 0 {
		return dst
	}

	scaleX := float64(srcW) / float64(dstW)
	scaleY := float64(srcH) / float64(dstH)

	for dy := 0; dy < dstH; dy++ {
		sy0 := float64(dy) * scaleY
		sy1 := float64(dy+1) * scaleY
		iy0 := int(math.Floor(sy0))
		iy1 := int(math.Ceil(sy1))

		for dx := 0; dx < dstW; dx++ {
			sx0 := float64(dx) * scaleX
			sx1 := float64(dx+1) * scaleX
			ix0 := int(math.Floor(sx0))
			ix1 := int(math.Ceil(sx1))

			var sumR, sumG, sumB, sumA, totalWeight float64

			for sy := iy0; sy < iy1; sy++ {
				if sy < 0 || sy >= srcH {
					continue
				}
				wy := math.Min(sy1, float64(sy+1)) - math.Max(sy0, float64(sy))
				if wy <= 0 {
					continue
				}
				for sx := ix0; sx < ix1; sx++ {
					if sx < 0 || sx >= srcW {
						continue
					}
					wx := math.Min(sx1, float64(sx+1)) - math.Max(sx0, float64(sx))
					if wx <= 0 {
						continue
					}
					w := wx * wy
					r, g, b, a := src.At(srcB.Min.X+sx, srcB.Min.Y+sy).RGBA()
					sumR += float64(r) * w
					sumG += float64(g) * w
					sumB += float64(b) * w
					sumA += float64(a) * w
					totalWeight += w
				}
			}

			if totalWeight > 0 && sumA > 0 {
				avgA := sumA / totalWeight
				r8 := math.Min(255, math.Max(0, (sumR/totalWeight)/avgA*255.0))
				g8 := math.Min(255, math.Max(0, (sumG/totalWeight)/avgA*255.0))
				b8 := math.Min(255, math.Max(0, (sumB/totalWeight)/avgA*255.0))
				a8 := math.Min(255, math.Max(0, avgA/65535.0*255.0))
				dst.SetNRGBA(dx, dy, color.NRGBA{
					R: uint8(math.Round(r8)),
					G: uint8(math.Round(g8)),
					B: uint8(math.Round(b8)),
					A: uint8(math.Round(a8)),
				})
			}
		}
	}
	return dst
}

package backend

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildWindowsICO_PreservesEdgeColorAndAlpha(t *testing.T) {
	// Create a synthetic 256x256 white circle on transparent black background (0,0,0,0)
	const srcSize = 256
	src := image.NewNRGBA(image.Rect(0, 0, srcSize, srcSize))
	center := float64(srcSize) / 2.0
	radius := float64(srcSize)*0.45

	for y := 0; y < srcSize; y++ {
		for x := 0; x < srcSize; x++ {
			dx := float64(x) + 0.5 - center
			dy := float64(y) + 0.5 - center
			if math.Hypot(dx, dy) <= radius {
				src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}

	var srcBuf bytes.Buffer
	if err := png.Encode(&srcBuf, src); err != nil {
		t.Fatalf("encode synthetic PNG: %v", err)
	}

	sizes := []int{16, 20, 24, 32}
	icoData, err := BuildWindowsICO(srcBuf.Bytes(), sizes...)
	if err != nil {
		t.Fatalf("BuildWindowsICO failed: %v", err)
	}

	if len(icoData) < 6+len(sizes)*16 {
		t.Fatalf("ICO data too short: %d bytes", len(icoData))
	}
	if string(icoData[:4]) != "\x00\x00\x01\x00" {
		t.Fatalf("invalid ICO magic header: %x", icoData[:4])
	}
	count := int(binary.LittleEndian.Uint16(icoData[4:6]))
	if count != len(sizes) {
		t.Fatalf("expected %d ICO entries, got %d", len(sizes), count)
	}

	for i, wantSize := range sizes {
		entryOffset := 6 + i*16
		entry := icoData[entryOffset : entryOffset+16]
		gotW := int(entry[0])
		if gotW == 0 {
			gotW = 256
		}
		gotH := int(entry[1])
		if gotH == 0 {
			gotH = 256
		}
		if gotW != wantSize || gotH != wantSize {
			t.Fatalf("entry %d dimensions = %dx%d, want %dx%d", i, gotW, gotH, wantSize, wantSize)
		}

		dataSize := binary.LittleEndian.Uint32(entry[8:12])
		dataOffset := binary.LittleEndian.Uint32(entry[12:16])
		pngSlice := icoData[dataOffset : dataOffset+dataSize]

		decoded, err := png.Decode(bytes.NewReader(pngSlice))
		if err != nil {
			t.Fatalf("entry %d (%dx%d) failed to decode as PNG: %v", i, wantSize, wantSize, err)
		}
		if decoded.Bounds().Dx() != wantSize || decoded.Bounds().Dy() != wantSize {
			t.Fatalf("decoded PNG size = %v, want %dx%d", decoded.Bounds(), wantSize, wantSize)
		}

		// Verify anti-aliased edge pixels exist and do not bleed dark halo from (0,0,0,0)
		semiTransparentCount := 0
		for y := 0; y < wantSize; y++ {
			for x := 0; x < wantSize; x++ {
				c := color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA)
				if c.A > 0 && c.A < 255 {
					semiTransparentCount++
					if c.R < 250 || c.G < 250 || c.B < 250 {
						t.Fatalf("entry %dx%d pixel (%d,%d) has dark halo bleed: %+v", wantSize, wantSize, x, y, c)
					}
				}
			}
		}
		if semiTransparentCount == 0 {
			t.Fatalf("entry %dx%d has no anti-aliased edge pixels", wantSize, wantSize)
		}
	}
}

func TestWindowsIconFileHasAllDPISizes(t *testing.T) {
	icoPath := filepath.Join("..", "build", "windows", "icon.ico")
	data, err := os.ReadFile(icoPath)
	if err != nil {
		t.Fatalf("read %s: %v", icoPath, err)
	}
	if len(data) < 6 || string(data[:4]) != "\x00\x00\x01\x00" {
		t.Fatalf("invalid ICO header in %s", icoPath)
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	foundSizes := make(map[int]bool, count)
	for i := 0; i < count; i++ {
		entry := data[6+i*16 : 6+(i+1)*16]
		w := int(entry[0])
		if w == 0 {
			w = 256
		}
		dataSize := binary.LittleEndian.Uint32(entry[8:12])
		dataOffset := binary.LittleEndian.Uint32(entry[12:16])
		img, err := png.Decode(bytes.NewReader(data[dataOffset : dataOffset+dataSize]))
		if err != nil {
			t.Fatalf("decode entry %d (%dpx) in icon.ico: %v", i, w, err)
		}
		if img.Bounds().Dx() != w || img.Bounds().Dy() != w {
			t.Fatalf("entry %d bounds = %v, want %dx%d", i, img.Bounds(), w, w)
		}
		foundSizes[w] = true
	}

	for _, want := range DefaultWindowsICOSizes {
		if !foundSizes[want] {
			t.Errorf("build/windows/icon.ico missing %dx%d entry", want, want)
		}
	}
}

package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativedialog "github.com/sqweek/dialog"
)

const (
	dotCellSize                    = 8
	fontSourceInitialDisplayWidth  = 360
	fontSourceInitialDisplayHeight = 180
)

type FontChar struct {
	Char        string
	Width       int
	Height      int
	HexData     string
	WhitePixels int
	Bitmap      [][]bool
}

type CharCell struct {
	BBox   image.Rectangle
	Bitmap [][]bool
	Char   string
}

type fontAutoPreprocessResult struct {
	Foreground       color.NRGBA
	Tolerance        color.NRGBA
	Background       color.NRGBA
	CropRect         image.Rectangle
	ForegroundPixels int
	Binary           *image.NRGBA
}

type fontColorBucket struct {
	count int
	sumR  int
	sumG  int
	sumB  int
}

func fontColorHex(c color.NRGBA) string {
	return fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

func fontColorParam(foreground, tolerance color.NRGBA) string {
	return fontColorHex(foreground) + "-" + fontColorHex(tolerance)
}

func fontColorBucketKey(c color.NRGBA) int {
	return int(c.R/16)<<8 | int(c.G/16)<<4 | int(c.B/16)
}

func fontBucketAverage(b *fontColorBucket) color.NRGBA {
	if b == nil || b.count == 0 {
		return color.NRGBA{0, 0, 0, 255}
	}
	return color.NRGBA{
		R: uint8(b.sumR / b.count),
		G: uint8(b.sumG / b.count),
		B: uint8(b.sumB / b.count),
		A: 255,
	}
}

func fontColorDistanceSq(a, b color.NRGBA) int {
	dr := int(a.R) - int(b.R)
	dg := int(a.G) - int(b.G)
	db := int(a.B) - int(b.B)
	return dr*dr + dg*dg + db*db
}

func fontAbsDiffInt(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func fontClampByte(v, minV, maxV int) uint8 {
	if v < minV {
		v = minV
	}
	if v > maxV {
		v = maxV
	}
	return uint8(v)
}

func imageColorNRGBA(img image.Image, x, y int) color.NRGBA {
	r, g, b, a := img.At(x, y).RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func estimateFontForegroundColor(src image.Image) (background, foreground color.NRGBA, ok bool) {
	if src == nil || src.Bounds().Empty() {
		return color.NRGBA{}, color.NRGBA{}, false
	}

	bounds := src.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()
	buckets := make(map[int]*fontColorBucket)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := imageColorNRGBA(src, x, y)
			key := fontColorBucketKey(c)
			bucket := buckets[key]
			if bucket == nil {
				bucket = &fontColorBucket{}
				buckets[key] = bucket
			}
			bucket.count++
			bucket.sumR += int(c.R)
			bucket.sumG += int(c.G)
			bucket.sumB += int(c.B)
		}
	}

	var bgBucket *fontColorBucket
	for _, bucket := range buckets {
		if bgBucket == nil || bucket.count > bgBucket.count {
			bgBucket = bucket
		}
	}
	background = fontBucketAverage(bgBucket)

	minForegroundCount := max(1, totalPixels/10000)
	if totalPixels >= 200 {
		minForegroundCount = max(2, minForegroundCount)
	}

	bestScore := -1
	bestDistance := 0
	var fgBucket *fontColorBucket
	const minContrastSq = 28 * 28
	for _, bucket := range buckets {
		if bucket == bgBucket || bucket.count < minForegroundCount {
			continue
		}
		avg := fontBucketAverage(bucket)
		distance := fontColorDistanceSq(avg, background)
		if distance < minContrastSq {
			continue
		}
		scoreCount := min(bucket.count, 1000)
		score := distance * scoreCount
		if score > bestScore || (score == bestScore && distance > bestDistance) {
			bestScore = score
			bestDistance = distance
			fgBucket = bucket
		}
	}

	if fgBucket == nil {
		return background, color.NRGBA{}, false
	}
	return background, fontBucketAverage(fgBucket), true
}

func estimateFontTolerance(src image.Image, foreground, background color.NRGBA) color.NRGBA {
	if src == nil || src.Bounds().Empty() {
		return color.NRGBA{16, 16, 16, 255}
	}

	bounds := src.Bounds()
	maxR, maxG, maxB := 0, 0, 0
	matched := 0
	const maxForegroundDistanceSq = 96 * 96 * 3
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := imageColorNRGBA(src, x, y)
			fgDistance := fontColorDistanceSq(c, foreground)
			if fgDistance > maxForegroundDistanceSq || fgDistance > fontColorDistanceSq(c, background) {
				continue
			}
			maxR = max(maxR, fontAbsDiffInt(c.R, foreground.R))
			maxG = max(maxG, fontAbsDiffInt(c.G, foreground.G))
			maxB = max(maxB, fontAbsDiffInt(c.B, foreground.B))
			matched++
		}
	}
	if matched == 0 {
		return color.NRGBA{16, 16, 16, 255}
	}

	return color.NRGBA{
		R: fontClampByte(maxR+8, 16, 80),
		G: fontClampByte(maxG+8, 16, 80),
		B: fontClampByte(maxB+8, 16, 80),
		A: 255,
	}
}

func fontForegroundBounds(binaryImg *image.NRGBA) (image.Rectangle, int, bool) {
	if binaryImg == nil || binaryImg.Bounds().Empty() {
		return image.Rectangle{}, 0, false
	}
	bounds := binaryImg.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if binaryImg.NRGBAAt(x, y).G <= 128 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
			count++
		}
	}
	if count == 0 {
		return image.Rectangle{}, 0, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), count, true
}

func expandRectWithin(rect, bounds image.Rectangle, pad int) image.Rectangle {
	return image.Rect(
		max(bounds.Min.X, rect.Min.X-pad),
		max(bounds.Min.Y, rect.Min.Y-pad),
		min(bounds.Max.X, rect.Max.X+pad),
		min(bounds.Max.Y, rect.Max.Y+pad),
	)
}

func autoPreprocessFontImage(src image.Image) (fontAutoPreprocessResult, bool) {
	background, foreground, ok := estimateFontForegroundColor(src)
	if !ok {
		return fontAutoPreprocessResult{}, false
	}
	tolerance := estimateFontTolerance(src, foreground, background)
	binary := createBinaryPreview(src, fontColorParam(foreground, tolerance))
	cropRect, foregroundPixels, ok := fontForegroundBounds(binary)
	if !ok {
		return fontAutoPreprocessResult{}, false
	}
	cropRect = expandRectWithin(cropRect, binary.Bounds(), 2)
	return fontAutoPreprocessResult{
		Foreground:       foreground,
		Tolerance:        tolerance,
		Background:       background,
		CropRect:         cropRect,
		ForegroundPixels: foregroundPixels,
		Binary:           binary,
	}, true
}

// ==================== 二值化 ====================

func parseColorWithOffset(colorStr string) (color.NRGBA, color.NRGBA) {
	arr := strings.Split(colorStr, "-")
	s := strings.TrimPrefix(arr[0], "#")
	if len(s) < 6 {
		s = "FFFFFF"
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	baseColor := color.NRGBA{uint8(r), uint8(g), uint8(b), 255}
	var offsetColor color.NRGBA
	if len(arr) > 1 {
		s2 := strings.TrimPrefix(arr[1], "#")
		if len(s2) >= 6 {
			or, _ := strconv.ParseUint(s2[0:2], 16, 8)
			og, _ := strconv.ParseUint(s2[2:4], 16, 8)
			ob, _ := strconv.ParseUint(s2[4:6], 16, 8)
			offsetColor = color.NRGBA{uint8(or), uint8(og), uint8(ob), 255}
		}
	}
	return baseColor, offsetColor
}

func colorAbsDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func isColorMatch(c1, base, offset color.NRGBA) bool {
	return colorAbsDiff(c1.R, base.R) <= offset.R &&
		colorAbsDiff(c1.G, base.G) <= offset.G &&
		colorAbsDiff(c1.B, base.B) <= offset.B
}

func createBinaryPreview(src image.Image, colorStr string) *image.NRGBA {
	bounds := src.Bounds()
	binary := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	base, offset := parseColorWithOffset(colorStr)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			pixel := color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
			if isColorMatch(pixel, base, offset) {
				binary.SetNRGBA(x-bounds.Min.X, y-bounds.Min.Y, color.NRGBA{0, 230, 0, 255})
			} else {
				binary.SetNRGBA(x-bounds.Min.X, y-bounds.Min.Y, color.NRGBA{15, 15, 15, 255})
			}
		}
	}
	return binary
}

func extractBitmapFromRect(binaryImg *image.NRGBA, rect image.Rectangle) [][]bool {
	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return nil
	}
	bitmap := make([][]bool, h)
	for y := 0; y < h; y++ {
		bitmap[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			bitmap[y][x] = binaryImg.NRGBAAt(rect.Min.X+x, rect.Min.Y+y).G > 128
		}
	}
	return bitmap
}

// ==================== 投影分割 ====================

func findCharacterBBoxes(binaryImg *image.NRGBA, colGap, rowGap int) []image.Rectangle {
	w := binaryImg.Bounds().Dx()
	h := binaryImg.Bounds().Dy()

	// Step 1: flood fill (4-connected) to find basic connected components
	visited := make([]bool, w*h)
	var compBB []image.Rectangle

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if visited[idx] || binaryImg.NRGBAAt(x, y).G <= 128 {
				continue
			}
			mnX, mnY, mxX, mxY := x, y, x, y
			stack := [][2]int{{x, y}}
			visited[idx] = true
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				px, py := p[0], p[1]
				if px < mnX {
					mnX = px
				}
				if px > mxX {
					mxX = px
				}
				if py < mnY {
					mnY = py
				}
				if py > mxY {
					mxY = py
				}
				for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					nx, ny := px+d[0], py+d[1]
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					ni := ny*w + nx
					if visited[ni] || binaryImg.NRGBAAt(nx, ny).G <= 128 {
						continue
					}
					visited[ni] = true
					stack = append(stack, [2]int{nx, ny})
				}
			}
			compBB = append(compBB, image.Rect(mnX, mnY, mxX+1, mxY+1))
		}
	}

	if len(compBB) == 0 {
		return nil
	}

	// Step 2: union-find + iterative proximity merge
	parent := make([]int, len(compBB))
	bbox := make([]image.Rectangle, len(compBB))
	for i := range compBB {
		parent[i] = i
		bbox[i] = compBB[i]
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}

	for changed := true; changed; {
		changed = false
		for i := 0; i < len(compBB); i++ {
			for j := i + 1; j < len(compBB); j++ {
				ri, rj := find(i), find(j)
				if ri == rj {
					continue
				}
				bi, bj := bbox[ri], bbox[rj]
				hGap := max(0, max(bi.Min.X, bj.Min.X)-min(bi.Max.X, bj.Max.X))
				vGap := max(0, max(bi.Min.Y, bj.Min.Y)-min(bi.Max.Y, bj.Max.Y))
				if hGap <= colGap && vGap <= rowGap {
					parent[ri] = rj
					bbox[rj] = bi.Union(bj)
					changed = true
				}
			}
		}
	}

	// Step 3: collect and sort
	seen := make(map[int]bool)
	var result []image.Rectangle
	for i := range compBB {
		root := find(i)
		if !seen[root] {
			seen[root] = true
			result = append(result, bbox[root])
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Min.X != result[j].Min.X {
			return result[i].Min.X < result[j].Min.X
		}
		return result[i].Min.Y < result[j].Min.Y
	})
	return result
}

// ==================== 点阵编码/解码 ====================

func encodeBitmapHex(bitmap [][]bool) (string, int) {
	height := len(bitmap)
	if height == 0 {
		return "", 0
	}
	width := len(bitmap[0])
	total := width * height
	numBytes := (total + 7) / 8
	byteArr := make([]byte, numBytes)
	whiteCount := 0
	idx := 0
	for _, row := range bitmap {
		for _, px := range row {
			if px {
				byteArr[idx/8] |= 1 << (idx % 8)
				whiteCount++
			}
			idx++
		}
	}
	return fmt.Sprintf("%x", byteArr), whiteCount
}

func decodeBitmapHex(hexData string, width, height int) [][]bool {
	byteArr := make([]byte, len(hexData)/2)
	for i := 0; i < len(hexData)-1; i += 2 {
		b, _ := strconv.ParseUint(hexData[i:i+2], 16, 8)
		byteArr[i/2] = byte(b)
	}
	bitmap := make([][]bool, height)
	idx := 0
	for y := 0; y < height; y++ {
		bitmap[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			if idx/8 < len(byteArr) {
				bitmap[y][x] = (byteArr[idx/8]>>(idx%8))&1 == 1
			}
			idx++
		}
	}
	return bitmap
}

func trimBitmapBounds(bitmap [][]bool) [][]bool {
	if len(bitmap) == 0 || len(bitmap[0]) == 0 {
		return nil
	}
	minX, minY := len(bitmap[0]), len(bitmap)
	maxX, maxY := -1, -1
	for y, row := range bitmap {
		for x, px := range row {
			if !px {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if maxX < minX || maxY < minY {
		return nil
	}

	trimmed := make([][]bool, maxY-minY+1)
	for y := minY; y <= maxY; y++ {
		row := make([]bool, maxX-minX+1)
		copy(row, bitmap[y][minX:maxX+1])
		trimmed[y-minY] = row
	}
	return trimmed
}

func fontBitmapMatchKey(bitmap [][]bool) string {
	trimmed := trimBitmapBounds(bitmap)
	if len(trimmed) == 0 || len(trimmed[0]) == 0 {
		return ""
	}
	hexData, _ := encodeBitmapHex(trimmed)
	return fmt.Sprintf("%dx%d:%s", len(trimmed[0]), len(trimmed), strings.ToLower(hexData))
}

func fontCharMatchKey(ch FontChar) string {
	if key := fontBitmapMatchKey(ch.Bitmap); key != "" {
		return key
	}
	if ch.Width <= 0 || ch.Height <= 0 || strings.TrimSpace(ch.HexData) == "" {
		return ""
	}
	return fontBitmapMatchKey(decodeBitmapHex(strings.TrimSpace(ch.HexData), ch.Width, ch.Height))
}

// ==================== 字库文件读写 ====================

func parseFontLib(content string) []FontChar {
	var chars []FontChar
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "$")
		if len(parts) < 5 {
			continue
		}
		charName := parts[0]
		hexData := parts[1]
		infoStr := parts[3]
		widthStr := parts[4]
		w, e1 := strconv.Atoi(widthStr)
		if e1 != nil || w <= 0 {
			continue
		}
		totalPixels := len(hexData) / 2 * 8
		h := totalPixels / w
		if h <= 0 {
			continue
		}
		wp := 0
		if dotIdx := strings.LastIndex(infoStr, "."); dotIdx >= 0 {
			wp, _ = strconv.Atoi(infoStr[dotIdx+1:])
		}
		bm := decodeBitmapHex(hexData, w, h)
		chars = append(chars, FontChar{
			Char: charName, Width: w, Height: h,
			HexData: hexData, WhitePixels: wp, Bitmap: bm,
		})
	}
	return chars
}

func exportFontLib(chars []FontChar, fgColor string) string {
	var sb strings.Builder
	sb.WriteString("# AutoGo FontLib v1.0\n")
	sb.WriteString(fmt.Sprintf("# Color:%s\n#\n", fgColor))
	for _, ch := range chars {
		sb.WriteString(fmt.Sprintf("%s$%s$1$0.0.%d$%d\n", ch.Char, ch.HexData, ch.WhitePixels, ch.Width))
	}
	return sb.String()
}

// ==================== 渲染 ====================

func fillGridPattern(img *image.NRGBA) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	stride := dotCellSize + 1
	gridColor := color.NRGBA{40, 40, 40, 255}
	cellColor := color.NRGBA{15, 15, 15, 255}
	for y := 0; y < h; y++ {
		idx := y * img.Stride
		for x := 0; x < w; x++ {
			var c color.NRGBA
			if x%stride == 0 || y%stride == 0 {
				c = gridColor
			} else {
				c = cellColor
			}
			img.Pix[idx] = c.R
			img.Pix[idx+1] = c.G
			img.Pix[idx+2] = c.B
			img.Pix[idx+3] = c.A
			idx += 4
		}
	}
}

func renderOriginalDotMatrix(src image.Image) *image.NRGBA {
	b := src.Bounds()
	pw, ph := b.Dx(), b.Dy()
	stride := dotCellSize + 1
	dispW := pw*stride + 1
	dispH := ph*stride + 1
	display := image.NewNRGBA(image.Rect(0, 0, dispW, dispH))
	fillGridPattern(display)
	for py := 0; py < ph; py++ {
		for px := 0; px < pw; px++ {
			r, g, bl, a := src.At(b.Min.X+px, b.Min.Y+py).RGBA()
			c := color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)}
			cx := px*stride + 1
			cy := py*stride + 1
			for dy := 0; dy < dotCellSize; dy++ {
				idx := (cy+dy)*display.Stride + cx*4
				for dx := 0; dx < dotCellSize; dx++ {
					display.Pix[idx] = c.R
					display.Pix[idx+1] = c.G
					display.Pix[idx+2] = c.B
					display.Pix[idx+3] = c.A
					idx += 4
				}
			}
		}
	}
	return display
}

func renderDotMatrix(binaryImg *image.NRGBA, colGap, rowGap int) (*image.NRGBA, []CharCell) {
	b := binaryImg.Bounds()
	pw, ph := b.Dx(), b.Dy()
	stride := dotCellSize + 1
	dispW := pw*stride + 1
	dispH := ph*stride + 1

	display := image.NewNRGBA(image.Rect(0, 0, dispW, dispH))

	fgColor := color.NRGBA{0, 230, 0, 255}
	for py := 0; py < ph; py++ {
		for px := 0; px < pw; px++ {
			if binaryImg.NRGBAAt(px, py).G > 128 {
				cx := px*stride + 1
				cy := py*stride + 1
				for dy := 0; dy < dotCellSize; dy++ {
					idx := (cy+dy)*display.Stride + cx*4
					for dx := 0; dx < dotCellSize; dx++ {
						display.Pix[idx] = fgColor.R
						display.Pix[idx+1] = fgColor.G
						display.Pix[idx+2] = fgColor.B
						display.Pix[idx+3] = fgColor.A
						idx += 4
					}
				}
			}
		}
	}

	bboxes := findCharacterBBoxes(binaryImg, colGap, rowGap)
	rectColor := color.NRGBA{255, 50, 0, 255}
	var cells []CharCell
	for _, bbox := range bboxes {
		dx1 := bbox.Min.X * stride
		dy1 := bbox.Min.Y * stride
		dx2 := bbox.Max.X * stride
		dy2 := bbox.Max.Y * stride
		drawRectOutlineOnImg(display, dx1, dy1, dx2, dy2, rectColor, 2)
		cells = append(cells, CharCell{BBox: bbox, Bitmap: extractBitmapFromRect(binaryImg, bbox)})
	}
	return display, cells
}

func drawRectOutlineOnImg(img *image.NRGBA, x1, y1, x2, y2 int, c color.NRGBA, thickness int) {
	bx, by := img.Bounds().Dx(), img.Bounds().Dy()
	for t := 0; t < thickness; t++ {
		for x := x1; x <= x2; x++ {
			if x >= 0 && x < bx {
				if y1+t >= 0 && y1+t < by {
					img.SetNRGBA(x, y1+t, c)
				}
				if y2-t >= 0 && y2-t < by {
					img.SetNRGBA(x, y2-t, c)
				}
			}
		}
		for y := y1; y <= y2; y++ {
			if y >= 0 && y < by {
				if x1+t >= 0 && x1+t < bx {
					img.SetNRGBA(x1+t, y, c)
				}
				if x2-t >= 0 && x2-t < bx {
					img.SetNRGBA(x2-t, y, c)
				}
			}
		}
	}
}

func createCharPreview(bitmap [][]bool, cellSize int) image.Image {
	if len(bitmap) == 0 || len(bitmap[0]) == 0 {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}
	h, w := len(bitmap), len(bitmap[0])
	img := image.NewNRGBA(image.Rect(0, 0, w*cellSize, h*cellSize))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var c color.NRGBA
			if bitmap[y][x] {
				c = color.NRGBA{0, 230, 0, 255}
			} else {
				c = color.NRGBA{25, 25, 25, 255}
			}
			for dy := 0; dy < cellSize; dy++ {
				for dx := 0; dx < cellSize; dx++ {
					img.SetNRGBA(x*cellSize+dx, y*cellSize+dy, c)
				}
			}
		}
	}
	return img
}

// ==================== 悬停容器 ====================

func createGridBackground(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	gridColor := color.NRGBA{40, 40, 40, 255}
	cellColor := color.NRGBA{15, 15, 15, 255}
	stride := dotCellSize + 1
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x%stride == 0 || y%stride == 0 {
				img.SetNRGBA(x, y, gridColor)
			} else {
				img.SetNRGBA(x, y, cellColor)
			}
		}
	}
	return img
}

// ==================== 动态网格背景 ====================

type gridBgWidget struct {
	widget.BaseWidget
}

func newGridBgWidget() *gridBgWidget {
	g := &gridBgWidget{}
	g.ExtendBaseWidget(g)
	return g
}

func (g *gridBgWidget) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromImage(nil)
	img.ScaleMode = canvas.ImageScalePixels
	img.FillMode = canvas.ImageFillOriginal
	return &gridBgRenderer{img: img}
}

type gridBgRenderer struct {
	img          *canvas.Image
	lastW, lastH int
}

func (r *gridBgRenderer) Layout(size fyne.Size) {
	w := int(size.Width)
	h := int(size.Height)
	if w > 0 && h > 0 && (w != r.lastW || h != r.lastH) {
		r.lastW = w
		r.lastH = h
		gridImg := image.NewNRGBA(image.Rect(0, 0, w, h))
		fillGridPattern(gridImg)
		r.img.Image = gridImg
		r.img.Refresh()
	}
	r.img.Resize(size)
	r.img.Move(fyne.NewPos(0, 0))
}

func (r *gridBgRenderer) MinSize() fyne.Size           { return fyne.NewSize(1, 1) }
func (r *gridBgRenderer) Refresh()                     {}
func (r *gridBgRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.img} }
func (r *gridBgRenderer) Destroy()                     {}

type tappableArea struct {
	widget.BaseWidget
	onTap func(fyne.Position)
}

func newTappableArea(onTap func(fyne.Position)) *tappableArea {
	t := &tappableArea{onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableArea) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(r)
}

func (t *tappableArea) Tapped(e *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap(e.Position)
	}
}

type hoverContainer struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onEnter func()
	onLeave func()
}

func newHoverContainer(content fyne.CanvasObject, onEnter, onLeave func()) *hoverContainer {
	h := &hoverContainer{content: content, onEnter: onEnter, onLeave: onLeave}
	h.ExtendBaseWidget(h)
	return h
}

func (h *hoverContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *hoverContainer) MouseIn(*desktop.MouseEvent) {
	if h.onEnter != nil {
		h.onEnter()
	}
}

func (h *hoverContainer) MouseMoved(*desktop.MouseEvent) {}

func (h *hoverContainer) MouseOut() {
	if h.onLeave != nil {
		h.onLeave()
	}
}

type fontImageDragMode int

const (
	fontImageDragNone fontImageDragMode = iota
	fontImageDragSelect
	fontImageDragPan
)

type fontImageViewer struct {
	widget.BaseWidget
	image              image.Image
	zoom               float32
	scroll             *container.Scroll
	selection          image.Rectangle
	tempSelection      image.Rectangle
	dragStart          image.Point
	dragCurrent        image.Point
	dragMode           fontImageDragMode
	lastDragAbs        fyne.Position
	selectionMode      bool
	magnifier          *MagnifierWidget
	onSelectionChanged func(image.Rectangle)
	onColorPicked      func(color.NRGBA)
}

func newFontImageViewer() *fontImageViewer {
	v := &fontImageViewer{zoom: 1}
	v.ExtendBaseWidget(v)
	return v
}

func (v *fontImageViewer) SetScroll(scroll *container.Scroll) {
	v.scroll = scroll
}

func (v *fontImageViewer) SetMagnifier(magnifier *MagnifierWidget) {
	v.magnifier = magnifier
}

func fontSourceInitialZoom(bounds image.Rectangle) float32 {
	if bounds.Empty() || bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return 1
	}
	if bounds.Dx() >= fontSourceInitialDisplayWidth || bounds.Dy() >= fontSourceInitialDisplayHeight {
		return 1
	}
	zoomX := float32(fontSourceInitialDisplayWidth) / float32(bounds.Dx())
	zoomY := float32(fontSourceInitialDisplayHeight) / float32(bounds.Dy())
	zoom := zoomX
	if zoomY > zoom {
		zoom = zoomY
	}
	if zoom < 1 {
		return 1
	}
	if zoom > maxImageZoom {
		return maxImageZoom
	}
	return zoom
}

func fontPreviewZoomForSourceZoom(sourceZoom float32) float32 {
	if sourceZoom <= 0 {
		sourceZoom = 1
	}
	zoom := sourceZoom / float32(dotCellSize+1)
	if zoom < minImageZoom {
		return minImageZoom
	}
	if zoom > maxImageZoom {
		return maxImageZoom
	}
	return zoom
}

func (v *fontImageViewer) SetImage(img image.Image) {
	if v.magnifier != nil {
		v.magnifier.Hide()
	}
	if img != nil {
		img = openCVImageToNRGBA(img)
	}
	v.image = img
	v.zoom = 1
	v.selection = image.Rectangle{}
	v.tempSelection = image.Rectangle{}
	v.dragMode = fontImageDragNone
	v.selectionMode = false
	v.Refresh()
	if v.scroll != nil {
		v.scroll.Refresh()
		v.scroll.ScrollToOffset(fyne.NewPos(0, 0))
	}
	if v.onSelectionChanged != nil {
		v.onSelectionChanged(image.Rectangle{})
	}
}

func (v *fontImageViewer) hideMagnifier() {
	if v.magnifier != nil {
		v.magnifier.Hide()
	}
}

func (v *fontImageViewer) updateMagnifier(pos fyne.Position) {
	if !magnifierEnabled || v.magnifier == nil || v.image == nil || v.scroll == nil || v.dragMode != fontImageDragNone {
		v.hideMagnifier()
		return
	}
	p, ok := v.imagePosition(pos)
	if !ok {
		v.hideMagnifier()
		return
	}
	viewPos := pos.Subtract(v.scroll.Offset)
	v.magnifier.Update(v.image, p.X, p.Y, viewPos.X, viewPos.Y)
}

func (v *fontImageViewer) updateMagnifierFromViewport(pos fyne.Position) {
	contentPos := pos
	if v.scroll != nil {
		contentPos = pos.Add(v.scroll.Offset)
	}
	p, ok := v.imagePosition(contentPos)
	if !ok {
		v.hideMagnifier()
		return
	}
	if !magnifierEnabled || v.magnifier == nil || v.image == nil || v.dragMode != fontImageDragNone {
		v.hideMagnifier()
		return
	}
	v.magnifier.Update(v.image, p.X, p.Y, pos.X, pos.Y)
}

func (v *fontImageViewer) ClearSelection() {
	if v.selection.Empty() && v.tempSelection.Empty() && v.dragMode == fontImageDragNone {
		return
	}
	v.selection = image.Rectangle{}
	v.tempSelection = image.Rectangle{}
	v.dragMode = fontImageDragNone
	v.Refresh()
	if v.onSelectionChanged != nil {
		v.onSelectionChanged(image.Rectangle{})
	}
}

func (v *fontImageViewer) StartSelectionMode() {
	v.selectionMode = true
	v.selection = image.Rectangle{}
	v.tempSelection = image.Rectangle{}
	v.dragMode = fontImageDragNone
	v.Refresh()
	if v.onSelectionChanged != nil {
		v.onSelectionChanged(image.Rectangle{})
	}
}

func (v *fontImageViewer) CancelSelectionMode() {
	v.selectionMode = false
	v.ClearSelection()
}

func (v *fontImageViewer) SelectionMode() bool {
	return v.selectionMode
}

func (v *fontImageViewer) SelectedRect() (image.Rectangle, bool) {
	if v.image == nil || v.selection.Empty() {
		return image.Rectangle{}, false
	}
	rect := v.selection.Intersect(v.image.Bounds())
	if rect.Empty() {
		return image.Rectangle{}, false
	}
	return rect, true
}

func (v *fontImageViewer) SetZoom(scale float32) {
	if v.image == nil {
		return
	}
	if scale < minImageZoom {
		scale = minImageZoom
	}
	if scale > maxImageZoom {
		scale = maxImageZoom
	}
	if scale == v.zoom {
		return
	}
	v.zoom = scale
	v.Refresh()
	if v.scroll != nil {
		v.scroll.Refresh()
	}
}

func (v *fontImageViewer) ResetZoom() {
	v.SetZoom(1)
	if v.scroll != nil {
		v.scroll.ScrollToOffset(fyne.NewPos(0, 0))
	}
}

func (v *fontImageViewer) zoomAt(pos fyne.Position, zoomIn bool) {
	if v.image == nil {
		return
	}

	oldScale := v.zoom
	if oldScale <= 0 {
		oldScale = 1
	}
	newScale := oldScale / zoomStepMultiplier
	if zoomIn {
		newScale = oldScale * zoomStepMultiplier
	}
	if newScale < minImageZoom {
		newScale = minImageZoom
	}
	if newScale > maxImageZoom {
		newScale = maxImageZoom
	}
	if newScale == oldScale {
		return
	}

	oldW, oldH := v.displayPixelSizeForZoom(oldScale)
	bounds := v.image.Bounds()
	contentPos := pos
	if v.scroll != nil {
		contentPos = pos.Add(v.scroll.Offset)
	}
	if contentPos.X < 0 {
		contentPos.X = 0
	} else if contentPos.X > float32(oldW) {
		contentPos.X = float32(oldW)
	}
	if contentPos.Y < 0 {
		contentPos.Y = 0
	} else if contentPos.Y > float32(oldH) {
		contentPos.Y = float32(oldH)
	}
	imageX := contentPos.X * float32(bounds.Dx()) / float32(oldW)
	imageY := contentPos.Y * float32(bounds.Dy()) / float32(oldH)

	v.zoom = newScale
	v.Refresh()
	if v.scroll != nil {
		newW, newH := v.displayPixelSizeForZoom(newScale)
		newContentX := imageX * float32(newW) / float32(bounds.Dx())
		newContentY := imageY * float32(newH) / float32(bounds.Dy())
		v.scroll.Refresh()
		v.scroll.ScrollToOffset(fyne.NewPos(newContentX-pos.X, newContentY-pos.Y))
	}
}

func (v *fontImageViewer) imagePosition(pos fyne.Position) (image.Point, bool) {
	if v.image == nil {
		return image.Point{}, false
	}
	bounds := v.image.Bounds()
	displayW, displayH := v.displayPixelSize()
	if pos.X < 0 || pos.Y < 0 || pos.X >= float32(displayW) || pos.Y >= float32(displayH) {
		return image.Point{}, false
	}
	x := bounds.Min.X + int(pos.X*float32(bounds.Dx())/float32(displayW))
	y := bounds.Min.Y + int(pos.Y*float32(bounds.Dy())/float32(displayH))
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return image.Point{}, false
	}
	return image.Pt(x, y), true
}

func (v *fontImageViewer) normalizedDragRect() image.Rectangle {
	if v.image == nil {
		return image.Rectangle{}
	}
	x0 := min(v.dragStart.X, v.dragCurrent.X)
	y0 := min(v.dragStart.Y, v.dragCurrent.Y)
	x1 := max(v.dragStart.X, v.dragCurrent.X) + 1
	y1 := max(v.dragStart.Y, v.dragCurrent.Y) + 1
	return image.Rect(x0, y0, x1, y1).Intersect(v.image.Bounds())
}

func (v *fontImageViewer) currentDisplaySelection() (image.Rectangle, bool) {
	if !v.tempSelection.Empty() {
		return v.tempSelection, true
	}
	if !v.selection.Empty() {
		return v.selection, true
	}
	return image.Rectangle{}, false
}

func (v *fontImageViewer) minSize() fyne.Size {
	w, h := v.displayPixelSize()
	return fyne.NewSize(float32(w), float32(h))
}

func (v *fontImageViewer) displayPixelSize() (int, int) {
	return v.displayPixelSizeForZoom(v.zoom)
}

func (v *fontImageViewer) displayPixelSizeForZoom(zoom float32) (int, int) {
	if v.image == nil || v.image.Bounds().Empty() {
		return 1, 1
	}
	bounds := v.image.Bounds()
	if zoom <= 0 {
		zoom = 1
	}
	w := int(math.Ceil(float64(float32(bounds.Dx()) * zoom)))
	h := int(math.Ceil(float64(float32(bounds.Dy()) * zoom)))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func (v *fontImageViewer) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromImage(v.image)
	img.FillMode = canvas.ImageFillStretch
	img.ScaleMode = canvas.ImageScalePixels
	selection := canvas.NewRectangle(color.Transparent)
	selection.StrokeColor = color.NRGBA{255, 170, 0, 255}
	selection.StrokeWidth = 2
	selection.Hide()
	return &fontImageViewerRenderer{
		viewer:    v,
		image:     img,
		selection: selection,
		objects:   []fyne.CanvasObject{img, selection},
	}
}

func (v *fontImageViewer) MouseDown(e *desktop.MouseEvent) {
	if v.image == nil {
		return
	}
	v.hideMagnifier()
	v.lastDragAbs = e.AbsolutePosition
	if e.Button == desktop.MouseButtonPrimary && e.Modifier&fyne.KeyModifierControl != 0 {
		p, ok := v.imagePosition(e.Position)
		if ok && v.onColorPicked != nil {
			v.onColorPicked(imageColorNRGBA(v.image, p.X, p.Y))
		}
		v.dragMode = fontImageDragNone
		v.updateMagnifier(e.Position)
		return
	}
	if e.Button == desktop.MouseButtonPrimary && v.selectionMode {
		p, ok := v.imagePosition(e.Position)
		if !ok {
			return
		}
		v.dragMode = fontImageDragSelect
		v.dragStart = p
		v.dragCurrent = p
		v.tempSelection = image.Rectangle{}
		return
	}
	if e.Button == desktop.MouseButtonPrimary {
		v.dragMode = fontImageDragPan
		return
	}
	v.dragMode = fontImageDragNone
}

func (v *fontImageViewer) MouseMoved(e *desktop.MouseEvent) {
	defer v.updateMagnifier(e.Position)
	switch v.dragMode {
	case fontImageDragSelect:
		p, ok := v.imagePosition(e.Position)
		if !ok {
			return
		}
		v.dragCurrent = p
		v.tempSelection = v.normalizedDragRect()
		v.Refresh()
	case fontImageDragPan:
		if v.scroll == nil {
			return
		}
		delta := e.AbsolutePosition.Subtract(v.lastDragAbs)
		offset := v.scroll.Offset
		v.scroll.ScrollToOffset(fyne.NewPos(offset.X-delta.X, offset.Y-delta.Y))
		v.lastDragAbs = e.AbsolutePosition
	}
}

func (v *fontImageViewer) MouseUp(e *desktop.MouseEvent) {
	if v.dragMode != fontImageDragSelect {
		v.dragMode = fontImageDragNone
		v.updateMagnifier(e.Position)
		return
	}
	p, ok := v.imagePosition(e.Position)
	if ok {
		v.dragCurrent = p
		rect := v.normalizedDragRect()
		if rect.Dx() <= 1 && rect.Dy() <= 1 {
			v.selection = image.Rectangle{}
		} else {
			v.selection = rect
		}
	}
	v.tempSelection = image.Rectangle{}
	v.dragMode = fontImageDragNone
	v.selectionMode = false
	v.Refresh()
	if v.onSelectionChanged != nil {
		v.onSelectionChanged(v.selection)
	}
	v.updateMagnifier(e.Position)
}

func (v *fontImageViewer) MouseIn(e *desktop.MouseEvent) {
	v.updateMagnifier(e.Position)
}

func (v *fontImageViewer) MouseOut() {
	v.hideMagnifier()
}

func (v *fontImageViewer) Cursor() desktop.Cursor {
	return desktop.CrosshairCursor
}

func (v *fontImageViewer) Scrolled(e *fyne.ScrollEvent) {
	driver, ok := fyne.CurrentApp().Driver().(desktop.Driver)
	if !ok || driver.CurrentKeyModifiers()&fyne.KeyModifierControl == 0 {
		if v.scroll != nil {
			v.scroll.Scrolled(e)
		}
		v.updateMagnifierFromViewport(e.Position)
		return
	}

	delta := e.Scrolled.DY
	if delta == 0 {
		delta = e.Scrolled.DX
	}
	if delta == 0 {
		return
	}
	v.zoomAt(e.Position, delta > 0)
	v.updateMagnifierFromViewport(e.Position)
}

type fontScrollOverlay struct {
	widget.BaseWidget
	onScroll func(*fyne.ScrollEvent)
}

func newFontScrollOverlay(onScroll func(*fyne.ScrollEvent)) *fontScrollOverlay {
	o := &fontScrollOverlay{onScroll: onScroll}
	o.ExtendBaseWidget(o)
	return o
}

func (o *fontScrollOverlay) CreateRenderer() fyne.WidgetRenderer {
	return fontScrollOverlayRenderer{}
}

func (o *fontScrollOverlay) Scrolled(e *fyne.ScrollEvent) {
	if o.onScroll != nil {
		o.onScroll(e)
	}
}

type fontScrollOverlayRenderer struct{}

func (fontScrollOverlayRenderer) Layout(fyne.Size)             {}
func (fontScrollOverlayRenderer) MinSize() fyne.Size           { return fyne.NewSize(1, 1) }
func (fontScrollOverlayRenderer) Refresh()                     {}
func (fontScrollOverlayRenderer) Objects() []fyne.CanvasObject { return nil }
func (fontScrollOverlayRenderer) Destroy()                     {}

type fontImageViewerRenderer struct {
	viewer    *fontImageViewer
	image     *canvas.Image
	selection *canvas.Rectangle
	objects   []fyne.CanvasObject
}

func (r *fontImageViewerRenderer) Layout(size fyne.Size) {
	r.image.Move(fyne.NewPos(0, 0))
	r.image.Resize(r.viewer.minSize())
	r.updateSelection()
}

func (r *fontImageViewerRenderer) MinSize() fyne.Size {
	return r.viewer.minSize()
}

func (r *fontImageViewerRenderer) Refresh() {
	r.image.Image = r.viewer.image
	r.image.Move(fyne.NewPos(0, 0))
	r.image.Resize(r.viewer.minSize())
	if r.viewer.image == nil {
		r.image.Hide()
	} else {
		r.image.Show()
	}
	r.updateSelection()
	canvas.Refresh(r.image)
	canvas.Refresh(r.selection)
}

func (r *fontImageViewerRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *fontImageViewerRenderer) Destroy() {}

func (r *fontImageViewerRenderer) updateSelection() {
	rect, ok := r.viewer.currentDisplaySelection()
	if !ok || r.viewer.image == nil {
		r.selection.Hide()
		return
	}
	bounds := r.viewer.image.Bounds()
	displayW, displayH := r.viewer.displayPixelSize()
	scaleX := float32(displayW) / float32(bounds.Dx())
	scaleY := float32(displayH) / float32(bounds.Dy())
	x := float32(rect.Min.X-bounds.Min.X) * scaleX
	y := float32(rect.Min.Y-bounds.Min.Y) * scaleY
	w := float32(rect.Dx()) * scaleX
	h := float32(rect.Dy()) * scaleY
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	r.selection.Move(fyne.NewPos(x, y))
	r.selection.Resize(fyne.NewSize(w, h))
	r.selection.Show()
}

// ==================== 字库制作窗口 ====================

func openFontLibWindow(parentWindow fyne.Window) {
	a := fyne.CurrentApp()
	w := a.NewWindow("AutoGo 字库制作")
	fontLibWindowSize := initialWindowSize(0.82, 0.78)

	var charCells []CharCell
	var charNameEntries []*widget.Entry
	var charHexCache []string
	var charWpCache []int
	var charMatchedLib []bool
	var binaryRegion *image.NRGBA
	var regionImg image.Image
	var fontLibChars []FontChar
	var suppressParamRefresh bool

	fgColorEntry := widget.NewEntry()
	fgColorEntry.SetText("000000")
	fgColorEntry.SetPlaceHolder("如: 000000")
	fgToleranceEntry := widget.NewEntry()
	fgToleranceEntry.SetText("101010")
	fgToleranceEntry.SetPlaceHolder("如: 101010")
	colGapEntry := widget.NewEntry()
	colGapEntry.SetText("1")
	rowGapEntry := widget.NewEntry()
	rowGapEntry.SetText("1")

	sourceInfoLabel := widget.NewLabel("请先去裁剪选取或加载图片")
	sourceInfoLabel.Wrapping = fyne.TextWrapWord
	previewInfoLabel := widget.NewLabel("绿色 = 文字前景，会入库；黑色 = 背景，会忽略")
	previewInfoLabel.Wrapping = fyne.TextWrapWord

	splitListBox := container.NewVBox()
	fontLibListBox := container.NewVBox()
	libHeaderLabel := widget.NewLabelWithStyle("字库内容 (0)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	libSearchEntry := widget.NewEntry()
	libSearchEntry.SetPlaceHolder("搜索字库字符")

	quickFillEntry := widget.NewEntry()
	quickFillEntry.SetPlaceHolder("快速填入: 主题壁纸")

	readInt := func(e *widget.Entry, def int) int {
		v, err := strconv.Atoi(strings.TrimSpace(e.Text))
		if err != nil || v < 0 {
			return def
		}
		return v
	}

	currentColorParam := func() string {
		fgHex := strings.TrimSpace(fgColorEntry.Text)
		if fgHex == "" {
			fgHex = "000000"
		}
		if strings.Contains(fgHex, "-") {
			return fgHex
		}
		toleranceHex := strings.TrimSpace(fgToleranceEntry.Text)
		if toleranceHex == "" {
			toleranceHex = "101010"
		}
		return fgHex + "-" + toleranceHex
	}

	var updateSourceInfo func()
	var refreshPreview func()
	var rebuildSplitList func()
	var rebuildLibList func()
	var setCropConfirmVisible func(bool)

	sourceViewer := newFontImageViewer()
	sourceViewer.onSelectionChanged = func(rect image.Rectangle) {
		if setCropConfirmVisible != nil {
			setCropConfirmVisible(!rect.Empty())
		}
		if updateSourceInfo != nil {
			updateSourceInfo()
		}
	}
	sourceViewer.onColorPicked = func(c color.NRGBA) {
		fgColorEntry.SetText(fontColorHex(c))
	}
	sourceScroll := container.NewScroll(sourceViewer)
	sourceViewer.SetScroll(sourceScroll)
	sourceScrollOverlay := newFontScrollOverlay(func(e *fyne.ScrollEvent) {
		sourceViewer.Scrolled(e)
	})
	sourceMagnifier := NewMagnifierWidget()
	sourceViewer.SetMagnifier(sourceMagnifier)

	previewViewer := newFontImageViewer()
	previewScroll := container.NewScroll(previewViewer)
	previewViewer.SetScroll(previewScroll)
	previewScrollOverlay := newFontScrollOverlay(func(e *fyne.ScrollEvent) {
		previewViewer.Scrolled(e)
	})

	updateSourceInfo = func() {
		if regionImg == nil {
			sourceInfoLabel.SetText("请先去裁剪选取或加载图片")
			return
		}
		bounds := regionImg.Bounds()
		text := fmt.Sprintf("当前图: %d×%d px | 左键拖动图片；Ctrl+左键取色；Ctrl+滚轮缩放",
			bounds.Dx(), bounds.Dy())
		if sourceViewer.SelectionMode() {
			text += " | 请拖拽选择裁剪区域"
		}
		if rect, ok := sourceViewer.SelectedRect(); ok {
			text += fmt.Sprintf(" | 选区: %d,%d - %d,%d (%d×%d)",
				rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y, rect.Dx(), rect.Dy())
		}
		sourceInfoLabel.SetText(text)
	}

	rebuildLibList = func() {
		fontLibListBox.RemoveAll()
		query := strings.ToLower(strings.TrimSpace(libSearchEntry.Text))
		shown := 0
		for i, ch := range fontLibChars {
			if query != "" && !strings.Contains(strings.ToLower(ch.Char), query) {
				continue
			}
			idx := i
			previewImg := canvas.NewImageFromImage(createCharPreview(ch.Bitmap, 2))
			previewImg.ScaleMode = canvas.ImageScalePixels
			previewImg.FillMode = canvas.ImageFillContain
			previewImg.SetMinSize(fyne.NewSize(30, 30))

			nameLabel := widget.NewLabelWithStyle(ch.Char, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			sizeLabel := widget.NewLabel(fmt.Sprintf("%dx%d | 白点:%d", ch.Width, ch.Height, ch.WhitePixels))
			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				fontLibChars = append(fontLibChars[:idx], fontLibChars[idx+1:]...)
				rebuildLibList()
				if refreshPreview != nil {
					refreshPreview()
				}
			})
			delBtn.Importance = widget.LowImportance

			row := container.NewBorder(nil, nil, previewImg, delBtn, container.NewVBox(nameLabel, sizeLabel))
			fontLibListBox.Add(row)
			fontLibListBox.Add(widget.NewSeparator())
			shown++
		}
		if shown == 0 {
			fontLibListBox.Add(widget.NewLabel("暂无字库内容"))
		}
		fontLibListBox.Refresh()
		if query == "" {
			libHeaderLabel.SetText(fmt.Sprintf("字库内容 (%d)", len(fontLibChars)))
		} else {
			libHeaderLabel.SetText(fmt.Sprintf("字库内容 (%d/%d)", shown, len(fontLibChars)))
		}
	}

	rebuildSplitList = func() {
		splitListBox.RemoveAll()
		charNameEntries = make([]*widget.Entry, len(charCells))
		charHexCache = make([]string, len(charCells))
		charWpCache = make([]int, len(charCells))
		charMatchedLib = make([]bool, len(charCells))

		if len(charCells) == 0 {
			splitListBox.Add(widget.NewLabel("暂无分割结果"))
			splitListBox.Refresh()
			return
		}

		libMatchNames := make(map[string]string, len(fontLibChars))
		for _, lc := range fontLibChars {
			key := fontCharMatchKey(lc)
			if key != "" && libMatchNames[key] == "" {
				libMatchNames[key] = lc.Char
			}
		}

		for i, cell := range charCells {
			hexData, wp := encodeBitmapHex(cell.Bitmap)
			charHexCache[i] = hexData
			charWpCache[i] = wp
			matchedName := libMatchNames[fontBitmapMatchKey(cell.Bitmap)]
			if matchedName != "" {
				charMatchedLib[i] = true
			}

			previewImg := canvas.NewImageFromImage(createCharPreview(cell.Bitmap, 2))
			previewImg.ScaleMode = canvas.ImageScalePixels
			previewImg.FillMode = canvas.ImageFillContain
			previewImg.SetMinSize(fyne.NewSize(42, 42))

			idLabel := widget.NewLabel(fmt.Sprintf("#%d  %dx%d", i, cell.BBox.Dx(), cell.BBox.Dy()))
			statusText := "未匹配字库"
			if matchedName != "" {
				statusText = "已匹配: " + matchedName
			}
			statusLabel := widget.NewLabel(statusText)
			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("字符")
			nameEntryContainer := container.New(&fixedWidthLayout{width: 74}, nameEntry)
			if matchedName != "" {
				nameEntry.SetText(matchedName)
			}
			charNameEntries[i] = nameEntry

			infoBox := container.NewVBox(idLabel, statusLabel)
			row := container.NewBorder(nil, nil, previewImg, nameEntryContainer, infoBox)
			splitListBox.Add(row)
			splitListBox.Add(widget.NewSeparator())
		}
		splitListBox.Refresh()
	}

	refreshPreview = func() {
		if regionImg == nil {
			binaryRegion = nil
			charCells = nil
			previewViewer.SetImage(nil)
			previewInfoLabel.SetText("请先去裁剪选取或加载图片")
			rebuildSplitList()
			return
		}

		cg := readInt(colGapEntry, 1)
		rg := readInt(rowGapEntry, 1)
		binaryRegion = createBinaryPreview(regionImg, currentColorParam())
		dotImg, cells := renderDotMatrix(binaryRegion, cg, rg)
		charCells = cells

		previewViewer.SetImage(dotImg)
		previewViewer.SetZoom(fontPreviewZoomForSourceZoom(sourceViewer.zoom))

		bw := binaryRegion.Bounds().Dx()
		bh := binaryRegion.Bounds().Dy()
		previewInfoLabel.SetText(fmt.Sprintf("二值预览: %d×%d px | 检测到 %d 个字符 | 列间距:%d 行间距:%d | 左键拖动；Ctrl+滚轮缩放",
			bw, bh, len(charCells), cg, rg))
		rebuildSplitList()
	}

	setRegionImage := func(img image.Image) {
		if img == nil || img.Bounds().Empty() {
			return
		}
		regionImg = img
		sourceViewer.SetImage(regionImg)
		sourceViewer.SetZoom(fontSourceInitialZoom(regionImg.Bounds()))
		updateSourceInfo()
		refreshPreview()
	}

	paramChanged := func(string) {
		if !suppressParamRefresh && refreshPreview != nil {
			refreshPreview()
		}
	}
	fgColorEntry.OnChanged = paramChanged
	fgToleranceEntry.OnChanged = paramChanged
	colGapEntry.OnChanged = paramChanged
	rowGapEntry.OnChanged = paramChanged
	libSearchEntry.OnChanged = func(string) { rebuildLibList() }

	quickFillBtn := widget.NewButtonWithIcon("快速填入", theme.ConfirmIcon(), func() {
		chars := []rune(strings.TrimSpace(quickFillEntry.Text))
		for i := 0; i < len(charNameEntries) && i < len(chars); i++ {
			charNameEntries[i].SetText(string(chars[i]))
		}
	})
	quickFillBtn.Importance = widget.HighImportance

	addToLibBtn := widget.NewButtonWithIcon("添加到字库", theme.ContentAddIcon(), func() {
		added := 0
		for i, cell := range charCells {
			if i >= len(charNameEntries) || i >= len(charMatchedLib) {
				break
			}
			if charMatchedLib[i] {
				continue
			}
			name := strings.TrimSpace(charNameEntries[i].Text)
			if name == "" {
				continue
			}
			bm := cell.Bitmap
			if len(bm) == 0 || len(bm[0]) == 0 {
				continue
			}
			fontLibChars = append(fontLibChars, FontChar{
				Char:        name,
				Width:       len(bm[0]),
				Height:      len(bm),
				HexData:     charHexCache[i],
				WhitePixels: charWpCache[i],
				Bitmap:      bm,
			})
			added++
		}
		if added > 0 {
			rebuildLibList()
			refreshPreview()
		}
	})
	addToLibBtn.Importance = widget.HighImportance

	getSelBtn := widget.NewButtonWithIcon("去裁剪选取", theme.VisibilityIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入", w)
			return
		}
		viewer := imageViewer
		sourceInfoLabel.SetText("请在主窗口图像上拖拽框选文字区域，松开后会自动回填到字库制作")
		viewer.SetRangeSelectModeWithCallback(func(rect image.Rectangle) {
			if imageViewer != viewer || viewer.image == nil {
				w.Show()
				w.RequestFocus()
				dialog.ShowInformation("提示", "当前图像已切换，请重新裁剪选取", w)
				return
			}
			selRect := normalizePickRect(viewer.image, rect)
			if selRect.Empty() {
				w.Show()
				w.RequestFocus()
				dialog.ShowInformation("提示", "主窗口选区无效，请重新裁剪选取", w)
				return
			}
			setRegionImage(cropImage(viewer.image, selRect))
			w.Show()
			w.RequestFocus()
		})
		parentWindow.Show()
		parentWindow.RequestFocus()
	})
	getSelBtn.Importance = widget.HighImportance

	loadImageBtn := widget.NewButtonWithIcon("加载图片", theme.FolderOpenIcon(), func() {
		go func() {
			fp, err := nativedialog.File().Filter("图片文件", "png", "jpg", "jpeg", "bmp").
				Title("加载字库图片").Load()
			if err != nil {
				return
			}
			data, err := os.ReadFile(fp)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("读取图片失败: %v", err), w) })
				return
			}
			img, err := decodeOpenCVTemplateBytes(data)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("图片解码失败: %v", err), w) })
				return
			}
			fyne.Do(func() { setRegionImage(img) })
		}()
	})

	autoPreprocessBtn := widget.NewButtonWithIcon("自动取色", theme.SearchIcon(), func() {
		if regionImg == nil {
			dialog.ShowInformation("提示", "请先去裁剪选取或加载图片", w)
			return
		}
		bounds := regionImg.Bounds()
		analysisRect := bounds
		if selected, ok := sourceViewer.SelectedRect(); ok {
			analysisRect = selected
		}
		analysisImg := cropImage(regionImg, analysisRect)
		background, foreground, ok := estimateFontForegroundColor(analysisImg)
		if !ok {
			dialog.ShowInformation("自动取色", "未识别到稳定的文字前景色，当前参数未修改", w)
			return
		}
		tolerance := estimateFontTolerance(analysisImg, foreground, background)

		suppressParamRefresh = true
		fgColorEntry.SetText(fontColorHex(foreground))
		fgToleranceEntry.SetText(fontColorHex(tolerance))
		suppressParamRefresh = false

		refreshPreview()
		previewInfoLabel.SetText(fmt.Sprintf("自动取色完成: 文字色 %s 容差 %s | 当前图未裁剪 | 检测到 %d 个字符",
			fontColorHex(foreground), fontColorHex(tolerance), len(charCells)))
	})
	autoPreprocessBtn.Importance = widget.HighImportance

	var cropConfirmRow *fyne.Container
	confirmCropBtn := widget.NewButtonWithIcon("确认裁剪", theme.ConfirmIcon(), func() {
		rect, ok := sourceViewer.SelectedRect()
		if !ok {
			dialog.ShowInformation("提示", "当前裁剪选区无效，请重新选择", w)
			return
		}
		setRegionImage(cropImage(regionImg, rect))
		if setCropConfirmVisible != nil {
			setCropConfirmVisible(false)
		}
	})
	confirmCropBtn.Importance = widget.HighImportance
	cancelCropBtn := widget.NewButton("取消", func() {
		sourceViewer.CancelSelectionMode()
		if setCropConfirmVisible != nil {
			setCropConfirmVisible(false)
		}
		updateSourceInfo()
	})
	cropConfirmRow = container.NewGridWithColumns(2, confirmCropBtn, cancelCropBtn)
	cropConfirmRow.Hide()
	setCropConfirmVisible = func(show bool) {
		if cropConfirmRow == nil {
			return
		}
		if show {
			cropConfirmRow.Show()
		} else {
			cropConfirmRow.Hide()
		}
		cropConfirmRow.Refresh()
	}

	cropBtn := widget.NewButtonWithIcon("裁剪", theme.ContentCutIcon(), func() {
		if regionImg == nil {
			dialog.ShowInformation("提示", "请先去裁剪选取或加载图片", w)
			return
		}
		sourceViewer.StartSelectionMode()
		if setCropConfirmVisible != nil {
			setCropConfirmVisible(false)
		}
		sourceInfoLabel.SetText("请在上方图片拖拽选择裁剪区域，松开后点击确认裁剪或取消")
	})

	refreshBtn := widget.NewButtonWithIcon("刷新预览", theme.ViewRefreshIcon(), func() {
		refreshPreview()
	})

	resetZoomBtn := widget.NewButton("重置缩放", func() {
		sourceViewer.ResetZoom()
		previewViewer.ResetZoom()
	})

	clearSelectionBtn := widget.NewButton("清除选区", func() {
		sourceViewer.CancelSelectionMode()
		if setCropConfirmVisible != nil {
			setCropConfirmVisible(false)
		}
		updateSourceInfo()
	})

	resetParamsBtn := widget.NewButton("重置参数", func() {
		suppressParamRefresh = true
		fgColorEntry.SetText("000000")
		fgToleranceEntry.SetText("101010")
		colGapEntry.SetText("1")
		rowGapEntry.SetText("1")
		suppressParamRefresh = false
		refreshPreview()
	})

	exportBtn := widget.NewButtonWithIcon("导出", theme.DocumentSaveIcon(), func() {
		if len(fontLibChars) == 0 {
			dialog.ShowInformation("提示", "字库为空", w)
			return
		}
		go func() {
			fp, err := nativedialog.File().Filter("字库文件", "txt").
				Title("导出字库").SetStartFile("fontlib.txt").Save()
			if err != nil {
				return
			}
			err = os.WriteFile(fp, []byte(exportFontLib(fontLibChars, currentColorParam())), 0644)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("保存失败: %v", err), w) })
				return
			}
			fyne.Do(func() {
				dialog.ShowInformation("成功", fmt.Sprintf("已导出 %d 个字符", len(fontLibChars)), w)
			})
		}()
	})

	copyBtn := widget.NewButtonWithIcon("复制", theme.ContentCopyIcon(), func() {
		if len(fontLibChars) == 0 {
			dialog.ShowInformation("提示", "字库为空", w)
			return
		}
		var sb strings.Builder
		for _, ch := range fontLibChars {
			sb.WriteString(fmt.Sprintf("%s$%s$1$0.0.%d$%d\n", ch.Char, ch.HexData, ch.WhitePixels, ch.Width))
		}
		w.Clipboard().SetContent(sb.String())
		dialog.ShowInformation("成功", fmt.Sprintf("已复制 %d 个字符到剪贴板", len(fontLibChars)), w)
	})

	importBtn := widget.NewButtonWithIcon("导入", theme.FolderOpenIcon(), func() {
		go func() {
			fp, err := nativedialog.File().Filter("字库文件", "txt", "lib").
				Title("导入字库").Load()
			if err != nil {
				return
			}
			data, err := os.ReadFile(fp)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("读取失败: %v", err), w) })
				return
			}
			imported := parseFontLib(string(data))
			if len(imported) == 0 {
				fyne.Do(func() { dialog.ShowInformation("提示", "未找到有效字库数据", w) })
				return
			}
			fyne.Do(func() {
				fontLibChars = append(fontLibChars, imported...)
				rebuildLibList()
				refreshPreview()
				dialog.ShowInformation("成功", fmt.Sprintf("已导入 %d 个字符", len(imported)), w)
			})
		}()
	})

	clearLibBtn := widget.NewButtonWithIcon("清空", theme.DeleteIcon(), func() {
		if len(fontLibChars) == 0 {
			return
		}
		dialog.ShowConfirm("确认", fmt.Sprintf("确定清空字库中的 %d 个字符？", len(fontLibChars)), func(ok bool) {
			if ok {
				fontLibChars = nil
				rebuildLibList()
				refreshPreview()
			}
		}, w)
	})

	fgColorRow := newFixedHeightContainer(container.NewBorder(nil, nil, widget.NewLabel("文字色:"), nil, fgColorEntry), 42)
	fgToleranceRow := newFixedHeightContainer(container.NewBorder(nil, nil, widget.NewLabel("偏色容差:"), nil, fgToleranceEntry), 42)
	colGapRow := newFixedHeightContainer(container.NewBorder(nil, nil, widget.NewLabel("列间距(像素):"), nil, colGapEntry), 42)
	rowGapRow := newFixedHeightContainer(container.NewBorder(nil, nil, widget.NewLabel("行间距(像素):"), nil, rowGapEntry), 42)

	leftPanel := container.New(&fixedWidthLayout{width: 190, padding: 10, verticalSpacing: 5},
		getSelBtn,
		loadImageBtn,
		widget.NewSeparator(),
		fgColorRow,
		fgToleranceRow,
		colGapRow,
		rowGapRow,
		widget.NewSeparator(),
		resetParamsBtn,
		widget.NewSeparator(),
		newFixedHeightContainer(container.NewGridWithColumns(2, cropBtn, refreshBtn), 44),
		newFixedHeightContainer(container.NewGridWithColumns(2, resetZoomBtn, clearSelectionBtn), 44),
		autoPreprocessBtn,
		layout.NewSpacer(),
	)

	sourcePanel := container.NewBorder(
		container.NewVBox(widget.NewLabelWithStyle("原图 / 当前裁剪图", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), sourceInfoLabel, cropConfirmRow),
		nil, nil, nil,
		container.NewStack(newGridBgWidget(), sourceScroll, sourceMagnifier, sourceScrollOverlay),
	)

	previewPanel := container.NewBorder(
		container.NewVBox(widget.NewLabelWithStyle("二值化预览 / 分割框", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), previewInfoLabel),
		nil, nil, nil,
		container.NewStack(newGridBgWidget(), previewScroll, previewScrollOverlay),
	)
	centerSplit := container.NewVSplit(sourcePanel, previewPanel)
	centerSplit.Offset = 0.52

	quickFillRow := container.NewBorder(nil, nil, nil, quickFillBtn, quickFillEntry)
	splitPanel := container.NewBorder(
		container.NewVBox(widget.NewLabelWithStyle("分割结果", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), quickFillRow, addToLibBtn),
		nil, nil,
		container.NewVScroll(splitListBox),
	)

	libButtons := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWithColumns(2, exportBtn, copyBtn),
		container.NewGridWithColumns(2, importBtn, clearLibBtn),
	)
	libPanel := container.NewBorder(
		container.NewVBox(libHeaderLabel, libSearchEntry, widget.NewSeparator()),
		libButtons,
		nil, nil,
		container.NewVScroll(fontLibListBox),
	)

	rightSplit := container.NewVSplit(splitPanel, libPanel)
	rightSplit.Offset = 0.55
	rightBg := canvas.NewRectangle(color.Transparent)
	rightBg.SetMinSize(fyne.NewSize(320, 0))
	rightPanel := container.NewStack(rightBg, container.NewPadded(rightSplit))
	centerAndRightSplit := container.NewHSplit(centerSplit, rightPanel)
	centerAndRightSplit.Offset = 0.68

	mainContent := container.NewBorder(nil, nil, leftPanel, nil, centerAndRightSplit)
	rebuildLibList()
	rebuildSplitList()
	w.SetContent(mainContent)
	w.Resize(fontLibWindowSize)
	w.CenterOnScreen()
	w.Show()
}

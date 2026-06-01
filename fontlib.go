package main

import (
	"fmt"
	"image"
	"image/color"
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

const dotCellSize = 8

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

// ==================== 字库制作窗口 ====================

func openFontLibWindow(parentWindow fyne.Window) {
	a := fyne.CurrentApp()
	w := a.NewWindow("AutoGo 字库制作")
	w.Resize(fyne.NewSize(1540, 850))

	var charCells []CharCell
	var charNameEntries []*widget.Entry
	var charHexCache []string
	var charWpCache []int
	var charMatchedLib []bool
	var binaryRegion *image.NRGBA
	var regionImg image.Image
	var fontLibChars []FontChar
	showingOriginal := false

	fgColorEntry := widget.NewEntry()
	fgColorEntry.SetText("000000-101010")
	fgColorEntry.SetPlaceHolder("如: 000000-101010")
	colGapEntry := widget.NewEntry()
	colGapEntry.SetText("1")
	rowGapEntry := widget.NewEntry()
	rowGapEntry.SetText("1")

	previewCanvasImg := canvas.NewImageFromImage(nil)
	previewCanvasImg.ScaleMode = canvas.ImageScalePixels
	previewCanvasImg.FillMode = canvas.ImageFillOriginal

	infoLabel := widget.NewLabel("请先获取选区或加载图片")
	infoLabel.Wrapping = fyne.TextWrapWord

	charCardHolder := container.NewHBox()
	scrollBarPad := canvas.NewRectangle(color.Transparent)
	scrollBarPad.SetMinSize(fyne.NewSize(1, 12))
	charCardInner := container.NewVBox(charCardHolder, scrollBarPad)
	charCardScroll := container.NewHScroll(charCardInner)

	quickFillEntry := widget.NewEntry()
	quickFillEntry.SetPlaceHolder("快速填入: 主题壁纸")
	quickFillBtn := widget.NewButtonWithIcon("快速填入", theme.ConfirmIcon(), func() {
		chars := []rune(strings.TrimSpace(quickFillEntry.Text))
		for i := 0; i < len(charNameEntries) && i < len(chars); i++ {
			charNameEntries[i].SetText(string(chars[i]))
		}
	})
	quickFillBtn.Importance = widget.HighImportance

	readInt := func(e *widget.Entry, def int) int {
		v, err := strconv.Atoi(strings.TrimSpace(e.Text))
		if err != nil || v < 0 {
			return def
		}
		return v
	}

	// ===== 右侧字库列表 =====
	libListBox := container.NewVBox()
	libListScroll := container.NewVScroll(libListBox)
	libHeaderLabel := widget.NewLabel("字库内容 (0)")
	libHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}

	var rebuildLibList func()
	rebuildLibList = func() {
		libListBox.RemoveAll()
		for i, ch := range fontLibChars {
			idx := i
			previewImg := canvas.NewImageFromImage(createCharPreview(ch.Bitmap, 2))
			previewImg.ScaleMode = canvas.ImageScalePixels
			previewImg.FillMode = canvas.ImageFillContain
			previewImg.SetMinSize(fyne.NewSize(26, 26))

			nameText := canvas.NewText(ch.Char, getTextColor(isDarkTheme))
			nameText.TextSize = 14
			nameText.TextStyle = fyne.TextStyle{Bold: true}

			sizeText := canvas.NewText(fmt.Sprintf("%dx%d", ch.Width, ch.Height), color.NRGBA{140, 140, 140, 255})
			sizeText.TextSize = 10

			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				fontLibChars = append(fontLibChars[:idx], fontLibChars[idx+1:]...)
				rebuildLibList()
			})
			delBtn.Importance = widget.LowImportance

			leftInfo := container.NewHBox(previewImg, container.NewCenter(nameText), container.NewCenter(sizeText))
			scrollPad := canvas.NewRectangle(color.Transparent)
			scrollPad.SetMinSize(fyne.NewSize(12, 1))
			row := container.NewBorder(nil, nil, leftInfo, container.NewHBox(delBtn, scrollPad))
			libListBox.Add(row)
		}
		libListBox.Refresh()
		libHeaderLabel.SetText(fmt.Sprintf("字库内容 (%d)", len(fontLibChars)))
	}

	// ===== 核心：开始切割 =====
	doSlice := func() {
		if regionImg == nil {
			return
		}
		showingOriginal = false
		fgHex := strings.TrimSpace(fgColorEntry.Text)
		if fgHex == "" {
			fgHex = "000000-101010"
		}
		cg := readInt(colGapEntry, 1)
		rg := readInt(rowGapEntry, 1)

		fullBinary := createBinaryPreview(regionImg, fgHex)
		bboxes := findCharacterBBoxes(fullBinary, cg, rg)
		if len(bboxes) > 0 {
			unionRect := bboxes[0]
			for _, bb := range bboxes[1:] {
				if bb.Min.X < unionRect.Min.X {
					unionRect.Min.X = bb.Min.X
				}
				if bb.Min.Y < unionRect.Min.Y {
					unionRect.Min.Y = bb.Min.Y
				}
				if bb.Max.X > unionRect.Max.X {
					unionRect.Max.X = bb.Max.X
				}
				if bb.Max.Y > unionRect.Max.Y {
					unionRect.Max.Y = bb.Max.Y
				}
			}
			pad := 2
			imgW, imgH := fullBinary.Bounds().Dx(), fullBinary.Bounds().Dy()
			x0 := max(0, unionRect.Min.X-pad)
			y0 := max(0, unionRect.Min.Y-pad)
			x1 := min(imgW, unionRect.Max.X+pad)
			y1 := min(imgH, unionRect.Max.Y+pad)
			expanded := image.Rect(x0, y0, x1, y1)
			cropped := image.NewNRGBA(image.Rect(0, 0, expanded.Dx(), expanded.Dy()))
			for y := 0; y < expanded.Dy(); y++ {
				for x := 0; x < expanded.Dx(); x++ {
					cropped.SetNRGBA(x, y, fullBinary.NRGBAAt(expanded.Min.X+x, expanded.Min.Y+y))
				}
			}
			binaryRegion = cropped
		} else {
			binaryRegion = fullBinary
		}

		dotImg, cells := renderDotMatrix(binaryRegion, cg, rg)
		charCells = cells

		previewCanvasImg.Image = dotImg
		previewCanvasImg.SetMinSize(fyne.NewSize(float32(dotImg.Bounds().Dx()), float32(dotImg.Bounds().Dy())))
		previewCanvasImg.Refresh()

		bw := binaryRegion.Bounds().Dx()
		bh := binaryRegion.Bounds().Dy()
		infoLabel.SetText(fmt.Sprintf("选区: %d×%d px | 检测到 %d 个字符 | 列间距:%d 行间距:%d",
			bw, bh, len(charCells), cg, rg))

		charCardHolder.RemoveAll()
		charNameEntries = make([]*widget.Entry, len(charCells))
		charHexCache = make([]string, len(charCells))
		charWpCache = make([]int, len(charCells))
		charMatchedLib = make([]bool, len(charCells))

		var cards []fyne.CanvasObject
		for i, cell := range charCells {
			hexData, wp := encodeBitmapHex(cell.Bitmap)
			charHexCache[i] = hexData
			charWpCache[i] = wp
			matchedName := ""
			for _, lc := range fontLibChars {
				if lc.HexData == hexData {
					matchedName = lc.Char
					charMatchedLib[i] = true
					break
				}
			}

			previewImg := canvas.NewImageFromImage(createCharPreview(cell.Bitmap, 2))
			previewImg.ScaleMode = canvas.ImageScalePixels
			previewImg.FillMode = canvas.ImageFillContain
			pw := float32(cell.BBox.Dx() * 2)
			ph := float32(cell.BBox.Dy() * 2)
			if pw < 20 {
				pw = 20
			}
			if pw > 70 {
				pw = 70
			}
			if ph < 20 {
				ph = 20
			}
			if ph > 50 {
				ph = 50
			}
			previewImg.SetMinSize(fyne.NewSize(pw, ph))

			idText := canvas.NewText(fmt.Sprintf("#%d  %dx%d", i, cell.BBox.Dx(), cell.BBox.Dy()), color.NRGBA{130, 170, 230, 255})
			idText.TextSize = 10

			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("字符")
			if matchedName != "" {
				nameEntry.SetText(matchedName)
			}
			charNameEntries[i] = nameEntry

			cardContent := container.NewBorder(
				idText, nameEntry, nil, nil,
				container.NewCenter(previewImg),
			)

			cardBg := canvas.NewRectangle(color.NRGBA{30, 30, 38, 255})
			if !isDarkTheme {
				cardBg = canvas.NewRectangle(color.NRGBA{235, 237, 242, 255})
			}
			cardMinW := canvas.NewRectangle(color.Transparent)
			cardMinW.SetMinSize(fyne.NewSize(100, 0))
			card := container.NewStack(cardMinW, cardBg, container.NewPadded(cardContent))
			cards = append(cards, card)
		}

		if len(cards) > 0 {
			row := container.NewHBox(cards...)
			charCardHolder.Add(row)
		}
		charCardHolder.Refresh()
	}

	// ===== 添加到字库 =====
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
			hexData := charHexCache[i]
			wp := charWpCache[i]
			newChar := FontChar{
				Char: name, Width: len(bm[0]), Height: len(bm),
				HexData: hexData, WhitePixels: wp, Bitmap: bm,
			}
			fontLibChars = append(fontLibChars, newChar)
			added++
		}
		if added > 0 {
			rebuildLibList()
		}
	})
	addToLibBtn.Importance = widget.HighImportance

	// ===== 按钮 =====
	getSelBtn := widget.NewButtonWithIcon("获取选区", theme.VisibilityIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入", w)
			return
		}
		if len(imageViewer.markRects) == 0 {
			dialog.ShowInformation("提示", "请先在主窗口图像上拖拽框选文字区域，然后再点此按钮", w)
			return
		}
		rect := imageViewer.markRects[0]
		selRect := image.Rect(
			min(rect.X1, rect.X2), min(rect.Y1, rect.Y2),
			max(rect.X1, rect.X2), max(rect.Y1, rect.Y2),
		)
		regionImg = cropImage(imageViewer.image, selRect)
		showingOriginal = true
		dotPreview := renderOriginalDotMatrix(regionImg)
		previewCanvasImg.Image = dotPreview
		previewCanvasImg.SetMinSize(fyne.NewSize(float32(dotPreview.Bounds().Dx()), float32(dotPreview.Bounds().Dy())))
		previewCanvasImg.Refresh()
		infoLabel.SetText(fmt.Sprintf("选区: %d×%d px | 请点击「开始切割」进行二值化分割",
			regionImg.Bounds().Dx(), regionImg.Bounds().Dy()))
	})
	getSelBtn.Importance = widget.HighImportance

	sliceBtn := widget.NewButtonWithIcon("开始切割", theme.MediaPlayIcon(), func() {
		if regionImg == nil {
			dialog.ShowInformation("提示", "请先获取选区或加载图片", w)
			return
		}
		doSlice()
	})
	sliceBtn.Importance = widget.HighImportance

	// ===== 字库操作按钮（右侧面板底部）=====
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
			fgHex := strings.TrimSpace(fgColorEntry.Text)
			err = os.WriteFile(fp, []byte(exportFontLib(fontLibChars, fgHex)), 0644)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("保存失败: %v", err), w) })
				return
			}
			fyne.Do(func() {
				dialog.ShowInformation("成功", fmt.Sprintf("已导出 %d 个字符", len(fontLibChars)), w)
			})
		}()
	})
	exportBtn.Importance = widget.MediumImportance

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
	copyBtn.Importance = widget.MediumImportance

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
				dialog.ShowInformation("成功", fmt.Sprintf("已导入 %d 个字符", len(imported)), w)
			})
		}()
	})
	importBtn.Importance = widget.MediumImportance

	clearLibBtn := widget.NewButtonWithIcon("清空", theme.DeleteIcon(), func() {
		if len(fontLibChars) == 0 {
			return
		}
		dialog.ShowConfirm("确认", fmt.Sprintf("确定清空字库中的 %d 个字符？", len(fontLibChars)), func(ok bool) {
			if ok {
				fontLibChars = nil
				rebuildLibList()
			}
		}, w)
	})
	clearLibBtn.Importance = widget.MediumImportance

	// ===== 布局 =====
	leftPanel := container.New(&fixedWidthLayout{width: 155, padding: 10, verticalSpacing: 4},
		layout.NewSpacer(),
		getSelBtn,
		widget.NewSeparator(),
		widget.NewLabel("文字颜色:"),
		fgColorEntry,
		widget.NewSeparator(),
		widget.NewLabel("列间距(像素):"),
		colGapEntry,
		widget.NewLabel("行间距(像素):"),
		rowGapEntry,
		widget.NewSeparator(),
		sliceBtn,
		layout.NewSpacer(),
	)

	gridBg := newGridBgWidget()
	previewTappable := newTappableArea(func(pos fyne.Position) {
		if !showingOriginal || regionImg == nil {
			return
		}
		stride := float32(dotCellSize + 1)
		px := int(pos.X / stride)
		py := int(pos.Y / stride)
		b := regionImg.Bounds()
		if px < 0 || py < 0 || px >= b.Dx() || py >= b.Dy() {
			return
		}
		r, g, bl, _ := regionImg.At(b.Min.X+px, b.Min.Y+py).RGBA()
		hexColor := fmt.Sprintf("%02X%02X%02X", r>>8, g>>8, bl>>8)
		cur := fgColorEntry.Text
		if idx := strings.Index(cur, "-"); idx >= 0 {
			fgColorEntry.SetText(hexColor + cur[idx:])
		} else {
			fgColorEntry.SetText(hexColor)
		}
	})
	previewScroll := container.NewScroll(container.NewStack(
		container.New(&topLeftLayout{}, previewCanvasImg),
		previewTappable,
	))
	previewArea := container.NewStack(gridBg, previewScroll)

	quickFillRow := container.NewBorder(nil, nil, nil,
		container.NewHBox(quickFillBtn, addToLibBtn),
		quickFillEntry,
	)
	charCardArea := newFixedHeightContainer(charCardScroll, 130)

	centerArea := container.NewBorder(
		infoLabel,
		container.NewVBox(
			widget.NewSeparator(),
			charCardArea,
			quickFillRow,
		),
		nil, nil,
		previewArea,
	)

	libBtnRow1 := container.NewGridWithColumns(2, exportBtn, copyBtn)
	libBtnRow2 := container.NewGridWithColumns(2, importBtn, clearLibBtn)

	rightContent := container.NewBorder(
		container.NewVBox(libHeaderLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), libBtnRow1, libBtnRow2),
		nil, nil,
		libListScroll,
	)
	rightBg := canvas.NewRectangle(color.Transparent)
	rightBg.SetMinSize(fyne.NewSize(200, 0))
	rightPanel := container.NewStack(rightBg, container.NewPadded(rightContent))

	mainContent := container.NewBorder(nil, nil, leftPanel, rightPanel, centerArea)
	w.SetContent(mainContent)
	w.Show()
}

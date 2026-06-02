package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativedialog "github.com/sqweek/dialog"
	"golang.org/x/image/bmp"
)

// 颜色点信息结构
type ColorPoint struct {
	ID       int
	Position string
	Color    string
	Offset   string
	Selected bool
}

// 标签页数据结构 - 保存每个标签页的独立数据
type TabData struct {
	colorPoints        []ColorPoint // 该标签页的颜色点列表
	markRects          []MarkRect   // 该标签页的矩形标记
	manualRectSelected bool         // 是否手动框选了区域
	imageViewer        *ImageViewer // 该标签页的图像查看器
	generatedCode      string       // 该标签页生成的代码
}

type UserConfig struct {
	Precision     string `json:"precision"`
	UniformOffset string `json:"uniform_offset"`
	PickCount     string `json:"pick_count"`
	PickMode      string `json:"pick_mode"`
	FunctionMode  string `json:"function_mode"`
	DirectionMode string `json:"direction_mode"`
	ShowMagnifier bool   `json:"show_magnifier"`
	AutoCopyRange bool   `json:"auto_copy_range"`
	ApplyRange    bool   `json:"apply_range"`
	GridMode      bool   `json:"grid_mode"`
	GridCols      int    `json:"grid_cols"`
	GridRows      int    `json:"grid_rows"`
	GridSpacing   int    `json:"grid_spacing"`

	FormatTemplates map[string]string `json:"format_templates"`
}

// 全局变量定义
var (
	headerBgColor      = color.NRGBA{30, 30, 30, 255}    // 表头深色背景
	lightHeaderBgColor = color.NRGBA{220, 220, 220, 255} // 表头浅色背景
	transparent        = color.NRGBA{0, 0, 0, 0}         // 透明色

	// 主题状态变量 - 初始值会在程序启动时根据系统主题设置
	isDarkTheme = false

	// 设备选择相关变量
	deviceSelect      *widget.Select
	selectedDevice    string
	deviceRefreshChan = make(chan bool)

	// 区域坐标显示
	rectCoordEntry *widget.Entry

	// 偏色值输入
	colorOffsetEntry *widget.Entry

	// 右侧表格新增点时使用的默认偏色
	defaultColorPointOffset = "202020"

	// 框选范围后是否自动复制
	autoCopyRangeEnabled = true

	// 找色模式选择
	colorModeRadio *widget.RadioGroup

	// 图像查看器（当前活动的）
	imageViewer *ImageViewer

	// 放大镜显示状态
	magnifierEnabled = true

	// 代码显示框
	codeDisplayEntry *widget.Entry

	// 图色面板 API 字段刷新回调
	updateImagesAPIFields func() string

	// 图色面板结果代码格式模板
	apiFormatTemplates = defaultAPIFormatTemplates()

	// 点阵模式状态
	gridModeEnabled  = false // 默认关闭点阵模式
	gridColsValue    = 4
	gridRowsValue    = 4
	gridSpacingValue = 7
	adb              = findADBPath()
)

func initialWindowSize(widthRatio, heightRatio float32) fyne.Size {
	const (
		fallbackWidth  = 1000
		fallbackHeight = 650
	)

	screenWidth, screenHeight := screenSizePixels()
	if screenWidth <= 0 || screenHeight <= 0 {
		return fyne.NewSize(fallbackWidth, fallbackHeight)
	}

	scale := screenScale()
	if scale <= 0 {
		scale = 1
	}

	return fyne.NewSize(
		float32(screenWidth)/scale*widthRatio,
		float32(screenHeight)/scale*heightRatio,
	)
}

// 固定高度的容器布局
type fixedHeightContainer struct {
	widget.BaseWidget
	content fyne.CanvasObject
	height  float32
}

func newFixedHeightContainer(content fyne.CanvasObject, height float32) *fixedHeightContainer {
	c := &fixedHeightContainer{content: content, height: height}
	c.ExtendBaseWidget(c)
	return c
}

func (c *fixedHeightContainer) CreateRenderer() fyne.WidgetRenderer {
	return &fixedHeightRenderer{container: c}
}

type fixedHeightRenderer struct {
	container *fixedHeightContainer
}

func (r *fixedHeightRenderer) Layout(size fyne.Size) {
	r.container.content.Resize(fyne.NewSize(size.Width, r.container.height))
	r.container.content.Move(fyne.NewPos(0, 0))
}

func (r *fixedHeightRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.container.content.MinSize().Width, r.container.height)
}

func (r *fixedHeightRenderer) Refresh() {
	r.container.content.Refresh()
}

func (r *fixedHeightRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container.content}
}

func (r *fixedHeightRenderer) Destroy() {}

// 获取当前主题下适合的文字颜色
func getTextColor(isDark bool) color.Color {
	if isDark {
		return color.White
	} else {
		return color.Black
	}
}

// 获取当前主题下适合的表头背景色
func getHeaderBgColor(isDark bool) color.Color {
	if isDark {
		return headerBgColor
	} else {
		return lightHeaderBgColor
	}
}

// 判断颜色亮度，决定是使用白色还是黑色文字
func getContrastColor(bgColor color.Color) color.Color {
	r, g, b, _ := bgColor.RGBA()
	// 转换为0-255范围
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)

	// 计算亮度 (亮度公式: 0.299*R + 0.587*G + 0.114*B)
	brightness := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)

	// 如果亮度大于阈值，使用黑色，否则使用白色
	if brightness > 128 {
		return color.Black
	}
	return color.White
}

// 自定义主题，使用微软雅黑字体的深色主题
type myTheme struct {
	fyne.Theme
}

// 设置深色主题
func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// 如果是深色主题，使用深色变体
	if variant == theme.VariantDark || isDarkTheme {
		// 为下拉框背景设置更明显的颜色
		if name == theme.ColorNameInputBackground {
			return color.NRGBA{60, 60, 60, 255} // 深色主题下使用深灰色作为输入框背景
		} else if name == theme.ColorNameFocus {
			// 消除焦点背景，使其与输入框背景相同
			return color.NRGBA{60, 60, 60, 255}
		}
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}

	// 如果是浅色主题，对特定元素进行自定义
	if name == theme.ColorNameBackground {
		return color.NRGBA{240, 240, 240, 255} // 使用浅灰色背景而非纯白色
	} else if name == theme.ColorNameButton {
		// 浅色主题下按钮背景使用浅灰色，以便与背景区分
		return color.NRGBA{220, 220, 220, 255}
	} else if name == theme.ColorNameInputBackground {
		// 浅色主题下为下拉框设置背景色
		return color.NRGBA{220, 220, 220, 255}
	} else if name == theme.ColorNameFocus {
		// 消除焦点背景，使其与输入框背景相同
		return color.NRGBA{220, 220, 220, 255}
	} else if name == theme.ColorNameShadow {
		// 增强浅色主题下的阴影可见度
		return color.NRGBA{0, 0, 0, 40}
	}

	// 其他颜色使用默认主题的浅色变体
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

// 设置尺寸
func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	// 设置滚动条始终为粗的样式
	if name == theme.SizeNameScrollBar {
		return 10
	}
	if name == theme.SizeNameScrollBarSmall {
		return 10
	}

	// 其他尺寸使用默认值
	return theme.DefaultTheme().Size(name)
}

// 自定义主题，只修改需要的主题属性
func newMyTheme() fyne.Theme {
	return &myTheme{Theme: theme.DarkTheme()}
}

// 自定义可点击表格行
type ClickableTableRow struct {
	widget.BaseWidget
	background    *canvas.Rectangle
	content       *fyne.Container
	onTapped      func()
	id            int
	isHighlighted bool
}

// 创建表格标题更新函数 - 前向声明
var updateTableHeader func()

// 创建表格选中更新函数 - 前向声明
var updateTableSelection func()
var tableContent *widget.List
var tableHeader *fyne.Container
var headerBg *canvas.Rectangle
var idHeader, posHeader, colorHeader, statusHeader *canvas.Text

func defaultUserConfig() UserConfig {
	return UserConfig{
		Precision:     "0.90",
		UniformOffset: "202020",
		PickCount:     "20个",
		PickMode:      "轮廓取点",
		FunctionMode:  "findMultiColor",
		DirectionMode: "0: 从左到右，从上到下",
		ShowMagnifier: true,
		AutoCopyRange: true,
		ApplyRange:    false,
		GridMode:      false,
		GridCols:      4,
		GridRows:      4,
		GridSpacing:   7,

		FormatTemplates: defaultAPIFormatTemplates(),
	}
}

func userConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "AutoGo图色助手", "config.json"), nil
}

func normalizeUserConfig(config UserConfig) UserConfig {
	defaults := defaultUserConfig()
	if strings.TrimSpace(config.Precision) == "" {
		config.Precision = defaults.Precision
	}
	if strings.TrimSpace(config.UniformOffset) == "" {
		config.UniformOffset = defaults.UniformOffset
	}
	if strings.TrimSpace(config.PickCount) == "" {
		config.PickCount = defaults.PickCount
	}
	if strings.TrimSpace(config.PickMode) == "" {
		config.PickMode = defaults.PickMode
	}
	if strings.TrimSpace(config.FunctionMode) == "" {
		config.FunctionMode = defaults.FunctionMode
	}
	if strings.TrimSpace(config.DirectionMode) == "" {
		config.DirectionMode = defaults.DirectionMode
	}
	if config.GridCols <= 0 {
		config.GridCols = defaults.GridCols
	}
	if config.GridRows <= 0 {
		config.GridRows = defaults.GridRows
	}
	if config.GridSpacing <= 0 {
		config.GridSpacing = defaults.GridSpacing
	}
	config.FormatTemplates = normalizeAPIFormatTemplates(config.FormatTemplates)
	return config
}

func loadUserConfig() UserConfig {
	config := defaultUserConfig()
	path, err := userConfigPath()
	if err != nil {
		return config
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultUserConfig()
	}
	return normalizeUserConfig(config)
}

func saveUserConfig(config UserConfig) error {
	path, err := userConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(normalizeUserConfig(config), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func saveUserConfigSilently(config UserConfig) {
	if err := saveUserConfig(config); err != nil {
		log.Printf("保存配置失败: %v", err)
	}
}

func newClickableTableRow(bg color.Color, content *fyne.Container, onTapped func()) *ClickableTableRow {
	row := &ClickableTableRow{
		background: canvas.NewRectangle(bg),
		content:    content,
		onTapped:   onTapped,
	}
	row.ExtendBaseWidget(row)
	return row
}

type commitEntry struct {
	widget.BaseWidget
	Text        string
	cursor      int
	focused     bool
	disabled    bool
	onCommit    func(string)
	OnSubmitted func(string)
}

func newCommitEntry() *commitEntry {
	entry := &commitEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *commitEntry) SetText(text string) {
	e.Text = text
	if e.cursor > len([]rune(e.Text)) {
		e.cursor = len([]rune(e.Text))
	}
	e.Refresh()
}

func (e *commitEntry) Enable() {
	e.disabled = false
	e.Refresh()
}

func (e *commitEntry) Disable() {
	e.disabled = true
	e.focused = false
	e.Refresh()
}

func (e *commitEntry) Tapped(*fyne.PointEvent) {
	if e.disabled {
		return
	}
	canvas := fyne.CurrentApp().Driver().CanvasForObject(e)
	if canvas != nil {
		canvas.Focus(e)
	}
}

func (e *commitEntry) FocusGained() {
	if e.disabled {
		return
	}
	e.focused = true
	e.cursor = len([]rune(e.Text))
	e.Refresh()
}

func (e *commitEntry) FocusLost() {
	if !e.focused {
		return
	}
	e.focused = false
	if e.onCommit != nil {
		e.onCommit(e.Text)
	}
	e.Refresh()
}

func (e *commitEntry) TypedRune(r rune) {
	if e.disabled {
		return
	}

	runes := []rune(e.Text)
	if e.cursor < 0 || e.cursor > len(runes) {
		e.cursor = len(runes)
	}
	runes = append(runes[:e.cursor], append([]rune{r}, runes[e.cursor:]...)...)
	e.Text = string(runes)
	e.cursor++
	e.Refresh()
}

func (e *commitEntry) TypedKey(key *fyne.KeyEvent) {
	if e.disabled {
		return
	}

	runes := []rune(e.Text)
	switch key.Name {
	case fyne.KeyLeft:
		if e.cursor > 0 {
			e.cursor--
		}
	case fyne.KeyRight:
		if e.cursor < len(runes) {
			e.cursor++
		}
	case fyne.KeyHome:
		e.cursor = 0
	case fyne.KeyEnd:
		e.cursor = len(runes)
	case fyne.KeyBackspace:
		if e.cursor > 0 {
			runes = append(runes[:e.cursor-1], runes[e.cursor:]...)
			e.cursor--
			e.Text = string(runes)
		}
	case fyne.KeyDelete:
		if e.cursor < len(runes) {
			runes = append(runes[:e.cursor], runes[e.cursor+1:]...)
			e.Text = string(runes)
		}
	case fyne.KeyReturn, fyne.KeyEnter:
		e.focused = false
		if e.OnSubmitted != nil {
			e.OnSubmitted(e.Text)
		}
		canvas := fyne.CurrentApp().Driver().CanvasForObject(e)
		if canvas != nil {
			canvas.Unfocus()
		}
	}
	e.Refresh()
}

func (e *commitEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if e.disabled {
		return
	}

	switch s := shortcut.(type) {
	case *fyne.ShortcutPaste:
		if s.Clipboard != nil {
			for _, r := range s.Clipboard.Content() {
				e.TypedRune(r)
			}
		}
	case *fyne.ShortcutCopy:
		if s.Clipboard != nil {
			s.Clipboard.SetContent(e.Text)
		}
	}
}

func (e *commitEntry) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	label := canvas.NewText("", getTextColor(isDarkTheme))
	label.TextSize = 12
	return &commitEntryRenderer{
		entry:   e,
		bg:      bg,
		label:   label,
		objects: []fyne.CanvasObject{bg, label},
	}
}

type commitEntryRenderer struct {
	entry   *commitEntry
	bg      *canvas.Rectangle
	label   *canvas.Text
	objects []fyne.CanvasObject
}

func (r *commitEntryRenderer) Destroy() {}

func (r *commitEntryRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	labelSize := r.label.MinSize()
	r.label.Move(fyne.NewPos(6, (size.Height-labelSize.Height)/2))
	r.label.Resize(fyne.NewSize(fyne.Max(0, size.Width-12), labelSize.Height))
}

func (r *commitEntryRenderer) MinSize() fyne.Size {
	return fyne.NewSize(58, 24)
}

func (r *commitEntryRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *commitEntryRenderer) Refresh() {
	if r.entry.disabled {
		r.bg.FillColor = transparent
		r.label.Text = ""
	} else {
		if isDarkTheme {
			r.bg.FillColor = color.NRGBA{55, 55, 55, 255}
		} else {
			r.bg.FillColor = color.NRGBA{220, 220, 220, 255}
		}
		r.label.Text = r.displayText()
	}
	r.label.Color = getTextColor(isDarkTheme)
	r.bg.Refresh()
	r.label.Refresh()
}

func (r *commitEntryRenderer) displayText() string {
	if !r.entry.focused {
		return r.entry.Text
	}

	runes := []rune(r.entry.Text)
	cursor := r.entry.cursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	runes = append(runes[:cursor], append([]rune{'|'}, runes[cursor:]...)...)
	return string(runes)
}

func (r *ClickableTableRow) TappedSecondary(*fyne.PointEvent) {
	// 右键点击，可以实现其他功能
}

func (r *ClickableTableRow) CreateRenderer() fyne.WidgetRenderer {
	r.background.SetMinSize(fyne.NewSize(250, 30))
	return &clickableTableRowRenderer{
		row:     r,
		objects: []fyne.CanvasObject{r.background, r.content},
	}
}

type clickableTableRowRenderer struct {
	row     *ClickableTableRow
	objects []fyne.CanvasObject
}

func (r *clickableTableRowRenderer) Destroy() {}

func (r *clickableTableRowRenderer) Layout(size fyne.Size) {
	r.row.background.Resize(size)
	r.row.content.Resize(size)
}

func (r *clickableTableRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(250, 30)
}

func (r *clickableTableRowRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *clickableTableRowRenderer) Refresh() {
	r.Layout(r.row.Size())
}

// 解析十六进制颜色字符串为RGBA颜色
func hexToColor(hex string) color.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return color.RGBA{100, 100, 255, 255} // 默认蓝色
	}

	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// 带动画的截图按钮
type AnimatedScreenshotButton struct {
	widget.BaseWidget
	button        *widget.Button
	animationView *canvas.Raster
	isLoading     bool
	scale         float32
	scaleDir      float32 // 1 为放大, -1 为缩小
	animationStop chan bool
	onTapped      func()
}

func NewAnimatedScreenshotButton(text string, icon fyne.Resource, tapped func()) *AnimatedScreenshotButton {
	btn := &AnimatedScreenshotButton{
		scale:    0.5,
		scaleDir: 1.0,
		onTapped: tapped,
	}
	btn.button = widget.NewButtonWithIcon(text, icon, func() {
		if !btn.isLoading {
			tapped()
		}
	})
	btn.button.Importance = widget.MediumImportance

	// 创建动画视图
	btn.animationView = canvas.NewRaster(btn.drawAnimation)
	btn.animationView.SetMinSize(fyne.NewSize(20, 20))
	btn.animationView.Hide()

	btn.ExtendBaseWidget(btn)
	return btn
}

// 绘制动画圆圈
func (b *AnimatedScreenshotButton) drawAnimation(w, h int) image.Image {
	if !b.isLoading || w == 0 || h == 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// 圆圈中心位置
	centerX := w / 2
	centerY := h / 2
	baseRadius := 6.0
	radius := baseRadius * float64(b.scale)

	// 绘制填充的圆
	fillColor := color.NRGBA{66, 150, 255, 200}
	strokeColor := color.NRGBA{66, 150, 255, 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= radius {
				img.Set(x, y, fillColor)
			} else if dist <= radius+1.5 {
				// 描边
				img.Set(x, y, strokeColor)
			}
		}
	}

	return img
}

func (b *AnimatedScreenshotButton) StartLoading() {
	if b.isLoading {
		return
	}
	b.isLoading = true
	b.button.Disable()
	b.scale = 0.5
	b.scaleDir = 1.0
	b.animationView.Show()
	b.animationStop = make(chan bool)

	// 启动动画
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // 约60fps，更流畅
		defer ticker.Stop()

		for {
			select {
			case <-b.animationStop:
				return
			case <-ticker.C:
				// 更新缩放 - 从0.5到1.5之间缩放，速度更快
				b.scale += b.scaleDir * 0.15
				if b.scale >= 1.6 {
					b.scale = 1.6
					b.scaleDir = -1.0
				} else if b.scale <= 0.4 {
					b.scale = 0.4
					b.scaleDir = 1.0
				}

				// 使用fyne.Do确保在主线程中刷新UI
				fyne.Do(func() {
					b.animationView.Refresh()
				})
			}
		}
	}()
}

func (b *AnimatedScreenshotButton) StopLoading() {
	if !b.isLoading {
		return
	}
	b.isLoading = false

	if b.animationStop != nil {
		close(b.animationStop)
		b.animationStop = nil
	}

	// 使用fyne.Do确保在主线程中更新UI
	fyne.Do(func() {
		b.button.Enable()
		b.scale = 0.5
		b.scaleDir = 1.0
		b.animationView.Hide()
		b.Refresh()
	})
}

func (b *AnimatedScreenshotButton) IsLoading() bool {
	return b.isLoading
}

func (b *AnimatedScreenshotButton) CreateRenderer() fyne.WidgetRenderer {
	return &animatedScreenshotButtonRenderer{
		button:        b,
		buttonWidget:  b.button,
		animationView: b.animationView,
		objects:       []fyne.CanvasObject{b.button, b.animationView},
	}
}

type animatedScreenshotButtonRenderer struct {
	button        *AnimatedScreenshotButton
	buttonWidget  *widget.Button
	animationView *canvas.Raster
	objects       []fyne.CanvasObject
}

func (r *animatedScreenshotButtonRenderer) Layout(size fyne.Size) {
	r.buttonWidget.Resize(size)

	// 动画圆圈位置：在按钮右侧
	animSize := float32(20)
	animX := size.Width - animSize - 8
	animY := (size.Height - animSize) / 2

	r.animationView.Move(fyne.NewPos(animX, animY))
	r.animationView.Resize(fyne.NewSize(animSize, animSize))
}

func (r *animatedScreenshotButtonRenderer) MinSize() fyne.Size {
	return r.buttonWidget.MinSize()
}

func (r *animatedScreenshotButtonRenderer) Refresh() {
	r.buttonWidget.Refresh()
	r.animationView.Refresh()
	r.Layout(r.button.Size())
}

func (r *animatedScreenshotButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *animatedScreenshotButtonRenderer) Destroy() {
	if r.button.animationStop != nil {
		close(r.button.animationStop)
	}
}

// 完全自定义的颜色复选框
type ColorCheck struct {
	widget.BaseWidget
	Checked   bool
	OnChanged func(bool)
	Color     color.Color
}

// 创建新的彩色复选框
func NewColorCheck(checked bool, fillColor color.Color, changed func(bool)) *ColorCheck {
	check := &ColorCheck{
		Checked:   checked,
		OnChanged: changed,
		Color:     fillColor,
	}
	check.ExtendBaseWidget(check)
	return check
}

// 处理点击事件
func (c *ColorCheck) Tapped(*fyne.PointEvent) {
	c.Checked = !c.Checked
	if c.OnChanged != nil {
		c.OnChanged(c.Checked)
	}
	c.Refresh()
}

// 鼠标进入时显示指针
func (c *ColorCheck) MouseIn(*desktop.MouseEvent) {
	c.Refresh()
}

// 鼠标离开时恢复
func (c *ColorCheck) MouseOut() {
	c.Refresh()
}

// 实现桌面鼠标悬停接口
func (c *ColorCheck) MouseMoved(*desktop.MouseEvent) {}
func (c *ColorCheck) CursorType() desktop.Cursor {
	return desktop.PointerCursor
}

// 创建渲染器
func (c *ColorCheck) CreateRenderer() fyne.WidgetRenderer {
	// 创建绘制组件
	box := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	box.StrokeWidth = 1                               // 边框保持1像素
	box.StrokeColor = color.NRGBA{160, 160, 160, 200} // 使用灰色边框，不那么刺眼

	// 创建选中标记 - 使用两条线组成勾号
	// 勾的颜色根据背景色而定，初始化时可以先用白色，Refresh时会更新
	checkLine1 := canvas.NewLine(color.White)
	checkLine1.StrokeWidth = 2 // 回到标准线宽

	checkLine2 := canvas.NewLine(color.White)
	checkLine2.StrokeWidth = 2 // 回到标准线宽

	renderer := &colorCheckRenderer{
		check:      c,
		box:        box,
		checkLine1: checkLine1,
		checkLine2: checkLine2,
		objects:    []fyne.CanvasObject{box, checkLine1, checkLine2},
	}

	// 如果已经选中，立即设置正确的对比色
	if c.Checked {
		contrastColor := getContrastColor(c.Color)
		checkLine1.StrokeColor = contrastColor
		checkLine2.StrokeColor = contrastColor
	}

	return renderer
}

// 自定义渲染器
type colorCheckRenderer struct {
	check      *ColorCheck
	box        *canvas.Rectangle
	checkLine1 *canvas.Line
	checkLine2 *canvas.Line
	objects    []fyne.CanvasObject
}

func (r *colorCheckRenderer) MinSize() fyne.Size {
	return fyne.NewSize(16, 16) // 从18x18减小到16x16
}

func (r *colorCheckRenderer) Layout(size fyne.Size) {
	// 计算复选框的尺寸和位置
	boxSize := fyne.Min(size.Width, size.Height)
	r.box.Resize(fyne.NewSize(boxSize, boxSize))
	r.box.Move(fyne.NewPos(0, (size.Height-boxSize)/2))

	// 调整勾选图标位置，使其在复选框中更加居中
	// 整体向右下方移动一些

	// 第一条线 - 从左往右稍微向下倾斜
	r.checkLine1.Position1 = fyne.NewPos(boxSize*0.25, boxSize*0.5)  // 右移并下移起点
	r.checkLine1.Position2 = fyne.NewPos(boxSize*0.45, boxSize*0.65) // 右移并下移终点

	// 第二条线 - 从中间向右上方延伸
	r.checkLine2.Position1 = fyne.NewPos(boxSize*0.45, boxSize*0.65) // 与第一条线终点一致
	r.checkLine2.Position2 = fyne.NewPos(boxSize*0.8, boxSize*0.35)  // 右移并下移终点
}

func (r *colorCheckRenderer) Refresh() {
	if r.check.Checked {
		// 如果选中，使用指定的颜色填充，边框设为0使其不可见
		r.box.FillColor = r.check.Color
		r.box.StrokeWidth = 0 // 选中时不显示边框
		r.checkLine1.Hidden = false
		r.checkLine2.Hidden = false

		// 根据背景色亮度决定勾的颜色
		contrastColor := getContrastColor(r.check.Color)
		r.checkLine1.StrokeColor = contrastColor
		r.checkLine2.StrokeColor = contrastColor
	} else {
		// 如果未选中，使用透明填充，恢复边框
		r.box.FillColor = color.NRGBA{0, 0, 0, 0}
		r.box.StrokeWidth = 1                               // 未选中时显示边框
		r.box.StrokeColor = color.NRGBA{160, 160, 160, 200} // 确保未选中时边框是灰色
		r.checkLine1.Hidden = true
		r.checkLine2.Hidden = true
	}

	r.box.Refresh()
	r.checkLine1.Refresh()
	r.checkLine2.Refresh()
}

func (r *colorCheckRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *colorCheckRenderer) Destroy() {}

// 获取所有连接的ADB设备（包括虚拟屏）
func getADBDevices() ([]string, error) {
	// 解析输出
	lines := strings.Split(adbExec("devices"), "\n")
	var baseDevices []string

	// 跳过第一行（标题行）
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.Contains(line, "offline") {
			continue
		}

		// 设备ID在每行的开始
		parts := strings.Fields(line)
		if len(parts) >= 1 && parts[0] != "" {
			deviceID := parts[0]
			baseDevices = append(baseDevices, deviceID)
		}
	}

	// 构建最终设备列表（包含虚拟屏）
	var devices []string

	for _, deviceID := range baseDevices {
		// 确保 cap.dex 已推送到设备（仅对基础设备，首次出现时推送）
		ensureCapDexOnDevice(deviceID)

		// 先添加原始设备
		devices = append(devices, deviceID)

		// 尝试获取虚拟屏ID
		virtualDisplays := getVirtualDisplays(deviceID)
		for _, displayID := range virtualDisplays {
			// 添加虚拟屏格式：设备ID[虚拟屏ID]
			devices = append(devices, fmt.Sprintf("%s[%s]", deviceID, displayID))
		}
	}

	return devices, nil
}

// 获取设备的虚拟屏ID列表
func getVirtualDisplays(deviceID string) []string {
	// 执行命令获取虚拟屏ID
	output := adbExec("-s", deviceID, "shell", "app_process", "-Djava.class.path=/data/local/tmp/cap.dex", "/", "com.autogo.vdm.Main", "1")

	// 如果输出为空或包含错误信息，返回空列表
	if output == "" || strings.Contains(output, "Error") || strings.Contains(output, "error") ||
		strings.Contains(output, "Exception") || strings.Contains(output, "not found") {
		return nil
	}

	// 解析输出，每行一个虚拟屏ID
	var displayIDs []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 检查是否是有效的数字ID
		if line != "" && isNumeric(line) {
			displayIDs = append(displayIDs, line)
		}
	}

	return displayIDs
}

// 检查字符串是否为纯数字
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func captureScreenWithADB(deviceID string) (img image.Image, err error) {
	deviceTempPath := "/sdcard/screenshot_temp.png"

	// 解析设备ID，检查是否包含虚拟屏ID（格式：baseDeviceID[displayID]）
	baseDeviceID := deviceID
	virtualDisplayID := ""

	if idx := strings.Index(deviceID, "["); idx != -1 {
		if endIdx := strings.Index(deviceID, "]"); endIdx > idx {
			baseDeviceID = deviceID[:idx]
			virtualDisplayID = deviceID[idx+1 : endIdx]
		}
	}

	// 根据是否有虚拟屏ID选择截图方式
	if virtualDisplayID != "" {
		// 使用 app_process 进行虚拟屏截图
		if adbExec("-s", baseDeviceID, "shell", "app_process", "-Djava.class.path=/data/local/tmp/cap.dex", "/", "com.autogo.vdm.Main", "2", virtualDisplayID, deviceTempPath) != "" {
			return nil, fmt.Errorf("虚拟屏截图失败: %v", err)
		}
	} else {
		// 使用常规 screencap 命令
		adbExec("-s", baseDeviceID, "shell", "screencap", deviceTempPath)
	}

	defer func() {
		go adbExec("-s", baseDeviceID, "shell", "rm", deviceTempPath)
	}()

	localTempFile, err := ioutil.TempFile("", "screenshot_*.png")
	localTempPath := localTempFile.Name()
	localTempFile.Close()

	// 确保清理本地临时文件
	defer os.Remove(localTempPath)

	adbExec("-s", baseDeviceID, "pull", deviceTempPath, localTempPath)

	data, err := ioutil.ReadFile(localTempPath)
	if err != nil {
		return nil, fmt.Errorf("读取临时文件失败: %v", err)
	}

	// 解码图片
	return png.Decode(bytes.NewReader(data))
}

// 设备监控线程
func deviceMonitor() {
	for {
		updateDeviceList()
		time.Sleep(2 * time.Second)
	}
}

// 更新设备列表
func updateDeviceList() {
	if deviceSelect == nil {
		return
	}

	// 获取当前已选择的设备ID
	currentSelectedDevice := deviceSelect.Selected
	currentOptions := deviceSelect.Options

	// 获取新设备列表
	devices, err := getADBDevices()
	if err != nil {
		// 错误处理
		fyne.Do(func() {
			// 先清空选项列表和选择
			deviceSelect.Options = []string{}
			deviceSelect.Selected = "" // 直接设置 Selected 字段
			deviceSelect.SetSelected("")
			deviceSelect.PlaceHolder = "获取设备失败"
			deviceSelect.Refresh()
		})
		return
	}

	// 无设备情况处理
	if len(devices) == 0 {
		fyne.Do(func() {
			// 先清空选项列表和选择
			deviceSelect.Options = []string{}
			deviceSelect.Selected = "" // 直接设置 Selected 字段
			deviceSelect.SetSelected("")
			deviceSelect.PlaceHolder = "无设备连接"
			selectedDevice = ""
			deviceSelect.Refresh()
		})
		return
	}

	// 检查设备列表是否有变化（包括虚拟屏的变化）
	listChanged := len(devices) != len(currentOptions)
	if !listChanged {
		for i, dev := range devices {
			if i >= len(currentOptions) || dev != currentOptions[i] {
				listChanged = true
				break
			}
		}
	}

	// 检查当前选择的设备是否仍在列表中
	deviceStillExists := false
	for _, dev := range devices {
		if dev == currentSelectedDevice {
			deviceStillExists = true
			break
		}
	}

	// 只有当列表变化或需要更新选择时才更新UI
	if listChanged || !deviceStillExists {
		fyne.Do(func() {
			// 更新设备列表
			deviceSelect.Options = devices

			if deviceStillExists && currentSelectedDevice != "" {
				// 如果之前选择的设备仍然存在，保持选择
				deviceSelect.SetSelected(currentSelectedDevice)
				deviceSelect.PlaceHolder = "选择设备"
				selectedDevice = currentSelectedDevice
			} else {
				// 否则选择第一个设备
				deviceSelect.SetSelected(devices[0])
				deviceSelect.PlaceHolder = "选择设备"
				selectedDevice = devices[0]
			}
			deviceSelect.Refresh()
		})
	}
}

// 定义点标记结构体
type MarkPoint struct {
	X, Y  int
	Color color.Color
}

// 定义矩形标记结构体
type MarkRect struct {
	X1, Y1 int // 起点
	X2, Y2 int // 终点
	Color  color.Color
}

type imageDragMode int

const (
	imageDragNone imageDragMode = iota
	imageDragPan
	imageDragRange
	imageDragPoint
)

const (
	minImageZoom       float32 = 0.1
	maxImageZoom       float32 = 8
	zoomStepMultiplier float32 = 1.1
	maxRightMenuPoints         = 20
	defaultRangeText           = "0,0,0,0"
)

// 自定义图像查看器，支持显示图像和鼠标事件跟踪
type ImageViewer struct {
	widget.BaseWidget
	image              image.Image // 当前显示的图像
	originalImage      image.Image // 保存的原始图像
	rotationDegrees    int         // 当前旋转角度 (0, 90, 180, 270)
	zoomScale          float32     // 当前缩放倍率
	displayImage       *canvas.Image
	contextMenu        fyne.CanvasObject
	markPoints         []MarkPoint // 存储点标记
	markRects          []MarkRect  // 存储矩形标记
	mouseDownX         int         // 鼠标按下时的X坐标
	mouseDownY         int         // 鼠标按下时的Y坐标
	isDragging         bool        // 是否正在拖动
	dragMode           imageDragMode
	lastDragAbs        fyne.Position
	tempRect           *MarkRect // 临时矩形，用于拖动过程的显示
	lastMouseX         int       // 上次鼠标X坐标（用于检测真实移动）
	lastMouseY         int       // 上次鼠标Y坐标（用于检测真实移动）
	onMouseMove        func(x, y int)
	onMouseDown        func(x, y int)
	onMouseUp          func(x, y int)
	onRightClick       func(x, y int)
	onRangeModeChanged func(enabled bool)
	getGridParams      func() (cols, rows, spacing int, hasParams bool) // 获取点阵参数的回调
	scrollContainer    *container.Scroll                                // 滚动容器引用
	magnifier          *MagnifierWidget                                 // 放大镜引用
	manualRectSelected bool                                             // 是否手动框选了区域
	rangeSelectMode    bool                                             // 是否等待框选范围
	mouseInWidget      bool                                             // 鼠标是否在图片框上
	window             fyne.Window                                      // 窗口引用，用于获取窗口位置
}

// 全局颜色点列表，供右侧表格使用
var colorPoints []ColorPoint

// 全局标签页数据映射：tabItem -> TabData
var tabDataMap = make(map[*container.TabItem]*TabData)

// 当前活动的标签页
var currentTab *container.TabItem

// 全局刷新表格函数，会在main中设置
var refreshColorList func()

// 全局触发生成代码函数，会在main中设置
var triggerGenerateCode func()

// 获取找色模式，返回"多点找色"或"多点比色"
func getColorMode() string {
	if colorModeRadio == nil {
		return "找色" // 默认模式
	}

	selected := colorModeRadio.Selected
	if selected == "" {
		return "找色" // 未选择时返回默认模式
	}

	return selected
}

// 保存当前标签页的数据
func saveCurrentTabData() {
	if currentTab == nil || imageViewer == nil {
		return
	}

	tabData, exists := tabDataMap[currentTab]
	if !exists {
		tabData = &TabData{}
		tabDataMap[currentTab] = tabData
	}

	// 保存颜色点列表（深拷贝）
	tabData.colorPoints = make([]ColorPoint, len(colorPoints))
	copy(tabData.colorPoints, colorPoints)

	// 保存矩形标记数据（深拷贝）
	tabData.markRects = make([]MarkRect, len(imageViewer.markRects))
	copy(tabData.markRects, imageViewer.markRects)
	tabData.manualRectSelected = imageViewer.manualRectSelected

	// 保存生成的代码
	if codeDisplayEntry != nil {
		tabData.generatedCode = codeDisplayEntry.Text
	}

	// 保存图像查看器引用
	tabData.imageViewer = imageViewer
}

// 恢复标签页数据
func restoreTabData(tab *container.TabItem) {
	tabData, exists := tabDataMap[tab]
	if !exists || tabData.imageViewer == nil {
		// 如果是欢迎页或没有数据，清空
		colorPoints = make([]ColorPoint, 0)
		imageViewer = nil
		setRectCoordText(defaultRangeText)
		if codeDisplayEntry != nil {
			codeDisplayEntry.SetText("")
		}
		if refreshColorList != nil {
			refreshColorList()
		}
		return
	}

	// 恢复颜色点列表（深拷贝）
	colorPoints = make([]ColorPoint, len(tabData.colorPoints))
	copy(colorPoints, tabData.colorPoints)

	// 恢复图像查看器
	imageViewer = tabData.imageViewer

	// 恢复矩形标记数据（深拷贝）
	if imageViewer != nil {
		imageViewer.markRects = make([]MarkRect, len(tabData.markRects))
		copy(imageViewer.markRects, tabData.markRects)
		imageViewer.manualRectSelected = tabData.manualRectSelected

		// 更新矩形坐标显示
		if tabData.manualRectSelected && len(tabData.markRects) > 0 && rectCoordEntry != nil {
			// 使用最后一个矩形的坐标
			lastRect := tabData.markRects[len(tabData.markRects)-1]
			rectCoordEntry.SetText(fmt.Sprintf("%d,%d,%d,%d",
				lastRect.X1, lastRect.Y1, lastRect.X2, lastRect.Y2))
		} else if rectCoordEntry != nil {
			// 如果没有手动框选，根据标记点自动计算矩形范围
			autoRect := calculateAutoRect()
			if autoRect != "" {
				rectCoordEntry.SetText(autoRect)
			} else {
				rectCoordEntry.SetText(defaultRangeText)
			}
		}

		// 重新绘制图像（显示标记点）
		imageViewer.Refresh()
	}

	// 恢复生成的代码
	if codeDisplayEntry != nil {
		codeDisplayEntry.SetText(tabData.generatedCode)
	}

	// 刷新表格显示
	if refreshColorList != nil {
		refreshColorList()
	}
}

func parsePointPosition(pos string) (int, int, bool) {
	parts := strings.Split(strings.TrimSpace(pos), ",")
	if len(parts) != 2 {
		return 0, 0, false
	}

	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}
	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}
	return x, y, true
}

func colorHexAtImage(img image.Image, x, y int) (string, color.Color, bool) {
	if img == nil {
		return "", nil, false
	}

	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return "", nil, false
	}

	pixelColor := img.At(x, y)
	r, g, b, _ := pixelColor.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", uint8(r>>8), uint8(g>>8), uint8(b>>8)), pixelColor, true
}

func setRectCoordText(text string) {
	if rectCoordEntry != nil {
		rectCoordEntry.SetText(text)
	}
}

func setColorPointAt(index, x, y int) bool {
	if index < 0 || index >= maxRightMenuPoints || imageViewer == nil || imageViewer.image == nil {
		return false
	}

	hexColor, _, ok := colorHexAtImage(imageViewer.image, x, y)
	if !ok {
		return false
	}

	point := ColorPoint{
		ID:       index,
		Position: fmt.Sprintf("%d, %d", x, y),
		Color:    hexColor,
		Offset:   defaultColorPointOffset,
		Selected: true,
	}

	if index < len(colorPoints) {
		point.Offset = colorPoints[index].Offset
		point.Selected = colorPoints[index].Selected
		colorPoints[index] = point
	} else if index == len(colorPoints) {
		colorPoints = append(colorPoints, point)
	} else {
		return false
	}

	syncImageViewerMarksFromColorPoints()
	updateTableSelection()
	return true
}

func normalizeColorOffset(text string) (string, bool) {
	text = strings.TrimPrefix(strings.TrimSpace(text), "#")
	if len(text) != 6 {
		return "", false
	}

	for _, c := range text {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return "", false
		}
	}
	return strings.ToUpper(text), true
}

func refreshColorPointFromImage(index int) bool {
	if index < 0 || index >= len(colorPoints) || imageViewer == nil || imageViewer.image == nil {
		return false
	}

	x, y, ok := parsePointPosition(colorPoints[index].Position)
	if !ok {
		return false
	}

	hexColor, _, ok := colorHexAtImage(imageViewer.image, x, y)
	if !ok {
		return false
	}

	colorPoints[index].Color = hexColor
	return true
}

func refreshAllColorPointsFromImage() {
	for i := range colorPoints {
		refreshColorPointFromImage(i)
	}
}

func syncImageViewerMarksFromColorPoints() {
	if imageViewer == nil {
		return
	}

	imageViewer.markPoints = imageViewer.markPoints[:0]
	for _, point := range colorPoints {
		x, y, ok := parsePointPosition(point.Position)
		if !ok {
			continue
		}
		imageViewer.markPoints = append(imageViewer.markPoints, MarkPoint{
			X:     x,
			Y:     y,
			Color: getInverseColor(hexToColor(point.Color)),
		})
	}

	if !imageViewer.manualRectSelected {
		imageViewer.updateBoundingBox()
	}
	imageViewer.Refresh()
}

func commitColorPointPosition(index int, value string) {
	if index < 0 || index >= len(colorPoints) {
		return
	}

	x, y, ok := parsePointPosition(value)
	if !ok {
		updateTableSelection()
		return
	}

	if imageViewer == nil || imageViewer.image == nil {
		colorPoints[index].Position = fmt.Sprintf("%d, %d", x, y)
		updateTableSelection()
		return
	}

	hexColor, _, ok := colorHexAtImage(imageViewer.image, x, y)
	if !ok {
		updateTableSelection()
		return
	}

	colorPoints[index].Position = fmt.Sprintf("%d, %d", x, y)
	colorPoints[index].Color = hexColor
	syncImageViewerMarksFromColorPoints()
	updateTableSelection()
}

func colorPointCoordinatesText() string {
	positions := make([]string, 0, len(colorPoints))
	for _, point := range colorPoints {
		x, y, ok := parsePointPosition(point.Position)
		if !ok {
			continue
		}
		positions = append(positions, fmt.Sprintf("%d, %d", x, y))
	}
	return strings.Join(positions, "\n")
}

func parsePointPositionsText(text string) []image.Point {
	lines := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ';'
	})

	points := make([]image.Point, 0, len(lines))
	for _, line := range lines {
		x, y, ok := parsePointPosition(line)
		if !ok {
			continue
		}
		points = append(points, image.Pt(x, y))
	}
	return points
}

func replaceColorPointsByPositions(points []image.Point, offset string) {
	if imageViewer == nil || imageViewer.image == nil {
		return
	}

	colorPoints = colorPoints[:0]
	for _, point := range points {
		hexColor, _, ok := colorHexAtImage(imageViewer.image, point.X, point.Y)
		if !ok {
			continue
		}

		colorPoints = append(colorPoints, ColorPoint{
			ID:       len(colorPoints),
			Position: fmt.Sprintf("%d, %d", point.X, point.Y),
			Color:    hexColor,
			Offset:   offset,
			Selected: true,
		})
	}

	syncImageViewerMarksFromColorPoints()
	updateTableSelection()
}

// 根据颜色点自动计算矩形范围
func calculateAutoRect() string {
	if len(colorPoints) == 0 {
		return ""
	}

	// 找到所有点的最小和最大坐标
	minX, minY := int(^uint(0)>>1), int(^uint(0)>>1) // int 最大值
	maxX, maxY := 0, 0
	validCount := 0

	for _, point := range colorPoints {
		x, y, ok := parsePointPosition(point.Position)
		if !ok {
			continue
		}
		validCount++

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
	if validCount == 0 {
		return ""
	}

	// 如果只有一个点，创建一个小的矩形范围
	if minX == maxX && minY == maxY {
		// 以该点为中心，创建一个10x10的矩形
		minX -= 5
		minY -= 5
		maxX += 5
		maxY += 5
	}

	// 添加一些边距（10像素）
	minX -= 10
	minY -= 10
	maxX += 10
	maxY += 10

	// 确保坐标不为负数
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}

	// 如果有图像，确保不超出图像范围
	if imageViewer != nil && imageViewer.image != nil {
		bounds := imageViewer.image.Bounds()
		if maxX > bounds.Max.X {
			maxX = bounds.Max.X
		}
		if maxY > bounds.Max.Y {
			maxY = bounds.Max.Y
		}
	}

	return fmt.Sprintf("%d,%d,%d,%d", minX, minY, maxX, maxY)
}

// 获取偏色值，返回字符串（十六进制格式），如果输入无效则返回默认值"101010"
func getColorOffset() string {
	if colorOffsetEntry == nil {
		return "101010" // 默认值
	}

	text := strings.TrimSpace(colorOffsetEntry.Text)
	if text == "" {
		return "101010" // 空值时返回默认值
	}

	// 移除可能的#前缀
	text = strings.TrimPrefix(text, "#")

	// 验证是否为有效的十六进制字符串（6位）
	if offset, ok := normalizeColorOffset(text); ok {
		return offset // 返回大写格式
	}

	return "101010" // 长度不对，返回默认值
}

func colorPointValue(point ColorPoint) string {
	colorValue := strings.ToLower(strings.TrimPrefix(point.Color, "#"))
	if offset, ok := normalizeColorOffset(point.Offset); ok {
		return fmt.Sprintf("%s-%s", colorValue, offset)
	}
	return colorValue
}

func selectedColorPoints() []ColorPoint {
	points := make([]ColorPoint, 0, len(colorPoints))
	for _, point := range colorPoints {
		if point.Selected {
			points = append(points, point)
		}
	}
	return points
}

func regionValuesFromEntry() (int, int, int, int) {
	if rectCoordEntry == nil {
		return 0, 0, 0, 0
	}

	parts := strings.Split(strings.TrimSpace(rectCoordEntry.Text), ",")
	if len(parts) != 4 {
		return 0, 0, 0, 0
	}

	values := [4]int{}
	for i, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return 0, 0, 0, 0
		}
		values[i] = value
	}
	return values[0], values[1], values[2], values[3]
}

func directionValue(directionText string) int {
	if strings.HasPrefix(strings.TrimSpace(directionText), "1") {
		return 1
	}
	if strings.HasPrefix(strings.TrimSpace(directionText), "2") {
		return 2
	}
	if strings.HasPrefix(strings.TrimSpace(directionText), "3") {
		return 3
	}
	return 0
}

func apiColorAlternatives(points []ColorPoint) string {
	colors := make([]string, 0, len(points))
	for _, point := range points {
		colors = append(colors, colorPointValue(point))
	}
	return strings.Join(colors, "|")
}

func apiMultiColorTemplate(points []ColorPoint) string {
	if len(points) == 0 {
		return ""
	}

	baseX, baseY, ok := parsePointPosition(points[0].Position)
	if !ok {
		return ""
	}

	parts := []string{colorPointValue(points[0])}
	for _, point := range points[1:] {
		x, y, ok := parsePointPosition(point.Position)
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d,%d,%s", x-baseX, y-baseY, colorPointValue(point)))
	}
	return strings.Join(parts, ",")
}

func apiRegionParamsObject(x1, y1, x2, y2 int, colorText, sim string, dir int) string {
	return fmt.Sprintf("{%d,%d,%d,%d,\"%s\",%s,%d,0}", x1, y1, x2, y2, colorText, sim, dir)
}

func defaultAPIFormatTemplates() map[string]string {
	return map[string]string{
		"FindColor":          "x, y := images.FindColor([参数])",
		"FindMultiColors":    "x, y := images.FindMultiColors([参数])",
		"FindMultiColorsAll": "points := images.FindMultiColorsAll([参数])",
		"CmpColor":           "matched := images.CmpColor([参数])",
	}
}

func copyAPIFormatTemplates(templates map[string]string) map[string]string {
	copied := make(map[string]string, len(templates))
	for key, value := range templates {
		copied[key] = value
	}
	return copied
}

func normalizeAPIFormatTemplates(templates map[string]string) map[string]string {
	normalized := defaultAPIFormatTemplates()
	for key, value := range templates {
		name := normalizeImagesFunctionName(key)
		if strings.TrimSpace(value) != "" {
			normalized[name] = value
		}
	}
	return normalized
}

func formatTemplateFor(functionName string) string {
	functionName = normalizeImagesFunctionName(functionName)
	template := strings.TrimSpace(apiFormatTemplates[functionName])
	if template != "" {
		return apiFormatTemplates[functionName]
	}
	return defaultAPIFormatTemplates()[functionName]
}

func applyFormatTemplate(template string, values map[string]string) string {
	result := template
	for _, placeholder := range apiFormatPlaceholders {
		result = strings.ReplaceAll(result, placeholder.token, values[placeholder.token])
	}
	return result
}

func renderImageAPICode(functionName string, values map[string]string) string {
	return applyFormatTemplate(formatTemplateFor(functionName), values)
}

func apiFormatValues(functionName, params, colorParams, colorText, sim string, dir, x1, y1, x2, y2, pointX, pointY int) map[string]string {
	return map[string]string{
		"[函数名]":      functionName,
		"[参数]":       params,
		"[颜色参数]":     colorParams,
		"[颜色值]":      colorText,
		"[相似度]":      sim,
		"[查找方向]":     strconv.Itoa(dir),
		"[屏幕ID]":     "0",
		"[范围_左]":     strconv.Itoa(x1),
		"[范围_上]":     strconv.Itoa(y1),
		"[范围_右]":     strconv.Itoa(x2),
		"[范围_下]":     strconv.Itoa(y2),
		"[坐标_X]":     strconv.Itoa(pointX),
		"[坐标_Y]":     strconv.Itoa(pointY),
		"[区域_左上]":    fmt.Sprintf("%d, %d", x1, y1),
		"[区域_右下]":    fmt.Sprintf("%d, %d", x2, y2),
		"[CmpColor]": fmt.Sprintf("%d, %d, \"%s\", %s, 0", pointX, pointY, colorText, sim),
	}
}

type apiFormatPlaceholder struct {
	token       string
	description string
}

var apiFormatPlaceholders = []apiFormatPlaceholder{
	{token: "[参数]", description: "完整函数参数"},
	{token: "[颜色参数]", description: "右侧颜色字段参数对象"},
	{token: "[颜色值]", description: "colorStr / colors"},
	{token: "[相似度]", description: "sim"},
	{token: "[查找方向]", description: "dir"},
	{token: "[屏幕ID]", description: "displayId"},
	{token: "[范围_左]", description: "x1"},
	{token: "[范围_上]", description: "y1"},
	{token: "[范围_右]", description: "x2"},
	{token: "[范围_下]", description: "y2"},
	{token: "[坐标_X]", description: "x"},
	{token: "[坐标_Y]", description: "y"},
	{token: "[函数名]", description: "函数名称"},
}

func sampleAPIFormatValues(functionName string) map[string]string {
	functionName = normalizeImagesFunctionName(functionName)
	switch functionName {
	case "FindColor":
		colorText := "FFFFFF|CCCCCC-101010"
		params := "0, 0, 0, 0, \"FFFFFF|CCCCCC-101010\", 0.9, 0, 0"
		return apiFormatValues(functionName, params, apiRegionParamsObject(0, 0, 0, 0, colorText, "0.9", 0), colorText, "0.9", 0, 0, 0, 0, 0, 100, 200)
	case "FindMultiColorsAll":
		colorText := "ffccff-151515,635,978,ffab2d-101010"
		params := "0, 0, 0, 0, \"ffccff-151515,635,978,ffab2d-101010\", 0.9, 0, 0"
		return apiFormatValues(functionName, params, apiRegionParamsObject(0, 0, 0, 0, colorText, "0.9", 0), colorText, "0.9", 0, 0, 0, 0, 0, 100, 200)
	case "CmpColor":
		colorText := "FFFFFF|CCCCCC-101010"
		params := "100, 200, \"FFFFFF|CCCCCC-101010\", 0.9, 0"
		return apiFormatValues(functionName, params, fmt.Sprintf("{100,200,\"%s\",0.9,0}", colorText), colorText, "0.9", 0, 0, 0, 0, 0, 100, 200)
	default:
		colorText := "ffccff-151515,635,978,ffab2d-101010"
		params := "0, 0, 0, 0, \"ffccff-151515,635,978,ffab2d-101010\", 0.9, 0, 0"
		return apiFormatValues("FindMultiColors", params, apiRegionParamsObject(0, 0, 0, 0, colorText, "0.9", 0), colorText, "0.9", 0, 0, 0, 0, 0, 100, 200)
	}
}

func showAPIFormatDialog(parent fyne.Window, selectedFunction string, saveConfig func()) {
	functions := []string{"FindMultiColors", "FindColor", "FindMultiColorsAll", "CmpColor"}
	localTemplates := copyAPIFormatTemplates(normalizeAPIFormatTemplates(apiFormatTemplates))
	defaultTemplates := defaultAPIFormatTemplates()
	currentFunction := normalizeImagesFunctionName(selectedFunction)

	templateEntry := widget.NewMultiLineEntry()
	templateEntry.Wrapping = fyne.TextWrapWord
	templateEntry.SetMinRowsVisible(8)

	previewEntry := widget.NewMultiLineEntry()
	previewEntry.Wrapping = fyne.TextWrapWord
	previewEntry.SetMinRowsVisible(7)
	previewEntry.Disable()

	refreshPreview := func() {
		localTemplates[currentFunction] = templateEntry.Text
		previewEntry.SetText(applyFormatTemplate(templateEntry.Text, sampleAPIFormatValues(currentFunction)))
	}
	templateEntry.OnChanged = func(string) {
		refreshPreview()
	}

	methodSelect := widget.NewSelect(functions, func(value string) {
		currentFunction = normalizeImagesFunctionName(value)
		templateEntry.SetText(localTemplates[currentFunction])
		refreshPreview()
	})

	placeholderButtons := container.NewVBox()
	for _, placeholder := range apiFormatPlaceholders {
		token := placeholder.token
		label := widget.NewLabel(placeholder.description)
		label.Wrapping = fyne.TextTruncate
		placeholderButtons.Add(container.NewBorder(nil, nil, widget.NewButton(token, func() {
			templateEntry.SetText(templateEntry.Text + token)
		}), nil, label))
	}

	leftPanel := container.NewBorder(
		container.NewVBox(
			container.NewBorder(nil, nil, widget.NewLabel("模板方法："), nil, methodSelect),
			widget.NewSeparator(),
			widget.NewLabel("模板参数"),
		),
		nil,
		nil,
		nil,
		container.NewVScroll(placeholderButtons),
	)

	templateBlock := container.NewBorder(widget.NewLabel("模板内容"), nil, nil, nil, templateEntry)
	previewBlock := container.NewBorder(widget.NewLabel("实时预览"), nil, nil, nil, previewEntry)
	rightSplit := container.NewVSplit(templateBlock, previewBlock)
	rightSplit.Offset = 0.55

	bodySplit := container.NewHSplit(container.New(&fixedWidthLayout{width: 230}, leftPanel), rightSplit)
	bodySplit.Offset = 0.34

	var formatDialog *dialog.CustomDialog
	restoreButton := widget.NewButton("还原配置", func() {
		localTemplates[currentFunction] = defaultTemplates[currentFunction]
		templateEntry.SetText(localTemplates[currentFunction])
		refreshPreview()
	})
	closeButton := widget.NewButton("关闭", func() {
		if formatDialog != nil {
			formatDialog.Hide()
		}
	})
	saveButton := widget.NewButton("保存配置", func() {
		localTemplates[currentFunction] = templateEntry.Text
		apiFormatTemplates = copyAPIFormatTemplates(normalizeAPIFormatTemplates(localTemplates))
		if saveConfig != nil {
			saveConfig()
		}
		refreshImagesAPIFields()
		if formatDialog != nil {
			formatDialog.Hide()
		}
	})
	saveButton.Importance = widget.HighImportance

	content := container.NewBorder(nil, container.NewHBox(layout.NewSpacer(), restoreButton, closeButton, saveButton), nil, nil, bodySplit)
	formatDialog = dialog.NewCustomWithoutButtons("自定义参数格式", content, parent)
	formatDialog.Resize(fyne.NewSize(720, 520))

	methodSelect.SetSelected(currentFunction)
	formatDialog.Show()
}

func refreshImagesAPIFields() {
	if updateImagesAPIFields != nil {
		updateImagesAPIFields()
	}
}

func normalizeImagesFunctionName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "findcolor":
		return "FindColor"
	case "findmulticolors", "findmulticolor":
		return "FindMultiColors"
	case "findmulticolorsall":
		return "FindMultiColorsAll"
	case "cmpcolor":
		return "CmpColor"
	default:
		return "FindMultiColors"
	}
}

func buildImagesAPICode(functionName, precisionText, directionText string) (string, string, string) {
	points := selectedColorPoints()
	if len(points) == 0 {
		return "", "", ""
	}

	sim := strings.TrimSpace(precisionText)
	if sim == "" {
		sim = "0.9"
	}
	dir := directionValue(directionText)
	x1, y1, x2, y2 := regionValuesFromEntry()
	functionName = normalizeImagesFunctionName(functionName)

	switch functionName {
	case "FindColor":
		colorText := apiColorAlternatives(points)
		colorParams := apiRegionParamsObject(x1, y1, x2, y2, colorText, sim, dir)
		params := fmt.Sprintf("%d, %d, %d, %d, \"%s\", %s, %d, 0", x1, y1, x2, y2, colorText, sim, dir)
		values := apiFormatValues(functionName, params, colorParams, colorText, sim, dir, x1, y1, x2, y2, 0, 0)
		return colorParams, params, renderImageAPICode(functionName, values)
	case "FindMultiColorsAll":
		colorText := apiMultiColorTemplate(points)
		colorParams := apiRegionParamsObject(x1, y1, x2, y2, colorText, sim, dir)
		params := fmt.Sprintf("%d, %d, %d, %d, \"%s\", %s, %d, 0", x1, y1, x2, y2, colorText, sim, dir)
		values := apiFormatValues(functionName, params, colorParams, colorText, sim, dir, x1, y1, x2, y2, 0, 0)
		return colorParams, params, renderImageAPICode(functionName, values)
	case "CmpColor":
		x, y, ok := parsePointPosition(points[0].Position)
		if !ok {
			return "", "", ""
		}
		colorText := apiColorAlternatives(points)
		colorParams := fmt.Sprintf("{%d,%d,\"%s\",%s,0}", x, y, colorText, sim)
		params := fmt.Sprintf("%d, %d, \"%s\", %s, 0", x, y, colorText, sim)
		values := apiFormatValues(functionName, params, colorParams, colorText, sim, dir, x1, y1, x2, y2, x, y)
		return colorParams, params, renderImageAPICode(functionName, values)
	default:
		colorText := apiMultiColorTemplate(points)
		colorParams := apiRegionParamsObject(x1, y1, x2, y2, colorText, sim, dir)
		params := fmt.Sprintf("%d, %d, %d, %d, \"%s\", %s, %d, 0", x1, y1, x2, y2, colorText, sim, dir)
		values := apiFormatValues(functionName, params, colorParams, colorText, sim, dir, x1, y1, x2, y2, 0, 0)
		return colorParams, params, renderImageAPICode(functionName, values)
	}
}

// 颜色匹配辅助函数
// 生成找色代码
func generateColorCode() string {
	// 检查是否有颜色点
	if len(colorPoints) == 0 {
		return ""
	}

	// 过滤出勾选的颜色点
	var selectedPoints []ColorPoint
	for _, point := range colorPoints {
		if point.Selected {
			selectedPoints = append(selectedPoints, point)
		}
	}

	// 检查是否有勾选的点
	if len(selectedPoints) == 0 {
		return ""
	}

	// 获取找色模式
	mode := getColorMode()

	if mode == "比色" {
		// 多点比色格式: "x1,y1,color1-offset,x2,y2,color2-offset,..."
		var parts []string
		for _, point := range selectedPoints {
			x, y, ok := parsePointPosition(point.Position)
			if !ok {
				continue
			}

			// 组合: x,y,color 或 x,y,color-offset
			parts = append(parts, fmt.Sprintf("%d,%d,%s", x, y, colorPointValue(point)))
		}

		colorStr := strings.Join(parts, ",")
		return fmt.Sprintf("images.DetectsMultiColors(\"%s\", 0.9, 0)", colorStr)

	} else {
		// 多点找色格式: 区域 + "color1-offset,dx2,dy2,color2-offset,dx3,dy3,color3-offset,..."

		// 获取区域坐标
		rectText := ""
		if rectCoordEntry != nil {
			rectText = strings.TrimSpace(rectCoordEntry.Text)
		}

		// 解析区域坐标 "x1,y1,x2,y2"
		x1, y1, x2, y2 := "0", "0", "0", "0"
		if rectText != "" {
			coords := strings.Split(rectText, ",")
			if len(coords) == 4 {
				x1 = strings.TrimSpace(coords[0])
				y1 = strings.TrimSpace(coords[1])
				x2 = strings.TrimSpace(coords[2])
				y2 = strings.TrimSpace(coords[3])
			}
		}

		// 获取第一个勾选的点作为基准点
		firstPoint := selectedPoints[0]
		baseX, baseY, ok := parsePointPosition(firstPoint.Position)
		if !ok {
			return ""
		}

		var parts []string
		parts = append(parts, colorPointValue(firstPoint))

		// 后续点使用相对坐标
		for i := 1; i < len(selectedPoints); i++ {
			point := selectedPoints[i]
			x, y, ok := parsePointPosition(point.Position)
			if !ok {
				continue
			}

			// 计算相对坐标
			dx := x - baseX
			dy := y - baseY

			// 组合: dx,dy,color 或 dx,dy,color-offset
			parts = append(parts, fmt.Sprintf("%d,%d,%s", dx, dy, colorPointValue(point)))
		}

		colorStr := strings.Join(parts, ",")
		return fmt.Sprintf("images.FindMultiColors(%s, %s, %s, %s, \"%s\", 0.9, 0, 0)", x1, y1, x2, y2, colorStr)
	}
}

// 添加点到颜色列表
func addColorPointToList(x, y int, colorHex string, selected bool) {
	// 创建新的颜色点
	newPoint := ColorPoint{
		ID:       len(colorPoints),
		Position: fmt.Sprintf("%d, %d", x, y),
		Color:    colorHex,
		Offset:   defaultColorPointOffset,
		Selected: selected,
	}

	// 添加到列表
	colorPoints = append(colorPoints, newPoint)

	// 刷新表格显示
	if refreshColorList != nil {
		refreshColorList()
	}

	// 如果当前没有手动框选矩形，自动更新矩形范围
	if imageViewer != nil && !imageViewer.manualRectSelected && rectCoordEntry != nil {
		autoRect := calculateAutoRect()
		if autoRect != "" {
			rectCoordEntry.SetText(autoRect)
		}
	}
}

// 添加矩形标记 - 确保只有一个矩形
func (v *ImageViewer) AddRect(x1, y1, x2, y2 int, c color.Color) {
	// 清空现有的所有矩形
	v.markRects = v.markRects[:0]

	// 添加新矩形
	v.markRects = append(v.markRects, MarkRect{X1: x1, Y1: y1, X2: x2, Y2: y2, Color: c})

	// 标记为手动框选
	v.manualRectSelected = true

	// 更新坐标显示框
	if rectCoordEntry != nil {
		// 确保坐标是左上角到右下角的顺序
		minX := min(x1, x2)
		minY := min(y1, y2)
		maxX := max(x1, x2)
		maxY := max(y1, y2)
		rectText := fmt.Sprintf("%d,%d,%d,%d", minX, minY, maxX, maxY)

		// 更新编辑框文本，仅显示四个坐标
		rectCoordEntry.SetText(rectText)
		if autoCopyRangeEnabled && v.window != nil {
			v.window.Clipboard().SetContent(rectText)
		}
	}

	v.Refresh() // 刷新视图以显示新矩形
}

// 更新区域坐标为所有点的外包围盒
func (v *ImageViewer) updateBoundingBox() {
	// 如果是手动框选的区域，不自动更新
	if v.manualRectSelected {
		return
	}

	// 如果没有点，不更新
	if len(v.markPoints) == 0 {
		return
	}

	// 找出所有点的最小和最大坐标
	minX := v.markPoints[0].X
	minY := v.markPoints[0].Y
	maxX := v.markPoints[0].X
	maxY := v.markPoints[0].Y

	for _, point := range v.markPoints[1:] {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	// 向外扩展10像素
	minX -= 10
	minY -= 10
	maxX += 10
	maxY += 10

	// 确保不超出图像边界
	if v.image != nil {
		bounds := v.image.Bounds()
		if minX < bounds.Min.X {
			minX = bounds.Min.X
		}
		if minY < bounds.Min.Y {
			minY = bounds.Min.Y
		}
		if maxX > bounds.Max.X {
			maxX = bounds.Max.X
		}
		if maxY > bounds.Max.Y {
			maxY = bounds.Max.Y
		}
	}

	// 更新坐标显示框
	if rectCoordEntry != nil {
		rectCoordEntry.SetText(fmt.Sprintf("%d,%d,%d,%d", minX, minY, maxX, maxY))
	}
}

// 清除所有标记
func (v *ImageViewer) ClearMarks() {
	v.markPoints = v.markPoints[:0]
	v.markRects = v.markRects[:0]

	// 重置手动框选标志
	v.manualRectSelected = false

	// 清空颜色点列表
	colorPoints = colorPoints[:0]

	// 使用fyne.Do确保在主线程中更新UI
	fyne.Do(func() {
		// 清空坐标显示框
		setRectCoordText(defaultRangeText)

		// 刷新表格显示
		if refreshColorList != nil {
			refreshColorList()
		}

		v.Refresh()
	})
}

// 设置左键按下回调函数
func (v *ImageViewer) SetOnMouseDown(callback func(x, y int)) {
	v.onMouseDown = callback
}

// 设置左键弹起回调函数
func (v *ImageViewer) SetOnMouseUp(callback func(x, y int)) {
	v.onMouseUp = callback
}

// 设置右键点击回调函数
func (v *ImageViewer) SetOnRightClick(callback func(x, y int)) {
	v.onRightClick = callback
}

func (v *ImageViewer) SetRangeSelectMode(enabled bool) {
	if v.rangeSelectMode == enabled {
		return
	}
	v.rangeSelectMode = enabled
	if !enabled {
		v.tempRect = nil
		v.dragMode = imageDragNone
	}
	if v.onRangeModeChanged != nil {
		v.onRangeModeChanged(enabled)
	}
	v.Refresh()
}

func (v *ImageViewer) hideContextMenu() {
	if v.contextMenu == nil {
		return
	}
	v.contextMenu = nil
	v.Refresh()
}

func (v *ImageViewer) currentZoomScale() float32 {
	if v.zoomScale <= 0 {
		return 1
	}
	return v.zoomScale
}

func (v *ImageViewer) imagePositionFromView(pos fyne.Position) (int, int, bool) {
	if v.image == nil {
		return 0, 0, false
	}

	scale := v.currentZoomScale()
	bounds := v.image.Bounds()
	x := bounds.Min.X + int(pos.X/scale)
	y := bounds.Min.Y + int(pos.Y/scale)
	return v.clampImagePosition(x, y)
}

func (v *ImageViewer) clampImagePosition(x, y int) (int, int, bool) {
	if v.image == nil {
		return 0, 0, false
	}

	bounds := v.image.Bounds()
	if x < bounds.Min.X {
		x = bounds.Min.X
	}
	if x >= bounds.Max.X {
		x = bounds.Max.X - 1
	}
	if y < bounds.Min.Y {
		y = bounds.Min.Y
	}
	if y >= bounds.Max.Y {
		y = bounds.Max.Y - 1
	}
	return x, y, true
}

func (v *ImageViewer) addPointAt(x, y int) {
	if v.getGridParams != nil {
		cols, rows, spacing, hasParams := v.getGridParams()
		if hasParams && cols > 0 && rows > 0 && spacing > 0 {
			v.AddGridPoints(x, y, cols, rows, spacing)
			return
		}
	}
	v.AddPoint(x, y, nil)
}

func (v *ImageViewer) panBy(delta fyne.Position) {
	if v.scrollContainer == nil {
		return
	}

	offset := v.scrollContainer.Offset
	v.scrollContainer.ScrollToOffset(fyne.NewPos(offset.X-delta.X, offset.Y-delta.Y))
}

func (v *ImageViewer) applyZoom(scale float32) {
	if v.image == nil {
		return
	}
	if scale < minImageZoom {
		scale = minImageZoom
	}
	if scale > maxImageZoom {
		scale = maxImageZoom
	}
	v.zoomScale = scale
	v.Refresh()
	if v.scrollContainer != nil {
		v.scrollContainer.Refresh()
		v.scrollContainer.ScrollToOffset(fyne.NewPos(0, 0))
	}
}

func (v *ImageViewer) FitToView() {
	if v.image == nil || v.scrollContainer == nil {
		return
	}

	bounds := v.image.Bounds()
	imgW := float32(bounds.Dx())
	imgH := float32(bounds.Dy())
	viewSize := v.scrollContainer.Size()
	if imgW <= 0 || imgH <= 0 || viewSize.Width <= 0 || viewSize.Height <= 0 {
		return
	}

	scale := fyne.Min(viewSize.Width/imgW, viewSize.Height/imgH)
	if scale > 1 {
		scale = 1
	}
	v.applyZoom(scale)
}

func (v *ImageViewer) ShowOriginalSize() {
	v.applyZoom(1)
}

func (v *ImageViewer) ResetRangeSelection() {
	v.markRects = v.markRects[:0]
	v.manualRectSelected = false
	v.tempRect = nil
	v.SetRangeSelectMode(false)
	setRectCoordText(defaultRangeText)
	v.Refresh()
}

func (v *ImageViewer) zoomAt(pos fyne.Position, zoomIn bool) {
	if v.image == nil {
		return
	}

	oldScale := v.currentZoomScale()
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

	scrollOffset := fyne.NewPos(0, 0)
	if v.scrollContainer != nil {
		scrollOffset = v.scrollContainer.Offset
	}
	viewPos := pos.Subtract(scrollOffset)
	imageX := pos.X / oldScale
	imageY := pos.Y / oldScale

	v.zoomScale = newScale
	v.Refresh()

	if v.scrollContainer != nil {
		v.scrollContainer.Refresh()
		v.scrollContainer.ScrollToOffset(fyne.NewPos(imageX*newScale-viewPos.X, imageY*newScale-viewPos.Y))
	}
}

// 实现MouseDown方法，处理左键按下
func (v *ImageViewer) MouseDown(e *desktop.MouseEvent) {
	if v.image == nil || e.Button != desktop.MouseButtonPrimary {
		return
	}
	v.hideContextMenu()

	mouseX, mouseY, ok := v.imagePositionFromView(e.Position)
	if !ok {
		return
	}

	// 记录按下的坐标
	v.mouseDownX = mouseX
	v.mouseDownY = mouseY
	v.lastDragAbs = e.AbsolutePosition
	v.isDragging = true
	v.dragMode = imageDragPan
	if v.rangeSelectMode {
		v.dragMode = imageDragRange
	} else if e.Modifier&fyne.KeyModifierControl != 0 {
		v.dragMode = imageDragPoint
	}

	// 清除任何现有的临时矩形
	v.tempRect = nil

	// 鼠标按下时隐藏放大镜
	if v.magnifier != nil {
		v.magnifier.Hide()
	}

	// 调用回调函数
	if v.onMouseDown != nil {
		v.onMouseDown(mouseX, mouseY)
	}
}

// 实现MouseUp方法，处理左键弹起
func (v *ImageViewer) MouseUp(e *desktop.MouseEvent) {
	if v.image == nil || e.Button != desktop.MouseButtonPrimary || !v.isDragging {
		return
	}

	mouseX, mouseY, ok := v.imagePositionFromView(e.Position)
	if !ok {
		return
	}

	// 计算按下和弹起位置的距离
	dist := distance(v.mouseDownX, v.mouseDownY, mouseX, mouseY)

	switch v.dragMode {
	case imageDragRange:
		if dist > 4 && v.tempRect != nil {
			v.AddRect(v.tempRect.X1, v.tempRect.Y1, v.tempRect.X2, v.tempRect.Y2, v.tempRect.Color)
		}
		v.SetRangeSelectMode(false)
	case imageDragPoint:
		if dist <= 4 {
			v.addPointAt(mouseX, mouseY)
		}
	}

	// 清除临时矩形
	v.tempRect = nil

	// 重置拖动状态
	v.isDragging = false
	v.dragMode = imageDragNone

	// 鼠标弹起时重新显示放大镜
	if magnifierEnabled && v.magnifier != nil {
		v.magnifier.Show()
	}

	// 刷新视图
	v.Refresh()

	// 调用回调函数
	if v.onMouseUp != nil {
		v.onMouseUp(mouseX, mouseY)
	}
}

// 实现TappedSecondary接口方法，处理右键点击
func (v *ImageViewer) TappedSecondary(e *fyne.PointEvent) {
	if v.image == nil || v.window == nil {
		return
	}

	mouseX, mouseY, ok := v.imagePositionFromView(e.Position)
	if !ok {
		return
	}
	colorHex, _, _ := colorHexAtImage(v.image, mouseX, mouseY)

	menuContent := container.NewVBox()
	addMenuButton := func(text string, enabled bool, action func()) {
		btn := widget.NewButton(text, func() {
			v.hideContextMenu()
			action()
		})
		if !enabled {
			btn.Disable()
		}
		menuContent.Add(btn)
	}

	addMenuButton("复制当前坐标", true, func() {
		v.window.Clipboard().SetContent(fmt.Sprintf("%d, %d", mouseX, mouseY))
	})
	addMenuButton("复制当前颜色", true, func() {
		v.window.Clipboard().SetContent(colorHex)
	})
	addMenuButton("清除所有选点", true, func() {
		v.ClearMarks()
	})
	menuContent.Add(widget.NewSeparator())

	for i := 0; i < maxRightMenuPoints; i++ {
		index := i
		addMenuButton(fmt.Sprintf("添加到点%d", index+1), index <= len(colorPoints), func() {
			setColorPointAt(index, mouseX, mouseY)
		})
	}
	menuSize := fyne.NewSize(170, 300)
	menuScroll := container.NewVScroll(menuContent)
	menuScroll.SetMinSize(menuSize)
	bg := canvas.NewRectangle(theme.BackgroundColor())
	bg.StrokeColor = theme.ShadowColor()
	bg.StrokeWidth = 1
	menu := container.NewStack(bg, menuScroll)

	pos := e.Position
	viewerSize := v.Size()
	if pos.X+menuSize.Width > viewerSize.Width {
		pos.X = viewerSize.Width - menuSize.Width
	}
	if pos.Y+menuSize.Height > viewerSize.Height {
		pos.Y = viewerSize.Height - menuSize.Height
	}
	if pos.X < 0 {
		pos.X = 0
	}
	if pos.Y < 0 {
		pos.Y = 0
	}
	menu.Move(pos)
	menu.Resize(menuSize)
	v.contextMenu = menu
	v.Refresh()

	// 调用回调函数
	if v.onRightClick != nil {
		v.onRightClick(mouseX, mouseY)
	}
}

func (v *ImageViewer) Scrolled(e *fyne.ScrollEvent) {
	driver, ok := fyne.CurrentApp().Driver().(desktop.Driver)
	if !ok || driver.CurrentKeyModifiers()&fyne.KeyModifierControl == 0 {
		if v.scrollContainer != nil {
			v.scrollContainer.Scrolled(e)
		}
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
}

// 设置图像
func (v *ImageViewer) SetImage(img image.Image) {
	v.image = img
	v.originalImage = img // 保存原始图像
	v.rotationDegrees = 0 // 重置旋转角度
	v.zoomScale = 1
	v.displayImage.Image = img
	v.Refresh()
}

// 旋转图像到指定角度
func (v *ImageViewer) RotateImage(degrees int) {
	if v.originalImage == nil {
		return
	}

	// 确保角度在0-359范围内
	degrees = degrees % 360

	// 如果角度是0，直接使用原图
	if degrees == 0 {
		v.image = v.originalImage
		v.displayImage.Image = v.image
		v.rotationDegrees = degrees
		v.Refresh()
		return
	}

	// 获取原始图像的尺寸
	bounds := v.originalImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var rotated *image.NRGBA // 使用NRGBA而不是RGBA

	// 根据角度执行不同的旋转
	switch degrees {
	case 90:
		// 创建新的画布，宽高互换
		rotated = image.NewNRGBA(image.Rect(0, 0, height, width))

		// 执行90度旋转 (x,y) -> (height-y-1, x)
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				rotated.Set(height-y-1, x, v.originalImage.At(x, y))
			}
		}

	case 180:
		// 创建新的画布，宽高保持不变
		rotated = image.NewNRGBA(image.Rect(0, 0, width, height))

		// 执行180度旋转 (x,y) -> (width-x-1, height-y-1)
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				rotated.Set(width-x-1, height-y-1, v.originalImage.At(x, y))
			}
		}

	case 270:
		// 创建新的画布，宽高互换
		rotated = image.NewNRGBA(image.Rect(0, 0, height, width))

		// 执行270度旋转 (x,y) -> (y, width-x-1)
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				rotated.Set(y, width-x-1, v.originalImage.At(x, y))
			}
		}
	}

	v.image = rotated
	v.displayImage.Image = rotated
	v.rotationDegrees = degrees

	// 清除所有绘制的标记
	v.ClearMarks()

	v.Refresh()
}

// 实现CreateRenderer方法
func (v *ImageViewer) CreateRenderer() fyne.WidgetRenderer {
	// 创建一个透明的临时矩形对象
	tempRectObj := canvas.NewRectangle(color.RGBA{255, 0, 0, 255})
	tempRectObj.StrokeWidth = 1
	tempRectObj.StrokeColor = color.RGBA{255, 0, 0, 255}
	tempRectObj.FillColor = color.RGBA{0, 0, 0, 0} // 透明填充
	tempRectObj.Hide()                             // 初始时隐藏

	return &imageViewerRenderer{
		viewer:   v,
		objects:  []fyne.CanvasObject{v.displayImage, tempRectObj},
		points:   []fyne.CanvasObject{},
		rects:    []fyne.CanvasObject{},
		tempRect: tempRectObj,
	}
}

// 处理鼠标移动事件
func (v *ImageViewer) MouseMoved(e *desktop.MouseEvent) {
	if v.image == nil {
		return
	}

	mouseX, mouseY, ok := v.imagePositionFromView(e.Position)
	if !ok {
		return
	}

	// 检测鼠标是否真的移动了（避免滚动导致的坐标变化）
	mouseMoved := (mouseX != v.lastMouseX || mouseY != v.lastMouseY)

	// 更新上次鼠标位置
	v.lastMouseX = mouseX
	v.lastMouseY = mouseY

	if v.isDragging && v.dragMode == imageDragPan {
		delta := e.AbsolutePosition.Subtract(v.lastDragAbs)
		v.panBy(delta)
		v.lastDragAbs = e.AbsolutePosition
	}

	// 范围模式拖动时更新临时矩形
	if v.isDragging && v.dragMode == imageDragRange {
		// 更新或创建临时矩形
		if v.tempRect == nil {
			v.tempRect = &MarkRect{
				X1:    v.mouseDownX,
				Y1:    v.mouseDownY,
				X2:    mouseX,
				Y2:    mouseY,
				Color: color.RGBA{255, 0, 0, 255}, // 红色
			}
		} else {
			// 更新临时矩形的终点
			v.tempRect.X2 = mouseX
			v.tempRect.Y2 = mouseY
		}

		// 刷新视图以更新临时矩形的显示
		v.Refresh()
	}

	// 更新放大镜（不在拖动状态时）
	if !v.isDragging && magnifierEnabled && v.magnifier != nil && v.scrollContainer != nil {
		// mouseX, mouseY 是相对于ImageViewer的坐标（图像坐标）
		// 需要转换为相对于可见窗口的坐标
		scrollOffset := v.scrollContainer.Offset
		scale := v.currentZoomScale()
		bounds := v.image.Bounds()
		viewX := float32(mouseX-bounds.Min.X)*scale - scrollOffset.X
		viewY := float32(mouseY-bounds.Min.Y)*scale - scrollOffset.Y

		// 检查图像坐标是否在范围内
		imgBounds := v.image.Bounds()
		if mouseX >= imgBounds.Min.X && mouseX < imgBounds.Max.X &&
			mouseY >= imgBounds.Min.Y && mouseY < imgBounds.Max.Y {
			// 更新放大镜，传递图像坐标和可见窗口坐标
			v.magnifier.Update(v.image, mouseX, mouseY, viewX, viewY)
		} else {
			v.magnifier.Hide()
		}
	}

	// 调用回调函数（如果有）
	if mouseMoved && v.onMouseMove != nil {
		v.onMouseMove(mouseX, mouseY)
	}
}

// 实现桌面光标接口的其他方法
func (v *ImageViewer) MouseIn(*desktop.MouseEvent) {
	v.mouseInWidget = true
	if v.window != nil {
		v.window.Canvas().Focus(v)
	}
	if magnifierEnabled && v.magnifier != nil && v.image != nil {
		v.magnifier.Show()
	}
}

func (v *ImageViewer) MouseOut() {
	v.mouseInWidget = false
	if v.magnifier != nil && !v.isDragging {
		v.magnifier.Hide()
	}
}

// 实现 Focusable 接口
func (v *ImageViewer) FocusGained()     {}
func (v *ImageViewer) FocusLost()       {}
func (v *ImageViewer) TypedRune(r rune) {}

// 处理键盘按键事件
func (v *ImageViewer) TypedKey(key *fyne.KeyEvent) {
	// 只有在鼠标在图片框上时才处理方向键
	if !v.mouseInWidget || v.image == nil || v.window == nil {
		return
	}

	// 如果 lastMouseX 和 lastMouseY 还没有初始化，使用图像中心作为起始位置
	if v.lastMouseX < 0 || v.lastMouseY < 0 {
		bounds := v.image.Bounds()
		v.lastMouseX = (bounds.Min.X + bounds.Max.X) / 2
		v.lastMouseY = (bounds.Min.Y + bounds.Max.Y) / 2
	}

	// 定义移动步长（像素）
	step := 1

	// 根据按键方向计算新的图像坐标
	newX := v.lastMouseX
	newY := v.lastMouseY

	switch key.Name {
	case fyne.KeyUp:
		newY -= step
	case fyne.KeyDown:
		newY += step
	case fyne.KeyLeft:
		newX -= step
	case fyne.KeyRight:
		newX += step
	case fyne.KeySpace:
		// 按空格键取当前鼠标位置的单个点
		if v.lastMouseX >= 0 && v.lastMouseY >= 0 {
			// 确保坐标在图像范围内
			bounds := v.image.Bounds()
			mouseX := v.lastMouseX
			mouseY := v.lastMouseY
			if mouseX < bounds.Min.X {
				mouseX = bounds.Min.X
			}
			if mouseX >= bounds.Max.X {
				mouseX = bounds.Max.X - 1
			}
			if mouseY < bounds.Min.Y {
				mouseY = bounds.Min.Y
			}
			if mouseY >= bounds.Max.Y {
				mouseY = bounds.Max.Y - 1
			}
			// 添加点（使用nil表示需要计算高对比度颜色）
			v.AddPoint(mouseX, mouseY, nil)
		}
		return // 空格键处理完毕，直接返回
	case fyne.KeyReturn, fyne.KeyEnter:
		// 按回车键触发生成按钮
		if triggerGenerateCode != nil {
			triggerGenerateCode()
		}
		return // 回车键处理完毕，直接返回
	default:
		return // 不是方向键、空格键或回车键，直接返回
	}

	// 限制坐标在图像范围内
	bounds := v.image.Bounds()
	if newX < bounds.Min.X {
		newX = bounds.Min.X
	}
	if newX >= bounds.Max.X {
		newX = bounds.Max.X - 1
	}
	if newY < bounds.Min.Y {
		newY = bounds.Min.Y
	}
	if newY >= bounds.Max.Y {
		newY = bounds.Max.Y - 1
	}

	// 计算相对移动量（相对于当前鼠标位置）
	deltaX := float64(newX - v.lastMouseX)
	deltaY := float64(newY - v.lastMouseY)

	// 更新内部状态
	v.lastMouseX = newX
	v.lastMouseY = newY

	// 相对移动系统鼠标光标
	// 注意：这里的 deltaX 和 deltaY 是像素级别的移动
	if deltaX != 0 || deltaY != 0 {
		moveMouseRelative(deltaX, deltaY)
	}

	// 主动更新放大镜（因为移动鼠标可能不会触发MouseMoved事件）
	if magnifierEnabled && v.magnifier != nil && v.scrollContainer != nil {
		// 计算相对于可见窗口的坐标
		scrollOffset := v.scrollContainer.Offset
		scale := v.currentZoomScale()
		bounds := v.image.Bounds()
		viewX := float32(newX-bounds.Min.X)*scale - scrollOffset.X
		viewY := float32(newY-bounds.Min.Y)*scale - scrollOffset.Y

		// 更新放大镜
		v.magnifier.Update(v.image, newX, newY, viewX, viewY)
	}
}

func (v *ImageViewer) CursorType() desktop.Cursor {
	return desktop.CrosshairCursor
}

// 获取图像尺寸
func (v *ImageViewer) ImageSize() (int, int) {
	if v.image == nil {
		return 0, 0
	}
	bounds := v.image.Bounds()
	return bounds.Max.X - bounds.Min.X, bounds.Max.Y - bounds.Min.Y
}

// 图像查看器渲染器
type imageViewerRenderer struct {
	viewer   *ImageViewer
	objects  []fyne.CanvasObject
	points   []fyne.CanvasObject // 用于绘制点的对象
	rects    []fyne.CanvasObject // 用于绘制矩形的对象
	tempRect fyne.CanvasObject   // 用于绘制临时矩形的对象
}

func (r *imageViewerRenderer) MinSize() fyne.Size {
	if r.viewer.image == nil {
		return fyne.NewSize(0, 0)
	}
	bounds := r.viewer.image.Bounds()
	scale := r.viewer.currentZoomScale()
	return fyne.NewSize(
		float32(bounds.Max.X-bounds.Min.X)*scale,
		float32(bounds.Max.Y-bounds.Min.Y)*scale,
	)
}

func (r *imageViewerRenderer) Layout(size fyne.Size) {
	r.viewer.displayImage.Resize(size)

	// 调整点标记的位置
	r.updatePointsLayout()

	// 调整矩形标记的位置
	r.updateRectsLayout()
}

func (r *imageViewerRenderer) Refresh() {
	r.viewer.displayImage.Refresh()

	// 更新点标记和矩形标记
	r.updatePoints()
	r.updateRects()
	r.updateTempRect() // 更新临时矩形

	// 确保我们有所有的对象
	allObjects := []fyne.CanvasObject{r.viewer.displayImage, r.tempRect}
	allObjects = append(allObjects, r.points...)
	allObjects = append(allObjects, r.rects...)
	if r.viewer.contextMenu != nil {
		allObjects = append(allObjects, r.viewer.contextMenu)
	}
	r.objects = allObjects

	// 强制刷新所有点和矩形的大小和位置
	r.updatePointsLayout()
	r.updateRectsLayout()

	// 刷新所有点和矩形
	for _, p := range r.points {
		p.Refresh()
	}
	for _, rect := range r.rects {
		rect.Refresh()
	}
	r.tempRect.Refresh()
}

// 更新点标记的位置
func (r *imageViewerRenderer) updatePointsLayout() {
	// 如果没有图像，不做任何事
	if r.viewer.image == nil {
		return
	}

	// 遍历所有点对象并调整它们的位置和大小
	bounds := r.viewer.image.Bounds()
	scale := r.viewer.currentZoomScale()
	pointSize := float32(2) * scale
	if pointSize < 2 {
		pointSize = 2
	}
	for i, p := range r.points {
		if i < len(r.viewer.markPoints) {
			point := r.viewer.markPoints[i]
			p.Resize(fyne.NewSize(pointSize, pointSize))
			// 移动点到正确的位置
			p.Move(fyne.NewPos(float32(point.X-bounds.Min.X)*scale, float32(point.Y-bounds.Min.Y)*scale))
		}
	}
}

// 更新矩形标记的位置
func (r *imageViewerRenderer) updateRectsLayout() {
	// 如果没有图像，不做任何事
	if r.viewer.image == nil {
		return
	}

	// 遍历所有矩形对象并调整它们的位置和大小
	bounds := r.viewer.image.Bounds()
	scale := r.viewer.currentZoomScale()
	for i, rect := range r.rects {
		if i < len(r.viewer.markRects) {
			markRect := r.viewer.markRects[i]

			// 计算矩形的位置和尺寸
			x := float32(min(markRect.X1, markRect.X2)-bounds.Min.X) * scale
			y := float32(min(markRect.Y1, markRect.Y2)-bounds.Min.Y) * scale
			width := float32(abs(markRect.X2-markRect.X1)) * scale
			height := float32(abs(markRect.Y2-markRect.Y1)) * scale

			// 确保宽高至少为1
			if width < 1 {
				width = 1
			}
			if height < 1 {
				height = 1
			}

			// 移动和调整矩形大小
			rect.Move(fyne.NewPos(x, y))
			rect.Resize(fyne.NewSize(width, height))
		}
	}
}

// 更新点标记（添加新点，移除旧点）
func (r *imageViewerRenderer) updatePoints() {
	// 清除现有点
	r.points = make([]fyne.CanvasObject, 0, len(r.viewer.markPoints))

	// 为每个标记点创建一个2x2像素的方块
	bounds := r.viewer.image.Bounds()
	scale := r.viewer.currentZoomScale()
	pointSize := float32(2) * scale
	if pointSize < 2 {
		pointSize = 2
	}
	for _, point := range r.viewer.markPoints {
		// 创建2x2的方块，左上角对应点击位置
		rect := canvas.NewRectangle(point.Color)
		rect.SetMinSize(fyne.NewSize(pointSize, pointSize))
		rect.Resize(fyne.NewSize(pointSize, pointSize))
		rect.Move(fyne.NewPos(float32(point.X-bounds.Min.X)*scale, float32(point.Y-bounds.Min.Y)*scale))
		r.points = append(r.points, rect)
	}
}

// 更新矩形标记（添加新矩形，移除旧矩形）
func (r *imageViewerRenderer) updateRects() {
	// 清除现有矩形
	r.rects = make([]fyne.CanvasObject, 0, len(r.viewer.markRects))

	// 为每个标记矩形创建一个矩形对象
	bounds := r.viewer.image.Bounds()
	scale := r.viewer.currentZoomScale()
	for _, markRect := range r.viewer.markRects {
		rect := canvas.NewRectangle(markRect.Color)
		rect.StrokeWidth = 1
		rect.StrokeColor = markRect.Color
		rect.FillColor = color.RGBA{0, 0, 0, 0} // 透明填充

		// 计算矩形的位置和尺寸
		x := float32(min(markRect.X1, markRect.X2)-bounds.Min.X) * scale
		y := float32(min(markRect.Y1, markRect.Y2)-bounds.Min.Y) * scale
		width := float32(abs(markRect.X2-markRect.X1)) * scale
		height := float32(abs(markRect.Y2-markRect.Y1)) * scale

		// 确保宽高至少为1
		if width < 1 {
			width = 1
		}
		if height < 1 {
			height = 1
		}

		// 设置矩形的位置和尺寸
		rect.Move(fyne.NewPos(x, y))
		rect.Resize(fyne.NewSize(width, height))

		r.rects = append(r.rects, rect)
	}
}

// 更新临时矩形
func (r *imageViewerRenderer) updateTempRect() {
	// 如果没有临时矩形数据，则隐藏临时矩形对象
	if r.viewer.tempRect == nil {
		r.tempRect.Hide()
		return
	}

	// 显示临时矩形
	r.tempRect.Show()

	// 获取临时矩形数据
	tempRect := r.viewer.tempRect

	// 计算矩形的位置和尺寸
	bounds := r.viewer.image.Bounds()
	scale := r.viewer.currentZoomScale()
	x := float32(min(tempRect.X1, tempRect.X2)-bounds.Min.X) * scale
	y := float32(min(tempRect.Y1, tempRect.Y2)-bounds.Min.Y) * scale
	width := float32(abs(tempRect.X2-tempRect.X1)) * scale
	height := float32(abs(tempRect.Y2-tempRect.Y1)) * scale

	// 确保宽高至少为1
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// 移动和调整矩形大小
	rect := r.tempRect.(*canvas.Rectangle)
	rect.Move(fyne.NewPos(x, y))
	rect.Resize(fyne.NewSize(width, height))
}

// 辅助函数：计算两点之间的欧几里得距离
func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}

// 辅助函数：取两个数中的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 辅助函数：取绝对值
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func (r *imageViewerRenderer) Objects() []fyne.CanvasObject {
	// 确保所有点对象和矩形对象都包含在返回的切片中
	return r.objects
}

func (r *imageViewerRenderer) Destroy() {}

// 计算颜色的反色（使用增强的反色算法）
func getInverseColor(c color.Color) color.Color {
	// 转换为RGBA
	r, g, b, a := c.RGBA()

	// 转换到0-255范围
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	a8 := uint8(a >> 8)

	// 使用增强的反色算法
	adjust := func(value uint8) uint8 {
		normalized := float64(value) / 255.0
		if normalized < 0.5 {
			normalized = normalized * normalized // 增强低值区域
		} else {
			normalized = 1 - (1-normalized)*(1-normalized) // 增强高值区域
		}
		return uint8((1 - normalized) * 255) // 取反并返回
	}

	invR := adjust(r8)
	invG := adjust(g8)
	invB := adjust(b8)

	return color.RGBA{invR, invG, invB, a8}
}

// 检查颜色是否接近中间灰色
func isNearMidGray(r, g, b uint8) bool {
	// 计算与中间灰色(127,127,127)的差距
	const midGray = 127
	const threshold = 30 // 阈值，可以调整

	rDiff := abs(int(r) - midGray)
	gDiff := abs(int(g) - midGray)
	bDiff := abs(int(b) - midGray)

	// 如果RGB三个通道都接近中间值，且彼此接近，则认为是接近中间灰色
	avgDiff := (rDiff + gDiff + bDiff) / 3
	maxDiff := max(max(rDiff, gDiff), bDiff)

	return avgDiff < threshold && maxDiff < threshold*1.5
}

// 计算颜色亮度
func getBrightness(r, g, b uint8) uint8 {
	// 使用感知亮度公式: 0.299*R + 0.587*G + 0.114*B
	return uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

// 取两个数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 创建新的图像查看器
func NewImageViewer() *ImageViewer {
	viewer := &ImageViewer{
		displayImage:    canvas.NewImageFromImage(nil),
		markPoints:      make([]MarkPoint, 0),
		markRects:       make([]MarkRect, 0),
		tempRect:        nil, // 初始化为nil，表示没有临时矩形
		rotationDegrees: 0,   // 初始化旋转角度为0
		zoomScale:       1,
		lastMouseX:      -1, // 初始化为-1，确保第一次移动会被检测到
		lastMouseY:      -1,
		onMouseMove: func(x, y int) {
			//fmt.Printf("鼠标移动: X=%d, Y=%d\n", x, y)
		},
		onMouseDown: func(x, y int) {
			//fmt.Printf("左键按下: X=%d, Y=%d\n", x, y)
		},
		onMouseUp: func(x, y int) {
			//fmt.Printf("左键弹起: X=%d, Y=%d\n", x, y)
		},
		onRightClick: func(x, y int) {
			//fmt.Printf("右键点击: X=%d, Y=%d\n", x, y)
		},
	}
	viewer.displayImage.FillMode = canvas.ImageFillContain
	viewer.ExtendBaseWidget(viewer)
	return viewer
}

// 添加点标记
func (v *ImageViewer) AddPoint(x, y int, c color.Color) {
	// 如果提供了颜色，使用提供的颜色
	// 如果没有提供颜色（nil）且有图像，则根据背景亮度选择高对比度颜色
	markColor := c
	var originalColor color.Color // 存储原始颜色

	if c == nil && v.image != nil {
		// 确保坐标在图像范围内
		bounds := v.image.Bounds()
		if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
			// 获取图像中该点的颜色
			pixelColor := v.image.At(x, y)
			originalColor = pixelColor // 保存原始颜色

			// 使用反色作为标记颜色
			markColor = getInverseColor(pixelColor)
		} else {
			// 如果坐标超出范围，使用默认黑色的反色（白色）
			originalColor = color.RGBA{0, 0, 0, 255} // 默认黑色
			markColor = getInverseColor(originalColor)
		}
	} else {
		originalColor = color.RGBA{0, 0, 0, 255} // 默认黑色
	}

	v.markPoints = append(v.markPoints, MarkPoint{X: x, Y: y, Color: markColor})

	// 获取原始颜色的十六进制表示
	r, g, b, _ := originalColor.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	hexColor := fmt.Sprintf("#%02X%02X%02X", r8, g8, b8)

	// 将点的信息添加到颜色点列表中，使用原始颜色
	addColorPointToList(x, y, hexColor, true)

	// 更新区域包围盒
	v.updateBoundingBox()

	v.Refresh() // 刷新视图以显示新点
}

// 批量添加NxN点阵的颜色点
func (v *ImageViewer) AddGridPoints(startX, startY, cols, rows, spacing int) {
	if v.image == nil {
		return
	}

	bounds := v.image.Bounds()

	// 双层循环遍历点阵
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			// 计算当前点的坐标
			x := startX + col*spacing
			y := startY + row*spacing

			// 检查坐标是否在图像范围内
			if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
				continue // 跳过超出范围的点
			}

			// 获取该点的颜色
			pixelColor := v.image.At(x, y)
			r, g, b, _ := pixelColor.RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			// 使用反色作为标记颜色
			markColor := getInverseColor(pixelColor)

			// 添加标记点
			v.markPoints = append(v.markPoints, MarkPoint{X: x, Y: y, Color: markColor})

			// 添加到颜色点列表
			hexColor := fmt.Sprintf("#%02X%02X%02X", r8, g8, b8)
			addColorPointToList(x, y, hexColor, true)
		}
	}

	// 更新区域包围盒
	v.updateBoundingBox()

	// 批量添加完成后刷新一次
	v.Refresh()
}

// 裁剪图像的辅助函数
func cropImage(img image.Image, rect image.Rectangle) image.Image {
	bounds := img.Bounds()
	croppedRect := image.Rect(0, 0, rect.Dx(), rect.Dy())
	croppedImg := image.NewNRGBA(croppedRect) // 使用NRGBA而不是RGBA

	// 复制像素
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
				croppedImg.Set(x-rect.Min.X, y-rect.Min.Y, img.At(x, y))
			}
		}
	}

	return croppedImg
}

// 将任意图像转换为NRGBA格式
func convertToNRGBA(src image.Image) *image.NRGBA {
	// 如果已经是NRGBA，直接返回
	if nrgba, ok := src.(*image.NRGBA); ok {
		return nrgba
	}

	// 创建新的NRGBA图像
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)

	// 复制像素
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}

	return dst
}

// 放大镜组件
type MagnifierWidget struct {
	widget.BaseWidget
	sourceImage   image.Image
	gridImage     *image.NRGBA // 使用NRGBA而不是RGBA
	gridRaster    *canvas.Raster
	gridSize      int
	cellSize      int
	infoText      *canvas.Text
	background    *canvas.Rectangle
	visible       bool
	mouseX        int     // 鼠标在图像中的X坐标（用于取色）
	mouseY        int     // 鼠标在图像中的Y坐标（用于取色）
	cursorX       float32 // 鼠标在可见区域中的X坐标（用于定位放大镜）
	cursorY       float32 // 鼠标在可见区域中的Y坐标（用于定位放大镜）
	containerSize fyne.Size
}

// 创建新的放大镜组件
func NewMagnifierWidget() *MagnifierWidget {
	m := &MagnifierWidget{
		gridSize: 15,
		cellSize: 15,
		visible:  false,
	}

	// 创建背景
	m.background = canvas.NewRectangle(color.NRGBA{40, 40, 40, 230})
	m.background.StrokeWidth = 2
	m.background.StrokeColor = color.NRGBA{200, 200, 200, 255}

	// 创建用于绘制网格的图像
	gridPixelSize := m.gridSize * m.cellSize
	m.gridImage = image.NewNRGBA(image.Rect(0, 0, gridPixelSize, gridPixelSize)) // 使用NRGBA而不是RGBA

	// 创建Raster来显示网格
	m.gridRaster = canvas.NewRaster(func(w, h int) image.Image {
		return m.gridImage
	})
	m.gridRaster.ScaleMode = canvas.ImageScalePixels

	// 创建信息文本
	m.infoText = canvas.NewText("X:0 Y:0 RGB:#000000", color.White)
	m.infoText.TextSize = 12
	m.infoText.Alignment = fyne.TextAlignCenter

	// 初始化时隐藏所有元素
	m.background.Hide()
	m.gridRaster.Hide()
	m.infoText.Hide()

	m.ExtendBaseWidget(m)
	return m
}

// 快速获取像素颜色（直接访问Pix数组，图像始终是NRGBA）
func getPixelColorFast(img image.Image, x, y int) (r, g, b uint8) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return 0, 0, 0 // 超出范围返回黑色
	}

	// 图像始终是NRGBA
	imgTyped := img.(*image.NRGBA)
	idx := (y-bounds.Min.Y)*imgTyped.Stride + (x-bounds.Min.X)*4
	return imgTyped.Pix[idx], imgTyped.Pix[idx+1], imgTyped.Pix[idx+2]
}

// 更新放大镜显示 - 完全复刻参考代码的逻辑
func (m *MagnifierWidget) Update(img image.Image, imageX, imageY int, cursorX, cursorY float32) {
	if !magnifierEnabled {
		m.Hide()
		return
	}
	if img == nil {
		return
	}

	m.sourceImage = img
	m.mouseX = imageX
	m.mouseY = imageY
	m.cursorX = cursorX
	m.cursorY = cursorY
	m.visible = true

	halfGrid := m.gridSize / 2

	// 获取中心点颜色并更新信息文本
	r8, g8, b8 := getPixelColorFast(img, imageX, imageY)
	colorStr := fmt.Sprintf("#%02X%02X%02X", r8, g8, b8)
	m.infoText.Text = fmt.Sprintf("X:%d Y:%d RGB:%s", imageX, imageY, colorStr)

	// 绘制网格到图像
	borderColor := color.RGBA{80, 80, 80, 255}

	for y := 0; y < m.gridSize; y++ {
		for x := 0; x < m.gridSize; x++ {
			pixelX := imageX - halfGrid + x
			pixelY := imageY - halfGrid + y

			// 快速获取像素颜色
			pr, pg, pb := getPixelColorFast(img, pixelX, pixelY)
			pixelColor := color.RGBA{pr, pg, pb, 255}

			// 填充格子区域
			for dy := 0; dy < m.cellSize; dy++ {
				for dx := 0; dx < m.cellSize; dx++ {
					imgX := x*m.cellSize + dx
					imgY := y*m.cellSize + dy
					gridIdx := imgY*m.gridImage.Stride + imgX*4

					// 绘制格子边框
					if dx == 0 || dy == 0 {
						m.gridImage.Pix[gridIdx] = borderColor.R
						m.gridImage.Pix[gridIdx+1] = borderColor.G
						m.gridImage.Pix[gridIdx+2] = borderColor.B
						m.gridImage.Pix[gridIdx+3] = borderColor.A
					} else {
						m.gridImage.Pix[gridIdx] = pixelColor.R
						m.gridImage.Pix[gridIdx+1] = pixelColor.G
						m.gridImage.Pix[gridIdx+2] = pixelColor.B
						m.gridImage.Pix[gridIdx+3] = pixelColor.A
					}
				}
			}

			// 中心位置添加特殊标识
			if x == halfGrid && y == halfGrid {
				// 使用中心像素的反色作为边框颜色
				centerColor := color.RGBA{pr, pg, pb, 255}
				inverseColor := getInverseColor(centerColor)
				r, g, b, _ := inverseColor.RGBA()
				borderR, borderG, borderB := uint8(r>>8), uint8(g>>8), uint8(b>>8)

				// 绘制加粗的边框
				for i := 0; i < m.cellSize; i++ {
					imgX := x * m.cellSize
					imgY := y * m.cellSize

					// 顶部和底部
					for _, dy := range []int{0, 1, m.cellSize - 2, m.cellSize - 1} {
						idx := (imgY+dy)*m.gridImage.Stride + (imgX+i)*4
						m.gridImage.Pix[idx] = borderR
						m.gridImage.Pix[idx+1] = borderG
						m.gridImage.Pix[idx+2] = borderB
						m.gridImage.Pix[idx+3] = 255
					}

					// 左侧和右侧
					for _, dx := range []int{0, 1, m.cellSize - 2, m.cellSize - 1} {
						idx := (imgY+i)*m.gridImage.Stride + (imgX+dx)*4
						m.gridImage.Pix[idx] = borderR
						m.gridImage.Pix[idx+1] = borderG
						m.gridImage.Pix[idx+2] = borderB
						m.gridImage.Pix[idx+3] = 255
					}
				}
			}
		}
	}

	m.Refresh()
}

// 隐藏放大镜
func (m *MagnifierWidget) Hide() {
	m.visible = false
	m.Refresh()
}

// 显示放大镜
func (m *MagnifierWidget) Show() {
	m.visible = true
	m.Refresh()
}

// 创建渲染器
func (m *MagnifierWidget) CreateRenderer() fyne.WidgetRenderer {
	return &magnifierRenderer{
		magnifier: m,
	}
}

// 放大镜渲染器
type magnifierRenderer struct {
	magnifier *MagnifierWidget
}

func (r *magnifierRenderer) MinSize() fyne.Size {
	// 不占用空间，因为是浮动在上层的
	return fyne.NewSize(0, 0)
}

func (r *magnifierRenderer) Layout(size fyne.Size) {
	// 保存容器尺寸
	r.magnifier.containerSize = size

	if !r.magnifier.visible {
		return
	}

	// 计算放大镜的尺寸
	gridWidth := float32(r.magnifier.gridSize * r.magnifier.cellSize)
	gridHeight := float32(r.magnifier.gridSize * r.magnifier.cellSize)
	infoHeight := float32(20)
	totalWidth := gridWidth + 10
	totalHeight := gridHeight + infoHeight + 15

	// 完全复刻参考代码的逻辑：放大镜位置 = 鼠标位置 + 偏移(20, 20)
	offsetX := float32(20)
	offsetY := float32(20)

	posX := r.magnifier.cursorX + offsetX
	posY := r.magnifier.cursorY + offsetY

	// 边界检查：如果会超出右边界，放到鼠标左边
	if posX+totalWidth > size.Width {
		posX = r.magnifier.cursorX - totalWidth - offsetX
	}

	// 边界检查：如果会超出下边界，放到鼠标上边
	if posY+totalHeight > size.Height {
		posY = r.magnifier.cursorY - totalHeight - offsetY
	}

	// 最终边界保护
	if posX < 5 {
		posX = 5
	}
	if posY < 5 {
		posY = 5
	}
	if posX+totalWidth > size.Width-5 {
		posX = size.Width - totalWidth - 5
	}
	if posY+totalHeight > size.Height-5 {
		posY = size.Height - totalHeight - 5
	}

	// 布局背景
	r.magnifier.background.Move(fyne.NewPos(posX, posY))
	r.magnifier.background.Resize(fyne.NewSize(totalWidth, totalHeight))

	// 布局网格（现在在顶部）
	gridX := posX + 5
	gridY := posY + 5
	r.magnifier.gridRaster.Move(fyne.NewPos(gridX, gridY))
	r.magnifier.gridRaster.Resize(fyne.NewSize(gridWidth, gridHeight))

	// 布局信息文本（现在在底部）
	infoY := posY + gridHeight + 10
	r.magnifier.infoText.Move(fyne.NewPos(posX+5, infoY))
	r.magnifier.infoText.Resize(fyne.NewSize(totalWidth-10, infoHeight))
}

func (r *magnifierRenderer) Refresh() {
	if !r.magnifier.visible {
		r.magnifier.background.Hide()
		r.magnifier.infoText.Hide()
		r.magnifier.gridRaster.Hide()
	} else {
		// 每次刷新时重新计算位置（跟随鼠标）
		if r.magnifier.containerSize.Width > 0 && r.magnifier.containerSize.Height > 0 {
			r.Layout(r.magnifier.containerSize)
		}

		r.magnifier.background.Show()
		r.magnifier.infoText.Show()
		r.magnifier.gridRaster.Show()
		r.magnifier.gridRaster.Refresh()
		r.magnifier.infoText.Refresh()
		r.magnifier.background.Refresh()
	}
}

func (r *magnifierRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.magnifier.background, r.magnifier.infoText, r.magnifier.gridRaster}
}

func (r *magnifierRenderer) Destroy() {}

func main() {
	// 释放嵌入的 cap.dex 到临时目录
	extractCapDex()

	// 创建应用 - 默认使用系统主题
	a := app.New()

	// 检测系统当前主题并更新isDarkTheme变量
	currentTheme := a.Settings().ThemeVariant()
	isDarkTheme = currentTheme == theme.VariantDark

	// 设置自定义主题
	a.Settings().SetTheme(newMyTheme())

	// 创建设备选择下拉框
	deviceSelect = widget.NewSelect([]string{}, func(value string) {
		if value != "" {
			selectedDevice = value
		}
	})
	// 设置下拉框占位符
	deviceSelect.PlaceHolder = "正在加载设备..."

	// 启动设备监控线程
	go deviceMonitor()

	// 延迟调用，先确保UI已经初始化完成
	go func() {
		time.Sleep(500 * time.Millisecond)
		deviceRefreshChan <- true
	}()

	// 创建窗口
	w := a.NewWindow("AutoGo图色助手")
	mainWindowSize := initialWindowSize(0.70, 0.70)
	userConfig := loadUserConfig()
	apiFormatTemplates = copyAPIFormatTemplates(userConfig.FormatTemplates)
	magnifierEnabled = userConfig.ShowMagnifier
	autoCopyRangeEnabled = userConfig.AutoCopyRange
	gridModeEnabled = userConfig.GridMode
	gridColsValue = userConfig.GridCols
	gridRowsValue = userConfig.GridRows
	gridSpacingValue = userConfig.GridSpacing

	// 创建标签页容器（使用修改后的DocTabs，无滚动条但支持关闭功能）
	tabs := container.NewDocTabs()
	tabs.SetTabLocation(container.TabLocationTop)
	var updateRangeButton func()

	// 设置标签页切换监听器
	tabs.OnSelected = func(tab *container.TabItem) {
		// 保存之前标签页的数据
		saveCurrentTabData()

		// 更新当前标签页
		currentTab = tab

		// 恢复新标签页的数据
		restoreTabData(tab)
		if updateRangeButton != nil {
			updateRangeButton()
		}
	}

	// 设置标签页关闭监听器
	tabs.OnClosed = func(tab *container.TabItem) {
		// 清理关闭标签页的数据
		delete(tabDataMap, tab)

		// 如果关闭的是当前标签页，清空当前引用
		if currentTab == tab {
			currentTab = nil
			imageViewer = nil
			colorPoints = make([]ColorPoint, 0)
			if rectCoordEntry != nil {
				rectCoordEntry.SetText(defaultRangeText)
			}
			if codeDisplayEntry != nil {
				codeDisplayEntry.SetText("")
			}
			if refreshColorList != nil {
				refreshColorList()
			}
			if updateRangeButton != nil {
				updateRangeButton()
			}
		}
	}

	// 标签页计数器
	tabCounter := 0

	// 创建第一个标签页（欢迎页）- 添加详细的使用说明
	welcomeTitle := widget.NewLabelWithStyle("欢迎使用 AutoGo 图色助手", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	instructionsText := `
📱 开始使用
  • 点击「截图」按钮从设备截取屏幕
  • 点击「载入」按钮导入本地图片
  • 每次截图/载入会在新标签页打开

⌨️ 快捷键
  • 左键点击/拖动：拖动图像
  • Ctrl + 左键点击：在图像上标记取色点
  • 范围按钮 / Ctrl+R：进入一次范围框选
  • Ctrl + 滚轮：缩放图像
  • 右键点击：打开坐标/颜色/点位菜单
  • ↑ ↓ ← →  移动鼠标（1像素）
  • Space     标记当前位置
  • Enter     生成代码

🎨 功能说明
  • 找色模式：多点找色（颜色匹配）
  • 比色模式：多点比色（精确比对）
  • 点阵模式：快速生成网格取色点
  • 偏色设置：容差值（如：101010）
  • 裁剪功能：框选后点击「裁剪」

💡 提示
  • 生成代码会自动复制到剪切板
  • 没有主动框选区域的时候会根据标记点自动生成区域
  • 放大镜会实时显示鼠标位置的像素放大图
  • 右侧表格显示所有取色点信息
  • 点击「生成」按钮获取代码
  • 标签页右侧有关闭按钮（❌）
`

	instructionsLabel := widget.NewLabel(instructionsText)
	instructionsLabel.Wrapping = fyne.TextWrapWord

	welcomeContent := container.NewVBox(
		layout.NewSpacer(),
		welcomeTitle,
		widget.NewSeparator(),
		instructionsLabel,
		layout.NewSpacer(),
	)

	welcomeScroll := container.NewScroll(welcomeContent)
	firstTab := container.NewTabItem("🏠", welcomeScroll)
	tabs.Append(firstTab)

	// 初始化当前标签页为欢迎页
	currentTab = firstTab

	// 初始化 imageViewer 为 nil，将在第一次截图或载入时创建
	imageViewer = nil

	// 初始化空的颜色点列表，不再使用generateSampleData
	colorPoints = make([]ColorPoint, 0)

	// 设置刷新表格的函数
	refreshColorList = func() {
		// 使用fyne.Do确保在主线程中执行UI更新
		fyne.Do(func() {
			if tableContent != nil {
				tableContent.Refresh()
			}
			refreshImagesAPIFields()
		})
	}

	// 主题切换功能
	toggleTheme := func() {
		if isDarkTheme {
			// 切换到亮色主题
			isDarkTheme = false
			a.Settings().SetTheme(newMyTheme())
		} else {
			// 切换到深色主题
			isDarkTheme = true
			a.Settings().SetTheme(newMyTheme())
		}

		// 切换主题后更新表头和列表
		updateTableHeader()
		updateTableSelection()
	}

	// 创建区域坐标显示控件
	rectCoordEntry = widget.NewEntry()
	rectCoordEntry.OnChanged = func(string) {
		refreshImagesAPIFields()
	}

	// 创建偏色值输入控件
	colorOffsetEntry = widget.NewEntry()
	colorOffsetEntry.SetPlaceHolder("偏色:101010") // 默认占位符为101010（十六进制格式）
	colorOffsetEntry.MultiLine = false           // 单行显示
	colorOffsetEntry.Wrapping = fyne.TextTruncate

	// 创建找色模式选择
	colorModeRadio = widget.NewRadioGroup([]string{"找色", "比色"}, nil)
	colorModeRadio.SetSelected("找色") // 默认选择多点找色
	colorModeRadio.Horizontal = true // 水平排列

	// 点阵参数获取回调函数（每次创建新的imageViewer时都需要设置）
	getGridParamsFunc := func() (cols, rows, spacing int, hasParams bool) {
		if gridModeEnabled {
			return gridColsValue, gridRowsValue, gridSpacingValue, true
		}
		return 0, 0, 0, false
	}
	configureImageViewer := func(v *ImageViewer) {
		v.window = w
		v.getGridParams = getGridParamsFunc
		v.onRangeModeChanged = func(bool) {
			if updateRangeButton != nil {
				updateRangeButton()
			}
		}
	}
	fitImageToView := func(v *ImageViewer) {
		v.FitToView()
		fyne.Do(func() {
			v.FitToView()
		})
	}

	// 创建左侧工具栏按钮 - 使用带动画的截图按钮
	var screenshotBtn *AnimatedScreenshotButton
	screenshotBtn = NewAnimatedScreenshotButton("截图", theme.ContentCopyIcon(), func() {
		// 启动加载动画
		screenshotBtn.StartLoading()

		// 异步执行截图操作
		go func() {
			capturedImg, err := captureScreenWithADB(selectedDevice)

			// 使用 fyne.Do 确保在主线程中更新UI
			fyne.Do(func() {
				// 停止动画
				screenshotBtn.StopLoading()

				if err != nil {
					log.Printf("截图失败: %v", err)
					dialog.ShowError(fmt.Errorf("截图失败: %v", err), w)
					return
				}

				// 转换为NRGBA格式
				capturedImg = convertToNRGBA(capturedImg)

				// 保存当前标签页的数据
				saveCurrentTabData()

				// 创建新的图像查看器
				newImageViewer := NewImageViewer()

				// 创建新的放大镜
				newMagnifier := NewMagnifierWidget()

				// 创建图像容器
				newImgContainer := container.New(&topLeftLayout{}, newImageViewer)
				newScrollContainer := container.NewScroll(newImgContainer)

				// 设置引用
				newImageViewer.scrollContainer = newScrollContainer
				newImageViewer.magnifier = newMagnifier
				configureImageViewer(newImageViewer)

				// 设置图像
				newImageViewer.SetImage(capturedImg)

				// 创建新标签页
				newScrollWithMagnifier := container.NewStack(newScrollContainer, newMagnifier)
				tabCounter++
				// 使用时间作为标签名称
				tabName := time.Now().Format("15:04:05")
				newTab := container.NewTabItem(tabName, newScrollWithMagnifier)

				// 初始化新标签页的数据
				tabDataMap[newTab] = &TabData{
					colorPoints:        make([]ColorPoint, 0),
					markRects:          make([]MarkRect, 0),
					manualRectSelected: false,
					imageViewer:        newImageViewer,
					generatedCode:      "",
				}

				tabs.Append(newTab)

				// 切换到新标签页
				tabs.Select(newTab)

				// 更新当前标签页引用
				currentTab = newTab

				// 更新当前的imageViewer引用为新标签页的viewer
				imageViewer = newImageViewer
				fitImageToView(newImageViewer)

				// 清空颜色点列表和矩形区域
				colorPoints = make([]ColorPoint, 0)
				if rectCoordEntry != nil {
					rectCoordEntry.SetText(defaultRangeText)
				}
				if codeDisplayEntry != nil {
					codeDisplayEntry.SetText("")
				}

				// 刷新表格
				if refreshColorList != nil {
					refreshColorList()
				}
			})
		}()
	})

	importBtn := widget.NewButtonWithIcon("载入", theme.FolderOpenIcon(), func() {
		// 使用系统原生文件打开对话框
		go func() {
			filePath, err := nativedialog.File().
				Filter("图片文件", "png", "jpg", "jpeg", "bmp").
				Title("选择图片文件").
				Load()

			if err != nil {
				// 用户取消或发生错误
				return
			}

			// 读取文件内容
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("读取文件失败: %v", err), w)
				})
				return
			}

			// 根据文件扩展名解码图像
			var img image.Image
			ext := strings.ToLower(filepath.Ext(filePath))

			switch ext {
			case ".png":
				img, err = png.Decode(bytes.NewReader(data))
			case ".jpg", ".jpeg":
				img, err = jpeg.Decode(bytes.NewReader(data))
			case ".bmp":
				img, err = bmp.Decode(bytes.NewReader(data))
			default:
				// 尝试自动检测格式
				img, _, err = image.Decode(bytes.NewReader(data))
			}

			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("解码图像失败: %v", err), w)
				})
				return
			}

			// 转换为NRGBA格式
			img = convertToNRGBA(img)

			// 在主线程中更新UI
			fyne.Do(func() {
				// 保存当前标签页的数据
				saveCurrentTabData()

				// 创建新的图像查看器和标签页
				newImageViewer := NewImageViewer()
				newMagnifier := NewMagnifierWidget()

				newImgContainer := container.New(&topLeftLayout{}, newImageViewer)
				newScrollContainer := container.NewScroll(newImgContainer)

				newImageViewer.scrollContainer = newScrollContainer
				newImageViewer.magnifier = newMagnifier
				configureImageViewer(newImageViewer)
				newImageViewer.SetImage(img)

				// 创建新标签页
				newScrollWithMagnifier := container.NewStack(newScrollContainer, newMagnifier)
				tabCounter++

				// 使用文件名作为标签名称
				tabName := filepath.Base(filePath)
				newTab := container.NewTabItem(tabName, newScrollWithMagnifier)

				// 初始化新标签页的数据
				tabDataMap[newTab] = &TabData{
					colorPoints:        make([]ColorPoint, 0),
					markRects:          make([]MarkRect, 0),
					manualRectSelected: false,
					imageViewer:        newImageViewer,
					generatedCode:      "",
				}

				tabs.Append(newTab)
				tabs.Select(newTab)

				// 更新当前标签页引用
				currentTab = newTab

				// 更新当前imageViewer引用
				imageViewer = newImageViewer
				fitImageToView(newImageViewer)

				// 清空颜色点列表和矩形区域
				colorPoints = make([]ColorPoint, 0)
				if rectCoordEntry != nil {
					rectCoordEntry.SetText(defaultRangeText)
				}
				if codeDisplayEntry != nil {
					codeDisplayEntry.SetText("")
				}

				// 刷新表格
				if refreshColorList != nil {
					refreshColorList()
				}
			})
		}()
	})
	importBtn.Importance = widget.MediumImportance

	saveBtn := widget.NewButtonWithIcon("保存", theme.DocumentSaveIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			// 没有图像可保存
			dialog.ShowInformation("提示", "当前没有可保存的图像", w)
			return
		}

		// 保存当前图像的引用，避免在 goroutine 中被修改
		imgToSave := imageViewer.image

		// 使用系统原生文件保存对话框
		go func() {
			filePath, err := nativedialog.File().
				Filter("PNG 图片", "png").
				Filter("JPEG 图片", "jpg", "jpeg").
				Title("保存图片").
				SetStartFile("screenshot.png").
				Save()

			if err != nil {
				// 用户取消或发生错误
				return
			}

			// 获取文件扩展名
			ext := strings.ToLower(filepath.Ext(filePath))

			// 如果没有扩展名，默认添加.png
			if ext == "" {
				filePath = filePath + ".png"
				ext = ".png"
			}

			// 创建文件
			file, err := os.Create(filePath)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("创建文件失败: %v", err), w)
				})
				return
			}
			defer file.Close()

			// 根据扩展名编码图像
			if ext == ".jpg" || ext == ".jpeg" {
				err = jpeg.Encode(file, imgToSave, &jpeg.Options{Quality: 100})
			} else {
				err = png.Encode(file, imgToSave)
			}

			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("保存图像失败: %v", err), w)
				})
				return
			}
		}()
	})
	saveBtn.Importance = widget.MediumImportance

	rotateBtn := widget.NewButtonWithIcon("旋转", theme.MediaReplayIcon(), func() {
		if imageViewer == nil || imageViewer.originalImage == nil {
			return // 如果没有图像，不执行任何操作
		}

		// 计算新的旋转角度 (每次点击增加90度)
		newDegrees := (imageViewer.rotationDegrees + 90) % 360

		// 执行旋转
		imageViewer.RotateImage(newDegrees)
	})
	rotateBtn.Importance = widget.MediumImportance

	// 底部额外按钮
	cutBtn := widget.NewButtonWithIcon("裁剪", theme.ContentCutIcon(), func() {
		// 检查是否有图像
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "当前没有可裁剪的图像", w)
			return
		}

		// 检查是否有选择矩形区域
		if len(imageViewer.markRects) == 0 {
			dialog.ShowInformation("提示", "请先使用鼠标在图像上拖动选择要裁剪的区域", w)
			return
		}

		// 获取选择区域的坐标
		rect := imageViewer.markRects[0]

		// 确保坐标是左上角到右下角的顺序
		minX := min(rect.X1, rect.X2)
		minY := min(rect.Y1, rect.Y2)
		maxX := max(rect.X1, rect.X2)
		maxY := max(rect.Y1, rect.Y2)

		// 确保裁剪区域在图像范围内
		bounds := imageViewer.image.Bounds()
		if minX < bounds.Min.X {
			minX = bounds.Min.X
		}
		if minY < bounds.Min.Y {
			minY = bounds.Min.Y
		}
		if maxX > bounds.Max.X {
			maxX = bounds.Max.X
		}
		if maxY > bounds.Max.Y {
			maxY = bounds.Max.Y
		}

		// 检查裁剪区域是否有效
		if minX >= maxX || minY >= maxY {
			dialog.ShowInformation("提示", "选择的裁剪区域无效", w)
			return
		}

		// 创建裁剪区域
		cropRect := image.Rect(minX, minY, maxX, maxY)

		// 裁剪图像
		croppedImg := cropImage(imageViewer.image, cropRect)

		// 保存当前标签页的数据
		saveCurrentTabData()

		// 创建新的图像查看器和标签页来展示裁剪后的图片
		newImageViewer := NewImageViewer()
		newMagnifier := NewMagnifierWidget()

		// 创建图像容器
		newImgContainer := container.New(&topLeftLayout{}, newImageViewer)
		newScrollContainer := container.NewScroll(newImgContainer)

		// 设置引用
		newImageViewer.scrollContainer = newScrollContainer
		newImageViewer.magnifier = newMagnifier
		configureImageViewer(newImageViewer)

		// 设置裁剪后的图像
		newImageViewer.SetImage(croppedImg)

		// 创建新标签页
		newScrollWithMagnifier := container.NewStack(newScrollContainer, newMagnifier)
		tabCounter++
		// 使用"裁剪_时间"作为标签名称
		tabName := "裁剪_" + time.Now().Format("15:04:05")
		newTab := container.NewTabItem(tabName, newScrollWithMagnifier)

		// 初始化新标签页的数据
		tabDataMap[newTab] = &TabData{
			colorPoints:        make([]ColorPoint, 0),
			markRects:          make([]MarkRect, 0),
			manualRectSelected: false,
			imageViewer:        newImageViewer,
			generatedCode:      "",
		}

		tabs.Append(newTab)

		// 切换到新标签页
		tabs.Select(newTab)

		// 更新当前标签页引用
		currentTab = newTab

		// 更新当前的imageViewer引用为新标签页的viewer
		imageViewer = newImageViewer
		fitImageToView(newImageViewer)

		// 清空颜色点列表和矩形区域
		colorPoints = make([]ColorPoint, 0)
		if rectCoordEntry != nil {
			rectCoordEntry.SetText(defaultRangeText)
		}
		if codeDisplayEntry != nil {
			codeDisplayEntry.SetText("")
		}

		// 刷新表格
		if refreshColorList != nil {
			refreshColorList()
		}
	})
	cutBtn.Importance = widget.MediumImportance

	// 字库制作按钮
	fontLibBtn := widget.NewButtonWithIcon("字库制作", theme.GridIcon(), func() {
		openFontLibWindow(w)
	})
	fontLibBtn.Importance = widget.MediumImportance

	// 主题切换按钮使用高重要性，使其更加突出
	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), toggleTheme)
	themeBtn.Importance = widget.MediumImportance

	// 创建代码显示框（多行只读文本框）
	codeDisplayEntry = widget.NewMultiLineEntry()
	codeDisplayEntry.SetPlaceHolder("生成的代码将显示在这里...")
	codeDisplayEntry.Wrapping = fyne.TextWrapWord
	codeDisplayEntry.TextStyle = fyne.TextStyle{Monospace: true}

	// 创建生成代码的函数
	generateCodeFunc := func() {
		// 生成代码并复制到剪贴板
		code := ""
		if updateImagesAPIFields != nil {
			code = updateImagesAPIFields()
		} else {
			code = generateColorCode()
		}
		if code != "" {
			// 复制到剪贴板
			w.Clipboard().SetContent(code + "\n")

			// 显示在编辑框中
			codeDisplayEntry.SetText(code)
		} else {
			codeDisplayEntry.SetText("")
			dialog.ShowError(fmt.Errorf("生成失败"), w)
		}
	}

	genBtn := widget.NewButtonWithIcon("生成", theme.DocumentCreateIcon(), generateCodeFunc)
	genBtn.Importance = widget.HighImportance

	// 设置全局触发生成代码函数，供键盘快捷键使用
	triggerGenerateCode = generateCodeFunc

	// 点阵模式主按钮
	var gridModeBtn *widget.Button
	var saveCurrentConfig func()
	updateGridBtn := func() {
		if gridModeEnabled {
			gridModeBtn.SetText("● 点阵模式")
			gridModeBtn.Importance = widget.HighImportance
		} else {
			gridModeBtn.SetText("○ 点阵模式")
			gridModeBtn.Importance = widget.MediumImportance
		}
		gridModeBtn.Refresh()
	}

	gridModeBtn = widget.NewButton("○ 点阵模式", func() {
		gridModeEnabled = !gridModeEnabled
		updateGridBtn()
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	updateGridBtn() // 初始化按钮状态

	// 点阵设置按钮（小按钮）
	gridSettingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		// 创建列数输入框和按钮
		colsEntry := widget.NewEntry()
		colsEntry.SetText(fmt.Sprintf("%d", gridColsValue))
		colsMinusBtn := widget.NewButton("-", func() {
			if val, err := strconv.Atoi(colsEntry.Text); err == nil && val > 1 {
				colsEntry.SetText(fmt.Sprintf("%d", val-1))
			}
		})

		colsPlusBtn := widget.NewButton("+", func() {
			if val, err := strconv.Atoi(colsEntry.Text); err == nil {
				colsEntry.SetText(fmt.Sprintf("%d", val+1))
			}
		})

		colsBtnContainer := container.NewGridWithColumns(2, colsMinusBtn, colsPlusBtn)
		colsRow := container.NewBorder(nil, nil, nil, colsBtnContainer, colsEntry)

		// 创建行数输入框和按钮
		rowsEntry := widget.NewEntry()
		rowsEntry.SetText(fmt.Sprintf("%d", gridRowsValue))
		rowsMinusBtn := widget.NewButton("-", func() {
			if val, err := strconv.Atoi(rowsEntry.Text); err == nil && val > 1 {
				rowsEntry.SetText(fmt.Sprintf("%d", val-1))
			}
		})

		rowsPlusBtn := widget.NewButton("+", func() {
			if val, err := strconv.Atoi(rowsEntry.Text); err == nil {
				rowsEntry.SetText(fmt.Sprintf("%d", val+1))
			}
		})

		rowsBtnContainer := container.NewGridWithColumns(2, rowsMinusBtn, rowsPlusBtn)
		rowsRow := container.NewBorder(nil, nil, nil, rowsBtnContainer, rowsEntry)

		// 创建间距输入框和按钮
		spacingEntry := widget.NewEntry()
		spacingEntry.SetText(fmt.Sprintf("%d", gridSpacingValue))
		spacingMinusBtn := widget.NewButton("-", func() {
			if val, err := strconv.Atoi(spacingEntry.Text); err == nil && val > 1 {
				spacingEntry.SetText(fmt.Sprintf("%d", val-1))
			}
		})

		spacingPlusBtn := widget.NewButton("+", func() {
			if val, err := strconv.Atoi(spacingEntry.Text); err == nil {
				spacingEntry.SetText(fmt.Sprintf("%d", val+1))
			}
		})

		spacingBtnContainer := container.NewGridWithColumns(2, spacingMinusBtn, spacingPlusBtn)
		spacingRow := container.NewBorder(nil, nil, nil, spacingBtnContainer, spacingEntry)

		// 创建表单
		formItems := []*widget.FormItem{
			widget.NewFormItem("列数:", colsRow),
			widget.NewFormItem("行数:", rowsRow),
			widget.NewFormItem("间距:", spacingRow),
		}

		// 创建对话框
		d := dialog.NewForm("点阵参数设置", "确定", "取消", formItems, func(confirmed bool) {
			if confirmed {
				// 解析并保存参数
				if cols, err := strconv.Atoi(strings.TrimSpace(colsEntry.Text)); err == nil && cols > 0 {
					gridColsValue = cols
				}
				if rows, err := strconv.Atoi(strings.TrimSpace(rowsEntry.Text)); err == nil && rows > 0 {
					gridRowsValue = rows
				}
				if spacing, err := strconv.Atoi(strings.TrimSpace(spacingEntry.Text)); err == nil && spacing > 0 {
					gridSpacingValue = spacing
				}

				log.Printf("点阵参数已更新: %dx%d/%d", gridColsValue, gridRowsValue, gridSpacingValue)
				if saveCurrentConfig != nil {
					saveCurrentConfig()
				}
			}
		}, w)

		d.Resize(fyne.NewSize(300, 240))
		d.Show()
	})

	// 使用Border布局，然后用固定高度容器包装（高度设为35）
	gridRowBorder := container.NewBorder(nil, nil, nil, gridSettingsBtn, gridModeBtn)
	gridRow := newFixedHeightContainer(gridRowBorder, 35)

	// 如果imageViewer已经存在（不为nil），设置回调
	if imageViewer != nil {
		configureImageViewer(imageViewer)
	}

	// 左侧工具栏布局：模拟 AutoGo 工具面板的窄栏按钮布局
	makeButton := func(text string) *widget.Button {
		btn := widget.NewButton(text, func() {})
		btn.Importance = widget.MediumImportance
		return btn
	}
	makeEntry := func(text string) *widget.Entry {
		entry := widget.NewEntry()
		entry.SetText(text)
		return entry
	}
	makeFixedPanel := func(width float32, content fyne.CanvasObject) *fyne.Container {
		minWidth := canvas.NewRectangle(color.Transparent)
		minWidth.SetMinSize(fyne.NewSize(width, 1))
		return container.NewStack(minWidth, content)
	}
	makeFixedWidthPanel := func(width float32, content fyne.CanvasObject) *fyne.Container {
		return container.New(&fixedWidthLayout{width: width}, content)
	}

	showMagnifierCheck := widget.NewCheck("显示放大镜", func(checked bool) {
		magnifierEnabled = checked
		if !checked && imageViewer != nil && imageViewer.magnifier != nil {
			imageViewer.magnifier.Hide()
		}
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	showMagnifierCheck.SetChecked(userConfig.ShowMagnifier)
	magnifierThemeRow := container.NewBorder(nil, nil, nil, themeBtn, showMagnifierCheck)

	screenshotBtn.button.SetText("截图 (CTRL+Z)")
	importBtn.SetText("加载 (CTRL+L)")

	rotateLeftBtn := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		if imageViewer == nil || imageViewer.originalImage == nil {
			return
		}
		imageViewer.RotateImage((imageViewer.rotationDegrees + 270) % 360)
	})
	rotateRightBtn := widget.NewButtonWithIcon("", theme.ContentRedoIcon(), func() {
		if imageViewer == nil || imageViewer.originalImage == nil {
			return
		}
		imageViewer.RotateImage((imageViewer.rotationDegrees + 90) % 360)
	})
	rotateRow := container.NewGridWithColumns(2, rotateLeftBtn, rotateRightBtn)

	rangeBtn := widget.NewButton("范围 (CTRL+R)", func() {
		if imageViewer == nil || imageViewer.image == nil {
			return
		}
		imageViewer.SetRangeSelectMode(!imageViewer.rangeSelectMode)
	})
	updateRangeButton = func() {
		if imageViewer != nil && imageViewer.rangeSelectMode {
			rangeBtn.SetText("选择范围中...")
			rangeBtn.Importance = widget.HighImportance
		} else {
			rangeBtn.SetText("范围 (CTRL+R)")
			rangeBtn.Importance = widget.MediumImportance
		}
		rangeBtn.Refresh()
	}
	updateRangeButton()
	w.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}, func(fyne.Shortcut) {
		if imageViewer == nil || imageViewer.image == nil {
			return
		}
		imageViewer.SetRangeSelectMode(!imageViewer.rangeSelectMode)
	})

	rectCoordEntry.SetText(defaultRangeText)
	coordDisplayEntry := rectCoordEntry
	copyResetRow := container.NewGridWithColumns(2, widget.NewButton("复制", func() {
		w.Clipboard().SetContent(coordDisplayEntry.Text)
	}), widget.NewButton("重置", func() {
		if imageViewer != nil {
			imageViewer.ResetRangeSelection()
			return
		}
		setRectCoordText(defaultRangeText)
	}))
	resetZoomBtn := widget.NewButton("重置缩放", func() {
		if imageViewer != nil {
			imageViewer.FitToView()
		}
	})
	originalSizeBtn := widget.NewButton("显示原始尺寸", func() {
		if imageViewer != nil {
			imageViewer.ShowOriginalSize()
		}
	})

	pickModeSelect := widget.NewSelect([]string{
		"随机取点",
		"轮廓取点",
		"重要取点",
		"最高和点",
		"最高和取点",
		"颜色分类轮廓",
		"颜色分类随机",
	}, func(string) {
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	pickModeSelect.SetSelected(userConfig.PickMode)
	pickCountEntry := makeEntry(userConfig.PickCount)
	pickCountEntry.OnChanged = func(string) {
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	}
	applyRangeCheck := widget.NewCheck("选取后应用范围", func(bool) {
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	applyRangeCheck.SetChecked(userConfig.ApplyRange)
	autoCopyRangeCheck := widget.NewCheck("自动复制范围", func(checked bool) {
		autoCopyRangeEnabled = checked
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	autoCopyRangeCheck.SetChecked(autoCopyRangeEnabled)

	leftControls := container.NewVBox(
		deviceSelect,
		magnifierThemeRow,
		screenshotBtn,
		rotateRow,
		importBtn,
		rangeBtn,
		coordDisplayEntry,
		copyResetRow,
		resetZoomBtn,
		originalSizeBtn,
		makeButton("抓取节点"),
		makeButton("清除查找标记"),
		makeButton("自动取色 (CTRL+A)"),
		pickModeSelect,
		container.NewBorder(nil, nil, widget.NewLabel("取色个数"), nil, pickCountEntry),
		applyRangeCheck,
		autoCopyRangeCheck,
		gridRow,
		fontLibBtn,
	)
	leftPanel := makeFixedWidthPanel(190, container.New(&compactPaddedLayout{padding: 2}, container.NewVScroll(container.New(&fixedContentWidthLayout{width: 170}, leftControls))))

	// 右侧工具栏布局：模拟图色工具 / 节点工具面板
	headerBg = canvas.NewRectangle(getHeaderBgColor(isDarkTheme))
	headerBg.SetMinSize(fyne.NewSize(360, 28))

	idHeader = canvas.NewText("", getTextColor(isDarkTheme))
	checkHeader := canvas.NewText("勾选", getTextColor(isDarkTheme))
	posHeader = canvas.NewText("坐标", getTextColor(isDarkTheme))
	colorHeader = canvas.NewText("取色", getTextColor(isDarkTheme))
	statusHeader = canvas.NewText("RGB", getTextColor(isDarkTheme))
	offsetHeader := canvas.NewText("偏色", getTextColor(isDarkTheme))
	for _, header := range []*canvas.Text{idHeader, checkHeader, posHeader, colorHeader, statusHeader, offsetHeader} {
		header.Alignment = fyne.TextAlignCenter
		header.TextSize = 12
		header.TextStyle = fyne.TextStyle{Bold: true}
	}

	tableHeader = container.NewStack(headerBg, container.New(&weightedGridLayout{
		cols:    6,
		weights: []float32{0.55, 1.0, 1.7, 1.6, 2.0, 1.45},
	},
		container.NewCenter(idHeader),
		container.NewCenter(checkHeader),
		container.NewCenter(posHeader),
		container.NewCenter(colorHeader),
		container.NewCenter(statusHeader),
		container.NewCenter(offsetHeader),
	))

	updateTableHeader = func() {
		headerBg.FillColor = getHeaderBgColor(isDarkTheme)
		headerBg.Refresh()
		textColor := getTextColor(isDarkTheme)
		for _, header := range []*canvas.Text{idHeader, checkHeader, posHeader, colorHeader, statusHeader, offsetHeader} {
			header.Color = textColor
			header.Refresh()
		}
	}

	updateTableSelection = func() {
		fyne.Do(func() {
			if tableContent != nil {
				tableContent.Refresh()
			}
			refreshImagesAPIFields()
		})
	}

	tableContent = widget.NewList(
		func() int {
			if len(colorPoints) > 11 {
				return len(colorPoints)
			}
			return 11
		},
		func() fyne.CanvasObject {
			idxText := canvas.NewText("", getTextColor(isDarkTheme))
			idxText.Alignment = fyne.TextAlignCenter
			idxText.TextSize = 11

			statusCheck := NewColorCheck(false, color.RGBA{240, 240, 240, 255}, func(bool) {})
			coordEntry := newCommitEntry()

			swatch := canvas.NewRectangle(color.Black)
			swatch.SetMinSize(fyne.NewSize(56, 24))

			rgbText := canvas.NewText("", getTextColor(isDarkTheme))
			rgbText.Alignment = fyne.TextAlignCenter
			rgbText.TextSize = 11
			offsetEntry := newCommitEntry()

			rowContent := container.New(&weightedGridLayout{
				cols:    6,
				weights: []float32{0.55, 1.0, 1.7, 1.6, 2.0, 1.45},
			},
				container.NewCenter(idxText),
				container.NewCenter(statusCheck),
				coordEntry,
				container.NewPadded(swatch),
				container.NewCenter(rgbText),
				offsetEntry,
			)
			return newClickableTableRow(transparent, rowContent, nil)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			row := item.(*ClickableTableRow)
			rowContent := row.content
			idxText := rowContent.Objects[0].(*fyne.Container).Objects[0].(*canvas.Text)
			statusCheck := rowContent.Objects[1].(*fyne.Container).Objects[0].(*ColorCheck)
			coordEntry := rowContent.Objects[2].(*commitEntry)
			swatch := rowContent.Objects[3].(*fyne.Container).Objects[0].(*canvas.Rectangle)
			rgbText := rowContent.Objects[4].(*fyne.Container).Objects[0].(*canvas.Text)
			offsetEntry := rowContent.Objects[5].(*commitEntry)

			textColor := getTextColor(isDarkTheme)
			idxText.Color = textColor
			rgbText.Color = textColor
			idxText.Text = strconv.Itoa(id + 1)

			if id < len(colorPoints) {
				point := colorPoints[id]
				coordEntry.Enable()
				coordEntry.onCommit = func(value string) {
					commitColorPointPosition(id, value)
				}
				coordEntry.OnSubmitted = coordEntry.onCommit
				coordEntry.SetText(point.Position)
				rgbText.Text = point.Color
				offsetEntry.Enable()
				offsetEntry.onCommit = func(value string) {
					if id >= len(colorPoints) {
						return
					}
					colorPoints[id].Offset = strings.TrimSpace(value)
					updateTableSelection()
				}
				offsetEntry.OnSubmitted = offsetEntry.onCommit
				offsetEntry.SetText(point.Offset)
				swatch.FillColor = hexToColor(point.Color)
				statusCheck.Color = hexToColor(point.Color)
				statusCheck.Checked = point.Selected
				statusCheck.OnChanged = func(checked bool) {
					if id >= len(colorPoints) {
						return
					}
					colorPoints[id].Selected = checked
					updateTableSelection()
				}
			} else {
				coordEntry.onCommit = nil
				coordEntry.OnSubmitted = nil
				coordEntry.SetText("")
				coordEntry.Disable()
				rgbText.Text = ""
				offsetEntry.onCommit = nil
				offsetEntry.OnSubmitted = nil
				offsetEntry.SetText("")
				offsetEntry.Disable()
				swatch.FillColor = color.Black
				statusCheck.Color = color.RGBA{240, 240, 240, 255}
				statusCheck.Checked = false
				statusCheck.OnChanged = func(bool) {}
			}

			idxText.Refresh()
			coordEntry.Refresh()
			rgbText.Refresh()
			offsetEntry.Refresh()
			swatch.Refresh()
			statusCheck.Refresh()
		},
	)
	tableScroll := container.NewVScroll(tableContent)
	tableBlock := container.NewBorder(tableHeader, nil, nil, nil, tableScroll)
	tableArea := newFixedHeightContainer(tableBlock, 230)

	clearAllBtn := widget.NewButton("清除所有 (CTRL+E)", func() {
		w.Canvas().Unfocus()
		if imageViewer != nil {
			imageViewer.ClearMarks()
			return
		}
		colorPoints = colorPoints[:0]
		updateTableSelection()
	})
	uniformOffsetEntry := makeEntry(userConfig.UniformOffset)
	uniformOffsetEntry.OnChanged = func(value string) {
		defaultColorPointOffset = strings.TrimSpace(value)
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	}
	defaultColorPointOffset = strings.TrimSpace(uniformOffsetEntry.Text)
	copyCoordsBtn := widget.NewButton("复制坐标", func() {
		w.Canvas().Unfocus()
		w.Clipboard().SetContent(colorPointCoordinatesText())
	})
	pasteCoordsBtn := widget.NewButton("粘贴坐标", func() {
		w.Canvas().Unfocus()
		points := parsePointPositionsText(w.Clipboard().Content())
		if len(points) == 0 {
			return
		}
		defaultColorPointOffset = strings.TrimSpace(uniformOffsetEntry.Text)
		replaceColorPointsByPositions(points, defaultColorPointOffset)
	})
	uniformOffsetBtn := widget.NewButton("统一偏色", func() {
		defaultColorPointOffset = strings.TrimSpace(uniformOffsetEntry.Text)
		for i := range colorPoints {
			colorPoints[i].Offset = defaultColorPointOffset
		}
		updateTableSelection()
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	clearOffsetBtn := widget.NewButton("清除偏色", func() {
		uniformOffsetEntry.SetText("000000")
		defaultColorPointOffset = "000000"
		for i := range colorPoints {
			colorPoints[i].Offset = defaultColorPointOffset
		}
		updateTableSelection()
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})

	precisionEntry := makeEntry(userConfig.Precision)
	precisionEntry.OnChanged = func(string) {
		refreshImagesAPIFields()
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	}
	functionSelect := widget.NewSelect([]string{"FindMultiColors", "FindColor", "FindMultiColorsAll", "CmpColor"}, func(string) {
		if updateImagesAPIFields != nil {
			updateImagesAPIFields()
		}
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	functionSelect.SetSelected(normalizeImagesFunctionName(userConfig.FunctionMode))
	directionSelect := widget.NewSelect([]string{
		"0: 从左到右，从上到下",
		"1: 从右到左，从上到下",
		"2: 从左到右，从下到上",
		"3: 从右到左，从下到上",
	}, func(string) {
		if updateImagesAPIFields != nil {
			updateImagesAPIFields()
		}
		if saveCurrentConfig != nil {
			saveCurrentConfig()
		}
	})
	if directionValue(userConfig.DirectionMode) == 1 {
		directionSelect.SetSelected("1: 从右到左，从上到下")
	} else if directionValue(userConfig.DirectionMode) == 2 {
		directionSelect.SetSelected("2: 从左到右，从下到上")
	} else if directionValue(userConfig.DirectionMode) == 3 {
		directionSelect.SetSelected("3: 从右到左，从下到上")
	} else {
		directionSelect.SetSelected("0: 从左到右，从上到下")
	}
	colorEntry := makeEntry("")
	paramsEntry := makeEntry("")
	resultEntry := widget.NewMultiLineEntry()
	resultEntry.SetPlaceHolder("结果")
	resultEntry.SetMinRowsVisible(6)
	updateImagesAPIFields = func() string {
		colorText, paramsText, code := buildImagesAPICode(functionSelect.Selected, precisionEntry.Text, directionSelect.Selected)
		colorEntry.SetText(colorText)
		paramsEntry.SetText(paramsText)
		resultEntry.SetText(code)
		return code
	}
	saveCurrentConfig = func() {
		saveUserConfigSilently(UserConfig{
			Precision:     strings.TrimSpace(precisionEntry.Text),
			UniformOffset: strings.TrimSpace(uniformOffsetEntry.Text),
			PickCount:     strings.TrimSpace(pickCountEntry.Text),
			PickMode:      pickModeSelect.Selected,
			FunctionMode:  functionSelect.Selected,
			DirectionMode: directionSelect.Selected,
			ShowMagnifier: showMagnifierCheck.Checked,
			AutoCopyRange: autoCopyRangeCheck.Checked,
			ApplyRange:    applyRangeCheck.Checked,
			GridMode:      gridModeEnabled,
			GridCols:      gridColsValue,
			GridRows:      gridRowsValue,
			GridSpacing:   gridSpacingValue,

			FormatTemplates: copyAPIFormatTemplates(apiFormatTemplates),
		})
	}

	toolForm := container.NewVBox(
		container.NewGridWithColumns(2, clearAllBtn, container.NewGridWithColumns(2, copyCoordsBtn, pasteCoordsBtn)),
		container.NewBorder(nil, nil, container.NewHBox(clearOffsetBtn, uniformOffsetBtn), nil, uniformOffsetEntry),
		container.NewAppTabs(
			container.NewTabItem("图色面板", container.NewVBox(
				container.NewBorder(nil, nil, widget.NewLabel("精度"), nil, precisionEntry),
				container.NewBorder(nil, nil, widget.NewLabel("函数"), nil, functionSelect),
				container.NewBorder(nil, nil, widget.NewLabel("方向"), nil, directionSelect),
				container.NewBorder(nil, nil, widget.NewLabel("颜色"), widget.NewButton("复制颜色", func() {
					w.Clipboard().SetContent(colorEntry.Text)
				}), colorEntry),
				container.NewBorder(nil, nil, widget.NewLabel("参数"), container.NewHBox(widget.NewButton("格式", func() {
					showAPIFormatDialog(w, functionSelect.Selected, saveCurrentConfig)
				}), widget.NewButton("复制参数", func() {
					w.Clipboard().SetContent(paramsEntry.Text)
				})), paramsEntry),
				container.NewBorder(nil, nil, widget.NewLabel("结果"), nil, resultEntry),
				container.NewGridWithColumns(4, genBtn, makeButton("找色测试"), makeButton("代码测试"), makeButton("图片查找")),
			)),
			container.NewTabItem("点阵OCR", container.NewCenter(widget.NewLabel("点阵OCR布局待实现"))),
			container.NewTabItem("光学OCR", container.NewCenter(widget.NewLabel("光学OCR布局待实现"))),
		),
	)

	rightToolPanel := container.NewBorder(tableArea, nil, nil, nil, toolForm)
	rightTabs := container.NewAppTabs(
		container.NewTabItem("图色工具", rightToolPanel),
		container.NewTabItem("节点工具", container.NewCenter(widget.NewLabel("节点工具布局待实现"))),
	)
	rightPanel := makeFixedPanel(340, container.NewVScroll(rightTabs))

	// 左工具栏固定宽度；中间图像区和右工具栏保持可拖拽
	centerRightSplit := container.NewHSplit(tabs, rightPanel)
	centerRightSplit.Offset = 0.70

	mainContent := container.NewBorder(nil, nil, leftPanel, nil, centerRightSplit)
	windowContent := container.NewPadded(mainContent)

	// 设置窗口内容并显示
	w.SetPadded(false)
	w.SetContent(windowContent)
	w.Resize(mainWindowSize)
	w.CenterOnScreen()

	// 设置拖放图片功能
	w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		// 处理拖放的文件
		for _, uri := range uris {
			filePath := uri.Path()
			ext := strings.ToLower(filepath.Ext(filePath))

			// 检查是否为支持的图片格式
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".bmp" {
				continue
			}

			// 在 goroutine 中加载图片
			go func(path string, extension string) {
				// 读取文件内容
				data, err := ioutil.ReadFile(path)
				if err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("读取文件失败: %v", err), w)
					})
					return
				}

				// 根据文件扩展名解码图像
				var img image.Image

				switch extension {
				case ".png":
					img, err = png.Decode(bytes.NewReader(data))
				case ".jpg", ".jpeg":
					img, err = jpeg.Decode(bytes.NewReader(data))
				case ".bmp":
					img, err = bmp.Decode(bytes.NewReader(data))
				default:
					// 尝试自动检测格式
					img, _, err = image.Decode(bytes.NewReader(data))
				}

				if err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("解码图像失败: %v", err), w)
					})
					return
				}

				// 转换为NRGBA格式
				img = convertToNRGBA(img)

				// 在主线程中更新UI
				fyne.Do(func() {
					// 保存当前标签页的数据
					saveCurrentTabData()

					// 创建新的图像查看器和标签页
					newImageViewer := NewImageViewer()
					newMagnifier := NewMagnifierWidget()

					newImgContainer := container.New(&topLeftLayout{}, newImageViewer)
					newScrollContainer := container.NewScroll(newImgContainer)

					newImageViewer.scrollContainer = newScrollContainer
					newImageViewer.magnifier = newMagnifier
					configureImageViewer(newImageViewer)
					newImageViewer.SetImage(img)

					// 创建新标签页
					newScrollWithMagnifier := container.NewStack(newScrollContainer, newMagnifier)
					tabCounter++

					// 使用文件名作为标签名称
					tabName := filepath.Base(path)
					newTab := container.NewTabItem(tabName, newScrollWithMagnifier)

					// 初始化新标签页的数据
					tabDataMap[newTab] = &TabData{
						colorPoints:        make([]ColorPoint, 0),
						markRects:          make([]MarkRect, 0),
						manualRectSelected: false,
						imageViewer:        newImageViewer,
						generatedCode:      "",
					}

					tabs.Append(newTab)
					tabs.Select(newTab)

					// 更新当前标签页引用
					currentTab = newTab

					// 更新当前imageViewer引用
					imageViewer = newImageViewer
					fitImageToView(newImageViewer)

					// 清空颜色点列表和矩形区域
					colorPoints = make([]ColorPoint, 0)
					if rectCoordEntry != nil {
						rectCoordEntry.SetText(defaultRangeText)
					}
					if codeDisplayEntry != nil {
						codeDisplayEntry.SetText("")
					}

					// 刷新表格
					if refreshColorList != nil {
						refreshColorList()
					}
				})
			}(filePath, ext)

			// 只处理第一个有效的图片文件
			break
		}
	})

	// 确保初始显示与当前系统主题匹配
	updateTableHeader()
	updateTableSelection()

	w.ShowAndRun()
}

// 带权重的网格布局，让不同列有不同的宽度
type weightedGridLayout struct {
	cols    int
	weights []float32 // 每列的权重
}

func (g *weightedGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for i, obj := range objects {
		if i >= g.cols {
			break
		}
		childSize := obj.MinSize()
		w += childSize.Width * g.weights[i]
		h = fyne.Max(h, childSize.Height)
	}
	return fyne.NewSize(w, h)
}

func (g *weightedGridLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	// 计算总权重
	totalWeight := float32(0)
	for _, weight := range g.weights {
		totalWeight += weight
	}

	// 计算单位宽度
	unitWidth := containerSize.Width / totalWeight

	// 布局对象
	x := float32(0)
	for i, obj := range objects {
		if i >= g.cols {
			break
		}

		// 计算此列宽度
		colWidth := unitWidth * g.weights[i]

		// 设置对象大小和位置
		obj.Move(fyne.NewPos(x, 0))
		obj.Resize(fyne.NewSize(colWidth, containerSize.Height))

		// 更新x坐标
		x += colWidth
	}
}

// 自定义布局，确保内容始终在左上角
type topLeftLayout struct{}

func (t *topLeftLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, obj := range objects {
		objSize := obj.MinSize()
		minSize.Width = fyne.Max(minSize.Width, objSize.Width)
		minSize.Height = fyne.Max(minSize.Height, objSize.Height)
	}
	return minSize
}

func (t *topLeftLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0) // 始终从左上角(0,0)开始放置
	for _, obj := range objects {
		size := obj.MinSize()
		obj.Resize(size)
		obj.Move(pos)
	}
}

type compactPaddedLayout struct {
	padding float32
}

func (p *compactPaddedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(p.padding*2, p.padding*2)
	}

	size := objects[0].MinSize()
	return size.Add(fyne.NewSize(p.padding*2, p.padding*2))
}

func (p *compactPaddedLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if len(objects) == 0 {
		return
	}

	innerSize := fyne.NewSize(
		fyne.Max(0, containerSize.Width-p.padding*2),
		fyne.Max(0, containerSize.Height-p.padding*2),
	)
	objects[0].Move(fyne.NewPos(p.padding, p.padding))
	objects[0].Resize(innerSize)
}

type fixedContentWidthLayout struct {
	width float32
}

func (l *fixedContentWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(1)
	if len(objects) > 0 {
		height = objects[0].MinSize().Height
	}
	return fyne.NewSize(l.width, height)
}

func (l *fixedContentWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(l.width, containerSize.Height))
}

// 固定宽度布局
type fixedWidthLayout struct {
	width           float32
	padding         float32 // 水平内边距
	verticalSpacing float32 // 垂直间距
}

func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(0)
	for i, o := range objects {
		childSize := o.MinSize()
		height += childSize.Height

		// 除了最后一个元素，其他元素后面都加上间距
		if i < len(objects)-1 {
			height += f.verticalSpacing
		}
	}
	return fyne.NewSize(f.width, height)
}

func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	for i, o := range objects {
		var size fyne.Size
		minSize := o.MinSize()

		if _, isBorder := o.(*fyne.Container); isBorder {
			// 如果是BorderLayout容器，让它占据整个高度
			size = fyne.NewSize(f.width, containerSize.Height)
		} else {
			// 所有控件宽度相同，但会应用内边距
			size = fyne.NewSize(f.width-(f.padding*2), minSize.Height)
		}

		// 对大多数控件应用水平内边距
		objPos := fyne.NewPos(pos.X+f.padding, pos.Y)

		// 布局和尺寸调整
		o.Move(objPos)
		o.Resize(size)

		// 更新下一个元素的位置，添加垂直间距
		pos = pos.Add(fyne.NewPos(0, minSize.Height))
		if i < len(objects)-1 {
			pos = pos.Add(fyne.NewPos(0, f.verticalSpacing))
		}
	}
}

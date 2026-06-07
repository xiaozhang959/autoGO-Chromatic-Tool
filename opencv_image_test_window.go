package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
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

const openCVLowSimWarningText = "相似度过低可能导致识别结果不准确"

var openCVImageTestWindow fyne.Window

func openOpenCVImageTestWindow(parent fyne.Window, defaultSimText string) {
	if openCVImageTestWindow != nil {
		openCVImageTestWindow.Show()
		openCVImageTestWindow.RequestFocus()
		return
	}

	a := fyne.CurrentApp()
	w := a.NewWindow("AutoGo 找图测试")
	openCVImageTestWindow = w
	w.SetOnClosed(func() {
		if openCVImageTestWindow == w {
			openCVImageTestWindow = nil
		}
	})

	functions := []string{
		openCVImageFuncFindImage,
		openCVImageFuncFindImageFromImage,
		openCVImageFuncFindImageAll,
	}

	var templateBytes []byte
	var templateImg image.Image
	templateName := "template.png"

	x1, y1, x2, y2 := regionValuesFromEntry()
	rangeEntry := newOpenCVTestEntry(formatOpenCVRange(x1, y1, x2, y2))

	if strings.TrimSpace(defaultSimText) == "" {
		defaultSimText = "0.9"
	}
	simEntry := newOpenCVTestEntry(defaultSimText)
	simWarning := newOpenCVLowSimWarningIcon(w.Canvas())
	updateSimWarning := func() {
		if shouldShowOpenCVLowSimWarning(simEntry.Text) {
			simWarning.Show()
			return
		}
		simWarning.Hide()
		simWarning.hidePopup()
	}
	simEntry.OnChanged = func(string) {
		updateSimWarning()
	}
	updateSimWarning()
	displayIDEntry := newOpenCVTestEntry("0")
	if isNumeric(selectedDisplayID) {
		displayIDEntry.SetText(selectedDisplayID)
	}

	methodSelect := widget.NewSelect(functions, nil)
	methodSelect.SetSelected(openCVImageFuncFindImage)
	grayCheck := widget.NewCheck("isGray 灰度匹配", nil)
	transparentCheck := widget.NewCheck("isTransparent 透明模板", nil)

	templatePreview := canvas.NewImageFromImage(nil)
	templatePreview.FillMode = canvas.ImageFillContain
	templatePreview.ScaleMode = canvas.ImageScalePixels
	templatePreview.SetMinSize(fyne.NewSize(260, 180))

	infoLabel := widget.NewLabel("请先从主图选区裁剪模板，或加载模板图片。")
	infoLabel.Wrapping = fyne.TextWrapWord

	resultEntry := widget.NewMultiLineEntry()
	resultEntry.TextStyle = fyne.TextStyle{Monospace: true}
	resultEntry.SetMinRowsVisible(5)
	resultEntry.SetPlaceHolder("测试结果会显示在这里")

	codeEntry := widget.NewMultiLineEntry()
	codeEntry.TextStyle = fyne.TextStyle{Monospace: true}
	codeEntry.SetMinRowsVisible(7)
	codeEntry.SetPlaceHolder("生成的 AutoGo opencv 调用代码会显示在这里")

	updateTemplate := func(img image.Image, data []byte, name string) {
		templateImg = img
		templateBytes = data
		if strings.TrimSpace(name) != "" {
			templateName = name
		}
		templatePreview.Image = templateImg
		if templateImg != nil {
			b := templateImg.Bounds()
			templatePreview.SetMinSize(fyne.NewSize(float32(max(120, b.Dx())), float32(max(90, b.Dy()))))
			infoLabel.SetText(fmt.Sprintf("模板: %s | %dx%d px | 后端: %s",
				templateName, b.Dx(), b.Dy(), openCVImageMatchBackendName()))
		}
		templatePreview.Refresh()
	}
	cropTemplateFromRect := func(img image.Image, rect image.Rectangle) {
		cropped := cropImage(img, rect)
		data, err := encodeOpenCVTemplatePNG(cropped)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		updateTemplate(cropped, data, fmt.Sprintf("选区_%dx%d.png", rect.Dx(), rect.Dy()))
	}

	cropTemplateButton := widget.NewButtonWithIcon("从当前选区裁剪模板", theme.ContentCutIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入。", w)
			return
		}
		rect, ok := currentOpenCVSelectedRect()
		if !ok {
			dialog.ShowInformation("提示", "请先在主窗口图像上使用“范围”工具框选区域", w)
			return
		}
		cropTemplateFromRect(imageViewer.image, rect)
	})
	cropTemplateButton.Importance = widget.HighImportance

	goCropTemplateButton := widget.NewButtonWithIcon("去裁剪模板", theme.ContentCutIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入。", w)
			return
		}
		if parent != nil {
			parent.RequestFocus()
		}
		viewer := imageViewer
		viewer.SetRangeSelectModeWithCallback(func(rect image.Rectangle) {
			if imageViewer != viewer || viewer.image == nil {
				fyne.Do(func() {
					dialog.ShowInformation("提示", "当前图像已切换，请重新框选模板区域。", w)
				})
				return
			}
			rect = normalizePickRect(viewer.image, rect)
			if rect.Empty() {
				fyne.Do(func() {
					dialog.ShowInformation("提示", "选择的模板区域无效。", w)
				})
				return
			}
			fyne.Do(func() {
				cropTemplateFromRect(viewer.image, rect)
				w.RequestFocus()
			})
		})
		infoLabel.SetText("请在主窗口图像上拖拽选择模板区域。")
	})

	loadTemplateButton := widget.NewButtonWithIcon("加载模板图片", theme.FolderOpenIcon(), func() {
		go func() {
			filePath, err := nativedialog.File().
				Filter("图片文件", "png", "jpg", "jpeg", "bmp").
				Title("选择模板图片").
				Load()
			if err != nil {
				return
			}
			data, err := os.ReadFile(filePath)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("读取模板失败: %v", err), w) })
				return
			}
			img, err := decodeOpenCVTemplateBytes(data)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}
			fyne.Do(func() {
				updateTemplate(img, data, filepath.Base(filePath))
			})
		}()
	})

	copyCodeButton := widget.NewButtonWithIcon("复制代码", theme.ContentCopyIcon(), func() {
		if strings.TrimSpace(codeEntry.Text) == "" {
			return
		}
		w.Clipboard().SetContent(codeEntry.Text)
		dialog.ShowInformation("已复制", "找图代码已复制到剪贴板。", w)
	})

	clearMarksButton := widget.NewButtonWithIcon("清除标记", theme.DeleteIcon(), func() {
		if imageViewer != nil {
			imageViewer.ClearFindTestHighlights()
		}
	})

	resetRangeButton := widget.NewButton("重置范围", func() {
		x1, y1, x2, y2 := regionValuesFromEntry()
		rangeEntry.SetText(formatOpenCVRange(x1, y1, x2, y2))
		infoLabel.SetText("已重置查找范围。")
	})
	rangeButton := widget.NewButton("范围 (Ctrl+R)", func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入。", w)
			return
		}
		if parent != nil {
			parent.RequestFocus()
		}
		viewer := imageViewer
		viewer.SetRangeSelectModeWithCallback(func(rect image.Rectangle) {
			if imageViewer != viewer || viewer.image == nil {
				fyne.Do(func() {
					dialog.ShowInformation("提示", "当前图像已切换，请重新框选查找范围。", w)
				})
				return
			}
			rect = normalizePickRect(viewer.image, rect)
			if rect.Empty() {
				fyne.Do(func() {
					dialog.ShowInformation("提示", "选择的查找范围无效。", w)
				})
				return
			}
			fyne.Do(func() {
				rangeEntry.SetText(formatOpenCVRange(rect.Min.X, rect.Min.Y, rect.Max.X-1, rect.Max.Y-1))
				infoLabel.SetText("已回填查找范围，可开始找图测试。")
				w.RequestFocus()
			})
		})
		infoLabel.SetText("请在主窗口图像上拖拽选择查找范围。")
	})

	runButton := widget.NewButtonWithIcon("开始找图测试", theme.MediaPlayIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入。", w)
			return
		}
		functionName := normalizeOpenCVImageFunctionName(methodSelect.Selected)
		opts, err := openCVOptionsFromEntries(functionName, rangeEntry, simEntry, grayCheck, transparentCheck)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		result, err := runOpenCVImageMatch(imageViewer.image, templateBytes, opts)
		if err != nil {
			resultEntry.SetText("错误：" + err.Error())
			if imageViewer != nil {
				imageViewer.ClearFindTestHighlights()
			}
		} else {
			resultEntry.SetText(formatOpenCVImageTestResult(functionName, result))
			setOpenCVFindTestHighlightRects(imageViewer, openCVMatchHighlightRects(result))
		}

		displayID := parseOpenCVIntEntry(displayIDEntry, 0)
		codeEntry.SetText(buildOpenCVImageTestCode(functionName, opts, displayID, templateName))
	})
	runButton.Importance = widget.HighImportance

	backendLabel := widget.NewLabel(openCVImageBackendStatusText())
	backendLabel.Wrapping = fyne.TextWrapWord

	leftPanel := container.NewPadded(container.NewVBox(
		backendLabel,
		widget.NewSeparator(),
		widget.NewLabel("函数"),
		methodSelect,
		widget.NewLabel("查找范围"),
		container.NewGridWithColumns(2, rangeButton, resetRangeButton),
		rangeEntry,
		container.NewBorder(nil, nil, container.NewHBox(widget.NewLabel("相似度 sim"), simWarning), nil, simEntry),
		container.NewBorder(nil, nil, widget.NewLabel("displayId"), nil, displayIDEntry),
		grayCheck,
		transparentCheck,
		widget.NewSeparator(),
		cropTemplateButton,
		goCropTemplateButton,
		loadTemplateButton,
		widget.NewSeparator(),
		runButton,
		clearMarksButton,
		layout.NewSpacer(),
	))

	templateArea := container.NewBorder(
		container.NewVBox(widget.NewLabelWithStyle("模板预览", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), infoLabel),
		nil, nil, nil,
		container.NewScroll(templatePreview),
	)
	resultArea := container.NewVSplit(
		container.NewBorder(widget.NewLabelWithStyle("测试结果", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, resultEntry),
		container.NewBorder(
			container.NewHBox(widget.NewLabelWithStyle("生成代码", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), copyCodeButton),
			nil, nil, nil, codeEntry,
		),
	)
	resultArea.Offset = 0.42

	body := container.NewHSplit(leftPanel, container.NewVSplit(templateArea, resultArea))
	body.Offset = 0.30
	content := container.NewPadded(body)
	w.SetContent(content)
	w.Resize(initialWindowSize(0.66, 0.70))
	w.CenterOnScreen()
	w.Show()
}

func newOpenCVTestEntry(value string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(value)
	return entry
}

func formatOpenCVRange(x1, y1, x2, y2 int) string {
	return fmt.Sprintf("%d,%d,%d,%d", x1, y1, x2, y2)
}

func shouldShowOpenCVLowSimWarning(text string) bool {
	sim, err := strconv.ParseFloat(strings.TrimSpace(text), 32)
	return err == nil && sim < 0.5
}

type openCVLowSimWarningIcon struct {
	widget.BaseWidget

	canvas fyne.Canvas
	icon   *widget.Icon
	popup  *widget.PopUp
}

func newOpenCVLowSimWarningIcon(canvas fyne.Canvas) *openCVLowSimWarningIcon {
	warning := &openCVLowSimWarningIcon{
		canvas: canvas,
		icon:   widget.NewIcon(theme.WarningIcon()),
	}
	warning.ExtendBaseWidget(warning)
	warning.Hide()
	return warning
}

func (w *openCVLowSimWarningIcon) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.icon)
}

func (w *openCVLowSimWarningIcon) MouseIn(event *desktop.MouseEvent) {
	if w.canvas == nil {
		return
	}
	w.hidePopup()

	content := container.NewPadded(widget.NewLabel(openCVLowSimWarningText))
	w.popup = widget.NewPopUp(content, w.canvas)

	pos := w.Position().Add(fyne.NewPos(w.Size().Width+theme.Padding(), 0))
	if event != nil {
		pos = event.AbsolutePosition.Add(fyne.NewPos(theme.Padding(), theme.Padding()))
	}
	w.popup.ShowAtPosition(pos)
}

func (w *openCVLowSimWarningIcon) MouseMoved(*desktop.MouseEvent) {}

func (w *openCVLowSimWarningIcon) MouseOut() {
	w.hidePopup()
}

func (w *openCVLowSimWarningIcon) hidePopup() {
	if w.popup == nil {
		return
	}
	w.popup.Hide()
	w.popup = nil
}

func currentOpenCVSelectedRect() (image.Rectangle, bool) {
	if imageViewer == nil || imageViewer.image == nil || len(imageViewer.markRects) == 0 {
		return image.Rectangle{}, false
	}
	mark := imageViewer.markRects[0]
	rect := image.Rect(
		min(mark.X1, mark.X2),
		min(mark.Y1, mark.Y2),
		max(mark.X1, mark.X2),
		max(mark.Y1, mark.Y2),
	)
	rect = rect.Intersect(imageViewer.image.Bounds())
	if rect.Empty() {
		return image.Rectangle{}, false
	}
	return rect, true
}

func openCVOptionsFromEntries(functionName string, rangeEntry, simEntry *widget.Entry, grayCheck, transparentCheck *widget.Check) (openCVMatchOptions, error) {
	x1, y1, x2, y2, err := parseOpenCVRangeEntry(rangeEntry)
	if err != nil {
		return openCVMatchOptions{}, err
	}
	sim, err := strconv.ParseFloat(strings.TrimSpace(simEntry.Text), 32)
	if err != nil {
		return openCVMatchOptions{}, fmt.Errorf("sim 必须是数字")
	}
	if functionName == openCVImageFuncFindImageFromImage {
		x1, y1, x2, y2 = 0, 0, 0, 0
	}
	return openCVMatchOptions{
		X1:            x1,
		Y1:            y1,
		X2:            x2,
		Y2:            y2,
		IsGray:        grayCheck.Checked,
		IsTransparent: transparentCheck.Checked,
		Sim:           float32(sim),
		FindAll:       functionName == openCVImageFuncFindImageAll,
		MaxResults:    defaultOpenCVImageMaxResults,
	}, nil
}

func parseOpenCVRangeEntry(entry *widget.Entry) (int, int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(entry.Text), ",")
	if len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("查找范围格式应为 x1,y1,x2,y2")
	}
	values := [4]int{}
	for i, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("查找范围格式应为 x1,y1,x2,y2，且坐标必须是整数")
		}
		values[i] = value
	}
	return values[0], values[1], values[2], values[3], nil
}

func parseOpenCVIntEntry(entry *widget.Entry, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(entry.Text))
	if err != nil {
		return fallback
	}
	return value
}

func formatOpenCVImageTestResult(functionName string, result openCVMatchResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("backend: %s\n", result.Backend))
	sb.WriteString(fmt.Sprintf("bestScore: %.6f\n", result.BestScore))
	sb.WriteString(fmt.Sprintf("template: %dx%d\n", result.TemplateSize.X, result.TemplateSize.Y))
	if len(result.Matches) == 0 {
		if functionName == openCVImageFuncFindImageAll {
			sb.WriteString("points: []")
		} else {
			sb.WriteString("result: -1,-1")
		}
		return sb.String()
	}

	if functionName != openCVImageFuncFindImageAll {
		match := result.Matches[0]
		sb.WriteString(fmt.Sprintf("result: %d,%d\nscore: %.6f", match.Point.X, match.Point.Y, match.Score))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("count: %d\n", len(result.Matches)))
	for i, match := range result.Matches {
		sb.WriteString(fmt.Sprintf("%d: %d,%d score=%.6f\n", i+1, match.Point.X, match.Point.Y, match.Score))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func buildOpenCVImageTestCode(functionName string, opts openCVMatchOptions, displayID int, templateName string) string {
	if strings.TrimSpace(templateName) == "" {
		templateName = "template.png"
	}
	readTemplate := fmt.Sprintf("templateBytes, err := os.ReadFile(%q)\nif err != nil {\n\tpanic(err)\n}", templateName)
	sim := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", opts.Sim), "0"), ".")
	if sim == "" {
		sim = "0"
	}
	switch functionName {
	case openCVImageFuncFindImageFromImage:
		return fmt.Sprintf("%s\nx, y := opencv.FindImageFromImage(img, &templateBytes, %t, %t, %s)",
			readTemplate, opts.IsGray, opts.IsTransparent, sim)
	case openCVImageFuncFindImageAll:
		return fmt.Sprintf("%s\npoints := opencv.FindImageAll(%d, %d, %d, %d, &templateBytes, %t, %t, %s, %d)",
			readTemplate, opts.X1, opts.Y1, opts.X2, opts.Y2, opts.IsGray, opts.IsTransparent, sim, displayID)
	default:
		return fmt.Sprintf("%s\nx, y := opencv.FindImage(%d, %d, %d, %d, &templateBytes, %t, %t, %s, %d)",
			readTemplate, opts.X1, opts.Y1, opts.X2, opts.Y2, opts.IsGray, opts.IsTransparent, sim, displayID)
	}
}

func openCVImageBackendStatusText() string {
	if openCVImageMatchBackendAvailable() {
		return "后端: OpenCV CGO 已启用"
	}
	return "后端: OpenCV CGO 未启用。使用真实找图前，请将 OpenCV 放入 third_party/opencv/windows-amd64 并用 -tags opencv_cgo 构建。"
}

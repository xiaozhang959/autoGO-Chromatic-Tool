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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativedialog "github.com/sqweek/dialog"
)

func openOpenCVImageTestWindow(parent fyne.Window, defaultSimText string) {
	a := fyne.CurrentApp()
	w := a.NewWindow("AutoGo 找图测试")

	functions := []string{
		openCVImageFuncFindImage,
		openCVImageFuncFindImageFromImage,
		openCVImageFuncFindImageAll,
	}

	var templateBytes []byte
	var templateImg image.Image
	templateName := "template.png"

	x1, y1, x2, y2 := regionValuesFromEntry()
	x1Entry := newOpenCVTestEntry(strconv.Itoa(x1))
	y1Entry := newOpenCVTestEntry(strconv.Itoa(y1))
	x2Entry := newOpenCVTestEntry(strconv.Itoa(x2))
	y2Entry := newOpenCVTestEntry(strconv.Itoa(y2))

	if strings.TrimSpace(defaultSimText) == "" {
		defaultSimText = "0.9"
	}
	simEntry := newOpenCVTestEntry(defaultSimText)
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

	cropTemplateButton := widget.NewButtonWithIcon("从当前选区裁剪模板", theme.ContentCutIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入。", w)
			return
		}
		rect, ok := currentOpenCVSelectedRect()
		if !ok {
			dialog.ShowInformation("提示", "请先在主窗口图像上拖拽框选模板区域。", w)
			return
		}
		cropped := cropImage(imageViewer.image, rect)
		data, err := encodeOpenCVTemplatePNG(cropped)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		updateTemplate(cropped, data, fmt.Sprintf("选区_%dx%d.png", rect.Dx(), rect.Dy()))
	})
	cropTemplateButton.Importance = widget.HighImportance

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

	runButton := widget.NewButtonWithIcon("开始找图测试", theme.MediaPlayIcon(), func() {
		if imageViewer == nil || imageViewer.image == nil {
			dialog.ShowInformation("提示", "主窗口没有图片，请先截图或载入。", w)
			return
		}
		functionName := normalizeOpenCVImageFunctionName(methodSelect.Selected)
		opts, err := openCVOptionsFromEntries(functionName, x1Entry, y1Entry, x2Entry, y2Entry, simEntry, grayCheck, transparentCheck)
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

	rangeGrid := container.NewGridWithColumns(4,
		container.NewBorder(widget.NewLabel("x1"), nil, nil, nil, x1Entry),
		container.NewBorder(widget.NewLabel("y1"), nil, nil, nil, y1Entry),
		container.NewBorder(widget.NewLabel("x2"), nil, nil, nil, x2Entry),
		container.NewBorder(widget.NewLabel("y2"), nil, nil, nil, y2Entry),
	)
	leftPanel := container.New(&fixedWidthLayout{width: 230, padding: 8, verticalSpacing: 6},
		backendLabel,
		widget.NewSeparator(),
		widget.NewLabel("函数"),
		methodSelect,
		widget.NewLabel("查找范围"),
		rangeGrid,
		widget.NewLabel("相似度 sim"),
		simEntry,
		widget.NewLabel("displayId"),
		displayIDEntry,
		grayCheck,
		transparentCheck,
		widget.NewSeparator(),
		cropTemplateButton,
		loadTemplateButton,
		widget.NewSeparator(),
		runButton,
		clearMarksButton,
		layout.NewSpacer(),
	)

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

func openCVOptionsFromEntries(functionName string, x1Entry, y1Entry, x2Entry, y2Entry, simEntry *widget.Entry, grayCheck, transparentCheck *widget.Check) (openCVMatchOptions, error) {
	x1, err := parseRequiredOpenCVIntEntry(x1Entry, "x1")
	if err != nil {
		return openCVMatchOptions{}, err
	}
	y1, err := parseRequiredOpenCVIntEntry(y1Entry, "y1")
	if err != nil {
		return openCVMatchOptions{}, err
	}
	x2, err := parseRequiredOpenCVIntEntry(x2Entry, "x2")
	if err != nil {
		return openCVMatchOptions{}, err
	}
	y2, err := parseRequiredOpenCVIntEntry(y2Entry, "y2")
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

func parseRequiredOpenCVIntEntry(entry *widget.Entry, name string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(entry.Text))
	if err != nil {
		return 0, fmt.Errorf("%s 必须是整数", name)
	}
	return value, nil
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

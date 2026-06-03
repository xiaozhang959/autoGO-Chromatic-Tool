package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AndroidAppInfo struct {
	Name         string   `json:"name"`
	PackageName  string   `json:"packageName"`
	ActivityName string   `json:"activityName"`
	Activities   []string `json:"activities"`
}

type appInfoCopyLabel struct {
	widget.Label
	onTapped func()
}

func newAppInfoCopyLabel(text string, onTapped func()) *appInfoCopyLabel {
	label := &appInfoCopyLabel{onTapped: onTapped}
	label.Text = text
	label.Wrapping = fyne.TextWrapOff
	label.ExtendBaseWidget(label)
	return label
}

func (l *appInfoCopyLabel) Tapped(*fyne.PointEvent) {
	if l.onTapped != nil {
		l.onTapped()
	}
}

func (l *appInfoCopyLabel) TappedSecondary(event *fyne.PointEvent) {
	l.Tapped(event)
}

type appInfoDisplayBox struct {
	widget.BaseWidget

	label    *widget.Label
	text     string
	onTapped func()
}

func newAppInfoDisplayBox(onTapped func()) *appInfoDisplayBox {
	box := &appInfoDisplayBox{
		label:    widget.NewLabel("-"),
		onTapped: onTapped,
	}
	box.label.Wrapping = fyne.TextWrapOff
	box.label.Truncation = fyne.TextTruncateEllipsis
	box.ExtendBaseWidget(box)
	return box
}

func (b *appInfoDisplayBox) Text() string {
	return b.text
}

func (b *appInfoDisplayBox) SetText(text string) {
	b.text = strings.TrimSpace(text)
	displayText := b.text
	if displayText == "" {
		displayText = "-"
	}
	b.label.SetText(displayText)
	b.Refresh()
}

func (b *appInfoDisplayBox) Tapped(*fyne.PointEvent) {
	if b.onTapped != nil {
		b.onTapped()
	}
}

func (b *appInfoDisplayBox) TappedSecondary(event *fyne.PointEvent) {
	b.Tapped(event)
}

func (b *appInfoDisplayBox) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.InputBackgroundColor())
	bg.StrokeColor = theme.InputBorderColor()
	bg.StrokeWidth = 1
	return &appInfoDisplayBoxRenderer{
		box:     b,
		bg:      bg,
		objects: []fyne.CanvasObject{bg, b.label},
	}
}

type appInfoDisplayBoxRenderer struct {
	box     *appInfoDisplayBox
	bg      *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *appInfoDisplayBoxRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	padding := theme.InnerPadding()
	labelSize := fyne.NewSize(size.Width-padding*2, size.Height-padding)
	if labelSize.Width < 0 {
		labelSize.Width = 0
	}
	if labelSize.Height < 0 {
		labelSize.Height = 0
	}
	r.box.label.Move(fyne.NewPos(padding, padding/2))
	r.box.label.Resize(labelSize)
}

func (r *appInfoDisplayBoxRenderer) MinSize() fyne.Size {
	padding := theme.InnerPadding()
	minSize := r.box.label.MinSize()
	return fyne.NewSize(minSize.Width+padding*2, minSize.Height+padding)
}

func (r *appInfoDisplayBoxRenderer) Refresh() {
	r.bg.FillColor = theme.InputBackgroundColor()
	r.bg.StrokeColor = theme.InputBorderColor()
	r.bg.Refresh()
	r.box.label.Refresh()
}

func (r *appInfoDisplayBoxRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *appInfoDisplayBoxRenderer) Destroy() {}

type AndroidAppInfoTool struct {
	window fyne.Window

	getSelectedDevice func() string

	root        fyne.CanvasObject
	searchEntry *widget.Entry
	queryBtn    *widget.Button
	statusLabel *widget.Label
	appList     *widget.List

	nameField      *appInfoDisplayBox
	packageField   *appInfoDisplayBox
	launcherField  *appInfoDisplayBox
	activitiesList *widget.List

	busy         bool
	apps         []AndroidAppInfo
	filteredApps []AndroidAppInfo

	hasSelectedApp       bool
	selectedApp          AndroidAppInfo
	selectedActivities   []string
	syncingListSelection bool
}

func newAndroidAppInfoTool(w fyne.Window, getSelectedDevice func() string) *AndroidAppInfoTool {
	tool := &AndroidAppInfoTool{
		window:            w,
		getSelectedDevice: getSelectedDevice,
		apps:              make([]AndroidAppInfo, 0),
		filteredApps:      make([]AndroidAppInfo, 0),
	}

	tool.searchEntry = widget.NewEntry()
	tool.searchEntry.SetPlaceHolder("搜索 应用名称 / 应用包名 / 界面名称")
	tool.searchEntry.OnChanged = func(string) {
		tool.applySearch()
	}

	tool.queryBtn = widget.NewButton("查询应用", func() {
		tool.Query()
	})
	tool.queryBtn.Importance = widget.MediumImportance

	tool.statusLabel = widget.NewLabel("未查询应用")
	tool.statusLabel.Wrapping = fyne.TextWrapOff

	tool.appList = widget.NewList(
		func() int {
			return len(tool.filteredApps)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabel("")
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			nameLabel.Wrapping = fyne.TextWrapOff

			packageLabel := widget.NewLabel("")
			packageLabel.Wrapping = fyne.TextWrapOff

			activityLabel := widget.NewLabel("")
			activityLabel.Wrapping = fyne.TextWrapOff

			countLabel := widget.NewLabel("")
			countLabel.Wrapping = fyne.TextWrapOff

			return container.NewVBox(nameLabel, packageLabel, activityLabel, countLabel)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(tool.filteredApps) {
				return
			}
			box, ok := item.(*fyne.Container)
			if !ok || len(box.Objects) < 4 {
				return
			}

			nameLabel, _ := box.Objects[0].(*widget.Label)
			packageLabel, _ := box.Objects[1].(*widget.Label)
			activityLabel, _ := box.Objects[2].(*widget.Label)
			countLabel, _ := box.Objects[3].(*widget.Label)
			if nameLabel == nil || packageLabel == nil || activityLabel == nil || countLabel == nil {
				return
			}

			app := tool.filteredApps[id]
			activityName := app.ActivityName
			if activityName == "" {
				activityName = "-"
			}
			nameLabel.SetText(app.Name)
			packageLabel.SetText("包名: " + app.PackageName)
			activityLabel.SetText("启动: " + activityName)
			countLabel.SetText(fmt.Sprintf("页面: %d 个 · 其它: %d 个", len(app.Activities), len(androidAppOtherActivities(app))))
		},
	)
	tool.appList.OnSelected = func(id widget.ListItemID) {
		if tool.syncingListSelection {
			return
		}
		tool.selectFilteredApp(id, false)
	}

	copyNameBtn := widget.NewButton("复制", func() {
		tool.copyText("应用名称", tool.nameField.Text())
	})
	copyPackageBtn := widget.NewButton("复制", func() {
		tool.copyText("应用包名", tool.packageField.Text())
	})
	copyLauncherBtn := widget.NewButton("复制", func() {
		tool.copyText("界面名称", tool.launcherField.Text())
	})
	copyActivitiesBtn := widget.NewButton("复制全部", func() {
		tool.copyText("其它界面", strings.Join(tool.selectedActivities, "\n"))
	})

	tool.nameField = newAppInfoDisplayBox(func() {
		tool.copyText("应用名称", tool.nameField.Text())
	})
	tool.packageField = newAppInfoDisplayBox(func() {
		tool.copyText("应用包名", tool.packageField.Text())
	})
	tool.launcherField = newAppInfoDisplayBox(func() {
		tool.copyText("界面名称", tool.launcherField.Text())
	})

	tool.activitiesList = widget.NewList(
		func() int {
			return len(tool.selectedActivities)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapOff
			copyBtn := widget.NewButton("复制", func() {})
			return container.NewBorder(nil, nil, nil, copyBtn, label)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(tool.selectedActivities) {
				return
			}
			row, ok := item.(*fyne.Container)
			if !ok || len(row.Objects) < 2 {
				return
			}
			label, _ := row.Objects[0].(*widget.Label)
			copyBtn, _ := row.Objects[1].(*widget.Button)
			if label == nil || copyBtn == nil {
				return
			}
			activity := tool.selectedActivities[id]
			label.SetText(activity)
			copyBtn.OnTapped = func() {
				tool.copyText("界面名称", activity)
			}
		},
	)

	detailPanel := container.NewVBox(
		widget.NewLabel("应用详情"),
		container.NewBorder(nil, nil, newAppInfoCopyLabel("名称", func() {
			tool.copyText("应用名称", tool.nameField.Text())
		}), copyNameBtn, tool.nameField),
		container.NewBorder(nil, nil, newAppInfoCopyLabel("包名", func() {
			tool.copyText("应用包名", tool.packageField.Text())
		}), copyPackageBtn, tool.packageField),
		container.NewBorder(nil, nil, newAppInfoCopyLabel("界面", func() {
			tool.copyText("界面名称", tool.launcherField.Text())
		}), copyLauncherBtn, tool.launcherField),
		container.NewBorder(nil, nil, nil, copyActivitiesBtn, newAppInfoCopyLabel("其它界面", func() {
			tool.copyText("其它界面", strings.Join(tool.selectedActivities, "\n"))
		})),
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, tool.activitiesList), 150),
	)

	tool.root = container.NewVBox(
		container.NewBorder(nil, nil, nil, tool.queryBtn, tool.searchEntry),
		container.NewHScroll(tool.statusLabel),
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, tool.appList), 260),
		detailPanel,
	)

	tool.clearSelectedApp()
	return tool
}

func (t *AndroidAppInfoTool) Content() fyne.CanvasObject {
	return t.root
}

func (t *AndroidAppInfoTool) Query() {
	if t.busy {
		return
	}

	device := ""
	if t.getSelectedDevice != nil {
		device = strings.TrimSpace(t.getSelectedDevice())
	}
	if device == "" {
		dialog.ShowInformation("提示", "请先选择已连接的 Android 设备", t.window)
		return
	}

	t.setBusy(true)
	t.setStatus("正在查询已安装应用...")

	go func() {
		apps, err := queryAndroidApps(device)
		fyne.Do(func() {
			t.setBusy(false)
			if err != nil {
				t.setStatus("查询应用失败")
				dialog.ShowError(fmt.Errorf("查询应用失败: %v", err), t.window)
				return
			}
			t.apps = apps
			t.applySearch()
		})
	}()
}

func (t *AndroidAppInfoTool) applySearch() {
	selectedPackage := ""
	if t.hasSelectedApp {
		selectedPackage = t.selectedApp.PackageName
	}

	keyword := ""
	if t.searchEntry != nil {
		keyword = t.searchEntry.Text
	}
	t.filteredApps = filterAndroidApps(t.apps, keyword)
	if t.appList != nil {
		t.appList.Refresh()
	}

	if len(t.filteredApps) == 0 {
		t.clearSelectedApp()
	} else {
		index := androidAppIndexByPackage(t.filteredApps, selectedPackage)
		if index < 0 {
			index = 0
		}
		t.selectFilteredApp(index, true)
	}

	total := len(t.apps)
	matched := len(t.filteredApps)
	if strings.TrimSpace(keyword) == "" {
		t.setStatus(fmt.Sprintf("已查询 %d 个应用", total))
		return
	}
	t.setStatus(fmt.Sprintf("已查询 %d 个应用 · 匹配 %d 个", total, matched))
}

func (t *AndroidAppInfoTool) setBusy(busy bool) {
	t.busy = busy
	if t.queryBtn == nil {
		return
	}
	if busy {
		t.queryBtn.Disable()
		return
	}
	t.queryBtn.Enable()
}

func (t *AndroidAppInfoTool) setStatus(text string) {
	if t.statusLabel != nil {
		t.statusLabel.SetText(text)
	}
}

func (t *AndroidAppInfoTool) selectFilteredApp(id int, updateList bool) {
	if id < 0 || id >= len(t.filteredApps) {
		t.clearSelectedApp()
		return
	}

	app := t.filteredApps[id]
	t.hasSelectedApp = true
	t.selectedApp = app
	t.selectedActivities = androidAppOtherActivities(app)

	if t.nameField != nil {
		t.nameField.SetText(app.Name)
	}
	if t.packageField != nil {
		t.packageField.SetText(app.PackageName)
	}
	if t.launcherField != nil {
		t.launcherField.SetText(app.ActivityName)
	}
	if t.activitiesList != nil {
		t.activitiesList.Refresh()
	}
	if updateList && t.appList != nil {
		t.syncingListSelection = true
		t.appList.Select(id)
		t.syncingListSelection = false
	}
}

func (t *AndroidAppInfoTool) clearSelectedApp() {
	t.hasSelectedApp = false
	t.selectedApp = AndroidAppInfo{}
	t.selectedActivities = nil
	if t.nameField != nil {
		t.nameField.SetText("")
	}
	if t.packageField != nil {
		t.packageField.SetText("")
	}
	if t.launcherField != nil {
		t.launcherField.SetText("")
	}
	if t.activitiesList != nil {
		t.activitiesList.Refresh()
	}
	if t.appList != nil {
		t.syncingListSelection = true
		t.appList.UnselectAll()
		t.syncingListSelection = false
	}
}

func (t *AndroidAppInfoTool) copyText(label, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		t.setStatus(label + "为空，无法复制")
		return
	}
	if t.window != nil {
		t.window.Clipboard().SetContent(text)
	}
	t.setStatus("已复制" + label)
}

func queryAndroidApps(deviceID string) ([]AndroidAppInfo, error) {
	baseDeviceID, _ := splitAndroidDeviceID(deviceID)
	if baseDeviceID == "" {
		return nil, fmt.Errorf("设备 ID 为空")
	}

	ensureCapDexOnDevice(baseDeviceID)
	args := append([]string{"-s", baseDeviceID}, androidCapDexAppInfoArgs()...)
	output, err := adbExecCombined(args...)
	if err != nil {
		return nil, fmt.Errorf("执行 App 信息查询失败: %v", adbErrorWithOutput(err, output))
	}

	apps, err := parseAndroidAppInfoLines(output)
	if err != nil {
		return nil, err
	}
	sortAndroidApps(apps)
	return apps, nil
}

func parseAndroidAppInfoLines(output string) ([]AndroidAppInfo, error) {
	apps := make([]AndroidAppInfo, 0)
	skipped := make([]string, 0)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "{") {
			skipped = append(skipped, line)
			continue
		}

		var app AndroidAppInfo
		if err := json.Unmarshal([]byte(line), &app); err != nil {
			return nil, fmt.Errorf("解析应用 JSON 失败: %v", err)
		}
		app.Name = strings.TrimSpace(app.Name)
		app.PackageName = strings.TrimSpace(app.PackageName)
		app.ActivityName = strings.TrimSpace(app.ActivityName)
		app.Activities = normalizeAndroidAppActivities(app.ActivityName, app.Activities)
		if app.PackageName == "" {
			return nil, fmt.Errorf("应用信息缺少包名: %s", line)
		}
		if app.Name == "" {
			app.Name = app.PackageName
		}
		apps = append(apps, app)
	}

	if len(apps) == 0 && len(skipped) > 0 {
		return nil, fmt.Errorf("未解析到应用信息，非 JSON 输出: %s", trimMiddle(strings.Join(skipped, "\n"), 300))
	}
	if len(skipped) > 0 {
		log.Printf("[app-info] 跳过非 JSON 输出: %q", logPreview(strings.Join(skipped, "\n"), 500))
	}
	return apps, nil
}

func filterAndroidApps(apps []AndroidAppInfo, keyword string) []AndroidAppInfo {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	filtered := make([]AndroidAppInfo, 0, len(apps))
	for _, app := range apps {
		if keyword == "" || androidAppMatches(app, keyword) {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

func androidAppMatches(app AndroidAppInfo, keyword string) bool {
	if strings.Contains(strings.ToLower(app.Name), keyword) ||
		strings.Contains(strings.ToLower(app.PackageName), keyword) ||
		strings.Contains(strings.ToLower(app.ActivityName), keyword) {
		return true
	}
	for _, activity := range app.Activities {
		if strings.Contains(strings.ToLower(activity), keyword) {
			return true
		}
	}
	return false
}

func sortAndroidApps(apps []AndroidAppInfo) {
	sort.SliceStable(apps, func(i, j int) bool {
		leftName := strings.ToLower(apps[i].Name)
		rightName := strings.ToLower(apps[j].Name)
		if leftName != rightName {
			return leftName < rightName
		}
		return apps[i].PackageName < apps[j].PackageName
	})
}

func normalizeAndroidAppActivities(launcher string, activities []string) []string {
	normalized := make([]string, 0, len(activities)+1)
	seen := make(map[string]struct{})
	add := func(activity string) {
		activity = strings.TrimSpace(activity)
		if activity == "" {
			return
		}
		if _, ok := seen[activity]; ok {
			return
		}
		seen[activity] = struct{}{}
		normalized = append(normalized, activity)
	}

	add(launcher)
	for _, activity := range activities {
		add(activity)
	}
	return normalized
}

func androidAppOtherActivities(app AndroidAppInfo) []string {
	launcher := strings.TrimSpace(app.ActivityName)
	others := make([]string, 0, len(app.Activities))
	seen := make(map[string]struct{})
	for _, activity := range app.Activities {
		activity = strings.TrimSpace(activity)
		if activity == "" || activity == launcher {
			continue
		}
		if _, ok := seen[activity]; ok {
			continue
		}
		seen[activity] = struct{}{}
		others = append(others, activity)
	}
	return others
}

func androidAppIndexByPackage(apps []AndroidAppInfo, packageName string) int {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return -1
	}
	for i, app := range apps {
		if app.PackageName == packageName {
			return i
		}
	}
	return -1
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type AndroidAppInfo struct {
	Name         string `json:"name"`
	PackageName  string `json:"packageName"`
	ActivityName string `json:"activityName"`
}

type AndroidAppInfoTool struct {
	window fyne.Window

	getSelectedDevice func() string

	root        fyne.CanvasObject
	searchEntry *widget.Entry
	queryBtn    *widget.Button
	statusLabel *widget.Label
	appList     *widget.List

	busy         bool
	apps         []AndroidAppInfo
	filteredApps []AndroidAppInfo
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

			return container.NewVBox(nameLabel, packageLabel, activityLabel)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(tool.filteredApps) {
				return
			}
			box, ok := item.(*fyne.Container)
			if !ok || len(box.Objects) < 3 {
				return
			}

			nameLabel, _ := box.Objects[0].(*widget.Label)
			packageLabel, _ := box.Objects[1].(*widget.Label)
			activityLabel, _ := box.Objects[2].(*widget.Label)
			if nameLabel == nil || packageLabel == nil || activityLabel == nil {
				return
			}

			app := tool.filteredApps[id]
			activityName := app.ActivityName
			if activityName == "" {
				activityName = "-"
			}
			nameLabel.SetText(app.Name)
			packageLabel.SetText("包名: " + app.PackageName)
			activityLabel.SetText("界面: " + activityName)
		},
	)

	tool.root = container.NewVBox(
		container.NewBorder(nil, nil, nil, tool.queryBtn, tool.searchEntry),
		container.NewHScroll(tool.statusLabel),
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, tool.appList), 520),
	)

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
	keyword := ""
	if t.searchEntry != nil {
		keyword = t.searchEntry.Text
	}
	t.filteredApps = filterAndroidApps(t.apps, keyword)
	if t.appList != nil {
		t.appList.Refresh()
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
	return strings.Contains(strings.ToLower(app.Name), keyword) ||
		strings.Contains(strings.ToLower(app.PackageName), keyword) ||
		strings.Contains(strings.ToLower(app.ActivityName), keyword)
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

package main

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const androidNodeDumpPath = "/sdcard/window_dump.xml"
const compactNodeToolAttrSelectWidth float32 = 40
const compactNodeToolAttrNameWidth float32 = 68
const compactNodeToolAttrFinderWidth float32 = 118
const compactNodeToolTreeRowHeight float32 = 18
const compactNodeToolAttrRowHeight float32 = 28
const compactNodeToolCheckOffsetY float32 = 5
const androidNodeTreeBottomSpacerRows = 2

const androidNodeSelectorFuncFindOnce = "FindOnce"
const androidNodeSelectorFuncFind = "Find"
const androidNodeSelectorFuncWaitFor = "WaitFor"
const androidNodeSelectorDefaultTimeout = "3000"
const androidNodeSelectorDefaultTemplate = "acc := uiacc.New({displayId})\nobj := acc{chain}.{call}"

type AndroidUINode struct {
	Number   int
	Depth    int
	Bounds   image.Rectangle
	Attrs    map[string]string
	Children []*AndroidUINode
}

type AndroidNodeSnapshot struct {
	Device     string
	BaseDevice string
	CapturedAt time.Time
	Nodes      []*AndroidUINode
}

type androidNodePageState struct {
	snapshot      *AndroidNodeSnapshot
	viewer        *ImageViewer
	searchText    string
	filteredNodes []*AndroidUINode
	treeChildren  map[string][]string
	treeNodeByID  map[string]*AndroidUINode
	treeIDByNode  map[*AndroidUINode]string
	treeParentID  map[string]string
}

func newAndroidNodePageState(snapshot *AndroidNodeSnapshot, viewer *ImageViewer) *androidNodePageState {
	state := &androidNodePageState{
		snapshot:      snapshot,
		viewer:        viewer,
		filteredNodes: make([]*AndroidUINode, 0),
	}
	state.ensureCollections()
	return state
}

func (s *androidNodePageState) ensureCollections() {
	if s.filteredNodes == nil {
		s.filteredNodes = make([]*AndroidUINode, 0)
	}
	if s.treeChildren == nil {
		s.treeChildren = make(map[string][]string)
	}
	if s.treeNodeByID == nil {
		s.treeNodeByID = make(map[string]*AndroidUINode)
	}
	if s.treeIDByNode == nil {
		s.treeIDByNode = make(map[*AndroidUINode]string)
	}
	if s.treeParentID == nil {
		s.treeParentID = make(map[string]string)
	}
}

type androidNodeAttrRow struct {
	Selected bool
	Name     string
	Value    string

	Finder  string
	XMLAttr string
	Kind    androidNodeAttrKind
	Method  string
	Methods []string
}

type androidNodeAttrKind int

const (
	androidNodeAttrUnsupported androidNodeAttrKind = iota
	androidNodeAttrString
	androidNodeAttrBool
	androidNodeAttrInt
	androidNodeAttrBounds
)

type androidNodeAttrMeta struct {
	Name    string
	XMLAttr string
	Finder  string
	Kind    androidNodeAttrKind
	Methods []string
}

var androidNodeAttrMetas = []androidNodeAttrMeta{
	{Name: "depth", Finder: "depth", Kind: androidNodeAttrInt, Methods: []string{"Depth"}},
	{Name: "index", XMLAttr: "index", Finder: "index", Kind: androidNodeAttrInt, Methods: []string{"Index"}},
	{Name: "drawingOrder", XMLAttr: "drawing-order", Finder: "drawing-order", Kind: androidNodeAttrInt, Methods: []string{"DrawingOrder"}},
	{Name: "class", XMLAttr: "class", Finder: "class", Kind: androidNodeAttrString, Methods: []string{"ClassName", "ClassNameContains", "ClassNameStartsWith", "ClassNameEndsWith", "ClassNameMatches"}},
	{Name: "package", XMLAttr: "package", Finder: "package", Kind: androidNodeAttrString, Methods: []string{"PackageName", "PackageNameContains", "PackageNameStartsWith", "PackageNameEndsWith", "PackageNameMatches"}},
	{Name: "text", XMLAttr: "text", Finder: "text", Kind: androidNodeAttrString, Methods: []string{"Text", "TextContains", "TextStartsWith", "TextEndsWith", "TextMatches"}},
	{Name: "desc", XMLAttr: "content-desc", Finder: "desc", Kind: androidNodeAttrString, Methods: []string{"Desc", "DescContains", "DescStartsWith", "DescEndsWith", "DescMatches"}},
	{Name: "id", XMLAttr: "resource-id", Finder: "id", Kind: androidNodeAttrString, Methods: []string{"Id", "IdContains", "IdStartsWith", "IdEndsWith", "IdMatches"}},
	{Name: "bounds", XMLAttr: "bounds", Kind: androidNodeAttrBounds, Methods: []string{"Bounds", "BoundsInside", "BoundsContains"}},
	{Name: "checkable", XMLAttr: "checkable", Finder: "checkable", Kind: androidNodeAttrBool, Methods: []string{"Checkable"}},
	{Name: "checked", XMLAttr: "checked", Finder: "checked", Kind: androidNodeAttrBool, Methods: []string{"Checked"}},
	{Name: "clickable", XMLAttr: "clickable", Finder: "clickable", Kind: androidNodeAttrBool, Methods: []string{"Clickable"}},
	{Name: "enabled", XMLAttr: "enabled", Finder: "enabled", Kind: androidNodeAttrBool, Methods: []string{"Enabled"}},
	{Name: "focusable", XMLAttr: "focusable", Finder: "focusable", Kind: androidNodeAttrBool, Methods: []string{"Focusable"}},
	{Name: "focused", XMLAttr: "focused", Finder: "focused", Kind: androidNodeAttrBool, Methods: []string{"Focused"}},
	{Name: "scrollable", XMLAttr: "scrollable", Finder: "scrollable", Kind: androidNodeAttrBool, Methods: []string{"Scrollable"}},
	{Name: "editable", XMLAttr: "editable", Finder: "editable", Kind: androidNodeAttrBool, Methods: []string{"Editable"}},
	{Name: "longClickable", XMLAttr: "long-clickable", Finder: "long-clickable", Kind: androidNodeAttrBool, Methods: []string{"LongClickable"}},
	{Name: "password", XMLAttr: "password", Finder: "password", Kind: androidNodeAttrBool, Methods: []string{"Password"}},
	{Name: "selected", XMLAttr: "selected", Finder: "selected", Kind: androidNodeAttrBool, Methods: []string{"Selected"}},
	{Name: "visible", XMLAttr: "visible", Finder: "visible", Kind: androidNodeAttrBool, Methods: []string{"Visible"}},
	{Name: "multiLine", XMLAttr: "multi-line", Finder: "multi-line", Kind: androidNodeAttrBool, Methods: []string{"MultiLine"}},
	{Name: "dismissable", XMLAttr: "dismissable", Finder: "dismissable", Kind: androidNodeAttrBool, Methods: []string{"Dismissable"}},
	{Name: "contextClickable", XMLAttr: "context-clickable", Finder: "context-clickable", Kind: androidNodeAttrBool, Methods: []string{"ContextClickable"}},
}

func newCompactNodeToolText(value string) *widget.Label {
	text := widget.NewLabel(value)
	text.SizeName = theme.SizeNameCaptionText
	text.Truncation = fyne.TextTruncateClip
	text.Wrapping = fyne.TextWrapOff
	return text
}

func newNodeTreeToolText() *widget.Label {
	text := widget.NewLabel("")
	text.SizeName = theme.SizeNameCaptionText
	text.Wrapping = fyne.TextWrapOff
	return text
}

func setCompactNodeToolText(text *widget.Label, value string) {
	text.SetText(value)
}

func newCompactNodeToolAttrRow(selected, name, value, finder fyne.CanvasObject) *fyne.Container {
	selectedBox := container.New(&fixedContentWidthLayout{width: compactNodeToolAttrSelectWidth}, selected)
	nameBox := container.New(&fixedContentWidthLayout{width: compactNodeToolAttrNameWidth}, name)
	finderBox := container.New(&fixedContentWidthLayout{width: compactNodeToolAttrFinderWidth}, finder)
	main := container.NewBorder(nil, nil, nameBox, nil, value)
	return container.NewBorder(nil, nil, selectedBox, finderBox, main)
}

func newCompactNodeToolAttrSeparator() fyne.CanvasObject {
	separator := canvas.NewRectangle(color.NRGBA{R: 210, G: 210, B: 210, A: 120})
	separator.SetMinSize(fyne.NewSize(1, 1))
	return separator
}

type verticalOffsetLayout struct {
	offsetY float32
}

func (l *verticalOffsetLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(1, 1)
	}
	return objects[0].MinSize()
}

func (l *verticalOffsetLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	height := size.Height - l.offsetY
	if height < 1 {
		height = 1
	}
	objects[0].Move(fyne.NewPos(0, l.offsetY))
	objects[0].Resize(fyne.NewSize(size.Width, height))
}

type minContentWidthLayout struct {
	minWidth float32
}

func (l *minContentWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(l.minWidth, 1)
	}
	min := objects[0].MinSize()
	min.Width = fyne.Max(min.Width, l.minWidth)
	return min
}

func (l *minContentWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
}

type flexibleMinWidthLayout struct {
	minWidth float32
}

func (l *flexibleMinWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(l.minWidth, 1)
	}
	min := objects[0].MinSize()
	min.Width = l.minWidth
	return min
}

func (l *flexibleMinWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
}

type compactNodeToolRow struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	height   float32
	minWidth *minContentWidthLayout
	onTapped func()
}

func newCompactNodeToolRow(content fyne.CanvasObject, height float32) *compactNodeToolRow {
	row := &compactNodeToolRow{
		content: content,
		height:  height,
	}
	row.ExtendBaseWidget(row)
	return row
}

func (r *compactNodeToolRow) Tapped(*fyne.PointEvent) {
	if r.onTapped != nil {
		r.onTapped()
	}
}

func (r *compactNodeToolRow) CreateRenderer() fyne.WidgetRenderer {
	return &compactNodeToolRowRenderer{
		row:     r,
		objects: []fyne.CanvasObject{r.content},
	}
}

type compactNodeToolRowRenderer struct {
	row     *compactNodeToolRow
	objects []fyne.CanvasObject
}

func (r *compactNodeToolRowRenderer) Destroy() {}

func (r *compactNodeToolRowRenderer) Layout(size fyne.Size) {
	contentSize := r.row.content.MinSize()
	if contentSize.Height > size.Height {
		contentSize.Height = size.Height
	}
	r.row.content.Move(fyne.NewPos(0, (size.Height-contentSize.Height)/2))
	r.row.content.Resize(fyne.NewSize(size.Width, contentSize.Height))
}

func (r *compactNodeToolRowRenderer) MinSize() fyne.Size {
	min := r.row.content.MinSize()
	if r.row.minWidth != nil {
		min.Width = fyne.Max(min.Width, r.row.minWidth.minWidth)
	}
	min.Height = r.row.height
	return min
}

func (r *compactNodeToolRowRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *compactNodeToolRowRenderer) Refresh() {
	r.row.content.Refresh()
}

func newCompactNodeTreeRow(minWidth *minContentWidthLayout) *compactNodeToolRow {
	row := newCompactNodeToolRow(container.NewHBox(newNodeTreeToolText()), compactNodeToolTreeRowHeight)
	row.minWidth = minWidth
	return row
}

func compactNodeTreeRowLabel(item fyne.CanvasObject) *widget.Label {
	row, ok := item.(*compactNodeToolRow)
	if !ok {
		return nil
	}
	content, ok := row.content.(*fyne.Container)
	if !ok || len(content.Objects) == 0 {
		return nil
	}
	label, _ := content.Objects[0].(*widget.Label)
	return label
}

func compactNodeTreeRow(item fyne.CanvasObject) *compactNodeToolRow {
	row, _ := item.(*compactNodeToolRow)
	return row
}

type AndroidNodeTool struct {
	window fyne.Window

	getSelectedDevice func() string
	getImageViewer    func() *ImageViewer
	openNodeImage     func(image.Image, func(int, int)) *ImageViewer
	onOpen            func()

	root             fyne.CanvasObject
	captureBtn       *widget.Button
	searchEntry      *widget.Entry
	statusLabel      *widget.Label
	nodeTree         *widget.Tree
	nodeTreeWidth    *minContentWidthLayout
	attrList         *fyne.Container
	selectorEntry    *widget.Entry
	selectorFunc     *widget.Select
	selectorFormat   *widget.Entry
	copySelectorBtn  *widget.Button
	copyAttrsBtn     *widget.Button
	selectAllBtn     *widget.Button
	clearSelectedBtn *widget.Button
	testSelectorBtn  *widget.Button

	busy          bool
	snapshot      *AndroidNodeSnapshot
	filteredNodes []*AndroidUINode
	selectedNode  *AndroidUINode
	attrRows      []androidNodeAttrRow
	nodeViewer    *ImageViewer
	nodePages     map[*ImageViewer]*androidNodePageState
	activePage    *androidNodePageState
	treeChildren  map[string][]string
	treeNodeByID  map[string]*AndroidUINode
	treeIDByNode  map[*AndroidUINode]string
	treeParentID  map[string]string
	syncingTree   bool
}

func newAndroidNodeTool(w fyne.Window, getSelectedDevice func() string, getImageViewer func() *ImageViewer, openNodeImage func(image.Image, func(int, int)) *ImageViewer) *AndroidNodeTool {
	tool := &AndroidNodeTool{
		window:            w,
		getSelectedDevice: getSelectedDevice,
		getImageViewer:    getImageViewer,
		openNodeImage:     openNodeImage,
		filteredNodes:     make([]*AndroidUINode, 0),
		attrRows:          make([]androidNodeAttrRow, 0),
		nodePages:         make(map[*ImageViewer]*androidNodePageState),
		treeChildren:      make(map[string][]string),
		treeNodeByID:      make(map[string]*AndroidUINode),
		treeIDByNode:      make(map[*AndroidUINode]string),
		treeParentID:      make(map[string]string),
	}

	tool.captureBtn = widget.NewButton("简单节点抓取", func() {
		tool.CaptureSimple()
	})

	tool.searchEntry = widget.NewEntry()
	tool.searchEntry.SetPlaceHolder("搜索 text / desc / id / class")
	searchBtn := widget.NewButton("搜索", func() {
		tool.applySearch()
	})
	prevBtn := widget.NewButton("上一个", func() {
		tool.selectRelative(-1)
	})
	nextBtn := widget.NewButton("下一个", func() {
		tool.selectRelative(1)
	})

	tool.statusLabel = widget.NewLabel("未抓取节点")
	tool.statusLabel.Wrapping = fyne.TextWrapWord
	tool.nodeTreeWidth = &minContentWidthLayout{minWidth: 1}

	tool.nodeTree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return tool.treeChildren[uid]
		},
		func(uid widget.TreeNodeID) bool {
			return uid == "" || len(tool.treeChildren[uid]) > 0
		},
		func(bool) fyne.CanvasObject {
			return newCompactNodeTreeRow(tool.nodeTreeWidth)
		},
		func(uid widget.TreeNodeID, branch bool, item fyne.CanvasObject) {
			label := compactNodeTreeRowLabel(item)
			if label == nil {
				return
			}
			row := compactNodeTreeRow(item)
			if uid == "" {
				setCompactNodeToolText(label, "")
				if row != nil {
					row.onTapped = nil
				}
				return
			}
			node := tool.treeNodeByID[uid]
			if node == nil {
				setCompactNodeToolText(label, "")
				if row != nil {
					row.onTapped = nil
				}
				return
			}
			prefix := "  "
			if node == tool.selectedNode {
				prefix = "▶ "
			}
			setCompactNodeToolText(label, prefix+androidNodeSummary(node))
			if row != nil {
				row.onTapped = func() {
					if !tool.syncingTree {
						tool.nodeTree.Select(uid)
					}
				}
			}
		},
	)
	tool.nodeTree.HideSeparators = true
	tool.nodeTree.OnSelected = func(uid widget.TreeNodeID) {
		if tool.syncingTree {
			return
		}
		node := tool.treeNodeByID[uid]
		if node == nil {
			return
		}
		tool.selectNode(node)
	}

	tool.attrList = container.NewVBox()

	tool.selectAllBtn = widget.NewButton("全部勾选", func() {
		tool.setAllAttrRows(true)
	})
	tool.clearSelectedBtn = widget.NewButton("清除全选", func() {
		tool.setAllAttrRows(false)
	})
	tool.testSelectorBtn = widget.NewButton("查找测试", func() {
		tool.testSelectedAttrs()
	})
	generateSelectorBtn := widget.NewButton("生成代码", func() {
		code, err := tool.selectedSelectorCode()
		if err != nil {
			tool.selectorEntry.SetText("生成失败: " + err.Error())
			tool.setStatus("生成代码失败: " + err.Error())
			return
		}
		if strings.TrimSpace(code) == "" {
			tool.setStatus("请先勾选至少一个有效属性")
			return
		}
		tool.selectorEntry.SetText(code)
		tool.setStatus("已生成 uiacc 代码")
	})
	tool.copySelectorBtn = widget.NewButton("复制代码", func() {
		tool.copySelectedSelectorCode()
	})
	tool.copyAttrsBtn = widget.NewButton("复制参数", func() {
		tool.copySelectedSelectorParams()
	})

	tool.selectorFunc = widget.NewSelect([]string{
		androidNodeSelectorFuncFindOnce,
		androidNodeSelectorFuncFind,
		androidNodeSelectorFuncWaitFor,
	}, nil)
	tool.selectorFunc.SetSelected(androidNodeSelectorFuncFindOnce)
	tool.selectorFunc.OnChanged = func(string) {
		tool.refreshSelector()
	}

	tool.selectorFormat = widget.NewMultiLineEntry()
	tool.selectorFormat.SetPlaceHolder("输出格式，支持 {displayId} / {chain} / {params} / {function} / {call} / {timeout}")
	tool.selectorFormat.SetMinRowsVisible(2)
	tool.selectorFormat.SetText(androidNodeSelectorDefaultTemplate)
	tool.selectorFormat.OnChanged = func(string) {
		tool.refreshSelector()
	}

	tool.selectorEntry = widget.NewMultiLineEntry()
	tool.selectorEntry.SetPlaceHolder("选择节点并勾选属性后生成 AutoGo uiacc 代码")
	tool.selectorEntry.SetMinRowsVisible(3)

	nodeHeader := newCompactNodeToolText("节点树")
	attrHeader := newCompactNodeToolAttrRow(
		newCompactNodeToolText("选"),
		newCompactNodeToolText("属性"),
		newCompactNodeToolText("值"),
		newCompactNodeToolText("函数"),
	)
	nodeTreeArea := container.New(&flexibleMinWidthLayout{minWidth: 1}, tool.nodeTree)

	tool.root = container.NewVBox(
		container.NewBorder(nil, nil, nil, container.NewHBox(tool.captureBtn, searchBtn, prevBtn, nextBtn), tool.searchEntry),
		tool.statusLabel,
		nodeHeader,
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, nodeTreeArea), 205),
		attrHeader,
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, container.NewVScroll(tool.attrList)), 140),
		container.NewGridWithColumns(4, tool.selectAllBtn, tool.clearSelectedBtn, tool.testSelectorBtn, generateSelectorBtn),
		container.NewBorder(nil, nil, widget.NewLabel("函数"), nil, tool.selectorFunc),
		container.NewBorder(nil, nil, widget.NewLabel("格式"), tool.copyAttrsBtn, tool.selectorFormat),
		container.NewBorder(nil, nil, widget.NewLabel("代码"), tool.copySelectorBtn, tool.selectorEntry),
	)

	return tool
}

func (t *AndroidNodeTool) Content() fyne.CanvasObject {
	return t.root
}

func (t *AndroidNodeTool) SetOnOpen(onOpen func()) {
	t.onOpen = onOpen
}

func (t *AndroidNodeTool) Capture() {
	t.capture(false)
}

func (t *AndroidNodeTool) CaptureSimple() {
	t.capture(true)
}

func (t *AndroidNodeTool) capture(compressed bool) {
	if t.busy {
		return
	}
	if t.onOpen != nil {
		t.onOpen()
	}

	device := strings.TrimSpace(t.getSelectedDevice())
	if device == "" {
		dialog.ShowInformation("提示", "请先选择已连接的 Android 设备", t.window)
		return
	}

	t.setBusy(true)
	t.setStatus("正在截图并抓取节点...")

	go func() {
		capturedImg, captureErr := captureScreenWithADB(device)
		if captureErr != nil {
			fyne.Do(func() {
				t.setBusy(false)
				t.setStatus("抓取节点截图失败")
				dialog.ShowError(fmt.Errorf("抓取节点截图失败: %v", captureErr), t.window)
			})
			return
		}
		capturedImg = convertToNRGBA(capturedImg)

		snapshot, err := captureAndroidNodeSnapshot(device, compressed)
		fyne.Do(func() {
			t.setBusy(false)
			if err != nil {
				t.setStatus("抓取节点失败")
				dialog.ShowError(fmt.Errorf("抓取节点失败: %v", err), t.window)
				return
			}

			var page *androidNodePageState
			var viewer *ImageViewer
			if t.openNodeImage != nil {
				viewer = t.openNodeImage(capturedImg, func(x, y int) {
					if page == nil {
						return
					}
					t.selectNodeAtPointOnPage(page, x, y)
				})
			}
			page = newAndroidNodePageState(snapshot, viewer)
			if viewer != nil {
				if t.nodePages == nil {
					t.nodePages = make(map[*ImageViewer]*androidNodePageState)
				}
				t.nodePages[viewer] = page
				viewer.onDoubleClick = func(x, y int) {
					t.selectSelectedNodeParentAtPointOnPage(page, x, y)
				}
				viewer.onActivated = func() {
					t.ActivateViewer(viewer)
				}
			}
			t.activateNodePage(page)
			t.setSnapshot(snapshot)
		})
	}()
}

func (t *AndroidNodeTool) setBusy(busy bool) {
	t.busy = busy
	if t.captureBtn == nil {
		return
	}
	if busy {
		t.captureBtn.Disable()
		return
	}
	t.captureBtn.Enable()
}

func (t *AndroidNodeTool) setStatus(text string) {
	if t.statusLabel != nil {
		t.statusLabel.SetText(text)
	}
}

func (t *AndroidNodeTool) ActivateViewer(viewer *ImageViewer) {
	if viewer == nil || t.nodePages == nil {
		return
	}

	page := t.nodePages[viewer]
	if page == nil || page == t.activePage {
		return
	}

	t.activateNodePage(page)
	if t.snapshot != nil {
		t.setStatus(fmt.Sprintf("已切换节点页 · %d 个节点 · 设备: %s", len(t.snapshot.Nodes), t.snapshot.Device))
	}
}

func (t *AndroidNodeTool) activateNodePage(page *androidNodePageState) {
	if page == nil {
		return
	}
	if t.activePage != page {
		t.saveActivePageState()
	}

	page.ensureCollections()
	t.activePage = page
	t.snapshot = page.snapshot
	t.filteredNodes = page.filteredNodes
	t.nodeViewer = page.viewer
	t.treeChildren = page.treeChildren
	t.treeNodeByID = page.treeNodeByID
	t.treeIDByNode = page.treeIDByNode
	t.treeParentID = page.treeParentID

	if t.searchEntry != nil && t.searchEntry.Text != page.searchText {
		t.searchEntry.SetText(page.searchText)
	}
	t.refreshNodeTree()
	t.refreshActivePageSelectionVisual()
}

func (t *AndroidNodeTool) refreshNodeTree() {
	if t.nodeTree != nil {
		t.nodeTree.Refresh()
	}
}

func (t *AndroidNodeTool) refreshActivePageSelectionVisual() {
	if t.selectedNodeBelongsToActivePage() {
		t.syncNodeListSelection(t.selectedNode)
		t.highlightSelectedNode()
		return
	}
	t.unselectNodeTree()
	t.clearSelectedNodeOverlay()
}

func (t *AndroidNodeTool) selectedNodeBelongsToActivePage() bool {
	if t.snapshot == nil || t.selectedNode == nil {
		return false
	}
	for _, node := range t.snapshot.Nodes {
		if node == t.selectedNode {
			return true
		}
	}
	return false
}

func (t *AndroidNodeTool) unselectNodeTree() {
	if t.nodeTree == nil {
		return
	}
	t.syncingTree = true
	t.nodeTree.UnselectAll()
	t.syncingTree = false
}

func (t *AndroidNodeTool) saveActivePageState() {
	if t.activePage == nil {
		return
	}

	t.activePage.snapshot = t.snapshot
	t.activePage.viewer = t.nodeViewer
	if t.searchEntry != nil {
		t.activePage.searchText = t.searchEntry.Text
	}
	t.activePage.filteredNodes = t.filteredNodes
	t.activePage.treeChildren = t.treeChildren
	t.activePage.treeNodeByID = t.treeNodeByID
	t.activePage.treeIDByNode = t.treeIDByNode
	t.activePage.treeParentID = t.treeParentID
}

func (t *AndroidNodeTool) setSnapshot(snapshot *AndroidNodeSnapshot) {
	selectFirstNode := !t.hasSelectorState()
	t.snapshot = snapshot
	if selectFirstNode {
		t.selectedNode = nil
		t.attrRows = t.attrRows[:0]
	}
	t.updateNodeOverlay()
	t.applySearchWithSelection(selectFirstNode)
	t.saveActivePageState()

	t.setStatus(fmt.Sprintf("已抓取 %d 个节点 · 设备: %s", len(snapshot.Nodes), snapshot.Device))
}

func (t *AndroidNodeTool) hasSelectorState() bool {
	return t.selectedNode != nil || len(t.attrRows) > 0
}

func (t *AndroidNodeTool) applySearch() {
	t.applySearchWithSelection(true)
}

func (t *AndroidNodeTool) applySearchWithSelection(selectFirstNode bool) {
	t.filteredNodes = t.filteredNodes[:0]
	if t.snapshot == nil {
		t.rebuildNodeTree("")
		if selectFirstNode {
			t.refreshLists()
		} else {
			t.refreshNodeTree()
		}
		t.saveActivePageState()
		return
	}

	keyword := strings.ToLower(strings.TrimSpace(t.searchEntry.Text))
	for _, node := range t.snapshot.Nodes {
		if keyword == "" || androidNodeMatches(node, keyword) {
			t.filteredNodes = append(t.filteredNodes, node)
		}
	}
	t.rebuildNodeTree(keyword)

	if len(t.filteredNodes) > 0 {
		if selectFirstNode {
			t.selectNode(t.filteredNodes[0])
		} else {
			t.refreshActivePageSelectionVisual()
			t.saveActivePageState()
		}
		return
	}

	if selectFirstNode {
		t.selectedNode = nil
		t.attrRows = t.attrRows[:0]
		t.refreshSelector()
		t.refreshLists()
	} else {
		t.refreshNodeTree()
		t.unselectNodeTree()
	}
	t.clearSelectedNodeOverlay()
	t.setStatus(fmt.Sprintf("没有匹配节点 · 总节点: %d", len(t.snapshot.Nodes)))
	t.saveActivePageState()
}

func (t *AndroidNodeTool) selectRelative(offset int) {
	if len(t.filteredNodes) == 0 {
		return
	}

	current := -1
	for i, node := range t.filteredNodes {
		if node == t.selectedNode {
			current = i
			break
		}
	}
	if current < 0 {
		current = 0
	} else {
		current = (current + offset + len(t.filteredNodes)) % len(t.filteredNodes)
	}
	t.selectNode(t.filteredNodes[current])
}

func (t *AndroidNodeTool) selectNode(node *AndroidUINode) {
	if node == nil {
		return
	}
	t.clearNodeFindTestHighlights()
	t.selectedNode = node
	t.attrRows = buildAndroidNodeAttrRows(node)
	t.refreshSelector()
	t.syncNodeListSelection(node)
	t.refreshLists()
	t.highlightSelectedNode()
	t.saveActivePageState()
}

func (t *AndroidNodeTool) refreshLists() {
	if t.nodeTree != nil {
		t.nodeTree.Refresh()
	}
	t.rebuildAttrList()
}

func (t *AndroidNodeTool) rebuildAttrList() {
	if t.attrList == nil {
		return
	}

	t.attrList.RemoveAll()
	for i, attr := range t.attrRows {
		index := i
		t.attrList.Add(t.newAndroidNodeAttrRowContent(index, attr))
	}
	t.attrList.Refresh()
}

func (t *AndroidNodeTool) newAndroidNodeAttrRowContent(index int, attr androidNodeAttrRow) fyne.CanvasObject {
	selectable := androidNodeAttrSelectable(attr)
	check := widget.NewCheck("", nil)
	check.SetChecked(selectable && attr.Selected)
	if !selectable {
		check.Disable()
	} else {
		check.OnChanged = func(selected bool) {
			if index < 0 || index >= len(t.attrRows) {
				return
			}
			t.attrRows[index].Selected = selected
			t.refreshSelector()
			t.saveActivePageState()
		}
	}

	methodSelect := widget.NewSelect(attr.Methods, nil)
	if selectable && len(attr.Methods) > 0 {
		methodSelect.SetSelected(attr.Method)
		methodSelect.OnChanged = func(method string) {
			if index < 0 || index >= len(t.attrRows) || method == "" {
				return
			}
			t.attrRows[index].Method = method
			t.refreshSelector()
			t.saveActivePageState()
		}
	} else {
		methodSelect.Disable()
	}

	checkBox := container.New(&verticalOffsetLayout{offsetY: compactNodeToolCheckOffsetY}, check)
	row := newCompactNodeToolAttrRow(
		checkBox,
		newCompactNodeToolText(attr.Name),
		newCompactNodeToolText(trimMiddle(attr.Value, 28)),
		methodSelect,
	)
	rowWithSeparator := container.NewBorder(nil, newCompactNodeToolAttrSeparator(), nil, nil, row)
	rowWidget := newCompactNodeToolRow(rowWithSeparator, compactNodeToolAttrRowHeight)
	if selectable {
		rowWidget.onTapped = func() {
			t.toggleAttrRowSelection(index)
		}
	}
	return rowWidget
}

func androidNodeAttrSelectable(attr androidNodeAttrRow) bool {
	method := androidNodeAttrMethod(attr)
	if method == "" {
		return attr.Value != "" && androidNodeFinderToXMLAttr(attr.Finder) != ""
	}
	return androidNodeAttrKindForMethod(attr, method) != androidNodeAttrUnsupported
}

func (t *AndroidNodeTool) toggleAttrRowSelection(index int) {
	if index < 0 || index >= len(t.attrRows) {
		return
	}
	if !androidNodeAttrSelectable(t.attrRows[index]) {
		return
	}
	t.attrRows[index].Selected = !t.attrRows[index].Selected
	t.refreshSelector()
	t.rebuildAttrList()
	t.saveActivePageState()
}

func (t *AndroidNodeTool) highlightSelectedNode() {
	if t.selectedNode == nil {
		return
	}
	viewer := t.activeNodeViewer()
	if viewer == nil || viewer.image == nil {
		return
	}
	viewer.SetNodeHighlightRect(t.selectedNode.Bounds)
}

func (t *AndroidNodeTool) clearSelectedNodeOverlay() {
	viewer := t.activeNodeViewer()
	if viewer == nil || viewer.image == nil {
		return
	}
	viewer.SetNodeHighlightRect(image.Rectangle{})
}

func (t *AndroidNodeTool) activeNodeViewer() *ImageViewer {
	if t.nodeViewer != nil {
		return t.nodeViewer
	}
	if t.getImageViewer == nil {
		return nil
	}
	return t.getImageViewer()
}

func (t *AndroidNodeTool) updateNodeOverlay() {
	viewer := t.activeNodeViewer()
	if viewer == nil || viewer.image == nil || t.snapshot == nil {
		return
	}

	rects := make([]image.Rectangle, 0, len(t.snapshot.Nodes))
	for _, node := range t.snapshot.Nodes {
		if node == nil || node.Bounds.Empty() {
			continue
		}
		rects = append(rects, node.Bounds)
	}
	viewer.SetNodeOverlayRects(rects)
}

func (t *AndroidNodeTool) syncNodeListSelection(node *AndroidUINode) {
	if t.nodeTree == nil || node == nil {
		return
	}

	uid := t.treeIDByNode[node]
	if uid == "" {
		return
	}

	for parentID := t.treeParentID[uid]; parentID != ""; parentID = t.treeParentID[parentID] {
		t.nodeTree.OpenBranch(parentID)
	}

	t.syncingTree = true
	t.nodeTree.Select(uid)
	t.syncingTree = false
}

func (t *AndroidNodeTool) filteredNodeIndex(node *AndroidUINode) int {
	for i, candidate := range t.filteredNodes {
		if candidate == node {
			return i
		}
	}
	return -1
}

func (t *AndroidNodeTool) rebuildNodeTree(keyword string) {
	t.treeChildren = make(map[string][]string)
	t.treeNodeByID = make(map[string]*AndroidUINode)
	t.treeIDByNode = make(map[*AndroidUINode]string)
	t.treeParentID = make(map[string]string)

	if t.snapshot == nil {
		t.updateNodeTreeScrollWidth()
		return
	}

	visible := make(map[*AndroidUINode]bool)
	if keyword == "" {
		for _, node := range t.snapshot.Nodes {
			visible[node] = true
		}
	} else {
		parentByNode := androidNodeParentMap(t.snapshot.Nodes)
		for _, node := range t.filteredNodes {
			for current := node; current != nil; current = parentByNode[current] {
				visible[current] = true
			}
		}
	}

	var addNode func(parentID string, node *AndroidUINode)
	addNode = func(parentID string, node *AndroidUINode) {
		if node == nil || !visible[node] {
			return
		}

		uid := androidNodeTreeID(node)
		t.treeChildren[parentID] = append(t.treeChildren[parentID], uid)
		t.treeNodeByID[uid] = node
		t.treeIDByNode[node] = uid
		t.treeParentID[uid] = parentID

		for _, child := range node.Children {
			addNode(uid, child)
		}
	}

	for _, root := range androidRootNodes(t.snapshot.Nodes) {
		addNode("", root)
	}
	for i := 0; i < androidNodeTreeBottomSpacerRows; i++ {
		t.treeChildren[""] = append(t.treeChildren[""], androidNodeTreeBottomSpacerID(i))
	}

	t.updateNodeTreeScrollWidth()
	if t.nodeTree != nil {
		t.nodeTree.Refresh()
		t.nodeTree.OpenAllBranches()
	}
}

func (t *AndroidNodeTool) updateNodeTreeScrollWidth() {
	if t.nodeTreeWidth == nil {
		return
	}

	minWidth := float32(1)
	for _, node := range t.treeNodeByID {
		textWidth := fyne.MeasureText("  "+androidNodeSummary(node), theme.CaptionTextSize(), fyne.TextStyle{}).Width
		indentWidth := float32(node.Depth+1) * (theme.IconInlineSize() + theme.Padding())
		minWidth = fyne.Max(minWidth, textWidth+indentWidth+theme.Padding()*4)
	}

	t.nodeTreeWidth.minWidth = minWidth
}

func (t *AndroidNodeTool) selectNodeAtPoint(x, y int) {
	node := t.smallestNodeAtPoint(image.Pt(x, y))
	if node == nil {
		return
	}

	if t.filteredNodeIndex(node) < 0 && t.snapshot != nil {
		t.searchEntry.SetText("")
		t.applySearch()
	}

	t.selectNode(node)
	t.setStatus(fmt.Sprintf("已选择节点 %03d · %s", node.Number, androidNodeSummary(node)))
}

func (t *AndroidNodeTool) selectNodeAtPointOnPage(page *androidNodePageState, x, y int) {
	if page == nil {
		return
	}
	t.activateNodePage(page)
	t.selectNodeAtPoint(x, y)
}

func (t *AndroidNodeTool) selectSelectedNodeParentAtPointOnPage(page *androidNodePageState, x, y int) {
	if page == nil {
		return
	}
	t.activateNodePage(page)
	t.selectSelectedNodeParentAtPoint(image.Pt(x, y))
}

func (t *AndroidNodeTool) selectSelectedNodeParentAtPoint(point image.Point) {
	if t.snapshot == nil || t.selectedNode == nil || t.selectedNode.Bounds.Empty() || !point.In(t.selectedNode.Bounds) {
		return
	}

	parent := t.parentNode(t.selectedNode)
	if parent == nil {
		t.setStatus(fmt.Sprintf("当前节点 %03d 已经是顶层节点", t.selectedNode.Number))
		return
	}

	if t.filteredNodeIndex(parent) < 0 {
		if t.searchEntry != nil {
			t.searchEntry.SetText("")
		}
		t.applySearchWithSelection(false)
	}

	t.selectNode(parent)
	t.setStatus(fmt.Sprintf("已选择父节点 %03d · %s", parent.Number, androidNodeSummary(parent)))
}

func (t *AndroidNodeTool) parentNode(node *AndroidUINode) *AndroidUINode {
	if t.snapshot == nil || node == nil {
		return nil
	}
	return androidNodeParentMap(t.snapshot.Nodes)[node]
}

func (t *AndroidNodeTool) smallestNodeAtPoint(point image.Point) *AndroidUINode {
	if t.snapshot == nil {
		return nil
	}

	var best *AndroidUINode
	bestArea := 0
	for _, node := range t.snapshot.Nodes {
		if node == nil || node.Bounds.Empty() || !point.In(node.Bounds) {
			continue
		}
		area := node.Bounds.Dx() * node.Bounds.Dy()
		if area <= 0 {
			continue
		}
		if best == nil || area < bestArea || (area == bestArea && node.Depth > best.Depth) {
			best = node
			bestArea = area
		}
	}
	return best
}

func (t *AndroidNodeTool) setAllAttrRows(selected bool) {
	for i := range t.attrRows {
		t.attrRows[i].Selected = selected && androidNodeAttrSelectable(t.attrRows[i])
	}
	t.refreshSelector()
	t.rebuildAttrList()
	t.saveActivePageState()
}

func (t *AndroidNodeTool) refreshSelector() {
	if t.selectorEntry == nil {
		return
	}
	code, err := t.selectedSelectorCode()
	if err != nil {
		t.selectorEntry.SetText("生成失败: " + err.Error())
		return
	}
	t.selectorEntry.SetText(code)
}

func (t *AndroidNodeTool) selectedSelectorCode() (string, error) {
	selected := selectedAndroidNodeAttrRows(t.attrRows)
	if len(selected) == 0 {
		return "", nil
	}
	return buildAndroidNodeSelectorCode(selected, androidNodeSelectorOptions{
		DisplayID: "0",
		Function:  t.selectedSelectorFunction(),
		Template:  t.selectedSelectorTemplate(),
		Timeout:   androidNodeSelectorDefaultTimeout,
	})
}

func (t *AndroidNodeTool) selectedSelectorFunction() string {
	if t.selectorFunc == nil || strings.TrimSpace(t.selectorFunc.Selected) == "" {
		return androidNodeSelectorFuncFindOnce
	}
	return strings.TrimSpace(t.selectorFunc.Selected)
}

func (t *AndroidNodeTool) selectedSelectorTemplate() string {
	if t.selectorFormat == nil {
		return androidNodeSelectorDefaultTemplate
	}
	return t.selectorFormat.Text
}

func (t *AndroidNodeTool) copySelectedSelectorCode() {
	if t.window == nil {
		return
	}
	code, err := t.selectedSelectorCode()
	if err != nil {
		t.setStatus("复制代码失败: " + err.Error())
		return
	}
	if strings.TrimSpace(code) == "" {
		t.setStatus("请先勾选至少一个有效属性")
		return
	}
	t.window.Clipboard().SetContent(code)
	t.setStatus("已复制 uiacc 代码")
}

func (t *AndroidNodeTool) copySelectedSelectorParams() {
	if t.window == nil {
		return
	}
	chain, err := buildAndroidNodeSelectorChain(selectedAndroidNodeAttrRows(t.attrRows))
	if err != nil {
		t.setStatus("复制参数失败: " + err.Error())
		return
	}
	if strings.TrimSpace(chain) == "" {
		t.setStatus("请先勾选至少一个有效属性")
		return
	}
	t.window.Clipboard().SetContent(chain)
	t.setStatus("已复制 uiacc 参数")
}

func (t *AndroidNodeTool) selectedXPath() string {
	if len(t.attrRows) == 0 {
		return ""
	}

	var parts []string
	for _, row := range t.attrRows {
		if !row.Selected || !androidNodeAttrSelectable(row) {
			continue
		}
		xmlAttr := row.XMLAttr
		if xmlAttr == "" {
			xmlAttr = androidNodeFinderToXMLAttr(row.Finder)
		}
		if xmlAttr == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("[@%s=%s]", xmlAttr, xpathLiteral(row.Value)))
	}
	if len(parts) == 0 {
		return ""
	}
	return "//*" + strings.Join(parts, "")
}

func (t *AndroidNodeTool) testSelectedAttrs() {
	if t.snapshot == nil {
		t.clearNodeFindTestHighlights()
		t.setStatus("请先抓取节点")
		return
	}

	selected := selectedAndroidNodeAttrRows(t.attrRows)
	if len(selected) == 0 {
		t.clearNodeFindTestHighlights()
		t.setStatus("请先勾选至少一个有效属性")
		return
	}

	function := t.selectedSelectorFunction()
	highlighted, total, err := androidNodeFindTestMatches(t.snapshot.Nodes, selected, function)
	if err != nil {
		t.clearNodeFindTestHighlights()
		t.setStatus("查找测试失败: " + err.Error())
		return
	}

	t.setNodeFindTestHighlights(highlighted)
	t.setStatus(androidNodeFindTestStatus(function, total, highlighted))
}

func (t *AndroidNodeTool) setNodeFindTestHighlights(nodes []*AndroidUINode) {
	viewer := t.activeNodeViewer()
	if viewer == nil {
		return
	}

	rects := make([]image.Rectangle, 0, len(nodes))
	for _, node := range nodes {
		if node == nil || node.Bounds.Empty() {
			continue
		}
		rects = append(rects, node.Bounds)
	}
	viewer.SetNodeFindTestHighlights(rects)
}

func (t *AndroidNodeTool) clearNodeFindTestHighlights() {
	viewer := t.activeNodeViewer()
	if viewer != nil {
		viewer.ClearFindTestHighlights()
	}
}

func androidNodeFindTestMatches(nodes []*AndroidUINode, selected []androidNodeAttrRow, function string) ([]*AndroidUINode, int, error) {
	highlightAll := normalizeAndroidNodeSelectorFunction(function) == androidNodeSelectorFuncFind
	highlighted := make([]*AndroidUINode, 0)
	total := 0
	for _, node := range nodes {
		if node == nil {
			continue
		}
		matched, err := androidNodeMatchesAttrs(node, selected)
		if err != nil {
			return nil, 0, err
		}
		if !matched {
			continue
		}
		total++
		if highlightAll || len(highlighted) == 0 {
			highlighted = append(highlighted, node)
		}
	}
	return highlighted, total, nil
}

func androidNodeFindTestStatus(function string, total int, highlighted []*AndroidUINode) string {
	switch normalizeAndroidNodeSelectorFunction(function) {
	case androidNodeSelectorFuncFind:
		return fmt.Sprintf("查找测试 Find: 匹配 %d 个节点", total)
	case androidNodeSelectorFuncWaitFor:
		if len(highlighted) == 0 {
			return "查找测试 WaitFor: 当前快照未找到节点"
		}
		return fmt.Sprintf("查找测试 WaitFor: 当前快照命中节点 %03d · 候选 %d 个", highlighted[0].Number, total)
	default:
		if len(highlighted) == 0 {
			return "查找测试 FindOnce: 未找到节点"
		}
		return fmt.Sprintf("查找测试 FindOnce: 命中节点 %03d · 候选 %d 个", highlighted[0].Number, total)
	}
}

func captureAndroidNodeSnapshot(deviceID string, compressed bool) (*AndroidNodeSnapshot, error) {
	baseDevice, virtualDisplayID := splitAndroidDeviceID(deviceID)
	if baseDevice == "" {
		return nil, fmt.Errorf("设备 ID 为空")
	}

	args := []string{"-s", baseDevice, "shell", "uiautomator", "dump"}
	if compressed {
		args = append(args, "--compressed")
	}
	args = append(args, androidNodeDumpPath)

	if output, err := adbExecCombined(args...); err != nil {
		return nil, fmt.Errorf("执行 uiautomator dump 失败: %v", adbErrorWithOutput(err, output))
	}

	xmlText, err := adbExecCombined("-s", baseDevice, "exec-out", "cat", androidNodeDumpPath)
	if err != nil || !strings.Contains(xmlText, "<hierarchy") {
		fallbackText, fallbackErr := adbExecCombined("-s", baseDevice, "shell", "cat", androidNodeDumpPath)
		if fallbackErr != nil {
			return nil, fmt.Errorf("读取节点 XML 失败: %v", adbErrorWithOutput(fallbackErr, fallbackText))
		}
		xmlText = fallbackText
	}
	_, _ = adbExecCombined("-s", baseDevice, "shell", "rm", androidNodeDumpPath)

	nodes, err := parseAndroidNodeXML(xmlText)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("未解析到节点")
	}

	displayDevice := deviceID
	if virtualDisplayID != "" {
		displayDevice = fmt.Sprintf("%s[%s] · 节点来自基础设备", baseDevice, virtualDisplayID)
	}

	return &AndroidNodeSnapshot{
		Device:     displayDevice,
		BaseDevice: baseDevice,
		CapturedAt: time.Now(),
		Nodes:      nodes,
	}, nil
}

func parseAndroidNodeXML(xmlText string) ([]*AndroidUINode, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlText))
	nodes := make([]*AndroidUINode, 0)
	stack := make([]*AndroidUINode, 0)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("解析节点 XML 失败: %v", err)
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if elem.Name.Local != "node" {
				continue
			}
			node := &AndroidUINode{
				Number: len(nodes) + 1,
				Depth:  len(stack),
				Attrs:  make(map[string]string),
			}
			for _, attr := range elem.Attr {
				node.Attrs[attr.Name.Local] = attr.Value
			}
			node.Bounds = parseAndroidBounds(node.Attrs["bounds"])

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
			nodes = append(nodes, node)

		case xml.EndElement:
			if elem.Name.Local == "node" && len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	return nodes, nil
}

func parseAndroidBounds(value string) image.Rectangle {
	cleaned := strings.NewReplacer("[", " ", "]", " ", ",", " ").Replace(value)
	fields := strings.Fields(cleaned)
	if len(fields) != 4 {
		return image.Rectangle{}
	}

	nums := make([]int, 4)
	for i, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			return image.Rectangle{}
		}
		nums[i] = n
	}
	return image.Rect(nums[0], nums[1], nums[2], nums[3])
}

func splitAndroidDeviceID(deviceID string) (baseDeviceID string, virtualDisplayID string) {
	baseDeviceID = strings.TrimSpace(deviceID)
	if idx := strings.Index(baseDeviceID, "["); idx != -1 {
		endIdx := strings.Index(baseDeviceID[idx:], "]")
		if endIdx > 0 {
			virtualDisplayID = baseDeviceID[idx+1 : idx+endIdx]
			baseDeviceID = baseDeviceID[:idx]
		}
	}
	return strings.TrimSpace(baseDeviceID), strings.TrimSpace(virtualDisplayID)
}

func buildAndroidNodeAttrRows(node *AndroidUINode) []androidNodeAttrRow {
	if node == nil {
		return nil
	}

	rows := make([]androidNodeAttrRow, 0, len(androidNodeAttrMetas))
	hasStableSelection := false
	for _, meta := range androidNodeAttrMetas {
		value := androidNodeAttrValueByMeta(node, meta)
		methods := append([]string(nil), meta.Methods...)
		method := ""
		if len(methods) > 0 {
			method = methods[0]
		}
		selected := value != "" && (meta.Name == "id" || meta.Name == "desc" || meta.Name == "text") && method != ""
		if selected {
			hasStableSelection = true
		}
		rows = append(rows, androidNodeAttrRow{
			Selected: selected,
			Name:     meta.Name,
			Value:    value,
			Finder:   meta.Finder,
			XMLAttr:  meta.XMLAttr,
			Kind:     meta.Kind,
			Method:   method,
			Methods:  methods,
		})
	}

	if !hasStableSelection {
		for i := range rows {
			if rows[i].Name == "class" && androidNodeAttrSelectable(rows[i]) {
				rows[i].Selected = true
				break
			}
		}
	}

	return rows
}

func androidNodeTreeID(node *AndroidUINode) string {
	if node == nil {
		return ""
	}
	return fmt.Sprintf("node-%d", node.Number)
}

func androidNodeTreeBottomSpacerID(index int) string {
	return fmt.Sprintf("node-tree-bottom-spacer-%d", index)
}

func androidRootNodes(nodes []*AndroidUINode) []*AndroidUINode {
	roots := make([]*AndroidUINode, 0)
	for _, node := range nodes {
		if node != nil && node.Depth == 0 {
			roots = append(roots, node)
		}
	}
	if len(roots) == 0 && len(nodes) > 0 && nodes[0] != nil {
		roots = append(roots, nodes[0])
	}
	return roots
}

func androidNodeParentMap(nodes []*AndroidUINode) map[*AndroidUINode]*AndroidUINode {
	parents := make(map[*AndroidUINode]*AndroidUINode)

	var walk func(parent *AndroidUINode, node *AndroidUINode)
	walk = func(parent *AndroidUINode, node *AndroidUINode) {
		if node == nil {
			return
		}
		if parent != nil {
			parents[node] = parent
		}
		for _, child := range node.Children {
			walk(node, child)
		}
	}

	for _, root := range androidRootNodes(nodes) {
		walk(nil, root)
	}
	return parents
}

func androidNodeSummary(node *AndroidUINode) string {
	if node == nil {
		return ""
	}

	className := singleLineNodeText(shortAndroidClassName(node.Attrs["class"]))
	primary := singleLineNodeText(firstNonEmpty(
		node.Attrs["text"],
		node.Attrs["content-desc"],
		node.Attrs["resource-id"],
		node.Attrs["bounds"],
	))
	indent := strings.Repeat("  ", min(node.Depth, 6))
	return fmt.Sprintf("%03d %s%s  %s", node.Number, indent, className, trimMiddle(primary, 48))
}

func androidNodeMatches(node *AndroidUINode, keyword string) bool {
	if strings.Contains(strings.ToLower(androidNodeSummary(node)), keyword) {
		return true
	}
	for name, value := range node.Attrs {
		if strings.Contains(strings.ToLower(name), keyword) || strings.Contains(strings.ToLower(value), keyword) {
			return true
		}
	}
	return false
}

func selectedAndroidNodeAttrRows(rows []androidNodeAttrRow) []androidNodeAttrRow {
	selected := make([]androidNodeAttrRow, 0, len(rows))
	for _, row := range rows {
		if row.Selected && androidNodeAttrSelectable(row) {
			selected = append(selected, row)
		}
	}
	return selected
}

type androidNodeSelectorOptions struct {
	DisplayID string
	Function  string
	Template  string
	Timeout   string
}

func buildAndroidNodeSelectorCode(rows []androidNodeAttrRow, options androidNodeSelectorOptions) (string, error) {
	chain, err := buildAndroidNodeSelectorChain(rows)
	if err != nil {
		return "", err
	}
	if chain == "" {
		return "", nil
	}

	function := normalizeAndroidNodeSelectorFunction(options.Function)
	timeout := strings.TrimSpace(options.Timeout)
	if timeout == "" {
		timeout = androidNodeSelectorDefaultTimeout
	}
	displayID := strings.TrimSpace(options.DisplayID)
	if displayID == "" {
		displayID = "0"
	}
	template := options.Template
	if strings.TrimSpace(template) == "" {
		template = androidNodeSelectorDefaultTemplate
	}
	call := androidNodeSelectorCall(function, timeout)

	return strings.NewReplacer(
		"{displayId}", displayID,
		"{chain}", chain,
		"{params}", chain,
		"{function}", function,
		"{call}", call,
		"{timeout}", timeout,
	).Replace(template), nil
}

func buildAndroidNodeSelectorChain(rows []androidNodeAttrRow) (string, error) {
	var parts []string
	for _, row := range rows {
		if !row.Selected || !androidNodeAttrSelectable(row) {
			continue
		}

		method := androidNodeAttrMethod(row)
		if method == "" {
			return "", fmt.Errorf("%s 没有可用 uiacc 函数", row.Name)
		}
		args, err := androidNodeSelectorArgs(row, method)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf(".%s(%s)", method, args))
	}
	return strings.Join(parts, ""), nil
}

func androidNodeSelectorArgs(row androidNodeAttrRow, method string) (string, error) {
	kind := androidNodeAttrKindForMethod(row, method)
	if strings.TrimSpace(row.Value) == "" {
		switch kind {
		case androidNodeAttrString:
			return strconv.Quote(row.Value), nil
		case androidNodeAttrBool:
			return "false", nil
		case androidNodeAttrInt:
			return "0", nil
		case androidNodeAttrBounds:
			return "0, 0, 0, 0", nil
		}
	}

	switch kind {
	case androidNodeAttrString:
		return strconv.Quote(row.Value), nil
	case androidNodeAttrBool:
		v, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(row.Value)))
		if err != nil {
			return "", fmt.Errorf("%s 布尔值无效: %q", row.Name, row.Value)
		}
		if v {
			return "true", nil
		}
		return "false", nil
	case androidNodeAttrInt:
		v, err := strconv.Atoi(strings.TrimSpace(row.Value))
		if err != nil {
			return "", fmt.Errorf("%s 数值无效: %q", row.Name, row.Value)
		}
		return strconv.Itoa(v), nil
	case androidNodeAttrBounds:
		rect, err := parseAndroidBoundsStrict(row.Value)
		if err != nil {
			return "", fmt.Errorf("%s 边界值无效: %v", row.Name, err)
		}
		return fmt.Sprintf("%d, %d, %d, %d", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y), nil
	default:
		return "", fmt.Errorf("%s 不支持生成 uiacc 选择器", row.Name)
	}
}

func normalizeAndroidNodeSelectorFunction(function string) string {
	switch strings.TrimSpace(function) {
	case androidNodeSelectorFuncFind, androidNodeSelectorFuncWaitFor:
		return strings.TrimSpace(function)
	default:
		return androidNodeSelectorFuncFindOnce
	}
}

func androidNodeSelectorCall(function, timeout string) string {
	switch normalizeAndroidNodeSelectorFunction(function) {
	case androidNodeSelectorFuncFind:
		return "Find()"
	case androidNodeSelectorFuncWaitFor:
		return fmt.Sprintf("WaitFor(%s)", timeout)
	default:
		return "FindOnce()"
	}
}

func androidNodeMatchesAttrs(node *AndroidUINode, selected []androidNodeAttrRow) (bool, error) {
	for _, row := range selected {
		matched, err := androidNodeMatchesAttr(node, row)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func androidNodeMatchesAttr(node *AndroidUINode, row androidNodeAttrRow) (bool, error) {
	method := androidNodeAttrMethod(row)
	value := androidNodeValueByAttrRow(node, row)
	if row.Value == "" {
		return value == "", nil
	}
	switch androidNodeAttrKindForMethod(row, method) {
	case androidNodeAttrString:
		return androidNodeStringMatches(method, value, row.Value)
	case androidNodeAttrBool:
		left, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(value)))
		if err != nil {
			return false, nil
		}
		right, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(row.Value)))
		if err != nil {
			return false, fmt.Errorf("%s 布尔值无效: %q", row.Name, row.Value)
		}
		return left == right, nil
	case androidNodeAttrInt:
		left, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return false, nil
		}
		right, err := strconv.Atoi(strings.TrimSpace(row.Value))
		if err != nil {
			return false, fmt.Errorf("%s 数值无效: %q", row.Name, row.Value)
		}
		return left == right, nil
	case androidNodeAttrBounds:
		left, err := parseAndroidBoundsStrict(value)
		if err != nil {
			return false, nil
		}
		right, err := parseAndroidBoundsStrict(row.Value)
		if err != nil {
			return false, fmt.Errorf("%s 边界值无效: %v", row.Name, err)
		}
		return androidNodeBoundsMatches(method, left, right), nil
	default:
		return false, fmt.Errorf("%s 不支持查找测试", row.Name)
	}
}

func androidNodeStringMatches(method, value, expected string) (bool, error) {
	switch {
	case strings.HasSuffix(method, "Contains"):
		return strings.Contains(value, expected), nil
	case strings.HasSuffix(method, "StartsWith"):
		return strings.HasPrefix(value, expected), nil
	case strings.HasSuffix(method, "EndsWith"):
		return strings.HasSuffix(value, expected), nil
	case strings.HasSuffix(method, "Matches"):
		re, err := regexp.Compile(expected)
		if err != nil {
			return false, fmt.Errorf("%s 正则无效: %v", method, err)
		}
		return re.MatchString(value), nil
	default:
		return value == expected, nil
	}
}

func androidNodeBoundsMatches(method string, value, expected image.Rectangle) bool {
	switch method {
	case "BoundsInside":
		return value.Min.X >= expected.Min.X &&
			value.Min.Y >= expected.Min.Y &&
			value.Max.X <= expected.Max.X &&
			value.Max.Y <= expected.Max.Y
	case "BoundsContains":
		return value.Min.X <= expected.Min.X &&
			value.Min.Y <= expected.Min.Y &&
			value.Max.X >= expected.Max.X &&
			value.Max.Y >= expected.Max.Y
	default:
		return value == expected
	}
}

func androidNodeAttrValueByMeta(node *AndroidUINode, meta androidNodeAttrMeta) string {
	return androidNodeAttrValue(node, meta.Name, meta.XMLAttr, meta.Finder, meta.Kind)
}

func androidNodeValueByAttrRow(node *AndroidUINode, row androidNodeAttrRow) string {
	return androidNodeAttrValue(node, row.Name, row.XMLAttr, row.Finder, row.Kind)
}

func androidNodeAttrValue(node *AndroidUINode, name, xmlAttr, finder string, kind androidNodeAttrKind) string {
	if node == nil {
		return ""
	}
	if name == "depth" {
		return strconv.Itoa(node.Depth)
	}
	for _, key := range androidNodeAttrXMLKeys(name, xmlAttr, finder) {
		if value, ok := node.Attrs[key]; ok {
			return value
		}
	}
	if kind == androidNodeAttrBool {
		if name == "visible" {
			return "true"
		}
		return "false"
	}
	return ""
}

func androidNodeAttrXMLKeys(name, xmlAttr, finder string) []string {
	keys := make([]string, 0, 4)
	add := func(key string) {
		if key == "" {
			return
		}
		for _, existing := range keys {
			if existing == key {
				return
			}
		}
		keys = append(keys, key)
	}

	add(xmlAttr)
	add(finder)
	switch name {
	case "desc":
		add("content-desc")
	case "id":
		add("resource-id")
	case "class":
		add("className")
	case "package":
		add("packageName")
	case "drawingOrder":
		add("drawing-order")
		add("drawingOrder")
	case "longClickable":
		add("long-clickable")
		add("longClickable")
		add("longClick")
	case "multiLine":
		add("multi-line")
		add("multiLine")
	case "contextClickable":
		add("context-clickable")
		add("contextClickable")
	}
	return keys
}

func androidNodeAttrMethod(row androidNodeAttrRow) string {
	if row.Method != "" {
		return row.Method
	}
	switch row.Finder {
	case "text":
		return "Text"
	case "desc":
		return "Desc"
	case "id":
		return "Id"
	case "class":
		return "ClassName"
	case "package":
		return "PackageName"
	case "index":
		return "Index"
	case "depth":
		return "Depth"
	case "drawing-order":
		return "DrawingOrder"
	case "checkable":
		return "Checkable"
	case "checked":
		return "Checked"
	case "clickable":
		return "Clickable"
	case "enabled":
		return "Enabled"
	case "focusable":
		return "Focusable"
	case "focused":
		return "Focused"
	case "scrollable":
		return "Scrollable"
	case "editable":
		return "Editable"
	case "long-clickable":
		return "LongClickable"
	case "password":
		return "Password"
	case "selected":
		return "Selected"
	case "visible":
		return "Visible"
	case "multi-line":
		return "MultiLine"
	case "dismissable":
		return "Dismissable"
	case "context-clickable":
		return "ContextClickable"
	default:
		return ""
	}
}

func androidNodeAttrKindForMethod(row androidNodeAttrRow, method string) androidNodeAttrKind {
	if row.Kind != androidNodeAttrUnsupported {
		return row.Kind
	}
	if method == "Depth" || method == "Index" || method == "DrawingOrder" {
		return androidNodeAttrInt
	}
	if method == "Bounds" || method == "BoundsInside" || method == "BoundsContains" {
		return androidNodeAttrBounds
	}
	if strings.HasPrefix(method, "Text") ||
		strings.HasPrefix(method, "Desc") ||
		strings.HasPrefix(method, "Id") ||
		strings.HasPrefix(method, "ClassName") ||
		strings.HasPrefix(method, "PackageName") {
		return androidNodeAttrString
	}
	return androidNodeAttrBool
}

func parseAndroidBoundsStrict(value string) (image.Rectangle, error) {
	cleaned := strings.NewReplacer("[", " ", "]", " ", ",", " ").Replace(value)
	fields := strings.Fields(cleaned)
	if len(fields) != 4 {
		return image.Rectangle{}, fmt.Errorf("需要 4 个坐标，实际 %d 个", len(fields))
	}

	nums := make([]int, 4)
	for i, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			return image.Rectangle{}, err
		}
		nums[i] = n
	}
	return image.Rect(nums[0], nums[1], nums[2], nums[3]), nil
}

func androidNodeValueByFinder(node *AndroidUINode, finder, name string) string {
	if node == nil {
		return ""
	}
	switch finder {
	case "text":
		return node.Attrs["text"]
	case "desc":
		return node.Attrs["content-desc"]
	case "id":
		return node.Attrs["resource-id"]
	case "class":
		return node.Attrs["class"]
	case "package":
		return node.Attrs["package"]
	case "index":
		return node.Attrs["index"]
	case "drawing-order":
		return node.Attrs["drawing-order"]
	default:
		if name == "depth" {
			return strconv.Itoa(node.Depth)
		}
		return node.Attrs[finder]
	}
}

func androidNodeFinderToXMLAttr(finder string) string {
	switch finder {
	case "text":
		return "text"
	case "desc":
		return "content-desc"
	case "id":
		return "resource-id"
	case "class":
		return "class"
	case "package":
		return "package"
	case "index":
		return "index"
	case "checkable", "checked", "clickable", "enabled", "focusable", "focused", "scrollable", "editable", "password", "selected", "visible", "dismissable":
		return finder
	case "drawing-order":
		return "drawing-order"
	case "long-clickable":
		return "long-clickable"
	case "multi-line":
		return "multi-line"
	case "context-clickable":
		return "context-clickable"
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func singleLineNodeText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func shortAndroidClassName(value string) string {
	if value == "" {
		return "<unknown>"
	}
	if idx := strings.LastIndex(value, "."); idx != -1 && idx+1 < len(value) {
		return value[idx+1:]
	}
	return value
}

func trimMiddle(value string, maxLen int) string {
	runes := []rune(value)
	if maxLen <= 0 || len(runes) <= maxLen {
		return value
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	left := maxLen/2 - 1
	right := maxLen - left - 3
	return string(runes[:left]) + "..." + string(runes[len(runes)-right:])
}

func xpathLiteral(value string) string {
	if !strings.Contains(value, "'") {
		return "'" + value + "'"
	}
	if !strings.Contains(value, "\"") {
		return "\"" + value + "\""
	}
	parts := strings.Split(value, "'")
	xpathParts := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if part != "" {
			xpathParts = append(xpathParts, "'"+part+"'")
		}
		if i < len(parts)-1 {
			xpathParts = append(xpathParts, "\"'\"")
		}
	}
	return "concat(" + strings.Join(xpathParts, ", ") + ")"
}

func adbErrorWithOutput(err error, output string) error {
	output = strings.TrimSpace(output)
	if output == "" {
		return err
	}
	return fmt.Errorf("%v: %s", err, output)
}

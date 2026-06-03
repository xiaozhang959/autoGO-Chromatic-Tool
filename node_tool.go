package main

import (
	"encoding/xml"
	"fmt"
	"image"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const androidNodeDumpPath = "/sdcard/window_dump.xml"
const compactNodeToolAttrSelectWidth float32 = 40
const compactNodeToolAttrNameWidth float32 = 68
const compactNodeToolAttrFinderWidth float32 = 118

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

type AndroidNodeTool struct {
	window fyne.Window

	getSelectedDevice func() string
	getImageViewer    func() *ImageViewer
	openNodeImage     func(image.Image, func(int, int)) *ImageViewer
	onOpen            func()

	root             fyne.CanvasObject
	simpleCheck      *widget.Check
	captureBtn       *widget.Button
	searchEntry      *widget.Entry
	statusLabel      *widget.Label
	nodeTree         *widget.Tree
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
		treeChildren:      make(map[string][]string),
		treeNodeByID:      make(map[string]*AndroidUINode),
		treeIDByNode:      make(map[*AndroidUINode]string),
		treeParentID:      make(map[string]string),
	}

	tool.simpleCheck = widget.NewCheck("获取简单节点", nil)
	tool.simpleCheck.SetChecked(false)

	invisibleCheck := widget.NewCheck("获取不可见节点", nil)
	invisibleCheck.Disable()

	tool.captureBtn = widget.NewButton("抓取节点", func() {
		tool.Capture()
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

	tool.nodeTree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return tool.treeChildren[uid]
		},
		func(uid widget.TreeNodeID) bool {
			return uid == "" || len(tool.treeChildren[uid]) > 0
		},
		func(bool) fyne.CanvasObject {
			return container.NewHBox(newNodeTreeToolText())
		},
		func(uid widget.TreeNodeID, branch bool, item fyne.CanvasObject) {
			row := item.(*fyne.Container)
			label := row.Objects[0].(*widget.Label)
			if uid == "" {
				setCompactNodeToolText(label, "")
				return
			}
			node := tool.treeNodeByID[uid]
			if node == nil {
				setCompactNodeToolText(label, "")
				return
			}
			prefix := "  "
			if node == tool.selectedNode {
				prefix = "▶ "
			}
			setCompactNodeToolText(label, prefix+androidNodeSummary(node))
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

	tool.root = container.NewVBox(
		container.NewHBox(tool.simpleCheck, invisibleCheck, tool.captureBtn),
		container.NewBorder(nil, nil, nil, container.NewHBox(searchBtn, prevBtn, nextBtn), tool.searchEntry),
		tool.statusLabel,
		nodeHeader,
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, tool.nodeTree), 230),
		attrHeader,
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, container.NewVScroll(tool.attrList)), 180),
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

		snapshot, err := captureAndroidNodeSnapshot(device, t.simpleCheck.Checked)
		fyne.Do(func() {
			t.setBusy(false)
			if err != nil {
				t.setStatus("抓取节点失败")
				dialog.ShowError(fmt.Errorf("抓取节点失败: %v", err), t.window)
				return
			}

			if t.openNodeImage != nil {
				t.nodeViewer = t.openNodeImage(capturedImg, func(x, y int) {
					t.selectNodeAtPoint(x, y)
				})
			}
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

func (t *AndroidNodeTool) setSnapshot(snapshot *AndroidNodeSnapshot) {
	t.snapshot = snapshot
	t.selectedNode = nil
	t.attrRows = t.attrRows[:0]
	t.updateNodeOverlay()
	t.applySearch()

	t.setStatus(fmt.Sprintf("已抓取 %d 个节点 · 设备: %s", len(snapshot.Nodes), snapshot.Device))
}

func (t *AndroidNodeTool) applySearch() {
	t.filteredNodes = t.filteredNodes[:0]
	if t.snapshot == nil {
		t.rebuildNodeTree("")
		t.refreshLists()
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
		t.selectNode(t.filteredNodes[0])
	} else {
		t.selectedNode = nil
		t.attrRows = t.attrRows[:0]
		t.refreshSelector()
		t.refreshLists()
		t.clearSelectedNodeOverlay()
		t.setStatus(fmt.Sprintf("没有匹配节点 · 总节点: %d", len(t.snapshot.Nodes)))
	}
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
	t.selectedNode = node
	t.attrRows = buildAndroidNodeAttrRows(node)
	t.refreshSelector()
	t.syncNodeListSelection(node)
	t.refreshLists()
	t.highlightSelectedNode()
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

func (t *AndroidNodeTool) newAndroidNodeAttrRowContent(index int, attr androidNodeAttrRow) *fyne.Container {
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
		}
	} else {
		methodSelect.Disable()
	}

	return newCompactNodeToolAttrRow(
		check,
		newCompactNodeToolText(attr.Name),
		newCompactNodeToolText(trimMiddle(attr.Value, 28)),
		methodSelect,
	)
}

func androidNodeAttrSelectable(attr androidNodeAttrRow) bool {
	if attr.Value == "" {
		return false
	}
	if attr.Method != "" {
		return attr.Kind != androidNodeAttrUnsupported
	}
	return androidNodeFinderToXMLAttr(attr.Finder) != ""
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

	if t.nodeTree != nil {
		t.nodeTree.Refresh()
		t.nodeTree.OpenAllBranches()
	}
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
	if t.snapshot == nil || t.selectedNode == nil {
		t.setStatus("请先抓取并选择节点")
		return
	}

	selected := selectedAndroidNodeAttrRows(t.attrRows)
	if len(selected) == 0 {
		t.setStatus("请先勾选至少一个有效属性")
		return
	}

	matches := 0
	for _, node := range t.snapshot.Nodes {
		matched, err := androidNodeMatchesAttrs(node, selected)
		if err != nil {
			t.setStatus("查找测试失败: " + err.Error())
			return
		}
		if matched {
			matches++
		}
	}

	level := "不唯一"
	if matches == 1 {
		level = "唯一"
	}
	t.setStatus(fmt.Sprintf("查找测试: 匹配 %d 个节点 · %s", matches, level))
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

	className := shortAndroidClassName(node.Attrs["class"])
	primary := firstNonEmpty(
		node.Attrs["text"],
		node.Attrs["content-desc"],
		node.Attrs["resource-id"],
		node.Attrs["bounds"],
	)
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
	switch androidNodeAttrKindForMethod(row, method) {
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

package main

import (
	"encoding/xml"
	"fmt"
	"image"
	"io"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const androidNodeDumpPath = "/sdcard/window_dump.xml"

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
	Finder   string
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
	attrList         *widget.List
	selectorEntry    *widget.Entry
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
			return container.NewHBox(widget.NewLabel(""))
		},
		func(uid widget.TreeNodeID, branch bool, item fyne.CanvasObject) {
			row := item.(*fyne.Container)
			label := row.Objects[0].(*widget.Label)
			if uid == "" {
				label.SetText("")
				return
			}
			node := tool.treeNodeByID[uid]
			if node == nil {
				label.SetText("")
				return
			}
			prefix := "  "
			if node == tool.selectedNode {
				prefix = "▶ "
			}
			label.SetText(prefix + androidNodeSummary(node))
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

	tool.attrList = widget.NewList(
		func() int {
			return len(tool.attrRows)
		},
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(
				4,
				widget.NewCheck("", nil),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			row := item.(*fyne.Container)
			check := row.Objects[0].(*widget.Check)
			nameLabel := row.Objects[1].(*widget.Label)
			valueLabel := row.Objects[2].(*widget.Label)
			finderLabel := row.Objects[3].(*widget.Label)

			if id < 0 || id >= len(tool.attrRows) {
				check.OnChanged = nil
				check.SetChecked(false)
				nameLabel.SetText("")
				valueLabel.SetText("")
				finderLabel.SetText("")
				return
			}

			attr := tool.attrRows[id]
			check.OnChanged = nil
			check.SetChecked(attr.Selected)
			check.OnChanged = func(checked bool) {
				if id < 0 || id >= len(tool.attrRows) {
					return
				}
				tool.attrRows[id].Selected = checked
				tool.refreshSelector()
			}
			nameLabel.SetText(attr.Name)
			valueLabel.SetText(trimMiddle(attr.Value, 32))
			finderLabel.SetText(attr.Finder)
		},
	)

	tool.selectAllBtn = widget.NewButton("全部勾选", func() {
		tool.setAllAttrRows(true)
	})
	tool.clearSelectedBtn = widget.NewButton("清除全选", func() {
		tool.setAllAttrRows(false)
	})
	tool.testSelectorBtn = widget.NewButton("查找测试", func() {
		tool.testSelectedAttrs()
	})
	tool.copySelectorBtn = widget.NewButton("复制选择器", func() {
		if tool.window != nil {
			tool.window.Clipboard().SetContent(tool.selectorEntry.Text)
		}
	})
	tool.copyAttrsBtn = widget.NewButton("复制属性", func() {
		if tool.window != nil {
			tool.window.Clipboard().SetContent(tool.selectedAttrsText())
		}
	})

	tool.selectorEntry = widget.NewMultiLineEntry()
	tool.selectorEntry.SetPlaceHolder("选择节点并勾选属性后生成 XPath 选择器")
	tool.selectorEntry.SetMinRowsVisible(4)

	nodeHeader := widget.NewLabel("节点树")
	attrHeader := container.NewGridWithColumns(
		4,
		widget.NewLabel("勾选"),
		widget.NewLabel("属性"),
		widget.NewLabel("值"),
		widget.NewLabel("查找函数"),
	)

	tool.root = container.NewVBox(
		container.NewHBox(tool.simpleCheck, invisibleCheck, tool.captureBtn),
		container.NewBorder(nil, nil, nil, container.NewHBox(searchBtn, prevBtn, nextBtn), tool.searchEntry),
		tool.statusLabel,
		nodeHeader,
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, tool.nodeTree), 190),
		attrHeader,
		newFixedHeightContainer(container.NewBorder(nil, nil, nil, nil, tool.attrList), 230),
		container.NewGridWithColumns(3, tool.selectAllBtn, tool.clearSelectedBtn, tool.testSelectorBtn),
		container.NewBorder(nil, nil, widget.NewLabel("选择器"), container.NewHBox(tool.copyAttrsBtn, tool.copySelectorBtn), tool.selectorEntry),
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
	if t.attrList != nil {
		t.attrList.Refresh()
	}
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
		t.attrRows[i].Selected = selected
	}
	t.refreshSelector()
	if t.attrList != nil {
		t.attrList.Refresh()
	}
}

func (t *AndroidNodeTool) refreshSelector() {
	if t.selectorEntry == nil {
		return
	}
	t.selectorEntry.SetText(t.selectedXPath())
}

func (t *AndroidNodeTool) selectedXPath() string {
	if len(t.attrRows) == 0 {
		return ""
	}

	var parts []string
	for _, row := range t.attrRows {
		if !row.Selected || row.Value == "" {
			continue
		}
		xmlAttr := androidNodeFinderToXMLAttr(row.Finder)
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

func (t *AndroidNodeTool) selectedAttrsText() string {
	var lines []string
	for _, row := range t.attrRows {
		if !row.Selected || row.Value == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s=%q", row.Name, row.Value))
	}
	return strings.Join(lines, "\n")
}

func (t *AndroidNodeTool) testSelectedAttrs() {
	if t.snapshot == nil || t.selectedNode == nil {
		t.setStatus("请先抓取并选择节点")
		return
	}

	selected := make([]androidNodeAttrRow, 0, len(t.attrRows))
	for _, row := range t.attrRows {
		if row.Selected && row.Value != "" {
			selected = append(selected, row)
		}
	}
	if len(selected) == 0 {
		t.setStatus("请先勾选至少一个有效属性")
		return
	}

	matches := 0
	for _, node := range t.snapshot.Nodes {
		if androidNodeMatchesAttrs(node, selected) {
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

	values := map[string]string{
		"depth":         strconv.Itoa(node.Depth),
		"index":         node.Attrs["index"],
		"class":         node.Attrs["class"],
		"package":       node.Attrs["package"],
		"text":          node.Attrs["text"],
		"desc":          node.Attrs["content-desc"],
		"id":            node.Attrs["resource-id"],
		"bounds":        node.Attrs["bounds"],
		"checkable":     node.Attrs["checkable"],
		"checked":       node.Attrs["checked"],
		"clickable":     node.Attrs["clickable"],
		"enabled":       node.Attrs["enabled"],
		"focusable":     node.Attrs["focusable"],
		"focused":       node.Attrs["focused"],
		"scrollable":    node.Attrs["scrollable"],
		"longClickable": node.Attrs["long-clickable"],
		"password":      node.Attrs["password"],
		"selected":      node.Attrs["selected"],
	}
	finders := map[string]string{
		"depth":         "",
		"index":         "index",
		"class":         "class",
		"package":       "package",
		"text":          "text",
		"desc":          "desc",
		"id":            "id",
		"bounds":        "",
		"checkable":     "checkable",
		"checked":       "checked",
		"clickable":     "clickable",
		"enabled":       "enabled",
		"focusable":     "focusable",
		"focused":       "focused",
		"scrollable":    "scrollable",
		"longClickable": "long-clickable",
		"password":      "password",
		"selected":      "selected",
	}

	preferred := []string{
		"depth", "index", "class", "package", "text", "desc", "id", "bounds",
		"checkable", "checked", "clickable", "enabled", "focusable", "focused",
		"scrollable", "longClickable", "password", "selected",
	}

	rows := make([]androidNodeAttrRow, 0, len(preferred))
	hasStableSelection := false
	for _, name := range preferred {
		value := values[name]
		if value == "" && name != "depth" {
			continue
		}

		selected := value != "" && (name == "id" || name == "desc" || name == "text")
		if selected {
			hasStableSelection = true
		}
		rows = append(rows, androidNodeAttrRow{
			Selected: selected,
			Name:     name,
			Value:    value,
			Finder:   finders[name],
		})
	}

	if !hasStableSelection {
		for i := range rows {
			if rows[i].Name == "class" && rows[i].Value != "" {
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

func androidNodeMatchesAttrs(node *AndroidUINode, selected []androidNodeAttrRow) bool {
	for _, row := range selected {
		if androidNodeValueByFinder(node, row.Finder, row.Name) != row.Value {
			return false
		}
	}
	return true
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
	case "checkable", "checked", "clickable", "enabled", "focusable", "focused", "scrollable", "password", "selected":
		return finder
	case "long-clickable":
		return "long-clickable"
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

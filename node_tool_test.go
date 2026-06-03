package main

import (
	"image"
	"strings"
	"testing"
)

func TestParseAndroidNodeXML(t *testing.T) {
	xmlText := `<hierarchy rotation="0">
		<node index="0" text="" resource-id="root" class="android.widget.FrameLayout" package="demo" content-desc="" bounds="[0,0][100,200]">
			<node index="1" text="登录" resource-id="demo:id/login" class="android.widget.Button" package="demo" content-desc="login button" bounds="[10,20][80,60]" />
		</node>
	</hierarchy>`

	nodes, err := parseAndroidNodeXML(xmlText)
	if err != nil {
		t.Fatalf("parseAndroidNodeXML returned error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	root := nodes[0]
	if root.Depth != 0 || root.Bounds != image.Rect(0, 0, 100, 200) {
		t.Fatalf("unexpected root node: depth=%d bounds=%v", root.Depth, root.Bounds)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected root to have 1 child, got %d", len(root.Children))
	}

	child := nodes[1]
	if child.Depth != 1 {
		t.Fatalf("expected child depth 1, got %d", child.Depth)
	}
	if child.Attrs["text"] != "登录" {
		t.Fatalf("expected child text 登录, got %q", child.Attrs["text"])
	}
	if child.Bounds != image.Rect(10, 20, 80, 60) {
		t.Fatalf("unexpected child bounds: %v", child.Bounds)
	}
}

func TestAndroidNodeSummaryUsesSingleLineText(t *testing.T) {
	node := &AndroidUINode{
		Number: 1,
		Depth:  1,
		Attrs: map[string]string{
			"class": "android.widget.TextView\nBad",
			"text":  "move vx,vy\n移动内容 vy 到 vx",
		},
	}

	summary := androidNodeSummary(node)
	if strings.Contains(summary, "\n") || strings.Contains(summary, "\t") {
		t.Fatalf("expected single-line summary, got %q", summary)
	}
	if !strings.Contains(summary, "TextView Bad") {
		t.Fatalf("expected class whitespace to be normalized, got %q", summary)
	}
	if !strings.Contains(summary, "move vx,vy 移动内容 vy 到 vx") {
		t.Fatalf("expected text whitespace to be normalized, got %q", summary)
	}
}

func TestBuildAndroidNodeAttrRowsSelectsStableAttrs(t *testing.T) {
	node := &AndroidUINode{
		Depth: 2,
		Attrs: map[string]string{
			"class":        "android.widget.Button",
			"text":         "登录",
			"content-desc": "login button",
			"resource-id":  "demo:id/login",
			"bounds":       "[10,20][80,60]",
		},
	}

	rows := buildAndroidNodeAttrRows(node)
	selected := map[string]bool{}
	methods := map[string]string{}
	for _, row := range rows {
		methods[row.Name] = row.Method
		if row.Selected {
			selected[row.Name] = true
		}
	}

	for _, name := range []string{"text", "desc", "id"} {
		if !selected[name] {
			t.Fatalf("expected %s to be selected by default", name)
		}
	}
	if selected["bounds"] {
		t.Fatal("bounds should not be selected by default")
	}
	if methods["text"] != "Text" || methods["desc"] != "Desc" || methods["id"] != "Id" {
		t.Fatalf("unexpected default methods: text=%q desc=%q id=%q", methods["text"], methods["desc"], methods["id"])
	}
	if methods["depth"] != "Depth" {
		t.Fatalf("expected depth default method Depth, got %q", methods["depth"])
	}
	if methods["bounds"] != "Bounds" {
		t.Fatalf("expected bounds default method Bounds, got %q", methods["bounds"])
	}
}

func TestBuildAndroidNodeAttrRowsIncludesUiaccMatchableAttrs(t *testing.T) {
	node := &AndroidUINode{
		Depth: 1,
		Attrs: map[string]string{
			"class": "android.widget.EditText",
		},
	}

	rows := buildAndroidNodeAttrRows(node)
	byName := map[string]androidNodeAttrRow{}
	for _, row := range rows {
		byName[row.Name] = row
	}

	for _, name := range []string{
		"depth", "index", "drawingOrder", "class", "package", "text", "desc", "id", "bounds",
		"checkable", "checked", "clickable", "enabled", "focusable", "focused",
		"scrollable", "editable", "longClickable", "password", "selected",
		"visible", "multiLine", "dismissable", "contextClickable",
	} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("expected uiacc matchable attr %s to be listed", name)
		}
	}
	if byName["editable"].Value != "false" {
		t.Fatalf("expected missing editable to default false, got %q", byName["editable"].Value)
	}
	if byName["visible"].Value != "true" {
		t.Fatalf("expected missing visible to default true, got %q", byName["visible"].Value)
	}
	if !androidNodeAttrSelectable(byName["drawingOrder"]) {
		t.Fatal("empty drawingOrder should be selectable for snapshot filtering")
	}
	if byName["desc"].Value != "" {
		t.Fatalf("expected missing desc to stay empty, got %q", byName["desc"].Value)
	}
	if !androidNodeAttrSelectable(byName["desc"]) {
		t.Fatal("empty desc should be selectable")
	}
	if !androidNodeAttrSelectable(byName["bounds"]) {
		t.Fatal("empty bounds should be selectable for snapshot filtering")
	}
}

func TestAndroidNodeAttrValueUsesAliases(t *testing.T) {
	node := &AndroidUINode{
		Attrs: map[string]string{
			"drawingOrder":     "3",
			"longClickable":    "true",
			"multiLine":        "true",
			"contextClickable": "true",
		},
	}

	if got := androidNodeAttrValue(node, "drawingOrder", "drawing-order", "drawing-order", androidNodeAttrInt); got != "3" {
		t.Fatalf("expected drawingOrder alias value 3, got %q", got)
	}
	if got := androidNodeAttrValue(node, "longClickable", "long-clickable", "long-clickable", androidNodeAttrBool); got != "true" {
		t.Fatalf("expected longClickable alias true, got %q", got)
	}
	if got := androidNodeAttrValue(node, "multiLine", "multi-line", "multi-line", androidNodeAttrBool); got != "true" {
		t.Fatalf("expected multiLine alias true, got %q", got)
	}
	if got := androidNodeAttrValue(node, "contextClickable", "context-clickable", "context-clickable", androidNodeAttrBool); got != "true" {
		t.Fatalf("expected contextClickable alias true, got %q", got)
	}
}

func TestSelectedXPath(t *testing.T) {
	tool := &AndroidNodeTool{
		attrRows: []androidNodeAttrRow{
			{Selected: true, Name: "id", Value: "demo:id/login", Finder: "id"},
			{Selected: true, Name: "text", Value: `登录'按钮"`, Finder: "text"},
			{Selected: false, Name: "class", Value: "android.widget.Button", Finder: "class"},
		},
	}

	xpath := tool.selectedXPath()
	if !strings.Contains(xpath, "[@resource-id='demo:id/login']") {
		t.Fatalf("expected resource-id condition in xpath, got %s", xpath)
	}
	if strings.Contains(xpath, "android.widget.Button") {
		t.Fatalf("unselected class should not appear in xpath: %s", xpath)
	}
	if !strings.Contains(xpath, "concat(") {
		t.Fatalf("expected concat literal for mixed quotes, got %s", xpath)
	}
}

func TestBuildAndroidNodeSelectorChain(t *testing.T) {
	rows := []androidNodeAttrRow{
		{Selected: true, Name: "depth", Value: "2", Kind: androidNodeAttrInt, Method: "Depth"},
		{Selected: true, Name: "text", Value: "登录", Kind: androidNodeAttrString, Method: "TextContains"},
		{Selected: true, Name: "desc", Value: "", Kind: androidNodeAttrString, Method: "Desc"},
		{Selected: true, Name: "clickable", Value: "true", Kind: androidNodeAttrBool, Method: "Clickable"},
		{Selected: true, Name: "bounds", Value: "[10,20][80,60]", Kind: androidNodeAttrBounds, Method: "Bounds"},
		{Selected: false, Name: "class", Value: "android.widget.Button", Kind: androidNodeAttrString, Method: "ClassName"},
	}

	chain, err := buildAndroidNodeSelectorChain(rows)
	if err != nil {
		t.Fatalf("buildAndroidNodeSelectorChain returned error: %v", err)
	}
	want := `.Depth(2).TextContains("登录").Desc("").Clickable(true).Bounds(10, 20, 80, 60)`
	if chain != want {
		t.Fatalf("expected chain %s, got %s", want, chain)
	}
}

func TestBuildAndroidNodeSelectorCodeUsesTemplate(t *testing.T) {
	rows := []androidNodeAttrRow{
		{Selected: true, Name: "id", Value: "demo:id/login", Kind: androidNodeAttrString, Method: "Id"},
	}

	code, err := buildAndroidNodeSelectorCode(rows, androidNodeSelectorOptions{
		DisplayID: "1",
		Function:  androidNodeSelectorFuncWaitFor,
		Template:  "acc := uiacc.New({displayId})\nobj := acc{params}.{call}",
		Timeout:   "5000",
	})
	if err != nil {
		t.Fatalf("buildAndroidNodeSelectorCode returned error: %v", err)
	}
	if !strings.Contains(code, "uiacc.New(1)") {
		t.Fatalf("expected display id in code, got %s", code)
	}
	if !strings.Contains(code, `.Id("demo:id/login").WaitFor(5000)`) {
		t.Fatalf("expected id wait chain in code, got %s", code)
	}
}

func TestBuildAndroidNodeSelectorChainGeneratesZeroValuesForEmptyArgs(t *testing.T) {
	rows := []androidNodeAttrRow{
		{Selected: true, Name: "drawingOrder", Value: "", Kind: androidNodeAttrInt, Method: "DrawingOrder"},
		{Selected: true, Name: "bounds", Value: "", Kind: androidNodeAttrBounds, Method: "Bounds"},
		{Selected: true, Name: "clickable", Value: "", Kind: androidNodeAttrBool, Method: "Clickable"},
	}

	chain, err := buildAndroidNodeSelectorChain(rows)
	if err != nil {
		t.Fatalf("buildAndroidNodeSelectorChain returned error: %v", err)
	}
	want := `.DrawingOrder(0).Bounds(0, 0, 0, 0).Clickable(false)`
	if chain != want {
		t.Fatalf("expected chain %s, got %s", want, chain)
	}
}

func TestAndroidNodeMatchesAttrsUsesMethodSemantics(t *testing.T) {
	node := &AndroidUINode{
		Depth: 2,
		Attrs: map[string]string{
			"text":      "登录按钮",
			"clickable": "true",
			"bounds":    "[10,20][80,60]",
		},
	}
	rows := []androidNodeAttrRow{
		{Selected: true, Name: "depth", Value: "2", Kind: androidNodeAttrInt, Method: "Depth"},
		{Selected: true, Name: "text", Value: "登录", XMLAttr: "text", Kind: androidNodeAttrString, Method: "TextContains"},
		{Selected: true, Name: "clickable", Value: "true", XMLAttr: "clickable", Kind: androidNodeAttrBool, Method: "Clickable"},
		{Selected: true, Name: "bounds", Value: "[0,0][100,100]", XMLAttr: "bounds", Kind: androidNodeAttrBounds, Method: "BoundsInside"},
	}

	matched, err := androidNodeMatchesAttrs(node, rows)
	if err != nil {
		t.Fatalf("androidNodeMatchesAttrs returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected node to match contains/bool/boundsInside selectors")
	}
}

func TestAndroidNodeMatchesAttrsSupportsEmptyAttrFilters(t *testing.T) {
	emptyNode := &AndroidUINode{
		Attrs: map[string]string{
			"text": "qq",
		},
	}
	nonEmptyNode := &AndroidUINode{
		Attrs: map[string]string{
			"text":          "qq",
			"content-desc":  "124",
			"drawing-order": "2",
			"bounds":        "[0,0][10,10]",
		},
	}
	rows := []androidNodeAttrRow{
		{Selected: true, Name: "text", Value: "qq", XMLAttr: "text", Kind: androidNodeAttrString, Method: "Text"},
		{Selected: true, Name: "desc", Value: "", XMLAttr: "content-desc", Kind: androidNodeAttrString, Method: "Desc"},
		{Selected: true, Name: "drawingOrder", Value: "", XMLAttr: "drawing-order", Kind: androidNodeAttrInt, Method: "DrawingOrder"},
		{Selected: true, Name: "bounds", Value: "", XMLAttr: "bounds", Kind: androidNodeAttrBounds, Method: "Bounds"},
	}

	matched, err := androidNodeMatchesAttrs(emptyNode, rows)
	if err != nil {
		t.Fatalf("androidNodeMatchesAttrs returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected node with missing attrs to match empty attr filters")
	}
	matched, err = androidNodeMatchesAttrs(nonEmptyNode, rows)
	if err != nil {
		t.Fatalf("androidNodeMatchesAttrs returned error: %v", err)
	}
	if matched {
		t.Fatal("expected node with non-empty attrs not to match empty attr filters")
	}
}

func TestAndroidNodeMatchesAttrsReportsInvalidRegex(t *testing.T) {
	node := &AndroidUINode{Attrs: map[string]string{"text": "登录"}}
	rows := []androidNodeAttrRow{
		{Selected: true, Name: "text", Value: "[", XMLAttr: "text", Kind: androidNodeAttrString, Method: "TextMatches"},
	}

	if _, err := androidNodeMatchesAttrs(node, rows); err == nil {
		t.Fatal("expected invalid regex error")
	}
}

func TestSmallestNodeAtPoint(t *testing.T) {
	root := &AndroidUINode{Number: 1, Depth: 0, Bounds: image.Rect(0, 0, 100, 100)}
	container := &AndroidUINode{Number: 2, Depth: 1, Bounds: image.Rect(10, 10, 90, 90)}
	button := &AndroidUINode{Number: 3, Depth: 2, Bounds: image.Rect(20, 20, 40, 40)}
	tool := &AndroidNodeTool{
		snapshot: &AndroidNodeSnapshot{
			Nodes: []*AndroidUINode{root, container, button},
		},
	}

	if got := tool.smallestNodeAtPoint(image.Pt(25, 25)); got != button {
		t.Fatalf("expected smallest button node, got %#v", got)
	}
	if got := tool.smallestNodeAtPoint(image.Pt(95, 95)); got != root {
		t.Fatalf("expected root node, got %#v", got)
	}
	if got := tool.smallestNodeAtPoint(image.Pt(120, 120)); got != nil {
		t.Fatalf("expected nil outside nodes, got %#v", got)
	}
}

func TestRebuildNodeTreeKeepsMatchedAncestors(t *testing.T) {
	root := &AndroidUINode{Number: 1, Depth: 0, Attrs: map[string]string{"class": "Root"}}
	container := &AndroidUINode{Number: 2, Depth: 1, Attrs: map[string]string{"class": "Container"}}
	button := &AndroidUINode{Number: 3, Depth: 2, Attrs: map[string]string{"text": "登录"}}
	root.Children = []*AndroidUINode{container}
	container.Children = []*AndroidUINode{button}

	tool := &AndroidNodeTool{
		snapshot:      &AndroidNodeSnapshot{Nodes: []*AndroidUINode{root, container, button}},
		filteredNodes: []*AndroidUINode{button},
	}
	tool.rebuildNodeTree("登录")

	rootID := androidNodeTreeID(root)
	containerID := androidNodeTreeID(container)
	buttonID := androidNodeTreeID(button)
	rootChildren := tool.treeChildren[""]
	if len(rootChildren) != 1+androidNodeTreeBottomSpacerRows || rootChildren[0] != rootID {
		t.Fatalf("expected root followed by bottom spacers, got %v", rootChildren)
	}
	if got := tool.treeChildren[rootID]; len(got) != 1 || got[0] != containerID {
		t.Fatalf("expected container ancestor visible, got %v", got)
	}
	if got := tool.treeChildren[containerID]; len(got) != 1 || got[0] != buttonID {
		t.Fatalf("expected matched button visible, got %v", got)
	}
}

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
	for _, row := range rows {
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
	if got := tool.treeChildren[""]; len(got) != 1 || got[0] != rootID {
		t.Fatalf("expected root visible, got %v", got)
	}
	if got := tool.treeChildren[rootID]; len(got) != 1 || got[0] != containerID {
		t.Fatalf("expected container ancestor visible, got %v", got)
	}
	if got := tool.treeChildren[containerID]; len(got) != 1 || got[0] != buttonID {
		t.Fatalf("expected matched button visible, got %v", got)
	}
}

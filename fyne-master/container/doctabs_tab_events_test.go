package container

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocTabs_TabButtonScrollAndSecondaryTapEvents(t *testing.T) {
	tabs := NewDocTabs(
		NewTabItem("Test1", widget.NewLabel("Test1")),
		NewTabItem("Test2", widget.NewLabel("Test2")),
	)
	renderer := test.TempWidgetRenderer(t, tabs).(*docTabsRenderer)
	buttons := docTabsTabButtons(renderer)
	require.Len(t, buttons, 2)

	var scrolledItem *TabItem
	var scrolledEvent *fyne.ScrollEvent
	tabs.OnTabScrolled = func(item *TabItem, event *fyne.ScrollEvent) {
		scrolledItem = item
		scrolledEvent = event
	}
	scrollEvent := &fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -1)}
	buttons[1].(*tabButton).Scrolled(scrollEvent)
	assert.Equal(t, tabs.Items[1], scrolledItem)
	assert.Equal(t, scrollEvent, scrolledEvent)

	var tappedItem *TabItem
	var tappedEvent *fyne.PointEvent
	tabs.OnTabSecondaryTapped = func(item *TabItem, event *fyne.PointEvent) {
		tappedItem = item
		tappedEvent = event
	}
	pointEvent := &fyne.PointEvent{Position: fyne.NewPos(3, 4)}
	buttons[0].(*tabButton).TappedSecondary(pointEvent)
	assert.Equal(t, tabs.Items[0], tappedItem)
	assert.Equal(t, pointEvent, tappedEvent)
}

func TestDocTabs_CloseTriggersOnClosed(t *testing.T) {
	tab1 := NewTabItem("Test1", widget.NewLabel("Test1"))
	tab2 := NewTabItem("Test2", widget.NewLabel("Test2"))
	tabs := NewDocTabs(tab1, tab2)

	var closed *TabItem
	tabs.OnClosed = func(tab *TabItem) {
		closed = tab
	}

	tabs.Close(tab1)

	assert.Equal(t, tab1, closed)
	assert.Equal(t, []*TabItem{tab2}, tabs.Items)
}

func docTabsTabButtons(renderer *docTabsRenderer) []fyne.CanvasObject {
	if scroll, ok := renderer.bar.Objects[0].(*Scroll); ok {
		return scroll.Content.(*fyne.Container).Objects
	}
	return renderer.bar.Objects[0].(*fyne.Container).Objects
}

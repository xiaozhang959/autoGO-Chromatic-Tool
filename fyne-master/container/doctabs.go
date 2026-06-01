package container

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Declare conformity with Widget interface.
var _ fyne.Widget = (*DocTabs)(nil)

// DocTabs container is used to display various pieces of content identified by tabs.
// The tabs contain text and/or an icon and allow the user to switch between the content specified in each TabItem.
// Each item is represented by a button at the edge of the container.
//
// Since: 2.1
type DocTabs struct {
	widget.BaseWidget

	Items []*TabItem

	CreateTab      func() *TabItem `json:"-"`
	CloseIntercept func(*TabItem)  `json:"-"`
	OnClosed       func(*TabItem)  `json:"-"`
	OnSelected     func(*TabItem)  `json:"-"`
	OnUnselected   func(*TabItem)  `json:"-"`

	current         int
	location        TabLocation
	isTransitioning bool

	popUpMenu *widget.PopUpMenu
}

// NewDocTabs creates a new tab container that allows the user to choose between various pieces of content.
//
// Since: 2.1
func NewDocTabs(items ...*TabItem) *DocTabs {
	tabs := &DocTabs{}
	tabs.ExtendBaseWidget(tabs)
	tabs.SetItems(items)
	return tabs
}

// Append adds a new TabItem to the end of the tab bar.
func (t *DocTabs) Append(item *TabItem) {
	t.SetItems(append(t.Items, item))
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
//
// Implements: fyne.Widget
func (t *DocTabs) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)
	th := t.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r := &docTabsRenderer{
		baseTabsRenderer: baseTabsRenderer{
			bar:       &fyne.Container{},
			divider:   canvas.NewRectangle(th.Color(theme.ColorNameShadow, v)),
			indicator: canvas.NewRectangle(th.Color(theme.ColorNamePrimary, v)),
		},
		docTabs:  t,
		scroller: &fyne.Container{Layout: layout.NewHBoxLayout()}, // 改为普通容器，不使用滚动
	}
	r.action = r.buildAllTabsButton()
	r.create = r.buildCreateTabsButton()
	r.tabs = t

	r.box = NewHBox(r.create, r.action)
	// 移除 OnScrolled 回调，因为不再是滚动容器
	r.updateAllTabs()
	r.updateCreateTab()
	r.updateTabs()
	r.updateIndicator(false)
	r.applyTheme(t)
	return r
}

// DisableIndex disables the TabItem at the specified index.
//
// Since: 2.3
func (t *DocTabs) DisableIndex(i int) {
	disableIndex(t, i)
}

// DisableItem disables the specified TabItem.
//
// Since: 2.3
func (t *DocTabs) DisableItem(item *TabItem) {
	disableItem(t, item)
}

// EnableIndex enables the TabItem at the specified index.
//
// Since: 2.3
func (t *DocTabs) EnableIndex(i int) {
	enableIndex(t, i)
}

// EnableItem enables the specified TabItem.
//
// Since: 2.3
func (t *DocTabs) EnableItem(item *TabItem) {
	enableItem(t, item)
}

// Hide hides the widget.
//
// Implements: fyne.CanvasObject
func (t *DocTabs) Hide() {
	if t.popUpMenu != nil {
		t.popUpMenu.Hide()
		t.popUpMenu = nil
	}
	t.BaseWidget.Hide()
}

// MinSize returns the size that this widget should not shrink below
//
// Implements: fyne.CanvasObject
func (t *DocTabs) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)
	return t.BaseWidget.MinSize()
}

// Remove tab by value.
func (t *DocTabs) Remove(item *TabItem) {
	removeItem(t, item)
	t.Refresh()
}

// RemoveIndex removes tab by index.
func (t *DocTabs) RemoveIndex(index int) {
	removeIndex(t, index)
	t.Refresh()
}

// Select sets the specified TabItem to be selected and its content visible.
func (t *DocTabs) Select(item *TabItem) {
	selectItem(t, item)
	t.Refresh()
}

// SelectIndex sets the TabItem at the specific index to be selected and its content visible.
func (t *DocTabs) SelectIndex(index int) {
	selectIndex(t, index)
}

// Selected returns the currently selected TabItem.
func (t *DocTabs) Selected() *TabItem {
	return selected(t)
}

// SelectedIndex returns the index of the currently selected TabItem.
func (t *DocTabs) SelectedIndex() int {
	return t.current
}

// SetItems sets the containers items and refreshes.
func (t *DocTabs) SetItems(items []*TabItem) {
	setItems(t, items)
	t.Refresh()
}

// SetTabLocation sets the location of the tab bar
func (t *DocTabs) SetTabLocation(l TabLocation) {
	t.location = tabsAdjustedLocation(l, t)
	t.Refresh()
}

// Show this widget, if it was previously hidden
//
// Implements: fyne.CanvasObject
func (t *DocTabs) Show() {
	t.BaseWidget.Show()
	t.SelectIndex(t.current)
}

func (t *DocTabs) close(item *TabItem) {
	if f := t.CloseIntercept; f != nil {
		f(item)
	} else {
		t.Remove(item)
		if f := t.OnClosed; f != nil {
			f(item)
		}
	}
}

func (t *DocTabs) onUnselected() func(*TabItem) {
	return t.OnUnselected
}

func (t *DocTabs) onSelected() func(*TabItem) {
	return t.OnSelected
}

func (t *DocTabs) items() []*TabItem {
	return t.Items
}

func (t *DocTabs) selected() int {
	return t.current
}

func (t *DocTabs) setItems(items []*TabItem) {
	t.Items = items
}

func (t *DocTabs) setSelected(selected int) {
	t.current = selected
}

func (t *DocTabs) setTransitioning(transitioning bool) {
	t.isTransitioning = transitioning
}

func (t *DocTabs) tabLocation() TabLocation {
	return t.location
}

func (t *DocTabs) transitioning() bool {
	return t.isTransitioning
}

// Declare conformity with WidgetRenderer interface.
var _ fyne.WidgetRenderer = (*docTabsRenderer)(nil)

type docTabsRenderer struct {
	baseTabsRenderer
	docTabs      *DocTabs
	scroller     *fyne.Container // 改为普通容器
	box          *fyne.Container
	create       *widget.Button
	lastSelected int
}

func (r *docTabsRenderer) Layout(size fyne.Size) {
	r.updateAllTabs()
	r.updateCreateTab()
	r.updateTabs()
	r.layout(r.docTabs, size)

	// lay out buttons before updating indicator, which is relative to their position
	buttons := r.scroller // 直接使用容器
	buttons.Layout.Layout(buttons.Objects, buttons.Size())
	r.updateIndicator(r.docTabs.transitioning())

	if r.docTabs.transitioning() {
		r.docTabs.setTransitioning(false)
	}
}

func (r *docTabsRenderer) MinSize() fyne.Size {
	return r.minSize(r.docTabs)
}

func (r *docTabsRenderer) Objects() []fyne.CanvasObject {
	return r.objects(r.docTabs)
}

func (r *docTabsRenderer) Refresh() {
	r.Layout(r.docTabs.Size())

	if c := r.docTabs.current; c != r.lastSelected {
		if c >= 0 && c < len(r.docTabs.Items) {
			r.scrollToSelected()
		}
		r.lastSelected = c
	}

	r.refresh(r.docTabs)

	canvas.Refresh(r.docTabs)
}

func (r *docTabsRenderer) buildAllTabsButton() (all *widget.Button) {
	all = &widget.Button{Importance: widget.LowImportance, OnTapped: func() {
		// Show pop up containing all tabs

		items := make([]*fyne.MenuItem, len(r.docTabs.Items))
		for i := 0; i < len(r.docTabs.Items); i++ {
			index := i // capture
			// FIXME MenuItem doesn't support icons (#1752)
			items[i] = fyne.NewMenuItem(r.docTabs.Items[i].Text, func() {
				r.docTabs.SelectIndex(index)
				if r.docTabs.popUpMenu != nil {
					r.docTabs.popUpMenu.Hide()
					r.docTabs.popUpMenu = nil
				}
			})
			items[i].Checked = index == r.docTabs.current
		}

		r.docTabs.popUpMenu = buildPopUpMenu(r.docTabs, all, items)
	}}

	return all
}

func (r *docTabsRenderer) buildCreateTabsButton() *widget.Button {
	create := widget.NewButton("", func() {
		if f := r.docTabs.CreateTab; f != nil {
			if tab := f(); tab != nil {
				r.docTabs.Append(tab)
				r.docTabs.SelectIndex(len(r.docTabs.Items) - 1)
			}
		}
	})
	create.Importance = widget.LowImportance
	return create
}

func (r *docTabsRenderer) buildTabButtons(count int, buttons *fyne.Container) {
	buttons.Objects = nil

	var iconPos buttonIconPosition
	if r.docTabs.location == TabLocationLeading || r.docTabs.location == TabLocationTrailing {
		buttons.Layout = layout.NewVBoxLayout()
		iconPos = buttonIconTop
	} else {
		buttons.Layout = layout.NewHBoxLayout()
		iconPos = buttonIconInline
	}

	for i := 0; i < count; i++ {
		item := r.docTabs.Items[i]
		if item.button == nil {
			item.button = &tabButton{
				onTapped: func() { r.docTabs.Select(item) },
				onClosed: func() { r.docTabs.close(item) },
				tabs:     r.tabs,
			}
		}
		button := item.button
		button.icon = item.Icon
		button.iconPosition = iconPos
		if i == r.docTabs.current {
			button.importance = widget.HighImportance
		} else {
			button.importance = widget.MediumImportance
		}
		button.text = item.Text
		button.textAlignment = fyne.TextAlignLeading
		button.Refresh()
		buttons.Objects = append(buttons.Objects, button)
	}
}

func (r *docTabsRenderer) scrollToSelected() {
	// 由于不使用滚动容器，这个函数现在只需要更新指示器
	r.updateIndicator(false)
}

func (r *docTabsRenderer) updateIndicator(animate bool) {
	th := r.docTabs.Theme()
	if r.docTabs.current < 0 {
		r.indicator.FillColor = color.Transparent
		r.moveIndicator(fyne.NewPos(0, 0), fyne.NewSize(0, 0), th, animate)
		return
	}

	var selectedPos fyne.Position
	var selectedSize fyne.Size

	buttons := r.scroller.Objects // 直接使用容器的Objects

	if r.docTabs.current >= len(buttons) {
		if a := r.action; a != nil {
			selectedPos = a.Position()
			selectedSize = a.Size()
			minSize := a.MinSize()
			if minSize.Width > selectedSize.Width {
				selectedSize = minSize
			}
		}
	} else {
		selected := buttons[r.docTabs.current]
		selectedPos = selected.Position()
		selectedSize = selected.Size()
		minSize := selected.MinSize()
		if minSize.Width > selectedSize.Width {
			selectedSize = minSize
		}
	}

	scrollOffset := fyne.NewPos(0, 0) // 不再使用滚动，offset总是0
	scrollSize := r.scroller.Size()

	var indicatorPos fyne.Position
	var indicatorSize fyne.Size
	pad := th.Size(theme.SizeNamePadding)

	switch r.docTabs.location {
	case TabLocationTop:
		indicatorPos = fyne.NewPos(selectedPos.X-scrollOffset.X, r.bar.MinSize().Height)
		indicatorSize = fyne.NewSize(fyne.Min(selectedSize.Width, scrollSize.Width-indicatorPos.X), pad)
	case TabLocationLeading:
		indicatorPos = fyne.NewPos(r.bar.MinSize().Width, selectedPos.Y-scrollOffset.Y)
		indicatorSize = fyne.NewSize(pad, fyne.Min(selectedSize.Height, scrollSize.Height-indicatorPos.Y))
	case TabLocationBottom:
		indicatorPos = fyne.NewPos(selectedPos.X-scrollOffset.X, r.bar.Position().Y-pad)
		indicatorSize = fyne.NewSize(fyne.Min(selectedSize.Width, scrollSize.Width-indicatorPos.X), pad)
	case TabLocationTrailing:
		indicatorPos = fyne.NewPos(r.bar.Position().X-pad, selectedPos.Y-scrollOffset.Y)
		indicatorSize = fyne.NewSize(pad, fyne.Min(selectedSize.Height, scrollSize.Height-indicatorPos.Y))
	}

	if indicatorPos.X < 0 {
		indicatorSize.Width = indicatorSize.Width + indicatorPos.X
		indicatorPos.X = 0
	}
	if indicatorPos.Y < 0 {
		indicatorSize.Height = indicatorSize.Height + indicatorPos.Y
		indicatorPos.Y = 0
	}
	if indicatorSize.Width < 0 || indicatorSize.Height < 0 {
		r.indicator.FillColor = color.Transparent
		r.indicator.Refresh()
		return
	}

	r.moveIndicator(indicatorPos, indicatorSize, th, animate)
}

func (r *docTabsRenderer) updateAllTabs() {
	if len(r.docTabs.Items) > 0 {
		r.action.Show()
	} else {
		r.action.Hide()
	}
}

func (r *docTabsRenderer) updateCreateTab() {
	if r.docTabs.CreateTab != nil {
		r.create.SetIcon(theme.ContentAddIcon())
		r.create.Show()
	} else {
		r.create.Hide()
	}
}

func (r *docTabsRenderer) updateTabs() {
	tabCount := len(r.docTabs.Items)
	r.buildTabButtons(tabCount, r.scroller) // 直接使用容器

	// Set layout of tab bar containing tab buttons and overflow action
	if r.docTabs.location == TabLocationLeading || r.docTabs.location == TabLocationTrailing {
		r.bar.Layout = layout.NewBorderLayout(nil, r.box, nil, nil)
		r.scroller.Layout = layout.NewVBoxLayout() // 垂直布局
	} else {
		r.bar.Layout = layout.NewBorderLayout(nil, nil, nil, r.box)
		r.scroller.Layout = layout.NewHBoxLayout() // 水平布局
	}

	r.bar.Objects = []fyne.CanvasObject{r.scroller, r.box}
	r.bar.Refresh()
}

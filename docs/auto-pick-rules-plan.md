# 自动取色取点规则实现计划

## 目标

实现左侧栏「自动取色」能力：用户点击「自动取色」后进入框选，框选完成后弹出「取消 / 确认」，确认后按当前下拉框规则在选区内生成取色点，并复用现有颜色点列表、图像标记和代码生成流程。

当前下拉框规则为：

1. 随机取点
2. 轮廓取点
3. 高亮取点
4. 高饱和取点
5. 颜色分类轮廓
6. 颜色分类随机

## 当前代码入口

主要文件：`main.go`

已存在的相关结构：

- `pickModeOptions`：自动取色模式选项。
- `pickModeSelect`：左侧栏模式下拉框。
- `pickCountEntry`：取色个数输入框，当前默认形如 `20个`。
- `ImageViewer.SetRangeSelectMode`：进入/退出框选模式。
- `ImageViewer.MouseDown / MouseUp`：当前框选结束后会调用 `AddRect` 并退出范围模式。
- `ImageViewer.AddRect`：保存框选矩形并更新 `rectCoordEntry`。
- `ImageViewer.AddPoint`：添加单点标记，并写入 `colorPoints`。
- `addColorPointToList`：写入右侧颜色点列表并刷新 UI。

当前缺口：

- 「自动取色」按钮目前是占位按钮，没有行为。
- 范围框选完成后没有区分“普通范围选择”和“自动取色范围选择”。
- 缺少批量添加点函数；如果循环调用 `AddPoint`，会重复刷新列表和图像。
- 缺少各取点规则的纯算法函数和测试。

## 总体设计

### 1. UI 流程

建议流程：

1. 点击「自动取色」。
2. 校验 `imageViewer != nil` 且 `imageViewer.image != nil`。
3. 读取并保存当前模式：`pickModeSelect.Selected`。
4. 解析取点个数：`pickCountEntry.Text`，例如 `20个` -> `20`。
5. 进入“自动取色框选模式”。
6. 用户拖拽选区。
7. 鼠标松开后生成选区矩形，弹出确认框：
   - 标题：`自动取色`
   - 内容：`确认在区域 x1,y1,x2,y2 内按「xxx」生成 N 个取色点？`
   - 按钮：`确认 / 取消`
8. 点击确认：执行对应取点规则，批量添加点。
9. 点击取消：不新增点，建议保留框选矩形，方便用户重新确认或手动使用该范围。

### 2. 框选回调改造

不要把自动取色逻辑硬塞进普通 `rangeBtn`。建议给 `ImageViewer` 增加一次性框选回调：

```go
type ImageViewer struct {
    // existing fields...
    onRangeSelected func(rect image.Rectangle)
}
```

新增方法：

```go
func (v *ImageViewer) SetRangeSelectModeWithCallback(callback func(image.Rectangle)) {
    v.onRangeSelected = callback
    v.SetRangeSelectMode(true)
}
```

`MouseUp` 在 `imageDragRange` 分支中：

1. 归一化矩形坐标。
2. 调用 `AddRect` 保留现有范围显示行为。
3. 退出范围模式。
4. 如果 `onRangeSelected != nil`，取出回调、清空字段、调用回调。

注意：回调必须是一次性的，避免下一次普通范围框选误触发自动取色。

### 3. 批量添加点

新增批量函数，避免每个点都刷新一次：

```go
func (v *ImageViewer) AddPoints(points []image.Point) {
    // 逐点读取原图颜色，追加 markPoints 和 colorPoints
    // 最后只 refreshColorList 一次，只 v.Refresh 一次
}
```

实现时建议拆一个内部 helper：

```go
func colorPointFromImage(img image.Image, p image.Point) (mark MarkPoint, item ColorPoint, ok bool)
```

验收要求：批量添加 20 个点时，右侧列表新增 20 行，图像只最终刷新一次。

## 公共算法接口

建议把算法先做成纯函数，便于测试：

```go
type autoPickRequest struct {
    Image       image.Image
    Rect        image.Rectangle
    Count       int
    Mode        string
    MinDistance int
}

func autoPickPoints(req autoPickRequest) []image.Point
```

内部按模式分发：

```go
switch req.Mode {
case "随机取点":
case "轮廓取点":
case "高亮取点":
case "高饱和取点":
case "颜色分类轮廓":
case "颜色分类随机":
}
```

公共 helper：

- `parsePickCount(text string) int`：解析 `20个`、`20`、空值。
- `normalizePickRect(img image.Image, rect image.Rectangle) image.Rectangle`：限制选区在图片范围内，并保证 `Min < Max`。
- `pointDistanceOK(points []image.Point, p image.Point, minDistance int) bool`：点间距过滤。
- `pickTopCandidates(candidates []autoPickCandidate, count, minDistance int) []image.Point`：按分数降序取点，带距离过滤。
- `rgbToHSV` 或 `rgbSaturation`：给高饱和/颜色分类使用。
- `luma`：灰度亮度，推荐 `0.299*r + 0.587*g + 0.114*b`。

候选结构：

```go
type autoPickCandidate struct {
    Point image.Point
    Score float64
    Class int
}
```

## 各取点规则

### 1. 随机取点

用途：快速覆盖选区，适合目标颜色分布比较均匀的区域。

规则：

1. 在选区内随机生成候选点。
2. 使用最小距离过滤，避免所有点挤在一起。
3. 如果随机尝试次数耗尽仍不足 N 个，降级为网格扫描补齐。

建议参数：

- 最大尝试次数：`count * 80`。
- `minDistance` 默认：根据选区面积和 count 自动估算，最小 3 像素。

测试：

- 固定随机种子，保证测试稳定。
- 返回点都在 rect 内。
- 数量不超过 count。
- 点不重复。

### 2. 轮廓取点

用途：抓 UI 边框、图标边缘、按钮边缘等结构明显的位置。

规则：

1. 在选区内计算灰度图。
2. 使用 Sobel 或简单梯度计算边缘强度。
3. 将边缘强度作为候选分数。
4. 按分数从高到低选点，并做最小距离过滤。

简化 Sobel：

- `gx = right - left`
- `gy = bottom - top`
- `score = abs(gx) + abs(gy)`

先用简化梯度即可，不要一开始引入复杂图像处理依赖。

测试：

- 构造黑白分界图，期望点落在分界线附近。
- 构造纯色图，结果可为空或少量降级点，但不能 panic。

### 3. 高亮取点

用途：抓亮色文字、发光按钮、高亮提示等。

规则：

1. 对选区每个像素计算亮度 `luma`。
2. 计算局部对比度，例如与上下左右邻居平均亮度差。
3. 分数建议：`score = luma * 0.7 + localContrast * 0.3`。
4. 选取得分最高的点，并做最小距离过滤。

注意：

- 只取“亮”不够，纯白背景会被大量选中；加入局部对比度可以更偏向文字/按钮边缘。
- 如果目标是亮色块内部，可降低对比度权重。

测试：

- 深色背景上一块亮色区域，点应落在亮色区域内。
- 白底黑字场景不应大量选纯白背景；如果效果不好，后续调权重。

### 4. 高饱和取点

用途：抓彩色按钮、图标、红点、蓝色高亮等高彩度元素。

规则：

1. 对每个像素计算饱和度。
2. 可加入局部对比度，避免纯彩色大背景被过度选中。
3. 分数建议：`score = saturation * 0.75 + localContrast * 0.25`。
4. 过滤低透明度像素；当前截图多为不透明，但 helper 应处理 alpha。
5. 按分数选点并做距离过滤。

测试：

- 灰色背景 + 红色小块，点应优先落在红色块。
- 高亮白色但低饱和区域不应优先于彩色区域。

### 5. 颜色分类轮廓

用途：目标区域有多个主色，希望每类颜色的边界都能取到代表点。

建议不要一开始做 k-means，先用颜色桶分类，简单稳定：

1. 对 RGB 降采样分桶，例如每通道除以 32：`r/32, g/32, b/32`。
2. 统计每个桶的像素数量。
3. 过滤过小桶，保留 top K 个主色桶。
4. 在每个主色桶内找边界像素：上下左右邻居属于不同桶即为边界。
5. 边界候选分数：梯度强度 + 桶数量权重。
6. 按颜色桶分配取点名额，至少每个主色桶 1 个，再按候选分数选点。

测试：

- 构造红/蓝/绿三块区域，结果应覆盖多个颜色类别。
- 点应偏向颜色块边界，而不是全部在纯色内部。

### 6. 颜色分类随机

用途：目标区域颜色多，但不一定需要边界，只希望每种主色都取到样本。

规则：

1. 复用颜色桶分类。
2. 保留 top K 主色桶。
3. 按桶像素占比给每个桶分配点数，至少 1 个。
4. 在每个桶内部随机取点，做全局最小距离过滤。
5. 不足时从剩余大桶补齐。

测试：

- 三色区域，返回点至少覆盖两个或三个主色桶。
- 固定随机种子，测试稳定。

## 实现阶段拆分

### 阶段 1：UI 框选确认流程

目标：点击「自动取色」可以进入框选，框选后弹确认框，但确认后暂时只弹出信息，不生成点。

改动：

- 给自动取色按钮绑定真实 handler。
- 给 `ImageViewer` 增加一次性 range selected callback。
- 解析 `pickCountEntry`。
- 确认框展示模式、数量、范围。

验收：

- 普通「范围」按钮行为不变。
- 自动取色框选后弹出确认框。
- 取消不新增点。
- `go test .` 通过。

### 阶段 2：批量添加点与随机取点

目标：先打通完整链路，确认后按「随机取点」新增点。

改动：

- 新增 `autoPickPoints` 框架。
- 新增 `pickRandomPoints`。
- 新增 `ImageViewer.AddPoints`。
- 自动取色确认后调用算法并批量添加。

验收：

- 随机取点能新增指定数量附近的点。
- 点全部在选区内。
- 右侧列表和图像标记同步。
- `go test .` 通过。

### 阶段 3：轮廓取点

目标：实现边缘/轮廓优先取点。

改动：

- 新增灰度和梯度 helper。
- 新增 `pickContourPoints`。
- 添加纯函数测试。

验收：

- 黑白分界图中，取点集中在分界附近。
- 纯色图不报错。
- `go test .` 通过。

### 阶段 4：高亮取点与高饱和取点

目标：实现亮度优先和彩度优先两类打分取点。

改动：

- 新增亮度、局部对比度、饱和度 helper。
- 新增 `pickHighlightPoints`。
- 新增 `pickHighSaturationPoints`。

验收：

- 高亮取点优先选择亮色/高对比区域。
- 高饱和取点优先选择彩色高饱和区域。
- `go test .` 通过。

### 阶段 5：颜色分类随机

目标：实现按主色分桶的随机取点。

改动：

- 新增颜色桶统计 helper。
- 新增按桶分配点数 helper。
- 新增 `pickColorClassRandomPoints`。

验收：

- 多色区域能覆盖多个主色。
- 固定随机种子后测试稳定。
- `go test .` 通过。

### 阶段 6：颜色分类轮廓

目标：实现按颜色分类后的边界取点。

改动：

- 复用颜色桶 helper。
- 新增桶边界检测。
- 新增 `pickColorClassContourPoints`。

验收：

- 多色块图中，点覆盖多个颜色边界。
- 不同颜色块之间的边界得分更高。
- `go test .` 通过。

### 阶段 7：体验打磨

目标：让功能更适合实际使用。

可选改进：

- 确认框里显示“预计生成 N 个点，实际可能因距离过滤略少”。
- 如果生成 0 个点，明确提示原因。
- 添加“清空原点后取色 / 追加取色”的选择；默认建议追加，避免误删。
- 对大选区显示短暂 loading，避免 UI 卡顿。
- 在帮助文案里补充自动取色说明。

## 推荐文件组织

短期保持单文件最小改动，避免大重构：

- `main.go`：UI 流程、ImageViewer 接入、算法 helper。
- `main_test.go`：纯算法测试。

如果后续算法代码明显变多，再拆分：

- `auto_pick.go`
- `auto_pick_test.go`

拆分条件：自动取色相关代码超过约 300 行，或 `main.go` 中相关 helper 已经影响阅读。

## 验证清单

每个阶段完成后至少执行：

```bash
go test .
```

手动 smoke test：

1. 加载图片。
2. 点击「自动取色」。
3. 框选区域。
4. 点击取消：不新增点。
5. 再次自动取色并点击确认：新增点。
6. 切换每个取点规则，确认不会 panic。
7. 点击「复制代码」，确认生成代码仍使用新增点。

## 关键约束

- 不引入大型图像处理依赖；优先标准库和简单算法。
- 取点算法必须是纯函数优先，方便测试。
- UI 逻辑只负责收集输入、调用算法、展示结果。
- 不要改变现有手动取点、范围按钮、代码生成行为。
- 每个规则独立实现、独立测试、独立提交，方便回滚。

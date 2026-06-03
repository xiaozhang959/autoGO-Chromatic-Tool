# 节点工具选择器区域改造方案（Plan）

## 1. 目标

将“节点工具”标签页底部当前简单的 XPath 选择器区域，改造成面向 AutoGo `uiacc` 的选择器生成区，参考用户截图中的交互能力：

- 属性列表中可勾选参与定位的节点属性。
- 每个属性可选择对应的 AutoGo `uiacc` 查找函数。
- 可选择最终执行函数，例如 `FindOnce`、`Find`、`WaitFor`。
- 可自由设置代码/参数输出格式。
- 支持复制参数、复制完整选择器/代码。
- 支持查找测试，验证当前组合在已抓取节点快照中的匹配数量。
- 保持现有节点抓取、节点树、属性展示、搜索、截图高亮逻辑不变。

本阶段只产出方案文档，后续再按文档分步改代码。

## 2. 已确认的现状

参考文件：

- `node_tool.go`
- `node_tool_test.go`
- `docs/autogo文档.md` 中 Android/iOS `uiacc` 章节

当前实现要点：

1. `AndroidNodeTool` 已包含节点树、属性列表、选择器输入框和按钮：
   - `selectorEntry`
   - `copySelectorBtn`
   - `copyAttrsBtn`
   - `selectAllBtn`
   - `clearSelectedBtn`
   - `testSelectorBtn`
2. 属性行结构 `androidNodeAttrRow` 当前只有：
   - `Selected`
   - `Name`
   - `Value`
   - `Finder`
3. `buildAndroidNodeAttrRows` 目前从节点 XML 属性生成行：
   - `depth`
   - `index`
   - `class`
   - `package`
   - `text`
   - `desc`
   - `id`
   - `bounds`
   - `checkable`
   - `checked`
   - `clickable`
   - `enabled`
   - `focusable`
   - `focused`
   - `scrollable`
   - `longClickable`
   - `password`
   - `selected`
4. 默认选择规则：优先勾选 `id`、`desc`、`text`；没有稳定属性时回退勾选 `class`。
5. `selectedXPath()` 当前只生成 XPath：
   - 示例：`//*[@resource-id='xxx'][@text='xxx']`
6. `testSelectedAttrs()` 当前只在本次抓取的 XML 快照里做等值匹配，并输出匹配数量/是否唯一。
7. 当前缺失能力：
   - 不能选择 AutoGo `uiacc` 函数。
   - 不能选择 `Contains`、`StartsWith`、`EndsWith`、`Matches` 等匹配方式。
   - 不能生成 AutoGo 代码链。
   - 不能编辑输出格式。
   - “复制属性”和“复制选择器”不能满足“复制参数/生成代码”的需求。

## 3. AutoGo `uiacc` 函数映射

### 3.1 字符串类属性

| 工具属性 | 节点 XML 属性 | AutoGo 方法候选 |
|---|---|---|
| `text` | `text` | `Text` / `TextContains` / `TextStartsWith` / `TextEndsWith` / `TextMatches` |
| `desc` | `content-desc` | `Desc` / `DescContains` / `DescStartsWith` / `DescEndsWith` / `DescMatches` |
| `id` | `resource-id` | `Id` / `IdContains` / `IdStartsWith` / `IdEndsWith` / `IdMatches` |
| `class` | `class` | `ClassName` / `ClassNameContains` / `ClassNameStartsWith` / `ClassNameEndsWith` / `ClassNameMatches` |
| `package` | `package` | `PackageName` / `PackageNameContains` / `PackageNameStartsWith` / `PackageNameEndsWith` / `PackageNameMatches` |

默认方法建议：

- `text` -> `Text`
- `desc` -> `Desc`
- `id` -> `Id`
- `class` -> `ClassName`
- `package` -> `PackageName`

### 3.2 范围类属性

| 工具属性 | 节点 XML 属性 | AutoGo 方法候选 |
|---|---|---|
| `bounds` | `bounds` | `Bounds` / `BoundsInside` / `BoundsContains` |

`bounds` 值格式来自 uiautomator XML，例如 `[32,168][688,1065]`，生成代码时转换为四个整数参数：

```go
acc.Bounds(32, 168, 688, 1065)
```

### 3.3 数值类属性

| 工具属性 | 节点 XML 属性 | AutoGo 方法候选 |
|---|---|---|
| `index` | `index` | `Index` |
| `drawingOrder` | `drawing-order` | `DrawingOrder` |
| `depth` | 内部解析深度 | 不建议默认参与 `uiacc` 生成 |

`depth` 是工具解析树结构得到的辅助值，不是文档里稳定的 `uiacc` 选择器方法，默认保留展示但不允许生成选择器。

### 3.4 布尔类属性

| 工具属性 | 节点 XML 属性 | AutoGo 方法候选 |
|---|---|---|
| `clickable` | `clickable` | `Clickable` |
| `longClickable` | `long-clickable` | `LongClickable` |
| `checkable` | `checkable` | `Checkable` |
| `checked` | `checked` | `Checked` |
| `selected` | `selected` | `Selected` |
| `enabled` | `enabled` | `Enabled` |
| `scrollable` | `scrollable` | `Scrollable` |
| `editable` | `editable` | `Editable` |
| `focusable` | `focusable` | `Focusable` |
| `focused` | `focused` | `Focused` |
| `password` | `password` | `Password` |
| `visible` | `visible` | `Visible` |
| `multiLine` | `multi-line` | `MultiLine` |
| `dismissable` | `dismissable` | `Dismissable` |
| `contextClickable` | `context-clickable` | `ContextClickable` |

当前抓取列表里已有一部分布尔属性，后续实现时优先接入已能从 XML 获取到的属性；不存在值的属性不展示或不可勾选。

### 3.5 终止函数

选择器链后可选择最终函数：

| UI 显示 | 生成示例 | 说明 |
|---|---|---|
| `FindOnce` | `acc.Text("登录").FindOnce()` | 查找第一个节点 |
| `Find` | `acc.TextContains("登录").Find()` | 查找所有节点 |
| `WaitFor` | `acc.Text("登录").WaitFor(3000)` | 等待节点出现，需超时参数 |

`Click(text)` 不是链式终止函数，而是快捷文本点击。为保持实现清晰，首版不放入同一个函数下拉中；如后续需要，可单独增加“生成点击代码”预设。

## 4. 推荐 UI 结构

保持上半部分节点树和属性明细区域不变，只替换/增强底部选择器区域。

### 4.1 属性列表

属性列表保留当前四列结构，但把“函数”列从只读文本改成可选择控件：

| 列 | 行为 |
|---|---|
| 勾选 | 点击整行或复选框切换是否参与生成 |
| 属性 | 显示 `text`、`desc`、`id` 等属性名 |
| 值 | 显示属性值，长文本中间截断 |
| 查找函数 | 每行一个下拉框，候选项来自属性类型映射 |

### 4.2 底部控制区

参考截图，建议从上到下布局：

1. 函数行
   - 标签：`函数`
   - 下拉框：`FindOnce` / `Find` / `WaitFor`
   - 可选：`使用正则` 复选框
     - 勾选后，把已选字符串属性的默认方法切换为 `*Matches`。
     - 取消后恢复为精确匹配方法。
2. 格式行
   - 标签：`格式`
   - 输入框：自定义输出模板。
   - 按钮：`复制参数`
3. 操作按钮行
   - `全部勾选`
   - `清除全选`
   - `查找测试`
   - `生成代码`
4. 输出区
   - 多行只读/可复制文本，显示按当前设置生成的结果。
   - 保留 `复制选择器` 或改名为 `复制代码`。

## 5. 输出格式设计

### 5.1 参数片段

“复制参数”输出链式参数片段，不包含 `acc` 和终止函数：

```go
.Text("懒人高级版").ClassName("android.widget.TextView")
```

或为了更适合粘贴到已有 `acc` 后面，允许去掉开头点的格式：

```go
Text("懒人高级版").ClassName("android.widget.TextView")
```

首版建议固定为带点链式片段，后续用格式模板自由调整。

### 5.2 默认完整代码

默认“生成代码”输出：

```go
acc := uiacc.New(0)
obj := acc.Text("懒人高级版").FindOnce()
```

`Find` 示例：

```go
acc := uiacc.New(0)
objs := acc.TextContains("登录").Find()
```

`WaitFor` 示例：

```go
acc := uiacc.New(0)
obj := acc.Text("登录").WaitFor(3000)
```

### 5.3 格式模板占位符

格式输入框支持以下占位符：

| 占位符 | 含义 |
|---|---|
| `{displayId}` | 屏幕 ID，默认 `0` |
| `{chain}` | 选择器链，例如 `.Text("登录").Id("demo:id/login")` |
| `{params}` | 同 `{chain}`，用于兼容“复制参数”命名 |
| `{function}` | 终止函数名，例如 `FindOnce` |
| `{call}` | 终止调用，例如 `FindOnce()` 或 `WaitFor(3000)` |
| `{timeout}` | `WaitFor` 超时时间，默认 `3000` |

默认模板：

```go
acc := uiacc.New({displayId})
obj := acc{chain}.{call}
```

当函数为 `Find` 时，变量名可保持 `obj`，避免模板逻辑过度复杂；如需要 `objs`，用户可手动调整格式。

## 6. 查找测试设计

首版查找测试仍基于本次抓取的 XML 快照，不直接执行 AutoGo 脚本。

原因：

- 速度快，不依赖额外运行环境。
- 不引入远程脚本执行、临时文件、权限差异等复杂度。
- 当前工具已有快照匹配基础，只需把等值匹配扩展为函数语义匹配。

匹配规则：

| 方法类型 | 本地匹配规则 |
|---|---|
| `Text` / `Desc` / `Id` 等精确方法 | 字符串等于 |
| `*Contains` | 字符串包含 |
| `*StartsWith` | 字符串前缀匹配 |
| `*EndsWith` | 字符串后缀匹配 |
| `*Matches` | Go 正则匹配 |
| 布尔方法 | `true` / `false` 字符串转 bool 后比较 |
| `Index` / `DrawingOrder` | int 比较 |
| `Bounds` | 四边界完全一致 |
| `BoundsInside` | 节点 bounds 在指定区域内 |
| `BoundsContains` | 节点 bounds 包含指定区域 |

测试结果显示：

```text
查找测试: 匹配 1 个节点 · 唯一
```

若正则无效，直接显示错误，不吞掉异常：

```text
查找测试失败: textMatches 正则无效: ...
```

## 7. 数据模型建议

将 `androidNodeAttrRow` 扩展为更明确的结构，避免继续把 XPath finder 和 AutoGo 方法混在一起：

```go
type androidNodeAttrKind int

const (
    androidNodeAttrString androidNodeAttrKind = iota
    androidNodeAttrBool
    androidNodeAttrInt
    androidNodeAttrBounds
    androidNodeAttrUnsupported
)

type androidNodeAttrRow struct {
    Selected bool
    Name     string
    Value    string

    XMLAttr string
    Kind    androidNodeAttrKind
    Method  string
    Methods []string
}
```

原则：

- `XMLAttr` 用于从 XML 快照中取值。
- `Method` 用于生成 AutoGo `uiacc` 代码。
- `Methods` 用于 UI 下拉候选项。
- 不再使用 `Finder` 同时承担多种含义。

为降低风险，也可以先保留 `Finder` 字段并新增 `Method/XMLAttr/Kind`，待功能稳定后再清理旧字段。

## 8. 实施步骤

### Step 1：增加选择器模型和纯函数测试

目标文件：

- `node_tool.go`
- `node_tool_test.go`

内容：

1. 增加属性元数据表：属性名、XML 属性名、类型、AutoGo 方法候选。
2. 修改/扩展 `buildAndroidNodeAttrRows`，让每行带默认 `Method` 和候选方法。
3. 增加代码生成纯函数：
   - 生成参数链 `{chain}`。
   - 生成终止调用 `{call}`。
   - 按模板生成完整代码。
4. 增加本地匹配纯函数：
   - 支持 exact/contains/startsWith/endsWith/matches。
   - 支持 bool/int/bounds。
5. 补充单元测试。

优先验证命令：

```powershell
go test ./... -run "TestBuildAndroidNodeAttrRows|TestSelected|TestAndroidNode"
```

### Step 2：替换底部 UI 控件

目标文件：

- `node_tool.go`

内容：

1. 属性行“函数”列改为 `widget.Select`。
2. 新增终止函数下拉框。
3. 新增格式输入框。
4. 新增或重命名按钮：
   - `复制参数`
   - `复制代码`
   - `生成代码`
5. 保持现有按钮行为可迁移：
   - `全部勾选`
   - `清除全选`
   - `查找测试`

### Step 3：接入生成、复制、测试行为

内容：

1. 选择属性、切换行函数、切换终止函数、编辑格式后，实时刷新输出区。
2. `复制参数` 复制 `{chain}`。
3. `复制代码` 复制当前输出区内容。
4. `查找测试` 使用新匹配规则，报告匹配数量和唯一性。
5. 无选中属性时明确提示，不生成假成功。

### Step 4：验证

按顺序运行：

```powershell
go test ./... -run "TestBuildAndroidNodeAttrRows|TestSelected|TestAndroidNode"
go test ./...
```

如 UI 运行环境可用，再做一次手动冒烟：

1. 打开节点工具。
2. 抓取节点。
3. 选择包含 `text/id/desc` 的节点。
4. 切换查找函数。
5. 复制参数。
6. 生成代码。
7. 查找测试。

## 9. 成功标准

本次改造完成后应满足：

- 选择节点后，属性表自动显示可用 AutoGo 查找函数。
- 勾选属性后，输出区生成 AutoGo `uiacc` 链式代码，而不是仅生成 XPath。
- 字符串属性可切换精确、包含、前缀、后缀、正则匹配。
- `FindOnce` / `Find` / `WaitFor` 至少三个终止函数可选。
- 格式输入框可控制最终输出文本。
- 复制参数和复制代码可用。
- 查找测试能根据所选函数语义给出匹配数量。
- 原有节点抓取、搜索、节点树选择、高亮不退化。
- 单元测试覆盖核心生成和匹配逻辑。

## 10. 明确不做的内容

首版不做以下内容，避免范围失控：

- 不执行真实 AutoGo 脚本验证，只做本地快照测试。
- 不重做整个节点工具布局。
- 不改图色工具标签页。
- 不接入 iOS 节点抓取；当前工具实现仍以 Android uiautomator XML 为基础。
- 不新增网络调用、遥测或额外大型依赖。
- 不删除 `docs/autogo文档.md`，该文件当前为未跟踪参考文档。

## 11. 风险与处理

| 风险 | 处理 |
|---|---|
| Fyne 表格中每行下拉控件布局可能拥挤 | 保持当前紧凑布局，必要时只让函数列显示短名称 |
| 正则匹配失败 | 不吞错，状态栏直接显示正则错误 |
| `bounds` 参数解析失败 | 该行不可生成，状态栏提示格式异常 |
| Android XML 属性名和 AutoGo 方法名不一致 | 使用元数据表统一维护映射 |
| `docs/autogo文档.md` 未跟踪 | 只读取作为参考，不纳入本次提交 |

# AutoGo 文档整合版

本文档由 https://autogo.cc 官方文档站点整理整合而成，适合直接提供给 AI 作为上下文使用。

整理说明：
1. 保留官方 README、更新日志、Android API、iOS API 的主要 Markdown 内容。
2. 统一章节结构，便于检索、引用和向量化。
3. 图片资源链接未全部内联，重点保留文字说明与接口文档。
4. 内容来源于抓取时官网可访问页面，如官网后续更新，请重新同步。

## 文档目录

### 前言
- 简介
- 更新日志

### Android API
- app - 应用 (API/app.md)
- device - 设备 (API/device.md)
- files - 文件系统 (API/files.md)
- uiacc - 节点操作 (API/uiacc.md)
- https - HTTP网络请求 (API/https.md)
- ime - 输入法 (API/ime.md)
- motion - 动作 (API/motion.md)
- images - 图像处理 (API/images.md)
- opencv - 图像处理 (API/opencv.md)
- dotocr - 点阵OCR (API/dotocr.md)
- media - 多媒体 (API/media.md)
- rhino - JS脚本引擎 (API/rhino.md)
- storages - 本地存储 (API/storages.md)
- plugin - 外部插件 (API/plugin.md)
- system - 系统 (API/system.md)
- utils - 工具函数 (API/utils.md)
- ppocr - 飞浆OCR (API/ppocr.md)
- yolo - 目标检测 (API/yolo.md)
- apkctl - APK外壳控制 (API/apkctl.md)
- imgui - 界面绘制 (API/imgui.md)
- console - 控制台 (API/console.md)
- hud - 悬浮显示 (API/hud.md)
- vdisplay - 虚拟屏幕 (API/vdisplay.md)

### iOS API
- app - 应用 (ios/app.md)
- device - 设备 (ios/device.md)
- files - 文件系统 (ios/files.md)
- uiacc - 节点操作 (ios/uiacc.md)
- https - HTTP网络请求 (ios/https.md)
- ime - 输入法 (ios/ime.md)
- motion - 动作 (ios/motion.md)
- images - 图像处理 (ios/images.md)
- opencv - 图像处理 (ios/opencv.md)
- dotocr - 点阵OCR (ios/dotocr.md)
- storages - 本地存储 (ios/storages.md)
- system - 系统 (ios/system.md)
- utils - 工具函数 (ios/utils.md)
- ppocr - 飞浆OCR (ios/ppocr.md)
- yolo - 目标检测 (ios/yolo.md)
- imgui - 界面绘制 (ios/imgui.md)
- console - 控制台 (ios/console.md)
- hud - 悬浮显示 (ios/hud.md)

---

## 简介

# 简介

**AutoGo** 是基于 Go 语言的 **Android 与 iOS** 自动化框架。**Android** 侧既可将脚本编译为 **原生二进制**，通过 ADB 或具备权限的 shell 在设备上直接执行，也可 **打包为 APK** 安装后以应用形态运行。**iOS** 侧在插件与工程流程下开发与调试，并支持 **打包为 IPA** 安装使用。

### QQ交流群:753399754&nbsp;&nbsp;&nbsp;<a href="https://qm.qq.com/q/mHAHFlgtXi" target="_blank" style="font-size:15px;">点击加入</a>

### 为什么选择 AutoGo？

- **双端支持**：同一套 Go 语言与相似 API 设计，按目标平台选用 SDK 与文档即可；部分模块仅某一端提供，以对应包文档为准。
- **Android 运行形态灵活**：二进制直跑适合 ADB / shell 场景；需要独立 App 体验时可 **打包 APK**，与插件文档中的外壳能力配合使用。
- **高安全性**：脚本经编译后以二进制或安装包（APK / IPA 等）形态交付，降低被直接篡改与逆向的风险。
- **可扩展性强**：可使用丰富的内置 API 以及各类 Go 开源库，定制自动化流程。
- **集成与兼容性（Android）**：在具备 shell 权限时，可作为独立工具或嵌入其他场景在 Android 上运行。
- **灵活的开发模式**：充分利用 Go 的并发与工程化能力，快速实现复杂自动化任务。
- **丰富的功能模块**：从应用控制到系统能力均有覆盖；Android 与 iOS 模块列表见侧栏，**iOS 环境要求、调试器与打包方式以 [更新日志](changelog.md) 为准**。

### 适用场景

- **应用自动化测试**：在 Android / iOS 上编写脚本，做功能、性能与兼容性验证。
- **跨应用操作**：在多应用协作、自动化回归等场景下驱动界面与系统能力（能力边界因平台而异）。
- **高安全性要求的自动化**：需要降低脚本被轻易破解或篡改风险的场景。
- **Android 深度集成**：可嵌入具备 shell 权限的 Android 使用场景，为应用或工具提供自动化能力。

### 安装指南

**1. 安装 IDE 开发工具：推荐使用 [GoLand](https://www.jetbrains.com/go/download/) 或 [IDEA](https://www.jetbrains.com/idea/download/)**

**2. 安装 [AutoGo](http://jgw.52ailin.com:7001/files/AutoGo/AutoGo-JetBrains-1.0.9.zip) 插件**

![alt text](img/1.png)

注意新版本的goland 不用安装中文插件 Customize 中可以直接设置

![alt text](img/2.png)

改了之后要求我们重启 我们点击 restart 重启即可

![alt text](img/3.png)

重启之后 我们可以看到 回到了熟悉的中文界面 点击插件管理 点击设置 点击 从磁盘安装插件

![alt text](img/4.png)

搜索 AutoGo 点击 安装 即可

![alt text](img/5.png)

重启之后 我们可以看到 插件已经安装成功了

**3. 初始化项目**

以下截图以 **Android 项目** 为例（含 **adb 路径**）。若目标为 **iOS**，请在插件中将项目初始化为 **iOS**，再按 [更新日志](changelog.md) 与侧栏 **iOS 文档** 连接设备、同步文件与打包；系统版本、巨魔、调试器 IPA 等要求以日志中说明为准。

我们新建一个项目 点击创建

![alt text](img/6.png)

![alt text](img/7.png)

点击检查更新下载最新版本SDK

![](img/13.png)

设置adb路径  右键选择插件 点击 其他设置

![](img/8.png)

![](img/9.png)

点击 确定 即可
初始化项目 右键选择插件 点击 初始化项目

![](img/10.png)

看到了 我们的项目已经初始化成功了

![alt text](img/11.png)

**4. 编写第一个项目：**
打开mian.go 文件 CTRL + a 全选 删除掉
粘贴以下代码

```go
package main

import "github.com/Dasongzi1366/AutoGo/motion"

func main() {
	motion.Home(0)  // 回到主界面
}
```

**Android**：连接模拟器或真机，执行 `adb devices` 能识别设备后，使用 **F7** 或点击运行即可。

**iOS**：在插件中完成 iOS 初始化与设备连接后，按插件与调试器说明运行（与 Android 的 adb 流程不同，详见更新日志）。

![alt text](img/12.png)

---

## 更新日志

# 更新日志

## [1.14.0] - 2026-05-28
- 安卓 APK 外壳新增远程调试功能
- 安卓 APK 外壳新增开机自启选项开关
- 安卓 APK 外壳新增查看日志功能
- 安卓 APK 模式运行脚本的情况下默认记录脚本日志到 /sdcard/logs/包名.log

## [1.13.13] - 2026-05-25
- 优化 YOLOv8 对不同 NCNN 导出格式的兼容性
- 快速调试新增对 paho.mqtt.golang gjson 这两个三方库的支持

## [1.13.12] - 2026-05-16
- 新增打包后的应用悬浮球暂停时改变主球为黄色方便区分脚本状态
- 修复 iOS 可能同时出现两个悬浮球的问题

## [1.13.11] - 2026-05-14
- 安卓13+通过网络adb自助激活一次后后续可以免手动激活运行
- 修复部分 miui 设备检查不到网络调试配对服务的问题
- iOS app.Install 方法新增支持安装 deb 需要设备已越狱
- iOS app.GetList 方法新增支持获取到巨魔安装的应用

## [1.13.10] - 2026-05-12
- 修复脚本代码中带有结构体指针可能导致无法快速调试的问题
- 修复安卓中 imgui 在部分设备出现乱码
- 修复安卓14雷电模拟器运行报错的问题
- 修复 iOS 中 imgui 弹出的输入框无法点击确定按钮的问题

## [1.13.9] - 2026-05-11
- 优化 dotocr 精度

## [1.13.8] - 2026-05-09
- imgui 适配安卓16
- 修复脚本在光速虚拟机无法正常运行的问题
- 修复光速虚拟机中 ppocr 无效的问题

## [1.13.7] - 2026-05-08
- 修复 dotocr 找字可能出现找不到的问题

## [1.13.6] - 2026-05-07
- 新增方法 opencv.FindImageFromImage 用于从图片对象中进行模板匹配
- 删除 opencv.FindImage 相关方法的大小缩放参数,新增指定是否是透明图的参数
- 删除 dotocr 识别和查找方法的行间距列间距参数
- 修复 dotocr 粘黏文字无法正常识别的问题

## [1.13.4] - 2026-05-05
- 修复 iOS 云手机中无法点击脚本悬浮球的问题

## [1.13.3] - 2026-04-30
- 新增 apkctl.RegEvent 方法用于注册脚本的暂停恢复和停止事件
- APK 冷启动接口新增启动脚本时传递参数给二进制,例如 "http://192.168.1.100:8989/task?cmd=start 参数1 参数2" 中的两个参数会被脚本接收到可以用 go 的 os.Args 进行读取
- 安卓项目的 Android.toml 内新增一个是否显示悬浮球的配置项,老项目删除这个文件重新初始化就可以看到
- 修复安卓 motion.KeyActionDown 无法正常持续按下的问题

## [1.13.1] - 2026-04-25
- 修复 iOS 部分设备截图少一截的问题
- 减小 iOS 打包的二进制体积
- 安卓 APK 外壳新增内置HiveMQ MQTT库,可以通过js使用

## [1.13.0] - 2026-04-23
- iOS 新增节点操作 uiacc 使用方式看最新文档
- 插件市场AutoGo插件升级到1.0.10后 iOS 项目才可以使用节点助手
- 手机端应用需要使用最新版本重新打包安装否则节点助手获取不到节点

## [1.12.3] - 2026-04-21
- 修复 apkctl.Eval 运行的 js 代码中调用 go.send 在部分设备中可能导致崩溃的问题
- 修复部分安卓11+的设备不会弹出自助激活引导的问题
- 优化 uiacc 偶尔出现节点不刷新问题

## [1.12.2] - 2026-04-20
- 修复 iOS 获取前台应用包名不准确的问题
- 修复 iOS 热更新不生效的问题
- 修复 iOS 多个不同包名的 deb 无法同时安装的问题
- 修复 iOS 的 hud 和 console 尺寸异常的问题
- iOS 越狱模式安装程序主页增加启动脚本和停止脚本的两个按钮用于解决云手机无法点击悬浮球的问题
- iOS 打包的应用增加调试模式的开关,后续调试代码通过任意脚本应用进行调试即可不在额外提供独立的调试器应用
- 修复雷电安卓7模拟器设置输入法失败的问题
- 修复雷电模拟器中 imgui 编辑框点击会导致程序崩溃的问题
- 修复上个版本导致的部分高版本安卓设备 html 的 UI 无法显示的问题

## [1.12.1] - 2026-04-18
- 安卓新增APK外壳控制的go包 apkctl 具体用法看文档
- 安卓新增方法 motion.KeyActionDown motion.KeyActionUp app.SelfPackage utils.InputAlert
- APK外壳新增内置脚本输入法,调用 ime.SetCurrentIME("") 即可把脚本输入法设置为当前输入法,后续脚本输入文字和剪切板操作优先使用脚本输入法
- 安卓imgui编辑框输入方式改为非APK安装模式运行时使用内置虚拟键盘,APK安装运行模式时使用弹出输入框的方式进行输入
- 优化安卓的 utils.Toast 在部分设备上导致崩溃的问题

## [1.11.12] - 2026-04-08
- 修复iOS的imgui在部分设备上出现崩溃问题
- 处理sdk报毒问题

## [1.11.09] - 2026-04-08
- iOS新增方法 app.GetBundlePath device.VpnOn device.VpnOff device.VpnStatus utils.ExecBinary
- 修复艾琳云控无法点击AutoGo编译的ipa脚本的悬浮球,需要重新打包脚本ipa安装后重启手机

## [1.11.07] - 2026-04-05
- 新增代码混淆功能,在AutoGo插件的其他设置选项中开启,需要在插件市场更新AutoGo插件版本到1.0.9
- 修复安卓14+虚拟屏幕预览画面出现重影问题

## [1.11.06] - 2026-04-03
- 修复 YOLOv8 单类/少类模型推理时 label 异常的问题
- 修复部分云手机远程调试时识别不到设备架构的问题

## [1.11.05] - 2026-04-03
- 安卓APK中申请root改为使用libsu库,之前某些无法请求到root的云手机可以更新后重新打包APK进行尝试
- 修复快速调试日志打印的文件名称和行号不准的问题

## [1.11.04] - 2026-04-01
- 快速调试新增支持 embed 嵌入文件
- 修复iOS打包的应用图标异常的问题

## [1.11.03] - 2026-03-31
- 新增方法 images.FindMultiColorsAll 返回多点找色所有符合条件的目标
- 新增方法 opencv.FindImageAll 返回找图所有符合条件的目标
- utils.Toast 修改为支持传入坐标及显示时长,时长未满再次调用会直接覆盖之前的消息

## [1.11.02] - 2026-03-30
- 修复iOS的 app.CurrentPackage 获取前台应用包名失败的问题
- 修复安卓的 media.PlayMP3 多次调用导致程序崩溃的问题
- iOS项目新增编译deb选项用于给iOS13+越狱设备进行安装使用,AutoGo插件需要在插件市场升级至1.0.7版本

## [1.11.01] - 2026-03-26
- iOS新增支持imgui 使用方式和安卓完全一致

## [1.10.06] - 2026-03-25
- 修复ipad设备出现图像异常问题
- 调试器更新到1.0.02版本 下载地址 https://1823847070.v.123pan.cn/1823847070/AutoGo/AutoGo-Debug-1.0.03.ipa

## [1.10.04] - 2026-03-23
- 修复iOS横屏情况下截取的图像也是竖的
- 修复iOS的Console Hud Toast 悬浮球不会随屏幕方向自动旋转的问题

## [1.10.03] - 2026-03-23
- iOS打包后的安装包增加一个开机自启的选项
- 修复iOS快速调试模式脚本工作区异常的问题

## [1.10.02] - 2026-03-22
- 新增适配支持windows打包ios二进制和ipa安装包

## [1.10.01] - 2026-03-22
- 修复部分iOS设备截图缺失一部分的问题,更新后需要重新同步文件
- 调试器最新版本下载地址 https://1823847070.v.123pan.cn/1823847070/AutoGo/AutoGo-Debug-1.0.03.ipa

## [1.10.0] - 2026-03-22
- 新增支持iOS设备
- AutoGo插件版本需要在插件市场更新到1.0.6
- 手机内需要安装巨魔(系统版本在14.0-17.0可以安装),调试器以及编译后的ipa安装包都只支持通过巨魔进行安装
- 插件中初始化项目为iOS之后才可以在连接设备选项中连接iOS设备
- 图色助手需要更新到1.0.6版本才可以截图iOS设备,图色助手在群文件下载
- 调试器下载地址 https://1823847070.v.123pan.cn/1823847070/AutoGo/AutoGo-Debug-1.0.03.ipa

## [1.9.06] - 2026-03-15
- 快速调试新增支持 github.com/gorilla/websocket
- 修复windows上编译可能出现命令行内容过长导致编译失败的问题

## [1.9.05] - 2026-03-09
- ppocr包集成v5模型,具体使用方法参考最新文档

## [1.9.04] - 2026-03-06
- 修复windows环境中快速调试无法显示脚本UI的问题

## [1.9.03] - 2026-03-06
- 快速调试功能从运行单个main.go文件改为支持运行完整项目,同时UI也支持了快速调试
- 编译器指令内容调整(AG二进制),需要在插件市场把AutoGo更新到1.0.5版本才可以正常使用1.9.03版本SDK
- 本次更新完毕后需要同步文件后最新的快速调试功能才能正常运行

## [1.8.27] - 2026-02-26
- dotocr包移除对图灵字库的支持
- AutoGo图色工具新增制作字库功能

## [1.8.26] - 2026-02-26
- 修复APK运行模式下前端localStorage方法保存不住数据的问题

## [1.8.25] - 2026-02-18
- imgui编辑框新增适配此输入法(http://jgw.52ailin.com:7001/files/AutoGo/qqpinyin.apk)进行输入,未安装和启用此输入法的依然使用内置虚拟键盘进行输入

## [1.8.24] - 2026-01-15
- app.Install方法适配xapk安装
- 修复偶尔出现脚本停止后imgui相关窗口依然存在的问题
- 修复上个版本更新导致打包后APK在安卓7无法运行的问题

## [1.8.23] - 2026-01-15
- 修复部分设备APK外壳悬浮窗UI关闭后点击任何地方都没有反应的问题
- 修复部分设备vdisplay创建虚拟屏可能导致崩溃的问题

## [1.8.22] - 2026-01-09
- APK外壳内置无线调试激活方式,需要安卓11及以上版本且连接了wifi才可用
- APK外壳增加强力保活机制,只有主动停止脚本才可以关闭APP
- 修复三星F711设备点击失效问题

## [1.8.18] - 2026-01-08
- 新增方法 vdisplay.SetTouchCallback 用于设置虚拟屏预览窗口的点击回调
- 修复console.Println无法打印百分号问题

## [1.8.17] - 2026-01-06
- uiacc包新增支持操作虚拟屏幕节点,需要安卓版本大于等于11
- 优化uiacc节点偶尔不刷新问题

## [1.8.16] - 2026-01-06
- vdisplay包新增SetTitle方法用于修改预览窗口标题

## [1.8.15] - 2026-01-04
- plugin包新增AssetManager参数类型支持
- imgui虚拟键盘新增按住拖动位置功能

## [1.8.14] - 2026-01-03
- 新增 plugin 外部插件模块,用于加载调用外部APK插件,具体使用方法参考文档和群文件示例代码

## [1.8.13] - 2026-01-02
- 新增方法 app.GetList 获取应用列表
- 新增方法 app.GetName 获取指定包名的应用名称
- 新增方法 imgui.DrawRect 绘制矩形
- 修复console打印内容长度超出组件宽度不会自动换行的问题

## [1.8.12] - 2025-12-26
- imgui 内置虚拟键盘用于处理部分设备点击编辑框不弹出输入法的问题

## [1.8.10] - 2025-12-23
- dotocr各种方法新增一个cutMode参数用于设置切割模式
- 修复部分设备系统自动更新webview导致app被杀死,需要重新打包APK
- 修复imgui的Button组件占用窗口焦点导致背景颜色异常的问题
- 修复vdisplay.Destroy()销毁虚拟屏导致程序崩溃的问题

## [1.8.09] - 2025-12-16
- 新增 dotocr 点阵OCR识别模块,兼容图灵字库和懒人字库
- AutoGo插件适配MacOS英特尔架构设备
- AutoGo插件删除内置Go版本
- AutoGo插件删除导入Go模块和编译Go模块功能

## [1.8.08] - 2025-12-13
- vdisplay 包内的相关方法改动,主要新增窗口预览虚拟屏画面功能
- Android.toml 配置文件内新增 autoRun 配置项表示打开APP后是否自动运行脚本,删除该配置文件后点击插件的初始化项目会出现该文件的最新版本
- 修复部分三星安卓14设备utils.Toast不显示的问题

## [1.8.07] - 2025-12-09
- 新增方法 app.GetVersion 获取应用版本号
- 新增方法 app.GetIcon 获取应用图标
- 新增方法 device.GetNotification 获取当前所有通知消息
- 修复雷电模拟器调用utils.Alert方法导致崩溃
- 修复部分设备打包APK报错

## [1.8.06] - 2025-12-02
- AutoGo插件增加检查更新功能,后续的SDK更新全部依赖此功能
- 修复部分设备imgui中文显示问号的问题
- 修复部分设备imgui.BeginTabBar需要双击才能切换选择夹的问题

---

## Android API 文档

### app - 应用

来源：https://autogo.cc/API/app.md

# app - 应用
---
app模块提供一系列函数，用于使用其他应用、与其他应用交互。例如启动应用、打开文件、发送意图等。

同时提供了方便的进阶函数startActivity和sendBroadcast，用他们可完成app模块没有内置的和其他应用的交互。

以下是 `app` 包中定义的 `IntentOptions` 结构体及其字段说明：

| **字段名**      | **类型**            | **说明**                           |
|------------------|---------------------|------------------------------------|
| `Action`         | `string`           | Intent 的动作，例如 `android.intent.action.VIEW`。 |
| `Type`           | `string`           | Intent 的数据类型，例如 `text/plain` 或 `image/*`。 |
| `Data`           | `string`           | Intent 的数据，例如文件路径或 URI。 |
| `Category`       | `[]string`         | Intent 的类别。                   |
| `PackageName`    | `string`           | 应用包名，用于指定目标应用。       |
| `Extras`         | `map[string]string`| Intent 的额外参数（键值对）。      |
| `Flags`          | `[]string`         | Intent 的标志，例如 `FLAG_ACTIVITY_NEW_TASK`。 |


## CurrentPackage
<hr style="margin: 0;">

获取当前页面的应用包名。

```go
packageName := app.CurrentPackage()
```

## CurrentActivity
<hr style="margin: 0;">

获取当前页面的应用类名。

```go
activityName := app.CurrentActivity()
```

## Launch
<hr style="margin: 0;">

通过应用包名启动应用。

- `packageName` {string} 应用包名，也支持"包名/类名"格式
- `displayId` {int} 屏幕ID

```go
success := app.Launch("com.tencent.mm", 0)
```

## SelfPackage
<hr style="margin: 0;">

获取当前脚本/宿主自身的包名。

```go
pkg := app.SelfPackage()
```

## GetList
<hr style="margin: 0;">

获取手机中所有应用列表。

- `includeSystemApps` {bool} 是否需要包含系统应用

```go
list := app.GetList(true)
```

## GetName
<hr style="margin: 0;">

获取指定包名应用的应用名称。

- `packageName` {string} 应用包名

```go
name := app.GetName("bin.mt.plus")
```

## GetIcon
<hr style="margin: 0;">

获取应用图标。

- `packageName` {string} 应用包名

```go
data := app.GetIcon("com.tencent.mm")
```

## GetVersion
<hr style="margin: 0;">

获取应用版本号。

- `packageName` {string} 应用包名

```go
version := app.GetVersion("com.tencent.mm")
```

## OpenSetting
<hr style="margin: 0;">

打开应用的详情页（设置页）。

- `packageName` {string} 应用包名

```go
success := app.OpenSetting("com.tencent.mm")
```

## ViewFile
<hr style="margin: 0;">

用其他应用查看文件。文件不存在的情况由查看文件的应用处理。

- `path` {string} 文件路径

```go
app.ViewFile("/sdcard/example.txt")
```

## EditFile
<hr style="margin: 0;">

用其他应用编辑文件。文件不存在的情况由编辑文件的应用处理。

- `path` {string} 文件路径

```go
app.EditFile("/sdcard/example.txt")
```

## Uninstall
<hr style="margin: 0;">

卸载应用。

- `packageName` {string} 应用包名

```go
app.Uninstall("com.tencent.mm")
```

## Install
<hr style="margin: 0;">

安装应用（也支持xapk）。

- `path` {string} APK 文件路径

```go
app.Install("/sdcard/app.apk")
```

## IsInstalled
<hr style="margin: 0;">

判断是否已经安装某个应用。

- `packageName` {string} 应用包名

```go
installed := app.IsInstalled("com.tencent.mm")
```

## Clear
<hr style="margin: 0;">

清除应用数据。

- `packageName` {string} 应用包名

```go
app.Clear("com.tencent.mm")
```

## ForceStop
<hr style="margin: 0;">

强制停止应用。

- `packageName` {string} 应用包名

```go
app.ForceStop("com.tencent.mm")
```

## Disable
<hr style="margin: 0;">

禁用应用。

- `packageName` {string} 应用包名

```go
app.Disable("com.tencent.mm")
```

## Enable
<hr style="margin: 0;">

启用应用。

- `packageName` {string} 应用包名

```go
app.Enable("com.tencent.mm")
```

## EnableAccessibility
<hr style="margin: 0;">

启用无障碍服务。

- `packageName` {string} 应用包名

```go
app.EnableAccessibility("org.autojs.autoxjs.v6")
```

## DisableAccessibility
<hr style="margin: 0;">

关闭无障碍服务。

- `packageName` {string} 应用包名

```go
app.DisableAccessibility("org.autojs.autoxjs.v6")
```

## IgnoreBattOpt
<hr style="margin: 0;">

忽略应用电池优化。

- `packageName` {string} 应用包名

```go
app.IgnoreBattOpt("com.tencent.mm")
```

## GetBrowserPackage
<hr style="margin: 0;">

获取系统默认浏览器包名。

```go
packageName := app.GetBrowserPackage()
```

## OpenUrl
<hr style="margin: 0;">

用浏览器打开指定的网址。

- `url` {string} 网站地址

```go
app.OpenUrl("https://example.com")
```

## StartActivity
<hr style="margin: 0;">

根据选项构造一个 Intent，并启动该 Activity。

- `options` {IntentOptions} Intent 选项

```go
app.StartActivity(app.IntentOptions{
	Action: "SEND",
	Type:   "text/plain",
	Data:   "file:///sdcard/1.txt",
})
```

## SendBroadcast
<hr style="margin: 0;">

根据选项构造一个 Intent，并发送广播。

- `options` {IntentOptions} Intent 选项

```go
app.SendBroadcast(options)
```

## StartService
<hr style="margin: 0;">

根据选项构造一个 Intent，并启动服务。

- `options` {IntentOptions} Intent 选项

```go
app.StartService(options)
```

---

### device - 设备

来源：https://autogo.cc/API/device.md

# device - 设备
---
device模块提供了与设备有关的信息与操作，例如获取设备宽高，内存使用率，IMEI，调整设备亮度、音量等。

以下是 `device` 包中定义的设备信息变量：

| **变量名**         | **类型**   | **说明**                              |
|-----------------|----------|-------------------------------------|
| `CpuAbi`        | `string` | 设备的CPU架构，如"arm64-v8a", "x86", "x86_64"等。 |
| `BuildId`       | `string` | 修订版本号，或者诸如"M4-rc20"的标识。             |
| `Broad`         | `string` | 设备的主板型号。                            |
| `Brand`         | `string` | 与产品或硬件相关的厂商品牌，如"Xiaomi", "Huawei"等。 |
| `Device`        | `string` | 设备在工业设计中的名称。                        |
| `Model`         | `string` | 设备型号。                               |
| `Product`       | `string` | 整个产品的名称。                            |
| `Bootloader`    | `string` | 设备 Bootloader 的版本。                  |
| `Hardware`      | `string` | 设备的硬件名称。                            |
| `Fingerprint`   | `string` | 构建 (build) 的唯一标识码。                  |
| `Serial`        | `string` | 硬件序列号。                              |
| `SdkInt`        | `int`    | 安卓系统 API 版本。例如安卓 4.4 的 sdkInt 为 19。 |
| `Incremental`   | `string` | 设备构建的内部版本号。                         |
| `Release`       | `string` | Android 系统版本号。例如 "5.0", "7.1.1"。    |
| `BaseOS`        | `string` | 设备的基础操作系统版本。                        |
| `SecurityPatch` | `string` | 安全补丁程序级别。                           |
| `Codename`      | `string` | 开发代号，例如发行版是"REL"。                   |

## GetDisplayInfo
<hr style="margin: 0;">

获取指定屏幕的分辨率信息。

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
width, height, dpi, rotation := device.GetDisplayInfo(0)
fmt.Printf("屏幕分辨率: %dx%d, DPI: %d, 旋转角度: %d\n", width, height, dpi, rotation)
```

## GetImei
<hr style="margin: 0;">

获取设备的 IMEI 码。

```go
imei := device.GetImei()
```

## GetAndroidId
<hr style="margin: 0;">

获取设备的 Android ID。

```go
androidId := device.GetAndroidId()
```

## GetWifiMac
<hr style="margin: 0;">

获取设备 WIFI 网卡的 MAC 地址。

```go
wifiMac := device.GetWifiMac()
```

## GetWlanMac
<hr style="margin: 0;">

获取设备以太网网卡的 MAC 地址。

```go
wlanMac := device.GetWlanMac()
```

## GetIp
<hr style="margin: 0;">

获取设备局域网 IP 地址。

```go
ip := device.GetIp()
```

## GetNotification
<hr style="margin: 0;">

获取设备当前所有通知消息。

```go
notifications := device.GetNotification()
for _, notification := range notifications {
	fmt.Println("ID:" + notification.Id)
	fmt.Println("包名:" + notification.PackageName)
	fmt.Println("标题:" + notification.Title)
	fmt.Println("内容:" + notification.Text)
	fmt.Println("标签:" + notification.Tag)
	fmt.Println()
}
```

## GetBrightness
<hr style="margin: 0;">

获取当前屏幕亮度值，范围为 0~255。

```go
brightness := device.GetBrightness()
```

## GetBrightnessMode
<hr style="margin: 0;">

获取当前屏幕亮度调节模式，0 为手动调节，1 为自动调节。

```go
mode := device.GetBrightnessMode()
```

## GetMusicVolume
<hr style="margin: 0;">

获取当前媒体音量。

```go
volume := device.GetMusicVolume()
```

## GetNotificationVolume
<hr style="margin: 0;">

获取当前通知音量。

```go
volume := device.GetNotificationVolume()
```

## GetAlarmVolume
<hr style="margin: 0;">

获取当前闹钟音量。

```go
volume := device.GetAlarmVolume()
```

## GetMusicMaxVolume
<hr style="margin: 0;">

获取媒体音量最大值。

```go
maxVolume := device.GetMusicMaxVolume()
```

## GetNotificationMaxVolume
<hr style="margin: 0;">

获取通知音量最大值。

```go
maxVolume := device.GetNotificationMaxVolume()
```

## GetAlarmMaxVolume
<hr style="margin: 0;">

获取闹钟音量最大值。

```go
maxVolume := device.GetAlarmMaxVolume()
```

## SetMusicVolume
<hr style="margin: 0;">

设置媒体音量。

- `volume` {int} 要设置的音量值

```go
device.SetMusicVolume(8)
```

## SetNotificationVolume
<hr style="margin: 0;">

设置通知音量。

- `volume` {int} 要设置的音量值

```go
device.SetNotificationVolume(8)
```

## SetAlarmVolume
<hr style="margin: 0;">

设置闹钟音量。

- `volume` {int} 要设置的音量值

```go
device.SetAlarmVolume(8)
```

## GetBattery
<hr style="margin: 0;">

获取当前电量百分比。

```go
battery := device.GetBattery()
```

## GetBatteryStatus
<hr style="margin: 0;">

获取电池状态。1：没有充电；2：正充电；3：没插充电器；4：不充电； 5：电池充满。

```go
status := device.GetBatteryStatus()
```

## SetBatteryStatus
<hr style="margin: 0;">

模拟设置电池状态。1：没有充电；2：正充电；5：电池充满。

- `value` {int} 要设置的电池状态

```go
device.SetBatteryStatus(2)
```

## SetBatteryLevel
<hr style="margin: 0;">

模拟设置电池电量百分比，范围 0-100。

- `value` {int} 要设置的电池电量百分比

```go
device.SetBatteryLevel(75)
```

## GetTotalMem
<hr style="margin: 0;">

获取设备总内存，单位 KB。

```go
totalMem := device.GetTotalMem()
```

## GetAvailMem
<hr style="margin: 0;">

获取设备当前可用内存，单位 KB。

```go
availMem := device.GetAvailMem()
```

## IsScreenOn
<hr style="margin: 0;">

判断屏幕是否点亮状态。

```go
isOn := device.IsScreenOn()
```

## IsScreenUnlock
<hr style="margin: 0;">

判断屏幕是否已解锁。

```go
isUnlock := device.IsScreenUnlock()
```

## SetDisplayPower
<hr style="margin: 0;">

设置屏幕电源模式，不影响脚本运行。

- `on` {bool} 是否点亮。

```go
device.SetDisplayPower(false)//熄屏挂机
```

## WakeUp
<hr style="margin: 0;">

唤醒设备，包括唤醒 CPU、屏幕等，可以用来点亮屏幕。

```go
device.WakeUp()
```

## Reboot
<hr style="margin: 0;">

重启设备。

```go
device.Reboot()
```

## KeepScreenOn
<hr style="margin: 0;">

保持屏幕常亮。

```go
device.KeepScreenOn()
```

## Vibrate
<hr style="margin: 0;">

使设备震动一段时间（单位毫秒，需要 root 权限）。

- `ms` {int} 要震动的时间（毫秒）。

```go
device.Vibrate(500)
```

## CancelVibration
<hr style="margin: 0;">

如果设备处于震动状态，则取消震动。

```go
device.CancelVibration()
```

---

### files - 文件系统

来源：https://autogo.cc/API/files.md

# files - 文件操作
---
提供文件和文件夹的操作接口，例如读取、写入、移动等。

## IsFile
<hr style="margin: 0;">

判断路径是否是文件。

- `path` {string} 路径

```go
isFile := files.IsFile("/sdcard/example.txt")
```

## IsDir
<hr style="margin: 0;">

判断路径是否是文件夹。

- `path` {string} 路径

```go
isDir := files.IsDir("/sdcard/example_folder")
```

## IsEmptyDir
<hr style="margin: 0;">

判断文件夹是否为空。如果路径不是文件夹，返回 false。

- `path` {string} 文件夹路径

```go
isEmpty := files.IsEmptyDir("/sdcard/example_folder")
```

## Create
<hr style="margin: 0;">

创建文件或文件夹。如果文件已存在，返回 true。

- `path` {string} 路径

```go
success := files.Create("/sdcard/new_file.txt")
```

## Exists
<hr style="margin: 0;">

判断路径是否存在。

- `path` {string} 路径

```go
exists := files.Exists("/sdcard/example.txt")
```

## EnsureDir
<hr style="margin: 0;">

确保文件夹存在，如果不存在则创建。

- `path` {string} 路径

```go
success := files.EnsureDir("/sdcard/new_folder")
```

## Read
<hr style="margin: 0;">

读取文本文件的内容。

- `path` {string} 文件路径

```go
content := files.Read("/sdcard/example.txt")
```

## ReadBytes
<hr style="margin: 0;">

读取文件的字节数据。

- `path` {string} 文件路径

```go
data := files.ReadBytes("/sdcard/example.txt")
```

## Write
<hr style="margin: 0;">

将文本写入文件。如果文件不存在则创建，存在则覆盖。

- `path` {string} 文件路径
- `text` {string} 要写入的文本

```go
files.Write("/sdcard/example.txt", "Hello, World!")
```

## WriteBytes
<hr style="margin: 0;">

将字节数据写入文件。如果文件不存在则创建，存在则覆盖。

- `path` {string} 文件路径
- `bytes` {[]byte} 要写入的字节数据

```go
files.WriteBytes("/sdcard/example.txt", []byte("Hello, World!"))
```

## Append
<hr style="margin: 0;">

将文本追加到文件末尾。如果文件不存在则创建。

- `path` {string} 文件路径
- `text` {string} 要追加的文本

```go
files.Append("/sdcard/example.txt", "Appended text")
```

## AppendBytes
<hr style="margin: 0;">

将字节数据追加到文件末尾。如果文件不存在则创建。

- `path` {string} 文件路径
- `bytes` {[]byte} 要追加的字节数据

```go
files.AppendBytes("/sdcard/example.txt", []byte("Appended bytes"))
```

## Copy
<hr style="margin: 0;">

复制文件。

- `fromPath` {string} 源文件路径
- `toPath` {string} 目标文件路径

```go
success := files.Copy("/sdcard/source.txt", "/sdcard/destination.txt")
```

## Move
<hr style="margin: 0;">

移动文件。

- `fromPath` {string} 源文件路径
- `toPath` {string} 目标文件路径

```go
success := files.Move("/sdcard/source.txt", "/sdcard/destination.txt")
```

## Rename
<hr style="margin: 0;">

重命名文件。

- `path` {string} 文件路径
- `newName` {string} 新文件名

```go
success := files.Rename("/sdcard/example.txt", "new_name.txt")
```

## GetName
<hr style="margin: 0;">

获取文件名。

- `path` {string} 文件路径

```go
name := files.GetName("/sdcard/example.txt")
```

## GetNameWithoutExtension
<hr style="margin: 0;">

获取不含扩展名的文件名。

- `path` {string} 文件路径

```go
name := files.GetNameWithoutExtension("/sdcard/example.txt")
```

## GetExtension
<hr style="margin: 0;">

获取文件的扩展名。

- `path` {string} 文件路径

```go
extension := files.GetExtension("/sdcard/example.txt")
```

## GetMd5
<hr style="margin: 0;">

获取文件的MD5值。

- `path` {string} 文件路径

```go
md5 := files.GetMd5("/sdcard/example.txt")
```

## Remove
<hr style="margin: 0;">

删除文件或文件夹。如果是文件夹，则删除其所有内容。

- `path` {string} 文件路径或文件夹路径

```go
success := files.Remove("/sdcard/example.txt")
```

## Path
<hr style="margin: 0;">

将相对路径转换为绝对路径。

- `relativePath` {string} 相对路径

```go
absolutePath := files.Path("./example.txt")
```

## ListDir
<hr style="margin: 0;">

列出文件夹下的所有文件和文件夹。

- `path` {string} 文件夹路径

```go
entries := files.ListDir("/sdcard/example_folder")
```

---

### uiacc - 节点操作

来源：https://autogo.cc/API/uiacc.md

# uiacc - 节点操作
---

提供基于辅助功能服务的控件定位、交互操作等功能。无需开启APP的无障碍服务。

以下是 `uiacc` 包中定义的 `Rect` 结构体及其字段说明：

| **字段名**  | **类型**    | **说明**                           |
|-------------|-------------|------------------------------------|
| `Left`      | `int`       | 矩形的左边界。                    |
| `Right`     | `int`       | 矩形的右边界。                    |
| `Top`       | `int`       | 矩形的上边界。                    |
| `Bottom`    | `int`       | 矩形的下边界。                    |
| `CenterX`   | `int`       | 矩形的中心 X 坐标。               |
| `CenterY`   | `int`       | 矩形的中心 Y 坐标。               |
| `Width`     | `int`       | 矩形的宽度。                      |
| `Height`    | `int`       | 矩形的高度。                      |

## New
<hr style="margin: 0;">

创建一个 Accessibility 对象。返回实例对象`*Uiacc`

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕，操作虚拟屏幕节点需要安卓版本大于等于11

```go
acc := uiacc.New(0)
```

## *Uiacc.Text
<hr style="margin: 0;">

设置选择器的 `text` 属性。

- `value` {string} 文本值。

```go
acc.Text("example text")
```

## *Uiacc.TextContains
<hr style="margin: 0;">

设置选择器的 `textContains` 属性。

- `value` {string} 包含的文本值。

```go
acc.TextContains("example")
```

## *Uiacc.TextStartsWith
<hr style="margin: 0;">

设置选择器的 `textStartsWith` 属性。

- `value` {string} 以此文本开头。

```go
acc.TextStartsWith("example")
```

## *Uiacc.TextEndsWith
<hr style="margin: 0;">

设置选择器的 `textEndsWith` 属性。

- `value` {string} 以此文本结尾。

```go
acc.TextEndsWith("example")
```

## *Uiacc.TextMatches
<hr style="margin: 0;">

设置选择器的 `textMatches` 属性，用于匹配符合指定正则表达式的控件。

- `value` {string} 正则表达式。

```go
acc.TextMatches("^example.*")
```

## *Uiacc.Desc
<hr style="margin: 0;">

设置选择器的 `desc` 属性，用于匹配描述等于指定文本的控件。

- `value` {string} 描述的文本值。

```go
acc.Desc("example description")
```

## *Uiacc.DescContains
<hr style="margin: 0;">

设置选择器的 `descContains` 属性，用于匹配描述包含指定文本的控件。

- `value` {string} 包含的描述文本值。

```go
acc.DescContains("example")
```

## *Uiacc.DescStartsWith
<hr style="margin: 0;">

设置选择器的 `descStartsWith` 属性，用于匹配描述以指定文本开头的控件。

- `value` {string} 描述文本的开头。

```go
acc.DescStartsWith("example")
```

## *Uiacc.DescEndsWith
<hr style="margin: 0;">

设置选择器的 `descEndsWith` 属性，用于匹配描述以指定文本结尾的控件。

- `value` {string} 描述文本的结尾。

```go
acc.DescEndsWith("example")
```

## *Uiacc.DescMatches
<hr style="margin: 0;">

设置选择器的 `descMatches` 属性，用于匹配描述符合指定正则表达式的控件。

- `value` {string} 正则表达式。

```go
acc.DescMatches("^example.*")
```

## *Uiacc.Id
<hr style="margin: 0;">

设置选择器的 `id` 属性，用于匹配 ID 等于指定值的控件。

- `value` {string} ID 值。

```go
acc.Id("example_id")
```

## *Uiacc.IdContains
<hr style="margin: 0;">

设置选择器的 `idContains` 属性，用于匹配 ID 包含指定值的控件。

- `value` {string} 包含的 ID 值。

```go
acc.IdContains("example")
```

## *Uiacc.IdStartsWith
<hr style="margin: 0;">

设置选择器的 `idStartsWith` 属性，用于匹配 ID 以指定值开头的控件。

- `value` {string} ID 的开头值。

```go
acc.IdStartsWith("example")
```

## *Uiacc.IdEndsWith
<hr style="margin: 0;">

设置选择器的 `idEndsWith` 属性，用于匹配 ID 以指定值结尾的控件。

- `value` {string} ID 的结尾值。

```go
acc.IdEndsWith("example")
```

## *Uiacc.IdMatches
<hr style="margin: 0;">

设置选择器的 `idMatches` 属性，用于匹配 ID 符合指定正则表达式的控件。

- `value` {string} 正则表达式。

```go
acc.IdMatches("^example.*")
```

## *Uiacc.ClassName
<hr style="margin: 0;">

设置选择器的 `className` 属性，用于匹配类名等于指定值的控件。

- `value` {string} 类名的值。

```go
acc.ClassName("example_class")
```

## *Uiacc.ClassNameContains
<hr style="margin: 0;">

设置选择器的 `classNameContains` 属性，用于匹配类名包含指定值的控件。

- `value` {string} 包含的类名值。

```go
acc.ClassNameContains("example")
```

## *Uiacc.ClassNameStartsWith
<hr style="margin: 0;">

设置选择器的 `classNameStartsWith` 属性，用于匹配类名以指定值开头的控件。

- `value` {string} 类名的开头值。

```go
acc.ClassNameStartsWith("example")
```

## *Uiacc.ClassNameEndsWith
<hr style="margin: 0;">

设置选择器的 `classNameEndsWith` 属性，用于匹配类名以指定值结尾的控件。

- `value` {string} 类名的结尾值。

```go
acc.ClassNameEndsWith("example")
```

## *Uiacc.ClassNameMatches
<hr style="margin: 0;">

设置选择器的 `classNameMatches` 属性，用于匹配类名符合指定正则表达式的控件。

- `value` {string} 正则表达式。

```go
acc.ClassNameMatches("^example.*")
```

## *Uiacc.PackageName
<hr style="margin: 0;">

设置选择器的 `packageName` 属性，用于匹配包名等于指定值的控件。

- `value` {string} 包名的值。

```go
acc.PackageName("com.example")
```

## *Uiacc.PackageNameContains
<hr style="margin: 0;">

设置选择器的 `packageNameContains` 属性，用于匹配包名包含指定值的控件。

- `value` {string} 包含的包名值。

```go
acc.PackageNameContains("example")
```

## *Uiacc.PackageNameStartsWith
<hr style="margin: 0;">

设置选择器的 `packageNameStartsWith` 属性，用于匹配包名以指定值开头的控件。

- `value` {string} 包名的开头值。

```go
acc.PackageNameStartsWith("com.example")
```

## *Uiacc.PackageNameEndsWith
<hr style="margin: 0;">

设置选择器的 `packageNameEndsWith` 属性，用于匹配包名以指定值结尾的控件。

- `value` {string} 包名的结尾值。

```go
acc.PackageNameEndsWith("example")
```

## *Uiacc.PackageNameMatches
<hr style="margin: 0;">

设置选择器的 `packageNameMatches` 属性，用于匹配包名符合指定正则表达式的控件。

- `value` {string} 正则表达式。

```go
acc.PackageNameMatches("^com\.example.*")
```

## *Uiacc.Bounds
<hr style="margin: 0;">

设置选择器的 `bounds` 属性，用于匹配控件在屏幕上的范围。

- `left, top, right, bottom` {int} 控件的屏幕边界。

```go
acc.Bounds(0, 0, 100, 100)
```

## *Uiacc.BoundsInside
<hr style="margin: 0;">

设置选择器的 `boundsInside` 属性，用于匹配控件在屏幕内的范围。

- `left, top, right, bottom` {int} 屏幕内的范围。

```go
acc.BoundsInside(0, 0, 500, 500)
```

## *Uiacc.BoundsContains
<hr style="margin: 0;">

设置选择器的 `boundsContains` 属性，用于匹配控件包含在指定范围内。

- `left, top, right, bottom` {int} 包含的范围。

```go
acc.BoundsContains(50, 50, 300, 300)
```

## *Uiacc.DrawingOrder
<hr style="margin: 0;">

设置选择器的 `drawingOrder` 属性，用于匹配控件在父控件中的绘制顺序。

- `value` {int} 绘制顺序。

```go
acc.DrawingOrder(2)
```

## *Uiacc.Clickable
<hr style="margin: 0;">

设置选择器的 `clickable` 属性，用于匹配控件是否可点击。

- `value` {bool} 是否可点击。

```go
acc.Clickable(true)
```

## *Uiacc.LongClickable
<hr style="margin: 0;">

设置选择器的 `longClickable` 属性，用于匹配控件是否可长按。

- `value` {bool} 是否可长按。

```go
acc.LongClickable(true)
```

## *Uiacc.Checkable
<hr style="margin: 0;">

设置选择器的 `checkable` 属性，用于匹配控件是否可选中。

- `value` {bool} 是否可选中。

```go
acc.Checkable(false)
```

## *Uiacc.Selected
<hr style="margin: 0;">

设置选择器的 `selected` 属性，用于匹配控件是否被选中。

- `value` {bool} 是否被选中。

```go
acc.Selected(true)
```

## *Uiacc.Enabled
<hr style="margin: 0;">

设置选择器的 `enabled` 属性，用于匹配控件是否启用。

- `value` {bool} 是否启用。

```go
acc.Enabled(true)
```

## *Uiacc.Scrollable
<hr style="margin: 0;">

设置选择器的 `scrollable` 属性，用于匹配控件是否可滚动。

- `value` {bool} 是否可滚动。

```go
acc.Scrollable(false)
```

## *Uiacc.Editable
<hr style="margin: 0;">

设置选择器的 `editable` 属性，用于匹配控件是否可编辑。

- `value` {bool} 是否可编辑。

```go
acc.Editable(true)
```

## *Uiacc.MultiLine
<hr style="margin: 0;">

设置选择器的 `multiLine` 属性，用于匹配控件是否多行。

- `value` {bool} 是否多行。

```go
acc.MultiLine(false)
```

## *Uiacc.Checked
<hr style="margin: 0;">

设置选择器的 `checked` 属性，用于匹配控件是否被勾选。

- `value` {bool} 是否勾选。

```go
acc.Checked(true)
```

## *Uiacc.Focusable
<hr style="margin: 0;">

设置选择器的 `focusable` 属性，用于匹配控件是否可聚焦。

- `value` {bool} 是否可聚焦。

```go
acc.Focusable(true)
```

## *Uiacc.Dismissable
<hr style="margin: 0;">

设置选择器的 `dismissable` 属性，用于匹配控件是否可解散。

- `value` {bool} 是否可解散。

```go
acc.Dismissable(false)
```

## *Uiacc.Focused
<hr style="margin: 0;">

设置选择器的 `focused` 属性，用于匹配控件是否是辅助功能焦点。

- `value` {bool} 是否为辅助功能焦点。

```go
acc.Focused(true)
```

## *Uiacc.ContextClickable
<hr style="margin: 0;">

设置选择器的 `contextClickable` 属性，用于匹配控件是否是上下文点击。

- `value` {bool} 是否为上下文点击。

```go
acc.ContextClickable(false)
```

## *Uiacc.Index
<hr style="margin: 0;">

设置选择器的 `index` 属性，用于匹配控件在父控件中的索引。

- `value` {int} 索引值。

```go
acc.Index(1)
```

## *Uiacc.Visible
<hr style="margin: 0;">

设置选择器的 `visible` 属性，用于匹配控件是否可见。

- `value` {bool} 是否可见。

```go
acc.Visible(true)
```

## *Uiacc.Password
<hr style="margin: 0;">

设置选择器的 `password` 属性，用于匹配控件是否为密码字段。

- `value` {bool} 是否为密码字段。

```go
acc.Password(false)
```

## *Uiacc.Click
<hr style="margin: 0;">

点击屏幕上的文本。

- `text` {string} 目标文本。

```go
acc.Click("目标文本")
```

## *Uiacc.WaitFor
<hr style="margin: 0;">

等待控件出现，返回控件对象 `*UiObject` 。

- `timeout` {int} 超时时间（毫秒）。`0` 表示无限等待。

```go
obj := acc.Text("hello").WaitFor(3000)
```

## *Uiacc.FindOnce
<hr style="margin: 0;">

查找单个控件，成功返回控件对象 `*UiObject` 。

```go
obj := acc.Text("hello").FindOnce()
```

## *Uiacc.Find
<hr style="margin: 0;">

查找所有符合条件的控件。返回控件对象数组 `[]*UiObject` 。

```go
objects := acc.Text("hello").Find()
```

## *Uiacc.Release
<hr style="margin: 0;">

释放无障碍服务资源 。

```go
uiacc.Release()
```

## *UiObject.Click
<hr style="margin: 0;">

点击该控件，并返回是否点击成功。

```go
success := uiObject.Click()
fmt.Println("点击成功:", success)
```

## *UiObject.ClickCenter
<hr style="margin: 0;">

使用控件坐标点击该控件的中点。

```go
success := uiObject.ClickCenter()
fmt.Println("点击中心成功:", success)
```

## *UiObject.ClickLongClick
<hr style="margin: 0;">

长按该控件，并返回是否点击成功。

```go
success := uiObject.ClickLongClick()
fmt.Println("长按成功:", success)
```

## *UiObject.Copy
<hr style="margin: 0;">

对输入框文本的选中内容进行复制，并返回是否操作成功。

```go
success := uiObject.Copy()
fmt.Println("复制成功:", success)
```

## *UiObject.Cut
<hr style="margin: 0;">

对输入框文本的选中内容进行剪切，并返回是否操作成功。

```go
success := uiObject.Cut()
fmt.Println("剪切成功:", success)
```

## *UiObject.Paste
<hr style="margin: 0;">

对输入框控件进行粘贴操作，把剪贴板内容粘贴到输入框中，并返回是否操作成功。

```go
success := uiObject.Paste()
fmt.Println("粘贴成功:", success)
```

## *UiObject.ScrollForward
<hr style="margin: 0;">

对控件执行向前滑动的操作，并返回是否操作成功。

```go
success := uiObject.ScrollForward()
fmt.Println("向前滑动成功:", success)
```

## *UiObject.ScrollBackward
<hr style="margin: 0;">

对控件执行向后滑动的操作，并返回是否操作成功。

```go
success := uiObject.ScrollBackward()
fmt.Println("向后滑动成功:", success)
```

## *UiObject.Collapse
<hr style="margin: 0;">

对控件执行折叠操作，并返回是否操作成功。

```go
success := uiObject.Collapse()
fmt.Println("折叠成功:", success)
```

## *UiObject.Expand
<hr style="margin: 0;">

对控件执行展开操作，并返回是否操作成功。

```go
success := uiObject.Expand()
fmt.Println("展开成功:", success)
```

## *UiObject.Show
<hr style="margin: 0;">

执行显示操作，并返回是否操作成功。

```go
success := uiObject.Show()
fmt.Println("显示成功:", success)
```

## *UiObject.Select
<hr style="margin: 0;">

对控件执行"选中"操作，并返回是否操作成功。

```go
selected := uiObject.Select()
fmt.Println("控件是否选中成功:", selected)
```

## *UiObject.ClearSelect
<hr style="margin: 0;">

清除控件的选中状态，并返回是否操作成功。

```go
cleared := uiObject.ClearSelect()
fmt.Println("控件是否清除选中成功:", cleared)
```

## *UiObject.SetSelection
<hr style="margin: 0;">

对输入框控件设置选中的文字内容，并返回是否操作成功。

- `start` {int} 选中内容的起始位置。
- `end` {int} 选中内容的结束位置。

```go
success := uiObject.SetSelection(0, 5)
fmt.Println("设置选中内容是否成功:", success)
```

## *UiObject.SetVisibleToUser
<hr style="margin: 0;">

设置控件是否可见。

- `isVisible` {bool} 是否可见。

```go
success := uiObject.SetVisibleToUser(false)
fmt.Println("设置控件不可见是否成功:", success)
```

## *UiObject.SetText
<hr style="margin: 0;">

设置输入框控件的文本内容，并返回是否设置成功。

- `str` {string} 文本内容。

```go
success := uiObject.SetText("example text")
fmt.Println("设置文本是否成功:", success)
```

## *UiObject.GetClickable
<hr style="margin: 0;">

获取控件的 `clickable` 属性。

```go
clickable := uiObject.GetClickable()
fmt.Println("控件是否可点击:", clickable)
```

## *UiObject.GetLongClickable
<hr style="margin: 0;">

获取控件的 `longClickable` 属性。

```go
longClickable := uiObject.GetLongClickable()
fmt.Println("控件是否支持长按:", longClickable)
```

## *UiObject.GetCheckable
<hr style="margin: 0;">

获取控件的 `checkable` 属性。

```go
checkable := uiObject.GetCheckable()
fmt.Println("控件是否可选中:", checkable)
```

## *UiObject.GetSelected
<hr style="margin: 0;">

获取控件的 `selected` 属性。

```go
selected := uiObject.GetSelected()
fmt.Println("控件是否被选中:", selected)
```

## *UiObject.GetEnabled
<hr style="margin: 0;">

获取控件的 `enabled` 属性。

```go
enabled := uiObject.GetEnabled()
fmt.Println("控件是否启用:", enabled)
```

## *UiObject.GetScrollable
<hr style="margin: 0;">

获取控件的 `scrollable` 属性。

```go
scrollable := uiObject.GetScrollable()
fmt.Println("控件是否可滚动:", scrollable)
```

## *UiObject.GetEditable
<hr style="margin: 0;">

获取控件的 `editable` 属性。

```go
editable := uiObject.GetEditable()
fmt.Println("控件是否可编辑:", editable)
```

## *UiObject.GetMultiLine
<hr style="margin: 0;">

获取控件的 `multiLine` 属性。

```go
multiLine := uiObject.GetMultiLine()
fmt.Println("控件是否多行:", multiLine)
```

## *UiObject.GetChecked
<hr style="margin: 0;">

获取控件的 `checked` 属性。

```go
checked := uiObject.GetChecked()
fmt.Println("控件是否被勾选:", checked)
```

## *UiObject.GetFocused
<hr style="margin: 0;">

获取控件的 `focused` 属性。

```go
focused := uiObject.GetFocused()
fmt.Println("控件是否获得了输入焦点:", focusable)
```

## *UiObject.GetFocusable
<hr style="margin: 0;">

获取控件的 `focusable` 属性。

```go
focusable := uiObject.GetFocusable()
fmt.Println("控件是否可聚焦:", focusable)
```

## *UiObject.GetDismissable
<hr style="margin: 0;">

获取控件的 `dismissable` 属性。

```go
dismissable := uiObject.GetDismissable()
fmt.Println("控件是否可解散:", dismissable)
```

## *UiObject.GetContextClickable
<hr style="margin: 0;">

获取控件的 `contextClickable` 属性。

```go
contextClickable := uiObject.GetContextClickable()
fmt.Println("控件是否支持上下文点击:", contextClickable)
```

## *UiObject.GetVisible
<hr style="margin: 0;">

获取控件的 `visible` 属性。

```go
visible := uiObject.GetVisible()
fmt.Println("控件是否可见:", visible)
```

## *UiObject.GetPassword
<hr style="margin: 0;">

获取控件的 `password` 属性。

```go
password := uiObject.GetPassword()
fmt.Println("控件是否为密码字段:", password)
```

## *UiObject.GetAccessibilityFocused
<hr style="margin: 0;">

获取控件的 `AccessibilityFocused` 属性。

```go
focused := uiObject.GetAccessibilityFocused()
fmt.Println("控件是否为辅助功能焦点:", focused)
```

## *UiObject.GetChildCount
<hr style="margin: 0;">

获取控件的子控件数目。

```go
childCount := uiObject.GetChildCount()
fmt.Println("子控件数量:", childCount)
```

## *UiObject.GetDrawingOrder
<hr style="margin: 0;">

获取控件在父控件中的绘制次序。

```go
drawingOrder := uiObject.GetDrawingOrder()
fmt.Println("控件绘制次序:", drawingOrder)
```

## *UiObject.GetIndex
<hr style="margin: 0;">

获取控件在父控件中的索引。

```go
index := uiObject.GetIndex()
fmt.Println("控件在父控件中的索引:", index)
```

## *UiObject.GetBounds
<hr style="margin: 0;">

获取控件在屏幕上的范围。

```go
bounds := uiObject.GetBounds()
fmt.Printf("控件范围: %v\n", bounds)
```

## *UiObject.GetBoundsInParent
<hr style="margin: 0;">

获取控件在父控件中的范围。

```go
bounds := uiObject.GetBoundsInParent()
fmt.Println("控件在父控件中的范围:", bounds)
```

## *UiObject.GetId
<hr style="margin: 0;">

获取控件的 ID。

```go
id := uiObject.GetId()
fmt.Println("控件 ID:", id)
```

## *UiObject.GetText
<hr style="margin: 0;">

获取控件的文本内容。

```go
text := uiObject.GetText()
fmt.Println("控件文本内容:", text)
```

## *UiObject.GetDesc
<hr style="margin: 0;">

获取控件的描述内容。

```go
desc := uiObject.GetDesc()
fmt.Println("控件描述内容:", desc)
```

## *UiObject.GetPackageName
<hr style="margin: 0;">

获取控件的包名。

```go
packageName := uiObject.GetPackageName()
fmt.Println("控件包名:", packageName)
```

## *UiObject.GetClassName
<hr style="margin: 0;">

获取控件的类名。

```go
className := uiObject.GetClassName()
fmt.Println("控件类名:", className)
```

## *UiObject.GetParent
<hr style="margin: 0;">

获取控件的父控件。

```go
parent := uiObject.GetParent()
fmt.Println("控件的父控件:", parent)
```

## *UiObject.GetChild
<hr style="margin: 0;">

获取控件的指定索引的子控件。

- `index` {int} 子控件的索引。

```go
child := uiObject.GetChild(0)
fmt.Println("第一个子控件:", child)
```

## *UiObject.GetChildren
<hr style="margin: 0;">

获取控件的所有子控件。返回控件对象数组 `[]*UiObject` 。

```go
children := uiObject.GetChildren()
for index, child := range children {
    fmt.Printf("子控件 %d: %v\n", index+1, child)
}
```

## *UiObject.ToString
<hr style="margin: 0;">

将节点对象转文本。

```go
str := uiObject.ToString()
fmt.Println("节点文本:", str)
```

---

### https - HTTP网络请求

来源：https://autogo.cc/API/https.md

# https - 网络请求
---
https模块提供了发送HTTP/HTTPS请求的功能，可用于与网络服务进行交互，获取网页内容，上传文件等。

## Get
<hr style="margin: 0;">

发送GET请求并返回响应状态码和数据。

- `url` {string} 请求的URL
- `timeout` {int} 请求的超时时间（毫秒），如果为0则不设置超时

```go
code, data := https.Get("https://example.com", 5000)
if code == 200 {
    fmt.Println("请求成功：", string(data))
} else {
    fmt.Println("请求失败，状态码：", code)
}
```

## Post
<hr style="margin: 0;">

发送POST请求并返回响应状态码和数据。支持自定义请求头和请求体，适用于发送JSON、XML等格式的数据。

- `url` {string} 请求的URL
- `data` {[]byte} 请求体数据（如JSON序列化后的字节数组）
- `headers` {map[string]string} 自定义请求头，如果为nil或未设置Content-Type，默认使用application/json
- `timeout` {int} 请求的超时时间（毫秒），如果为0则不设置超时

```go
// 构造JSON请求体
reqData := map[string]interface{}{
    "name": "张三",
    "age":  25,
}
jsonData, _ := json.Marshal(reqData)

// 设置请求头
headers := map[string]string{
    "Content-Type":  "application/json",
    "Authorization": "Bearer your-token",
}

// 发送请求
code, data := https.Post("https://example.com/api/user", jsonData, headers, 10000)
if code == 200 {
    fmt.Println("请求成功：", string(data))
} else {
    fmt.Println("请求失败，状态码：", code)
}
``` 

## PostMultipart
<hr style="margin: 0;">

发送带有文件的POST请求（multipart/form-data格式）并返回响应状态码和数据。

- `url` {string} 请求的URL
- `fileName` {string} 文件名
- `fileData` {[]byte} 文件数据（字节数组）
- `timeout` {int} 请求的超时时间（毫秒），如果为0则不设置超时

```go
// 读取文件数据
fileData := files.ReadBytes("/sdcard/image.jpg")
// 发送请求
code, data := https.PostMultipart("https://example.com/upload", "image.jpg", fileData, 10000)
if code == 200 {
    fmt.Println("上传成功：", string(data))
} else {
    fmt.Println("上传失败，状态码：", code)
}
```

---

### ime - 输入法

来源：https://autogo.cc/API/ime.md

# ime - 输入法
---
ime模块提供了一系列函数，用于控制输入法行为，实现文本输入等功能。

## InputText
<hr style="margin: 0;">

使用输入法输入文本。

- `text` {string} 需要输入的文本

```go
ime.InputText("Hello, World!")
```

## GetClipText
<hr style="margin: 0;">

获取剪贴板文本内容。

```go
text := ime.GetClipText()
fmt.Println("剪贴板内容:", text)
```

## SetClipText
<hr style="margin: 0;">

设置剪贴板文本内容。

- `text` {string} 要设置的文本

```go
ime.SetClipText("这是要复制到剪贴板的内容")
``` 

## GetIMEList
<hr style="margin: 0;">

获取输入法列表。

```go
fmt.Println("输入法列表:",ime.GetIMEList())
``` 

## SetCurrentIME
<hr style="margin: 0;">

设置系统当前输入法。

- `packageName` {string} 要设置为当前输入法的应用包名，如果为空且当前是APK安装模式在运行会将脚本的输入法设置为当前输入法

```go
ime.SetCurrentIME("com.android.inputmethod.latin")
```

---

### motion - 动作

来源：https://autogo.cc/API/motion.md

# motion - 操作
---
motion模块提供了一系列模拟用户操作的函数，如点击、滑动、按键等。

## TouchDown
<hr style="margin: 0;">

模拟触摸屏按下操作。

- `x` {int} 触摸点的X坐标
- `y` {int} 触摸点的Y坐标
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.TouchDown(500, 600, 0, 0)
```

## TouchMove
<hr style="margin: 0;">

模拟触摸屏移动操作。

- `x` {int} 移动到的X坐标
- `y` {int} 移动到的Y坐标
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.TouchMove(550, 650, 0, 0)
```

## TouchUp
<hr style="margin: 0;">

模拟触摸屏抬起操作。

- `x` {int} 抬起点的X坐标
- `y` {int} 抬起点的Y坐标
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.TouchUp(550, 650, 0, 0)
```

## Click
<hr style="margin: 0;">

模拟单击操作。

- `x` {int} 单击点的X坐标
- `y` {int} 单击点的Y坐标
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.Click(500, 600, 0, 0)
```

## LongClick
<hr style="margin: 0;">

模拟长按操作。

- `x` {int} 长按点的X坐标
- `y` {int} 长按点的Y坐标
- `duration` {int} 长按持续时间（毫秒）
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.LongClick(500, 600, 1000, 0, 0)  // 长按1秒
```

## Swipe
<hr style="margin: 0;">

模拟滑动操作。

- `x1` {int} 起始点的X坐标
- `y1` {int} 起始点的Y坐标
- `x2` {int} 结束点的X坐标
- `y2` {int} 结束点的Y坐标
- `duration` {int} 滑动持续时间（毫秒）
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.Swipe(300, 800, 300, 200, 500, 0, 0)  // 从下往上滑动
```

## Swipe2
<hr style="margin: 0;">

使用贝塞尔曲线方式进行滑动（轨迹更加自然）。

- `x1` {int} 起始点的X坐标
- `y1` {int} 起始点的Y坐标
- `x2` {int} 结束点的X坐标
- `y2` {int} 结束点的Y坐标
- `duration` {int} 滑动持续时间（毫秒）
- `fingerID` {int} 触摸点的指针ID（0-9）
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.Swipe2(300, 800, 300, 200, 500, 0, 0)  // 从下往上滑动，轨迹更自然
```

## Home
<hr style="margin: 0;">

模拟按下Home键。

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.Home(0)
```

## Back
<hr style="margin: 0;">

模拟按下返回键。

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.Back(0)
```

## Recents
<hr style="margin: 0;">

显示最近任务。

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.Recents(0)
```

## PowerDialog
<hr style="margin: 0;">

弹出电源键菜单。

```go
motion.PowerDialog()
```

## Notifications
<hr style="margin: 0;">

拉出通知栏。

```go
motion.Notifications()
```

## QuickSettings
<hr style="margin: 0;">

显示快速设置（下拉通知栏到底）。

```go
motion.QuickSettings()
```

## VolumeUp
<hr style="margin: 0;">

按下音量上键。

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.VolumeUp(0)
```

## VolumeDown
<hr style="margin: 0;">

按下音量下键。

- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.VolumeDown(0)
```

## Camera
<hr style="margin: 0;">

模拟按下照相键。

```go
motion.Camera()
```

## KeyAction
<hr style="margin: 0;">

模拟单击指定按键。

- `code` {int} 按键代码，参考KEYCODE_常量
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.KeyAction(motion.KEYCODE_ENTER, 0)  // 单击回车键
``` 

## KeyActionDown
<hr style="margin: 0;">

模拟按下指定按键。

- `code` {int} 按键代码，参考KEYCODE_常量
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.KeyActionDown(motion.KEYCODE_ENTER, 0)  // 按下回车键
``` 

## KeyActionUp
<hr style="margin: 0;">

模拟弹起指定按键。

- `code` {int} 按键代码，参考KEYCODE_常量
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
motion.KeyActionUp(motion.KEYCODE_ENTER, 0)  // 弹起回车键
```

---

### images - 图像处理

来源：https://autogo.cc/API/images.md

# images - 图像处理
---
images模块提供了截图、图像处理、颜色查找等功能。

## SetCallback
<hr style="margin: 0;">

设置一个新图像数据到达的回调。

- `callback` {function} 当新图像数据到达时调用的函数，格式为 `func(img *image.NRGBA, displayId int)`，如果传入 `nil`，则会移除当前设置的回调

```go
images.SetCallback(func(img *image.NRGBA, displayId int) {
    // 处理新图像数据
})
```

**注意事项：**
- 回调函数应避免执行耗时操作，否则可能导致后续图像数据处理延迟
- 回调函数内部如需进行耗时操作（如文件写入或网络请求），建议启动新的 goroutine 处理，避免阻塞回调执行

## CaptureScreen
<hr style="margin: 0;">

截取屏幕的指定区域。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
img := images.CaptureScreen(0, 0, 0, 0, 0) // 截取主屏幕全屏
```

## Pixel
<hr style="margin: 0;">

获取指定坐标点的颜色值。

- `x` {int} 坐标点的 x 位置
- `y` {int} 坐标点的 y 位置
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
color := images.Pixel(100, 200, 0)
```

## CmpColor
<hr style="margin: 0;">

比较指定坐标点 (x, y) 的颜色。

- `x` {int} 坐标点的 x 位置
- `y` {int} 坐标点的 y 位置
- `colorStr` {string} 颜色字符串，格式如 "FFFFFF|CCCCCC-101010"，每种颜色用 "|" 分割，"-" 后表示偏色
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
matched := images.CmpColor(100, 200, "FFFFFF|CCCCCC-101010", 0.9, 0)
```

## FindColor
<hr style="margin: 0;">

在指定区域内查找目标颜色。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colorStr` {string} 颜色格式串，例如 "FFFFFF|CCCCCC-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `dir` {int} 查找方向，0-3 分别表示不同的查找方向
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

**查找方向说明：**
- 0 - 从左到右，从上到下
- 1 - 从右到左，从上到下
- 2 - 从左到右，从下到上
- 3 - 从右到左，从下到上

```go
x, y := images.FindColor(0, 0, 0, 0, "FFFFFF", 0.9, 0, 0)
```

## GetColorCountInRegion
<hr style="margin: 0;">

计算指定区域内符合颜色条件的像素数量。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colorStr` {string} 要查找的颜色字符串，例如 "FFFFFF|CCCCCC-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
count := images.GetColorCountInRegion(0, 0, 0, 0, "FFFFFF", 0.9, 0)
```

## DetectsMultiColors
<hr style="margin: 0;">

根据指定的颜色串信息在屏幕进行多点颜色比对（多点比色）。

- `colors` {string} 颜色模板字符串，例如 "369,1220,ffab2d-101010,370,1221,24b1ff-101010,380,390,907efd-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
matched := images.DetectsMultiColors("369,1220,ffab2d-101010,370,1221,24b1ff-101010", 0.9, 0)
```

## FindMultiColors
<hr style="margin: 0;">

在指定区域内查找匹配的多点颜色序列（多点找色）。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colors` {string} 颜色模板字符串，例如 "ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `dir` {int} 查找方向，0-3 分别表示不同的查找方向
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
x, y := images.FindMultiColors(0, 0, 0, 0, "ffccff-151515,635,978,ffab2d-101010", 0.9, 0, 0)
```

## FindMultiColorsAll
<hr style="margin: 0;">

在指定区域内查找匹配的多点颜色序列并返回所有符合条件的坐标。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colors` {string} 颜色模板字符串，例如 "ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `dir` {int} 查找方向，0-3 分别表示不同的查找方向
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
points := images.FindMultiColorsAll(0, 0, 0, 0, "ffccff-151515,635,978,ffab2d-101010", 0.9, 0, 0)
```

## ReadFromPath
<hr style="margin: 0;">

读取路径指定的图片文件并返回图像对象。

- `path` {string} 要读取的图片文件路径

```go
img := images.ReadFromPath("/sdcard/image.png")
```

## ReadFromBase64
<hr style="margin: 0;">

解码 Base64 数据并返回解码后的图片对象。

- `base64Str` {string} 要解码的 Base64 字符串

```go
img := images.ReadFromBase64("iVBORw0KGgoAAAANSUhEUgAAAAUA...")
```

## ReadFromBytes
<hr style="margin: 0;">

解码字节数组并返回解码后的图片对象。

- `data` {[]byte} 要解码的字节数组

```go
img := images.ReadFromBytes(bytes)
```

## Save
<hr style="margin: 0;">

把图片保存到指定路径。

- `img` {*image.NRGBA} 要保存的图像对象
- `path` {string} 保存图片的文件路径
- `quality` {int} 保存图片的质量（如果适用）

```go
success := images.Save(img, "/sdcard/saved.png", 100)
```

## EncodeToBase64
<hr style="margin: 0;">

把图像对象编码为 Base64 数据。

- `img` {*image.NRGBA} 要编码的图像对象
- `format` {string} 编码的图片格式（如 "png", "jpg" 等）
- `quality` {int} 编码的图片质量

```go
base64Str := images.EncodeToBase64(img, "png", 100)
```

## EncodeToBytes
<hr style="margin: 0;">

把图片编码为字节数组。

- `img` {*image.NRGBA} 要编码的图像对象
- `format` {string} 编码的图片格式（如 "png", "jpg" 等）
- `quality` {int} 编码的图片质量

```go
bytes := images.EncodeToBytes(img, "png", 100)
```

## ToNrgba
<hr style="margin: 0;">

将任意 image.Image 对象转换为 *image.NRGBA。

- `img` {image.Image} 要转换的图像对象

```go
nrgbaImg := images.ToNrgba(anyImg)
```

## Clip
<hr style="margin: 0;">

从源图像中裁剪指定区域并返回新图像。

- `img` {*image.NRGBA} 要裁剪的图像
- `x1` {int} 裁剪区域左上角 x 坐标
- `y1` {int} 裁剪区域左上角 y 坐标
- `x2` {int} 裁剪区域右下角 x 坐标
- `y2` {int} 裁剪区域右下角 y 坐标

```go
clippedImg := images.Clip(img, 100, 100, 300, 300)
```

## Resize
<hr style="margin: 0;">

调整图像大小。

- `img` {*image.NRGBA} 要调整的图像
- `width` {int} 目标宽度
- `height` {int} 目标高度

```go
resizedImg := images.Resize(img, 800, 600)
```

## Rotate
<hr style="margin: 0;">

旋转图像。

- `img` {*image.NRGBA} 要旋转的图像
- `degree` {int} 旋转角度（顺时针方向）

```go
rotatedImg := images.Rotate(img, 90)
```

## Grayscale
<hr style="margin: 0;">

将彩色图像转换为灰度图像。

- `img` {*image.NRGBA} 要转换的彩色图像

```go
grayImg := images.Grayscale(img)
```

## ApplyThreshold
<hr style="margin: 0;">

对图像应用阈值处理。

- `img` {*image.NRGBA} 要处理的图像
- `threshold` {int} 阈值
- `maxVal` {int} 超过阈值后应用的值
- `typ` {string} 阈值类型，如 "BINARY", "BINARY_INV" 等

```go
thresholdImg := images.ApplyThreshold(img, 128, 255, "BINARY")
```

## ApplyAdaptiveThreshold
<hr style="margin: 0;">

应用自适应阈值处理。

- `img` {*image.NRGBA} 要处理的图像
- `maxValue` {float64} 最大值
- `adaptiveMethod` {string} 自适应方法，如 "MEAN_C", "GAUSSIAN_C"
- `thresholdType` {string} 阈值类型，如 "BINARY", "BINARY_INV"
- `blockSize` {int} 用于计算阈值的像素邻域大小
- `C` {float64} 从平均值或加权平均值中减去的常量

```go
adaptiveImg := images.ApplyAdaptiveThreshold(img, 255, "MEAN_C", "BINARY", 11, 2)
```

## ApplyBinarization
<hr style="margin: 0;">

应用二值化处理。

- `img` {*image.NRGBA} 要处理的图像
- `threshold` {int} 阈值

```go
binaryImg := images.ApplyBinarization(img, 128)
```

---

### opencv - 图像处理

来源：https://autogo.cc/API/opencv.md

# opencv - 图像处理
---
提供基于 OpenCV 的图像处理功能。由于 OpenCV 方法数量太多，剩余的方法全部参照 [官方文档](https://docs.opencv.org/4.10.0/)

## FindImage
<hr style="margin: 0;">

在指定区域内查找匹配的图片模板。返回找到的图片左上角坐标，如果未找到则返回 (-1, -1)。

- `x1`, `y1` {int}: 区域左上角的坐标。
- `x2`, `y2` {int}: 区域右下角的坐标。当 `x2` 或 `y2` 为 0 时，表示使用图像的最大宽度或高度。
- `template` {*[]byte}: 模板图片的字节数组指针，表示要在区域内查找的图片。
- `isGray` {bool}: 是否将图像转换为灰度图进行匹配。
- `isTransparent` {bool}: 是否按透明图处理。为 `true` 时，模板左上角第一个像素的 RGB 颜色会被当作透明色，匹配时忽略模板中所有同色像素。
- `sim` {float32}: 相似度阈值，取值范围为 0.1 到 1.0，值越高表示匹配要求越精确。
- `displayId` {int}: 屏幕ID。

说明：

- `isTransparent` 为 `false` 时，按普通图片匹配，不生成遮罩。
- `isTransparent` 为 `true` 时，按透明图片匹配，透明色由模板左上角像素决定。

```go
x, y := opencv.FindImage(0, 0, 1920, 1080, &templateBytes, false, false, 0.8, 0)
if x != -1 && y != -1 {
    fmt.Printf("模板匹配成功，坐标为: (%d, %d)\n", x, y)
} else {
    fmt.Println("未找到匹配的模板。")
}
```

## FindImageFromImage
<hr style="margin: 0;">

在给定图像中查找匹配的图片模板。参数含义与 `FindImage` 相同，但 `img` 直接作为待匹配图像，不会进行屏幕截图。

- `img` {*image.NRGBA}: 待匹配图像。
- `template` {*[]byte}: 模板图片的字节数组指针。
- `isGray` {bool}: 是否将图像转换为灰度图进行匹配。
- `isTransparent` {bool}: 是否按透明图处理。为 `true` 时，模板左上角第一个像素的 RGB 颜色会被当作透明色。
- `sim` {float32}: 相似度阈值，取值范围为 0.1 到 1.0。

```go
x, y := opencv.FindImageFromImage(img, &templateBytes, false, false, 0.8)
```

## FindImageAll
<hr style="margin: 0;">

在指定区域内查找匹配的图片模板，返回所有符合条件的坐标。

- `x1`, `y1` {int}: 区域左上角的坐标。
- `x2`, `y2` {int}: 区域右下角的坐标。当 `x2` 或 `y2` 为 0 时，表示使用图像的最大宽度或高度。
- `template` {*[]byte}: 模板图片的字节数组指针，表示要在区域内查找的图片。
- `isGray` {bool}: 是否将图像转换为灰度图进行匹配。
- `isTransparent` {bool}: 是否按透明图处理。为 `true` 时，模板左上角第一个像素的 RGB 颜色会被当作透明色。
- `sim` {float32}: 相似度阈值，取值范围为 0.1 到 1.0，值越高表示匹配要求越精确。
- `displayId` {int}: 屏幕ID。

```go
points := opencv.FindImageAll(0, 0, 1920, 1080, &templateBytes, false, false, 0.8, 0)
```

---

### dotocr - 点阵OCR

来源：https://autogo.cc/API/dotocr.md

# dotocr - 点阵OCR
---
提供基于模板匹配的文字识别和查找功能。

## SetDict
<hr style="margin: 0;">

设置字库。字库内容按行分割，每行一条模板记录。

- `name` {string} 字库名称，为空字符串时使用 "default"
- `dict` {string} 字库内容字符串，按行分割，每行一条模板记录

```go
dotocr.SetDict("字库1", dictContent)
```

## Ocr
<hr style="margin: 0;">

在屏幕指定区域进行OCR文字识别。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `threshold` {string} 阈值字符串，例如 "ffffff-101010"
- `sim` {float32} 匹配相似度阈值，取值范围 0.0-1.0，例如 0.8
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称，为空字符串时使用 "default"
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

**返回值：**
- 当 `asJSON == false` 时，返回纯文本字符串，按检测顺序拼接所有识别到的字符
- 当 `asJSON == true` 时，返回 JSON 数组字符串，每个元素包含 `{"x":坐标x, "y":坐标y, "width":宽度, "height":高度, "text":文字内容, "sim":相似度}`

```go
// 返回纯文本
text := dotocr.Ocr(0, 0, 100, 50, "ffffff-101010", 0.8, false, "字库1", 0)

// 返回 JSON 格式
jsonStr := dotocr.Ocr(0, 0, 100, 50, "ffffff-101010", 0.8, true, "字库1", 0)
```

## OcrFromImage
<hr style="margin: 0;">

从图像对象进行OCR文字识别。

- `img` {*image.NRGBA} 要识别的图像对象
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称

**返回值：** 识别结果字符串（纯文本或 JSON 格式）

```go
img := images.ReadFromPath("/sdcard/screenshot.png")
text := dotocr.OcrFromImage(img, "ffffff-101010", 0.8, false, "字库1")
```

## OcrFromBase64
<hr style="margin: 0;">

从Base64编码的图像数据进行OCR文字识别。

- `b64` {string} Base64编码的图像数据
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称

**返回值：** 识别结果字符串（纯文本或 JSON 格式）

```go
text := dotocr.OcrFromBase64(base64String, "ffffff-101010", 0.8, false, "字库1")
```

## OcrFromPath
<hr style="margin: 0;">

从图像文件路径进行OCR文字识别。

- `path` {string} 图像文件路径
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称

**返回值：** 识别结果字符串（纯文本或 JSON 格式）

```go
text := dotocr.OcrFromPath("/sdcard/screenshot.png", "ffffff-101010", 0.8, false, "字库1")
```

## FindStr
<hr style="margin: 0;">

在屏幕指定区域中查找指定字符串的位置。

- `x1` {int} 查找区域的左上角 x 坐标
- `y1` {int} 查找区域的左上角 y 坐标
- `x2` {int} 查找区域的右下角 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 查找区域的右下角 y 坐标，当为 0 时表示使用屏幕最大高度
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
x, y := dotocr.FindStr(0, 0, 0, 0, "商店", "ffffff-101010", 0.8, "字库1", 0)
if x != -1 && y != -1 {
    fmt.Printf("找到字符串，坐标: (%d, %d)\n", x, y)
}
```

## FindStrFromImage
<hr style="margin: 0;">

在图像对象中查找指定字符串的位置。

- `img` {*image.NRGBA} 要查找的图像对象
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
img := images.ReadFromPath("/sdcard/screenshot.png")
x, y := dotocr.FindStrFromImage(img, "商店", "ffffff-101010", 0.8, "字库1")
```

## FindStrFromBase64
<hr style="margin: 0;">

在Base64编码的图像中查找指定字符串的位置。

- `b64` {string} Base64编码的图像数据
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
x, y := dotocr.FindStrFromBase64(base64String, "商店", "ffffff-101010", 0.8, "字库1")
```

## FindStrFromPath
<hr style="margin: 0;">

在图像文件中查找指定字符串的位置。

- `path` {string} 图像文件路径
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
x, y := dotocr.FindStrFromPath("/sdcard/screenshot.png", "商店", "ffffff-101010", 0.8, "字库1")
```

---

### media - 多媒体

来源：https://autogo.cc/API/media.md

# media - 多媒体
---
提供多媒体编程的支持，目前仅支持媒体文件扫描，后续会加入更多功能。

## ScanFile
<hr style="margin: 0;">

扫描指定文件，将其加入媒体库中。

- `path` {string} 要扫描的文件路径。

```go
media.ScanFile("/sdcard/1.png")
```

## PlayMP3
<hr style="margin: 0;">

播放指定路径的 MP3 音频文件。

- `path` {string} 要播放的 MP3 文件路径。

```go
media.PlayMP3("/sdcard/music.mp3")
```

## SendSMS
<hr style="margin: 0;">

向指定手机号发送短信。

- `number` {string} 接收短信的目标手机号。
- `message` {string} 要发送的短信内容。

```go
media.SendSMS("10086", "Hello AutoGo")
```

---

### rhino - JS脚本引擎

来源：https://autogo.cc/API/rhino.md

# rhino - JS脚本引擎
---
提供 JavaScript 脚本执行能力，基于 Rhino 引擎实现。

## Eval
<hr style="margin: 0;">

执行指定的 JavaScript 脚本，并返回执行结果。

- `contextId` {string} 执行上下文的标识符，用于区分不同的脚本运行环境（可用于隔离变量作用域或缓存）。
- `script` {string} 要执行的 JavaScript 代码字符串。

```go
fmt.Println(rhino.Eval("script", `importClass(android.os.Build);Build.MODEL`))
```

---

### storages - 本地存储

来源：https://autogo.cc/API/storages.md

# storages - 本地存储
---
提供了保存简单数据、用户配置等的支持。保存的数据除非应用被卸载或者被主动删除，否则会一直保留。

## Get
<hr style="margin: 0;">

从本地存储中取出键值为 `key` 的数据。

- `table` {string} 要操作的表。
- `key` {string} 要查询的键。

```go
value := storages.Get("data", "username")
fmt.Println("获取到的值:", value)
```

## Put
<hr style="margin: 0;">

把值 `value` 保存到本地存储中。

- `table` {string} 要操作的表。
- `key` {string} 要保存的键。
- `value` {string} 要保存的值。

```go
storages.Put("data", "username", "JohnDoe")
```

## Remove
<hr style="margin: 0;">

移除键值为 `key` 的数据。

- `table` {string} 要操作的表。
- `key` {string} 要移除的键。

```go
storages.Remove("data", "username")
```

## Contains
<hr style="margin: 0;">

返回该本地存储是否包含键值为 `key` 的数据。

- `table` {string} 要操作的表。
- `key` {string} 要检查的键。

```go
exists := storages.Contains("data", "username")
if exists {
    fmt.Println("键存在!")
} else {
    fmt.Println("键不存在!")
}
```

## GetAll
<hr style="margin: 0;">

获取该本地存储所有键值对。

- `table` {string} 要操作的表。

```go
data := storages.GetAll("data")
for k, v := range data {
	fmt.Printf("%s = %s\n", k, v)
}
```

## Clear
<hr style="margin: 0;">

移除该本地存储的所有数据。

- `table` {string} 要操作的表。

```go
storages.Clear("data")
fmt.Println("所有存储数据已被清除!")
```

---

### plugin - 外部插件

来源：https://autogo.cc/API/plugin.md

# plugin - APK插件加载
---
plugin模块提供了动态加载外部APK插件的功能，通过JNI实现Go代码与Java代码的交互。可以加载APK中的类、创建实例并调用方法，支持参数自动转换和Bitmap内存管理。

## 类型说明

### Context
<hr style="margin: 0;">

`Context` 是一个占位符类型，用于表示Android Context对象。当作为参数传递时，plugin会自动获取Android应用上下文。

```go
type Context struct{}
```

**使用示例：**
```go
// 传递Context参数给Java方法
instance.Call("methodNeedContext", plugin.Context{}, "otherParam")
```

### AssetManager
<hr style="margin: 0;">

`AssetManager` 是一个占位符类型，用于表示Android AssetManager对象。当作为参数传递时，plugin会自动创建AssetManager实例并添加ClassLoader中保存的APK路径。

```go
type AssetManager struct{}
```

**使用示例：**
```go
// 创建AssetManager参数
assetMgr := plugin.NewAssetManager()

// 传递AssetManager参数给Java方法
instance.Call("loadResources", assetMgr, "resource_name")
```

### ClassLoader
<hr style="margin: 0;">

`ClassLoader` 表示一个已加载的APK类加载器，用于创建APK中的类实例。

**方法：**
- `NewInstance(className string, args ...interface{}) (*Instance, error)` - 创建类实例
- `Release()` - 释放类加载器资源

### Instance
<hr style="margin: 0;">

`Instance` 表示一个Java对象实例，用于调用对象的方法。

**方法：**
- `Call(methodName string, args ...interface{}) (interface{}, error)` - 调用实例方法
- `CallInt(methodName string, args ...interface{}) (int, error)` - 调用返回int的方法
- `CallLong(methodName string, args ...interface{}) (int64, error)` - 调用返回long的方法
- `CallFloat(methodName string, args ...interface{}) (float32, error)` - 调用返回float的方法
- `CallDouble(methodName string, args ...interface{}) (float64, error)` - 调用返回double的方法
- `CallBool(methodName string, args ...interface{}) (bool, error)` - 调用返回boolean的方法
- `CallString(methodName string, args ...interface{}) (string, error)` - 调用返回String的方法
- `CallVoid(methodName string, args ...interface{}) error` - 调用无返回值的方法
- `Release()` - 释放实例资源

## LoadApk
<hr style="margin: 0;">

加载外部APK插件文件，返回一个类加载器对象。加载时会自动提取APK中与当前Go二进制架构对应的SO库到Go二进制所在目录。

- `apkPath` {string} APK文件的绝对路径

**返回值：**
- `*ClassLoader` 类加载器对象，加载失败返回 `nil`

**架构映射：**
- `arm64` → `arm64-v8a`
- `amd64` → `x86_64`
- `386` → `x86`

```go
cl := plugin.LoadApk("/data/local/tmp/assets/my_plugin.apk")
if cl != nil {
    defer cl.Release()
    fmt.Println("APK加载成功")
}
```

## NewInstance
<hr style="margin: 0;">

创建APK中指定类的实例。支持普通类和静态内部类（使用`$`分隔）。

- `className` {string} 完整的类名，例如 "com.example.MyClass" 或 "com.example.MyClass$InnerClass"
- `args` {...interface{}} 构造函数参数（可选）

**返回值：**
- `*Instance` 实例对象
- `error` 错误信息，成功返回 `nil`

**支持的参数类型：**
- `int`, `int32` → Java `int`
- `int64` → Java `long`
- `float32` → Java `float`
- `float64` → Java `double`
- `bool` → Java `boolean`
- `string` → Java `String`
- `[]byte` → Java `byte[]`
- `*image.NRGBA` → `android.graphics.Bitmap`（自动转换）
- `plugin.Context` → `android.content.Context`（自动获取）
- `plugin.AssetManager` → `android.content.res.AssetManager`（自动创建并添加APK路径）

```go
// 创建无参构造的实例
instance, err := cl.NewInstance("com.example.MyClass")
if err != nil {
    fmt.Println("创建实例失败:", err)
    return
}
defer instance.Release()

// 创建带参数的实例
instance, err := cl.NewInstance("com.example.MyClass", "param1", int64(123), true)

// 创建静态内部类实例
instance, err := cl.NewInstance("com.example.MyClass$InnerClass")
```

## Call
<hr style="margin: 0;">

调用实例的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `interface{}` 方法返回值（可能为 `nil`）
- `error` 错误信息，成功返回 `nil`

**注意事项：**
- 如果Java方法需要 `long` 类型参数，Go端需要显式传递 `int64` 类型，不能直接传递 `int`
- 传递 `*image.NRGBA` 作为参数时，会自动转换为 `android.graphics.Bitmap`，并在方法调用后自动调用 `recycle()` 释放内存
- 返回值类型需要根据实际情况进行类型断言或使用类型安全的调用方法

```go
// 调用无参方法
result, err := instance.Call("getStatus")

// 调用带参数的方法
result, err := instance.Call("process", "input", int64(100))

// 传递图像参数（自动转为Bitmap）
img := images.CaptureScreen(0, 0, 0, 0, 0)
result, err := instance.Call("analyzeImage", img)

// 传递Context参数
result, err := instance.Call("methodNeedContext", plugin.Context{}, "param2")

// 传递AssetManager参数
assetMgr := plugin.NewAssetManager()
result, err := instance.Call("loadAsset", assetMgr, "asset_file.txt")
```

## CallInt
<hr style="margin: 0;">

调用返回 `int` 类型的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `int` 方法返回的整数值
- `error` 错误信息，成功返回 `nil`

```go
count, err := instance.CallInt("getCount")
if err != nil {
    fmt.Println("调用失败:", err)
} else {
    fmt.Printf("计数: %d\n", count)
}
```

## CallLong
<hr style="margin: 0;">

调用返回 `long` 类型的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `int64` 方法返回的长整数值
- `error` 错误信息，成功返回 `nil`

```go
timestamp, err := instance.CallLong("getTimestamp")
if err != nil {
    fmt.Println("调用失败:", err)
} else {
    fmt.Printf("时间戳: %d\n", timestamp)
}
```

## CallFloat
<hr style="margin: 0;">

调用返回 `float` 类型的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `float32` 方法返回的浮点数值
- `error` 错误信息，成功返回 `nil`

```go
score, err := instance.CallFloat("getScore")
if err != nil {
    fmt.Println("调用失败:", err)
} else {
    fmt.Printf("分数: %.2f\n", score)
}
```

## CallDouble
<hr style="margin: 0;">

调用返回 `double` 类型的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `float64` 方法返回的双精度浮点数值
- `error` 错误信息，成功返回 `nil`

```go
ratio, err := instance.CallDouble("getRatio")
if err != nil {
    fmt.Println("调用失败:", err)
} else {
    fmt.Printf("比率: %.4f\n", ratio)
}
```

## CallBool
<hr style="margin: 0;">

调用返回 `boolean` 类型的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `bool` 方法返回的布尔值
- `error` 错误信息，成功返回 `nil`

```go
isValid, err := instance.CallBool("validate", "input")
if err != nil {
    fmt.Println("调用失败:", err)
} else if isValid {
    fmt.Println("验证通过")
}
```

## CallString
<hr style="margin: 0;">

调用返回 `String` 类型的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `string` 方法返回的字符串
- `error` 错误信息，成功返回 `nil`

```go
name, err := instance.CallString("getName")
if err != nil {
    fmt.Println("调用失败:", err)
} else {
    fmt.Printf("名称: %s\n", name)
}
```

## CallVoid
<hr style="margin: 0;">

调用无返回值的方法。

- `methodName` {string} 方法名称
- `args` {...interface{}} 方法参数（可选）

**返回值：**
- `error` 错误信息，成功返回 `nil`

```go
err := instance.CallVoid("initialize", "config.xml")
if err != nil {
    fmt.Println("初始化失败:", err)
}
```

## Release (ClassLoader)
<hr style="margin: 0;">

释放类加载器占用的资源。调用后不能再使用该类加载器创建新实例。

```go
cl := plugin.LoadApk("/path/to/plugin.apk")
if cl != nil {
    defer cl.Release()
}
```

## Release (Instance)
<hr style="margin: 0;">

释放实例占用的资源。调用后不能再使用该实例调用方法。

```go
instance, err := cl.NewInstance("com.example.MyClass")
if err == nil {
    defer instance.Release()
}
```

## 完整示例
<hr style="margin: 0;">

```go
package main

import (
    "fmt"
    "github.com/Dasongzi1366/AutoGo/plugin"
    "github.com/Dasongzi1366/AutoGo/images"
)

func main() {
    // 加载APK插件
    cl := plugin.LoadApk("/data/local/tmp/assets/ocr_plugin.apk")
    if cl == nil {
        fmt.Println("加载APK失败")
        return
    }
    defer cl.Release()

    // 创建OCR引擎实例
    engine, err := cl.NewInstance("com.ddddocr.OCREngine")
    if err != nil {
        fmt.Printf("创建实例失败: %v\n", err)
        return
    }
    defer engine.Release()

    // 截取屏幕
    img := images.CaptureScreen(0, 0, 0, 0, 0)
    
    // 调用OCR识别（图像自动转为Bitmap）
    result, err := engine.CallString("recognition", img)
    if err != nil {
        fmt.Printf("识别失败: %v\n", err)
        return
    }
    
    fmt.Printf("识别结果: %s\n", result)
    
    // 调用需要Context的方法
    err = engine.CallVoid("saveToCache", plugin.Context{}, "data.txt")
    if err != nil {
        fmt.Printf("保存失败: %v\n", err)
    }
    
    // 使用AssetManager加载APK中的资源
    assetMgr := plugin.NewAssetManager()
    resourceData, err := engine.CallString("loadResource", assetMgr, "config.json")
    if err != nil {
        fmt.Printf("加载资源失败: %v\n", err)
    } else {
        fmt.Printf("资源内容: %s\n", resourceData)
    }
}
```

## 类型转换说明
<hr style="margin: 0;">

### Go类型 → Java类型映射

| Go类型 | Java类型 | 说明 |
|--------|----------|------|
| `int`, `int32` | `int` | 32位整数 |
| `int64` | `long` | 64位整数，**注意必须显式使用int64** |
| `float32` | `float` | 单精度浮点数 |
| `float64` | `double` | 双精度浮点数 |
| `bool` | `boolean` | 布尔值 |
| `string` | `String` | 字符串 |
| `[]byte` | `byte[]` | 字节数组 |
| `*image.NRGBA` | `android.graphics.Bitmap` | 图像（自动转换，自动释放） |
| `plugin.Context` | `android.content.Context` | 应用上下文（自动获取） |
| `plugin.AssetManager` | `android.content.res.AssetManager` | 资源管理器（自动创建并添加APK路径） |

---

### system - 系统

来源：https://autogo.cc/API/system.md

# system - 系统功能
---
提供与系统相关的功能，包括进程管理和资源查询。

## GetPid
<hr style="margin: 0;">

获取指定进程的 PID。

- `processName` {string} 进程名。如果为空则返回当前进程的 PID。

```go
pid := system.GetPid("com.example.app")
fmt.Println("Process ID:", pid)
```

## GetMemoryUsage
<hr style="margin: 0;">

获取指定进程的内存使用量。

- `pid` {int} 进程的 PID。如果为 0 则获取当前进程的内存使用量。

```go
memoryUsage := system.GetMemoryUsage(12345)
fmt.Println("Memory Usage (KB):", memoryUsage)
```

## GetCpuUsage
<hr style="margin: 0;">

获取指定进程的 CPU 使用率。

- `pid` {int} 进程的 PID。如果为 0 则获取当前进程的 CPU 使用率。

```go
cpuUsage := system.GetCpuUsage(12345)
fmt.Println("CPU Usage (%):", cpuUsage)
```

## RestartSelf
<hr style="margin: 0;">

重启当前脚本进程。

```go
system.RestartSelf()
```

## SetBootStart
<hr style="margin: 0;">

设置脚本开机自动运行，需要root权限。

```go
system.SetBootStart(true)
```

---

### utils - 工具函数

来源：https://autogo.cc/API/utils.md

# utils - 工具函数
---
提供一组常用的工具函数，包括日志记录、字符串与数据类型的转换、随机数生成等功能。

## Shell
<hr style="margin: 0;">

执行 shell 命令并返回输出。

- `cmd` {string} 要执行的命令。

```go
output := utils.Shell("ls -l")
fmt.Println("Command Output:", output)
```

## Toast
<hr style="margin: 0;">

显示 Toast 提示信息。

- `message` {string} 要显示的提示信息。
- `x` {int} 在界面上显示的 X 坐标（传递-1使用默认坐标）。
- `y` {int} 在界面上显示的 Y 坐标（传递-1使用默认坐标）。
- `duration` {int} 提示显示的持续时间，单位为毫秒（传递-1使用默认2000毫秒）。

```go
utils.Toast("Hello AutoGo", -1, -1, -1)
```

## Alert
<hr style="margin: 0;">

显示带标题、内容和按钮的弹窗，阻塞等待用户点击后返回按钮索引，安卓15及以上只有在APK模式下才能正常弹出。

- `title` {string} 弹窗标题。
- `content` {string} 弹窗内容。
- `btn1Text` {string} 第一个按钮的文字(通常为"取消")，输入空字符串默认不显示该按钮。
- `btn2Text` {string} 第二个按钮的文字(通常为"确定")。

```go
btnIndex := utils.Alert("确认操作", "重置脚本数据？", "取消", "确认")
if btnIndex == 1 {
	fmt.Println("用户点击了确认")
} else {
	fmt.Println("用户点击了取消")
}
```

## InputAlert
<hr style="margin: 0;">

弹出带输入框的对话框，阻塞等待用户操作，此方法仅在APK安装模式运行时有效。

- `title` {string} 弹窗标题。
- `content` {string} 说明文字。
- `placeholder` {string} 输入框占位提示。
- `defaultText` {string} 输入框默认文本。
- `btn1Text` {string} 取消类按钮文字。
- `btn2Text` {string} 确认类按钮文字；为空则只显示一个按钮。

**返回** `(string, bool)`：点击确认返回 `(输入内容, true)`，点击取消返回 `("", false)`。

```go
text, ok := utils.InputAlert("输入", "请输入备注", "备注", "", "取消", "确定")
if ok {
	fmt.Println("输入:", text)
}
```

## Sleep
<hr style="margin: 0;">

让当前线程暂停执行指定的时间。

- `i` {int} 暂停时间（毫秒）。

```go
utils.Sleep(500) // 暂停 500 毫秒
```

## Random
<hr style="margin: 0;">

返回指定范围内的真随机整数，包含最小值和最大值。

- `min` {int} 最小值。
- `max` {int} 最大值。

```go
randNum := utils.Random(1, 10)
fmt.Println("Random Number:", randNum)
```

## LogI
<hr style="margin: 0;">

记录一条 INFO 级别的日志。

- `label` {string} 日志标签，用于标识日志类别。
- `message` {...interface{}} 日志消息，描述具体的日志内容。

```go
utils.LogI("AppStart", "Application has started successfully.")
```

## LogE
<hr style="margin: 0;">

记录一条 ERROR 级别的日志。

- `label` {string} 日志标签，用于标识日志类别。
- `message` {...interface{}} 日志消息，描述具体的日志内容。

```go
utils.LogE("AppCrash", "Application encountered an unexpected error.")
```

## I2s
<hr style="margin: 0;">

将整数转换为字符串。

- `i` {int} 要转换的整数。

```go
str := utils.I2s(123)
fmt.Println("String Value:", str)
```

## S2i
<hr style="margin: 0;">

将字符串转换为整数。

- `s` {string} 要转换的字符串。

```go
num := utils.S2i("123")
fmt.Println("Integer Value:", num)
```

## F2s
<hr style="margin: 0;">

将浮点数转换为字符串。

- `f` {float64} 要转换的浮点数。

```go
str := utils.F2s(123.45)
fmt.Println("String Value:", str)
```

## S2f
<hr style="margin: 0;">

将字符串转换为浮点数。如果转换失败返回 0.0。

- `s` {string} 要转换的字符串。

```go
num := utils.S2f("123.45")
fmt.Println("Float Value:", num)
```

## B2s
<hr style="margin: 0;">

将布尔值转换为字符串 ("true" 或 "false")。

- `b` {bool} 要转换的布尔值。

```go
str := utils.B2s(true)
fmt.Println("Boolean as String:", str)
```

## S2b
<hr style="margin: 0;">

将字符串转换为布尔值。如果无法转换则返回 false。

- `s` {string} 要转换的字符串。

```go
boolVal := utils.S2b("true")
fmt.Println("Boolean Value:", boolVal)
```

---

### ppocr - 飞浆OCR

来源：https://autogo.cc/API/ppocr.md

# ppocr - 飞桨OCR
---
提供基于 PaddleOCR 的文字检测和识别功能。

以下是 `ppocr` 包中定义的 `Result` 结构体及其字段说明：

| **字段名**  | **类型**    | **说明**                              |
|-------------|-------------|---------------------------------------|
| `X`         | `int`       | 检测结果的左上角 X 坐标。             |
| `Y`         | `int`       | 检测结果的左上角 Y 坐标。             |
| `Width`     | `int`       | 检测结果的宽度。                      |
| `Height`    | `int`       | 检测结果的高度。                      |
| `Label`     | `string`    | 检测到的文字内容或标签。              |
| `Score`     | `float64`   | 检测结果的置信度，取值范围为 0-1。   |
| `CenterX`   | `int`       | 检测结果的中心 X 坐标。               |
| `CenterY`   | `int`       | 检测结果的中心 Y 坐标。               |

## New
<hr style="margin: 0;">

创建一个 Ppocr 实例对象。成功返回实例对象`*Ppocr`，如果加载失败则返回 nil。

- `version` {string} 模型版本，目前仅支持`v2`和`v5`

```go
ocr := ppocr.New("v5")
if ocr == nil {
    fmt.Println("初始化失败")
    return
}
fmt.Println("初始化成功")
```

## *Ppocr.Ocr
<hr style="margin: 0;">

在屏幕指定区域进行OCR文字识别。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"，"-" 后表示偏色范围，如果不需要指定则直接传入空字符串`""`
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
results := ocr.Ocr(0, 0, 0, 0, "000000", 0) // 识别主屏幕全屏的黑色文字
for _, result := range results {
    fmt.Println(result.Label) // 打印识别到的文字
}
```

## *Ppocr.OcrFromImage
<hr style="margin: 0;">

从图像对象进行OCR文字识别。

- `img` {*image.NRGBA} 要识别的图像对象
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"

```go
img := images.ReadFromPath("/sdcard/screenshot.png")
results := ocr.OcrFromImage(img, "000000")
```

## *Ppocr.OcrFromBase64
<hr style="margin: 0;">

从Base64编码的图像数据进行OCR文字识别。

- `b64` {string} Base64编码的图像数据
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"

```go
results := ocr.OcrFromBase64(base64String, "000000")
```

## *Ppocr.OcrFromPath
<hr style="margin: 0;">

从图像文件路径进行OCR文字识别。

- `path` {string} 图像文件路径
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"

```go
results := ocr.OcrFromPath("/sdcard/screenshot.png", "000000")
```

## *Ppocr.Close
<hr style="margin: 0;">

关闭 PPOCR 实例。

```go
ocr.Close()
```

---

### yolo - 目标检测

来源：https://autogo.cc/API/yolo.md

# yolo - 目标检测
---
提供基于 YOLO 的目标检测功能。

以下是 `yolo` 包中定义的 `Result` 结构体及其字段说明：

| **字段名**  | **类型**    | **说明**                              |
|-------------|-------------|---------------------------------------|
| `X`         | `int`       | 检测结果的左上角 X 坐标。             |
| `Y`         | `int`       | 检测结果的左上角 Y 坐标。             |
| `Width`     | `int`       | 检测结果的宽度。                      |
| `Height`    | `int`       | 检测结果的高度。                      |
| `Label`     | `string`    | 检测到的文字内容或标签。              |
| `Score`     | `float64`   | 检测结果的置信度，取值范围为 0-1。   |
| `CenterX`   | `int`       | 检测结果的中心 X 坐标。               |
| `CenterY`   | `int`       | 检测结果的中心 Y 坐标。               |

## New
<hr style="margin: 0;">

创建一个 Yolo 实例对象。成功返回实例对象`*Yolo`，如果加载失败则返回 nil。

- `version` {string} 模型版本，目前仅支持`v5`和`v8`
- `cpuThreadNum` {int} 用于模型推理的 CPU 线程数。
- `paramPath` {string} 模型参数文件路径。
- `binPath` {string} 模型二进制文件路径。
- `labels` {string} 标签文本，多个标签使用`,`进行隔开。

```go
yolo := yolo.New("v8", 4, "/data/local/tmp/param", "/data/local/tmp/bin", "person,bicycle,car")
if yolo == nil {
    fmt.Println("模型加载失败")
    return
}
fmt.Println("模型加载成功")
```

## *Yolo.Detect
<hr style="margin: 0;">

在屏幕指定区域进行目标检测。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

```go
results := detector.Detect(0, 0, 0, 0, 0) // 在主屏幕全屏范围内检测目标
for _, result := range results {
    fmt.Printf("检测到 %s，置信度: %.2f\n", result.Label, result.Score)
}
```

## *Yolo.DetectFromImage
<hr style="margin: 0;">

从图像对象进行目标检测。

- `img` {*image.NRGBA} 要检测的图像对象

```go
img := images.ReadFromPath("/sdcard/photo.jpg")
results := detector.DetectFromImage(img)
```

## *Yolo.DetectFromBase64
<hr style="margin: 0;">

从Base64编码的图像数据进行目标检测。

- `b64` {string} Base64编码的图像数据
- `colorStr` {string} 保留参数，可以传入空字符串

```go
results := detector.DetectFromBase64(base64String, "")
```

## *Yolo.DetectFromPath
<hr style="margin: 0;">

从图像文件路径进行目标检测。

- `path` {string} 图像文件路径
- `colorStr` {string} 保留参数，可以传入空字符串

```go
results := detector.DetectFromPath("/sdcard/photo.png", "")
```

## *Yolo.Close
<hr style="margin: 0;">

关闭 YOLO 模型实例，释放相关资源。

```go
yolo.Close()
```

---

### apkctl - APK外壳控制

来源：https://autogo.cc/API/apkctl.md

# apkctl - APK外壳控制
---
通过 `apkctl` 模块可以从 Go 侧直接调用 APK 内部能力。

## Eval
<hr style="margin: 0;">

执行指定的 JavaScript 脚本，并返回执行结果。

- `contextID` {string} 执行上下文 ID。相同的 `contextID` 会复用同一个 JS 作用域，因此可以在多次 `Eval` 之间保留变量、控件引用和页面状态。
- `script` {string} 要执行的 JavaScript 代码字符串。

默认注入到 JS 作用域中的对象：

- `context` / `service`：当前 APK 内的 `FloatingBallService` 实例。
- `appContext`：Application Context。
- `go`：桥接对象，可调用 `go.send("消息")` 把字符串消息发回 Go。
- `scriptUi`：原生脚本页面控制器，用来打开/关闭一个专门给 JS 使用的空 Activity。
- `uiActivity` / `uiContext` / `uiRoot`：当前脚本页面已经就绪时，可直接拿来创建原生控件。

当前 APK 已内置一些常用原生 View 库，JS 中可直接 `importClass(...)` 使用：

- `Packages.com.google.android.material.button.MaterialButton`
- `Packages.com.google.android.material.textfield.TextInputLayout`
- `Packages.com.google.android.material.textfield.TextInputEditText`
- `Packages.com.google.android.material.card.MaterialCardView`
- `Packages.androidx.recyclerview.widget.RecyclerView`
- `Packages.androidx.swiperefreshlayout.widget.SwipeRefreshLayout`
- `Packages.androidx.viewpager2.widget.ViewPager2`

`scriptUi` 常用方法：

- `scriptUi.open()` 打开原生脚本页面，并默认等待页面就绪。
- `scriptUi.getActivity()` / `scriptUi.getContext()`
- `scriptUi.getRoot()`
- `scriptUi.runOnUiThread(runnable)` 在主线程创建/更新原生控件。
- `scriptUi.clear()` 清空页面上的所有原生控件。
- `scriptUi.close()` 关闭脚本页面。

```go
result := apkctl.Eval("demo", `
importClass(android.widget.Button);
importClass(java.lang.Runnable);
importClass(Packages.com.google.android.material.button.MaterialButton);

if (!scriptUi.open()) {
    throw new Error("ScriptUiActivity not ready");
}

var activity = scriptUi.getActivity();
var root = scriptUi.getRoot();
scriptUi.runOnUiThread(new Runnable({
    run: function () {
        root.removeAllViews();
        var button = new MaterialButton(activity);
        button.setText("点我");
        root.addView(button);
    }
}));

"ok";
`)
fmt.Println(result)
```

## SetCallback
<hr style="margin: 0;">

设置一个接收 JS 主动消息的回调。

当 APK 内执行的 JavaScript 调用 `go.send("你的消息")` 时，Java 层会把这条消息主动推回 Go，然后这里注册的 callback 就会收到：

- `contextID` {string} 发送这条消息的 JS 执行上下文 ID。
- `message` {string} JS 里传给 `go.send(...)` 的原始字符串。

```go
apkctl.SetCallback(func(contextID, message string) {
    fmt.Println("from js:", contextID, message)
})

apkctl.Eval("page", `
go.send("button_clicked")
`)
```

如果没有设置 callback，JS 发来的消息会被接收但不会进一步处理。复杂数据建议 JS 先 `JSON.stringify` 后再发送。

## RegEvent
<hr style="margin: 0;">

注册脚本控制事件回调。

- `EventPause`：脚本暂停事件。
- `EventResume`：脚本恢复事件。
- `EventStop`：脚本停止事件。

```go
apkctl.RegEvent(apkctl.EventPause, func() {
    apkctl.Toast("准备暂停")
})

apkctl.RegEvent(apkctl.EventStop, func() {
    apkctl.Toast("准备停止")
})
```

传 `nil` 可以取消对应事件的回调：

```go
apkctl.RegEvent(apkctl.EventPause, nil)
```

说明：`EventPause`、`EventStop` 会等待回调执行完成后再继续对应操作；`EventResume` 是恢复后的通知事件，不等待回调返回。

---

### imgui - 界面绘制

来源：https://autogo.cc/API/imgui.md

# imgui - 即时模式图形用户界面
---
提供基于 Dear ImGui 的图形用户界面功能。由于 ImGui 方法数量众多，完整的方法列表请参照 [Dear ImGui 官方文档](https://github.com/ocornut/imgui)

## 基础示例

```go
package main

import (
    "fmt"
    "github.com/Dasongzi1366/AutoGo/imgui"
)

func main() {
    // 初始化
    imgui.Init()
    
    // 状态变量
    counter := 0
    showWindow := true
    
    // 主循环
    imgui.Run(func() {
        // 设置窗口
        imgui.SetNextWindowSizeV(imgui.Vec2{X: 500, Y: 400}, imgui.CondOnce)
        imgui.SetNextWindowPosV(imgui.Vec2{X: 100, Y: 100}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})
        
        // 创建窗口
        imgui.BeginV("示例窗口", &showWindow, 0)
        
        // 标题
        imgui.Text("ImGui 示例程序")
        imgui.Separator()
        imgui.Spacing()
        
        // 计数器
        imgui.Text(fmt.Sprintf("计数器: %d", counter))
        
        // 按钮
        if imgui.Button("增加") {
            counter++
        }
        imgui.SameLine()
        if imgui.Button("减少") {
            counter--
        }
        imgui.SameLine()
        if imgui.Button("重置") {
            counter = 0
        }
        
        imgui.Spacing()
        imgui.Separator()
        imgui.Spacing()
        
        // 样式化按钮
        imgui.Text("样式化按钮：")
        imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0})
        if imgui.Button("绿色按钮") {
            // 绿色按钮的操作
        }
        imgui.PopStyleColor()
        
        imgui.SameLine()
        imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{X: 0.8, Y: 0.2, Z: 0.2, W: 1.0})
        if imgui.Button("红色按钮") {
            // 红色按钮的操作
        }
        imgui.PopStyleColor()
        
        // 结束窗口
        imgui.End()
    })

    // 阻塞主进程防止程序退出
	select {}
}
```

---

### console - 控制台

来源：https://autogo.cc/API/console.md

# console - 控制台
---
提供用于控制台悬浮窗的控制接口，支持多实例、位置、大小、颜色设置以及内容打印等功能。

## New
<hr style="margin: 0;">

创建一个新的控制台实例。

**返回** {*Console} 控制台实例指针

```go
c := console.New()
```

## SetWindowSize
<hr style="margin: 0;">

设置控制台窗口的宽高。

- `width` {int} 控制台窗口的宽度
- `height` {int} 控制台窗口的高度

**返回** {*Console} 控制台实例指针

```go
c.SetWindowSize(800, 600)
```

## SetWindowPosition
<hr style="margin: 0;">

设置控制台窗口的位置。

- `x` {int} 控制台窗口左上角的横坐标
- `y` {int} 控制台窗口左上角的纵坐标

**返回** {*Console} 控制台实例指针

```go
c.SetWindowPosition(100, 200)
```

## SetWindowColor
<hr style="margin: 0;">

设置控制台窗口的背景颜色。

- `color` {string} 背景颜色的十六进制字符串，格式如 "#1E1F22"

**返回** {*Console} 控制台实例指针

```go
c.SetWindowColor("#1E1F22")
```

## SetTextColor
<hr style="margin: 0;">

设置控制台文字颜色。

- `color` {string} 文字颜色的十六进制字符串，格式如 "#FFFFFF"

**返回** {*Console} 控制台实例指针

```go
c.SetTextColor("#FFFFFF")
```

## SetTextSize
<hr style="margin: 0;">

设置控制台文字大小。

- `size` {int} 文字大小

```go
c.SetTextSize(50)
```

## Println
<hr style="margin: 0;">

打印文本到控制台。

- `a` {any} 要打印的参数，支持多个参数，行为类似 fmt.Println

```go
c.Println("Hello, world!")
c.Println("用户ID:", 123, "状态:", "在线")
```

## Clear
<hr style="margin: 0;">

清空控制台内容。

```go
c.Clear()
```

## Show
<hr style="margin: 0;">

显示控制台窗口。

```go
c.Show()
```

## Hide
<hr style="margin: 0;">

隐藏控制台窗口。

```go
c.Hide()
```

## IsVisible
<hr style="margin: 0;">

检查控制台是否可见。

**返回** {bool} true 表示可见，false 表示隐藏

```go
if c.IsVisible() {
    // 控制台当前可见
}
```

## Destroy
<hr style="margin: 0;">

销毁控制台实例，释放资源。

```go
c.Destroy()
```

## 示例
<hr style="margin: 0;">

```go
// 创建并配置控制台
c := console.New()
c.SetWindowPosition(50, 50)
c.SetWindowSize(700, 500)
c.SetWindowColor("#1E1F22")
c.SetTextColor("#00FF00")
c.SetTextSize(45)

// 打印日志
c.Println("控制台已就绪")

for {
    c.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05"))
    utils.Sleep(1000)
}
```

---

### hud - 悬浮显示

来源：https://autogo.cc/API/hud.md

# hud - 悬浮显示
---
提供悬浮显示功能，支持多实例、彩色文本显示等功能。

以下是 `hud` 包中定义的 `TextItem` 结构体及其字段说明：

| **字段名**   | **类型**      | **说明**                                      |
|--------------|---------------|-----------------------------------------------|
| `TextColor`  | `color.Color` | 文字颜色。格式如 `"#FFFFFF"`。 |
| `Text`       | `string`      | 显示的文本内容。                              |

## New
<hr style="margin: 0;">

创建一个新的 HUD 实例。

**返回** {*HUD} HUD 实例指针

```go
h := hud.New()
```

## SetPosition
<hr style="margin: 0;">

设置 HUD 的位置和大小。

- `x1` {int} 左上角横坐标
- `y1` {int} 左上角纵坐标
- `x2` {int} 右下角横坐标
- `y2` {int} 右下角纵坐标

**返回** {*HUD} HUD 实例指针

```go
h.SetPosition(100, 100, 400, 150)
```

## SetBackgroundColor
<hr style="margin: 0;">

设置 HUD 的背景颜色。

- `color` {string} 背景颜色的十六进制字符串，格式如 "#2D2D30" 或 "#2D2D3080"（带透明度）

**返回** {*HUD} HUD 实例指针

```go
h.SetBackgroundColor("#2D2D30")
h.SetBackgroundColor("#00000080")  // 半透明黑色
```

## SetTextSize
<hr style="margin: 0;">

设置 HUD 的字体大小。

- `size` {int} 字体大小（推荐范围：30-60）

**返回** {*HUD} HUD 实例指针

```go
h.SetTextSize(45)
```

## SetText
<hr style="margin: 0;">

设置 HUD 显示的文本内容（支持多色文本）。

- `items` {[]TextItem} 文本项数组，每个元素包含颜色和文本

**返回** {*HUD} HUD 实例指针

```go
h.SetText([]hud.TextItem{
    {TextColor: "#00FF00", Text: "HP: "},
    {TextColor: "#FFFFFF", Text: "100/100"},
})
```

## Show
<hr style="margin: 0;">

显示 HUD。

**返回** {*HUD} HUD 实例指针

```go
h.Show()
```

## Hide
<hr style="margin: 0;">

隐藏 HUD。

**返回** {*HUD} HUD 实例指针

```go
h.Hide()
```

## IsVisible
<hr style="margin: 0;">

检查 HUD 是否可见。

**返回** {bool} true 表示可见，false 表示隐藏

```go
if h.IsVisible() {
    // HUD 当前可见
}
```

## Destroy
<hr style="margin: 0;">

销毁 HUD 实例，释放资源。

```go
h.Destroy()
```

## 示例
<hr style="margin: 0;">

```go
// 创建并配置 HUD
h := hud.New()
h.SetPosition(50, 700, 450, 750)
h.SetBackgroundColor("#3D000080")
h.SetTextSize(45)

for {
    // 设置多色文本
    h.SetText([]hud.TextItem{
        {TextColor: "#00FF00", Text: "当前时间: "},
        {TextColor: "#FFFFFF", Text: time.Now().Format("2006-01-02 15:04:05")},
    })
    utils.Sleep(1000)
}
```

---

### vdisplay - 虚拟屏幕

来源：https://autogo.cc/API/vdisplay.md

# vdisplay - 虚拟屏幕
---
vdisplay模块提供了虚拟屏幕的创建和管理功能，需要Android 10及以上版本。通过该模块可以创建虚拟屏幕，用于在虚拟屏幕中启动和操作应用程序，实现多应用并行操作和完美全分辨率。

## Create
<hr style="margin: 0;">

创建一个虚拟屏幕。

- `width` {int} 屏幕宽度
- `height` {int} 屏幕高度
- `dpi` {int} 屏幕的像素密度

**返回值：**
- `*Vdisplay` 虚拟屏幕对象，如果创建失败则返回 `nil`

**注意：**
- 可以通过Scrcpy执行命令查看虚拟屏幕的实时画面：`scrcpy --display-id=虚拟屏幕ID`

```go
v := vdisplay.Create(720, 1280, 320)
if v != nil {
    fmt.Printf("创建成功,屏幕ID:%d\n", v.GetDisplayId())
}
```

## GetDisplayId
<hr style="margin: 0;">

获取虚拟屏幕的ID。

```go
displayId := v.GetDisplayId()
fmt.Printf("虚拟屏幕ID: %d\n", displayId)
```

## LaunchApp
<hr style="margin: 0;">

在虚拟屏幕中启动指定的应用程序。

- `packageName` {string} 要启动的应用包名

```go
success := v.LaunchApp("com.example.app")
if success {
    fmt.Println("应用启动成功")
}
```

## SetTitle
<hr style="margin: 0;">

设置虚拟屏预览窗口的标题。

- `title` {string} 标题

```go
v.SetTitle("MT管理器")
```

## SetTouchCallback
<hr style="margin: 0;">

设置一个点击回调。

- `callback` {function} 当用户点击虚拟屏预览窗口时调用的函数，格式为 `func(x, y, action, displayId int)`，如果传入 `nil`，则会移除当前设置的回调

```go
v.SetTouchCallback(func(x, y, action, displayId int) {
    // 处理用户点击事件
})
```

## ShowPreviewWindow
<hr style="margin: 0;">

显示虚拟屏幕的预览窗口。预览窗口可以在ImgUI界面中显示虚拟屏幕的实时画面，并支持触摸操作。

- `rotated` {bool} 是否将预览窗口旋转90度显示(显示横屏游戏画面时可能需要)

```go
v.ShowPreviewWindow(false)
```

## HidePreviewWindow
<hr style="margin: 0;">

隐藏虚拟屏幕的预览窗口。

```go
v.HidePreviewWindow()
```

## SetPreviewWindowSize
<hr style="margin: 0;">

设置预览窗口的大小。

- `width` {int} 预览窗口宽度（像素）
- `height` {int} 预览窗口高度（像素）

```go
v.SetPreviewWindowSize(800, 600)
```

## SetPreviewWindowPos
<hr style="margin: 0;">

设置预览窗口的位置。

- `x` {int} 预览窗口X坐标
- `y` {int} 预览窗口Y坐标

```go
v.SetPreviewWindowPos(100, 100)
```

## Destroy
<hr style="margin: 0;">

销毁虚拟屏幕，释放相关资源。

```go
v.Destroy()
```

---

## iOS API 文档

### app - 应用

来源：https://autogo.cc/ios/app.md

# app - 应用
---
app 模块用于启动与管理应用、查询安装信息、打开链接等。脚本中的包名字段对应应用的 **Bundle Identifier**。

以下是 `app` 包中定义的 `AppInfo` 结构体及其字段说明：

| **字段名** | **类型** | **说明** |
|-----------|---------|---------|
| `PackageName` | `string` | 应用的 Bundle Identifier。 |
| `AppName` | `string` | 应用显示名称。 |
| `VersionName` | `string` | 版本号（如 CFBundleShortVersionString）。 |
| `VersionCode` | `string` | 构建号（如 CFBundleVersion）。 |
| `IsSystemApp` | `bool` | 是否为系统应用。 |

## CurrentPackage
<hr style="margin: 0;">

获取当前前台应用的 Bundle Identifier。

```go
bundleID := app.CurrentPackage()
```

## Launch
<hr style="margin: 0;">

通过 Bundle Identifier 启动应用。

- `packageName` {string} 应用 Bundle ID。

**返回** {bool} 是否启动成功。

```go
ok := app.Launch("com.apple.mobilesafari")
```

## ForceStop
<hr style="margin: 0;">

关闭指定应用（系统级终止进程）。

- `packageName` {string} 应用 Bundle ID。

```go
app.ForceStop("com.example.app")
```

## GetList
<hr style="margin: 0;">

获取已安装应用列表。

- `includeSystemApps` {bool} 是否包含系统应用。

```go
list := app.GetList(true)
for _, a := range list {
	fmt.Println(a.PackageName, a.AppName)
}
```

## GetBundlePath
<hr style="margin: 0;">

获取应用安装包(.app)路径

- `packageName` {string} 应用 Bundle ID。

```go
path := app.GetBundlePath("com.auto.go")
```

## SelfPackage
<hr style="margin: 0;">

获取当前脚本/宿主自身的包名。

```go
pkg := app.SelfPackage()
```

## GetName
<hr style="margin: 0;">

通过 Bundle ID 获取应用显示名称。

- `packageName` {string} 应用 Bundle ID。

```go
name := app.GetName("com.tencent.xin")
```

## GetVersion
<hr style="margin: 0;">

获取应用版本号字符串。

- `packageName` {string} 应用 Bundle ID。

```go
ver := app.GetVersion("com.tencent.xin")
```

## GetIcon
<hr style="margin: 0;">

获取应用图标 PNG 二进制数据。

- `packageName` {string} 应用 Bundle ID。

**返回** {[]byte} PNG 数据。

```go
data := app.GetIcon("com.apple.mobilesafari")
```

## IsInstalled
<hr style="margin: 0;">

判断是否已安装指定应用。

- `packageName` {string} 应用 Bundle ID。

```go
if app.IsInstalled("com.example.app") {
	fmt.Println("已安装")
}
```

## Uninstall
<hr style="margin: 0;">

卸载应用。

- `packageName` {string} 应用 Bundle ID。

```go
app.Uninstall("com.example.app")
```

## Install
<hr style="margin: 0;">

安装 IPA 文件。

- `path` {string} IPA 文件路径。

```go
app.Install("/var/mobile/Documents/app.ipa")
```

## Clear
<hr style="margin: 0;">

清除应用数据。

- `packageName` {string} 应用 Bundle ID。

```go
app.Clear("com.example.app")
```

## OpenUrl
<hr style="margin: 0;">

使用系统默认方式打开 URL。若未带 `http://` 或 `https://` 前缀，会自动补上 `http://`。

- `url` {string} 网址或 URL Scheme。

```go
app.OpenUrl("https://example.com")
```

---

### device - 设备

来源：https://autogo.cc/ios/device.md

# device - 设备
---
device 模块提供与设备相关的信息与操作，例如屏幕参数、电量、内存、网络地址、亮度、唤醒与重启等。

以下是 `device` 包中定义的设备信息变量：

| **变量名** | **类型**   | **说明** |
|-----------|-----------|---------|
| `Model`   | `string`  | 设备机型标识。 |
| `Release` | `string`  | 系统版本，如 `"16.5"`。 |
| `Serial`  | `string`  | 设备序列号。 |

## GetDisplayInfo
<hr style="margin: 0;">

获取主屏幕的分辨率信息。

**返回** `width` {int} 逻辑宽度，`height` {int} 逻辑高度，`scale` {float64} 屏幕缩放因子，`rotation` {int} 旋转角度。

```go
width, height, scale, rotation := device.GetDisplayInfo()
fmt.Printf("屏幕: %dx%d, scale: %v, rotation: %d\n", width, height, scale, rotation)
```

## VpnStatus
<hr style="margin: 0;">

获取VPN状态(开启或关闭)

```go
status := device.VpnStatus()
```

## VpnOn
<hr style="margin: 0;">

开启VPN

```go
device.VpnOn()
```

## VpnOff
<hr style="margin: 0;">

关闭VPN

```go
device.VpnOff()
```

## GetBattery
<hr style="margin: 0;">

获取当前电量百分比。

```go
battery := device.GetBattery()
```

## GetBatteryStatus
<hr style="margin: 0;">

获取电池状态：`0` 未知，`1` 未充电，`2` 充电中，`3` 已充满。

```go
status := device.GetBatteryStatus()
```

## IsScreenOn
<hr style="margin: 0;">

判断屏幕是否点亮。

```go
isOn := device.IsScreenOn()
```

## IsScreenUnlock
<hr style="margin: 0;">

判断屏幕是否已解锁（未处于锁屏状态）。

```go
isUnlock := device.IsScreenUnlock()
```

## GetBrightness
<hr style="margin: 0;">

获取当前屏幕亮度，范围为 0~255。

```go
brightness := device.GetBrightness()
```

## GetTotalMem
<hr style="margin: 0;">

获取设备总内存，单位 KB。

```go
totalMem := device.GetTotalMem()
```

## GetAvailMem
<hr style="margin: 0;">

获取设备当前可用内存，单位 KB。

```go
availMem := device.GetAvailMem()
```

## WakeUp
<hr style="margin: 0;">

唤醒设备（点亮屏幕等）。

```go
device.WakeUp()
```

## KeepScreenOn
<hr style="margin: 0;">

保持屏幕常亮。

```go
device.KeepScreenOn()
```

## GetIp
<hr style="margin: 0;">

获取设备局域网 IPv4 地址（优先返回私有网段地址）。

```go
ip := device.GetIp()
```

## GetWifiMac
<hr style="margin: 0;">

获取 Wi‑Fi 网卡 MAC 地址（常见为 `en0` 接口）。

```go
wifiMac := device.GetWifiMac()
```

## Reboot
<hr style="margin: 0;">

重启设备。

```go
device.Reboot()
```

---

### files - 文件系统

来源：https://autogo.cc/ios/files.md

# files - 文件操作
---
提供文件和文件夹的操作接口，例如读取、写入、移动等。

## IsFile
<hr style="margin: 0;">

判断路径是否是文件。

- `path` {string} 路径

```go
isFile := files.IsFile("/sdcard/example.txt")
```

## IsDir
<hr style="margin: 0;">

判断路径是否是文件夹。

- `path` {string} 路径

```go
isDir := files.IsDir("/sdcard/example_folder")
```

## IsEmptyDir
<hr style="margin: 0;">

判断文件夹是否为空。如果路径不是文件夹，返回 false。

- `path` {string} 文件夹路径

```go
isEmpty := files.IsEmptyDir("/sdcard/example_folder")
```

## Create
<hr style="margin: 0;">

创建文件或文件夹。如果文件已存在，返回 true。

- `path` {string} 路径

```go
success := files.Create("/sdcard/new_file.txt")
```

## Exists
<hr style="margin: 0;">

判断路径是否存在。

- `path` {string} 路径

```go
exists := files.Exists("/sdcard/example.txt")
```

## EnsureDir
<hr style="margin: 0;">

确保文件夹存在，如果不存在则创建。

- `path` {string} 路径

```go
success := files.EnsureDir("/sdcard/new_folder")
```

## Read
<hr style="margin: 0;">

读取文本文件的内容。

- `path` {string} 文件路径

```go
content := files.Read("/sdcard/example.txt")
```

## ReadBytes
<hr style="margin: 0;">

读取文件的字节数据。

- `path` {string} 文件路径

```go
data := files.ReadBytes("/sdcard/example.txt")
```

## Write
<hr style="margin: 0;">

将文本写入文件。如果文件不存在则创建，存在则覆盖。

- `path` {string} 文件路径
- `text` {string} 要写入的文本

```go
files.Write("/sdcard/example.txt", "Hello, World!")
```

## WriteBytes
<hr style="margin: 0;">

将字节数据写入文件。如果文件不存在则创建，存在则覆盖。

- `path` {string} 文件路径
- `bytes` {[]byte} 要写入的字节数据

```go
files.WriteBytes("/sdcard/example.txt", []byte("Hello, World!"))
```

## Append
<hr style="margin: 0;">

将文本追加到文件末尾。如果文件不存在则创建。

- `path` {string} 文件路径
- `text` {string} 要追加的文本

```go
files.Append("/sdcard/example.txt", "Appended text")
```

## AppendBytes
<hr style="margin: 0;">

将字节数据追加到文件末尾。如果文件不存在则创建。

- `path` {string} 文件路径
- `bytes` {[]byte} 要追加的字节数据

```go
files.AppendBytes("/sdcard/example.txt", []byte("Appended bytes"))
```

## Copy
<hr style="margin: 0;">

复制文件。

- `fromPath` {string} 源文件路径
- `toPath` {string} 目标文件路径

```go
success := files.Copy("/sdcard/source.txt", "/sdcard/destination.txt")
```

## Move
<hr style="margin: 0;">

移动文件。

- `fromPath` {string} 源文件路径
- `toPath` {string} 目标文件路径

```go
success := files.Move("/sdcard/source.txt", "/sdcard/destination.txt")
```

## Rename
<hr style="margin: 0;">

重命名文件。

- `path` {string} 文件路径
- `newName` {string} 新文件名

```go
success := files.Rename("/sdcard/example.txt", "new_name.txt")
```

## GetName
<hr style="margin: 0;">

获取文件名。

- `path` {string} 文件路径

```go
name := files.GetName("/sdcard/example.txt")
```

## GetNameWithoutExtension
<hr style="margin: 0;">

获取不含扩展名的文件名。

- `path` {string} 文件路径

```go
name := files.GetNameWithoutExtension("/sdcard/example.txt")
```

## GetExtension
<hr style="margin: 0;">

获取文件的扩展名。

- `path` {string} 文件路径

```go
extension := files.GetExtension("/sdcard/example.txt")
```

## GetMd5
<hr style="margin: 0;">

获取文件的MD5值。

- `path` {string} 文件路径

```go
md5 := files.GetMd5("/sdcard/example.txt")
```

## Remove
<hr style="margin: 0;">

删除文件或文件夹。如果是文件夹，则删除其所有内容。

- `path` {string} 文件路径或文件夹路径

```go
success := files.Remove("/sdcard/example.txt")
```

## Path
<hr style="margin: 0;">

将相对路径转换为绝对路径。

- `relativePath` {string} 相对路径

```go
absolutePath := files.Path("./example.txt")
```

## ListDir
<hr style="margin: 0;">

列出文件夹下的所有文件和文件夹。

- `path` {string} 文件夹路径

```go
entries := files.ListDir("/sdcard/example_folder")
```

---

### uiacc - 节点操作

来源：https://autogo.cc/ios/uiacc.md

# uiacc - 节点操作（iOS）
---

提供基于 iOS 辅助功能（Accessibility）的控件定位与基础交互能力。

## Rect
<hr style="margin: 0;">

`uiacc.Rect` 与 Android 版字段对齐，**单位为物理像素**（px）。

| **字段名**  | **类型** | **说明** |
|-------------|----------|----------|
| `Left`      | `int`    | 左边界 |
| `Right`     | `int`    | 右边界 |
| `Top`       | `int`    | 上边界 |
| `Bottom`    | `int`    | 下边界 |
| `CenterX`   | `int`    | 中心 X |
| `CenterY`   | `int`    | 中心 Y |
| `Width`     | `int`    | 宽度 |
| `Height`    | `int`    | 高度 |

## New
<hr style="margin: 0;">

创建一个 `*Uiacc` 实例。

```go
acc := uiacc.New()
defer acc.Release()
```

## *Uiacc.Release
<hr style="margin: 0;">

释放当前 `Uiacc` 对象持有的引用。

```go
acc := uiacc.New()
acc.Release()
```

## 选择器（Selector）
<hr style="margin: 0;">

选择器方法会返回 `*Uiacc`，支持链式调用，最终通常配合 `WaitFor` / `FindOnce` / `Find` 使用。

### 文本类

## *Uiacc.Text
<hr style="margin: 0;">

设置选择器的 `text` 属性。

- `v` {string} 文本值

```go
obj := acc.Text("登录").FindOnce()
```

## *Uiacc.TextContains
<hr style="margin: 0;">

设置选择器的 `textContains` 属性。

- `v` {string} 包含的文本片段

```go
obj := acc.TextContains("登录").FindOnce()
```

## *Uiacc.TextStartsWith
<hr style="margin: 0;">

设置选择器的 `textStartsWith` 属性。

- `v` {string} 前缀

```go
obj := acc.TextStartsWith("登").FindOnce()
```

## *Uiacc.TextEndsWith
<hr style="margin: 0;">

设置选择器的 `textEndsWith` 属性。

- `v` {string} 后缀

```go
obj := acc.TextEndsWith("录").FindOnce()
```

## *Uiacc.TextMatches
<hr style="margin: 0;">

设置选择器的 `textMatches` 属性（正则匹配）。

- `pattern` {string} 正则表达式

```go
obj := acc.TextMatches("^登.*录$").FindOnce()
```

## *Uiacc.Desc
<hr style="margin: 0;">

设置选择器的 `desc` 属性。

- `v` {string} 描述文本

```go
obj := acc.Desc("确认").FindOnce()
```

## *Uiacc.DescContains
<hr style="margin: 0;">

设置选择器的 `descContains` 属性。

- `v` {string} 描述包含片段

```go
obj := acc.DescContains("确认").FindOnce()
```

## *Uiacc.DescStartsWith
<hr style="margin: 0;">

设置选择器的 `descStartsWith` 属性。

- `v` {string} 描述前缀

```go
obj := acc.DescStartsWith("确").FindOnce()
```

## *Uiacc.DescEndsWith
<hr style="margin: 0;">

设置选择器的 `descEndsWith` 属性。

- `v` {string} 描述后缀

```go
obj := acc.DescEndsWith("认").FindOnce()
```

## *Uiacc.DescMatches
<hr style="margin: 0;">

设置选择器的 `descMatches` 属性（正则匹配）。

- `pattern` {string} 正则表达式

```go
obj := acc.DescMatches("^确.*认$").FindOnce()
```

## *Uiacc.Id
<hr style="margin: 0;">

设置选择器的 `id` 属性。

- `v` {string} id 值

```go
obj := acc.Id("login_button").FindOnce()
```

## *Uiacc.IdContains
<hr style="margin: 0;">

设置选择器的 `idContains` 属性。

- `v` {string} id 包含片段

```go
obj := acc.IdContains("login").FindOnce()
```

## *Uiacc.IdStartsWith
<hr style="margin: 0;">

设置选择器的 `idStartsWith` 属性。

- `v` {string} id 前缀

```go
obj := acc.IdStartsWith("login_").FindOnce()
```

## *Uiacc.IdEndsWith
<hr style="margin: 0;">

设置选择器的 `idEndsWith` 属性。

- `v` {string} id 后缀

```go
obj := acc.IdEndsWith("_button").FindOnce()
```

## *Uiacc.IdMatches
<hr style="margin: 0;">

设置选择器的 `idMatches` 属性（正则匹配）。

- `pattern` {string} 正则表达式

```go
obj := acc.IdMatches("^login_.*$").FindOnce()
```

## *Uiacc.ClassName
<hr style="margin: 0;">

设置选择器的 `className` 属性。

- `v` {string} 类名

```go
obj := acc.ClassName("UIButton").FindOnce()
```

## *Uiacc.ClassNameContains
<hr style="margin: 0;">

设置选择器的 `classNameContains` 属性。

- `v` {string} 类名包含片段

```go
obj := acc.ClassNameContains("Button").FindOnce()
```

## *Uiacc.ClassNameStartsWith
<hr style="margin: 0;">

设置选择器的 `classNameStartsWith` 属性。

- `v` {string} 类名前缀

```go
obj := acc.ClassNameStartsWith("UI").FindOnce()
```

## *Uiacc.ClassNameEndsWith
<hr style="margin: 0;">

设置选择器的 `classNameEndsWith` 属性。

- `v` {string} 类名后缀

```go
obj := acc.ClassNameEndsWith("Button").FindOnce()
```

## *Uiacc.ClassNameMatches
<hr style="margin: 0;">

设置选择器的 `classNameMatches` 属性（正则匹配）。

- `pattern` {string} 正则表达式

```go
obj := acc.ClassNameMatches("^UI.*Button$").FindOnce()
```

## *Uiacc.PackageName
<hr style="margin: 0;">

设置选择器的 `packageName` 属性。

- `v` {string} 包名

```go
obj := acc.PackageName("com.example.ios").FindOnce()
```

## *Uiacc.PackageNameContains
<hr style="margin: 0;">

设置选择器的 `packageNameContains` 属性。

- `v` {string} 包名包含片段

```go
obj := acc.PackageNameContains("example").FindOnce()
```

## *Uiacc.PackageNameStartsWith
<hr style="margin: 0;">

设置选择器的 `packageNameStartsWith` 属性。

- `v` {string} 包名前缀

```go
obj := acc.PackageNameStartsWith("com.example").FindOnce()
```

## *Uiacc.PackageNameEndsWith
<hr style="margin: 0;">

设置选择器的 `packageNameEndsWith` 属性。

- `v` {string} 包名后缀

```go
obj := acc.PackageNameEndsWith(".ios").FindOnce()
```

## *Uiacc.PackageNameMatches
<hr style="margin: 0;">

设置选择器的 `packageNameMatches` 属性（正则匹配）。

- `pattern` {string} 正则表达式

```go
obj := acc.PackageNameMatches("^com\\.example\\..*$").FindOnce()
```

### bounds（物理像素）

## *Uiacc.Bounds
<hr style="margin: 0;">

设置选择器的 `bounds` 属性。

- `left, top, right, bottom` {int} 边界（物理像素）

```go
obj := acc.Bounds(0, 0, 100, 100).FindOnce()
```

## *Uiacc.BoundsInside
<hr style="margin: 0;">

设置选择器的 `boundsInside` 属性。

- `left, top, right, bottom` {int} 内部范围（物理像素）

```go
obj := acc.BoundsInside(0, 0, 500, 500).FindOnce()
```

## *Uiacc.BoundsContains
<hr style="margin: 0;">

设置选择器的 `boundsContains` 属性。

- `left, top, right, bottom` {int} 包含范围（物理像素）

```go
obj := acc.BoundsContains(50, 50, 300, 300).FindOnce()
```

### 布尔/数值类

## *Uiacc.Clickable
<hr style="margin: 0;">

设置选择器的 `clickable` 属性。

- `v` {bool}

```go
obj := acc.Clickable(true).FindOnce()
```

## *Uiacc.Selected
<hr style="margin: 0;">

设置选择器的 `selected` 属性。

- `v` {bool}

```go
obj := acc.Selected(true).FindOnce()
```

## *Uiacc.Enabled
<hr style="margin: 0;">

设置选择器的 `enabled` 属性。

- `v` {bool}

```go
obj := acc.Enabled(true).FindOnce()
```

## *Uiacc.Scrollable
<hr style="margin: 0;">

设置选择器的 `scrollable` 属性。

- `v` {bool}

```go
obj := acc.Scrollable(true).FindOnce()
```

## *Uiacc.Editable
<hr style="margin: 0;">

设置选择器的 `editable` 属性。

- `v` {bool}

```go
obj := acc.Editable(true).FindOnce()
```

## *Uiacc.Checked
<hr style="margin: 0;">

设置选择器的 `checked` 属性。

- `v` {bool}

```go
obj := acc.Checked(true).FindOnce()
```

## *Uiacc.Password
<hr style="margin: 0;">

设置选择器的 `password` 属性。

- `v` {bool}

```go
obj := acc.Password(true).FindOnce()
```

## *Uiacc.Index
<hr style="margin: 0;">

设置选择器的 `index` 属性。

- `v` {int}

```go
obj := acc.Index(0).FindOnce()
```

## 查找与等待
<hr style="margin: 0;">

## *Uiacc.Click
<hr style="margin: 0;">

点击屏幕上的文本。iOS 版逻辑为：**先按 `text` 匹配，未命中再按 `desc` 匹配**。

- `text` {string} 目标文本

```go
ok := acc.Click("确定")
```

## *Uiacc.WaitFor
<hr style="margin: 0;">

等待控件出现并返回节点对象。

- `timeout` {int} 超时时间（毫秒）。`0` 表示无限等待，超时返回 `nil`

```go
obj := acc.Text("登录").WaitFor(3000)
```

## *Uiacc.FindOnce
<hr style="margin: 0;">

查找第一个符合条件的节点，返回它在查询瞬间的属性快照（找不到返回 `nil`）。

```go
obj := acc.Text("登录").FindOnce()
```

## *Uiacc.Find
<hr style="margin: 0;">

查找所有符合条件的节点，返回它们在查询瞬间的属性快照数组。

```go
objs := acc.TextContains("登录").Find()
```

## 节点对象（UiObject）
<hr style="margin: 0;">

## *UiObject.Click
<hr style="margin: 0;">

点击该控件，并返回是否点击成功。

```go
obj := acc.Text("登录").WaitFor(3000)
if obj != nil {
  ok := obj.Click()
  fmt.Println("点击成功:", ok)
}
```

## *UiObject.GetBounds / GetBoundsInParent
<hr style="margin: 0;">

获取控件范围，单位为物理像素（px）。

```go
rect := obj.GetBounds()
rect2 := obj.GetBoundsInParent()
fmt.Printf("控件范围: %v\n", rect)
fmt.Printf("控件在父控件中的范围: %v\n", rect2)
```

## *UiObject.GetId
<hr style="margin: 0;">

获取控件的 `accessibilityIdentifier`。

```go
id := obj.GetId()
fmt.Println("控件 ID:", id)
```

## *UiObject.GetDesc
<hr style="margin: 0;">

获取控件的描述内容。iOS 版将 `accessibilityValue` 对齐到 Android 的 `contentDescription` 语义。

```go
desc := obj.GetDesc()
fmt.Println("控件描述内容:", desc)
```

## *UiObject.GetClassName
<hr style="margin: 0;">

获取控件的类名：当 `class2` 为空时回退到 `role`。

```go
cls := obj.GetClassName()
fmt.Println("控件类名:", cls)
```

## *UiObject.GetClickable
<hr style="margin: 0;">

获取控件的 `clickable` 属性。

```go
clickable := obj.GetClickable()
fmt.Println("控件是否可点击:", clickable)
```

## *UiObject.GetSelected
<hr style="margin: 0;">

获取控件的 `selected` 属性。

```go
selected := obj.GetSelected()
fmt.Println("控件是否被选中:", selected)
```

## *UiObject.GetEnabled
<hr style="margin: 0;">

获取控件的 `enabled` 属性。

```go
enabled := obj.GetEnabled()
fmt.Println("控件是否启用:", enabled)
```

## *UiObject.GetScrollable
<hr style="margin: 0;">

获取控件的 `scrollable` 属性。

```go
scrollable := obj.GetScrollable()
fmt.Println("控件是否可滚动:", scrollable)
```

## *UiObject.GetEditable
<hr style="margin: 0;">

获取控件的 `editable` 属性。

```go
editable := obj.GetEditable()
fmt.Println("控件是否可编辑:", editable)
```

## *UiObject.GetChecked
<hr style="margin: 0;">

获取控件的 `checked` 属性。

```go
checked := obj.GetChecked()
fmt.Println("控件是否被勾选:", checked)
```

## *UiObject.GetPassword
<hr style="margin: 0;">

获取控件的 `password` 属性。

```go
password := obj.GetPassword()
fmt.Println("控件是否为密码字段:", password)
```

## *UiObject.GetChildCount
<hr style="margin: 0;">

获取控件的子控件数目。

```go
childCount := obj.GetChildCount()
fmt.Println("子控件数量:", childCount)
```

## *UiObject.GetIndex
<hr style="margin: 0;">

获取控件在父控件中的索引。

```go
index := obj.GetIndex()
fmt.Println("控件在父控件中的索引:", index)
```

## *UiObject.GetText
<hr style="margin: 0;">

获取控件的文本内容。

```go
text := obj.GetText()
fmt.Println("控件文本内容:", text)
```

## *UiObject.GetPackageName
<hr style="margin: 0;">

获取控件的包名。

```go
packageName := obj.GetPackageName()
fmt.Println("控件包名:", packageName)
```

## *UiObject.ToString
<hr style="margin: 0;">

将节点对象转文本。

```go
str := obj.ToString()
fmt.Println("节点文本:", str)
```

---

### https - HTTP网络请求

来源：https://autogo.cc/ios/https.md

# https - 网络请求
---
https模块提供了发送HTTP/HTTPS请求的功能，可用于与网络服务进行交互，获取网页内容，上传文件等。

## Get
<hr style="margin: 0;">

发送GET请求并返回响应状态码和数据。

- `url` {string} 请求的URL
- `timeout` {int} 请求的超时时间（毫秒），如果为0则不设置超时

```go
code, data := https.Get("https://example.com", 5000)
if code == 200 {
    fmt.Println("请求成功：", string(data))
} else {
    fmt.Println("请求失败，状态码：", code)
}
```

## Post
<hr style="margin: 0;">

发送POST请求并返回响应状态码和数据。支持自定义请求头和请求体，适用于发送JSON、XML等格式的数据。

- `url` {string} 请求的URL
- `data` {[]byte} 请求体数据（如JSON序列化后的字节数组）
- `headers` {map[string]string} 自定义请求头，如果为nil或未设置Content-Type，默认使用application/json
- `timeout` {int} 请求的超时时间（毫秒），如果为0则不设置超时

```go
// 构造JSON请求体
reqData := map[string]interface{}{
    "name": "张三",
    "age":  25,
}
jsonData, _ := json.Marshal(reqData)

// 设置请求头
headers := map[string]string{
    "Content-Type":  "application/json",
    "Authorization": "Bearer your-token",
}

// 发送请求
code, data := https.Post("https://example.com/api/user", jsonData, headers, 10000)
if code == 200 {
    fmt.Println("请求成功：", string(data))
} else {
    fmt.Println("请求失败，状态码：", code)
}
``` 

## PostMultipart
<hr style="margin: 0;">

发送带有文件的POST请求（multipart/form-data格式）并返回响应状态码和数据。

- `url` {string} 请求的URL
- `fileName` {string} 文件名
- `fileData` {[]byte} 文件数据（字节数组）
- `timeout` {int} 请求的超时时间（毫秒），如果为0则不设置超时

```go
// 读取文件数据
fileData := files.ReadBytes("/sdcard/image.jpg")
// 发送请求
code, data := https.PostMultipart("https://example.com/upload", "image.jpg", fileData, 10000)
if code == 200 {
    fmt.Println("上传成功：", string(data))
} else {
    fmt.Println("上传失败，状态码：", code)
}
```

---

### ime - 输入法

来源：https://autogo.cc/ios/ime.md

# ime - 输入法与剪贴板
---
ime 模块提供剪贴板读写以及向当前焦点控件输入文本等能力。

## InputText
<hr style="margin: 0;">

向当前焦点控件输入文本。

- `text` {string} 要输入的文本。

```go
ime.InputText("Hello, World!")
```

## GetClipText
<hr style="margin: 0;">

获取剪贴板文本内容。

```go
text := ime.GetClipText()
fmt.Println("剪贴板:", text)
```

## SetClipText
<hr style="margin: 0;">

设置剪贴板文本内容。

- `text` {string} 要写入的文本。

**返回** {bool} 是否设置成功。

```go
ime.SetClipText("要复制的内容")
```

---

### motion - 动作

来源：https://autogo.cc/ios/motion.md

# motion - 操作
---
motion 模块提供模拟用户操作的函数，如点击、滑动、按键等。

以下为 `motion` 包中的按键常量（用于 `KeyAction`）：

| **常量** | **值** | **说明** |
|---------|--------|---------|
| `KEYCODE_HOME` | 3 | Home |
| `KEYCODE_DPAD_UP` | 19 | 方向键上 |
| `KEYCODE_DPAD_DOWN` | 20 | 方向键下 |
| `KEYCODE_DPAD_LEFT` | 21 | 方向键左 |
| `KEYCODE_DPAD_RIGHT` | 22 | 方向键右 |
| `KEYCODE_VOLUME_UP` | 24 | 音量加 |
| `KEYCODE_VOLUME_DOWN` | 25 | 音量减 |
| `KEYCODE_POWER` | 26 | 电源 |
| `KEYCODE_TAB` | 61 | Tab |
| `KEYCODE_SPACE` | 62 | 空格 |
| `KEYCODE_ENTER` | 66 | 回车 |
| `KEYCODE_DEL` | 67 | 退格 |
| `KEYCODE_SEARCH` | 84 | 搜索/回车（搜索框场景） |
| `KEYCODE_ESCAPE` | 111 | Escape |
| `KEYCODE_FORWARD_DEL` | 112 | 向前删除 |
| `KEYCODE_APP_SWITCH` | 187 | 应用切换 |

## TouchDown
<hr style="margin: 0;">

模拟触摸按下。

- `x` {int} 触摸点 X 坐标。
- `y` {int} 触摸点 Y 坐标。
- `fingerID` {int} 触摸点指针 ID（0~9）。

```go
motion.TouchDown(500, 600, 0)
```

## TouchMove
<hr style="margin: 0;">

模拟触摸移动。

- `x` {int} 目标 X 坐标。
- `y` {int} 目标 Y 坐标。
- `fingerID` {int} 触摸点指针 ID（0~9）。

```go
motion.TouchMove(550, 650, 0)
```

## TouchUp
<hr style="margin: 0;">

模拟触摸抬起。

- `x` {int} 抬起点 X 坐标。
- `y` {int} 抬起点 Y 坐标。
- `fingerID` {int} 触摸点指针 ID（0~9）。

```go
motion.TouchUp(550, 650, 0)
```

## Click
<hr style="margin: 0;">

模拟单击（内部为按下与抬起，带短随机间隔）。

- `x` {int} 点击 X 坐标。
- `y` {int} 点击 Y 坐标。
- `fingerID` {int} 触摸点指针 ID（0~9）。

```go
motion.Click(500, 600, 0)
```

## LongClick
<hr style="margin: 0;">

模拟长按。

- `x` {int} 长按点 X 坐标。
- `y` {int} 长按点 Y 坐标。
- `duration` {int} 长按持续时间（毫秒）。
- `fingerID` {int} 触摸点指针 ID（0~9）。

```go
motion.LongClick(500, 600, 1000, 0)
```

## Swipe
<hr style="margin: 0;">

模拟直线滑动。

- `x1` {int} 起点 X。
- `y1` {int} 起点 Y。
- `x2` {int} 终点 X。
- `y2` {int} 终点 Y。
- `duration` {int} 持续时间（毫秒）。
- `fingerID` {int} 触摸点指针 ID（0~9）。

```go
motion.Swipe(300, 800, 300, 200, 500, 0)
```

## Swipe2
<hr style="margin: 0;">

使用贝塞尔等方式滑动，轨迹更接近自然手势。

参数与 `Swipe` 相同。

```go
motion.Swipe2(300, 800, 300, 200, 500, 0)
```

## Home
<hr style="margin: 0;">

模拟按下 Home 键。

```go
motion.Home()
```

## Recents
<hr style="margin: 0;">

打开最近任务（实现为连续 Home 相关按键组合，行为与系统相关）。

```go
motion.Recents()
```

## VolumeUp
<hr style="margin: 0;">

按下音量加键。

```go
motion.VolumeUp()
```

## VolumeDown
<hr style="margin: 0;">

按下音量减键。

```go
motion.VolumeDown()
```

## KeyAction
<hr style="margin: 0;">

模拟按下指定按键。

- `code` {int} 按键码，使用上表 `KEYCODE_*` 常量。

```go
motion.KeyAction(motion.KEYCODE_ENTER)
```

---

### images - 图像处理

来源：https://autogo.cc/ios/images.md

# images - 图像处理
---
images模块提供了截图、图像处理、颜色查找等功能。

## SetCallback
<hr style="margin: 0;">

设置一个新图像数据到达的回调。

- `callback` {function} 当新图像数据到达时调用的函数，格式为 `func(img *image.NRGBA)`，如果传入 `nil`，则会移除当前设置的回调

```go
images.SetCallback(func(img *image.NRGBA) {
    // 处理新图像数据
})
```

**注意事项：**
- 回调函数应避免执行耗时操作，否则可能导致后续图像数据处理延迟
- 回调函数内部如需进行耗时操作（如文件写入或网络请求），建议启动新的 goroutine 处理，避免阻塞回调执行

## CaptureScreen
<hr style="margin: 0;">

截取屏幕的指定区域。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度

```go
img := images.CaptureScreen(0, 0, 0, 0) // 截取全屏
```

## Pixel
<hr style="margin: 0;">

获取指定坐标点的颜色值。

- `x` {int} 坐标点的 x 位置
- `y` {int} 坐标点的 y 位置

```go
color := images.Pixel(100, 200)
```

## CmpColor
<hr style="margin: 0;">

比较指定坐标点 (x, y) 的颜色。

- `x` {int} 坐标点的 x 位置
- `y` {int} 坐标点的 y 位置
- `colorStr` {string} 颜色字符串，格式如 "FFFFFF|CCCCCC-101010"，每种颜色用 "|" 分割，"-" 后表示偏色
- `sim` {float32} 相似度，取值范围 0.1-1.0

```go
matched := images.CmpColor(100, 200, "FFFFFF|CCCCCC-101010", 0.9)
```

## FindColor
<hr style="margin: 0;">

在指定区域内查找目标颜色。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colorStr` {string} 颜色格式串，例如 "FFFFFF|CCCCCC-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `dir` {int} 查找方向，0-3 分别表示不同的查找方向

**查找方向说明：**
- 0 - 从左到右，从上到下
- 1 - 从右到左，从上到下
- 2 - 从左到右，从下到上
- 3 - 从右到左，从下到上

```go
x, y := images.FindColor(0, 0, 0, 0, "FFFFFF", 0.9, 0)
```

## GetColorCountInRegion
<hr style="margin: 0;">

计算指定区域内符合颜色条件的像素数量。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colorStr` {string} 要查找的颜色字符串，例如 "FFFFFF|CCCCCC-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0

```go
count := images.GetColorCountInRegion(0, 0, 0, 0, "FFFFFF", 0.9)
```

## DetectsMultiColors
<hr style="margin: 0;">

根据指定的颜色串信息在屏幕进行多点颜色比对（多点比色）。

- `colors` {string} 颜色模板字符串，例如 "369,1220,ffab2d-101010,370,1221,24b1ff-101010,380,390,907efd-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0

```go
matched := images.DetectsMultiColors("369,1220,ffab2d-101010,370,1221,24b1ff-101010", 0.9)
```

## FindMultiColors
<hr style="margin: 0;">

在指定区域内查找匹配的多点颜色序列（多点找色）。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colors` {string} 颜色模板字符串，例如 "ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `dir` {int} 查找方向，0-3 分别表示不同的查找方向

```go
x, y := images.FindMultiColors(0, 0, 0, 0, "ffccff-151515,635,978,ffab2d-101010", 0.9, 0)
```

## FindMultiColorsAll
<hr style="margin: 0;">

在指定区域内查找匹配的多点颜色序列并返回所有符合条件的坐标。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colors` {string} 颜色模板字符串，例如 "ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010"
- `sim` {float32} 相似度，取值范围 0.1-1.0
- `dir` {int} 查找方向，0-3 分别表示不同的查找方向

```go
points := images.FindMultiColorsAll(0, 0, 0, 0, "ffccff-151515,635,978,ffab2d-101010", 0.9, 0)
```

## ReadFromPath
<hr style="margin: 0;">

读取路径指定的图片文件并返回图像对象。

- `path` {string} 要读取的图片文件路径

```go
img := images.ReadFromPath("/sdcard/image.png")
```

## ReadFromBase64
<hr style="margin: 0;">

解码 Base64 数据并返回解码后的图片对象。

- `base64Str` {string} 要解码的 Base64 字符串

```go
img := images.ReadFromBase64("iVBORw0KGgoAAAANSUhEUgAAAAUA...")
```

## ReadFromBytes
<hr style="margin: 0;">

解码字节数组并返回解码后的图片对象。

- `data` {[]byte} 要解码的字节数组

```go
img := images.ReadFromBytes(bytes)
```

## Save
<hr style="margin: 0;">

把图片保存到指定路径。

- `img` {*image.NRGBA} 要保存的图像对象
- `path` {string} 保存图片的文件路径
- `quality` {int} 保存图片的质量（如果适用）

```go
success := images.Save(img, "/sdcard/saved.png", 100)
```

## EncodeToBase64
<hr style="margin: 0;">

把图像对象编码为 Base64 数据。

- `img` {*image.NRGBA} 要编码的图像对象
- `format` {string} 编码的图片格式（如 "png", "jpg" 等）
- `quality` {int} 编码的图片质量

```go
base64Str := images.EncodeToBase64(img, "png", 100)
```

## EncodeToBytes
<hr style="margin: 0;">

把图片编码为字节数组。

- `img` {*image.NRGBA} 要编码的图像对象
- `format` {string} 编码的图片格式（如 "png", "jpg" 等）
- `quality` {int} 编码的图片质量

```go
bytes := images.EncodeToBytes(img, "png", 100)
```

## ToNrgba
<hr style="margin: 0;">

将任意 image.Image 对象转换为 *image.NRGBA。

- `img` {image.Image} 要转换的图像对象

```go
nrgbaImg := images.ToNrgba(anyImg)
```

## Clip
<hr style="margin: 0;">

从源图像中裁剪指定区域并返回新图像。

- `img` {*image.NRGBA} 要裁剪的图像
- `x1` {int} 裁剪区域左上角 x 坐标
- `y1` {int} 裁剪区域左上角 y 坐标
- `x2` {int} 裁剪区域右下角 x 坐标
- `y2` {int} 裁剪区域右下角 y 坐标

```go
clippedImg := images.Clip(img, 100, 100, 300, 300)
```

## Resize
<hr style="margin: 0;">

调整图像大小。

- `img` {*image.NRGBA} 要调整的图像
- `width` {int} 目标宽度
- `height` {int} 目标高度

```go
resizedImg := images.Resize(img, 800, 600)
```

## Rotate
<hr style="margin: 0;">

旋转图像。

- `img` {*image.NRGBA} 要旋转的图像
- `degree` {int} 旋转角度（顺时针方向）

```go
rotatedImg := images.Rotate(img, 90)
```

## Grayscale
<hr style="margin: 0;">

将彩色图像转换为灰度图像。

- `img` {*image.NRGBA} 要转换的彩色图像

```go
grayImg := images.Grayscale(img)
```

## ApplyThreshold
<hr style="margin: 0;">

对图像应用阈值处理。

- `img` {*image.NRGBA} 要处理的图像
- `threshold` {int} 阈值
- `maxVal` {int} 超过阈值后应用的值
- `typ` {string} 阈值类型，如 "BINARY", "BINARY_INV" 等

```go
thresholdImg := images.ApplyThreshold(img, 128, 255, "BINARY")
```

## ApplyAdaptiveThreshold
<hr style="margin: 0;">

应用自适应阈值处理。

- `img` {*image.NRGBA} 要处理的图像
- `maxValue` {float64} 最大值
- `adaptiveMethod` {string} 自适应方法，如 "MEAN_C", "GAUSSIAN_C"
- `thresholdType` {string} 阈值类型，如 "BINARY", "BINARY_INV"
- `blockSize` {int} 用于计算阈值的像素邻域大小
- `C` {float64} 从平均值或加权平均值中减去的常量

```go
adaptiveImg := images.ApplyAdaptiveThreshold(img, 255, "MEAN_C", "BINARY", 11, 2)
```

## ApplyBinarization
<hr style="margin: 0;">

应用二值化处理。

- `img` {*image.NRGBA} 要处理的图像
- `threshold` {int} 阈值

```go
binaryImg := images.ApplyBinarization(img, 128)
```

---

### opencv - 图像处理

来源：https://autogo.cc/ios/opencv.md

# opencv - 图像处理
---
提供基于 OpenCV 的图像处理功能。由于 OpenCV 方法数量太多，剩余的方法全部参照 [官方文档](https://docs.opencv.org/4.10.0/)

## FindImage
<hr style="margin: 0;">

在指定区域内查找匹配的图片模板。返回找到的图片左上角坐标，如果未找到则返回 (-1, -1)。

- `x1`, `y1` {int}: 区域左上角的坐标。
- `x2`, `y2` {int}: 区域右下角的坐标。当 `x2` 或 `y2` 为 0 时，表示使用图像的最大宽度或高度。
- `template` {*[]byte}: 模板图片的字节数组指针，表示要在区域内查找的图片。
- `isGray` {bool}: 是否将图像转换为灰度图进行匹配。
- `isTransparent` {bool}: 是否按透明图处理。为 `true` 时，模板左上角第一个像素的 RGB 颜色会被当作透明色，匹配时忽略模板中所有同色像素。
- `sim` {float32}: 相似度阈值，取值范围为 0.1 到 1.0，值越高表示匹配要求越精确。

说明：

- `isTransparent` 为 `false` 时，按普通图片匹配，不生成遮罩。
- `isTransparent` 为 `true` 时，按透明图片匹配，透明色由模板左上角像素决定。

```go
x, y := opencv.FindImage(0, 0, 1920, 1080, &templateBytes, false, false, 0.8)
if x != -1 && y != -1 {
    fmt.Printf("模板匹配成功，坐标为: (%d, %d)\n", x, y)
} else {
    fmt.Println("未找到匹配的模板。")
}
```

## FindImageFromImage
<hr style="margin: 0;">

在给定图像中查找匹配的图片模板。参数含义与 `FindImage` 相同，但 `img` 直接作为待匹配图像，不会进行屏幕截图。

- `img` {*image.NRGBA}: 待匹配图像。
- `template` {*[]byte}: 模板图片的字节数组指针。
- `isGray` {bool}: 是否将图像转换为灰度图进行匹配。
- `isTransparent` {bool}: 是否按透明图处理。为 `true` 时，模板左上角第一个像素的 RGB 颜色会被当作透明色。
- `sim` {float32}: 相似度阈值，取值范围为 0.1 到 1.0。

```go
x, y := opencv.FindImageFromImage(img, &templateBytes, false, false, 0.8)
```

## FindImageAll
<hr style="margin: 0;">

在指定区域内查找匹配的图片模板，返回所有符合条件的坐标。

- `x1`, `y1` {int}: 区域左上角的坐标。
- `x2`, `y2` {int}: 区域右下角的坐标。当 `x2` 或 `y2` 为 0 时，表示使用图像的最大宽度或高度。
- `template` {*[]byte}: 模板图片的字节数组指针，表示要在区域内查找的图片。
- `isGray` {bool}: 是否将图像转换为灰度图进行匹配。
- `isTransparent` {bool}: 是否按透明图处理。为 `true` 时，模板左上角第一个像素的 RGB 颜色会被当作透明色。
- `sim` {float32}: 相似度阈值，取值范围为 0.1 到 1.0，值越高表示匹配要求越精确。

```go
points := opencv.FindImageAll(0, 0, 1920, 1080, &templateBytes, false, false, 0.8)
```

---

### dotocr - 点阵OCR

来源：https://autogo.cc/ios/dotocr.md

# dotocr - 点阵OCR
---
提供基于模板匹配的文字识别和查找功能。

## SetDict
<hr style="margin: 0;">

设置字库。字库内容按行分割，每行一条模板记录。

- `name` {string} 字库名称，为空字符串时使用 "default"
- `dict` {string} 字库内容字符串，按行分割，每行一条模板记录

```go
dotocr.SetDict("字库1", dictContent)
```

## Ocr
<hr style="margin: 0;">

在屏幕指定区域进行OCR文字识别。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `threshold` {string} 阈值字符串，例如 "ffffff-101010"
- `sim` {float32} 匹配相似度阈值，取值范围 0.0-1.0，例如 0.8
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称，为空字符串时使用 "default"
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

**返回值：**
- 当 `asJSON == false` 时，返回纯文本字符串，按检测顺序拼接所有识别到的字符
- 当 `asJSON == true` 时，返回 JSON 数组字符串，每个元素包含 `{"x":坐标x, "y":坐标y, "width":宽度, "height":高度, "text":文字内容, "sim":相似度}`

```go
// 返回纯文本
text := dotocr.Ocr(0, 0, 100, 50, "ffffff-101010", 0.8, false, "字库1", 0)

// 返回 JSON 格式
jsonStr := dotocr.Ocr(0, 0, 100, 50, "ffffff-101010", 0.8, true, "字库1", 0)
```

## OcrFromImage
<hr style="margin: 0;">

从图像对象进行OCR文字识别。

- `img` {*image.NRGBA} 要识别的图像对象
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称

**返回值：** 识别结果字符串（纯文本或 JSON 格式）

```go
img := images.ReadFromPath("/sdcard/screenshot.png")
text := dotocr.OcrFromImage(img, "ffffff-101010", 0.8, false, "字库1")
```

## OcrFromBase64
<hr style="margin: 0;">

从Base64编码的图像数据进行OCR文字识别。

- `b64` {string} Base64编码的图像数据
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称

**返回值：** 识别结果字符串（纯文本或 JSON 格式）

```go
text := dotocr.OcrFromBase64(base64String, "ffffff-101010", 0.8, false, "字库1")
```

## OcrFromPath
<hr style="margin: 0;">

从图像文件路径进行OCR文字识别。

- `path` {string} 图像文件路径
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `asJSON` {bool} 是否以 JSON 格式返回
- `dictName` {string} 使用的字库名称

**返回值：** 识别结果字符串（纯文本或 JSON 格式）

```go
text := dotocr.OcrFromPath("/sdcard/screenshot.png", "ffffff-101010", 0.8, false, "字库1")
```

## FindStr
<hr style="margin: 0;">

在屏幕指定区域中查找指定字符串的位置。

- `x1` {int} 查找区域的左上角 x 坐标
- `y1` {int} 查找区域的左上角 y 坐标
- `x2` {int} 查找区域的右下角 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 查找区域的右下角 y 坐标，当为 0 时表示使用屏幕最大高度
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称
- `displayId` {int} 屏幕ID，0表示主屏幕，其他值表示虚拟屏幕

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
x, y := dotocr.FindStr(0, 0, 0, 0, "商店", "ffffff-101010", 0.8, "字库1", 0)
if x != -1 && y != -1 {
    fmt.Printf("找到字符串，坐标: (%d, %d)\n", x, y)
}
```

## FindStrFromImage
<hr style="margin: 0;">

在图像对象中查找指定字符串的位置。

- `img` {*image.NRGBA} 要查找的图像对象
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
img := images.ReadFromPath("/sdcard/screenshot.png")
x, y := dotocr.FindStrFromImage(img, "商店", "ffffff-101010", 0.8, "字库1")
```

## FindStrFromBase64
<hr style="margin: 0;">

在Base64编码的图像中查找指定字符串的位置。

- `b64` {string} Base64编码的图像数据
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
x, y := dotocr.FindStrFromBase64(base64String, "商店", "ffffff-101010", 0.8, "字库1")
```

## FindStrFromPath
<hr style="margin: 0;">

在图像文件中查找指定字符串的位置。

- `path` {string} 图像文件路径
- `text` {string} 要查找的字符串
- `threshold` {string} 阈值字符串
- `sim` {float32} 匹配相似度阈值
- `dictName` {string} 使用的字库名称

**返回值：** 
- `x` {int} 找到时返回字符串第一个字符的 x 坐标，未找到返回 -1
- `y` {int} 找到时返回字符串第一个字符的 y 坐标，未找到返回 -1

```go
x, y := dotocr.FindStrFromPath("/sdcard/screenshot.png", "商店", "ffffff-101010", 0.8, "字库1")
```

---

### storages - 本地存储

来源：https://autogo.cc/ios/storages.md

# storages - 本地存储
---
提供了保存简单数据、用户配置等的支持。保存的数据除非应用被卸载或者被主动删除，否则会一直保留。

## Get
<hr style="margin: 0;">

从本地存储中取出键值为 `key` 的数据。

- `table` {string} 要操作的表。
- `key` {string} 要查询的键。

```go
value := storages.Get("data", "username")
fmt.Println("获取到的值:", value)
```

## Put
<hr style="margin: 0;">

把值 `value` 保存到本地存储中。

- `table` {string} 要操作的表。
- `key` {string} 要保存的键。
- `value` {string} 要保存的值。

```go
storages.Put("data", "username", "JohnDoe")
```

## Remove
<hr style="margin: 0;">

移除键值为 `key` 的数据。

- `table` {string} 要操作的表。
- `key` {string} 要移除的键。

```go
storages.Remove("data", "username")
```

## Contains
<hr style="margin: 0;">

返回该本地存储是否包含键值为 `key` 的数据。

- `table` {string} 要操作的表。
- `key` {string} 要检查的键。

```go
exists := storages.Contains("data", "username")
if exists {
    fmt.Println("键存在!")
} else {
    fmt.Println("键不存在!")
}
```

## GetAll
<hr style="margin: 0;">

获取该本地存储所有键值对。

- `table` {string} 要操作的表。

```go
data := storages.GetAll("data")
for k, v := range data {
	fmt.Printf("%s = %s\n", k, v)
}
```

## Clear
<hr style="margin: 0;">

移除该本地存储的所有数据。

- `table` {string} 要操作的表。

```go
storages.Clear("data")
fmt.Println("所有存储数据已被清除!")
```

---

### system - 系统

来源：https://autogo.cc/ios/system.md

# system - 系统功能
---
提供与进程相关的查询与脚本进程重启。

## GetPid
<hr style="margin: 0;">

获取进程 PID。

- `processName` {string} 进程名；若为空字符串，返回**当前进程**的 PID。

```go
pid := system.GetPid("SpringBoard")
selfPid := system.GetPid("")
```

## GetMemoryUsage
<hr style="margin: 0;">

获取指定进程的内存使用量（KB）。

- `pid` {int} 进程 PID。

```go
kb := system.GetMemoryUsage(pid)
fmt.Println("Memory (KB):", kb)
```

## GetCpuUsage
<hr style="margin: 0;">

获取指定进程的 CPU 使用率。

- `pid` {int} 进程 PID。

```go
cpu := system.GetCpuUsage(pid)
fmt.Println("CPU (%):", cpu)
```

## RestartSelf
<hr style="margin: 0;">

重启当前脚本进程。

```go
system.RestartSelf()
```

---

### utils - 工具函数

来源：https://autogo.cc/ios/utils.md

# utils - 工具函数
---
提供常用的工具函数，包括 Toast、系统对话框、随机数、休眠以及字符串与数值类型转换等。

## Toast
<hr style="margin: 0;">

显示 Toast 提示信息。

- `message` {string} 要显示的提示信息。
- `x` {int} 在界面上显示的 X 坐标（传递-1使用默认坐标）。
- `y` {int} 在界面上显示的 Y 坐标（传递-1使用默认坐标）。
- `duration` {int} 提示显示的持续时间，单位为毫秒（传递-1使用默认2000毫秒）。

```go
utils.Toast("Hello AutoGo", -1, -1, -1)
```

## Alert
<hr style="margin: 0;">

弹出系统级对话框，阻塞直到用户点击按钮。

- `title` {string} 弹窗标题。
- `content` {string} 弹窗内容。
- `btn1Text` {string} 第一个按钮文字；`btn2Text` 为空时通常只显示该按钮。
- `btn2Text` {string} 第二个按钮文字；传空字符串则只显示一个按钮。

**返回** {int} `0` 表示点击了第一个按钮，`1` 表示点击了第二个按钮。

```go
btnIndex := utils.Alert("确认操作", "是否继续？", "取消", "确定")
if btnIndex == 1 {
	fmt.Println("用户点击了确定")
}
```

## InputAlert
<hr style="margin: 0;">

弹出带输入框的系统级对话框，阻塞等待用户操作。

- `title` {string} 弹窗标题。
- `content` {string} 说明文字。
- `placeholder` {string} 输入框占位提示。
- `defaultText` {string} 输入框默认文本。
- `btn1Text` {string} 取消类按钮文字。
- `btn2Text` {string} 确认类按钮文字；为空则只显示一个按钮。

**返回** `(string, bool)`：点击确认返回 `(输入内容, true)`，点击取消返回 `("", false)`。

```go
text, ok := utils.InputAlert("输入", "请输入备注", "备注", "", "取消", "确定")
if ok {
	fmt.Println("输入:", text)
}
```

## ExecBinary
<hr style="margin: 0;">

执行指定路径的二进制可执行文件。

- `path` {string} 二进制文件的路径。
- `args` {...string} 传递给二进制文件的命令行参数。

**返回** {int} 返回进程的PID,失败的话返回-1。

```go
pid := utils.ExecBinary(files.Path("./app"))
```


## Sleep
<hr style="margin: 0;">

让当前协程暂停指定时间。

- `i` {int} 暂停时间（毫秒）。

```go
utils.Sleep(500)
```

## Random
<hr style="margin: 0;">

返回指定范围内的随机整数（含最小值与最大值）。

- `min` {int} 最小值。
- `max` {int} 最大值。

```go
n := utils.Random(1, 10)
```

## I2s
<hr style="margin: 0;">

将整数转换为字符串。

- `i` {int} 要转换的整数。

```go
str := utils.I2s(123)
```

## S2i
<hr style="margin: 0;">

将字符串转换为整数（解析失败时为 0）。

- `s` {string} 要转换的字符串。

```go
num := utils.S2i("123")
```

## F2s
<hr style="margin: 0;">

将浮点数转换为字符串。

- `f` {float64} 要转换的浮点数。

```go
str := utils.F2s(123.45)
```

## S2f
<hr style="margin: 0;">

将字符串转换为浮点数。转换失败返回 `0.0`。

- `s` {string} 要转换的字符串。

```go
num := utils.S2f("123.45")
```

## B2s
<hr style="margin: 0;">

将布尔值转换为字符串（`"true"` 或 `"false"`）。

- `b` {bool} 要转换的布尔值。

```go
str := utils.B2s(true)
```

## S2b
<hr style="margin: 0;">

将字符串转换为布尔值。无法解析时返回 `false`。

- `s` {string} 要转换的字符串。

```go
b := utils.S2b("true")
```

---

### ppocr - 飞浆OCR

来源：https://autogo.cc/ios/ppocr.md

# ppocr - 飞桨OCR
---
提供基于 PaddleOCR 的文字检测和识别功能。

以下是 `ppocr` 包中定义的 `Result` 结构体及其字段说明：

| **字段名**  | **类型**    | **说明**                              |
|-------------|-------------|---------------------------------------|
| `X`         | `int`       | 检测结果的左上角 X 坐标。             |
| `Y`         | `int`       | 检测结果的左上角 Y 坐标。             |
| `Width`     | `int`       | 检测结果的宽度。                      |
| `Height`    | `int`       | 检测结果的高度。                      |
| `Label`     | `string`    | 检测到的文字内容或标签。              |
| `Score`     | `float64`   | 检测结果的置信度，取值范围为 0-1。   |
| `CenterX`   | `int`       | 检测结果的中心 X 坐标。               |
| `CenterY`   | `int`       | 检测结果的中心 Y 坐标。               |

## New
<hr style="margin: 0;">

创建一个 Ppocr 实例对象。成功返回实例对象`*Ppocr`，如果加载失败则返回 nil。

- `version` {string} 模型版本，目前仅支持`v2`和`v5`

```go
ocr := ppocr.New("v5")
if ocr == nil {
    fmt.Println("初始化失败")
    return
}
fmt.Println("初始化成功")
```

## *Ppocr.Ocr
<hr style="margin: 0;">

在屏幕指定区域进行OCR文字识别。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"，"-" 后表示偏色范围，如果不需要指定则直接传入空字符串`""`

```go
results := ocr.Ocr(0, 0, 0, 0, "000000-101010") // 识别主屏幕全屏的黑色文字
for _, result := range results {
    fmt.Println(result.Label) // 打印识别到的文字
}
```

## *Ppocr.OcrFromImage
<hr style="margin: 0;">

从图像对象进行OCR文字识别。

- `img` {*image.NRGBA} 要识别的图像对象
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"

```go
img := images.ReadFromPath("/sdcard/screenshot.png")
results := ocr.OcrFromImage(img, "000000")
```

## *Ppocr.OcrFromBase64
<hr style="margin: 0;">

从Base64编码的图像数据进行OCR文字识别。

- `b64` {string} Base64编码的图像数据
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"

```go
results := ocr.OcrFromBase64(base64String, "000000")
```

## *Ppocr.OcrFromPath
<hr style="margin: 0;">

从图像文件路径进行OCR文字识别。

- `path` {string} 图像文件路径
- `colorStr` {string} 文字颜色范围，格式如 "FFFFFF" 或 "FFFFFF-101010"

```go
results := ocr.OcrFromPath("/sdcard/screenshot.png", "000000")
```

## *Ppocr.Close
<hr style="margin: 0;">

关闭 PPOCR 实例。

```go
ocr.Close()
```

---

### yolo - 目标检测

来源：https://autogo.cc/ios/yolo.md

# yolo - 目标检测
---
提供基于 YOLO 的目标检测功能。

以下是 `yolo` 包中定义的 `Result` 结构体及其字段说明：

| **字段名**  | **类型**    | **说明**                              |
|-------------|-------------|---------------------------------------|
| `X`         | `int`       | 检测结果的左上角 X 坐标。             |
| `Y`         | `int`       | 检测结果的左上角 Y 坐标。             |
| `Width`     | `int`       | 检测结果的宽度。                      |
| `Height`    | `int`       | 检测结果的高度。                      |
| `Label`     | `string`    | 检测到的文字内容或标签。              |
| `Score`     | `float64`   | 检测结果的置信度，取值范围为 0-1。   |
| `CenterX`   | `int`       | 检测结果的中心 X 坐标。               |
| `CenterY`   | `int`       | 检测结果的中心 Y 坐标。               |

## New
<hr style="margin: 0;">

创建一个 Yolo 实例对象。成功返回实例对象`*Yolo`，如果加载失败则返回 nil。

- `version` {string} 模型版本，目前仅支持`v5`和`v8`
- `cpuThreadNum` {int} 用于模型推理的 CPU 线程数。
- `paramPath` {string} 模型参数文件路径。
- `binPath` {string} 模型二进制文件路径。
- `labels` {string} 标签文本，多个标签使用`,`进行隔开。

```go
yolo := yolo.New("v8", 4, "/data/local/tmp/param", "/data/local/tmp/bin", "person,bicycle,car")
if yolo == nil {
    fmt.Println("模型加载失败")
    return
}
fmt.Println("模型加载成功")
```

## *Yolo.Detect
<hr style="margin: 0;">

在屏幕指定区域进行目标检测。

- `x1` {int} 区域左上角的 x 坐标
- `y1` {int} 区域左上角的 y 坐标
- `x2` {int} 区域右下角的 x 坐标，当为 0 时表示使用屏幕最大宽度
- `y2` {int} 区域右下角的 y 坐标，当为 0 时表示使用屏幕最大高度

```go
results := detector.Detect(0, 0, 0, 0) // 在主屏幕全屏范围内检测目标
for _, result := range results {
    fmt.Printf("检测到 %s，置信度: %.2f\n", result.Label, result.Score)
}
```

## *Yolo.DetectFromImage
<hr style="margin: 0;">

从图像对象进行目标检测。

- `img` {*image.NRGBA} 要检测的图像对象

```go
img := images.ReadFromPath("/sdcard/photo.jpg")
results := detector.DetectFromImage(img)
```

## *Yolo.DetectFromBase64
<hr style="margin: 0;">

从Base64编码的图像数据进行目标检测。

- `b64` {string} Base64编码的图像数据
- `colorStr` {string} 保留参数，可以传入空字符串

```go
results := detector.DetectFromBase64(base64String, "")
```

## *Yolo.DetectFromPath
<hr style="margin: 0;">

从图像文件路径进行目标检测。

- `path` {string} 图像文件路径
- `colorStr` {string} 保留参数，可以传入空字符串

```go
results := detector.DetectFromPath("/sdcard/photo.png", "")
```

## *Yolo.Close
<hr style="margin: 0;">

关闭 YOLO 模型实例，释放相关资源。

```go
yolo.Close()
```

---

### imgui - 界面绘制

来源：https://autogo.cc/ios/imgui.md

# imgui - 即时模式图形用户界面
---
提供基于 Dear ImGui 的图形用户界面功能。由于 ImGui 方法数量众多，完整的方法列表请参照 [Dear ImGui 官方文档](https://github.com/ocornut/imgui)

## 基础示例

```go
package main

import (
    "fmt"
    "github.com/Dasongzi1366/AutoGo/imgui"
)

func main() {
    // 初始化
    imgui.Init()
    
    // 状态变量
    counter := 0
    showWindow := true
    
    // 主循环
    imgui.Run(func() {
        // 设置窗口
        imgui.SetNextWindowSizeV(imgui.Vec2{X: 500, Y: 400}, imgui.CondOnce)
        imgui.SetNextWindowPosV(imgui.Vec2{X: 100, Y: 100}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})
        
        // 创建窗口
        imgui.BeginV("示例窗口", &showWindow, 0)
        
        // 标题
        imgui.Text("ImGui 示例程序")
        imgui.Separator()
        imgui.Spacing()
        
        // 计数器
        imgui.Text(fmt.Sprintf("计数器: %d", counter))
        
        // 按钮
        if imgui.Button("增加") {
            counter++
        }
        imgui.SameLine()
        if imgui.Button("减少") {
            counter--
        }
        imgui.SameLine()
        if imgui.Button("重置") {
            counter = 0
        }
        
        imgui.Spacing()
        imgui.Separator()
        imgui.Spacing()
        
        // 样式化按钮
        imgui.Text("样式化按钮：")
        imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0})
        if imgui.Button("绿色按钮") {
            // 绿色按钮的操作
        }
        imgui.PopStyleColor()
        
        imgui.SameLine()
        imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{X: 0.8, Y: 0.2, Z: 0.2, W: 1.0})
        if imgui.Button("红色按钮") {
            // 红色按钮的操作
        }
        imgui.PopStyleColor()
        
        // 结束窗口
        imgui.End()
    })

    // 阻塞主进程防止程序退出
	select {}
}
```

---

### console - 控制台

来源：https://autogo.cc/ios/console.md

# console - 控制台
---
提供用于控制台悬浮窗的控制接口，支持多实例、位置、大小、颜色设置以及内容打印等功能。

## New
<hr style="margin: 0;">

创建一个新的控制台实例。

**返回** {*Console} 控制台实例指针

```go
c := console.New()
```

## SetWindowSize
<hr style="margin: 0;">

设置控制台窗口的宽高。

- `width` {int} 控制台窗口的宽度
- `height` {int} 控制台窗口的高度

**返回** {*Console} 控制台实例指针

```go
c.SetWindowSize(800, 600)
```

## SetWindowPosition
<hr style="margin: 0;">

设置控制台窗口的位置。

- `x` {int} 控制台窗口左上角的横坐标
- `y` {int} 控制台窗口左上角的纵坐标

**返回** {*Console} 控制台实例指针

```go
c.SetWindowPosition(100, 200)
```

## SetWindowColor
<hr style="margin: 0;">

设置控制台窗口的背景颜色。

- `color` {string} 背景颜色的十六进制字符串，格式如 "#1E1F22"

**返回** {*Console} 控制台实例指针

```go
c.SetWindowColor("#1E1F22")
```

## SetTextColor
<hr style="margin: 0;">

设置控制台文字颜色。

- `color` {string} 文字颜色的十六进制字符串，格式如 "#FFFFFF"

**返回** {*Console} 控制台实例指针

```go
c.SetTextColor("#FFFFFF")
```

## SetTextSize
<hr style="margin: 0;">

设置控制台文字大小。

- `size` {int} 文字大小

```go
c.SetTextSize(50)
```

## Println
<hr style="margin: 0;">

打印文本到控制台。

- `a` {any} 要打印的参数，支持多个参数，行为类似 fmt.Println

```go
c.Println("Hello, world!")
c.Println("用户ID:", 123, "状态:", "在线")
```

## Clear
<hr style="margin: 0;">

清空控制台内容。

```go
c.Clear()
```

## Show
<hr style="margin: 0;">

显示控制台窗口。

```go
c.Show()
```

## Hide
<hr style="margin: 0;">

隐藏控制台窗口。

```go
c.Hide()
```

## IsVisible
<hr style="margin: 0;">

检查控制台是否可见。

**返回** {bool} true 表示可见，false 表示隐藏

```go
if c.IsVisible() {
    // 控制台当前可见
}
```

## Destroy
<hr style="margin: 0;">

销毁控制台实例，释放资源。

```go
c.Destroy()
```

## 示例
<hr style="margin: 0;">

```go
// 创建并配置控制台
c := console.New()
c.SetWindowPosition(50, 50)
c.SetWindowSize(700, 500)
c.SetWindowColor("#1E1F22")
c.SetTextColor("#00FF00")
c.SetTextSize(45)

// 打印日志
c.Println("控制台已就绪")

for {
    c.Println("当前时间:", time.Now().Format("2006-01-02 15:04:05"))
    utils.Sleep(1000)
}
```

---

### hud - 悬浮显示

来源：https://autogo.cc/ios/hud.md

# hud - 悬浮显示
---
提供悬浮显示功能，支持多实例、彩色文本显示等功能。

以下是 `hud` 包中定义的 `TextItem` 结构体及其字段说明：

| **字段名**   | **类型**      | **说明**                                      |
|--------------|---------------|-----------------------------------------------|
| `TextColor`  | `color.Color` | 文字颜色。格式如 `"#FFFFFF"`。 |
| `Text`       | `string`      | 显示的文本内容。                              |

## New
<hr style="margin: 0;">

创建一个新的 HUD 实例。

**返回** {*HUD} HUD 实例指针

```go
h := hud.New()
```

## SetPosition
<hr style="margin: 0;">

设置 HUD 的位置和大小。

- `x1` {int} 左上角横坐标
- `y1` {int} 左上角纵坐标
- `x2` {int} 右下角横坐标
- `y2` {int} 右下角纵坐标

**返回** {*HUD} HUD 实例指针

```go
h.SetPosition(100, 100, 400, 150)
```

## SetBackgroundColor
<hr style="margin: 0;">

设置 HUD 的背景颜色。

- `color` {string} 背景颜色的十六进制字符串，格式如 "#2D2D30" 或 "#2D2D3080"（带透明度）

**返回** {*HUD} HUD 实例指针

```go
h.SetBackgroundColor("#2D2D30")
h.SetBackgroundColor("#00000080")  // 半透明黑色
```

## SetTextSize
<hr style="margin: 0;">

设置 HUD 的字体大小。

- `size` {int} 字体大小（推荐范围：30-60）

**返回** {*HUD} HUD 实例指针

```go
h.SetTextSize(45)
```

## SetText
<hr style="margin: 0;">

设置 HUD 显示的文本内容（支持多色文本）。

- `items` {[]TextItem} 文本项数组，每个元素包含颜色和文本

**返回** {*HUD} HUD 实例指针

```go
h.SetText([]hud.TextItem{
    {TextColor: "#00FF00", Text: "HP: "},
    {TextColor: "#FFFFFF", Text: "100/100"},
})
```

## Show
<hr style="margin: 0;">

显示 HUD。

**返回** {*HUD} HUD 实例指针

```go
h.Show()
```

## Hide
<hr style="margin: 0;">

隐藏 HUD。

**返回** {*HUD} HUD 实例指针

```go
h.Hide()
```

## IsVisible
<hr style="margin: 0;">

检查 HUD 是否可见。

**返回** {bool} true 表示可见，false 表示隐藏

```go
if h.IsVisible() {
    // HUD 当前可见
}
```

## Destroy
<hr style="margin: 0;">

销毁 HUD 实例，释放资源。

```go
h.Destroy()
```

## 示例
<hr style="margin: 0;">

```go
// 创建并配置 HUD
h := hud.New()
h.SetPosition(50, 700, 450, 750)
h.SetBackgroundColor("#3D000080")
h.SetTextSize(45)

for {
    // 设置多色文本
    h.SetText([]hud.TextItem{
        {TextColor: "#00FF00", Text: "当前时间: "},
        {TextColor: "#FFFFFF", Text: time.Now().Format("2006-01-02 15:04:05")},
    })
    utils.Sleep(1000)
}
```

---

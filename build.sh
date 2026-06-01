#!/bin/bash

# 配置参数
APP_NAME="AutoGo图色助手"
ICON_PATH="./build/logo.png"
ICON_ICNS="./build/logo.icns"
BUILD_DIR="./build"
VERSION="1.0.5"
BUILD_NUMBER="5"

echo "开始打包 macOS 应用..."
echo ""

# 编译参数：去除调试符号和调试信息，减小体积
LDFLAGS="-s -w"

# 1. 编译 ARM64 架构（用于 DMG）
echo "步骤 1/4: 编译 ARM64 (Apple Silicon) 架构..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="$LDFLAGS" -o "${APP_NAME}_arm64" .
if [ $? -ne 0 ]; then
    echo "✗ ARM64 架构编译失败"
    exit 1
fi
echo "✓ ARM64 架构编译完成"

# 2. 编译 AMD64 架构（独立二进制文件）
echo "步骤 2/4: 编译 AMD64 (Intel) 架构..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/${APP_NAME}_Intel" .
if [ $? -ne 0 ]; then
    echo "✗ AMD64 架构编译失败"
    rm -f "${APP_NAME}_arm64"
    exit 1
fi
echo "✓ AMD64 架构编译完成 -> $BUILD_DIR/${APP_NAME}_Intel"

# 3. 打包 ARM64 版本的 .app 应用
echo "步骤 3/4: 使用 fyne 打包 .app 文件 (ARM64)..."

# 先用 fyne 创建 .app 结构
fyne package -os darwin -name "$APP_NAME" -icon "$ICON_PATH" -appVersion "$VERSION" -appBuild "$BUILD_NUMBER"

if [ ! -d "${APP_NAME}.app" ]; then
    echo "✗ 打包 .app 失败"
    rm -f "${APP_NAME}_arm64"
    exit 1
fi

# 替换可执行文件为 ARM64 二进制
EXEC_PATH="${APP_NAME}.app/Contents/MacOS/app"
if [ -f "$EXEC_PATH" ]; then
    cp "${APP_NAME}_arm64" "$EXEC_PATH"
    chmod +x "$EXEC_PATH"
    echo "  ✓ 已使用 ARM64 二进制"
else
    echo "  ✗ 找不到可执行文件路径"
    rm -f "${APP_NAME}_arm64"
    exit 1
fi

# 验证最终的可执行文件
echo "  验证 .app 中的二进制架构："
lipo -info "$EXEC_PATH" | sed 's/^/    /'

# 清理临时文件
rm -f "${APP_NAME}_arm64"

echo "✓ .app 打包完成 (ARM64)"

# 4. 移动到 build 目录并创建 DMG
echo "步骤 4/4: 移动到 build 目录并创建 DMG..."

# 删除旧的 .app 和 DMG（如果存在）
if [ -d "$BUILD_DIR/${APP_NAME}.app" ]; then
    rm -rf "$BUILD_DIR/${APP_NAME}.app"
fi

mv -f "${APP_NAME}.app" "$BUILD_DIR/${APP_NAME}.app"
echo "  ✓ 已移动到 $BUILD_DIR"

# 创建 DMG 安装包
DMG_PATH="$BUILD_DIR/${APP_NAME}.dmg"

# 删除旧的 DMG（如果存在）
if [ -f "$DMG_PATH" ]; then
    rm -f "$DMG_PATH"
fi

# 使用 create-dmg 创建漂亮的 DMG
create-dmg \
  --volname "$APP_NAME" \
  --volicon "$ICON_ICNS" \
  --window-pos 200 120 \
  --window-size 700 400 \
  --icon-size 100 \
  --icon "${APP_NAME}.app" 200 160 \
  --hide-extension "${APP_NAME}.app" \
  --app-drop-link 500 160 \
  "$DMG_PATH" \
  "$BUILD_DIR/${APP_NAME}.app" \
  > /dev/null 2>&1

# 清理临时文件
if [ -f "${APP_NAME}" ]; then
    rm -f "${APP_NAME}"
fi

# 验证结果
echo ""
echo "================================"
if [ -f "$DMG_PATH" ]; then
    echo "✓ macOS 打包完成！"
    echo ""
    echo "生成的文件："
    echo "  ARM64 (Apple Silicon):"
    echo "    .app 文件: $BUILD_DIR/${APP_NAME}.app"
    echo "    DMG 文件: $DMG_PATH ($(du -h "$DMG_PATH" | cut -f1))"
    echo ""
    echo "  Intel Mac:"
    echo "    二进制文件: $BUILD_DIR/${APP_NAME}_Intel"
    echo ""
    echo "安装说明："
    echo "  Apple Silicon: 双击 DMG，拖入 Applications"
    echo "  Intel Mac: 直接运行 ${APP_NAME}_Intel 二进制文件"
    echo ""
    echo "注意：首次运行时，右键点击应用选择'打开'，然后在弹出窗口中再次点击'打开'"
else
    echo "✗ DMG 生成失败"
    exit 1
fi
echo "================================"

# 6. 创建 Intel 版 DMG（使用模板替换二进制）
echo ""
echo "================================"
echo "创建 Intel Mac DMG..."
echo "================================"

INTEL_DMG_TEMPLATE="$BUILD_DIR/AutoGo图色助手_MacOs_Intel_tmp.dmg"
INTEL_DMG_OUTPUT="$BUILD_DIR/${APP_NAME}_Intel.dmg"

if [ -f "$INTEL_DMG_TEMPLATE" ]; then
    echo "  使用模板: $INTEL_DMG_TEMPLATE"
    
    # 编译 Intel 二进制
    echo "  编译 Intel 二进制..."
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="$LDFLAGS" -o "${APP_NAME}_intel_temp" .
    
    if [ $? -eq 0 ]; then
        # 将模板 DMG 转换为可读写格式
        INTEL_DMG_RW="${BUILD_DIR}/${APP_NAME}_Intel_rw.dmg"
        rm -f "$INTEL_DMG_RW" "$INTEL_DMG_OUTPUT"
        hdiutil convert "$INTEL_DMG_TEMPLATE" -format UDRW -o "$INTEL_DMG_RW" > /dev/null 2>&1
        
        # 挂载可读写 DMG
        INTEL_MOUNT=$(hdiutil attach "$INTEL_DMG_RW" -readwrite -nobrowse 2>/dev/null | grep "Volumes" | tail -1 | awk '{for(i=3;i<=NF;i++) printf "%s ", $i; print ""}' | sed 's/ *$//')
        
        if [ -n "$INTEL_MOUNT" ]; then
            # 查找并替换二进制文件
            INTEL_EXEC=$(find "$INTEL_MOUNT" -name "app" -type f 2>/dev/null | head -1)
            if [ -n "$INTEL_EXEC" ]; then
                cp "${APP_NAME}_intel_temp" "$INTEL_EXEC"
                chmod +x "$INTEL_EXEC"
                echo "  ✓ 已替换 Intel 二进制"
            fi
            
            # 卸载 DMG
            hdiutil detach "$INTEL_MOUNT" > /dev/null 2>&1
            
            # 转回压缩只读格式
            hdiutil convert "$INTEL_DMG_RW" -format UDZO -o "$INTEL_DMG_OUTPUT" > /dev/null 2>&1
            rm -f "$INTEL_DMG_RW"
            
            if [ -f "$INTEL_DMG_OUTPUT" ]; then
                echo "  ✓ Intel DMG 创建完成: $INTEL_DMG_OUTPUT"
            else
                echo "  ⚠️ DMG 压缩转换失败"
            fi
        else
            echo "  ⚠️ 无法挂载模板 DMG"
            rm -f "$INTEL_DMG_RW"
        fi
        
        # 清理临时文件
        rm -f "${APP_NAME}_intel_temp"
    else
        echo "  ⚠️ Intel 二进制编译失败"
    fi
else
    echo "  ⚠️ 未找到 Intel DMG 模板: $INTEL_DMG_TEMPLATE"
    echo "  跳过 Intel DMG 创建"
fi
echo "================================"

# 7. 编译 Windows 版本
echo ""
echo "================================"
echo "开始编译 Windows 版本..."
echo "================================"
echo ""

# 步骤 0: 准备 Windows 图标资源
echo "步骤 0/3: 准备 Windows 图标资源..."

# 检查 rsrc 工具是否安装
if ! command -v rsrc &> /dev/null; then
    echo "  ⚠️  rsrc 工具未安装，正在安装..."
    go install github.com/akavel/rsrc@latest
    if [ $? -ne 0 ]; then
        echo "  ✗ rsrc 安装失败，将使用无图标版本编译"
        ICO_FILE=""
    fi
fi

# 检查是否需要转换 PNG 到 ICO
ICO_FILE="$BUILD_DIR/logo.ico"
if [ ! -f "$ICO_FILE" ]; then
    echo "  转换 PNG 图标到 ICO 格式（添加圆角）..."
    python3 convert_icon.py "$ICON_PATH" "$ICO_FILE" 2>&1 | sed 's/^/  /'
    if [ ${PIPESTATUS[0]} -ne 0 ]; then
        ICO_FILE=""
    fi
else
    echo "  ✓ ICO 图标已存在"
fi

# 生成 Windows 资源文件
if [ -n "$ICO_FILE" ] && [ -f "$ICO_FILE" ] && command -v rsrc &> /dev/null; then
    echo "  生成 Windows 资源文件..."
    rsrc -ico "$ICO_FILE" -manifest windows_icon.manifest -arch amd64 -o rsrc_windows_amd64.syso 2>&1 | grep -v "^$" || true
    if [ -f "rsrc_windows_amd64.syso" ]; then
        echo "  ✓ 资源文件生成成功"
    else
        echo "  ⚠️  资源文件生成失败，将使用无图标版本"
        rm -f rsrc_windows_amd64.syso
    fi
else
    echo "  ⚠️  跳过图标资源生成（缺少 ICO 文件或 rsrc 工具）"
fi

echo "步骤 1/3: 编译 Windows AMD64 架构..."
# Windows 特定参数：-H windowsgui 隐藏控制台窗口
WIN_LDFLAGS="-s -w -H windowsgui"
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="$WIN_LDFLAGS" -o "$BUILD_DIR/${APP_NAME}.exe" .

BUILD_STATUS=$?
if [ $BUILD_STATUS -ne 0 ]; then
    echo "  ⚠️  Windows 编译失败（可能缺少 MinGW-w64）"
    echo "  提示：如需编译 Windows 版本，请先安装 MinGW-w64："
    echo "    brew install mingw-w64"
    echo ""
    echo "  已跳过 Windows 版本编译"
    # 清理资源文件
    rm -f rsrc_windows_amd64.syso
else
    echo "✓ Windows AMD64 架构编译完成"
    
    # 清理资源文件（已链接到 exe 中）
    rm -f rsrc_windows_amd64.syso
    
    # 验证生成的文件
    if [ -f "$BUILD_DIR/${APP_NAME}.exe" ]; then
        WIN_SIZE=$(du -h "$BUILD_DIR/${APP_NAME}.exe" | cut -f1)
        echo "  生成文件: $BUILD_DIR/${APP_NAME}.exe ($WIN_SIZE)"
        echo "  优化说明: 已去除调试符号，隐藏控制台窗口，已嵌入图标"
        
        # 复制 build 目录下的图标到同一目录（可选）
        if [ -f "$ICON_PATH" ]; then
            cp "$ICON_PATH" "$BUILD_DIR/logo.png"
        fi
        
        echo ""
        echo "Windows 版本说明："
        echo "  - 将 ${APP_NAME}.exe 和 adb.exe 放在同一目录"
        echo "  - 或确保 adb.exe 在系统 PATH 中"
        echo "  - 运行时不会显示控制台窗口"
        echo "  - 已优化体积（去除调试符号）"
        echo "  - 已嵌入应用图标"
    fi
fi

echo ""
echo "================================"
echo "✓ 所有构建任务完成！"
echo ""
echo "生成的文件："
echo "  macOS: $BUILD_DIR/${APP_NAME}.app"
echo "  macOS: $BUILD_DIR/${APP_NAME}.dmg"
if [ -f "$BUILD_DIR/${APP_NAME}.exe" ]; then
    echo "  Windows: $BUILD_DIR/${APP_NAME}.exe"
fi
echo "================================"

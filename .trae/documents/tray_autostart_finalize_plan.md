# 托盘图标与开机自启动功能最终实施计划

## 摘要
用户要求为 SugarPlayer 新增：
1. 设置中可开关的“开机自动启动”功能。
2. 系统托盘图标，支持设置开启/关闭。
3. 设置“点击关闭按钮”行为：隐藏到托盘或退出应用。
4. 右键托盘菜单显示当前歌曲、上一首、下一首、退出。

经代码库检查，上述功能的大部分实现已存在于 `tray.go`、`autostart_windows.go`、`autostart_stub.go`、`app.go`、`Settings.vue`、`App.vue`、`useSession.ts` 等文件中。本计划以**验证、补齐、修复**为主，确保功能可编译、可运行。

## 当前状态分析

### 已完成内容
- `config.go`：`ConfigSettings` 已包含 `AutoStart`、`TrayEnabled`、`CloseToTray`。
- `useConfig.ts`：`AppSettings` 与 `load()` 默认值已包含三项新设置。
- `tray.go`：已创建托盘、设置图标与提示、构建右键菜单、发射 `tray:prev`/`tray:next`/`tray:exit` 事件、左键点击显示主窗口。
- `autostart_windows.go`：已实现 Windows 注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\SugarMusic` 的写入/删除。
- `autostart_stub.go`：非 Windows 平台 `ApplyAutoStart` 空实现已存在。
- `app.go`：`App` 结构体已增加托盘字段，`SetCloseToTray` 已暴露。
- `useSession.ts`：`handleClose(forceQuit)` 已支持托盘隐藏逻辑。
- `Settings.vue`：已新增“系统”设置卡片，含三个开关。
- `App.vue`：已监听设置变化、同步后端、处理托盘事件、同步当前歌曲到托盘标签。
- `frontend/scripts/prepare-bindings-dts.js` 与 `frontend/bindings/sugarplayer/app.js`：四个新绑定方法已生成。
- `main.go`：主窗口 `Name: "main"` 已设置。

### 需要验证/修复的潜在风险
1. **服务启动时窗口引用时机**：`app.go` 的 `ServiceStartup` 中通过 `a.app.Window.GetByName("main")` 获取主窗口并注册 `WindowClosing` 钩子。若服务启动发生在窗口创建之前，钩子会注册失败，导致“关闭按钮最小化到托盘”失效。需要确认 Wails v3 生命周期，必要时改用事件监听窗口创建或在更晚生命周期注册钩子。
2. **托盘事件无数据发射的兼容性**：`tray.go` 使用 `a.app.Event.Emit("tray:prev")` 等无数据事件。根据项目经验，`RegisterEvent[any]` 无数据会触发 nil panic，但未注册事件监听可正常。当前前端使用 `Events.On` 监听，应可工作，但需编译与运行验证。
3. **绑定类型文件缺失**：`frontend/bindings/sugarplayer/models.d.ts` 不存在，但 `prepare-bindings-dts.js` 会在构建时生成。需确认前端构建流程正确执行该脚本。
4. **构建验证**：需要完整运行前端构建与 Go 构建，确认无编译错误。
5. **运行时验证**：需要实际启动应用，测试注册表、托盘图标、关闭行为、菜单功能。

## 修改计划

### 1. 验证并修复 `app.go` 窗口关闭钩子注册
- **文件**：`d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\app.go`
- **操作**：
  - 确认 `ServiceStartup` 中 `GetByName("main")` 是否能正确获取窗口。
  - 若无法获取，改用 `a.app.Events.On("wails:window:created")` 或类似事件，在窗口创建后再注册 `WindowClosing` 钩子；或将主窗口引用保存延迟到首次需要时（如 `showMainWindow` / `SetCloseToTray`）重新获取。
  - 确保 `SetCloseToTray` 在启用时即时生效。

### 2. 验证 `tray.go` 与 Wails v3 API 兼容性
- **文件**：`d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\tray.go`
- **操作**：
  - 确认 `a.app.SystemTray.New()`、`tray.SetIcon`、`tray.SetTooltip`、`tray.SetMenu`、`tray.OnClick`、`tray.Run`、`tray.Destroy` 在当前 Wails v3 alpha 版本中可用。
  - 确认 `menu.Add(...)` 返回的 `*application.MenuItem` 的 `OnClick` 签名正确。
  - 若编译报错，根据 Wails v3 实际 API 调整。

### 3. 重新生成前端绑定类型
- **文件**：`d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\frontend\scripts\prepare-bindings-dts.js`
- **操作**：
  - 运行 `cd frontend && npm run build`（或项目约定的构建命令），确保 `prepare-bindings-dts.js` 被执行，`app.d.ts` 与 `models.d.ts` 正确生成。
  - 如有缺失，手动执行 `node scripts/prepare-bindings-dts.js` 或 `wails3 generate bindings`。

### 4. 完整构建验证
- **操作**：
  - 运行 `cd frontend && npm run build`。
  - 运行 `wails3 build`（项目约定使用 `wails3 build` 而非 `wails build`）。
  - 修复所有编译错误。

### 5. 运行时功能验证
- **操作**：
  - 启动应用，打开设置 → “系统”卡片。
  - 开启“开机自动启动”，检查注册表项是否正确写入/删除。
  - 开启“启用系统托盘”，确认托盘图标出现，左键点击显示主窗口，右键菜单包含当前歌曲、上一首、下一首、退出。
  - 开启“关闭按钮最小化到托盘”，点击标题栏关闭按钮，确认窗口隐藏且进程仍在；托盘右键“退出”正常保存配置并退出。
  - 禁用“启用系统托盘”，确认关闭按钮直接退出应用。

## 假设与决策
- **平台范围**：主要针对 Windows。非 Windows 平台通过 `autostart_stub.go` 保持编译通过，但托盘功能行为取决于 Wails v3 跨平台支持。
- **关闭逻辑**：关闭按钮行为统一由前端 `useSession.ts::handleClose` 判断，后端 `WindowClosing` 钩子作为兜底/辅助，确保即使前端事件未拦截也能隐藏到托盘。
- **图标资源**：依赖 `build/windows/icon.ico` 已存在。若缺失，需补充图标文件。
- **绑定生成**：依赖 `prepare-bindings-dts.js` 与 `wails3 generate bindings` 生成一致的前后端绑定。

## 验证步骤
1. 前端构建通过，无 TypeScript/ESLint 错误。
2. Go 构建通过，无编译错误。
3. 注册表写入/删除验证（Windows）。
4. 托盘图标显示、左键显示窗口、右键菜单功能验证。
5. 关闭按钮行为验证（隐藏到托盘 vs 退出）。
6. 关闭“启用系统托盘”后，关闭按钮恢复直接退出。

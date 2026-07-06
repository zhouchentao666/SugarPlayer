# 托盘图标与开机自启动功能实现计划

## 背景与目标
用户希望在 SugarPlayer 中新增：
1. 开机自动启动（Windows 注册表 Run 键）
2. 系统托盘图标，支持设置开启/关闭
3. 关闭按钮行为设置：隐藏到托盘或退出应用
4. 右键托盘菜单显示当前歌曲、上一首、下一首、退出

## 关键设计决策
- 后端负责：创建/销毁托盘、读写 Windows 启动项注册表、拦截窗口关闭事件、向前端发送托盘控制事件
- 前端负责：持久化三项新设置、在歌曲变化时同步托盘标签、监听托盘事件执行播放控制
- 关闭逻辑统一放在 `useSession.ts::handleClose`，托盘图标禁用时强制退出
- 启动项和托盘功能在 Windows 外的平台提供空实现，保证编译通过

## 修改文件清单

### Go 后端
1. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\config.go`
   - `ConfigSettings` 新增 `AutoStart`、`TrayEnabled`、`CloseToTray` 字段

2. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\app.go`
   - `App` 结构体新增托盘相关字段（`tray *application.SystemTray`、`traySongLabel *application.MenuItem`、`trayIcon []byte`）
   - `ServiceStartup` 中读取并缓存图标字节，通过 `GetByName("main")` 保存主窗口引用，注册 `WindowClosing` hook

3. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\tray.go`（新建）
   - `//go:embed build/windows/icon.ico` 嵌入托盘图标
   - 暴露 `EnableTray(enabled bool) error`、`SetTraySongInfo(label string) error`
   - 构建托盘菜单：当前歌曲（禁用）、分隔线、上一首、下一首、分隔线、退出
   - 菜单事件通过 `a.app.Event.Emit("tray:prev" / "tray:next" / "tray:exit")` 通知前端
   - 左键点击显示并聚焦主窗口

4. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\autostart_windows.go`（新建）
   - `//go:build windows`
   - 暴露 `ApplyAutoStart(enabled bool) error`
   - 操作 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\SugarMusic`

5. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\autostart_stub.go`（新建）
   - `//go:build !windows`
   - `ApplyAutoStart` 直接返回 `nil`

6. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\main.go`
   - 主窗口 `Name: "main"`

### 前端
7. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\frontend\src\composables\useConfig.ts`
   - `AppSettings` 新增 `autoStart`、`trayEnabled`、`closeToTray`
   - `load()` 默认值全部设为 `false`

8. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\frontend\src\composables\useSession.ts`
   - `handleClose(forceQuit = false)`
   - 非强制退出且 `trayEnabled && closeToTray` 时调用 `Window.Hide()`
   - 否则走原有保存并 `Application.Quit()` 逻辑

9. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\frontend\src\components\Settings.vue`
   - 新增“常规”或“系统托盘”设置卡片
   - 开机自动启动开关（`autoStart`）
   - 启用系统托盘开关（`trayEnabled`）
   - 关闭按钮最小化到托盘开关（`closeToTray`，托盘禁用时禁用）

10. `d:\zhouchentao\biancheng\Web\SugarPlayer\sugarplayer\frontend\src\App.vue`
    - 默认 `settings` 添加三项新字段
    - 引入绑定 `ApplyAutoStart`、`EnableTray`、`SetTraySongInfo`
    - `onMounted` 加载配置后应用 `ApplyAutoStart` 和 `EnableTray`
    - `watch` `settings.autoStart`、`settings.trayEnabled`，变化时同步后端
    - `watch` `audio.currentSong`，构造 `歌曲名 - 艺术家` 字符串调用 `SetTraySongInfo`
    - 监听 `Events.On('tray:prev', playPrev)`、`tray:next`、`tray:exit`，并在 `onUnmounted` 取消

## 事件约定
| 事件名 | 方向 | 用途 |
|---|---|---|
| `tray:prev` | 后端 → 前端 | 托盘菜单“上一首” |
| `tray:next` | 后端 → 前端 | 托盘菜单“下一首” |
| `tray:exit` | 后端 → 前端 | 托盘菜单“退出” |

## 验证步骤
1. 运行 `cd frontend && npm run build` 与 `go build ./...`，确认无编译错误
2. 启动应用后，在设置中开启“开机自动启动”，检查注册表 `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\SugarMusic` 是否存在且指向当前 exe
3. 开启“启用系统托盘”，确认任务栏托盘出现 SugarMusic 图标
4. 左键点击托盘图标，主窗口显示并聚焦
5. 右键托盘图标，确认菜单包含当前歌曲、上一首、下一首、退出；点击上一首/下一首测试播放切换
6. 开启“关闭按钮最小化到托盘”，点击标题栏关闭按钮，窗口隐藏但进程仍在；托盘右键退出可正常退出并保存配置
7. 禁用“系统托盘”后，关闭按钮应直接退出应用，无论“关闭按钮最小化到托盘”状态如何

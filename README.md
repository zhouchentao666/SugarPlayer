# SugarPlayer / SugarMusic

一款基于 Wails v3 + Vue 3 开发的本地音乐播放器，支持歌词、播放列表管理与现代化窗口效果。

## 项目截图

<img width="1707" height="1150" alt="image" src="https://github.com/user-attachments/assets/b1eb7f07-d112-4343-a8b8-818a397ef20f" />
<img width="1392" height="818" alt="ea04e758c845cac74b3edceeccdab225" src="https://github.com/user-attachments/assets/4df6529d-9def-462a-b74f-1d75d56514eb" />
<img width="1345" height="872" alt="image" src="https://github.com/user-attachments/assets/3cc56508-b299-49f5-8952-6f4f8285454e" />
<img width="1343" height="877" alt="image" src="https://github.com/user-attachments/assets/db74df47-a0ca-4757-9fc4-90e7769250e2" />
<img width="1526" height="1008" alt="image" src="https://github.com/user-attachments/assets/b51edbbf-5ce1-4d88-90c3-f8066ae5abc5" />
<img width="1410" height="823" alt="image" src="https://github.com/user-attachments/assets/d41703c7-62c8-4f0a-8d1b-92a022aaf2bb" />


## 功能特性

- 本地音乐播放：支持 MP3、FLAC、WAV、AAC、OGG、M4A、WMA、OPUS 等常见格式
- 歌词显示：支持普通 LRC、增强 LRC、YRC、LRC A2，集成 AMLL 歌词组件
- 播放列表：文件夹扫描、创建与管理多个播放列表
- 全屏播放器：封面 FLIP 动画、自适应布局、AMLL 歌词、倍速播放
- 倍速播放：0.25x ~ 16x 滑块调节
- 窗口效果：Windows Acrylic、自定义背景图、歌曲主题色、无边框窗口
- 状态持久化：可保存播放列表、当前歌曲、窗口位置与播放进度

## 技术栈

- 后端：Go 1.25 + Wails v3 (`v3.0.0-alpha2.108`)
- 前端：Vue 3 + TypeScript + Vite
- 歌词：@applemusic-like-lyrics
- 元数据：go.senan.xyz/taglib

## 开发环境

- 安装 Go 1.25+
- 安装 Node.js 与 npm
- 安装 Wails v3 CLI：`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`

## 常用命令

```bash
# 安装前端依赖
cd frontend && npm install

# 前端开发（Vite）
npm run dev

# 前端构建
npm run build

# 完整构建可执行程序（使用 wails3）
wails3 build
```

## 项目结构

```
sugarplayer/
├── app.go           # Wails Service 入口与生命周期
├── audio.go         # 本地音频流服务、元数据读取
├── config.go        # 配置持久化
├── dialogs.go       # 文件/文件夹选择对话框
├── watcher.go       # 文件夹监控
├── main.go          # 应用入口与窗口配置
├── wails.json       # Wails 项目配置
├── Taskfile.yml     # 构建任务
├── build/           # 平台构建资源与安装程序
└── frontend/        # Vue 前端源码
    ├── src/components/   # 界面组件
    ├── src/composables/  # 业务逻辑组合式函数
    ├── src/utils/        # 工具函数
    └── package.json
```

## 注意事项

- 请使用 `wails3 build` 进行打包，不要使用旧版 `wails build`。
- Windows 亚克力效果需要在 `main.go` 中设置 `BackgroundTypeTranslucent` 与 `BackdropType: application.Acrylic`。
- `frontend/vite.config.ts` 中已固定 `target: 'es2021'` 与 `cssTarget: 'chrome100'`，以兼容 AMLL 依赖。
- 构建时若出现 AMLL CSS nesting 警告，不影响运行。

## 版本

当前版本：`0.0.2`

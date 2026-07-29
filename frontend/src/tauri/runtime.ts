// @wailsio/runtime 兼容层：Events / Window / System / Application
// 使前端业务代码在改 import 路径后无需调整调用方式。
import { emit, listen } from '@tauri-apps/api/event'
import { getCurrentWindow } from '@tauri-apps/api/window'
import { invoke } from '@tauri-apps/api/core'

// ---------- Events（Wails 风格：回调收到 { data }，On 返回 off 函数） ----------

type OffFn = () => void

const listenersByName = new Map<string, Set<OffFn>>()

export const Events = {
  On(name: string, callback: (event: { data: any }) => void): OffFn {
    let unlisten: OffFn | null = null
    let cancelled = false
    listen(name, (e) => callback({ data: e.payload })).then((un) => {
      if (cancelled) un()
      else unlisten = un
    })
    const off: OffFn = () => {
      cancelled = true
      unlisten?.()
      unlisten = null
      listenersByName.get(name)?.delete(off)
    }
    let set = listenersByName.get(name)
    if (!set) {
      set = new Set()
      listenersByName.set(name, set)
    }
    set.add(off)
    return off
  },
  Off(name: string) {
    const set = listenersByName.get(name)
    if (set) {
      for (const off of [...set]) off()
      listenersByName.delete(name)
    }
  },
  Emit(name: string, data?: any): Promise<void> {
    return emit(name, data)
  },
}

// ---------- Window ----------

export const Window = {
  Minimise: () => getCurrentWindow().minimize(),
  ToggleMaximise: () => getCurrentWindow().toggleMaximize(),
  IsMaximised: () => getCurrentWindow().isMaximized(),
  Close: () => getCurrentWindow().close(),
  Hide: () => getCurrentWindow().hide(),
  Show: () => getCurrentWindow().show(),
  Fullscreen: () => getCurrentWindow().setFullscreen(true),
  UnFullscreen: () => getCurrentWindow().setFullscreen(false),
  SetAlwaysOnTop: (onTop: boolean) => getCurrentWindow().setAlwaysOnTop(onTop),
  async Position(): Promise<{ x: number; y: number }> {
    const win = getCurrentWindow()
    const scale = await win.scaleFactor()
    const pos = (await win.outerPosition()).toLogical(scale)
    return { x: Math.round(pos.x), y: Math.round(pos.y) }
  },
  async Size(): Promise<{ width: number; height: number }> {
    const win = getCurrentWindow()
    const scale = await win.scaleFactor()
    const size = (await win.outerSize()).toLogical(scale)
    return { width: Math.round(size.width), height: Math.round(size.height) }
  },
  SetPosition: (x: number, y: number) =>
    getCurrentWindow().setPosition({ type: 'Logical', data: { x, y } } as any),
  SetSize: (width: number, height: number) =>
    getCurrentWindow().setSize({ type: 'Logical', data: { width, height } } as any),
}

// ---------- System / Application ----------

export const System = {
  IsWindows: () => navigator.userAgent.includes('Windows'),
  IsMac: () => navigator.userAgent.includes('Mac'),
  IsLinux: () => navigator.userAgent.includes('Linux'),
}

export const Application = {
  Quit: () => invoke('quit_app'),
}

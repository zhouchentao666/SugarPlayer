//! 系统托盘（对齐 Go tray.go）：当前歌曲、上一首/下一首、桌面歌词开关/锁定、主界面、退出。

use std::sync::Mutex;
use tauri::menu::{MenuBuilder, MenuItem};
use tauri::tray::{MouseButton, MouseButtonState, TrayIcon, TrayIconBuilder, TrayIconEvent};
use tauri::{AppHandle, Emitter, Manager, Wry};

use crate::config;
use crate::lyric_window;

const TRAY_ID: &str = "sugarplayer-tray";

#[derive(Default)]
pub struct TrayState {
    pub tray: Mutex<Option<TrayIcon>>,
    pub song_item: Mutex<Option<MenuItem<Wry>>>,
    pub lyric_toggle_item: Mutex<Option<MenuItem<Wry>>>,
    pub lyric_lock_item: Mutex<Option<MenuItem<Wry>>>,
    pub close_to_tray: Mutex<bool>,
}

impl TrayState {
    pub fn has_tray(&self) -> bool {
        self.tray.lock().unwrap().is_some()
    }
}

pub fn enable_tray(app: &AppHandle, enabled: bool) -> Result<(), String> {
    let state = app.state::<TrayState>();
    if enabled {
        if state.has_tray() {
            return Ok(());
        }
        create_tray(app)?;
        update_lyric_menu(app);
        return Ok(());
    }
    // 关闭托盘
    *state.tray.lock().unwrap() = None;
    *state.song_item.lock().unwrap() = None;
    *state.lyric_toggle_item.lock().unwrap() = None;
    *state.lyric_lock_item.lock().unwrap() = None;
    if let Some(tray) = app.remove_tray_by_id(TRAY_ID) {
        drop(tray);
    }
    Ok(())
}

fn create_tray(app: &AppHandle) -> Result<(), String> {
    let song = MenuItem::with_id(app, "tray-song", "未在播放", false, None::<&str>)
        .map_err(|e| e.to_string())?;
    let prev = MenuItem::with_id(app, "tray-prev", "上一首", true, None::<&str>)
        .map_err(|e| e.to_string())?;
    let next = MenuItem::with_id(app, "tray-next", "下一首", true, None::<&str>)
        .map_err(|e| e.to_string())?;
    let lyric_toggle =
        MenuItem::with_id(app, "tray-lyric-toggle", "显示桌面歌词", true, None::<&str>)
            .map_err(|e| e.to_string())?;
    let lyric_lock =
        MenuItem::with_id(app, "tray-lyric-lock", "锁定桌面歌词", true, None::<&str>)
            .map_err(|e| e.to_string())?;
    let show_main = MenuItem::with_id(app, "tray-show-main", "主界面", true, None::<&str>)
        .map_err(|e| e.to_string())?;
    let exit = MenuItem::with_id(app, "tray-exit", "退出", true, None::<&str>)
        .map_err(|e| e.to_string())?;

    let menu = MenuBuilder::new(app)
        .item(&song)
        .separator()
        .item(&prev)
        .item(&next)
        .separator()
        .item(&lyric_toggle)
        .item(&lyric_lock)
        .separator()
        .item(&show_main)
        .separator()
        .item(&exit)
        .build()
        .map_err(|e| e.to_string())?;

    let mut builder = TrayIconBuilder::with_id(TRAY_ID)
        .tooltip("SugarPlayer")
        .menu(&menu)
        .show_menu_on_left_click(false)
        .on_menu_event(|app, event| match event.id().as_ref() {
            "tray-prev" => {
                let _ = app.emit("tray:prev", ());
            }
            "tray-next" => {
                let _ = app.emit("tray:next", ());
            }
            "tray-lyric-toggle" => {
                let cfg = config::load_desktop_lyric();
                let enabled = cfg
                    .get("enabled")
                    .and_then(|v| v.as_bool())
                    .unwrap_or(false);
                let _ = lyric_window::toggle(app, !enabled);
                update_lyric_menu(app);
            }
            "tray-lyric-lock" => {
                let cfg = config::load_desktop_lyric();
                let locked = cfg.get("isLock").and_then(|v| v.as_bool()).unwrap_or(false);
                let _ = lyric_window::set_ignore_mouse_events(app, !locked);
                update_lyric_menu(app);
            }
            "tray-show-main" => show_main_window(app),
            "tray-exit" => {
                let _ = app.emit("tray:exit", ());
            }
            _ => {}
        })
        .on_tray_icon_event(|tray, event| {
            if let TrayIconEvent::Click {
                button: MouseButton::Left,
                button_state: MouseButtonState::Up,
                ..
            } = event
            {
                show_main_window(tray.app_handle());
            }
        });

    if let Some(icon) = app.default_window_icon() {
        builder = builder.icon(icon.clone());
    }

    let tray = builder.build(app).map_err(|e| e.to_string())?;

    let state = app.state::<TrayState>();
    *state.tray.lock().unwrap() = Some(tray);
    *state.song_item.lock().unwrap() = Some(song);
    *state.lyric_toggle_item.lock().unwrap() = Some(lyric_toggle);
    *state.lyric_lock_item.lock().unwrap() = Some(lyric_lock);
    Ok(())
}

pub fn set_song_info(app: &AppHandle, label: &str) {
    let state = app.state::<TrayState>();
    if let Some(item) = state.song_item.lock().unwrap().as_ref() {
        let _ = item.set_text(label);
    }
}

/// 按桌面歌词配置刷新托盘菜单文案。
pub fn update_lyric_menu(app: &AppHandle) {
    let state = app.state::<TrayState>();
    if !state.has_tray() {
        return;
    }
    let cfg = config::load_desktop_lyric();
    let enabled = cfg
        .get("enabled")
        .and_then(|v| v.as_bool())
        .unwrap_or(false);
    let locked = cfg.get("isLock").and_then(|v| v.as_bool()).unwrap_or(false);
    if let Some(item) = state.lyric_toggle_item.lock().unwrap().as_ref() {
        let _ = item.set_text(if enabled { "隐藏桌面歌词" } else { "显示桌面歌词" });
    }
    if let Some(item) = state.lyric_lock_item.lock().unwrap().as_ref() {
        let _ = item.set_text(if locked { "解锁桌面歌词" } else { "锁定桌面歌词" });
    }
}

pub fn show_main_window(app: &AppHandle) {
    if let Some(win) = app.get_webview_window("main") {
        let _ = win.show();
        let _ = win.unminimize();
        let _ = win.set_focus();
    }
}

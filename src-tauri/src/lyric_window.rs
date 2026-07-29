//! 桌面歌词窗口（对齐 Go lyric_window.go）：
//! 无边框、透明、置顶、不进任务栏，位置尺寸持久化到 settings.desktopLyric。

use serde_json::{json, Value};
use std::sync::atomic::{AtomicU64, Ordering};
use tauri::{
    AppHandle, Emitter, LogicalPosition, LogicalSize, Manager, WebviewUrl, WebviewWindowBuilder,
};

use crate::config;
use crate::tray;

pub const LABEL: &str = "desktop-lyric";

fn num(cfg: &serde_json::Map<String, Value>, key: &str, default: f64) -> f64 {
    cfg.get(key).and_then(|v| v.as_f64()).unwrap_or(default)
}

fn flag(cfg: &serde_json::Map<String, Value>, key: &str) -> bool {
    cfg.get(key).and_then(|v| v.as_bool()).unwrap_or(false)
}

/// 显示或隐藏桌面歌词窗口。
pub fn toggle(app: &AppHandle, enabled: bool) -> Result<(), String> {
    if enabled {
        if let Some(win) = app.get_webview_window(LABEL) {
            let _ = win.show();
            let _ = win.set_always_on_top(true);
            return Ok(());
        }
        return open(app);
    }
    if let Some(win) = app.get_webview_window(LABEL) {
        let _ = win.hide();
    }
    let mut cfg = config::load_desktop_lyric();
    cfg.insert("enabled".into(), Value::Bool(false));
    config::save_desktop_lyric(&cfg);
    tray::update_lyric_menu(app);
    Ok(())
}

fn open(app: &AppHandle) -> Result<(), String> {
    let cfg = config::load_desktop_lyric();
    let mut w = num(&cfg, "width", 800.0);
    let mut h = num(&cfg, "height", 180.0);
    if w <= 0.0 {
        w = 800.0;
    }
    if h <= 0.0 {
        h = 180.0;
    }
    let mut x = num(&cfg, "x", 0.0);
    let mut y = num(&cfg, "y", 0.0);

    // 位置兜底：首次显示放在主屏底部居中（对齐 Go ensureDesktopLyricPosition）
    let (mx, my, mw, mh) = app
        .primary_monitor()
        .map(|m| {
            let scale = m.scale_factor();
            let pos = m.position().to_logical::<f64>(scale);
            let size = m.size().to_logical::<f64>(scale);
            (pos.x, pos.y, size.width, size.height)
        })
        .unwrap_or((0.0, 0.0, 1920.0, 1080.0));
    if x == 0.0 && y == 0.0 {
        x = mx + mw / 2.0 - w / 2.0;
        y = my + mh - h - 90.0;
    }
    if x + w > mx + mw {
        x = mx + mw - w;
    }
    if y + h > my + mh {
        y = my + mh - h;
    }
    if x < mx {
        x = mx;
    }
    if y < my {
        y = my;
    }

    let win = WebviewWindowBuilder::new(
        app,
        LABEL,
        WebviewUrl::App("index.html?desktopLyric=1".into()),
    )
    .title("桌面歌词")
    .decorations(false)
    .transparent(true)
    .shadow(false)
    .always_on_top(true)
    .skip_taskbar(true)
    .resizable(true)
    .min_inner_size(440.0, 120.0)
    .max_inner_size(1600.0, 300.0)
    .inner_size(w, h)
    .position(x, y)
    .visible(false)
    .build()
    .map_err(|e| e.to_string())?;

    let _ = win.set_ignore_cursor_events(flag(&cfg, "isLock"));

    // 移动/缩放后防抖保存位置
    let handle = app.clone();
    win.on_window_event(move |event| match event {
        tauri::WindowEvent::Moved(_) | tauri::WindowEvent::Resized(_) => {
            schedule_save_bounds(&handle);
        }
        _ => {}
    });

    let _ = win.show();
    let _ = win.set_always_on_top(true);
    tray::update_lyric_menu(app);
    Ok(())
}

/// 关闭并销毁桌面歌词窗口，enabled 置 false。
pub fn close(app: &AppHandle) -> Result<(), String> {
    if let Some(win) = app.get_webview_window(LABEL) {
        let _ = win.destroy();
    }
    let mut cfg = config::load_desktop_lyric();
    cfg.insert("enabled".into(), Value::Bool(false));
    config::save_desktop_lyric(&cfg);
    tray::update_lyric_menu(app);
    Ok(())
}

/// 设置窗口位置和大小并持久化。
pub fn set_bounds(app: &AppHandle, x: f64, y: f64, width: f64, height: f64) -> Result<(), String> {
    if let Some(win) = app.get_webview_window(LABEL) {
        let _ = win.set_position(LogicalPosition::new(x, y));
        let _ = win.set_size(LogicalSize::new(width, height));
    }
    let mut cfg = config::load_desktop_lyric();
    cfg.insert("x".into(), json!(x));
    cfg.insert("y".into(), json!(y));
    cfg.insert("width".into(), json!(width));
    cfg.insert("height".into(), json!(height));
    config::save_desktop_lyric(&cfg);
    Ok(())
}

/// 锁定/解锁（鼠标穿透）并持久化，广播 lock-changed 事件。
pub fn set_ignore_mouse_events(app: &AppHandle, ignore: bool) -> Result<(), String> {
    if let Some(win) = app.get_webview_window(LABEL) {
        let _ = win.set_ignore_cursor_events(ignore);
    }
    let mut cfg = config::load_desktop_lyric();
    cfg.insert("isLock".into(), Value::Bool(ignore));
    config::save_desktop_lyric(&cfg);
    tray::update_lyric_menu(app);
    let _ = app.emit("desktop-lyric:lock-changed", json!({ "locked": ignore }));
    Ok(())
}

static SAVE_GEN: AtomicU64 = AtomicU64::new(0);

fn schedule_save_bounds(app: &AppHandle) {
    let gen = SAVE_GEN.fetch_add(1, Ordering::SeqCst) + 1;
    let app = app.clone();
    std::thread::spawn(move || {
        std::thread::sleep(std::time::Duration::from_millis(250));
        if SAVE_GEN.load(Ordering::SeqCst) != gen {
            return;
        }
        save_current_bounds(&app);
    });
}

fn save_current_bounds(app: &AppHandle) {
    if let Some(win) = app.get_webview_window(LABEL) {
        let scale = win.scale_factor().unwrap_or(1.0);
        if let (Ok(pos), Ok(size)) = (win.outer_position(), win.outer_size()) {
            let lp = pos.to_logical::<f64>(scale);
            let ls = size.to_logical::<f64>(scale);
            let mut cfg = config::load_desktop_lyric();
            cfg.insert("x".into(), json!(lp.x));
            cfg.insert("y".into(), json!(lp.y));
            cfg.insert("width".into(), json!(ls.width));
            cfg.insert("height".into(), json!(ls.height));
            config::save_desktop_lyric(&cfg);
        }
    }
}

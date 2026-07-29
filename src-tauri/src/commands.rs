//! Tauri 命令：与原 Wails 版 App 绑定方法一一对应（去除在线音乐部分）。

use std::collections::HashMap;
use std::path::Path;

use serde_json::Value;
use tauri::{AppHandle, Emitter, Manager, State, WebviewUrl, WebviewWindowBuilder};

use crate::config;
use crate::lyric_window;
use crate::models::SongMetadata;
use crate::music;
use crate::tray;
use crate::watcher::FolderWatcherState;
use crate::AudioPort;

// ---------- 基础 ----------

#[tauri::command]
pub fn version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

#[tauri::command]
pub fn audio_server_url(state: State<AudioPort>) -> String {
    format!("http://127.0.0.1:{}", state.0)
}

#[tauri::command]
pub fn get_donate_image_urls(state: State<AudioPort>) -> HashMap<String, String> {
    let base = format!("http://127.0.0.1:{}", state.0);
    HashMap::from([
        ("wechat".to_string(), format!("{base}/cover?name=wechat")),
        ("alipay".to_string(), format!("{base}/cover?name=alipay")),
    ])
}

#[tauri::command]
pub fn open_url(app: AppHandle, u: String) -> Result<(), String> {
    use tauri_plugin_opener::OpenerExt;
    app.opener()
        .open_url(u, None)
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub fn quit_app(app: AppHandle) {
    app.exit(0);
}

// ---------- 本地音乐 ----------

#[tauri::command]
pub fn scan_music_folder(path: String) -> Result<Vec<String>, String> {
    music::scan_folder(&path)
}

#[tauri::command]
pub fn read_metadata(path: String) -> Result<SongMetadata, String> {
    music::read_metadata(Path::new(&path))
}

#[tauri::command]
pub fn read_lyrics(path: String) -> String {
    music::read_lyrics(Path::new(&path))
}

#[tauri::command]
pub fn read_cover_art(path: String) -> String {
    music::read_cover_art(Path::new(&path))
}

#[tauri::command]
pub fn read_image_file(path: String) -> Result<String, String> {
    music::read_image_file(Path::new(&path))
}

#[tauri::command]
pub fn read_audio_file(path: String) -> Result<String, String> {
    use base64::{engine::general_purpose::STANDARD as B64, Engine as _};
    let data = std::fs::read(Path::new(&path)).map_err(|e| e.to_string())?;
    let mime = match Path::new(&path)
        .extension()
        .and_then(|e| e.to_str())
        .map(|e| e.to_ascii_lowercase())
        .as_deref()
    {
        Some("mp3") => "audio/mpeg",
        Some("flac") => "audio/flac",
        Some("wav") => "audio/wav",
        Some("ogg") | Some("opus") => "audio/ogg",
        Some("m4a") => "audio/mp4",
        Some("aac") => "audio/aac",
        _ => "application/octet-stream",
    };
    Ok(format!("data:{};base64,{}", mime, B64.encode(&data)))
}

// ---------- 配置 ----------

#[tauri::command]
pub fn load_config() -> Value {
    config::load_config()
}

#[tauri::command]
pub fn save_config(config: Value) -> Result<(), String> {
    config::save_config(&config)
}

// ---------- 对话框 ----------

#[tauri::command]
pub fn open_music_files(app: AppHandle) -> Vec<String> {
    use tauri_plugin_dialog::DialogExt;
    app.dialog()
        .file()
        .add_filter(
            "音频",
            &["mp3", "flac", "wav", "ogg", "m4a", "aac", "opus", "wma"],
        )
        .blocking_pick_files()
        .map(|v| {
            v.into_iter()
                .filter_map(|p| p.into_path().ok())
                .map(|pb| pb.to_string_lossy().to_string())
                .collect()
        })
        .unwrap_or_default()
}

#[tauri::command]
pub fn open_music_folder(app: AppHandle) -> String {
    use tauri_plugin_dialog::DialogExt;
    app.dialog()
        .file()
        .blocking_pick_folder()
        .and_then(|p| p.into_path().ok())
        .map(|pb| pb.to_string_lossy().to_string())
        .unwrap_or_default()
}

#[tauri::command]
pub fn open_image_file(app: AppHandle) -> String {
    use tauri_plugin_dialog::DialogExt;
    app.dialog()
        .file()
        .add_filter("图片", &["png", "jpg", "jpeg", "webp", "bmp", "gif"])
        .blocking_pick_file()
        .and_then(|p| p.into_path().ok())
        .map(|pb| pb.to_string_lossy().to_string())
        .unwrap_or_default()
}

#[tauri::command]
pub fn open_in_explorer(path: String) -> Result<(), String> {
    #[cfg(target_os = "windows")]
    {
        std::process::Command::new("explorer")
            .args(["/select,", &path])
            .spawn()
            .map_err(|e| e.to_string())?;
        Ok(())
    }
    #[cfg(target_os = "macos")]
    {
        std::process::Command::new("open")
            .args(["-R", &path])
            .spawn()
            .map_err(|e| e.to_string())?;
        Ok(())
    }
    #[cfg(all(not(target_os = "windows"), not(target_os = "macos")))]
    {
        let dir = Path::new(&path)
            .parent()
            .map(|p| p.to_string_lossy().to_string())
            .unwrap_or(path.clone());
        std::process::Command::new("xdg-open")
            .arg(dir)
            .spawn()
            .map_err(|e| e.to_string())?;
        Ok(())
    }
}

// ---------- 歌曲编辑器窗口 ----------

#[tauri::command]
pub fn open_song_editor(app: AppHandle, path: String) -> Result<(), String> {
    let url = format!("index.html?editor=1&path={}", urlencode(&path));
    if let Some(win) = app.get_webview_window("song-editor") {
        let _ = win.destroy();
    }
    WebviewWindowBuilder::new(&app, "song-editor", WebviewUrl::App(url.as_str().into()))
        .title("编辑歌曲信息")
        .inner_size(560.0, 680.0)
        .min_inner_size(400.0, 500.0)
        .center()
        .build()
        .map_err(|e| e.to_string())?;
    Ok(())
}

#[tauri::command]
pub fn emit_metadata_changed(app: AppHandle) {
    let _ = app.emit("localmetadata:changed", ());
}

// ---------- 文件夹监听 ----------

#[tauri::command]
pub fn watch_music_folder(
    app: AppHandle,
    state: State<FolderWatcherState>,
    path: String,
) -> Result<(), String> {
    state.watch(app, path)
}

#[tauri::command]
pub fn stop_watching(state: State<FolderWatcherState>) {
    state.stop();
}

// ---------- 桌面歌词 ----------

#[tauri::command]
pub fn toggle_desktop_lyric(app: AppHandle, enabled: bool) -> Result<(), String> {
    lyric_window::toggle(&app, enabled)
}

#[tauri::command]
pub fn close_desktop_lyric(app: AppHandle) -> Result<(), String> {
    lyric_window::close(&app)
}

#[tauri::command]
pub fn get_desktop_lyric_config() -> String {
    config::desktop_lyric_config_json()
}

#[tauri::command]
pub fn set_desktop_lyric_bounds(
    app: AppHandle,
    x: f64,
    y: f64,
    width: f64,
    height: f64,
) -> Result<(), String> {
    lyric_window::set_bounds(&app, x, y, width, height)
}

#[tauri::command]
pub fn set_desktop_lyric_ignore_mouse_events(app: AppHandle, ignore: bool) -> Result<(), String> {
    lyric_window::set_ignore_mouse_events(&app, ignore)
}

// ---------- 托盘 / 系统 ----------

#[tauri::command]
pub fn enable_tray(app: AppHandle, enabled: bool) -> Result<(), String> {
    tray::enable_tray(&app, enabled)
}

#[tauri::command]
pub fn set_tray_song_info(app: AppHandle, label: String) {
    tray::set_song_info(&app, &label);
}

#[tauri::command]
pub fn set_close_to_tray(app: AppHandle, enabled: bool) {
    let state = app.state::<tray::TrayState>();
    *state.close_to_tray.lock().unwrap() = enabled;
}

#[tauri::command]
pub fn show_main_window(app: AppHandle) {
    tray::show_main_window(&app);
}

#[tauri::command]
pub fn apply_auto_start(app: AppHandle, enabled: bool) -> Result<(), String> {
    use tauri_plugin_autostart::ManagerExt;
    let manager = app.autolaunch();
    if enabled {
        manager.enable().map_err(|e| e.to_string())
    } else {
        manager.disable().map_err(|e| e.to_string())
    }
}

// ---------- 工具 ----------

fn urlencode(s: &str) -> String {
    let mut out = String::new();
    for b in s.as_bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                out.push(*b as char)
            }
            _ => out.push_str(&format!("%{:02X}", b)),
        }
    }
    out
}

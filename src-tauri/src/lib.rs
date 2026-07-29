mod commands;
mod config;
mod http_server;
mod lyric_window;
mod models;
mod music;
mod tray;
mod watcher;

use tauri::Manager;

pub struct AudioPort(pub u16);

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let port = http_server::start();
    tauri::Builder::default()
        .plugin(tauri_plugin_window_state::Builder::default().build())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_autostart::init(
            tauri_plugin_autostart::MacosLauncher::LaunchAgent,
            None,
        ))
        .manage(AudioPort(port))
        .manage(tray::TrayState::default())
        .manage(watcher::FolderWatcherState::default())
        .on_window_event(|window, event| {
            // 主窗口：勾选“关闭进入托盘”且托盘已启用时，点关闭改为隐藏
            if window.label() == "main" {
                if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                    let state = window.app_handle().state::<tray::TrayState>();
                    let close_to_tray = *state.close_to_tray.lock().unwrap();
                    if close_to_tray && state.has_tray() {
                        api.prevent_close();
                        let _ = window.hide();
                    }
                }
            }
        })
        .invoke_handler(tauri::generate_handler![
            commands::version,
            commands::audio_server_url,
            commands::get_donate_image_urls,
            commands::open_url,
            commands::quit_app,
            commands::scan_music_folder,
            commands::read_metadata,
            commands::read_lyrics,
            commands::read_cover_art,
            commands::read_image_file,
            commands::read_audio_file,
            commands::load_config,
            commands::save_config,
            commands::open_music_files,
            commands::open_music_folder,
            commands::open_image_file,
            commands::open_in_explorer,
            commands::open_song_editor,
            commands::emit_metadata_changed,
            commands::watch_music_folder,
            commands::stop_watching,
            commands::toggle_desktop_lyric,
            commands::close_desktop_lyric,
            commands::get_desktop_lyric_config,
            commands::set_desktop_lyric_bounds,
            commands::set_desktop_lyric_ignore_mouse_events,
            commands::enable_tray,
            commands::set_tray_song_info,
            commands::set_close_to_tray,
            commands::show_main_window,
            commands::apply_auto_start,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

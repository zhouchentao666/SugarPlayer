//! 音乐文件夹监听：文件增删时向前端广播 "folder:changed" 事件（对齐 Go FolderWatcher）。

use notify::{Event, EventKind, RecommendedWatcher, RecursiveMode, Watcher};
use std::path::Path;
use std::sync::Mutex;
use tauri::{AppHandle, Emitter};

#[derive(Default)]
pub struct FolderWatcherState {
    watcher: Mutex<Option<RecommendedWatcher>>,
    paths: Mutex<Vec<String>>,
}

impl FolderWatcherState {
    /// 监听一个文件夹（可多次调用叠加多个文件夹）。
    pub fn watch(&self, app: AppHandle, path: String) -> Result<(), String> {
        let mut guard = self.watcher.lock().unwrap();
        if guard.is_none() {
            let handle = app.clone();
            let watcher = notify::recommended_watcher(move |res: Result<Event, notify::Error>| {
                if let Ok(event) = res {
                    match event.kind {
                        EventKind::Create(_) | EventKind::Remove(_) => {
                            let p = event
                                .paths
                                .first()
                                .map(|p| p.to_string_lossy().to_string())
                                .unwrap_or_default();
                            let _ = handle.emit("folder:changed", p);
                        }
                        _ => {}
                    }
                }
            })
            .map_err(|e| e.to_string())?;
            *guard = Some(watcher);
        }
        if let Some(w) = guard.as_mut() {
            w.watch(Path::new(&path), RecursiveMode::Recursive)
                .map_err(|e| e.to_string())?;
            self.paths.lock().unwrap().push(path);
        }
        Ok(())
    }

    /// 停止全部监听。
    pub fn stop(&self) {
        *self.watcher.lock().unwrap() = None;
        self.paths.lock().unwrap().clear();
    }
}

//! 配置持久化：直接透传前端的 JSON（与原 Wails 版 config.json 结构兼容），
//! 避免 Rust 结构体字段与前端不一致导致丢字段。

use serde_json::{json, Map, Value};
use std::fs;
use std::path::PathBuf;

/// 返回配置文件路径：<config_dir>/SugarPlayer/config.json
pub fn config_path() -> PathBuf {
    let mut dir = dirs::config_dir().unwrap_or_else(|| PathBuf::from("."));
    dir.push("SugarPlayer");
    let _ = fs::create_dir_all(&dir);
    dir.push("config.json");
    dir
}

pub fn load_config() -> Value {
    let p = config_path();
    match fs::read_to_string(&p) {
        Ok(s) => serde_json::from_str(&s).unwrap_or(Value::Null),
        Err(_) => Value::Null,
    }
}

pub fn save_config(config: &Value) -> Result<(), String> {
    let p = config_path();
    if let Some(parent) = p.parent() {
        let _ = fs::create_dir_all(parent);
    }
    let s = serde_json::to_string_pretty(config).map_err(|e| e.to_string())?;
    fs::write(&p, s).map_err(|e| e.to_string())
}

/// 桌面歌词默认配置（与 Go defaultDesktopLyricConfig 一致）。
fn default_desktop_lyric() -> Value {
    json!({
        "enabled": false,
        "fontSize": 30,
        "mainColor": "#73BCFC",
        "unplayedColor": "rgba(255, 255, 255, 0.5)",
        "shadowColor": "rgba(255, 255, 255, 0.5)",
        "fontWeight": 600,
        "position": "center",
        "alwaysShowPlayInfo": false,
        "animation": true,
        "showYrc": true,
        "showTran": false,
        "isDoubleLine": true,
        "textBackgroundMask": false,
        "backgroundMaskColor": "rgba(0,0,0,0.2)",
        "fontFamily": "PingFangSC-Semibold, system-ui, -apple-system, sans-serif",
        "x": 0,
        "y": 0,
        "width": 800,
        "height": 180,
        "isLock": false
    })
}

/// 读取合并默认值后的桌面歌词配置对象。
pub fn load_desktop_lyric() -> Map<String, Value> {
    let mut merged = default_desktop_lyric().as_object().cloned().unwrap();
    let cfg = load_config();
    if let Some(saved) = cfg
        .get("settings")
        .and_then(|s| s.get("desktopLyric"))
        .and_then(|v| v.as_object())
    {
        for (k, v) in saved {
            // 数值/字符串为空的字段保留默认值（对齐 Go mergeDesktopLyricConfig）
            let keep_default = match v {
                Value::String(s) => s.is_empty(),
                Value::Number(n) => {
                    matches!(
                        k.as_str(),
                        "fontSize" | "fontWeight" | "width" | "height"
                    ) && n.as_f64().unwrap_or(0.0) <= 0.0
                }
                Value::Null => true,
                _ => false,
            };
            if !keep_default {
                merged.insert(k.clone(), v.clone());
            }
        }
    }
    merged
}

/// 保存桌面歌词配置（写回 settings.desktopLyric）。
pub fn save_desktop_lyric(dl: &Map<String, Value>) {
    let mut cfg = load_config();
    if !cfg.is_object() {
        cfg = json!({});
    }
    let obj = cfg.as_object_mut().unwrap();
    let settings = obj
        .entry("settings".to_string())
        .or_insert_with(|| json!({}));
    if !settings.is_object() {
        *settings = json!({});
    }
    settings
        .as_object_mut()
        .unwrap()
        .insert("desktopLyric".to_string(), Value::Object(dl.clone()));
    let _ = save_config(&cfg);
}

/// 返回桌面歌词配置 JSON 字符串（含默认值合并）。
pub fn desktop_lyric_config_json() -> String {
    Value::Object(load_desktop_lyric()).to_string()
}

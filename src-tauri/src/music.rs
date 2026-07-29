//! 本地音乐：扫描、元数据、歌词、封面（lofty 实现，对齐 Go 后端行为）。

use base64::{engine::general_purpose::STANDARD as B64, Engine as _};
use lofty::picture::PictureType;
use lofty::prelude::*;
use lofty::probe::Probe;
use std::fs;
use std::path::Path;

use crate::models::SongMetadata;

pub const AUDIO_EXTS: &[&str] = &[
    "mp3", "flac", "wav", "aac", "ogg", "m4a", "wma", "opus",
];

/// 递归扫描目录，返回所有音频文件路径（与 Go ScanMusicFolder 一致）。
pub fn scan_folder(root: &str) -> Result<Vec<String>, String> {
    let root_path = Path::new(root);
    if !root_path.is_dir() {
        return Err(format!("不是目录: {root}"));
    }
    let mut files = Vec::new();
    for entry in walkdir::WalkDir::new(root_path)
        .into_iter()
        .filter_map(|e| e.ok())
    {
        let path = entry.path();
        if path.is_file() {
            if let Some(ext) = path.extension().and_then(|s| s.to_str()) {
                if AUDIO_EXTS.contains(&ext.to_ascii_lowercase().as_str()) {
                    files.push(path.to_string_lossy().to_string());
                }
            }
        }
    }
    Ok(files)
}

/// 读取单个音频文件的元数据。
pub fn read_metadata(path: &Path) -> Result<SongMetadata, String> {
    let tagged = Probe::open(path)
        .map_err(|e| e.to_string())?
        .guess_file_type()
        .map_err(|e| e.to_string())?
        .read()
        .map_err(|e| e.to_string())?;

    let props = tagged.properties();
    let duration = props.duration().as_secs_f64();
    let bitrate = props.audio_bitrate().unwrap_or(0);

    let tag = tagged.primary_tag().or_else(|| tagged.first_tag());
    let (title, artist, album, genre, year) = match tag {
        Some(t) => (
            t.title().map(|s| s.to_string()),
            t.artist().map(|s| s.to_string()),
            t.album().map(|s| s.to_string()),
            t.genre().map(|s| s.to_string()),
            t.get_string(&ItemKey::RecordingDate)
                .or_else(|| t.get_string(&ItemKey::Year))
                .map(|s| s.to_string()),
        ),
        None => (None, None, None, None, None),
    };

    Ok(SongMetadata {
        title: title.unwrap_or_default(),
        artist: artist.unwrap_or_default(),
        album: album.unwrap_or_default(),
        genre: genre.unwrap_or_default(),
        year: year.unwrap_or_default(),
        duration,
        bitrate,
    })
}

/// 读取歌词：优先内嵌歌词，其次同目录同名 .lrc 文件。
pub fn read_lyrics(path: &Path) -> String {
    if let Ok(tagged) = Probe::open(path)
        .and_then(|p| p.guess_file_type())
        .and_then(|p| p.read())
    {
        if let Some(tag) = tagged.primary_tag().or_else(|| tagged.first_tag()) {
            if let Some(lyrics) = tag.get_string(&ItemKey::Lyrics) {
                if !lyrics.trim().is_empty() {
                    return lyrics.to_string();
                }
            }
        }
    }
    // 同名 .lrc 兜底
    let lrc = path.with_extension("lrc");
    if let Ok(text) = fs::read_to_string(&lrc) {
        return text;
    }
    String::new()
}

/// 读取封面图，返回 data URL（base64），无封面返回空字符串。
pub fn read_cover_art(path: &Path) -> String {
    let tagged = match Probe::open(path)
        .and_then(|p| p.guess_file_type())
        .and_then(|p| p.read())
    {
        Ok(t) => t,
        Err(_) => return String::new(),
    };
    let tag = match tagged.primary_tag().or_else(|| tagged.first_tag()) {
        Some(t) => t,
        None => return String::new(),
    };
    // 优先正面封面，退回任意图片
    let pic = tag
        .pictures()
        .iter()
        .find(|p| p.pic_type() == PictureType::CoverFront)
        .or_else(|| tag.pictures().first());
    match pic {
        Some(p) => format!(
            "data:{};base64,{}",
            detect_mime(p.data()),
            B64.encode(p.data())
        ),
        None => String::new(),
    }
}

/// 读取图片文件为 data URL。
pub fn read_image_file(path: &Path) -> Result<String, String> {
    let data = fs::read(path).map_err(|e| e.to_string())?;
    Ok(format!(
        "data:{};base64,{}",
        detect_mime(&data),
        B64.encode(&data)
    ))
}

fn detect_mime(data: &[u8]) -> &'static str {
    if data.starts_with(&[0xFF, 0xD8, 0xFF]) {
        "image/jpeg"
    } else if data.starts_with(&[0x89, 0x50, 0x4E, 0x47]) {
        "image/png"
    } else if data.starts_with(b"RIFF") {
        "image/webp"
    } else if data.starts_with(b"BM") {
        "image/bmp"
    } else if data.starts_with(b"GIF8") {
        "image/gif"
    } else {
        "image/jpeg"
    }
}

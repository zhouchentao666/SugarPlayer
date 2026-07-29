//! 数据模型：与原 Go 后端 JSON 字段保持一致。

use serde::{Deserialize, Serialize};

/// 与 Go 版 SongMetadata 字段一致（title/artist/album/genre/year/duration/bitrate）。
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SongMetadata {
    pub title: String,
    pub artist: String,
    pub album: String,
    pub genre: String,
    pub year: String,
    pub duration: f64,
    pub bitrate: u32,
}

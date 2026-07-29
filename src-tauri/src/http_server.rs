//! 本地音频 HTTP 服务器（移植自原 Go AudioServer）。
//! 前端 HTML5 <audio> 通过 http://127.0.0.1:<port>/audio?path=<encoded> 播放本地文件，
//! 捐赠二维码通过 /cover?name=wechat|alipay 提供。

use std::fs;
use std::io::Read;
use std::net::Ipv4Addr;

const WX: &[u8] = include_bytes!("../../assets/微信.jpg");
const ALIPAY: &[u8] = include_bytes!("../../assets/支付宝.jpg");

/// 在随机空闲端口启动本地服务器，返回端口号。
pub fn start() -> u16 {
    let server = tiny_http::Server::http((Ipv4Addr::LOCALHOST, 0)).expect("audio server bind");
    let port = server.server_addr().port();
    std::thread::spawn(move || {
        for req in server.incoming_requests() {
            let resp = handle(req.url());
            let _ = req.respond(resp);
        }
    });
    port
}

fn handle(url: &str) -> tiny_http::Response<Vec<u8>> {
    if let Some((path, query)) = url.split_once('?') {
        match path {
            "/audio" => {
                if let Some(val) = query.strip_prefix("path=") {
                    return serve_file(&percent_decode(val));
                }
                return not_found();
            }
            "/cover" => {
                let name = query.strip_prefix("name=").unwrap_or("");
                let (data, mime) = if name == "wechat" {
                    (WX, "image/jpeg")
                } else if name == "alipay" {
                    (ALIPAY, "image/jpeg")
                } else {
                    (&[][..], "image/jpeg")
                };
                return data_response(data, mime);
            }
            _ => {}
        }
    }
    not_found()
}

fn serve_file(path: &str) -> tiny_http::Response<Vec<u8>> {
    match fs::File::open(path) {
        Ok(mut f) => {
            let mut buf = Vec::new();
            if f.read_to_end(&mut buf).is_ok() {
                return data_response(&buf, mime_for(path));
            }
            error_response()
        }
        Err(_) => error_response(),
    }
}

fn data_response(data: &[u8], mime: &str) -> tiny_http::Response<Vec<u8>> {
    let header =
        tiny_http::Header::from_bytes(&b"Content-Type"[..], mime.as_bytes()).unwrap();
    tiny_http::Response::from_data(data.to_vec()).with_header(header)
}

fn not_found() -> tiny_http::Response<Vec<u8>> {
    tiny_http::Response::from_string("not found").with_status_code(404)
}

fn error_response() -> tiny_http::Response<Vec<u8>> {
    tiny_http::Response::from_string("error").with_status_code(500)
}

fn mime_for(path: &str) -> &'static str {
    match path
        .rsplit('.')
        .next()
        .map(|e| e.to_ascii_lowercase())
        .as_deref()
    {
        Some("mp3") => "audio/mpeg",
        Some("flac") => "audio/flac",
        Some("wav") => "audio/wav",
        Some("ogg") => "audio/ogg",
        Some("m4a") => "audio/mp4",
        Some("aac") => "audio/aac",
        Some("opus") => "audio/ogg",
        Some("wma") => "audio/x-ms-wma",
        Some("ape") => "audio/x-ape",
        _ => "application/octet-stream",
    }
}

fn percent_decode(s: &str) -> String {
    let bytes = s.as_bytes();
    let mut out = Vec::new();
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'%' && i + 2 < bytes.len() {
            let hi = (bytes[i + 1] as char).to_digit(16);
            let lo = (bytes[i + 2] as char).to_digit(16);
            if let (Some(h), Some(l)) = (hi, lo) {
                out.push((h * 16 + l) as u8);
                i += 3;
                continue;
            }
        }
        out.push(bytes[i]);
        i += 1;
    }
    String::from_utf8_lossy(&out).to_string()
}

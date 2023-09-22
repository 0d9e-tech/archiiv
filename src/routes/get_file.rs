use std::{io::ErrorKind, sync::Arc};

use axum::{
    body::StreamBody,
    extract::State,
    http::{
        header::{CONTENT_LENGTH, CONTENT_TYPE},
        StatusCode,
    },
    response::{IntoResponse, Response},
    Extension,
};
use serde::Serialize;
use serde_json::json;
use tokio::fs::{self, File};
use tokio_util::io::ReaderStream;

use crate::{
    global::Global,
    utils::{err_response, sanitize_path, ErrorReason, Json, Username},
};

pub async fn get_file(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    axum::extract::Path(file): axum::extract::Path<String>,
) -> Response {
    let path = match sanitize_path(username, &global, file) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("get_file:{}", path.display());
    let file = match File::open(&path).await {
        Ok(f) => f,
        Err(e) => {
            let (code, reason) =
                if e.kind() == ErrorKind::NotFound || e.kind() == ErrorKind::NotADirectory {
                    (StatusCode::NOT_FOUND, ErrorReason::NotFound404)
                } else {
                    eprintln!("{e}");
                    (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
                };
            return err_response(code, reason).into_response();
        }
    };
    let len = file.metadata().await.unwrap().len();
    let mut resp = StreamBody::new(ReaderStream::new(file)).into_response();
    resp.headers_mut()
        .insert(CONTENT_LENGTH, len.to_string().parse().unwrap());
    if let Some(mime) = mime_guess::from_path(path).first() {
        resp.headers_mut()
            .insert(CONTENT_TYPE, mime.to_string().parse().unwrap());
    }
    dbg!(resp)
}

pub async fn get_meta(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    axum::extract::Path(file): axum::extract::Path<String>,
) -> Response {
    let path = match sanitize_path(username, &global, file) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("get_meta:{}", path.display());
    let meta = match fs::metadata(&path).await {
        Ok(f) => f,
        Err(e) => {
            let (code, reason) =
                if e.kind() == ErrorKind::NotFound || e.kind() == ErrorKind::NotADirectory {
                    (StatusCode::NOT_FOUND, ErrorReason::NotFound404)
                } else {
                    eprintln!("{e}");
                    (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
                };
            return err_response(code, reason).into_response();
        }
    };
    if meta.is_file() {
        let mut result = serde_json::to_value(MetaItem::File {
            name: path.file_name().unwrap().to_string_lossy().into_owned(),
            size: meta.len(),
        })
        .unwrap();
        result["ok"] = serde_json::Value::Bool(true);
        Json(result).into_response()
    } else if meta.is_dir() {
        let mut results = Vec::new();
        let mut dir = fs::read_dir(path).await.unwrap();
        loop {
            let Ok(x) = dir.next_entry().await else {
                continue;
            };
            let Some(x) = x else {
                break;
            };
            let name = x.file_name().to_string_lossy().into_owned();
            let meta = match x.metadata().await {
                Ok(meta) => meta,
                Err(e) => {
                    eprintln!("Error reading metadata for {name}: {e}");
                    continue;
                }
            };
            results.push(if meta.is_dir() {
                MetaItem::Dir { name }
            } else {
                MetaItem::File {
                    name,
                    size: meta.len(),
                }
            });
        }
        Json(json!({
            "ok": true,
            "type": "directory",
            "contents": results,
        }))
        .into_response()
    } else {
        err_response(StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500).into_response()
    }
}

#[derive(Serialize)]
#[serde(tag = "type")]
enum MetaItem {
    #[serde(rename = "file")]
    File { name: String, size: u64 },
    #[serde(rename = "directory")]
    Dir { name: String },
}

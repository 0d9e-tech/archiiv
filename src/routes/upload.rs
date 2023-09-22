use std::{
    io::{self, ErrorKind},
    sync::Arc,
};

use axum::{
    extract::{BodyStream, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Extension,
};
use futures_util::TryStreamExt;
use serde::Deserialize;
use serde_json::json;
use tokio::fs::{self, File};
use tokio_util::io::StreamReader;

use crate::{
    global::Global,
    utils::{err_response, sanitize_path, ErrorReason, Json, Username},
};

pub async fn upload(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    axum::extract::Path(file): axum::extract::Path<String>,
    stream: BodyStream,
) -> Response {
    let path = match sanitize_path(&username, &global, file) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("upload:{}", path.display());
    let mut file = match File::options()
        .create_new(true)
        .write(true)
        .open(path)
        .await
    {
        Ok(f) => f,
        Err(e) => {
            let (code, reason) = match e.kind() {
                ErrorKind::NotFound => (StatusCode::BAD_REQUEST, ErrorReason::ParentDoesntExist),
                ErrorKind::AlreadyExists | ErrorKind::NotADirectory => {
                    (StatusCode::BAD_REQUEST, ErrorReason::PathExists)
                }
                _ => {
                    eprintln!("{e}");
                    (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
                }
            };
            return err_response(code, reason).into_response();
        }
    };
    if let Err(e) = tokio::io::copy(
        &mut StreamReader::new(stream.map_err(|e| std::io::Error::new(ErrorKind::Other, e))),
        &mut file,
    )
    .await
    {
        eprintln!("{e}");
        return err_response(StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
            .into_response();
    }
    Json(json!({
        "ok": true
    }))
    .into_response()
}

pub async fn delete(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    axum::extract::Path(file): axum::extract::Path<String>,
) -> Response {
    let path = match sanitize_path(&username, &global, file) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("delete:{}", path.display());
    // FIXME: race condition
    let meta = match fs::metadata(&path).await {
        Ok(f) => f,
        Err(e) => {
            let (code, reason) = match e.kind() {
                ErrorKind::NotFound | ErrorKind::NotADirectory => {
                    (StatusCode::NOT_FOUND, ErrorReason::NotFound404)
                }
                _ => {
                    eprintln!("{e}");
                    (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
                }
            };
            return err_response(code, reason).into_response();
        }
    };
    let result = if meta.is_file() {
        fs::remove_file(path).await
    } else if meta.is_dir() {
        fs::remove_dir_all(path).await
    } else {
        Err(io::Error::new(
            ErrorKind::Other,
            "Neither a file nor a directory",
        ))
    };
    match result {
        Ok(()) => Json(json!({ "ok": true })).into_response(),
        Err(e) => {
            let (code, reason) = match e.kind() {
                ErrorKind::NotFound => (StatusCode::BAD_REQUEST, ErrorReason::ParentDoesntExist),
                ErrorKind::AlreadyExists | ErrorKind::NotADirectory => {
                    (StatusCode::BAD_REQUEST, ErrorReason::PathExists)
                }
                _ => {
                    eprintln!("{e}");
                    (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
                }
            };
            err_response(code, reason).into_response()
        }
    }
}

pub async fn mkdir(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    Json(query): Json<MkDirRequest>,
) -> Response {
    let path = match sanitize_path(&username, &global, query.path) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("mkdir:{}", path.display());
    if let Err(e) = tokio::fs::create_dir(&path).await {
        let (code, reason) = match e.kind() {
            ErrorKind::NotFound => (StatusCode::BAD_REQUEST, ErrorReason::ParentDoesntExist),
            ErrorKind::AlreadyExists | ErrorKind::NotADirectory => {
                (StatusCode::BAD_REQUEST, ErrorReason::PathExists)
            }
            _ => {
                eprintln!("{e}");
                (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
            }
        };
        return err_response(code, reason).into_response();
    }
    Json(json!({
        "ok": true
    }))
    .into_response()
}

#[derive(Deserialize)]
pub struct MkDirRequest {
    path: String,
}

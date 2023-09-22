use std::{io::ErrorKind, sync::Arc};

use axum::{
    extract::{BodyStream, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Extension,
};
use futures_util::TryStreamExt;
use serde_json::json;
use tokio::fs::File;
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
    let path = match sanitize_path(username, &global, file) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("Uploading to {}", path.display());
    let mut file = match File::options()
        .create_new(true)
        .write(true)
        .open(path)
        .await
    {
        Ok(f) => f,
        Err(e) => {
            let (code, reason) = match e.kind() {
                ErrorKind::NotFound => (StatusCode::BAD_REQUEST, ErrorReason::DirectoryDoesntExist),
                ErrorKind::AlreadyExists => (StatusCode::BAD_REQUEST, ErrorReason::FileExists),
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

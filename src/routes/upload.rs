use std::{io::ErrorKind, path::Path, sync::Arc};

use axum::{
    extract::{BodyStream, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Extension,
};
use futures_util::TryStreamExt;
use serde::Deserialize;
use serde_json::json;
use tokio::fs::File;
use tokio_util::io::StreamReader;

use crate::{
    global::Global,
    utils::{err_response, err_response_with_info, ErrorReason, Json, Query, Username},
};

#[axum::debug_handler]
pub async fn upload(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    Query(UploadQuery { file }): Query<UploadQuery>,
    stream: BodyStream,
) -> Response {
    let Ok(sanitized_filename) = Path::new(&file).strip_prefix("/") else {
        return err_response_with_info(
            StatusCode::BAD_REQUEST,
            ErrorReason::InvalidPath,
            "Path must start with /",
        )
        .into_response();
    };
    if !sanitized_filename
        .components()
        .all(|c| matches!(c, std::path::Component::Normal(_)))
    {
        return err_response_with_info(
            StatusCode::BAD_REQUEST,
            ErrorReason::InvalidPath,
            "Path can't contain `..` or `/./`",
        )
        .into_response();
    }
    let mut path = global.data_dir.to_path_buf();
    path.push("users");
    path.push(username);
    path = path.join(sanitized_filename);

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

#[derive(Deserialize)]
pub struct UploadQuery {
    file: String,
}

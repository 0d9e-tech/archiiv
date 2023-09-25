use std::{io::ErrorKind, sync::Arc};

use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    Extension,
};
use serde::Deserialize;
use serde_json::json;
use tokio::fs::{self};

use crate::{
    global::Global,
    utils::{err_response, sanitize_path, ErrorReason, Json, Username},
};

/// This currently follows linux semantics, and doesn't detect fine-grained error info.
///
/// TODO: I'd prefer to change this, but that will likely require fs locking.
pub async fn rename(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    Json(request): Json<RenameRequest>,
) -> Response {
    let from = match sanitize_path(&username, &global, request.from) {
        Ok(x) => x,
        Err(e) => return e,
    };
    let to = match sanitize_path(&username, &global, request.to) {
        Ok(x) => x,
        Err(e) => return e,
    };
    eprintln!("rename\n  from:{}\n  to:{}", from.display(), to.display());
    match fs::rename(from, to).await {
        Ok(()) => Json(json!({ "ok": true })).into_response(),
        Err(e) => {
            eprintln!("{e}");
            let (code, reason) = match e.kind() {
                ErrorKind::PermissionDenied | ErrorKind::Other => {
                    (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
                }
                _ => (StatusCode::BAD_REQUEST, ErrorReason::InvalidRenameOperation),
            };
            err_response(code, reason).into_response()
        }
    }
}

#[derive(Deserialize)]
pub struct RenameRequest {
    from: String,
    to: String,
}

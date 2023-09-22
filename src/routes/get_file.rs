use std::{io::ErrorKind, sync::Arc};

use axum::{
    body::StreamBody,
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    Extension,
};
use tokio::fs::File;
use tokio_util::io::ReaderStream;

use crate::{
    global::Global,
    utils::{err_response, sanitize_path, ErrorReason, Username},
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
    let file = match File::open(&path).await {
        Ok(f) => f,
        Err(e) => {
            let (code, reason) = if e.kind() == ErrorKind::NotFound {
                (StatusCode::NOT_FOUND, ErrorReason::NotFound404)
            } else {
                eprintln!("{e}");
                (StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
            };
            return err_response(code, reason).into_response();
        }
    };
    let mut resp = StreamBody::new(ReaderStream::new(file)).into_response();
    if let Some(mime) = mime_guess::from_path(path).first() {
        resp.headers_mut()
            .insert("Content-Type", mime.to_string().parse().unwrap());
    }
    resp
}

use std::{collections::hash_map::Entry, sync::Arc};

use axum::{
    extract::{DefaultBodyLimit, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing as axr, Extension,
};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::{
    consts::UPLOAD_LIMIT_BYTES,
    global::Global,
    routes::{
        get_file::{get_file, get_meta},
        upload::{mkdir, upload},
    },
    utils::{err_response, handle_method_not_allowed, ErrorReason, Json, Username},
};

pub fn create_app() -> axr::Router<Arc<Global>> {
    axr::Router::new()
        .route(
            "/whoami",
            axr::get(whoami).fallback(handle_method_not_allowed),
        )
        .route(
            "/delete-auth-token",
            axr::post(del_token).fallback(handle_method_not_allowed),
        )
        .route(
            "/f/*path",
            axr::get(get_file)
                .post(upload)
                .layer(DefaultBodyLimit::max(UPLOAD_LIMIT_BYTES))
                .fallback(handle_method_not_allowed),
        )
        .route(
            "/mkdir",
            axr::post(mkdir).fallback(handle_method_not_allowed),
        )
        .route(
            "/meta/*path",
            axr::get(get_meta).fallback(handle_method_not_allowed),
        )
}

async fn whoami(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
) -> Response {
    let tokens = global.get_users().await[&username]
        .session_tokens
        .iter()
        .map(|(device, data)| TokenInfo {
            device: device.clone(),
            created_on: data.created_on,
        })
        .collect::<Vec<_>>();

    Json(serde_json::json!({
        "ok": true,
        "username": username,
        "tokens": tokens,
    }))
    .into_response()
}

#[derive(Serialize)]
struct TokenInfo {
    device: String,
    created_on: DateTime<Utc>,
}

enum DelTokenResult {
    Ok,
    DeviceNameNotFound,
    CannotDeleteLastToken,
}

async fn del_token(
    State(global): State<Arc<Global>>,
    Extension(Username(username)): Extension<Username>,
    Json(req): Json<DelTokenRequest>,
) -> Response {
    let result = global
        .modify_users(|us| {
            if let Some(u) = us.get_mut(&username) {
                let is_last = u.session_tokens.len() == 1;
                if let Entry::Occupied(e) = u.session_tokens.entry(req.device_name) {
                    if is_last {
                        return DelTokenResult::CannotDeleteLastToken;
                    }
                    e.remove();
                    return DelTokenResult::Ok;
                }
            }
            DelTokenResult::DeviceNameNotFound
        })
        .await;
    match result {
        DelTokenResult::Ok => Json(serde_json::json!({
            "ok": true,
        }))
        .into_response(),

        DelTokenResult::DeviceNameNotFound => {
            err_response(StatusCode::BAD_REQUEST, ErrorReason::DeviceNameNotFound).into_response()
        }

        DelTokenResult::CannotDeleteLastToken => err_response(
            StatusCode::BAD_REQUEST,
            ErrorReason::CannotDeleteLastAuthToken,
        )
        .into_response(),
    }
}

#[derive(Deserialize)]
struct DelTokenRequest {
    device_name: String,
}

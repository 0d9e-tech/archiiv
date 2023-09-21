use std::sync::Arc;

use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing as axr, Extension, Json,
};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::{
    global::Global,
    utils::{err_response, ErrorReason, Username},
};

pub fn create_app() -> axr::Router<Arc<Global>> {
    axr::Router::new()
        .route("/whoami", axr::get(whoami))
        .route("/delete-auth-token", axr::post(del_token))
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
                if u.session_tokens.len() == 1 {
                    return DelTokenResult::CannotDeleteLastToken;
                }
                if u.session_tokens.remove(&req.device_name).is_some() {
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

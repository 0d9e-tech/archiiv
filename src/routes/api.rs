use std::sync::Arc;

use axum::{
    extract::State,
    response::{IntoResponse, Response},
    routing as axr, Extension, Json,
};
use chrono::{DateTime, Utc};
use serde::Serialize;

use crate::{global::Global, utils::Username};

pub fn create_app() -> axr::Router<Arc<Global>> {
    axr::Router::new().route("/whoami", axr::get(whoami))
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

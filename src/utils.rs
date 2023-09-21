use std::sync::Arc;

use axum::{
    extract::State,
    http::{header::AUTHORIZATION, Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
    Json,
};

use crate::global::Global;

pub fn err_response(code: StatusCode, reason: ErrorReason) -> impl IntoResponse {
    (
        code,
        Json(serde_json::json!({
            "ok": false,
            "reason": reason.s(),
        })),
    )
}

#[derive(Clone, Copy)]
pub enum ErrorReason {
    RegistrationDisabled,
    UsernameExists,
    InvalidUsername,
    IncorrectLoginOrOtp,
    MissingAuthToken,
    InvalidAuthToken,
    DeviceNameExists,
    DeviceNameNotFound,
    CannotDeleteLastAuthToken,
}

impl ErrorReason {
    pub const fn s(self) -> &'static str {
        match self {
            Self::RegistrationDisabled => "error.registration_disabled",
            Self::UsernameExists => "error.username_exists",
            Self::InvalidUsername => "error.invalid_username",
            Self::IncorrectLoginOrOtp => "error.incorrect_login_or_otp",
            Self::MissingAuthToken => "error.missing_auth_token",
            Self::InvalidAuthToken => "error.invalid_auth_token",
            Self::DeviceNameExists => "error.device_name_exists",
            Self::DeviceNameNotFound => "error.device_name_not_found",
            Self::CannotDeleteLastAuthToken => "error.cannot_delete_last_auth_token",
        }
    }
}

pub async fn auth_middleware<B: Send>(
    State(state): State<Arc<Global>>,
    mut request: Request<B>,
    next: Next<B>,
) -> Response {
    let Some(auth) = request.headers().get(AUTHORIZATION) else {
        return err_response(StatusCode::UNAUTHORIZED, ErrorReason::MissingAuthToken)
            .into_response();
    };
    let parts = auth.to_str().unwrap().splitn(2, ' ').collect::<Vec<_>>();
    if parts.len() != 2 || parts[0] != "Bearer" {
        return err_response(StatusCode::UNAUTHORIZED, ErrorReason::InvalidAuthToken)
            .into_response();
    }
    let token = parts[1];
    let user = state
        .get_users()
        .await
        .iter()
        .find_map(|(username, u)| {
            u.session_tokens
                .iter()
                .find(|(_, t)| t.token == token)
                .map(|_| username)
        })
        .cloned();
    let Some(username) = user else {
        return err_response(StatusCode::UNAUTHORIZED, ErrorReason::InvalidAuthToken)
            .into_response();
    };
    request.extensions_mut().insert(Username(username));
    next.run(request).await
}

#[derive(Clone)]
pub struct Username(pub String);

use std::{
    any::{Any, TypeId},
    sync::Arc,
};

use axum::{
    extract::{rejection::JsonRejection, FromRequest, State},
    http::{header::AUTHORIZATION, Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
};
use serde::Serialize;
use serde_json::Value as JsonValue;

use crate::global::Global;

pub fn err_response(code: StatusCode, reason: ErrorReason) -> (StatusCode, Json<JsonValue>) {
    err_response_with_info::<()>(code, reason, ())
}

pub fn err_response_with_info<T: Serialize + Any>(
    code: StatusCode,
    reason: ErrorReason,
    extra_info: T,
) -> (StatusCode, Json<JsonValue>) {
    let mut value = serde_json::json!({
        "ok": false,
        "reason": reason.s(),
    });
    if extra_info.type_id() != TypeId::of::<()>() {
        value
            .as_object_mut()
            .unwrap()
            .insert("info".to_owned(), serde_json::to_value(extra_info).unwrap());
    }
    (code, Json(value))
}

#[derive(Clone, Copy)]
pub enum ErrorReason {
    NotFound404,
    MethodNotAllowed405,
    Error500,
    NetworkError,
    MalformedRequest,

    // Registration
    RegistrationDisabled,
    UsernameExists,
    InvalidUsername,

    // Login
    IncorrectLoginOrOtp,
    DeviceNameExists,

    // Auth middleware
    MissingAuthToken,
    InvalidAuthToken,

    // Delete token
    DeviceNameNotFound,
    CannotDeleteLastAuthToken,
}

impl ErrorReason {
    pub const fn s(self) -> &'static str {
        match self {
            Self::NotFound404 => "error.404_not_found",
            Self::MethodNotAllowed405 => "error.405_method_not_allowed",
            Self::Error500 => "error.500_internal_server_error",
            Self::NetworkError => "error.network_error",
            Self::MalformedRequest => "error.malformed_request",
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

pub async fn handle_404() -> impl IntoResponse {
    err_response(StatusCode::NOT_FOUND, ErrorReason::NotFound404)
}

pub async fn handle_method_not_allowed() -> impl IntoResponse {
    err_response(
        StatusCode::METHOD_NOT_ALLOWED,
        ErrorReason::MethodNotAllowed405,
    )
}

/// Custom Json request extractor + response creator which returns all extraction errors as JSON response
///
/// Adapted from [axum example]
///
/// [axum example]: https://github.com/tokio-rs/axum/blob/c8cf147657093bff3aad5cbf2dafa336235a37c6/examples/customize-extractor-error/src/custom_extractor.rs
pub struct Json<T>(pub T);

#[axum::async_trait]
impl<T, S, B> FromRequest<S, B> for Json<T>
where
    axum::Json<T>: FromRequest<S, B, Rejection = JsonRejection>,
    S: Send + Sync,
    B: Send + 'static,
{
    type Rejection = (StatusCode, Json<JsonValue>);

    async fn from_request(req: Request<B>, state: &S) -> Result<Self, Self::Rejection> {
        let res = axum::Json::<T>::from_request(req, state).await;
        match res {
            Ok(x) => Ok(Self(x.0)),
            Err(JsonRejection::BytesRejection(_)) => Err(err_response(
                StatusCode::BAD_REQUEST,
                ErrorReason::NetworkError,
            )),
            Err(JsonRejection::JsonDataError(e)) => Err(err_response_with_info(
                StatusCode::BAD_REQUEST,
                ErrorReason::MalformedRequest,
                e.body_text(),
            )),
            Err(JsonRejection::JsonSyntaxError(e)) => Err(err_response_with_info(
                StatusCode::BAD_REQUEST,
                ErrorReason::MalformedRequest,
                e.body_text(),
            )),
            Err(JsonRejection::MissingJsonContentType(e)) => Err(err_response_with_info(
                StatusCode::BAD_REQUEST,
                ErrorReason::MalformedRequest,
                e.body_text(),
            )),
            Err(_) => Err(err_response(
                StatusCode::INTERNAL_SERVER_ERROR,
                ErrorReason::Error500,
            )),
        }
    }
}

impl<T> IntoResponse for Json<T>
where
    axum::Json<T>: IntoResponse,
{
    fn into_response(self) -> Response {
        axum::Json(self.0).into_response()
    }
}

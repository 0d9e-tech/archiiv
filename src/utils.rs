use std::{
    any::Any,
    ffi::OsStr,
    path::{Path, PathBuf},
    sync::Arc,
};

use axum::{
    body::BoxBody,
    extract::{
        rejection::{JsonRejection, QueryRejection},
        FromRequest, FromRequestParts, State,
    },
    http::{header::AUTHORIZATION, Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
};
use serde::Serialize;
use serde_json::Value as JsonValue;
use tower_http::catch_panic::ResponseForPanic;

use crate::global::Global;

pub fn err_response(code: StatusCode, reason: ErrorReason) -> (StatusCode, Json<JsonValue>) {
    err_response_with_info::<()>(code, reason, ())
}

pub fn err_response_with_info<T: Serialize + Any>(
    code: StatusCode,
    reason: ErrorReason,
    extra_info: T,
) -> (StatusCode, Json<JsonValue>) {
    let value = serde_json::json!({
        "ok": false,
        "reason": reason.s(),
        "info": extra_info,
    });
    (code, Json(value))
}

#[derive(Clone, Copy)]
pub enum ErrorReason {
    NotFound404,
    MethodNotAllowed405,
    Error500,
    NetworkError,
    MalformedRequest,
    InvalidPath,

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

    // Upload/mkdir
    ParentDoesntExist,
    PathExists,

    // Rename
    /// See docs for [`crate::routes::rename::rename`]
    InvalidRenameOperation,
}

impl ErrorReason {
    pub const fn s(self) -> &'static str {
        match self {
            Self::NotFound404 => "error.generic.404_not_found",
            Self::MethodNotAllowed405 => "error.generic.405_method_not_allowed",
            Self::Error500 => "error.generic.500_internal_server_error",
            Self::NetworkError => "error.generic.network_error",
            Self::MalformedRequest => "error.generic.malformed_request",
            Self::InvalidPath => "error.generic.invalid_path",

            Self::RegistrationDisabled => "error.registration.registration_disabled",
            Self::UsernameExists => "error.registration.username_exists",
            Self::InvalidUsername => "error.registration.invalid_username",

            Self::IncorrectLoginOrOtp => "error.login.incorrect_login_or_otp",
            Self::DeviceNameExists => "error.login.device_name_exists",

            Self::MissingAuthToken => "error.auth.missing_auth_token",
            Self::InvalidAuthToken => "error.auth.invalid_auth_token",

            Self::DeviceNameNotFound => "error.delete_token.device_name_not_found",
            Self::CannotDeleteLastAuthToken => "error.delete_token.cannot_delete_last_auth_token",

            Self::ParentDoesntExist => "error.upload.parent_doesnt_exist",
            Self::PathExists => "error.upload.path_exists",

            Self::InvalidRenameOperation => "error.rename.invalid_rename_operation",
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

#[derive(Clone, Copy)]
pub struct PanicHandler;

impl ResponseForPanic for PanicHandler {
    type ResponseBody = BoxBody;

    fn response_for_panic(
        &mut self,
        err: Box<dyn Any + Send + 'static>,
    ) -> axum::http::Response<Self::ResponseBody> {
        let details = err.downcast::<String>().map_or_else(
            |err| {
                err.downcast::<&str>()
                    .map_or_else(|_| "Unknown panic message", |s| *s)
                    .to_owned()
            },
            |err| *err,
        );
        eprintln!("Panic in HTTP handler: {details}");
        let (parts, body) = err_response(StatusCode::INTERNAL_SERVER_ERROR, ErrorReason::Error500)
            .into_response()
            .into_parts();
        axum::http::Response::from_parts(parts, body)
    }
}

pub fn sanitize_path(
    username: &str,
    global: &Global,
    path: impl AsRef<OsStr>,
) -> Result<PathBuf, Response> {
    let sanitized_filename = Path::new(&path);

    if !sanitized_filename
        .components()
        .all(|c| matches!(c, std::path::Component::Normal(_)))
    {
        return Err(err_response_with_info(
            StatusCode::BAD_REQUEST,
            ErrorReason::InvalidPath,
            "Path can't contain `..` or `/./`, or start with `/`",
        )
        .into_response());
    }
    let mut path = global.data_dir.to_path_buf();
    path.push("users");
    path.push(username);
    Ok(path.join(sanitized_filename))
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
            Err(e) => {
                eprintln!("{e}");
                Err(err_response(
                    StatusCode::INTERNAL_SERVER_ERROR,
                    ErrorReason::Error500,
                ))
            }
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

/// Custom Query request extractor
///
/// Adapted in similar fashion as Json above
pub struct Query<T>(pub T);

#[axum::async_trait]
impl<T, S> FromRequestParts<S> for Query<T>
where
    axum::extract::Query<T>: FromRequestParts<S, Rejection = QueryRejection>,
    S: Send + Sync,
{
    type Rejection = (StatusCode, Json<JsonValue>);

    async fn from_request_parts(
        parts: &mut axum::http::request::Parts,
        state: &S,
    ) -> Result<Self, Self::Rejection> {
        let res = axum::extract::Query::<T>::from_request_parts(parts, state).await;
        match res {
            Ok(x) => Ok(Self(x.0)),
            Err(QueryRejection::FailedToDeserializeQueryString(e)) => Err(err_response_with_info(
                StatusCode::BAD_REQUEST,
                ErrorReason::MalformedRequest,
                e.body_text(),
            )),
            Err(e) => {
                eprintln!("{e}");
                Err(err_response(
                    StatusCode::INTERNAL_SERVER_ERROR,
                    ErrorReason::Error500,
                ))
            }
        }
    }
}

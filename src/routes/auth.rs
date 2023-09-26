use std::{
    collections::{hash_map::Entry, HashMap},
    sync::Arc,
};

use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing as axr,
};
use chrono::Utc;
use rand::{thread_rng, Rng};
use serde::Deserialize;
use totp_rs::Secret;

use crate::{
    global::{Global, SessionToken, User},
    utils::{err_response, handle_method_not_allowed, ErrorReason, Json},
};

pub fn create_app() -> axr::Router<Arc<Global>> {
    axr::Router::new()
        .route(
            "/register",
            axr::post(register).fallback(handle_method_not_allowed),
        )
        .route(
            "/login",
            axr::post(login).fallback(handle_method_not_allowed),
        )
}

async fn register(State(global): State<Arc<Global>>, Json(req): Json<RegisterRequest>) -> Response {
    if !req.username.chars().all(|c| c.is_ascii_alphanumeric())
        || req.username.len() > 64
        || req.username.len() < 2
    {
        return err_response(StatusCode::BAD_REQUEST, ErrorReason::InvalidUsername).into_response();
    }
    if !global.config.enable_registration && global.get_users().await.len() > 0 {
        return err_response(StatusCode::FORBIDDEN, ErrorReason::RegistrationDisabled)
            .into_response();
    }
    let user = global
        .modify_users(|us| {
            if let Entry::Vacant(e) = us.entry(req.username.clone()) {
                let u = User {
                    registered_on: Utc::now(),
                    otp: totp_rs::TOTP::new(
                        totp_rs::Algorithm::SHA1,
                        6,
                        1,
                        30,
                        Secret::generate_secret().to_bytes().unwrap(),
                        Some("Archív".to_owned()),
                        req.username,
                    )
                    .unwrap(),
                    session_tokens: HashMap::new(),
                };
                e.insert(u.clone());
                return Some(u);
            }
            None
        })
        .await;
    match user {
        None => err_response(StatusCode::BAD_REQUEST, ErrorReason::UsernameExists).into_response(),
        Some(u) => Json(serde_json::json!({
            "ok": true,
            "otp": u.otp.get_url(),
        }))
        .into_response(),
    }
}

#[derive(Deserialize)]
struct RegisterRequest {
    username: String,
}

enum LoginResult {
    IncorrectLoginOrOtp,
    DeviceNameExists,
    Ok(String),
}

async fn login(State(global): State<Arc<Global>>, Json(req): Json<LoginRequest>) -> Response {
    let result = global
        .modify_users(|us| {
            if let Some(u) = us.get_mut(&req.username) {
                if u.otp.check_current(&req.otp).unwrap() {
                    let s = thread_rng()
                        .sample_iter(rand::distributions::Alphanumeric)
                        .take(32)
                        .map(|c| c as char)
                        .collect::<String>();
                    return match u.session_tokens.entry(req.device_name) {
                        Entry::Occupied(_) => LoginResult::DeviceNameExists,
                        Entry::Vacant(e) => {
                            e.insert(SessionToken {
                                token: s.clone(),
                                created_on: Utc::now(),
                            });
                            LoginResult::Ok(s)
                        }
                    };
                }
            }
            LoginResult::IncorrectLoginOrOtp
        })
        .await;
    match result {
        LoginResult::IncorrectLoginOrOtp => {
            err_response(StatusCode::UNAUTHORIZED, ErrorReason::IncorrectLoginOrOtp).into_response()
        }
        LoginResult::DeviceNameExists => {
            err_response(StatusCode::BAD_REQUEST, ErrorReason::DeviceNameExists).into_response()
        }
        LoginResult::Ok(t) => {
            tokio::fs::create_dir_all(global.data_dir.join(format!("users/{}", req.username)))
                .await
                .unwrap();
            Json(serde_json::json!({
                "ok": true,
                "token": t,
            }))
            .into_response()
        }
    }
}

#[derive(Deserialize)]
struct LoginRequest {
    username: String,
    otp: String,
    device_name: String,
}

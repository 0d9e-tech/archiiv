use std::{
    collections::{hash_map::Entry, HashMap},
    sync::Arc,
};

use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing as axr, Json,
};
use chrono::Utc;
use rand::{thread_rng, Rng};
use totp_rs::Secret;

use crate::{
    global::{Global, SessionToken, User},
    utils::{err_response, ErrorReason},
};

pub fn create_app() -> axr::Router<Arc<Global>> {
    axr::Router::new()
        .route("/register", axr::post(register))
        .route("/login", axr::post(login))
}

async fn register(State(global): State<Arc<Global>>, Json(req): Json<RegisterRequest>) -> Response {
    if !req.username.chars().all(|c| c.is_ascii_alphanumeric()) {
        return err_response(StatusCode::BAD_REQUEST, ErrorReason::InvalidUsername).into_response();
    }
    if !global.config.enable_registration && global.get_users().await.len() > 0 {
        return err_response(StatusCode::FORBIDDEN, ErrorReason::RegistrationDisabled)
            .into_response();
    }
    let mut user = None;
    global
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
                user = Some(u);
            }
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

#[derive(serde::Deserialize)]
struct RegisterRequest {
    username: String,
}

enum LoginResult {
    IncorrectLoginOrOtp,
    DeviceNameExists,
    Ok(String),
}

async fn login(State(global): State<Arc<Global>>, Json(req): Json<LoginRequest>) -> Response {
    let mut token = LoginResult::IncorrectLoginOrOtp;
    global
        .modify_users(|us| {
            if let Some(u) = us.get_mut(&req.username) {
                if u.otp.check_current(&req.otp).unwrap() {
                    let s = thread_rng()
                        .sample_iter(rand::distributions::Alphanumeric)
                        .take(32)
                        .map(|c| c as char)
                        .collect::<String>();
                    match u.session_tokens.entry(req.device_name) {
                        Entry::Occupied(_) => {
                            token = LoginResult::DeviceNameExists;
                            return;
                        }
                        Entry::Vacant(e) => {
                            e.insert(SessionToken {
                                token: s.clone(),
                                created_on: Utc::now(),
                            });
                        }
                    }
                    token = LoginResult::Ok(s);
                }
            }
        })
        .await;
    match token {
        LoginResult::IncorrectLoginOrOtp => {
            err_response(StatusCode::UNAUTHORIZED, ErrorReason::IncorrectLoginOrOtp).into_response()
        }
        LoginResult::DeviceNameExists => {
            err_response(StatusCode::BAD_REQUEST, ErrorReason::DeviceNameExists).into_response()
        }
        LoginResult::Ok(t) => Json(serde_json::json!({
            "ok": true,
            "token": t,
        }))
        .into_response(),
    }
}

#[derive(serde::Deserialize)]
struct LoginRequest {
    username: String,
    otp: String,
    device_name: String,
}

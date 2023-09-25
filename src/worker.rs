use std::sync::Arc;

use chrono::{Duration, Utc};

use crate::{
    consts::{ACCOUNT_WITHOUT_TOKEN_EXPIRATION, WORKER_INTERVAL},
    global::Global,
};

pub async fn worker(global: Arc<Global>) {
    loop {
        tokio::time::sleep(WORKER_INTERVAL).await;
        global
            .modify_users(|us| {
                us.retain(|_, u| {
                    !u.session_tokens.is_empty()
                        || Utc::now() - u.registered_on
                            < Duration::from_std(ACCOUNT_WITHOUT_TOKEN_EXPIRATION).unwrap()
                });
            })
            .await;
    }
}

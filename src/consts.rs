use std::time::Duration;

pub const WORKER_INTERVAL: Duration = Duration::from_secs(60 * 5);
pub const ACCOUNT_WITHOUT_TOKEN_EXPIRATION: Duration = Duration::from_secs(60 * 10);
pub const UPLOAD_LIMIT_BYTES: usize = 1024 * 1024 * 1024 * 10;

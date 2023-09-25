use std::{collections::HashMap, io, path::Path};

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use tokio::{
    fs,
    sync::{RwLock, RwLockReadGuard},
};

#[derive(Serialize, Deserialize, Clone)]
pub struct User {
    pub registered_on: DateTime<Utc>,
    pub otp: totp_rs::TOTP,
    pub session_tokens: HashMap<String, SessionToken>,
}

#[derive(Serialize, Deserialize, Clone)]
pub struct SessionToken {
    pub token: String,
    pub created_on: DateTime<Utc>,
}

#[derive(Serialize, Deserialize)]
pub struct Config {
    pub port: u16,
    pub enable_registration: bool,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            port: 4000,
            enable_registration: false,
        }
    }
}

pub struct Global {
    users: RwLock<HashMap<String, User>>,
    pub config: Config,
    pub data_dir: Box<Path>,
}

impl Global {
    pub fn load(data_dir: Box<Path>) -> Self {
        let users = RwLock::new(load_or_default(&data_dir.join("users.json")));
        let config = load_or_default(&data_dir.join("config.json"));
        Self {
            users,
            config,
            data_dir,
        }
    }

    pub async fn get_users(&self) -> RwLockReadGuard<'_, HashMap<String, User>> {
        self.users.read().await
    }

    pub async fn modify_users<T: Send>(
        &self,
        f: impl FnOnce(&mut HashMap<String, User>) -> T + Send,
    ) -> T {
        let mut guard = self.users.write().await;
        let res = f(&mut guard);
        let guard = guard.downgrade();
        let path = self.data_dir.join("users.json");
        if let Err(e) = fs::write(&path, serde_json::to_vec_pretty(&*guard).unwrap()).await {
            eprintln!("Failed to write {}: {}", path.display(), e);
        }
        res
    }
}

fn load_or_default<'a, T>(path: &Path) -> T
where
    T: Serialize + Default + for<'de> serde::Deserialize<'de>,
{
    match std::fs::File::open(path) {
        Ok(file) => serde_json::from_reader(file)
            .map_err(|e| format!("Failed to parse {}: {}", path.display(), e))
            .unwrap(),
        Err(e) if e.kind() == io::ErrorKind::NotFound => {
            eprintln!("Creating default {}", path.display());
            let default = T::default();
            serde_json::to_writer_pretty(
                std::fs::File::create(path)
                    .map_err(|e| format!("Failed to create {}: {}", path.display(), e))
                    .unwrap(),
                &default,
            )
            .unwrap();
            default
        }
        Err(e) => panic!("Failed to open {}: {}", path.display(), e),
    }
}

#![feature(async_fn_in_trait)]

mod consts;
mod global;
mod routes;
mod utils;
mod worker;

use std::{
    env,
    ffi::OsString,
    fs,
    io::ErrorKind,
    net::{Ipv4Addr, SocketAddr, SocketAddrV4},
    path::PathBuf,
    sync::Arc,
};

use axum::routing as axr;
use tokio::{
    select,
    signal::unix::{signal, SignalKind},
};
use tower_http::cors::{AllowHeaders, CorsLayer};
use utils::handle_method_not_allowed;

use crate::{global::Global, utils::handle_404};

#[tokio::main]
async fn main() {
    let data_dir: PathBuf = env::var("DATA_DIR")
        .unwrap_or_else(|_| "data".to_owned())
        .try_into()
        .expect("DATA_DIR is not a valid path");
    let global = Arc::new(Global::load(data_dir.into_boxed_path()));
    ensure_dirs(&global).await;
    let addr = SocketAddr::V4(SocketAddrV4::new(Ipv4Addr::UNSPECIFIED, global.config.port));

    let app = create_app(Arc::clone(&global))
        .with_state(Arc::clone(&global))
        .fallback(handle_404)
        .layer(CorsLayer::permissive().allow_headers(AllowHeaders::mirror_request()));
    let server = axum::Server::bind(&addr);
    println!("Listening on {}", addr);

    let mut sigterm = signal(SignalKind::terminate()).unwrap();
    let mut sigint = signal(SignalKind::interrupt()).unwrap();
    select! {
        _ = sigint.recv() => {
            println!("Received SIGINT!");
        }
        _ = sigterm.recv() => {
            println!("Received SIGTERM!");
        }
        _ = server.serve(app.into_make_service()) => {
            panic!("Server stopped!");
        }
        _ = worker::worker(global) => {
            panic!("Worker stopped!");
        }
    };
}

fn create_app(global: Arc<Global>) -> axr::Router<Arc<Global>> {
    axr::Router::new()
        .route(
            "/",
            axr::get(|| async { "Welcome to Archív.\n\ngithub.com.com/0d9e-tech/archiiv-rs" })
                .fallback(handle_method_not_allowed),
        )
        .nest("/auth", routes::auth::create_app())
        .nest(
            "/api",
            routes::api::create_app().route_layer(axum::middleware::from_fn_with_state(
                Arc::clone(&global),
                utils::auth_middleware,
            )),
        )
}

async fn ensure_dirs(global: &Arc<Global>) {
    let tmp = global.data_dir.join("_tmp");
    if let Err(e) = fs::remove_dir_all(&tmp) {
        assert!(
            e.kind() == ErrorKind::NotFound,
            "Failed to remove data/_tmp"
        );
    }
    fs::create_dir(&tmp).expect("Failed to create data/_tmp");

    let userdirs = global.data_dir.join("users");
    fs::create_dir_all(&userdirs).expect("Failed to create data/users");
    let listing = fs::read_dir(&userdirs)
        .expect("Failed to read data/users")
        .map(Result::unwrap)
        .collect::<Vec<_>>();

    let keys = global
        .get_users()
        .await
        .iter()
        .map(|(name, data)| (OsString::from(name), data.clone()))
        .collect::<Vec<_>>();
    for (username, data) in &keys {
        if data.session_tokens.is_empty() {
            continue;
        }
        let dir = userdirs.join(username);
        let found = listing.iter().find(|x| x.file_name() == *username);
        if let Some(x) = found {
            assert!(
                x.file_type().unwrap().is_dir(),
                "Path {} is not a directory",
                dir.display()
            );
        } else {
            eprintln!(
                "Directory for user {} didn't exist, creating it",
                username.to_str().unwrap()
            );
            fs::create_dir(&dir).expect(&format!("Failed to create directory {}", dir.display()));
        }
    }
    for item in listing {
        if !keys.iter().any(|(name, _)| *name == item.file_name()) {
            eprintln!(
                "Found directory for user {} who doesn't have account",
                item.file_name().to_string_lossy()
            );
        }
    }
}

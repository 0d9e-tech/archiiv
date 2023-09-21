mod consts;
mod global;
mod routes;
mod utils;
mod worker;

use std::{
    env,
    net::{Ipv4Addr, SocketAddr, SocketAddrV4},
    path::PathBuf,
    sync::Arc,
};

use axum::routing as axr;
use tokio::{
    select,
    signal::unix::{signal, SignalKind},
};

use crate::global::Global;

#[tokio::main]
async fn main() {
    let data_dir: PathBuf = env::var("DATA_DIR")
        .unwrap_or_else(|_| "data".to_owned())
        .try_into()
        .expect("DATA_DIR is not a valid path");
    let global = Arc::new(Global::load(data_dir.into_boxed_path()));
    let addr = SocketAddr::V4(SocketAddrV4::new(Ipv4Addr::UNSPECIFIED, global.config.port));

    let app = create_app(Arc::clone(&global)).with_state(Arc::clone(&global));
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
            axr::get(|| async { "Welcome to Archív.\n\ngithub.com.com/0d9e-tech/archiiv-rs" }),
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

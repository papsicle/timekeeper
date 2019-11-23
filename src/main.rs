use actix_web::{web, App, HttpServer, Responder, HttpRequest};
use serde::{Deserialize};

async fn rules() -> impl Responder {

    let rules_str = " \
        *Rules* Everyone starts at 21 pts. You gain 2 pts per perfect week. \
        You loose pts based on how many minutes you are late to a meeting. \
        You can buy back points by proposing stuff to the team and them telling you how much that is worth.\n\
        *Exceptions* You won't be counted as being late if you are actively fixing prod. \
        You won't be counted as being late if you can't attend and post your status update in time.";

    rules_str
}

async fn leaderboard() -> impl Responder {
    "This is a leaderboard"
}

async fn reset_all() -> impl Responder {
    "Nuking the DB"
}

async fn handle_member(info: web::Path<String>) -> impl Responder {
    format!("Hello {}!", info)
}

async fn assign_points(info: web::Path<(String, u32)>) -> impl Responder {
    format!("Giving {} points to {}!", info.1, info.0)
}

async fn remove_points(info: web::Path<(String, u32)>) -> impl Responder {
    format!("Penalizing {} for {} minutes", info.0, info.1)
}

// TODO: Perfect week as a "cleanup function (if no late in X, reset-auto (on point-request and leaderboard?))"

fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
            App::new()
              .route("/rules", web::get().to(rules))
              .route("/leaderboard", web::get().to(leaderboard))
              .route("/leaderboard/reset", web::get().to(reset_all))
              .route("/members/{name}", web::get().to(handle_member))
              .route("/members/{name}/points/{points}", web::post().to(assign_points))
              .route("/members/{name}/late/{minutes}", web::post().to(remove_points))

        })
        .bind("127.0.0.1:8080")?
        .run()
}

use actix_web::{web, middleware, App, HttpServer, Responder, HttpRequest};
use serde::{Serialize, Deserialize};
use bincode;
use sled::Db;
use std::sync::Arc;

#[derive(Serialize, Deserialize, Debug)]
struct MemberInfo {
    member_name: String,
    points: u32
}

async fn rules() -> impl Responder {

    let rules_str = " \
        *Rules* Everyone starts at 21 pts. You gain 2 pts per perfect week. \
        You loose pts based on how many minutes you are late to a meeting. \
        You can buy back points by proposing stuff to the team and them telling you how much that is worth.\n\
        *Exceptions* You won't be counted as being late if you are actively fixing prod. \
        You won't be counted as being late if you can't attend and post your status update in time.";

    rules_str
}

async fn leaderboard(st: web::Data<Arc<Db>>) -> impl Responder {

    let mut vec = Vec::new();

    st.iter().for_each(|x| {
        let (_, value) = x.unwrap();
        let unserialized = bincode::deserialize::<MemberInfo>(&value).unwrap();
        vec.push(format!("{}, {}", unserialized.member_name, unserialized.points));
    });

    format!("{:?}", vec)
}

async fn reset_all() -> impl Responder {
    "Nuking the DB"
}

async fn handle_member_add(st: web::Data<Arc<Db>>, member_name: web::Path<String>) -> impl Responder {

    let new_member = MemberInfo {
        member_name: member_name.clone(),
        points: 0,
    };

    st.insert(member_name.as_ref(), bincode::serialize(&new_member).unwrap()).unwrap(); // Unwrap like there is no tomorrow
    st.flush().unwrap(); // Unwrap like there is no tomorrow
    format!("Added {}!", member_name)
}

async fn assign_points(info: web::Path<(String, u32)>) -> impl Responder {
    format!("Giving {} points to {}!", info.1, info.0)
}

async fn remove_points(info: web::Path<(String, u32)>) -> impl Responder {
    format!("Penalizing {} for {} minutes", info.0, info.1)
}

// TODO: Perfect week as a "cleanup function (if no late in X, reset-auto (on point-request and leaderboard?))"

fn main() -> std::io::Result<()> {

    std::env::set_var("RUST_LOG", "actix_web=info");
    env_logger::init();

    let tree = Arc::new(Db::open("./timekeeper.sled.db").unwrap());

    HttpServer::new(move || {
            App::new()
              .wrap(middleware::Logger::default())
              .data(tree.clone())
              .route("/rules", web::get().to(rules))
              .route("/leaderboard", web::get().to(leaderboard))
              .route("/leaderboard/reset", web::get().to(reset_all))
              .route("/members/{name}/add", web::get().to(handle_member_add))
              .route("/members/{name}/points/{points}", web::post().to(assign_points))
              .route("/members/{name}/late/{minutes}", web::post().to(remove_points))

        })
        .bind("127.0.0.1:8080")?
        .run()
}

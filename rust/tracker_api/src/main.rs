use lambda_http::{run, service_fn, tracing, Body, Error, Request, RequestExt, Response};
use serde_json::json;

use template_utils::load_tera;
use tera::Context;
use tracker_analyzer::helpers::analyze_user_portfolio;

mod template_utils;

async fn function_handler(event: Request) -> Result<Response<Body>, Error> {
    if let Some(user_id) = event.query_string_parameters().first("user_id") {
        let portfolio = analyze_user_portfolio(user_id).await.unwrap();
        // let js = serde_json::to_value(&portfolio).unwrap();

        let mut ctx = Context::new();
        // ctx.insert("portfolio", &portfolio);
        ctx.insert("accounts", &portfolio.accounts_metadata);
        ctx.insert("accounts_stat", &portfolio.accounts);
        ctx.insert("portfolio", &portfolio.portfolio);

        let tera = load_tera();
        let result = tera.render("index.html", &ctx);
        let resp = Response::builder()
            .status(200)
            .header("content-type", "text/html")
            // .header("content-type", "application/json")
            // .body(Body::from(serde_json::to_vec(&js).unwrap()))
            .body(Body::Text(result.unwrap()))
            .map_err(Box::new)?;
        Ok(resp)
    } else {
        let resp = Response::builder()
            .status(400)
            .body(Body::from(json!({"error": "Missing user_id"}).to_string()))
            .map_err(Box::new)?;
        Ok(resp)
    }
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing::init_default_subscriber();

    run(service_fn(function_handler)).await
}

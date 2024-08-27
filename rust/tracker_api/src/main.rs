use lambda_http::{
    http::{response::Builder, Method, StatusCode},
    run, service_fn, tracing, Body, Error, Request, RequestExt, Response,
};
use serde_json::json;

use template_utils::load_tera;
use tera::Context;
use tracker_analyzer::helpers::analyze_user_portfolio;

mod template_utils;

fn base_response(req: &Request, status: StatusCode) -> Builder {
    let mut allow_origin = "https://tracker.arrakisholdings.com";

    if let Some(origin) = req.headers().get("Origin") {
        let s = origin.to_str().unwrap();
        if [
            allow_origin,
            "https://portfolio-tracker-8nd.pages.dev",
            "http://localhost:8000",
        ]
        .contains(&s)
        {
            allow_origin = s;
        }
    }

    Response::builder()
        .status(status)
        .header("content-type", "text/html")
        .header("access-control-allow-origin", allow_origin)
        .header("access-control-allow-methods", "get, post, options")
        .header(
            "access-control-allow-headers",
            "content-type, hx-request, hx-current-url, hx-trigger, hx-target",
        )
        .header("access-control-max-age", "7200")
}

async fn function_handler(event: Request) -> Result<Response<Body>, Error> {
    if event.method() == Method::OPTIONS {
        let resp = base_response(&event, StatusCode::OK)
            // .header("", "")
            .body(Body::default())
            .map_err(Box::new)?;

        return Ok(resp);
    }

    let currency = event
        .query_string_parameters_ref()
        .and_then(|params| params.first("currency"))
        .unwrap_or_else(|| "USD");

    if let Some(user_id) = event.query_string_parameters().first("user_id") {
        let portfolio = analyze_user_portfolio(user_id, currency).await.unwrap();
        // let js = serde_json::to_value(&portfolio).unwrap();

        let mut ctx = Context::new();
        // ctx.insert("portfolio", &portfolio);
        ctx.insert("accounts", &portfolio.accounts_metadata);
        ctx.insert("accounts_stat", &portfolio.accounts);
        ctx.insert("portfolio", &portfolio.portfolio);
        ctx.insert("currency", &portfolio.currency_md);
        ctx.insert("rate", &portfolio.rate);

        let tera = load_tera();
        // let result = tera.render("index.html", &ctx);
        let result = tera.render("accounts-table.html", &ctx);
        let resp = base_response(&event, StatusCode::OK)
            // .header("content-type", "application/json")
            // .body(Body::from(serde_json::to_vec(&js).unwrap()))
            .body(Body::Text(result.unwrap()))
            .map_err(Box::new)?;
        Ok(resp)
    } else {
        let resp = base_response(&event, StatusCode::FORBIDDEN)
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

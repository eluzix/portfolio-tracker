use std::collections::HashMap;

use lambda_http::{run, service_fn, tracing, Body, Error, Request, RequestExt, Response};
use numfmt::{Formatter, Precision};
use serde_json::json;

use tera::{to_value, Context, Tera, Value};
use tracker_analyzer::helpers::analyze_user_portfolio;

pub mod filters;

pub fn currency_filter(value: &Value, args: &HashMap<String, Value>) -> tera::Result<Value> {
    let currency = match args.get("sign") {
        Some(currency) => currency.as_str().unwrap(),
        _ => "$",
    };

    let mut f = Formatter::new() // start with blank representation
        .separator(',')
        .unwrap()
        .prefix(currency)
        .unwrap()
        .precision(Precision::Decimals(2));
    if let Some(val) = value.as_f64() {
        return Ok(to_value(f.fmt2(val)).unwrap());
    }

    Ok(to_value(value.to_string()).unwrap())
}

pub fn percent_filter(value: &Value, _: &HashMap<String, Value>) -> tera::Result<Value> {
    let mut f = Formatter::new() // start with blank representation
        .separator(',')
        .unwrap()
        .suffix("%")
        .unwrap()
        .precision(Precision::Decimals(2));
    if let Some(val) = value.as_f64() {
        return Ok(to_value(f.fmt2(val * 100.0)).unwrap());
    }

    Ok(to_value(value.to_string()).unwrap())
}
fn load_tera() -> Tera {
    let mut tera = match Tera::new("templates/**/*.html") {
        Ok(t) => t,
        Err(e) => {
            println!("Parsing error(s): {}", e);
            panic!("EEEE");
        }
    };
    tera.register_filter("currency_filter", currency_filter);
    tera.register_filter("percent_filter", percent_filter);

    tera
}

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

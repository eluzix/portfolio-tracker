use std::collections::HashMap;

use lambda_http::{
    http::{response::Builder, Method, StatusCode},
    run, service_fn, tracing, Body, Error, Request, RequestExt, Response,
};
use serde_json::json;

use template_utils::load_tera;
use tera::Context;
use tracker_analyzer::{
    helpers::{analyze_user_portfolio, transactions_by_account},
    store::{
        cache::{default_cache, Cache, DynamoCache},
        user_data::{self, load_user_data},
    },
    types::{
        account::AccountMetadata,
        transactions::{Transaction, TransactionType},
    },
};
use url::form_urlencoded;

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

async fn handle_index(user_id: &str, event: Request) -> Result<Response<Body>, Error> {
    let currency = event
        .query_string_parameters_ref()
        .and_then(|params| params.first("currency"))
        .unwrap_or_else(|| "USD");

    match event.query_string_parameters().first("ac") {
        Some("clean-cache") => {
            let c: std::sync::Arc<DynamoCache> = default_cache();
            c.clear("prices").await;
            c.clear("rates").await;
        }
        _ => {}
    }

    let portfolio = analyze_user_portfolio(user_id, currency).await.unwrap();

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
}

fn get_query_param_from_current_url(event: &Request, param: &str) -> Option<String> {
    if let Some(cur_req_url) = event.headers().get("HX-Current-Url") {
        if let Ok(u) = url::Url::parse(cur_req_url.to_str().unwrap()) {
            for (k, v) in u.query_pairs() {
                if k == param {
                    return Some(v.to_string());
                }
            }
        }
    }

    None
}

async fn handle_add_transaction(
    user_id: &str,
    account_id: &str,
    event: &Request,
) -> Option<String> {
    let body: String = match event.body() {
        Body::Text(b) => b.clone(),
        Body::Binary(b) => String::from_utf8(b.clone()).unwrap_or_else(|_| "".to_string()),
        _ => return None,
    };

    let mut params: HashMap<String, String> = HashMap::new();
    for (key, value) in form_urlencoded::parse(body.as_bytes()) {
        params.insert(key.into(), value.into());
    }

    let tr = Transaction {
        id: "".to_string(),
        account_id: account_id.to_string(),
        symbol: params.get("symbol").unwrap().clone(),
        date: params.get("date").unwrap().clone(),
        transaction_type: TransactionType::from(params.get("type").unwrap().clone()),
        quantity: params.get("quantity").unwrap().parse::<u32>().unwrap(),
        pps: params.get("pps").unwrap().parse::<f64>().unwrap(),
    };

    match user_data::add_transaction(user_id, &tr).await {
        Ok(tr_id) => {
            println!("New transaction id ::::> {:?}", tr_id);
            return Some(tr_id);
        }
        Err(err) => println!("Error adding transaction: {:?}", err),
    }

    None
}

async fn handle_delete_transaction(user_id: &str, event: &Request) -> Option<String> {
    if let Some(tr_id) = get_query_param_from_current_url(&event, "tr_id") {
        match user_data::delete_transaction(user_id, tr_id.as_str()).await {
            Ok(tr_id) => {
                println!("Deleted transaction id ::::> {:?}", tr_id);
                return Some(tr_id);
            }
            Err(err) => println!("Error deleting transaction: {:?}", err),
        }
    }

    None
}

async fn handle_transactions(user_id: &str, event: Request) -> Result<Response<Body>, Error> {
    if let Some(account_id) = get_query_param_from_current_url(&event, "account_id") {
        let account_id: String = account_id;
        match event.query_string_parameters().first("ac") {
            Some("add-transaction") => {
                if let None = handle_add_transaction(user_id, account_id.as_str(), &event).await {
                    println!("Error adding transaction");
                }
            }
            Some("delete-transaction") => {
                if let None = handle_delete_transaction(user_id, &event).await {
                    println!("Error deleting transaction");
                }
            }
            _ => {}
        }

        let resp = load_user_data(user_id).await.unwrap();
        let account: AccountMetadata = resp.1.into_iter().find(|a| a.id == account_id).unwrap();
        let transactions: HashMap<String, Vec<Transaction>> = transactions_by_account(&resp.0);

        let mut ctx = Context::new();
        ctx.insert("account", &account);
        ctx.insert("transactions", &transactions.get(&account_id).unwrap());

        let tera = load_tera();
        let result = tera.render("account-transactions.html", &ctx);
        let resp = base_response(&event, StatusCode::OK)
            .body(Body::Text(result.unwrap()))
            .map_err(Box::new)?;
        Ok(resp)
    } else {
        let resp = base_response(&event, StatusCode::FORBIDDEN)
            .body(Body::from(
                json!({"error": "Missing account id"}).to_string(),
            ))
            .map_err(Box::new)?;
        Ok(resp)
    }
}

async fn function_handler(event: Request) -> Result<Response<Body>, Error> {
    if event.method() == Method::OPTIONS {
        let resp = base_response(&event, StatusCode::OK)
            // .header("", "")
            .body(Body::default())
            .map_err(Box::new)?;

        return Ok(resp);
    }

    let page = event
        .query_string_parameters_ref()
        .and_then(|params| params.first("page"))
        .unwrap_or_else(|| "index");

    if let Some(user_id) = event.query_string_parameters().first("user_id") {
        match page {
            "index" => {
                return handle_index(user_id, event).await;
            }
            "transactions" => {
                return handle_transactions(user_id, event).await;
            }
            _ => {}
        }
    }

    let resp = base_response(&event, StatusCode::FORBIDDEN)
        .body(Body::from(json!({"error": "Missing user_id"}).to_string()))
        .map_err(Box::new)?;
    Ok(resp)
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing::init_default_subscriber();

    run(service_fn(function_handler)).await
}

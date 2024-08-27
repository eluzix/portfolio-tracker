use std::collections::HashMap;

use numfmt::{Formatter, Precision};
use serde::Deserialize;
use serde_json;
use tera::{to_value, Context, Tera, Value};
use tracker_analyzer::helpers::{
    analyze_user_portfolio, merge_dividends, sort_transactions_by_date,
};
use tracker_analyzer::store::cache::{self, default_cache};
use tracker_analyzer::store::market::MarketStackResponse;
use tracker_analyzer::store::user_data::load_user_data;
use tracker_analyzer::store::{market, tracker_config};
use tracker_analyzer::types::transactions::Transaction;

async fn print_all() {
    let portfolio = analyze_user_portfolio("1", "USD").await.unwrap();

    for (account_id, portfolio_data) in portfolio.accounts.iter() {
        let account_data = portfolio.accounts_metadata.get(account_id).unwrap();
        println!("--------\nAccount: {}", account_data.name);
        println!("Portfolio: {:?}", portfolio_data);
    }

    let all_portfolio_data = portfolio.portfolio;
    println!("--------\nAll Accounts");
    println!("Portfolio: {:?}", all_portfolio_data);
}

async fn test_price() {
    // let inp = vec!["AAPL".to_string(), "GOOG".to_string()];
    // let inp2 = vec!["AAPL", "GOOG"];
    let inp2 = vec![
        "INTC", "SCHD", "SPY", "WIX", "HOG", "AAPL", "VT", "HDV", "PFE", "VTI", "TSLA", "BRK.B",
    ];
    let cache = default_cache();
    let prices = market::load_prices(&*cache, &inp2).await;
    // let prices = PricesClient::fetch_prices(&inp).await;
    println!("prices ====>>>> {:?}", prices);
}

async fn test_market() {
    let symbols = [
        "INTC", "SCHD", "SPY", "WIX", "HOG", "AAPL", "VT", "HDV", "PFE", "VTI", "TSLA", "BRK.B",
    ];
    let key: String = tracker_config::get("marketstack_key").unwrap();
    let client = reqwest::Client::new();
    let res: String = client
        .get("https://api.marketstack.com/v1/eod/latest")
        .query(&[("symbols", symbols.join(",")), ("access_key", key)])
        .send()
        .await
        .unwrap()
        .text()
        .await
        .unwrap();

    let res = res.replace(",[]", "");
    println!("----->> res: {:?}", res);

    let js = serde_json::from_str::<MarketStackResponse>(&res);

    // .unwrap()
    // .json::<MarketStackResponse>()
    // .await;
    // .text()
    // .await;

    println!(">>>>>>>>>>>> {:?}", js);
}

async fn test_dividends() {
    // let symbols = ["AAPL", "BRK-B"];
    let symbols = ["TSLA", "BRK-B", "WIX"];

    let cache = default_cache();
    let d = market::load_dividends(&*cache.clone(), &symbols)
        .await
        .unwrap();
    let dflt: Vec<Transaction> = vec![];
    let tsla = d.get("TSLA").unwrap_or_else(|| &dflt);

    println!(">>>>>>>> tsla dividends: {:?}", tsla);
}

async fn test_transactions() {
    let user_id = "1";
    let cache = default_cache();

    let resp = load_user_data(user_id).await.unwrap();
    let mut transactions = resp.0;
    let d = market::load_dividends(&*cache.clone(), &["VT"]).await;
    merge_dividends(&mut transactions, &d);
    sort_transactions_by_date(&mut transactions);
    let first_tr = transactions.first().unwrap();
    println!(">>>>>>> {:?}", first_tr);
}

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

async fn test_template() {
    let mut tera = match Tera::new("templates/**/*.html") {
        Ok(t) => t,
        Err(e) => {
            println!("Parsing error(s): {}", e);
            panic!("EEEE");
        }
    };
    tera.register_filter("currency_filter", currency_filter);
    tera.register_filter("percent_filter", percent_filter);

    let portfolio = analyze_user_portfolio("1", "ILS").await.unwrap();
    let mut ctx = Context::new();
    ctx.insert("portfolio", &portfolio.portfolio);
    ctx.insert("accounts", &portfolio.accounts_metadata);
    ctx.insert("accounts_stat", &portfolio.accounts);
    ctx.insert("currency", &portfolio.currency);
    ctx.insert("rate", &portfolio.rate);

    // let result = tera.render("index.html", &ctx);
    let result = tera.render("accounts-table.html", &ctx);
    println!("{:?}", result);
}

async fn test_exhange() {
    let c = "ILS";
    let cache = default_cache();
    if let Ok(res) = market::load_exhnage_rate(&*cache, c).await {
        println!("for {} rate: {}", c, res);
    }
}

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), ()> {
    // print_all().await;
    // test_price().await;
    // test_market().await;
    // test_dividends().await;
    // test_transactions().await;
    // test_exhange().await;
    test_template().await;
    Ok(())
}

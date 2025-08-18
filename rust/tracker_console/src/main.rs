use std::collections::HashMap;
use std::fs::File;
use std::io::{BufWriter, Write};

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
    let resp = load_user_data("1").await.unwrap();
    
    // Dump transactions
    let file = File::create("transactions.jsonl").expect("Failed to create transactions.jsonl");
    let mut writer = BufWriter::new(file);
    
    for transaction in resp.0 {
        let mut json_value = serde_json::to_value(&transaction).expect("Failed to serialize transaction");
        
        // Convert pps from f64 to u32 cents (e.g., 32.42 -> 3242, 22.1 -> 2210)
        if let Some(pps_val) = json_value.get_mut("pps") {
            if let Some(pps_f64) = pps_val.as_f64() {
                let pps_cents = (pps_f64 * 100.0).round() as u32;
                *pps_val = serde_json::Value::Number(serde_json::Number::from(pps_cents));
            }
        }
        
        let json_line = serde_json::to_string(&json_value).expect("Failed to serialize modified transaction");
        writeln!(writer, "{}", json_line).expect("Failed to write to file");
    }
    
    writer.flush().expect("Failed to flush writer");
    println!("Transactions dumped to transactions.jsonl");

    // Dump account metadata
    let account_file = File::create("accounts.jsonl").expect("Failed to create accounts.jsonl");
    let mut account_writer = BufWriter::new(account_file);
    
    for account in resp.1 {
        let json_line = serde_json::to_string(&account).expect("Failed to serialize account");
        writeln!(account_writer, "{}", json_line).expect("Failed to write to file");
    }
    
    account_writer.flush().expect("Failed to flush account writer");
    println!("Account metadata dumped to accounts.jsonl");

    // let portfolio = analyze_user_portfolio("1", "USD").await.unwrap();
    //
    // for (account_id, portfolio_data) in portfolio.accounts.iter() {
    //     let account_data = portfolio.accounts_metadata.get(account_id).unwrap();
    //     println!("--------\nAccount: {:?}", account_data);
    //     // println!("Portfolio: {:?}", portfolio_data);
    // }

    // let all_portfolio_data = portfolio.portfolio;
    // println!("--------\nAll Accounts");
    // println!("Portfolio: {:?}", all_portfolio_data);
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
    if let Ok(d) = market::load_dividends(&*cache.clone(), &["VT"]).await {
        merge_dividends(&mut transactions, &d);
        sort_transactions_by_date(&mut transactions);
        let first_tr = transactions.first().unwrap();
        println!(">>>>>>> {:?}", first_tr);
    }
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

    let portfolio = analyze_user_portfolio("1", "USD").await.unwrap();
    let mut ctx = Context::new();
    ctx.insert("portfolio", &portfolio.portfolio);
    ctx.insert("accounts", &portfolio.accounts_metadata);
    ctx.insert("accounts_stat", &portfolio.accounts);
    ctx.insert("currency", &portfolio.currency_md);
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

async fn test_cur_metadata() {
    let c = "ILS";
    let cache = default_cache();
    let res = market::load_currency_metadata(&*cache, c).await;
    println!("for {:?} MD: {:?}", c, res)
}

async fn test_splits() {
    // let symbols = ["AAPL", "BRK-B"];
    let symbols = ["SCHD"];

    let cache = default_cache();
    let d = market::load_splits(&*cache.clone(), &symbols)
        .await
        .unwrap();
    let dflt: Vec<Transaction> = vec![];
    let s = d.get("SCHD").unwrap_or_else(|| &dflt);

    println!(">>>>>>>> SCHD splits: {:?}", s);
}

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), ()> {
    print_all().await;
    // test_price().await;
    // test_market().await;
    // test_dividends().await;
    // test_splits().await;
    // test_transactions().await;
    // test_exhange().await;
    // test_cur_metadata().await;
    // test_template().await;
    Ok(())
}

use serde_json;
use tracker_analyzer::helpers::analyze_user_portfolio;
use tracker_analyzer::store::cache::{self, default_cache};
use tracker_analyzer::store::market::MarketStackResponse;
use tracker_analyzer::store::{market, tracker_config};

async fn print_all() {
    let portfolio = analyze_user_portfolio("1").await.unwrap();

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
    let symbols = ["AAPL", "BRK-B"];
    let cache = default_cache();
    let d = market::load_dividends(&*cache.clone(), &symbols).await;
    println!(">>>>>>>> dividends: {:?}", d);
}

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), ()> {
    // print_all().await;
    // test_price().await;
    // test_market().await;
    test_dividends().await;
    Ok(())
}

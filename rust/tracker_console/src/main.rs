use tracker_analyzer::helpers::analyze_user_portfolio;
use tracker_analyzer::store::market::{PriceFetcher, PricesClient};

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
    let inp = vec!["AAPL".to_string(), "GOOG".to_string()];
    let prices = PricesClient::fetch_prices(&inp).await;
    println!("prices ====>>>> {:?}", prices);
}

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), ()> {
    print_all().await;
    // test_price().await;
    Ok(())
}

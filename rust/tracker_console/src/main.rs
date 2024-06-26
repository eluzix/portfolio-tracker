use tracker_analyzer::helpers::analyze_user_portfolio;

/// Lists your DynamoDB tables in the default Region or us-east-1 if a default Region isn't set.
#[tokio::main]
async fn main() -> Result<(), ()> {
    let portfolio = analyze_user_portfolio("1").await.unwrap();

    for (account_id, portfolio_data) in portfolio.accounts.iter() {
        println!("--------\nAccount: {}", account_id);
        println!("Portfolio: {:?}", portfolio_data);
    }

    let all_portfolio_data = portfolio.portfolio;
    println!("--------\nAll Accounts");
    println!("Portfolio: {:?}", all_portfolio_data);

    Ok(())
}

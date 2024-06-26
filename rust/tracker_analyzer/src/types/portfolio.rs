use std::collections::HashSet;
use serde::{Deserialize, Serialize};

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct AnalyzedPortfolio {
    // stats and variables
    pub exchange_rate: f64,
    pub avg_pps: f64,
    pub symbols: HashSet<String>,

    // totals and gains
    pub total_invested: f64,
    pub total_withdrawn: f64,
    pub total_dividends: f64,

    pub current_portfolio_value: f64,

    // gains and yields

    // gain in percentage
    pub portfolio_gain: f64,
    // gain in value
    pub portfolio_gain_value: f64,
    pub annualized_yield: f64,
    pub modified_dietz_yield: f64,
}

impl AnalyzedPortfolio {
    pub fn new() -> AnalyzedPortfolio {
        AnalyzedPortfolio {
            symbols: HashSet::with_capacity(0),
            exchange_rate: 1.0,
            avg_pps: 0.0,
            total_invested: 0.0,
            total_withdrawn: 0.0,
            total_dividends: 0.0,
            current_portfolio_value: 0.0,
            portfolio_gain: 0.0,
            portfolio_gain_value: 0.0,
            annualized_yield: 0.0,
            modified_dietz_yield: 0.0,
        }
    }
}

impl Default for AnalyzedPortfolio {
    fn default() -> Self {
        AnalyzedPortfolio::new()
    }
}


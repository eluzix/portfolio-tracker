use crate::types::portfolio::AnalyzedPortfolio;
use crate::{store::market::CurrencyMetadata, types::account::AccountMetadata};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct UserPortfolio {
    pub accounts_metadata: HashMap<String, AccountMetadata>,
    pub accounts: HashMap<String, AnalyzedPortfolio>,
    pub portfolio: AnalyzedPortfolio,
    pub rate: f64,
    pub currency_md: CurrencyMetadata,
}

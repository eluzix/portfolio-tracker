use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use crate::types::account::AccountMetadata;
use crate::types::portfolio::AnalyzedPortfolio;

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct UserPortfolio {
    pub accounts_metadata: HashMap<String, AccountMetadata>,
    pub accounts: HashMap<String, AnalyzedPortfolio>,
    pub portfolio: AnalyzedPortfolio,
}

use crate::types::account::AccountMetadata;
use crate::types::portfolio::AnalyzedPortfolio;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct UserPortfolio {
    pub accounts_metadata: HashMap<String, AccountMetadata>,
    pub accounts: HashMap<String, AnalyzedPortfolio>,
    pub portfolio: AnalyzedPortfolio,
}

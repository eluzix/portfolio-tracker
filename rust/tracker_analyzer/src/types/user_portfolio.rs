use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use crate::types::portfolio::AnalyzedPortfolio;

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct UserPortfolio {
    pub accounts: HashMap<String, AnalyzedPortfolio>,
    pub portfolio: AnalyzedPortfolio,
}

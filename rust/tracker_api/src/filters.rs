use tera::Filter;

pub struct CurrencyFilter {}

impl Filter for CurrencyFilter {
    fn filter(
        &self,
        value: &tera::Value,
        args: &std::collections::HashMap<String, tera::Value>,
    ) -> tera::Result<tera::Value> {
        tera::Result::Ok(value.clone())
    }
}

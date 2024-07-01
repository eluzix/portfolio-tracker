use std::collections::HashMap;
use aws_sdk_dynamodb::types::AttributeValue;
use serde::{Deserialize, Serialize};

#[derive(Debug, PartialEq, Serialize, Deserialize, Default)]
pub struct AccountMetadata {
    pub id: String,
    pub name: String,
    pub owner: String,
    pub institution: String,
    pub institution_id: Option<String>,
    pub description: Option<String>,
    pub tags: Vec<String>,
    pub created_at: Option<u32>,
    pub updated_at: Option<u32>,
}

impl AccountMetadata {
    pub fn from_dynamodb(item: HashMap<String, AttributeValue>) -> Self {
        let id = item.get("id").unwrap().as_s().unwrap().to_string();
        let name = item.get("name").unwrap().as_s().unwrap().to_string();
        let owner = item.get("owner").unwrap().as_s().unwrap().to_string();
        let institution = item.get("institution").unwrap().as_s().unwrap().to_string();
        let institution_id = item.get("institution_id").map(|v| v.as_s().unwrap().to_string());
        let description = item.get("description").map(|v| v.as_s().unwrap().to_string());
        let created_at = item.get("created_at").map(|v| v.as_n().unwrap().parse().unwrap());
        let updated_at = item.get("updated_at").map(|v| v.as_n().unwrap().parse().unwrap());
        let mut tags = Vec::new();
        if let Some(tags_attr) = item.get("tags") {
            for tag in tags_attr.as_l().unwrap() {
                tags.push(tag.as_s().unwrap().to_string());
            }
        }

        AccountMetadata {
            id,
            name,
            owner,
            institution,
            institution_id,
            description,
            tags,
            created_at,
            updated_at,
        }
    }
}

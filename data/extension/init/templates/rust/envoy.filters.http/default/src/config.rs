use std::convert::TryFrom;

use serde::Deserialize;

use envoy_sdk::extension;

/// Configuration for a sample HTTP filter.
#[derive(Deserialize, Debug)]
pub struct SampleHttpFilterConfig {
    #[serde(default)]
    pub param: String,
}

impl TryFrom<&[u8]> for SampleHttpFilterConfig {
    type Error = extension::Error;

    /// Parses filter configuration from JSON.
    fn try_from(value: &[u8]) -> extension::Result<Self> {
        serde_json::from_slice(value).map_err(extension::Error::new)
    }
}

impl Default for SampleHttpFilterConfig {
    /// Creates the default configuration.
    fn default() -> Self {
        SampleHttpFilterConfig {
            param: String::new(),
        }
    }
}

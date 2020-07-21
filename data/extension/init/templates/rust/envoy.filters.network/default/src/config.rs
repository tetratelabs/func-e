use std::convert::TryFrom;

use serde::Deserialize;

use envoy_sdk::extension;

/// Configuration for a sample network filter.
#[derive(Deserialize, Debug)]
pub struct SampleNetworkFilterConfig {
    #[serde(default)]
    pub param: String,
}

impl TryFrom<&[u8]> for SampleNetworkFilterConfig {
    type Error = extension::Error;

    /// Parses filter configuration from JSON.
    fn try_from(value: &[u8]) -> extension::Result<Self> {
        serde_json::from_slice(value).map_err(extension::Error::new)
    }
}

impl Default for SampleNetworkFilterConfig {
    /// Creates the default configuration.
    fn default() -> Self {
        SampleNetworkFilterConfig {
            param: String::new(),
        }
    }
}

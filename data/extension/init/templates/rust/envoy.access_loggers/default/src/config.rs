use std::convert::TryFrom;

use serde::Deserialize;

use envoy_sdk::extension;

/// Configuration for a sample access logger.
#[derive(Deserialize, Debug)]
pub struct SampleAccessLoggerConfig {
    #[serde(default)]
    pub param: String,
}

impl TryFrom<&[u8]> for SampleAccessLoggerConfig {
    type Error = extension::Error;

    fn try_from(value: &[u8]) -> extension::Result<Self> {
        serde_json::from_slice(value).map_err(extension::Error::new)
    }
}

impl Default for SampleAccessLoggerConfig {
    /// Creates the default configuration.
    fn default() -> Self {
        SampleAccessLoggerConfig {
            param: String::new(),
        }
    }
}

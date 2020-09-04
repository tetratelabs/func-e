use std::convert::TryFrom;

use serde::Deserialize;

use envoy::extension;

/// Configuration for a Sample HTTP Filter.
#[derive(Debug, Default, Deserialize)]
pub struct SampleHttpFilterConfig {}

impl TryFrom<&[u8]> for SampleHttpFilterConfig {
    type Error = extension::Error;

    /// Parses filter configuration from JSON.
    fn try_from(value: &[u8]) -> extension::Result<Self> {
        serde_json::from_slice(value).map_err(Self::Error::from)
    }
}

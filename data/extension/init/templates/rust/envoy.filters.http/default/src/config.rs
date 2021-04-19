use std::convert::TryFrom;

use nanoserde::DeJson;

use envoy::extension;

/// Configuration for a Sample HTTP Filter.
#[derive(Debug, Default, DeJson)]
pub struct SampleHttpFilterConfig {}

impl TryFrom<&[u8]> for SampleHttpFilterConfig {
    type Error = extension::Error;

    /// Parses filter configuration from JSON.
    fn try_from(value: &[u8]) -> extension::Result<Self> {
        let json = String::from_utf8(value.to_vec())?;
        DeJson::deserialize_json(&json).map_err(extension::Error::from)
    }
}

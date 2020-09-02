use std::convert::TryFrom;

use serde::Deserialize;

use envoy::extension;

/// Configuration for a Sample Access Logger.
#[derive(Debug, Default, Deserialize)]
pub struct SampleAccessLoggerConfig {}

impl TryFrom<&[u8]> for SampleAccessLoggerConfig {
    type Error = extension::Error;

    fn try_from(value: &[u8]) -> extension::Result<Self> {
        serde_json::from_slice(value).map_err(extension::Error::from)
    }
}

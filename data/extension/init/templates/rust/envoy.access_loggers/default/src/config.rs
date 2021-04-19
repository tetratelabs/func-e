use std::convert::TryFrom;

use nanoserde::DeJson;

use envoy::extension;

/// Configuration for a Sample Access Logger.
#[derive(Debug, Default, DeJson)]
pub struct SampleAccessLoggerConfig {}

impl TryFrom<&[u8]> for SampleAccessLoggerConfig {
    type Error = extension::Error;

    fn try_from(value: &[u8]) -> extension::Result<Self> {
        let json = String::from_utf8(value.to_vec())?;
        DeJson::deserialize_json(&json).map_err(extension::Error::from)
    }
}

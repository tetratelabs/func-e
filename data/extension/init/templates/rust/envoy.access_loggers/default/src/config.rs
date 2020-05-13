/// Configuration for a sample access logger.
pub struct SampleAccessLoggerConfig {
    pub value: String,
}

impl SampleAccessLoggerConfig {
    /// Creates a new configuration.
    pub fn new<T: Into<String>>(value: T) -> SampleAccessLoggerConfig {
        SampleAccessLoggerConfig {
            value: value.into(),
        }
    }
}

impl Default for SampleAccessLoggerConfig {
    /// Creates the default configuration.
    fn default() -> Self {
        SampleAccessLoggerConfig {
            value: String::new(),
        }
    }
}

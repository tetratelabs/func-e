/// Configuration for a sample HTTP filter.
pub struct SampleHttpFilterConfig {
    pub value: String,
}

impl SampleHttpFilterConfig {
    /// Creates a new configuration.
    pub fn new<T: Into<String>>(value: T) -> SampleHttpFilterConfig {
        SampleHttpFilterConfig {
            value: value.into(),
        }
    }
}

impl Default for SampleHttpFilterConfig {
    /// Creates the default configuration.
    fn default() -> Self {
        SampleHttpFilterConfig {
            value: String::new(),
        }
    }
}

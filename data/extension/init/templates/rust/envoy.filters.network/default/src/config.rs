/// Configuration for a sample network filter.
pub struct SampleNetworkFilterConfig {
    pub value: String,
}

impl SampleNetworkFilterConfig {
    /// Creates a new configuration.
    pub fn new<T: Into<String>>(value: T) -> SampleNetworkFilterConfig {
        SampleNetworkFilterConfig {
            value: value.into(),
        }
    }
}

impl Default for SampleNetworkFilterConfig {
    /// Creates the default configuration.
    fn default() -> Self {
        SampleNetworkFilterConfig {
            value: String::new(),
        }
    }
}

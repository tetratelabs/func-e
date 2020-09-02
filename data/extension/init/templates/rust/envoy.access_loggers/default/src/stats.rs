use envoy::host::stats::Counter;

// Sample stats.
pub struct SampleAccessLoggerStats {
    log_entries_total: Box<dyn Counter>,
}

impl SampleAccessLoggerStats {
    pub fn new(log_entries_total: Box<dyn Counter>) -> Self {
        SampleAccessLoggerStats { log_entries_total }
    }

    pub fn log_entries_total(&self) -> &dyn Counter {
        &*self.log_entries_total
    }
}

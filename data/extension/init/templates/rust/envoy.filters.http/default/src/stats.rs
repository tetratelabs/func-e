use envoy::host::stats::Counter;

// Sample stats.
pub struct SampleHttpFilterStats {
    requests_total: Box<dyn Counter>,
}

impl SampleHttpFilterStats {
    pub fn new(requests_total: Box<dyn Counter>) -> Self {
        SampleHttpFilterStats { requests_total }
    }

    pub fn requests_total(&self) -> &dyn Counter {
        &*self.requests_total
    }
}

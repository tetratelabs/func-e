use envoy_sdk::host::services::metrics::{Counter, Gauge, Histogram};

// Sample stats.
pub struct SampleHttpFilterStats {
    requests_total: Box<dyn Counter>,
    requests_active: Box<dyn Gauge>,
    response_body_size_bytes: Box<dyn Histogram>,
}

impl SampleHttpFilterStats {
    pub fn new(
        requests_total: Box<dyn Counter>,
        requests_active: Box<dyn Gauge>,
        response_body_size_bytes: Box<dyn Histogram>,
    ) -> Self {
        SampleHttpFilterStats {
            requests_total,
            requests_active,
            response_body_size_bytes,
        }
    }

    pub fn requests_total(&self) -> &dyn Counter {
        &*self.requests_total
    }
    pub fn requests_active(&self) -> &dyn Gauge {
        &*self.requests_active
    }
    pub fn response_body_size_bytes(&self) -> &dyn Histogram {
        &*self.response_body_size_bytes
    }
}

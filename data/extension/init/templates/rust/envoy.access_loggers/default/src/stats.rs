use envoy_sdk::host::services::metrics::{Counter, Gauge};

// Sample stats.
pub struct SampleAccessLoggerStats {
    requests_total: Box<dyn Counter>,
    reports_active: Box<dyn Gauge>,
    reports_total: Box<dyn Counter>,
}

impl SampleAccessLoggerStats {
    pub fn new(
        requests_total: Box<dyn Counter>,
        reports_active: Box<dyn Gauge>,
        reports_total: Box<dyn Counter>,
    ) -> Self {
        SampleAccessLoggerStats {
            requests_total,
            reports_active,
            reports_total,
        }
    }

    pub fn requests_total(&self) -> &dyn Counter {
        &*self.requests_total
    }
    pub fn reports_active(&self) -> &dyn Gauge {
        &*self.reports_active
    }
    pub fn reports_total(&self) -> &dyn Counter {
        &*self.reports_total
    }
}

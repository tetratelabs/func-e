use envoy::host::stats::Counter;

// Sample stats.
pub struct SampleNetworkFilterStats {
    connections_total: Box<dyn Counter>,
}

impl SampleNetworkFilterStats {
    pub fn new(connections_total: Box<dyn Counter>) -> Self {
        SampleNetworkFilterStats { connections_total }
    }

    pub fn connections_total(&self) -> &dyn Counter {
        &*self.connections_total
    }
}

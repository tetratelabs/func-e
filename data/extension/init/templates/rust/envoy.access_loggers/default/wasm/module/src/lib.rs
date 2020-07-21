use proxy_wasm::traits::RootContext;
use proxy_wasm::types::LogLevel;

use envoy_sdk::extension;
use envoy_sdk::start;

use envoy_sample_access_logger::SampleAccessLogger;

// Generate a `_start` function that is called by Envoy
// when a new instance of WebAssembly module is created.
start! { on_module_start(); }

/// Does one-time initialization.
fn on_module_start() {
    proxy_wasm::set_log_level(LogLevel::Info);

    // Register Access logger extension
    proxy_wasm::set_root_context(|_| -> Box<dyn RootContext> {
        // Inject dependencies on Envoy host APIs
        let logger =
            SampleAccessLogger::with_default_ops().expect("unable to initialize extension");

        // Bridge between Access logger abstraction and Envoy ABI
        Box::new(extension::access_logger::LoggerContext::with_default_ops(
            logger,
        ))
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_start() {
        on_module_start()
    }
}

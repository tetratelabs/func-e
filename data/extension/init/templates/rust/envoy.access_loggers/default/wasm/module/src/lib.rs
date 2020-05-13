use proxy_wasm::traits::RootContext;
use proxy_wasm::types::LogLevel;

use envoy_sdk::extension;
use envoy_sdk::host::services::clients;
use envoy_sdk::host::services::time;

use envoy_sample_access_logger::SampleAccessLogger;

#[cfg_attr(
    all(
        target_arch = "wasm32",
        target_vendor = "unknown",
        target_os = "unknown"
    ),
    export_name = "_start"
)]
#[no_mangle]
/// Is called when a new instance of WebAssembly module gets created.
///
/// In general, a single WebAssembly module can include multiple extensions.
/// Envoy will instantiate the module on first use of any of them.
///
/// A module with an Access Logger extension (like this one) will be
/// instantiated at a time Envoy creates a Listener that makes
/// use of that logger.
/// Nonetheless, lifecycle of a WebAssembly module instance is separate
/// from the one of a Listener.
/// A single WebAssembly module instance can be shared by multiple Listeners
/// and will outlive all of them.
extern "C" fn start() {
    proxy_wasm::set_log_level(LogLevel::Info);
    proxy_wasm::set_root_context(|_| -> Box<dyn RootContext> {
        // Inject dependencies on Envoy host APIs
        let logger = SampleAccessLogger::new(&time::ops::Host, &clients::http::ops::Host);
        Box::new(extension::access_logger::LoggerContext::new(
            logger,
            &extension::access_logger::ops::Host,
            &clients::http::ops::Host,
        ))
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_start() {
        start()
    }
}

use proxy_wasm::traits::{ChildContext, RootContext};
use proxy_wasm::types::LogLevel;

use envoy_sdk::extension;
use envoy_sdk::extension::factory;
use envoy_sdk::extension::filter::http;
use envoy_sdk::start;

use envoy_sample_http_filter::SampleHttpFilterFactory;

// Generate a `_start` function that is called by Envoy
// when a new instance of WebAssembly module is created.
start! { on_module_start(); }

/// Does one-time initialization.
fn on_module_start() {
    proxy_wasm::set_log_level(LogLevel::Info);

    // Register HTTP filter extension
    proxy_wasm::set_root_context(|_| -> Box<dyn RootContext> {
        // Inject dependencies on Envoy host APIs
        let http_filter_factory =
            SampleHttpFilterFactory::with_default_ops().expect("unable to initialize extension");

        // Bridge between HTTP filter factory abstraction and Envoy ABI
        Box::new(factory::FactoryContext::with_default_ops(
            http_filter_factory,
            |http_filter_factory, instance_id| -> ChildContext {
                let http_filter = <_ as extension::factory::Factory>::new_extension(
                    http_filter_factory,
                    instance_id,
                )
                .unwrap();

                // Bridge between HTTP filter abstraction and Envoy ABI
                ChildContext::HttpContext(Box::new(http::FilterContext::with_default_ops(
                    http_filter,
                )))
            },
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

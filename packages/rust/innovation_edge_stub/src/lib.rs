//! Optional Rust edge inference/cache stub for innovation edge nodes.

pub fn edge_cache_key(tenant: &str, key: &str) -> String {
    format!("edge:{tenant}:{key}")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn key_format() {
        assert_eq!(edge_cache_key("t", "n"), "edge:t:n");
    }
}

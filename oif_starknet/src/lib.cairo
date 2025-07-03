pub mod permit2 {
    pub mod interface;
    pub mod permit2;
    pub mod allowance_transfer {
        pub mod allowance_transfer;
        pub mod interface;
    }
    pub mod signature_transfer {
        pub mod interface;
        pub mod signature_transfer;
    }

    pub mod unordered_nonces {
        pub mod interface;
        pub mod unordered_nonces;
    }
}

pub mod libraries {
    pub mod allowance;
    pub mod bitmap;
    pub mod permit_hash;
    pub mod utils;
    pub mod mocks {
        pub mod erc20;
    }
}

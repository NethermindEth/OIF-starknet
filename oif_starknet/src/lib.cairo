pub mod libraries {
    pub mod allowance;
    pub mod bitmap;
    pub mod permit_hash;
    pub mod unordered_nonces;
    pub mod utils;
}
pub mod mocks {
    pub mod mock_account;
    pub mod mock_erc20;
    pub mod mock_erc20_permit;
    pub mod mock_witness;
}
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
}

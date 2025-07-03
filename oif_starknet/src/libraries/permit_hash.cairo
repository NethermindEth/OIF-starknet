use core::hash::{HashStateExTrait, HashStateTrait};
use core::poseidon::PoseidonTrait;
use oif_starknet::libraries::utils::selector;
use oif_starknet::permit2::allowance_transfer::interface::{
    PermitBatch, PermitDetails, PermitSingle,
};
use oif_starknet::permit2::signature_transfer::interface::{
    PermitBatchTransferFrom, PermitTransferFrom, TokenPermissions,
};
use openzeppelin_utils::cryptography::snip12::{
    SNIP12HashSpanImpl, SNIP12Metadata, StarknetDomain, StructHash,
};
use starknet::{ContractAddress, get_caller_address, get_tx_info};

/// TYPE_HASHES & TYPE_STRINGS
pub const _U256_TYPE_HASH: felt252 = selector!("\"u256\"(\"low\":\"u128\",\"high\":\"u128\")");

// @dev There's no u8 in SNIP-12, we use u128
pub const _PERMIT_DETAILS_TYPE_HASH: felt252 = selector!(
    "\"PermitDetails\"(\"token\":\"ContractAddress\",\"amount\":\"u256\",\"expiration\":\"u128\",\"nonce\":\"u128\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const _PERMIT_SINGLE_TYPE_HASH: felt252 = selector!(
    "\"PermitSingle\"(\"details\":\"PermitDetails\",\"spender\":\"ContractAddress\",\"sig_deadline\":\"u256\")\"PermitDetails\"(\"token\":\"ContractAddress\",\"amount\":\"u256\",\"expiration\":\"u128\",\"nonce\":\"u128\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const _PERMIT_BATCH_TYPE_HASH: felt252 = selector!(
    "\"PermitBatch\"(\"details\":\"PermitDetails*\",\"spender\":\"ContractAddress\",\"sig_deadline\":\"u256\")\"PermitDetails\"(\"token\":\"ContractAddress\",\"amount\":\"u256\",\"expiration\":\"u128\",\"nonce\":\"u128\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const _TOKEN_PERMISSIONS_TYPE_HASH: felt252 = selector!(
    "\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const _PERMIT_TRANSFER_FROM_TYPE_HASH: felt252 = selector!(
    "\"PermitTransferFrom\"(\"permitted\":\"TokenPermissions\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\")\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub const _PERMIT_BATCH_TRANSFER_FROM_TYPE_HASH: felt252 = selector!(
    "\"PermitBatchTransferFrom\"(\"permitted\":\"TokenPermissions*\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\")\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")",
);

pub fn _TOKEN_PERMISSIONS_TYPE_STRING() -> ByteArray {
    "\"TokenPermissions\"(\"token\":\"ContractAddress\",\"amount\":\"u256\")\"u256\"(\"low\":\"u128\",\"high\":\"u128\")"
}

pub fn _PERMIT_WITNESS_TRANSFER_FROM_TYPE_HASH_STUB() -> ByteArray {
    "\"PermitWitnessTransferFrom\"(\"permitted\":\"TokenPermissions\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\","
}

pub fn _PERMIT_WITNESS_BATCH_TRANSFER_FROM_TYPE_HASH_STUB() -> ByteArray {
    "\"PermitWitnessBatchTransferFrom\"(\"permitted\":\"TokenPermissions*\",\"spender\":\"ContractAddress\",\"nonce\":\"felt\",\"deadline\":\"u256\","
}

pub fn _PERMIT_WITNESS_TRANSFER_FROM_TYPE_HASH(witness_type_string: ByteArray) -> felt252 {
    let stub = _PERMIT_WITNESS_TRANSFER_FROM_TYPE_HASH_STUB();
    selector(format!("{stub}{witness_type_string}"))
}

pub fn _PERMIT_BATCH_WITNESS_TRANSFER_FROM_TYPE_HASH(witness_type_string: ByteArray) -> felt252 {
    let stub = _PERMIT_WITNESS_BATCH_TRANSFER_FROM_TYPE_HASH_STUB();
    selector(format!("{stub}{witness_type_string}"))
}

/// HASHING STRUCTS ///

pub impl U256StructHash of StructHash<u256> {
    fn hash_struct(self: @u256) -> felt252 {
        PoseidonTrait::new().update_with(_U256_TYPE_HASH).update_with(*self).finalize()
    }
}

pub impl PermitSingleStructHash of StructHash<PermitSingle> {
    fn hash_struct(self: @PermitSingle) -> felt252 {
        PoseidonTrait::new()
            .update_with(_PERMIT_SINGLE_TYPE_HASH)
            .update_with(self.details.hash_struct())
            .update_with(*self.spender)
            .update_with(self.sig_deadline.hash_struct())
            .finalize()
    }
}

pub impl PermitBatchStructHash of StructHash<PermitBatch> {
    fn hash_struct(self: @PermitBatch) -> felt252 {
        let hashed_details = self
            .details
            .into_iter()
            .map(|detail| detail.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            .update_with(_PERMIT_BATCH_TYPE_HASH)
            .update_with(hashed_details)
            .update_with(*self.spender)
            .update_with(self.sig_deadline.hash_struct())
            .finalize()
    }
}

pub impl StructHashPermitTransferFrom of StructHash<PermitTransferFrom> {
    fn hash_struct(self: @PermitTransferFrom) -> felt252 {
        PoseidonTrait::new()
            .update_with(_PERMIT_TRANSFER_FROM_TYPE_HASH)
            .update_with(self.permitted.hash_struct())
            .update_with(get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .finalize()
    }
}

pub impl StructHashPermitBatchTransferFrom of StructHash<PermitBatchTransferFrom> {
    fn hash_struct(self: @PermitBatchTransferFrom) -> felt252 {
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            .update_with(_PERMIT_BATCH_TRANSFER_FROM_TYPE_HASH)
            .update_with(hashed_permissions)
            .update_with(get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .finalize()
    }
}

pub trait StructHashWitnessTrait<T> {
    fn hash_with_witness(self: @T, witness: felt252, witness_type_string: ByteArray) -> felt252;
}

pub impl StructHashWitnessPermitTransferFrom of StructHashWitnessTrait<PermitTransferFrom> {
    fn hash_with_witness(
        self: @PermitTransferFrom, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252 {
        PoseidonTrait::new()
            .update_with(_PERMIT_WITNESS_TRANSFER_FROM_TYPE_HASH(witness_type_string))
            .update_with(self.permitted.hash_struct())
            .update_with(get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}

pub impl StructHashWitnessPermitBatchTransferFrom of StructHashWitnessTrait<
    PermitBatchTransferFrom,
> {
    fn hash_with_witness(
        self: @PermitBatchTransferFrom, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252 {
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            .update_with(_PERMIT_BATCH_WITNESS_TRANSFER_FROM_TYPE_HASH(witness_type_string))
            .update_with(hashed_permissions)
            .update_with(get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}

pub impl PermitDetailsStructHash of StructHash<PermitDetails> {
    fn hash_struct(self: @PermitDetails) -> felt252 {
        PoseidonTrait::new()
            .update_with(_PERMIT_DETAILS_TYPE_HASH)
            .update_with(*self.token)
            .update_with(self.amount.hash_struct())
            .update_with(*self.expiration)
            .update_with(*self.nonce)
            .finalize()
    }
}

pub impl TokenPermissionsStructHash of StructHash<TokenPermissions> {
    fn hash_struct(self: @TokenPermissions) -> felt252 {
        PoseidonTrait::new()
            .update_with(_TOKEN_PERMISSIONS_TYPE_HASH)
            .update_with(*self.token)
            .update_with(self.amount.hash_struct())
            .finalize()
    }
}

/// OFFCHAIN MESSAGE HASHING WITH WITNESS ///

pub trait OffchainMessageHashWitnessTrait<T> {
    fn get_message_hash_with_witness(
        self: @T, signer: ContractAddress, witness: felt252, witness_type_string: ByteArray,
    ) -> felt252;
}

pub impl OffChainMessageHashWitnessPermitTransferFrom<
    impl metadata: SNIP12Metadata,
> of OffchainMessageHashWitnessTrait<PermitTransferFrom> {
    fn get_message_hash_with_witness(
        self: @PermitTransferFrom,
        signer: ContractAddress,
        witness: felt252,
        witness_type_string: ByteArray,
    ) -> felt252 {
        let domain = StarknetDomain {
            name: metadata::name(),
            version: metadata::version(),
            chain_id: get_tx_info().unbox().chain_id,
            revision: 1,
        };
        PoseidonTrait::new()
            // Domain
            .update_with('StarkNet Message')
            .update_with(domain.hash_struct())
            // Account
            .update_with(signer)
            // Message
            .update_with(_PERMIT_WITNESS_TRANSFER_FROM_TYPE_HASH(witness_type_string))
            .update_with(self.permitted.hash_struct())
            .update_with(get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}

pub impl OffChainMessageHashWitnessPermitBatchTransferFrom<
    impl metadata: SNIP12Metadata,
> of OffchainMessageHashWitnessTrait<PermitBatchTransferFrom> {
    fn get_message_hash_with_witness(
        self: @PermitBatchTransferFrom,
        signer: ContractAddress,
        witness: felt252,
        witness_type_string: ByteArray,
    ) -> felt252 {
        let domain = StarknetDomain {
            name: metadata::name(),
            version: metadata::version(),
            chain_id: get_tx_info().unbox().chain_id,
            revision: 1,
        };
        let hashed_permissions = self
            .permitted
            .into_iter()
            .map(|permission| permission.hash_struct())
            .collect::<Array<felt252>>()
            .span();

        PoseidonTrait::new()
            // Domain
            .update_with('StarkNet Message')
            .update_with(domain.hash_struct())
            // Account
            .update_with(signer)
            // Message
            .update_with(_PERMIT_BATCH_WITNESS_TRANSFER_FROM_TYPE_HASH(witness_type_string))
            .update_with(hashed_permissions)
            .update_with(get_caller_address())
            .update_with(*self.nonce)
            .update_with(self.deadline.hash_struct())
            .update_with(witness)
            .finalize()
    }
}
//pub impl OffChainMessageHashPermitTransferFrom of OffchainMessageHash<PermitTransferFrom> {
//    fn get_message_hash(self: @PermitTransferFrom, signer: ContractAddress) -> felt252 {
//        PoseidonTrait::new()
//            // Domain
//            .update_with('StarkNetDomain')
//            .update_with(DOMAIN())
//            // Account
//            .update_with(signer)
//            // Message
//            .update_with(_TOKEN_PERMISSIONS_TYPE_HASH)
//            .update_with(self.permitted.hash_struct())
//            .update_with(get_caller_address())
//            .update_with(*self.nonce)
//            .update_with(self.deadline.hash_struct())
//            .finalize()
//    }
//}
//
//pub impl OffChainMessageHashPermitBatchTransferFrom of OffchainMessageHash<
//    PermitBatchTransferFrom,
//> {
//    fn get_message_hash(self: @PermitBatchTransferFrom, signer: ContractAddress) -> felt252 {
//        let hashed_permissions = self
//            .permitted
//            .into_iter()
//            .map(|permission| permission.hash_struct())
//            .collect::<Array<felt252>>()
//            .span();
//
//        PoseidonTrait::new()
//            // Domain
//            .update_with('StarkNetDomain')
//            .update_with(DOMAIN())
//            // Account
//            .update_with(signer)
//            // Message
//            .update_with(_PERMIT_BATCH_TRANSFER_FROM_TYPEHASH)
//            .update_with(hashed_permissions)
//            .update_with(get_caller_address())
//            .update_with(*self.nonce)
//            .update_with(self.deadline.hash_struct())
//            .finalize()
//    }
//}
//



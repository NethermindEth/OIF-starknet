use alexandria_bytes::{Bytes, BytesTrait};
use core::num::traits::Zero;
use starknet::ContractAddress;

#[derive(Serde, Default, Drop)]
pub struct OrderData {
    pub sender: ContractAddress,
    pub recipient: ContractAddress,
    pub input_token: ContractAddress,
    pub output_token: ContractAddress,
    pub amount_in: u256,
    pub amount_out: u256,
    pub sender_nonce: felt252,
    pub origin_domain: u32,
    pub destination_domain: u32,
    pub destination_settler: ContractAddress,
    pub fill_deadline: u64,
    pub data: Bytes,
}

pub mod OrderEncoder {
    use alexandria_bytes::{Bytes, BytesTrait};
    use core::keccak::compute_keccak_byte_array;
    use super::OrderData;

    pub const ORDER_DATA_TYPE_HASH: felt252 = selector!(
        "\"Order Data\"(\"Sender\":\"ContractAddress\",\"Recipient\":\"ContractAddress\",\"Input Token\":\"ContractAddress\",\"Output Token\":\"ContractAddress\",\"Amount In\":\"u256\",\"Amount Out\":\"u256\",\"Sender Nonce\":\"felt\",\"Origin Domain\":\"u128\",\"Destination Domain\":\"u128\",\"Destination Settler\":\"ContractAddress\",\"Fill Deadline\":\"timestamp\",\"Data\":\"Bytes\")\"Bytes\"(\"Size\": \"u128\",\"Data\":\"u128*\")\"u256\"(\"low\": \"u128\",\"high\":\"u128*\")",
    );

    /// Typehash for OrderData struct
    pub fn order_data_type_hash() -> felt252 {
        ORDER_DATA_TYPE_HASH
    }

    /// Compute the ID of an OrderData struct
    pub fn id(order: @OrderData) -> u256 {
        compute_keccak_byte_array(@encode(order).into())
    }


    /// Encode an OrderData struct into Bytes
    pub fn encode(order: @OrderData) -> Bytes {
        let OrderData {
            sender,
            recipient,
            input_token,
            output_token,
            amount_in,
            amount_out,
            sender_nonce,
            origin_domain,
            destination_domain,
            destination_settler,
            fill_deadline,
            data,
        } = order;

        let mut encoded: Bytes = BytesTrait::new_empty();
        encoded.append_address(*sender);
        encoded.append_address(*recipient);
        encoded.append_address(*input_token);
        encoded.append_address(*output_token);
        encoded.append_u256(*amount_in);
        encoded.append_u256(*amount_out);
        encoded.append_felt252(*sender_nonce);
        encoded.append_u32(*origin_domain); //u64
        encoded.append_u32(*destination_domain); //u64
        encoded.append_address(*destination_settler);
        encoded.append_u64(*fill_deadline);
        encoded.concat(data);

        encoded
    }

    /// Decode OrderData struct from Bytes
    pub fn decode(order_data: @Bytes) -> OrderData {
        let (offset, sender) = order_data.read_address(0);
        let (offset, recipient) = order_data.read_address(offset);
        let (offset, input_token) = order_data.read_address(offset);
        let (offset, output_token) = order_data.read_address(offset);
        let (offset, amount_in) = order_data.read_u256(offset);
        let (offset, amount_out) = order_data.read_u256(offset);
        let (offset, sender_nonce) = order_data.read_felt252(offset);
        let (offset, origin_domain) = order_data.read_u32(offset);
        let (offset, destination_domain) = order_data.read_u32(offset);
        let (offset, destination_settler) = order_data.read_address(offset);
        let (offset, fill_deadline) = order_data.read_u64(offset);
        let (offset, data_size) = order_data.read_usize(offset);
        let (_, data) = order_data.read_bytes(offset, data_size);

        OrderData {
            sender,
            recipient,
            input_token,
            output_token,
            amount_in,
            amount_out,
            sender_nonce,
            origin_domain,
            destination_domain,
            destination_settler,
            fill_deadline,
            data: data,
        }
    }
}

pub trait OpenOrderEncoder<T> {
    fn encode(self: T) -> Bytes;
    fn decode(self: Bytes) -> T;
}

pub impl OpenOrderEncoderImplAt of OpenOrderEncoder<(felt252, @Bytes)> {
    /// Encodes an order_data_type and @order_data into Bytes
    fn encode(self: (felt252, @Bytes)) -> Bytes {
        let (order_data_type, data) = self;
        let mut encoded = BytesTrait::new_empty();

        encoded.append_felt252(order_data_type);
        encoded.append_usize(data.size());
        encoded.concat(data);

        encoded
    }

    /// Decodes an order_data_type and @order_data from Bytes
    fn decode(self: Bytes) -> (felt252, @Bytes) {
        let (offset, order_data_type) = self.read_felt252(0);
        let (offset, data_size) = self.read_usize(offset);
        let (_, data) = self.read_bytes(offset, data_size);

        (order_data_type, @data)
    }
}

pub impl OpenOrderEncoderImpl of OpenOrderEncoder<(felt252, Bytes)> {
    fn encode(self: (felt252, Bytes)) -> Bytes {
        /// Encodes an order_data_type and order_data into Bytes
        let (order_data_type, data) = self;
        let mut encoded = BytesTrait::new_empty();

        encoded.append_felt252(order_data_type);
        encoded.append_usize(data.size());
        encoded.concat(@data);

        encoded
    }

    /// Decodes an order_data_type and order_data from Bytes
    fn decode(self: Bytes) -> (felt252, Bytes) {
        let (offset, order_data_type) = self.read_felt252(0);
        let (offset, data_size) = self.read_usize(offset);
        let (_, data) = self.read_bytes(offset, data_size);

        (order_data_type, data)
    }
}

/// Sets the default value of `ContractAddress` to zero.
pub impl ContractAddressDefault of Default<ContractAddress> {
    fn default() -> ContractAddress {
        Zero::zero()
    }
}

/// Sets the default value of `Bytes` to zero.
pub impl BytesDefault of Default<Bytes> {
    fn default() -> Bytes {
        BytesTrait::new_empty()
    }
}


pub mod Hyperlane7683Message {
    use alexandria_bytes::{Bytes, BytesTrait};

    /// Returns formatted Router7683 message
    /// @dev This function should only be used in memory message construction.
    ///
    /// Parameters:
    /// - `settle`: Flag to indicate if the message is a settlement or refund
    /// - `order_ids`: The orderIds to settle or refund
    /// - `orders_filler_data`: Each element should contain the bytes32 encoded address of the
    /// settlement receiver.
    ///
    /// Returns: Formatted message body
    pub fn encode(settle: bool, order_ids: Span<u256>, orders_filler_data: Span<Bytes>) -> Bytes {
        let mut encoded: Bytes = BytesTrait::new_empty();

        match settle {
            true => encoded.append_u8(1),
            false => encoded.append_u8(0),
        };

        encoded.append_usize(order_ids.len());
        for order_id in order_ids {
            encoded.append_u256(*order_id);
        };

        encoded.append_usize(orders_filler_data.len());
        for order_filler_data in orders_filler_data {
            encoded.append_usize(order_filler_data.size());
            encoded.concat(order_filler_data);
        };

        encoded
    }

    /// Parses and returns the calls from the provided message
    ///
    /// Parameters:
    /// - `message`: The interchain message
    ///
    /// Returns The array of calls
    pub fn decode(message: Bytes) -> (bool, Span<u256>, Span<Bytes>) {
        // Settle
        let (offset, settle) = message.read_u8(0);

        // Order IDs
        let mut order_ids = array![];
        let (mut offset, order_ids_size) = message.read_usize(offset);
        for _ in 0..order_ids_size {
            let (_offset, order_id) = message.read_u256(offset);
            order_ids.append(order_id);

            offset = _offset;
        };

        // Orders Filler Data
        let mut orders_filler_data = array![];
        let (mut offset, orders_filler_data_len) = message.read_usize(offset);
        for _ in 0..orders_filler_data_len {
            let (_offset, filler_data_size) = message.read_usize(offset);
            let (_offset, filler_data) = message.read_bytes(_offset, filler_data_size);
            orders_filler_data.append(filler_data);

            offset = _offset;
        };

        (settle == 1, order_ids.span(), orders_filler_data.span())
    }

    pub fn encode_settle(order_ids: Span<u256>, orders_filler_data: Span<Bytes>) -> Bytes {
        encode(true, order_ids, orders_filler_data)
    }

    pub fn encode_refund(order_ids: Span<u256>) -> Bytes {
        encode(false, order_ids, array![BytesTrait::new_empty()].span())
    }
}

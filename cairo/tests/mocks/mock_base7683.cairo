#[starknet::interface]
pub trait IMockBase7683<TState> {
    fn set_native(ref self: TState, is_native: bool);
    fn set_counter_part_ba(ref self: TState, counter_part: starknet::ContractAddress);
    fn local_domain(self: @TState) -> u32;
    fn counter_part_ba(self: @TState) -> starknet::ContractAddress;
}

#[starknet::contract]
pub mod MockBase7683 {
    use alexandria_bytes::{Bytes, BytesStore, BytesTrait};
    use core::keccak::compute_keccak_byte_array;
    use core::num::traits::Bounded;
    use oif_starknet::base7683::Base7683Component;
    use oif_starknet::base7683::Base7683Component::{DestinationSettler, OriginSettler};
    use oif_starknet::basic_swap7683::BasicSwap7683Component;
    use oif_starknet::erc7683::interface::{
        FillInstruction, GaslessCrossChainOrder, OnchainCrossChainOrder, Output,
        ResolvedCrossChainOrder,
    };
    use openzeppelin_utils::cryptography::snip12::StructHashStarknetDomainImpl;
    use starknet::ContractAddress;
    use starknet::storage::{
        Map, MutableVecTrait, StoragePathEntry, StoragePointerReadAccess, StoragePointerWriteAccess,
        Vec, VecTrait,
    };

    /// COMPONENT INJECTION ///
    component!(path: Base7683Component, storage: base7683, event: Base7683Event);
    component!(path: BasicSwap7683Component, storage: basic_swap7683, event: BasicSwap7683Event);

    /// EXTERNAL ///

    /// Base7683
    #[abi(embed_v0)]
    pub impl OriginSettlerImpl =
        Base7683Component::OriginSettlerImpl<ContractState>;
    #[abi(embed_v0)]
    impl DestinationSettlerImpl =
        Base7683Component::DestinationSettlerImpl<ContractState>;
    #[abi(embed_v0)]
    pub impl ExtraImpl = Base7683Component::ERC7683ExtraImpl<ContractState>;

    impl BaseInternalImpl = Base7683Component::InternalImpl<ContractState>;

    /// BasicSwap7683
    impl BasicSwapInternalImpl = BasicSwap7683Component::InternalImpl<ContractState>;

    /// STORAGE ///
    #[storage]
    pub struct Storage {
        native: bool,
        counter_part_ba: ContractAddress,
        input_token: ContractAddress,
        output_token: ContractAddress,
        counterpart_ba: ContractAddress,
        origin: u32,
        destination: u32,
        filled_id: u256,
        filled_origin_data: Bytes,
        filled_filler_data: Bytes,
        settled_order_ids: Map<usize, u256>,
        settled_orders_origin_data: Map<usize, Bytes>,
        settled_orders_filler_data: Map<usize, Bytes>,
        settled_order_ids_len: usize,
        settled_orders_origin_data_len: usize,
        settled_orders_filler_data_len: usize,
        refunded_order_ids: Map<usize, u256>,
        refunded_order_ids_len: usize,
        /// COMPONENT STORAGE ///
        #[substorage(v0)]
        base7683: Base7683Component::Storage,
        #[substorage(v0)]
        basic_swap7683: BasicSwap7683Component::Storage,
    }

    /// CONSTRUCTOR ///
    #[constructor]
    fn constructor(
        ref self: ContractState,
        permit2: ContractAddress,
        local: u32,
        remote: u32,
        input_token: ContractAddress,
        output_token: ContractAddress,
    ) {
        self.base7683._initialize(permit2);
        self.origin.write(local);
        self.destination.write(remote);
        self.input_token.write(input_token);
        self.output_token.write(output_token);
    }

    /// EVENTS ///
    #[event]
    #[derive(Drop, starknet::Event)]
    pub enum Event {
        #[flat]
        Base7683Event: Base7683Component::Event,
        #[flat]
        BasicSwap7683Event: BasicSwap7683Component::Event,
    }

    /// EXTRA PUBLIC ///
    #[abi(embed_v0)]
    pub impl MockBase7683Impl of super::IMockBase7683<ContractState> {
        fn counter_part_ba(self: @ContractState) -> ContractAddress {
            self.counter_part_ba.read()
        }

        fn set_native(ref self: ContractState, is_native: bool) {
            self.set_native(is_native);
        }

        fn set_counter_part_ba(ref self: ContractState, counter_part: ContractAddress) {
            self.counter_part_ba.write(counter_part);
        }

        fn local_domain(self: @ContractState) -> u32 {
            self.origin.read()
        }
    }

    /// INTERNAL ///
    #[generate_trait]
    pub impl InternalImpl of InternalTrait {
        fn __resolved_order(
            self: @Base7683Component::ComponentState<ContractState>,
            sender: ContractAddress,
            open_deadline: u64,
            fill_deadline: u64,
            order_data: Bytes,
        ) -> (ResolvedCrossChainOrder, u256, felt252) {
            let self = self.get_contract();

            let max_spent = array![
                Output {
                    token: self.output_token.read(),
                    amount: 100,
                    recipient: self.counter_part_ba.read(),
                    chain_id: self.destination.read(),
                },
            ];

            let min_received = array![
                Output {
                    token: self.input_token.read(),
                    amount: 100,
                    recipient: 0.try_into().unwrap(),
                    chain_id: self.origin.read(),
                },
            ];

            let fill_instructions = array![
                FillInstruction {
                    destination_chain_id: self.destination.read(),
                    destination_settler: self.counter_part_ba.read(),
                    origin_data: order_data,
                },
            ];

            let order_id: u256 = 123456789;

            (
                ResolvedCrossChainOrder {
                    user: sender,
                    origin_chain_id: self.origin.read(),
                    open_deadline,
                    fill_deadline,
                    order_id,
                    min_received,
                    max_spent,
                    fill_instructions,
                },
                order_id,
                1,
            )
        }
    }

    /// BASE OVERRIDES ///
    pub impl Base7686VirtualImpl of Base7683Component::Virtual<ContractState> {
        fn _fill_order(
            ref self: Base7683Component::ComponentState<ContractState>,
            order_id: u256,
            origin_data: @Bytes,
            filler_data: @Bytes,
        ) {
            let mut self = self.get_contract_mut();
            self.filled_id.write(order_id);
            self.filled_origin_data.write(origin_data.clone());
            self.filled_filler_data.write(filler_data.clone());
        }

        fn _resolve_onchain_order(
            self: @Base7683Component::ComponentState<ContractState>, order: @OnchainCrossChainOrder,
        ) -> (ResolvedCrossChainOrder, u256, felt252) {
            self
                .__resolved_order(
                    starknet::get_caller_address(),
                    Bounded::<u64>::MAX,
                    *order.fill_deadline,
                    order.order_data.clone(),
                )
        }

        fn _resolve_gasless_order(
            self: @Base7683Component::ComponentState<ContractState>,
            order: @GaslessCrossChainOrder,
            origin_filler_data: @Bytes,
        ) -> (ResolvedCrossChainOrder, u256, felt252) {
            self
                .__resolved_order(
                    *order.user,
                    *order.open_deadline,
                    *order.fill_deadline,
                    order.order_data.clone(),
                )
        }

        fn _settle_orders(
            ref self: Base7683Component::ComponentState<ContractState>,
            order_ids: @Array<u256>,
            orders_origin_data: @Array<Bytes>,
            orders_filler_data: @Array<Bytes>,
            value: u256,
        ) {
            let mut self = self.get_contract_mut();

            self.settled_order_ids_len.write(order_ids.len());
            self.settled_orders_origin_data_len.write(orders_origin_data.len());
            self.settled_orders_filler_data_len.write(orders_filler_data.len());

            for i in 0..order_ids.len() {
                self.settled_order_ids.entry(i).write(*order_ids[i]);
            };

            for i in 0..orders_origin_data.len() {
                self.settled_orders_origin_data.entry(i).write(orders_origin_data[i].clone());
            };

            for i in 0..orders_filler_data.len() {
                self.settled_orders_filler_data.entry(i).write(orders_filler_data[i].clone());
            };
        }


        fn _refund_onchain_orders(
            ref self: Base7683Component::ComponentState<ContractState>,
            orders: @Array<OnchainCrossChainOrder>,
            order_ids: @Array<u256>,
            value: u256,
        ) {
            let mut self = self.get_contract_mut();

            self.refunded_order_ids_len.write(order_ids.len());

            for i in 0..order_ids.len() {
                self.refunded_order_ids.entry(i).write(*order_ids[i]);
            };
        }

        fn _refund_gasless_orders(
            ref self: Base7683Component::ComponentState<ContractState>,
            orders: @Array<GaslessCrossChainOrder>,
            order_ids: @Array<u256>,
            value: u256,
        ) {
            let mut contract_state = self.get_contract_mut();
            BasicSwapInternalImpl::_refund_gasless_orders(
                ref contract_state.basic_swap7683, orders, order_ids, value,
            );
        }

        fn _local_domain(self: @Base7683Component::ComponentState<ContractState>) -> u32 {
            // // MailboxClientImpl::get_local_domain(self.get_contract())
            self.get_contract().origin.read()
        }

        fn _get_gasless_order_id(
            self: @Base7683Component::ComponentState<ContractState>, order: @GaslessCrossChainOrder,
        ) -> u256 {
            compute_keccak_byte_array(@Into::<Bytes, ByteArray>::into(order.order_data.clone()))
        }

        fn _get_onchain_order_id(
            self: @Base7683Component::ComponentState<ContractState>, order: @OnchainCrossChainOrder,
        ) -> u256 {
            compute_keccak_byte_array(@Into::<Bytes, ByteArray>::into(order.order_data.clone()))
        }
    }

    /// BASIC SWAP OVERRIDES ///
    impl BasicSwapVirtualImpl of BasicSwap7683Component::BasicSwapVirtual<ContractState> {
        /// Dispatches a settlement message to the specified domain.
        /// @dev Encodes the settle message using Hyperlane7683Message and dispatches it via the
        /// GasRouter.
        ///
        /// Parameters:
        /// - `origin_domain`: The domain to which the settlement message is sent.
        /// - `order_ids`: The IDs of the orders to settle.
        /// - `orders_filler_data`: The filler data for the orders.
        fn _dispatch_settle(
            ref self: BasicSwap7683Component::ComponentState<ContractState>,
            origin_domain: u32,
            order_ids: @Array<u256>,
            orders_filler_data: @Array<Bytes>,
            value: u256,
        ) { //
        // // let mut contract_state = self.get_contract_mut();
        // // contract_state
        // //     .gas_router
        // //     ._Gas_router_dispatch(
        // //         origin_domain.try_into().unwrap(),
        // //         value,
        // //         Hyperlane7683Message::encode_settle(
        // //             order_ids.span(), orders_filler_data.span(),
        // //         ),
        // //         contract_state.mailbox_client.get_hook(),
        // //     );
        }

        /// Dispatches a refund message to the specified domain.
        /// @dev Encodes the refund message using Hyperlane7683Message and dispatches it via the
        /// GasRouter.
        ///
        /// Parameters:
        /// - `origin_domain`: The domain to which the refund message is sent.
        /// - `order_dds`: The IDs of the orders to refund.
        fn _dispatch_refund(
            ref self: BasicSwap7683Component::ComponentState<ContractState>,
            origin_domain: u32,
            order_ids: @Array<u256>,
            value: u256,
        ) { //
        // // let mut contract_state = self.get_contract_mut();
        // // contract_state
        // //     .gas_router
        // //     ._Gas_router_dispatch(
        // //         origin_domain.try_into().unwrap(),
        // //         value,
        // //         Hyperlane7683Message::encode_refund(order_ids.span()),
        // //         contract_state.mailbox_client.get_hook(),
        // //     );
        }

        fn _handle_settle_order(
            ref self: BasicSwap7683Component::ComponentState<ContractState>,
            message_origin: u32,
            message_sender: ContractAddress,
            order_id: u256,
            receiver: ContractAddress,
        ) {
            let mut contract_state = self.get_contract_mut();

            BasicSwapInternalImpl::_handle_settle_order(
                ref contract_state.basic_swap7683,
                message_origin,
                message_sender,
                order_id,
                receiver,
            );
        }

        fn _handle_refund_order(
            ref self: BasicSwap7683Component::ComponentState<ContractState>,
            message_origin: u32,
            message_sender: ContractAddress,
            order_id: u256,
        ) {
            let mut contract_state = self.get_contract_mut();
            BasicSwapInternalImpl::_handle_refund_order(
                ref contract_state.basic_swap7683, message_origin, message_sender, order_id,
            );
        }
    }
}

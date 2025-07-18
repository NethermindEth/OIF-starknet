use starknet::ContractAddress;

/// Signals that an order has been opened
/// @param order_id: A unique order identifier within this settlement system
/// @param resolved_order: Resolved order that would be returned by resolve if called instead of
/// Open
#[derive(Drop, starknet::Event)]
pub struct Open {
    pub order_id: felt252,
    pub resolved_order: ResolvedCrossChainOrder,
}

/// Standard interface for settlement contracts on the origin chain
#[starknet::interface]
pub trait IOriginSettler<TState> {
    /// Opens a gasless cross-chain order on behalf of a user.
    /// @dev To be called by the filler.
    /// @dev This method must emit the Open event
    ///
    /// Parameters:
    /// - `order`: The GaslessCrossChainOrder definition
    /// - `signature`: The user's signature over the order
    /// - `origin_filler_data`: Any filler-defined data required by the settler
    fn open_for(
        ref self: TState,
        order: GaslessCrossChainOrder,
        signature: Array<felt252>,
        origin_filler_data: ByteArray,
    );

    /// Opens a cross-chain order
    /// @dev To be called by the user
    /// @dev This method must emit the Open event
    ///
    /// Parameter:
    /// - `order`: The OnchainCrossChainOrder definition
    fn open(ref self: TState, order: OnchainCrossChainOrder);

    /// Resolves a specific GaslessCrossChainOrder into a generic ResolvedCrossChainOrder
    /// @dev Intended to improve standardized integration of various order types and settlement
    /// contracts
    ///
    /// Parameters:
    /// - ``order` The GaslessCrossChainOrder definition
    /// - `origin_filler_data` Any filler-defined data required by the settler
    ///
    /// Returns: ResolvedCrossChainOrder hydrated order data including the inputs and outputs of the
    /// order
    fn resolve_for(
        self: @TState, order: GaslessCrossChainOrder, origin_filler_data: ByteArray,
    ) -> ResolvedCrossChainOrder;

    /// Resolves a specific OnchainCrossChainOrder into a generic ResolvedCrossChainOrder
    /// @dev Intended to improve standardized integration of various order types and settlement
    /// contracts
    ///
    /// Parameters:
    /// - `order`: The OnchainCrossChainOrder definition
    ///
    /// Returns: ResolvedCrossChainOrder hydrated order data including the inputs and outputs of the
    /// order
    fn resolve(self: @TState, order: OnchainCrossChainOrder) -> ResolvedCrossChainOrder;
}

/// Standard interface for settlement contracts on the destination chain
#[starknet::interface]
pub trait IDestinationSettler<TState> {
    /// Fills a single leg of a particular order on the destination chain
    ///
    /// Parameters
    /// - `order_id`: Unique order identifier for this order
    /// - `origin_data`: Data emitted on the origin to parameterize the fill
    /// - `filler_data`: Data provided by the filler to inform the fill or express their preferences
    fn fill(ref self: TState, order_id: felt252, origin_data: ByteArray, filler_data: ByteArray);
}

#[starknet::interface]
pub trait IERC7683Extra<TState> {
    /// READS ///

    fn witness_hash(self: @TState, resolved_order: ResolvedCrossChainOrder) -> felt252;
    fn used_nonces(self: @TState, user: ContractAddress, nonce: felt252) -> bool;
    fn open_orders(self: @TState, order_id: felt252) -> ByteArray;
    fn filled_orders(self: @TState, order_id: felt252) -> FilledOrder;
    fn order_status(self: @TState, order_id: felt252) -> felt252;

    // { //
    //
    // call hash_struct on order

    /// WRITES ///

    /// Settles a batch of filled orders on the chain where the orders were opened.
    /// @dev Pays the filler the amount locked when the orders were opened.
    /// The settled status should not be changed here but rather on the origin chain. To allow the
    /// filler to retry in case some error occurs. Ensuring the order is eligible for settling in
    /// the origin chain is the responsibility of the caller.
    ///
    /// Parameters:
    /// - `order_ids`: An array of IDs for the orders to settle.
    fn settle(ref self: TState, order_ids: Array<felt252>);

    /// Refunds a batch of expired GaslessCrossChainOrders on the chain where the orders were
    /// opened. The refunded status should not be changed here but rather on the origin chain. To
    /// allow the user to retry in case some error occurs. Ensuring the order is eligible for
    /// refunding in the origin chain is the responsibility of the caller.
    ///
    /// Parameters:
    /// - `orders`: An array of GaslessCrossChainOrders to refund.
    fn refund_gasless_cross_chain_order(ref self: TState, orders: Array<GaslessCrossChainOrder>);


    /// Refunds a batch of expired OnchainCrossChainOrder on the chain where the orders were opened.
    /// The refunded status should not be changed here but rather on the origin chain. To allow the
    /// user to retry in case some error occurs. Ensuring the order is eligible for refunding the
    /// origin chain is the responsibility of the caller.
    ///
    /// Parameters:
    /// - `orders`: An array of GaslessCrossChainOrders to refund.
    fn refund_onchain_cross_chain_order(ref self: TState, orders: Array<OnchainCrossChainOrder>);

    /// Invalidates a nonce for the user calling the function.
    ///
    /// Parameters:
    /// - `nonce`: The nonce to invalidate.
    fn invalidate_nonces(ref self: TState, nonce: felt252);

    /// Checks whether a given nonce is valid.
    ///
    /// Parameters
    /// - `from`: The address whose nonce validity is being checked.
    /// - `nonce`: The nonce to check.
    ///
    /// Returns: `true` if the nonce is valid, `false` otherwise.
    fn in_valid_nonce(self: @TState, from: ContractAddress, nonce: felt252) -> bool;
}

/// Standard order struct to be signed by users, disseminated to fillers, and submitted to origin
/// settler contracts by fillers
#[derive(Serde, Drop, Clone)]
pub struct GaslessCrossChainOrder {
    pub origin_settler: ContractAddress,
    pub user: ContractAddress,
    pub nonce: felt252, //u256,
    pub origin_chain_id: u256,
    pub open_deadline: u64, //u32,
    pub fill_deadline: u64, //u32,
    pub order_data_type: felt252,
    pub order_data: ByteArray,
}

/// Standard order struct for user-opened orders, where the user is the one submitting the order
/// creation transaction
#[derive(Serde, Drop, Clone)]
pub struct OnchainCrossChainOrder {
    pub fill_deadline: u64, //u32,
    pub order_data_type: felt252,
    pub order_data: ByteArray,
}

/// An implementation-generic representation of an order intended for filler consumption
/// @dev Defines all requirements for filling an order by unbundling the implementation-specific
/// orderData.
/// @dev Intended to improve integration generalization by allowing fillers to compute the exact
/// input and output information of any order
#[derive(Serde, Drop)]
pub struct ResolvedCrossChainOrder {
    pub user: ContractAddress,
    pub origin_chain_id: u256,
    pub open_deadline: u64,
    pub fill_deadline: u64,
    pub order_id: felt252,
    pub max_spent: Array<Output>,
    pub min_received: Array<Output>,
    pub fill_instructions: Array<FillInstruction>,
}

/// Tokens that must be received for a valid order fulfillment
#[derive(Serde, Clone, Drop)]
pub struct Output {
    pub token: ContractAddress,
    pub amount: u256,
    pub recipient: ContractAddress,
    pub chain_id: u256,
}

/// Instructions to parameterize each leg of the fill
#[derive(Serde, Drop)]
pub struct FillInstruction {
    pub destination_chain_id: u256,
    pub destination_settler: ContractAddress,
    pub origin_data: ByteArray,
}

/// Represents data for an order that has been filled.
#[derive(starknet::Store, Drop, Serde)]
pub struct FilledOrder {
    pub origin_data: ByteArray,
    pub filler_data: ByteArray,
}


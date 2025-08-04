use alexandria_bytes::{Bytes, BytesTrait, BytesStore};
use crate::common::{
    Account, deal, deal_multiple, deploy_permit2, deploy_eth, deploy_erc20,
    deploy_mock_basic_swap7683, generate_account,
};
use permit2::interfaces::permit2::{IPermit2Dispatcher, IPermit2DispatcherTrait};
use core::num::traits::{Bounded, Pow};
use core::keccak::compute_keccak_byte_array;
use snforge_std::signature::stark_curve::{
    StarkCurveKeyPairImpl, StarkCurveSignerImpl, StarkCurveVerifierImpl,
};
use permit2::snip12_utils::permits::{TokenPermissionsStructHash, U256StructHash};
use openzeppelin_utils::cryptography::snip12::{SNIP12HashSpanImpl, StructHash};
use oif_starknet::libraries::order_encoder::{OrderData, OrderEncoder};
use oif_starknet::base7683::{
    SpanFelt252StructHash, ArrayFelt252StructHash, Base7683Component, Base7683Component::Filled,
    Base7683Component::Settle, Base7683Component::Refund,
};
use oif_starknet::erc7683::interface::{
    Output, FilledOrder, ResolvedCrossChainOrder, GaslessCrossChainOrder, Open,
    Base7683ABIDispatcher, Base7683ABIDispatcherTrait,
};
use oif_starknet::basic_swap7683::BasicSwap7683Component;
use oif_starknet::libraries::order_encoder::{BytesDefault};
use openzeppelin_token::erc20::interface::{IERC20Dispatcher, IERC20DispatcherTrait};
use starknet::ContractAddress;
use snforge_std::{
    start_cheat_caller_address, start_cheat_caller_address_global,
    start_cheat_block_timestamp_global, stop_cheat_block_timestamp_global, EventSpyAssertionsTrait,
    stop_cheat_caller_address_global, stop_cheat_caller_address, spy_events, EventSpyTrait,
    EventsFilterTrait,
};
use crate::mocks::mock_base7683::{IMockBase7683Dispatcher, IMockBase7683DispatcherTrait};
use crate::mocks::mock_basic_swap7683::{
    IMockBasicSwap7683Dispatcher, IMockBasicSwap7683DispatcherTrait,
};
use crate::base_test::{
    BaseTestSetup, setup as base_setup, _prepare_gasless_order as __prepare_gasless_order,
    _balances, _assert_open_order, _get_signature, _prepare_onchain_order,
};

#[derive(Drop, Clone)]
pub struct BasicSwapTestSetup {
    pub base_full: Base7683ABIDispatcher,
    pub base_swap: IMockBasicSwap7683Dispatcher,
    pub permit2: ContractAddress,
    pub input_token: IERC20Dispatcher,
    pub output_token: IERC20Dispatcher,
    pub kaka: Account,
    pub karp: Account,
    pub veg: Account,
    pub counterpart: ContractAddress,
    pub origin: u32,
    pub destination: u32,
    pub amount: u256,
    pub DOMAIN_SEPARATOR: felt252,
    pub fork_id: u256,
    pub users: Array<ContractAddress>,
    pub wrong_msg_origin: u32,
    pub wrong_msg_sender: ContractAddress,
}

pub fn _assert_resolved_order(
    resolved_order: ResolvedCrossChainOrder,
    order_data: Bytes,
    user: ContractAddress,
    fill_deadline: u64,
    open_deadline: u64,
    to: ContractAddress,
    destination_settler: ContractAddress,
    origin_chain_id: u32,
    input_token: ContractAddress,
    output_token: ContractAddress,
    setup: BasicSwapTestSetup,
) {
    assert_eq!(resolved_order.max_spent.len(), 1);
    assert_eq!(*resolved_order.max_spent.at(0).token, output_token);
    assert_eq!(*resolved_order.max_spent.at(0).amount, setup.amount);
    assert_eq!(*resolved_order.max_spent.at(0).recipient, to);
    assert_eq!(*resolved_order.max_spent.at(0).chain_id, setup.destination);

    assert_eq!(resolved_order.min_received.len(), 1);
    assert_eq!(*resolved_order.min_received.at(0).token, input_token);
    assert_eq!(*resolved_order.min_received.at(0).amount, setup.amount);
    assert_eq!(*resolved_order.min_received.at(0).recipient, 0.try_into().unwrap());
    assert_eq!(*resolved_order.min_received.at(0).chain_id, setup.origin);

    assert_eq!(resolved_order.fill_instructions.len(), 1);
    assert_eq!(*resolved_order.fill_instructions.at(0).destination_chain_id, setup.destination);
    assert_eq!(*resolved_order.fill_instructions.at(0).destination_settler, destination_settler);
    assert(
        resolved_order.fill_instructions.at(0).origin_data == @order_data,
        'Fill instructions do not match',
    );

    assert_eq!(resolved_order.user, user);
    assert_eq!(resolved_order.origin_chain_id, origin_chain_id);
    assert_eq!(resolved_order.open_deadline, open_deadline);
    assert_eq!(resolved_order.fill_deadline, fill_deadline);
}

pub fn _balance_id(user: ContractAddress, setup: BasicSwapTestSetup) -> usize {
    let kaka = setup.kaka.account.contract_address;
    let karp = setup.karp.account.contract_address;
    let veg = setup.veg.account.contract_address;
    let counter_part = setup.counterpart;
    let base = setup.base_swap.contract_address;

    if user == kaka {
        0
    } else if user == karp {
        1
    } else if user == veg {
        2
    } else if user == counter_part {
        3
    } else if user == base {
        4
    } else {
        999999999
    }
}

pub fn setup() -> BasicSwapTestSetup {
    let permit2 = deploy_permit2();
    let _eth = deploy_eth();
    let input_token = deploy_erc20("Input Token", "IN");
    let output_token = deploy_erc20("Output Token", "OUT");
    let base_swap = deploy_mock_basic_swap7683(permit2);
    let base_full = Base7683ABIDispatcher { contract_address: base_swap.contract_address };

    let DOMAIN_SEPARATOR = IPermit2Dispatcher { contract_address: permit2 }.DOMAIN_SEPARATOR();
    let kaka = generate_account();
    let karp = generate_account();
    let veg = generate_account();
    let counterpart: ContractAddress = 'counterpart'.try_into().unwrap();
    let users = array![
        kaka.account.contract_address,
        karp.account.contract_address,
        veg.account.contract_address,
        counterpart,
        base_swap.contract_address,
    ];

    deal_multiple(
        array![
            input_token.contract_address,
            output_token.contract_address,
            _eth.contract_address,
            _eth.contract_address,
        ],
        array![
            kaka.account.contract_address,
            karp.account.contract_address,
            veg.account.contract_address,
        ],
        1_000_000 * 10_u256.pow(18),
    );

    start_cheat_block_timestamp_global(123456789);

    BasicSwapTestSetup {
        base_full,
        base_swap,
        permit2,
        input_token,
        output_token,
        kaka,
        karp,
        veg,
        counterpart,
        origin: 1,
        destination: 2,
        amount: 100,
        DOMAIN_SEPARATOR,
        fork_id: 0,
        users,
        wrong_msg_origin: 678.try_into().unwrap(),
        wrong_msg_sender: 'wrongMsgSender'.try_into().unwrap(),
    }
}


// fn setup() -> BasicSwapTestSetup {
//     let BasicSwapTestSetup {
//         base_full,
//         base_swap,
//         permit2,
//         input_token,
//         output_token,
//         kaka,
//         karp,
//         veg,
//         counterpart,
//         origin,
//         destination,
//         amount,
//         DOMAIN_SEPARATOR,
//         fork_id,
//         mut users,
//         wrong_msg_origin,
//         wrong_msg_sender,
//     } = basic_swap_setup();
//
//
//
//     BasicSwapTestSetup {
//         base_full,
//         base_swap,
//         permit2,
//         input_token,
//         output_token,
//         kaka,
//         karp,
//         veg,
//         counterpart,
//         origin,
//         destination,
//         amount,
//         DOMAIN_SEPARATOR,
//         fork_id,
//         users,
//         wrong_msg_origin,
//         wrong_msg_sender,
//     }
// }

fn prepare_order_data(setup: BasicSwapTestSetup) -> OrderData {
    OrderData {
        sender: setup.kaka.account.contract_address,
        recipient: setup.karp.account.contract_address,
        input_token: setup.input_token.contract_address,
        output_token: setup.output_token.contract_address,
        amount_in: setup.amount,
        amount_out: setup.amount,
        sender_nonce: 1,
        origin_domain: setup.origin,
        destination_domain: setup.destination,
        destination_settler: setup.counterpart,
        fill_deadline: starknet::get_block_timestamp() + 100,
        data: BytesTrait::new_empty(),
    }
}

fn _prepare_gasless_order(
    setup: BasicSwapTestSetup,
    order_data: Bytes,
    permit_nonce: felt252,
    open_deadline: u64,
    fill_deadline: u64,
) -> GaslessCrossChainOrder {
    __prepare_gasless_order(
        setup.base_swap.contract_address,
        setup.kaka.account.contract_address,
        setup.origin,
        order_data,
        permit_nonce,
        open_deadline,
        fill_deadline,
        OrderEncoder::order_data_type_hash(),
    )
}

#[test]
fn test__settle_orders_works() {
    let setup = setup();
    let order_data1 = prepare_order_data(setup.clone());
    let mut order_data2 = prepare_order_data(setup.clone());
    order_data2.origin_domain = setup.destination;

    let order_ids: Array<u256> = array!['order1'.into(), 'order2'.into()];
    let orders_origin_data: Array<Bytes> = array![
        OrderEncoder::encode(@order_data1), OrderEncoder::encode(@order_data2),
    ];
    let orders_filler_data: Array<Bytes> = array![
        Into::<ByteArray, Bytes>::into("some filler data1"),
        Into::<ByteArray, Bytes>::into("some filler data2"),
    ];

    setup
        .base_swap
        .settle_orders(
            order_ids.clone(), orders_origin_data.clone(), orders_filler_data.clone(), 0,
        );

    assert_eq!(setup.base_swap.dispatched_origin_domain(), setup.origin);
    assert_eq!(setup.base_swap.dispatched_order_ids()[0], order_ids[0]);
    assert_eq!(setup.base_swap.dispatched_order_ids()[1], order_ids[1]);
    assert(
        setup.base_swap.dispatched_orders_filler_data()[0] == orders_filler_data[0],
        'Origin data mismatch 1',
    );
    assert(
        setup.base_swap.dispatched_orders_filler_data()[1] == orders_filler_data[1],
        'Origin data mismatch 2',
    );
}

#[test]
fn test__refund_orders_onchain_works() {
    let setup = setup();
    let order_data1 = prepare_order_data(setup.clone());
    let mut order_data2 = prepare_order_data(setup.clone());
    order_data2.origin_domain = setup.destination;

    let order1 = _prepare_onchain_order(
        OrderEncoder::encode(@order_data1),
        order_data1.fill_deadline,
        OrderEncoder::order_data_type_hash(),
    );

    let order2 = _prepare_onchain_order(
        OrderEncoder::encode(@order_data2),
        order_data2.fill_deadline,
        OrderEncoder::order_data_type_hash(),
    );

    let order_ids: Array<u256> = array!['order1'.into(), 'order2'.into()];
    let orders = array![order1.clone(), order2.clone()];
    setup.base_swap.refund_onchain_orders(orders.clone(), order_ids.clone(), 0);

    assert_eq!(setup.base_swap.dispatched_origin_domain(), setup.origin);
    assert_eq!(setup.base_swap.dispatched_order_ids()[0], order_ids[0]);
    assert_eq!(setup.base_swap.dispatched_order_ids()[1], order_ids[1]);
}

#[test]
fn test__refund_orders_gasless_works() {
    let setup = setup();
    let permit_nonce = 0;
    let order_data1 = prepare_order_data(setup.clone());
    let mut order_data2 = prepare_order_data(setup.clone());
    order_data2.origin_domain = setup.destination;

    let order1 = _prepare_gasless_order(
        setup.clone(), OrderEncoder::encode(@order_data1), permit_nonce, 0, 0,
    );

    let order2 = _prepare_gasless_order(
        setup.clone(), OrderEncoder::encode(@order_data2), permit_nonce, 0, 0,
    );

    let order_ids: Array<u256> = array!['order1'.into(), 'order2'.into()];
    let orders = array![order1.clone(), order2.clone()];

    setup.base_swap.refund_gasless_orders(orders.clone(), order_ids.clone(), 0);
    setup.base_swap.refund_gasless_orders(orders.clone(), order_ids.clone(), 0);

    assert_eq!(setup.base_swap.dispatched_origin_domain(), setup.origin);
    assert_eq!(setup.base_swap.dispatched_order_ids()[0], order_ids[0]);
    assert_eq!(setup.base_swap.dispatched_order_ids()[1], order_ids[1]);
}

#[test]
fn test__handle_settle_order_works() {
    let setup = setup();
    let order_data = prepare_order_data(setup.clone());
    let order_id = 'order1'.into();

    // Set order to opened
    setup.base_swap.set_order_opened(order_id, order_data);

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    let mut spy = spy_events();
    setup
        .base_swap
        .handle_settle_order(
            setup.destination, setup.counterpart, order_id, setup.karp.account.contract_address,
        );
    spy
        .assert_emitted(
            @array![
                (
                    setup.base_swap.contract_address,
                    BasicSwap7683Component::Event::Settled(
                        BasicSwap7683Component::Settled {
                            order_id: order_id.clone(),
                            receiver: setup.karp.account.contract_address,
                        },
                    ),
                ),
            ],
        );

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.SETTLED());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())]
            - setup.amount,
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())]
            + setup.amount,
    );
}

#[test]
fn test__handle_settle_order_not_OPENED() {
    let setup = setup();
    let order_id = 'order1'.into();
    // don't set the order as opened

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    setup
        .base_swap
        .handle_settle_order(
            setup.destination, setup.counterpart, order_id, setup.karp.account.contract_address,
        );

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.UNKNOWN());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())],
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())],
    );
}

#[test]
fn test__handle_settle_order_wrong_msg_origin() {
    let setup = setup();
    let order_id = 'order1'.into();
    // don't set the order as opened

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    setup
        .base_swap
        .handle_settle_order(
            setup.wrong_msg_origin,
            setup.counterpart,
            order_id,
            setup.karp.account.contract_address,
        );

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.UNKNOWN());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())],
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())],
    );
}

#[test]
fn test__handle_settle_order_wrong_msg_sender() {
    let setup = setup();
    let order_id = 'order1'.into();
    // don't set the order as opened

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    setup
        .base_swap
        .handle_settle_order(
            setup.destination,
            setup.wrong_msg_sender,
            order_id,
            setup.karp.account.contract_address,
        );

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.UNKNOWN());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())],
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())],
    );
}

#[test]
fn test__handle_refund_order_works() {
    let setup = setup();
    let order_data = prepare_order_data(setup.clone());
    let order_id = 'order1'.into();

    // set the order as opened
    setup.base_swap.set_order_opened(order_id, order_data);

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    let mut spy = spy_events();
    setup.base_swap.handle_refund_order(setup.destination, setup.counterpart, order_id);
    spy
        .assert_emitted(
            @array![
                (
                    setup.base_swap.contract_address,
                    BasicSwap7683Component::Event::Refunded(
                        BasicSwap7683Component::Refunded {
                            order_id: order_id.clone(),
                            receiver: setup.kaka.account.contract_address,
                        },
                    ),
                ),
            ],
        );

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.REFUNDED());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())]
            - setup.amount,
    );
    assert_eq!(
        *balances_after[_balance_id(setup.kaka.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.kaka.account.contract_address, setup.clone())]
            + setup.amount,
    );
}

#[test]
fn test__handle_refund_order_not_OPENED() {
    let setup = setup();
    let order_id = 'order1'.into();

    // don't set the order as opened

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    setup.base_swap.handle_refund_order(setup.destination, setup.counterpart, order_id);

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.UNKNOWN());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())],
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())],
    );
}

#[test]
fn test__handle_refund_order_wrong_msg_origin() {
    let setup = setup();
    let order_id = 'order1'.into();

    // don't set the order as opened

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    setup.base_swap.handle_refund_order(setup.wrong_msg_origin, setup.counterpart, order_id);

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.UNKNOWN());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())],
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())],
    );
}

#[test]
fn test__handle_refund_order_wrong_msg_sender() {
    let setup = setup();
    let order_id = 'order1'.into();

    // don't set the order as opened

    deal(setup.input_token.contract_address, setup.base_swap.contract_address, 1_000_000);
    let balances_before = _balances(setup.input_token, setup.users.clone());

    setup.base_swap.handle_refund_order(setup.destination, setup.wrong_msg_sender, order_id);

    let balances_after = _balances(setup.input_token, setup.users.clone());

    assert_eq!(setup.base_full.order_status(order_id), setup.base_full.UNKNOWN());
    assert_eq!(
        *balances_after[_balance_id(setup.base_swap.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.base_swap.contract_address, setup.clone())],
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())],
    );
}

#[test]
fn test__resolve_order_onchain_works() {
    let setup = setup();
    let order_data = prepare_order_data(setup.clone());
    let order = _prepare_onchain_order(
        OrderEncoder::encode(@order_data),
        order_data.fill_deadline,
        OrderEncoder::order_data_type_hash(),
    );

    start_cheat_caller_address(
        setup.base_swap.contract_address, setup.kaka.account.contract_address,
    );
    let (resolved_order, _, _) = setup.base_swap.resolve_onchain_order(order.clone());

    _assert_resolved_order(
        resolved_order,
        order.order_data,
        setup.kaka.account.contract_address,
        order_data.fill_deadline,
        Bounded::<u64>::MAX,
        setup.counterpart,
        setup.counterpart,
        1,
        setup.input_token.contract_address,
        setup.output_token.contract_address,
        setup.clone(),
    );
}

#[test]
fn test__resolve_order_gasless_works() {
    let setup = setup();
    let permit_nonce = 0;
    let order_data = prepare_order_data(setup.clone());
    let open_deadline = starknet::get_block_timestamp() + 10;
    let order = _prepare_gasless_order(
        setup.clone(),
        OrderEncoder::encode(@order_data),
        permit_nonce,
        open_deadline,
        order_data.fill_deadline,
    );

    let (resolved_order, _, _) = setup
        .base_swap
        .resolve_gasless_order(order.clone(), BytesTrait::new_empty());

    _assert_resolved_order(
        resolved_order,
        order.order_data,
        setup.kaka.account.contract_address,
        order_data.fill_deadline,
        open_deadline,
        setup.counterpart,
        setup.counterpart,
        1,
        setup.input_token.contract_address,
        setup.output_token.contract_address,
        setup.clone(),
    );
}

#[test]
#[should_panic(expected: "Invalid order type: 2422673239666974525189380806111333")]
fn test__resolve_order_INVALID_ORDER_TYPE() {
    let setup = setup();
    let wrong_order_type = 'wrongOrderType';
    let order_data = prepare_order_data(setup.clone());
    let order = _prepare_onchain_order(
        OrderEncoder::encode(@order_data), order_data.fill_deadline, wrong_order_type,
    );

    setup.base_swap.resolve_onchain_order(order);
}

#[test]
#[should_panic(expected: "Invalid origin domain: 0")]
fn test__resolve_order_INVALID_ORIGIN_DOMAIN() {
    let setup = setup();
    let mut order_data = prepare_order_data(setup.clone());
    order_data.origin_domain = 0;
    let order = _prepare_onchain_order(
        OrderEncoder::encode(@order_data),
        order_data.fill_deadline,
        OrderEncoder::order_data_type_hash(),
    );

    start_cheat_caller_address(
        setup.base_swap.contract_address, setup.kaka.account.contract_address,
    );
    setup.base_swap.resolve_onchain_order(order);
    stop_cheat_caller_address(setup.base_swap.contract_address);
}

#[test]
fn test__get_order_id_gasless_works() {
    let setup = setup();
    let order_data = prepare_order_data(setup.clone());

    let order = _prepare_gasless_order(setup.clone(), OrderEncoder::encode(@order_data), 0, 0, 0);

    assert_eq!(setup.base_swap.get_gasless_order_id(order), OrderEncoder::id(@order_data));
}

#[test]
fn test__get_order_id_onchain_works() {
    let setup = setup();
    let order_data = prepare_order_data(setup.clone());

    let order = _prepare_onchain_order(
        OrderEncoder::encode(@order_data),
        order_data.fill_deadline,
        OrderEncoder::order_data_type_hash(),
    );

    assert_eq!(setup.base_swap.get_onchain_order_id(order), OrderEncoder::id(@order_data));
}

#[test]
#[should_panic(expected: "Invalid order type: 2422673239666974525189380806111333")]
fn test__get_order_id_onchain_INVALID_ORDER_TYPE() {
    let setup = setup();
    let wrong_order_type = 'wrongOrderType';
    let order_data = prepare_order_data(setup.clone());

    let order = _prepare_onchain_order(
        OrderEncoder::encode(@order_data), order_data.fill_deadline, wrong_order_type,
    );

    setup.base_swap.get_onchain_order_id(order);
}


#[test]
fn test__fill_order_ERC20_works() {
    let setup = setup();
    let mut order_data = prepare_order_data(setup.clone());
    order_data.destination_domain = setup.origin;
    let order_id = OrderEncoder::id(@order_data);
    let origin_data = OrderEncoder::encode(@order_data);

    let balances_before = _balances(setup.output_token, setup.users.clone());

    start_cheat_caller_address(
        setup.output_token.contract_address, setup.kaka.account.contract_address,
    );
    setup.output_token.approve(setup.base_swap.contract_address, setup.amount);
    stop_cheat_caller_address(setup.output_token.contract_address);

    start_cheat_caller_address(
        setup.base_swap.contract_address, setup.kaka.account.contract_address,
    );
    setup.base_swap.fill_order(order_id, origin_data, BytesTrait::new_empty());
    stop_cheat_caller_address(setup.base_swap.contract_address);

    let balances_after = _balances(setup.output_token, setup.users.clone());

    assert_eq!(
        *balances_after[_balance_id(setup.kaka.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.kaka.account.contract_address, setup.clone())]
            - setup.amount,
    );
    assert_eq!(
        *balances_after[_balance_id(setup.karp.account.contract_address, setup.clone())],
        *balances_before[_balance_id(setup.karp.account.contract_address, setup.clone())]
            + setup.amount,
    );
}

#[test]
#[should_panic(expected: 'Invalid order ID')]
fn test__fill_order_INVALID_ORDER_ID() {
    let setup = setup();
    let mut order_data = prepare_order_data(setup.clone());
    order_data.destination_domain = setup.origin;
    let order_id = 'wrongId'.into();
    let origin_data = OrderEncoder::encode(@order_data);

    start_cheat_caller_address(
        setup.base_swap.contract_address, setup.kaka.account.contract_address,
    );
    setup.base_swap.fill_order(order_id, origin_data, BytesTrait::new_empty());
    stop_cheat_caller_address(setup.base_swap.contract_address);
}

#[test]
#[should_panic(expected: 'Order fill expired')]
fn test__fill_order_ORDER_FILL_EXPIRED() {
    let setup = setup();
    let mut order_data = prepare_order_data(setup.clone());
    order_data.fill_deadline = starknet::get_block_timestamp() - 1;
    order_data.destination_domain = setup.origin;
    let order_id = OrderEncoder::id(@order_data);
    let origin_data = OrderEncoder::encode(@order_data);

    start_cheat_caller_address(
        setup.base_swap.contract_address, setup.kaka.account.contract_address,
    );
    setup.base_swap.fill_order(order_id, origin_data, BytesTrait::new_empty());
    stop_cheat_caller_address(setup.base_swap.contract_address);
}

#[test]
#[should_panic(expected: 'Invalid order domain')]
fn test__fill_order_INVALID_ORDER_DOMAIN() {
    let setup = setup();
    let order_data = prepare_order_data(setup.clone());
    let order_id = OrderEncoder::id(@order_data);
    let origin_data = OrderEncoder::encode(@order_data);

    start_cheat_caller_address(
        setup.base_swap.contract_address, setup.kaka.account.contract_address,
    );
    setup.base_swap.fill_order(order_id, origin_data, BytesTrait::new_empty());
    stop_cheat_caller_address(setup.base_swap.contract_address);
}


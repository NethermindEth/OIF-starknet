use alexandria_bytes::{Bytes, BytesStore};
use core::num::traits::Bounded;
use snforge_std::signature::stark_curve::{
    StarkCurveKeyPairImpl, StarkCurveSignerImpl, StarkCurveVerifierImpl,
};
use crate::common::pop_event;
use permit2::snip12_utils::permits::{TokenPermissionsStructHash, U256StructHash};
use openzeppelin_utils::cryptography::snip12::{SNIP12HashSpanImpl, StructHash};
use oif_starknet::base7683::{SpanFelt252StructHash, ArrayFelt252StructHash};
use oif_starknet::erc7683::interface::{GaslessCrossChainOrder, Open, Base7683ABIDispatcherTrait};
use oif_starknet::libraries::order_encoder::{BytesDefault};
use openzeppelin_token::erc20::interface::{IERC20DispatcherTrait};
use snforge_std::{
    start_cheat_caller_address, start_cheat_caller_address_global, stop_cheat_caller_address_global,
    stop_cheat_caller_address, spy_events, EventSpyTrait,
};
use crate::mocks::mock_base7683::{IMockBase7683Dispatcher, IMockBase7683DispatcherTrait};
use crate::_base_test::{
    BaseTestSetup, setup as base_setup, _prepare_onchain_order, _balances, _assert_open_order,
    _assert_resolved_order, _get_signature,
};

pub fn setup() -> BaseTestSetup {
    let BaseTestSetup {
        _base7683,
        base,
        permit2,
        input_token,
        output_token,
        kaka,
        karp,
        veg,
        counter_part_addr,
        origin,
        destination,
        amount,
        DOMAIN_SEPARATOR,
        fork_id,
        mut users,
    } = base_setup();

    users.append(base.contract_address);

    BaseTestSetup {
        _base7683,
        base,
        permit2,
        input_token,
        output_token,
        kaka,
        karp,
        veg,
        counter_part_addr,
        origin,
        destination,
        amount,
        DOMAIN_SEPARATOR,
        fork_id,
        users,
    }
}

pub fn _prepare_gasless_order(
    order_data: Bytes,
    nonce: felt252,
    open_deadline: u64,
    fill_deadline: u64,
    order_data_type: felt252,
    setup: BaseTestSetup,
) -> GaslessCrossChainOrder {
    GaslessCrossChainOrder {
        origin_settler: setup.base.contract_address,
        user: setup.kaka.account.contract_address,
        origin_chain_id: setup.origin,
        order_data,
        nonce,
        open_deadline,
        fill_deadline,
        order_data_type,
    }
}


#[test]
#[fuzzer]
fn test_open_works(fill_deadline: u64) {
    let setup = setup();
    let order_data: Bytes = Into::<ByteArray, Bytes>::into("some order data");
    let order_type: felt252 = 'some order type';
    let order = _prepare_onchain_order(order_data.clone(), fill_deadline, order_type);

    start_cheat_caller_address(
        setup.input_token.contract_address, setup.kaka.account.contract_address,
    );
    setup.input_token.approve(setup.base.contract_address, setup.amount);
    stop_cheat_caller_address(setup.input_token.contract_address);

    start_cheat_caller_address(setup.base.contract_address, setup.kaka.account.contract_address);
    assert(
        setup._base7683.is_valid_nonce(setup.kaka.account.contract_address, 1),
        'Nonce is not valid',
    );
    let balances_before = _balances(setup.input_token, setup.users.clone().into());

    /// Open order and catch event
    let mut spy = spy_events();
    setup._base7683.open(order);
    let Open {
        order_id, resolved_order,
    } = pop_event::<Open>(setup.base.contract_address, selector!("Open"), spy.get_events().events);

    _assert_resolved_order(
        resolved_order,
        order_data.clone(),
        setup.kaka.account.contract_address,
        fill_deadline,
        Bounded::<u64>::MAX,
        setup.base.counter_part_ba(),
        setup.base.counter_part_ba(),
        setup.base.local_domain(),
        setup.input_token.contract_address,
        setup.output_token.contract_address,
        setup.clone(),
    );

    _assert_open_order(
        order_id,
        setup.kaka.account.contract_address,
        order_data,
        balances_before,
        setup.kaka.account.contract_address,
        setup.clone(),
    );

    stop_cheat_caller_address(setup.base.contract_address);
}


#[test]
#[fuzzer]
#[should_panic(expected: 'Invalid nonce')]
fn test_open_invalid_nonce(fill_deadline: u64) {
    let setup = setup();
    let order_data: Bytes = Into::<ByteArray, Bytes>::into("some order data");
    let order_type: felt252 = 'some order type';
    let order = _prepare_onchain_order(order_data.clone(), fill_deadline, order_type);

    start_cheat_caller_address(setup.base.contract_address, setup.kaka.account.contract_address);
    setup._base7683.invalidate_nonces(1);
    setup._base7683.open(order);
    stop_cheat_caller_address(setup.input_token.contract_address);
}

#[test]
#[fuzzer]
fn test_open_for_works(mut open_deadline: u64, fill_deadline: u64) {
    let setup = setup();
    if (open_deadline <= starknet::get_block_timestamp()) {
        open_deadline += starknet::get_block_timestamp();
    }

    start_cheat_caller_address(
        setup.input_token.contract_address, setup.kaka.account.contract_address,
    );
    setup.input_token.approve(setup.permit2, Bounded::<u256>::MAX);
    stop_cheat_caller_address(setup.input_token.contract_address);

    let nonce = 0;
    let order_data = Into::<ByteArray, Bytes>::into("some order data");
    let order_data_type = 'some order data';
    let order = _prepare_gasless_order(
        order_data.clone(), nonce, open_deadline, fill_deadline, order_data_type, setup.clone(),
    );
    let witness = setup
        ._base7683
        .witness_hash(setup._base7683.resolve_for(order.clone(), Default::default()));
    let sig = _get_signature(
        setup.kaka,
        setup.base.contract_address,
        witness,
        setup.input_token.contract_address,
        nonce,
        open_deadline,
        setup.clone(),
    );

    assert(
        setup._base7683.is_valid_nonce(setup.kaka.account.contract_address, nonce),
        'Nonce is not valid',
    );
    let balances_before = _balances(setup.input_token, setup.users.clone().into());

    /// Open order and catch event
    let mut spy = spy_events();
    start_cheat_caller_address(setup.base.contract_address, setup.karp.account.contract_address);
    setup._base7683.open_for(order, sig, Default::default());
    let Open {
        order_id, resolved_order,
    } = pop_event::<Open>(setup.base.contract_address, selector!("Open"), spy.get_events().events);

    _assert_resolved_order(
        resolved_order,
        order_data.clone(),
        setup.kaka.account.contract_address,
        fill_deadline,
        open_deadline,
        setup.base.counter_part_ba(),
        setup.base.counter_part_ba(),
        setup.base.local_domain(),
        setup.input_token.contract_address,
        setup.output_token.contract_address,
        setup.clone(),
    );

    _assert_open_order(
        order_id,
        setup.kaka.account.contract_address,
        order_data,
        balances_before,
        setup.kaka.account.contract_address,
        setup.clone(),
    );

    stop_cheat_caller_address(setup.base.contract_address);
}


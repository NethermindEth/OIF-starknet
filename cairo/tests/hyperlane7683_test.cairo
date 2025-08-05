use alexandria_bytes::{Bytes, BytesTrait, BytesStore};
use crate::common::{
    Account, deal, deal_multiple, deploy_permit2, deploy_eth, deploy_erc20,
    deploy_mock_hyperlane7683, generate_account,
};
use contracts::client::router_component::{IRouterDispatcher, IRouterDispatcherTrait};
use contracts::client::gas_router_component::{
    IGasRouterDispatcher, IGasRouterDispatcherTrait, GasRouterComponent::GasRouterConfig,
};
use snforge_std::{
    start_cheat_block_timestamp_global, start_cheat_caller_address_global,
    stop_cheat_caller_address_global, ContractClassTrait, DeclareResultTrait, declare,
};
use permit2::interfaces::permit2::{IPermit2Dispatcher, IPermit2DispatcherTrait};
use core::num::traits::{Bounded, Pow};
use core::keccak::compute_keccak_byte_array;
use snforge_std::signature::stark_curve::{
    StarkCurveKeyPairImpl, StarkCurveSignerImpl, StarkCurveVerifierImpl,
};
use permit2::snip12_utils::permits::{TokenPermissionsStructHash, U256StructHash};
use openzeppelin_utils::cryptography::snip12::{SNIP12HashSpanImpl, StructHash};
use oif_starknet::libraries::order_encoder::{OrderData, OrderEncoder, ContractAddressDefault};
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
use starknet::{ContractAddress, ClassHash};
use snforge_std::{
    start_cheat_caller_address, EventSpyAssertionsTrait, stop_cheat_caller_address, spy_events,
    EventSpyTrait,
};
use crate::mocks::mock_hyperlane_environment::{
    IMockHyperlaneEnvironmentDispatcher, IMockHyperlaneEnvironmentDispatcherTrait,
};
use crate::mocks::mock_base7683::{IMockBase7683Dispatcher, IMockBase7683DispatcherTrait};
use crate::mocks::mock_hyperlane7683::{
    IMockHyperlane7683Dispatcher, IMockHyperlane7683DispatcherTrait,
};
use crate::mocks::mock_basic_swap7683::{
    IMockBasicSwap7683Dispatcher, IMockBasicSwap7683DispatcherTrait,
};
use crate::base_test::{
    BaseTestSetup, setup as base_setup, _prepare_gasless_order as __prepare_gasless_order,
    _balances, _assert_open_order, _get_signature, _prepare_onchain_order,
};
use mocks::test_interchain_gas_payment::{
    ITestInterchainGasPaymentDispatcher, ITestInterchainGasPaymentDispatcherTrait,
};
use contracts::interfaces::{IMailboxClient, IMailboxDispatcher, IMailboxDispatcherTrait};
use crate::mocks::mock_mailbox::{MockMailbox};
use crate::mocks::test_interchain_gas_payment::{TestInterchainGasPayment};


const GAS_LIMIT: u256 = 60_000;

#[derive(Drop, Clone)]
pub struct HyperlaneTestSetup {
    pub environment: IMockHyperlaneEnvironmentDispatcher,
    pub igp: ITestInterchainGasPaymentDispatcher,
    pub origin_router: IMockHyperlane7683Dispatcher,
    pub destination_router: IMockHyperlane7683Dispatcher,
    ////
    pub origin_router_b32: u256,
    pub destination_router_b32: u256,
    pub destination_router_override_b32: u256,
    pub gas_payment_quote: u256,
    pub gas_payment_quote_override: u256,
    ////
    pub admin: ContractAddress,
    pub owner: ContractAddress,
    pub sender: ContractAddress,
    ////////
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
    setup: HyperlaneTestSetup,
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

pub fn _balance_id(user: ContractAddress, setup: HyperlaneTestSetup) -> usize {
    let kaka = setup.kaka.account.contract_address;
    let karp = setup.karp.account.contract_address;
    let veg = setup.veg.account.contract_address;
    let counter_part = setup.counterpart;
    let origin_router = setup.origin_router.contract_address;
    let destination_router = setup.destination_router.contract_address;
    let igp = setup.igp.contract_address;

    if user == kaka {
        0
    } else if user == karp {
        1
    } else if user == veg {
        2
    } else if user == counter_part {
        3
    } else if user == origin_router {
        4
    } else if user == destination_router {
        5
    } else if user == igp {
        6
    } else {
        999999999
    }
}

pub fn deploy_environment(
    origin: u32, destination: u32, mailbox_class_hash: ClassHash, ism_class_hash: ClassHash,
) -> IMockHyperlaneEnvironmentDispatcher {
    let contract = declare("MockHyperlaneEnvironment").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    origin.serialize(ref ctor_calldata);
    destination.serialize(ref ctor_calldata);
    mailbox_class_hash.serialize(ref ctor_calldata);
    ism_class_hash.serialize(ref ctor_calldata);

    let (contract_address, _) = contract.deploy(@ctor_calldata).expect('mock env failed');

    IMockHyperlaneEnvironmentDispatcher { contract_address }
}

pub fn deploy_igp() -> ITestInterchainGasPaymentDispatcher {
    let contract = declare("TestInterchainGasPayment").unwrap().contract_class();
    let (contract_address, _) = contract.deploy(@array![]).expect('mock igp env failed');

    ITestInterchainGasPaymentDispatcher { contract_address }
}

pub fn setup() -> HyperlaneTestSetup {
    let admin = 'admin'.try_into().unwrap();
    let owner = 'owner'.try_into().unwrap();
    let sender = 'sender'.try_into().unwrap();

    // Declare mock contracts
    start_cheat_caller_address_global(owner);
    let _contract = declare("MockMailbox").unwrap();
    let mock_mailbox_class_hash = _contract.contract_class().class_hash;
    let _contract = declare("TestInterchainGasPayment").unwrap();
    let mock_ism_class_hash = _contract.contract_class().class_hash;

    let origin = 1;
    let destination = 2;

    let environment = deploy_environment(
        origin, destination, *mock_mailbox_class_hash, *mock_ism_class_hash,
    );
    let igp = deploy_igp();
    let gas_payment_quote = igp.quote_gas_payment(GAS_LIMIT);

    let permit2 = deploy_permit2();
    let _eth = deploy_eth();
    let input_token = deploy_erc20("Input Token", "IN");
    let output_token = deploy_erc20("Output Token", "OUT");

    let origin_router = deploy_mock_hyperlane7683(
        permit2,
        environment.mailboxes(origin).contract_address,
        owner,
        igp.contract_address,
        environment.isms(origin).contract_address,
    );
    let destination_router = deploy_mock_hyperlane7683(
        permit2,
        environment.mailboxes(destination).contract_address,
        owner,
        igp.contract_address,
        environment.isms(destination).contract_address,
    );

    IMailboxDispatcher { contract_address: environment.mailboxes(origin).contract_address }
        .set_default_hook(igp.contract_address);
    IMailboxDispatcher { contract_address: environment.mailboxes(destination).contract_address }
        .set_default_hook(igp.contract_address);

    let origin_router_b32: u256 = Into::<
        felt252, u256,
    >::into(origin_router.contract_address.into());
    let destination_router_b32: u256 = Into::<
        felt252, u256,
    >::into(destination_router.contract_address.into());
    let destination_router_override_b32: u256 = Default::default();

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
        origin_router.contract_address,
        destination_router.contract_address,
        igp.contract_address,
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

    HyperlaneTestSetup {
        environment,
        igp,
        origin_router,
        destination_router,
        origin_router_b32,
        destination_router_b32,
        destination_router_override_b32,
        gas_payment_quote,
        gas_payment_quote_override: Default::default(),
        admin,
        owner,
        sender,
        permit2,
        input_token,
        output_token,
        kaka,
        karp,
        veg,
        counterpart,
        origin,
        destination,
        amount: 100,
        DOMAIN_SEPARATOR,
        fork_id: 0,
        users,
    }
}

fn enroll_routers(setup: HyperlaneTestSetup) {
    let origin_router_address = setup.origin_router.contract_address;
    let destination_router_address = setup.destination_router.contract_address;

    start_cheat_caller_address(origin_router_address, setup.owner);
    IRouterDispatcher { contract_address: origin_router_address }
        .enroll_remote_router(setup.destination, setup.destination_router_b32);
    IGasRouterDispatcher { contract_address: origin_router_address }
        .set_destination_gas(Option::None, Option::Some(setup.destination), Option::Some(60_000));
    stop_cheat_caller_address(origin_router_address);

    start_cheat_caller_address(destination_router_address, setup.admin);
    IRouterDispatcher { contract_address: destination_router_address }
        .enroll_remote_router(setup.origin, setup.origin_router_b32);
    IGasRouterDispatcher { contract_address: destination_router_address }
        .set_destination_gas(Option::None, Option::Some(setup.origin), Option::Some(60_000));
    stop_cheat_caller_address(destination_router_address);
}

#[test]
fn test_local_domain() {
    let setup = setup();

    assert_eq!(setup.origin_router.get_7683_local_domain(), setup.origin);
    assert_eq!(setup.destination_router.get_7683_local_domain(), setup.destination);
}

#[test]
#[fuzzer]
fn test_fuzz_enroll_remote_routers(mut count: u8, mut domain: u32, mut router: u256) { //
    let setup = setup();

    if (count.into() >= router) {
        router = count.into() + 1
    }
    if (count.into() >= domain) {
        domain = count.into() + 1
    }
    if (router == 0) {
        router = 1;
    }

    let mut domains: Array<u32> = array![];
    let mut routers: Array<u256> = array![];

    for i in 0..count {
        domains.append(domain - i.into());
        routers.append(router - i.into());
    };

    start_cheat_caller_address(setup.origin_router.contract_address, setup.owner);
    IRouterDispatcher { contract_address: setup.origin_router.contract_address }
        .enroll_remote_routers(domains.clone(), routers.clone());
    stop_cheat_caller_address(setup.origin_router.contract_address);

    let actual_domains = IRouterDispatcher {
        contract_address: setup.origin_router.contract_address,
    }
        .domains();

    assert_eq!(actual_domains.len(), domains.len());
    assert_eq!(
        IRouterDispatcher { contract_address: setup.origin_router.contract_address }.domains(),
        domains,
    );
    for i in 0..count {
        let actual_router = IRouterDispatcher {
            contract_address: setup.origin_router.contract_address,
        }
            .routers(*domains.at(i.into()));

        assert_eq!(@actual_router, routers.at(i.into()));
        assert_eq!(actual_domains.at(i.into()), domains.at(i.into()));
    }
}


use openzeppelin_account::interface::AccountABIDispatcher;
use openzeppelin_token::erc20::interface::{IERC20Dispatcher};
use snforge_std::signature::stark_curve::{
    StarkCurveKeyPairImpl, StarkCurveSignerImpl, StarkCurveVerifierImpl,
};
use snforge_std::signature::{KeyPair, KeyPairTrait};
use snforge_std::{Event, ContractClassTrait, DeclareResultTrait, declare};
use starknet::{ContractAddress};
use starknet::event::Event as _Event;
use crate::mocks::mock_base7683::{IMockBase7683Dispatcher};
use crate::mocks::mock_basic_swap7683::{IMockBasicSwap7683Dispatcher};
use crate::mocks::mock_hyperlane7683::{IMockHyperlane7683Dispatcher};
use crate::mocks::interfaces::{IMintableDispatcher, IMintableDispatcherTrait};

pub fn ETH_ADDRESS() -> ContractAddress {
    0x049D36570D4e46f48e99674bd3fcc84644DdD6b96F7C741B1562B82f9e004dC7.try_into().unwrap()
}

pub fn deploy_eth() -> IERC20Dispatcher {
    let mock_erc20_contract = declare("MockERC20").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    let name: ByteArray = "Ethereum";
    let symbol: ByteArray = "ETH";
    name.serialize(ref ctor_calldata);
    symbol.serialize(ref ctor_calldata);

    let (erc20_address, _) = mock_erc20_contract.deploy_at(@ctor_calldata, ETH_ADDRESS()).unwrap();
    IERC20Dispatcher { contract_address: erc20_address }
}

pub fn deploy_erc20(name: ByteArray, symbol: ByteArray) -> IERC20Dispatcher {
    let mock_erc20_contract = declare("MockERC20").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    name.serialize(ref ctor_calldata);
    symbol.serialize(ref ctor_calldata);

    let (erc20_address, _) = mock_erc20_contract.deploy(@ctor_calldata).unwrap();

    IERC20Dispatcher { contract_address: erc20_address }
}

pub fn deploy_permit2() -> ContractAddress {
    let mock_permit2_contract = declare("MockPermit2").unwrap().contract_class();
    let (mock_permit2_address, _) = mock_permit2_contract
        .deploy(@array![])
        .expect('mock permit2 deployment failed');

    mock_permit2_address
}

pub fn deploy_mock_base7683(
    permit2: ContractAddress,
    local: u32,
    remote: u32,
    input_token: ContractAddress,
    output_token: ContractAddress,
) -> IMockBase7683Dispatcher {
    let contract = declare("MockBase7683").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    permit2.serialize(ref ctor_calldata);
    local.serialize(ref ctor_calldata);
    remote.serialize(ref ctor_calldata);
    input_token.serialize(ref ctor_calldata);
    output_token.serialize(ref ctor_calldata);

    let (contract_address, _) = contract
        .deploy(@ctor_calldata)
        .expect('mock permit2 deployment failed');

    IMockBase7683Dispatcher { contract_address }
}

pub fn deploy_mock_basic_swap7683(permit2: ContractAddress) -> IMockBasicSwap7683Dispatcher {
    let contract = declare("MockBasicSwap7683").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    permit2.serialize(ref ctor_calldata);

    let (contract_address, _) = contract.deploy(@ctor_calldata).expect('mock basic swap failed');

    IMockBasicSwap7683Dispatcher { contract_address }
}

pub fn deploy_mock_hyperlane7683(
    permit2: ContractAddress,
    mailbox: ContractAddress,
    owner: ContractAddress,
    hook: ContractAddress,
    ism: ContractAddress,
) -> IMockHyperlane7683Dispatcher {
    let contract = declare("MockHyperlane7683").unwrap().contract_class();
    let mut ctor_calldata: Array<felt252> = array![];
    permit2.serialize(ref ctor_calldata);
    mailbox.serialize(ref ctor_calldata);
    owner.serialize(ref ctor_calldata);
    hook.serialize(ref ctor_calldata);
    ism.serialize(ref ctor_calldata);

    let (contract_address, _) = contract.deploy(@ctor_calldata).expect('mock hyperlane failed');

    IMockHyperlane7683Dispatcher { contract_address }
}


pub fn deal(token: ContractAddress, to: ContractAddress, amount: u256) {
    IMintableDispatcher { contract_address: token }.mint(to, amount);
}

pub fn deal_multiple(tokens: Array<ContractAddress>, tos: Array<ContractAddress>, amount: u256) {
    for token in tokens.span() {
        for to in tos.span() {
            deal(*token, *to, amount);
        };
    };
}


/// Accounts ///

#[derive(Drop, Copy)]
pub struct Account {
    pub account: AccountABIDispatcher,
    pub key_pair: KeyPair<felt252, felt252>,
}

pub fn generate_account() -> Account {
    let mock_account_contract = declare("MockAccount").unwrap().contract_class();
    let key_pair = KeyPairTrait::<felt252, felt252>::generate();
    let (account_address, _) = mock_account_contract.deploy(@array![key_pair.public_key]).unwrap();
    let account = AccountABIDispatcher { contract_address: account_address };
    Account { account, key_pair }
}

/// Utils ///
pub fn pop_event<T, +Drop<T>, +Default<T>, +PartialEq<T>, impl TEvent: starknet::Event<T>>(
    target: ContractAddress, selector: felt252, events: Array<(ContractAddress, Event)>,
) -> T {
    let mut popped: T = Default::default();
    for (source, e) in events {
        if (source == target) {
            let Event { mut keys, mut data } = e;
            if (*keys[0] == selector) {
                let _ = keys.pop_front();
                let mut keys = keys.span();
                let mut data = data.span();

                popped = _Event::<T>::deserialize(ref keys, ref data)
                    .expect('Failed to build event');
            }
        }
    };

    if (popped == Default::default()) {
        panic!("Failed to find event")
    }

    popped
}

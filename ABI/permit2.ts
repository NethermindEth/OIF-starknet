export const ABI = [
  {
    "type": "impl",
    "name": "AllowedTransferImpl",
    "interface_name": "oif_starknet::permit2::allowance_transfer::interface::IAllowanceTransfer"
  },
  {
    "type": "struct",
    "name": "core::integer::u256",
    "members": [
      {
        "name": "low",
        "type": "core::integer::u128"
      },
      {
        "name": "high",
        "type": "core::integer::u128"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::allowance_transfer::interface::PermitDetails",
    "members": [
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "amount",
        "type": "core::integer::u256"
      },
      {
        "name": "expiration",
        "type": "core::integer::u64"
      },
      {
        "name": "nonce",
        "type": "core::integer::u64"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::allowance_transfer::interface::PermitSingle",
    "members": [
      {
        "name": "details",
        "type": "oif_starknet::permit2::allowance_transfer::interface::PermitDetails"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "sig_deadline",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "struct",
    "name": "core::array::Span::<oif_starknet::permit2::allowance_transfer::interface::PermitDetails>",
    "members": [
      {
        "name": "snapshot",
        "type": "@core::array::Array::<oif_starknet::permit2::allowance_transfer::interface::PermitDetails>"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::allowance_transfer::interface::PermitBatch",
    "members": [
      {
        "name": "details",
        "type": "core::array::Span::<oif_starknet::permit2::allowance_transfer::interface::PermitDetails>"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "sig_deadline",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::allowance_transfer::interface::AllowanceTransferDetails",
    "members": [
      {
        "name": "from",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "to",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "amount",
        "type": "core::integer::u256"
      },
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::allowance_transfer::interface::TokenSpenderPair",
    "members": [
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress"
      }
    ]
  },
  {
    "type": "interface",
    "name": "oif_starknet::permit2::allowance_transfer::interface::IAllowanceTransfer",
    "items": [
      {
        "type": "function",
        "name": "allowance",
        "inputs": [
          {
            "name": "user",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "token",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "spender",
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "outputs": [
          {
            "type": "(core::integer::u256, core::integer::u64, core::integer::u64)"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "approve",
        "inputs": [
          {
            "name": "token",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "spender",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "amount",
            "type": "core::integer::u256"
          },
          {
            "name": "expiration",
            "type": "core::integer::u64"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "permit",
        "inputs": [
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "permit_single",
            "type": "oif_starknet::permit2::allowance_transfer::interface::PermitSingle"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "permit_batch",
        "inputs": [
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "permit_batch",
            "type": "oif_starknet::permit2::allowance_transfer::interface::PermitBatch"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "transfer_from",
        "inputs": [
          {
            "name": "from",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "to",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "amount",
            "type": "core::integer::u256"
          },
          {
            "name": "token",
            "type": "core::starknet::contract_address::ContractAddress"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "batch_transfer_from",
        "inputs": [
          {
            "name": "transfer_details",
            "type": "core::array::Array::<oif_starknet::permit2::allowance_transfer::interface::AllowanceTransferDetails>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "lockdown",
        "inputs": [
          {
            "name": "approvals",
            "type": "core::array::Array::<oif_starknet::permit2::allowance_transfer::interface::TokenSpenderPair>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "invalidate_nonces",
        "inputs": [
          {
            "name": "token",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "spender",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "new_nonce",
            "type": "core::integer::u64"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      }
    ]
  },
  {
    "type": "impl",
    "name": "SignatureTransferImpl",
    "interface_name": "oif_starknet::permit2::signature_transfer::interface::ISignatureTransfer"
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::signature_transfer::interface::TokenPermissions",
    "members": [
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "amount",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::signature_transfer::interface::PermitTransferFrom",
    "members": [
      {
        "name": "permitted",
        "type": "oif_starknet::permit2::signature_transfer::interface::TokenPermissions"
      },
      {
        "name": "nonce",
        "type": "core::felt252"
      },
      {
        "name": "deadline",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails",
    "members": [
      {
        "name": "to",
        "type": "core::starknet::contract_address::ContractAddress"
      },
      {
        "name": "requested_amount",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "struct",
    "name": "core::array::Span::<oif_starknet::permit2::signature_transfer::interface::TokenPermissions>",
    "members": [
      {
        "name": "snapshot",
        "type": "@core::array::Array::<oif_starknet::permit2::signature_transfer::interface::TokenPermissions>"
      }
    ]
  },
  {
    "type": "struct",
    "name": "oif_starknet::permit2::signature_transfer::interface::PermitBatchTransferFrom",
    "members": [
      {
        "name": "permitted",
        "type": "core::array::Span::<oif_starknet::permit2::signature_transfer::interface::TokenPermissions>"
      },
      {
        "name": "nonce",
        "type": "core::felt252"
      },
      {
        "name": "deadline",
        "type": "core::integer::u256"
      }
    ]
  },
  {
    "type": "struct",
    "name": "core::array::Span::<oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails>",
    "members": [
      {
        "name": "snapshot",
        "type": "@core::array::Array::<oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails>"
      }
    ]
  },
  {
    "type": "struct",
    "name": "core::byte_array::ByteArray",
    "members": [
      {
        "name": "data",
        "type": "core::array::Array::<core::bytes_31::bytes31>"
      },
      {
        "name": "pending_word",
        "type": "core::felt252"
      },
      {
        "name": "pending_word_len",
        "type": "core::integer::u32"
      }
    ]
  },
  {
    "type": "interface",
    "name": "oif_starknet::permit2::signature_transfer::interface::ISignatureTransfer",
    "items": [
      {
        "type": "function",
        "name": "permit_transfer_from",
        "inputs": [
          {
            "name": "permit",
            "type": "oif_starknet::permit2::signature_transfer::interface::PermitTransferFrom"
          },
          {
            "name": "transfer_details",
            "type": "oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails"
          },
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "permit_batch_transfer_from",
        "inputs": [
          {
            "name": "permit",
            "type": "oif_starknet::permit2::signature_transfer::interface::PermitBatchTransferFrom"
          },
          {
            "name": "transfer_details",
            "type": "core::array::Span::<oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails>"
          },
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "permit_witness_transfer_from",
        "inputs": [
          {
            "name": "permit",
            "type": "oif_starknet::permit2::signature_transfer::interface::PermitTransferFrom"
          },
          {
            "name": "transfer_details",
            "type": "oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails"
          },
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "witness",
            "type": "core::felt252"
          },
          {
            "name": "witness_type_string",
            "type": "core::byte_array::ByteArray"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      },
      {
        "type": "function",
        "name": "permit_witness_batch_transfer_from",
        "inputs": [
          {
            "name": "permit",
            "type": "oif_starknet::permit2::signature_transfer::interface::PermitBatchTransferFrom"
          },
          {
            "name": "transfer_details",
            "type": "core::array::Span::<oif_starknet::permit2::signature_transfer::interface::SignatureTransferDetails>"
          },
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "witness",
            "type": "core::felt252"
          },
          {
            "name": "witness_type_string",
            "type": "core::byte_array::ByteArray"
          },
          {
            "name": "signature",
            "type": "core::array::Array::<core::felt252>"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      }
    ]
  },
  {
    "type": "impl",
    "name": "UnorderedNoncesImpl",
    "interface_name": "oif_starknet::libraries::unordered_nonces::IUnorderedNonces"
  },
  {
    "type": "enum",
    "name": "core::bool",
    "variants": [
      {
        "name": "False",
        "type": "()"
      },
      {
        "name": "True",
        "type": "()"
      }
    ]
  },
  {
    "type": "interface",
    "name": "oif_starknet::libraries::unordered_nonces::IUnorderedNonces",
    "items": [
      {
        "type": "function",
        "name": "nonce_bitmap",
        "inputs": [
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "nonce_space",
            "type": "core::felt252"
          }
        ],
        "outputs": [
          {
            "type": "core::felt252"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "is_nonce_usable",
        "inputs": [
          {
            "name": "owner",
            "type": "core::starknet::contract_address::ContractAddress"
          },
          {
            "name": "nonce",
            "type": "core::felt252"
          }
        ],
        "outputs": [
          {
            "type": "core::bool"
          }
        ],
        "state_mutability": "view"
      },
      {
        "type": "function",
        "name": "invalidate_unordered_nonces",
        "inputs": [
          {
            "name": "nonce_space",
            "type": "core::felt252"
          },
          {
            "name": "mask",
            "type": "core::felt252"
          }
        ],
        "outputs": [],
        "state_mutability": "external"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::allowance_transfer::interface::events::NonceInvalidation",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "new_nonce",
        "type": "core::integer::u64",
        "kind": "data"
      },
      {
        "name": "old_nonce",
        "type": "core::integer::u64",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::allowance_transfer::interface::events::Approval",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "amount",
        "type": "core::integer::u256",
        "kind": "data"
      },
      {
        "name": "expiration",
        "type": "core::integer::u64",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::allowance_transfer::interface::events::Permit",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "amount",
        "type": "core::integer::u256",
        "kind": "data"
      },
      {
        "name": "expiration",
        "type": "core::integer::u64",
        "kind": "data"
      },
      {
        "name": "nonce",
        "type": "core::integer::u64",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::allowance_transfer::interface::events::Lockdown",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "token",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "data"
      },
      {
        "name": "spender",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::allowance_transfer::interface::events::AllowanceTransferEvent",
    "kind": "enum",
    "variants": [
      {
        "name": "NonceInvalidation",
        "type": "oif_starknet::permit2::allowance_transfer::interface::events::NonceInvalidation",
        "kind": "nested"
      },
      {
        "name": "Approval",
        "type": "oif_starknet::permit2::allowance_transfer::interface::events::Approval",
        "kind": "nested"
      },
      {
        "name": "Permit",
        "type": "oif_starknet::permit2::allowance_transfer::interface::events::Permit",
        "kind": "nested"
      },
      {
        "name": "Lockdown",
        "type": "oif_starknet::permit2::allowance_transfer::interface::events::Lockdown",
        "kind": "nested"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::allowance_transfer::allowance_transfer::AllowanceTransferComponent::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "AllowanceTransferEvent",
        "type": "oif_starknet::permit2::allowance_transfer::interface::events::AllowanceTransferEvent",
        "kind": "flat"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::signature_transfer::signature_transfer::SignatureTransferComponent::Event",
    "kind": "enum",
    "variants": []
  },
  {
    "type": "event",
    "name": "oif_starknet::libraries::unordered_nonces::UnorderedNoncesComponent::UnorderedNonceInvalidation",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "nonce_space",
        "type": "core::felt252",
        "kind": "data"
      },
      {
        "name": "mask",
        "type": "core::felt252",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::libraries::unordered_nonces::UnorderedNoncesComponent::NonceInvalidated",
    "kind": "struct",
    "members": [
      {
        "name": "owner",
        "type": "core::starknet::contract_address::ContractAddress",
        "kind": "key"
      },
      {
        "name": "nonce",
        "type": "core::felt252",
        "kind": "data"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::libraries::unordered_nonces::UnorderedNoncesComponent::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "UnorderedNonceInvalidation",
        "type": "oif_starknet::libraries::unordered_nonces::UnorderedNoncesComponent::UnorderedNonceInvalidation",
        "kind": "nested"
      },
      {
        "name": "NonceInvalidated",
        "type": "oif_starknet::libraries::unordered_nonces::UnorderedNoncesComponent::NonceInvalidated",
        "kind": "nested"
      }
    ]
  },
  {
    "type": "event",
    "name": "oif_starknet::permit2::permit2::Permit2::Event",
    "kind": "enum",
    "variants": [
      {
        "name": "AllowedTransferEvent",
        "type": "oif_starknet::permit2::allowance_transfer::allowance_transfer::AllowanceTransferComponent::Event",
        "kind": "flat"
      },
      {
        "name": "SignatureTransferEvent",
        "type": "oif_starknet::permit2::signature_transfer::signature_transfer::SignatureTransferComponent::Event",
        "kind": "flat"
      },
      {
        "name": "UnorderedNoncesEvent",
        "type": "oif_starknet::libraries::unordered_nonces::UnorderedNoncesComponent::Event",
        "kind": "flat"
      }
    ]
  }
] as const;

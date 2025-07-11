package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/typedData"
	"github.com/NethermindEth/starknet.go/utils"
)

// Get type, struct, and message hashes for a type
// Returns hex string
func getHexHashes(typedData *typedData.TypedData, account string) (typeHash *felt.Felt, structHash *felt.Felt, messageHash *felt.Felt, err error) {
	zero := &felt.Zero

	typeHash, err = typedData.GetTypeHash(typedData.PrimaryType)
	if err != nil {
		return zero, zero, zero, err
	}

	structHash, err = typedData.GetStructHash(typedData.PrimaryType)
	if err != nil {
		return zero, zero, zero, err
	}

	messageHash, err = typedData.GetMessageHash(account)
	if err != nil {
		return zero, zero, zero, err
	}

	return typeHash, structHash, messageHash, nil
}

// Get type, struct, and message hashes for a type
// Returns hex string
func getIntHashes(typedData *typedData.TypedData, account string) (typeHash *big.Int, structHash *big.Int, messageHash *big.Int, err error) {
	zeroFelt := felt.Zero
	zero := utils.HexToBN(zeroFelt.String())

	typeHashHex, structHashHex, messageHashHex, err := getHexHashes(typedData, account)
	if err != nil {
		return zero, zero, zero, err
	}

	return toBn(typeHashHex),
		toBn(structHashHex),
		toBn(messageHashHex),
		nil
}

func printHexHashes(typedData *typedData.TypedData, account string) {
	typeHash, structHash, messageHash, err := getHexHashes(typedData, account)

	if err != nil {
		log.Fatal("Error getting (int) hashes:", err)
		return
	}

	fmt.Println("-----", typedData.PrimaryType, "-----")
	fmt.Println("Type hash: ", typeHash)
	fmt.Println("Struct hash: ", structHash)
	fmt.Println("Message hash: ", messageHash)
}

func printIntHashes(typedData *typedData.TypedData, account string) {
	typeHash, structHash, messageHash, err := getIntHashes(typedData, account)

	if err != nil {
		log.Fatal("Error getting (int) hashes:", err)
		return
	}

	fmt.Println("-----", typedData.PrimaryType, "-----")
	fmt.Println("Type hash: ", typeHash)
	fmt.Println("Struct hash: ", structHash)
	fmt.Println("Message hash: ", messageHash)
}

func toBn(el *felt.Felt) *big.Int {
	return utils.HexToBN(el.String())
}

func main() {
	account0 := "0x127fd5f1fe78a71f8bcd1fec63e3fe2f0486b6ecd5c86a0466c3a21fa5cfcec"

	permitSingle, _ := utils.UnmarshalJSONFileToType[typedData.TypedData]("../typedData/PermitSingle.json", "")
	permitBatch, _ := utils.UnmarshalJSONFileToType[typedData.TypedData]("../typedData/PermitBatch.json", "")
	permitTransferFrom, _ := utils.UnmarshalJSONFileToType[typedData.TypedData]("../typedData/PermitTransferFrom.json", "")
	permitBatchTransferFrom, _ := utils.UnmarshalJSONFileToType[typedData.TypedData]("../typedData/PermitBatchTransferFrom.json", "")
	permitWitnessTransferFrom, _ := utils.UnmarshalJSONFileToType[typedData.TypedData]("../typedData/PermitWitnessTransferFrom.json", "")
	permitWitnessBatchTransferFrom, _ := utils.UnmarshalJSONFileToType[typedData.TypedData]("../typedData/PermitWitnessBatchTransferFrom.json", "")

	printHexHashes(permitSingle, account0)
	printHexHashes(permitBatch, account0)
	printHexHashes(permitTransferFrom, account0)
	printHexHashes(permitBatchTransferFrom, account0)
	printHexHashes(permitWitnessTransferFrom, account0)
	printHexHashes(permitWitnessBatchTransferFrom, account0)

	printIntHashes(permitSingle, account0)
	printIntHashes(permitBatch, account0)
	printIntHashes(permitTransferFrom, account0)
	printIntHashes(permitBatchTransferFrom, account0)
	printIntHashes(permitWitnessTransferFrom, account0)
	printIntHashes(permitWitnessBatchTransferFrom, account0)

	// // MockWitness
	// mockWitnessStructHash, _ := permitWitnessBatchTransferFrom.GetStructHash("Mock Witness", "Witness")
	// fmt.Println("Mock Witness Struct Hash (hex):\n-\t\t", mockWitnessStructHash)
	// fmt.Println("Mock Witness Struct Hash (int):\n-\t\t", utils.HexToBN(mockWitnessStructHash.String()))
	//
	// // Struct hashes
	// // u256StructHash, _ := permitWitnessTransferFrom.GetStructHash("U256")
	// // betaStructHash, _ := permitWitnessTransferFrom.GetStructHash("Beta" ,"Witness", "B")
	// // zetaStructHash, _ := permitWitnessTransferFrom.GetStructHash("Zeta", "Witness", "Z")
}

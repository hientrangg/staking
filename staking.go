package staking

import (
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-edge/helper/common"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/helper/keccak"
	"github.com/0xPolygon/polygon-edge/types"
)

var (
	MinValidatorCount = uint64(1)
	MaxValidatorCount = common.MaxSafeJSInt
)

// getAddressMapping returns the key for the SC storage mapping (address => something)
//
// More information:
// https://docs.soliditylang.org/en/latest/internals/layout_in_storage.html
func getAddressMapping(address types.Address, slot int64) []byte {
	bigSlot := big.NewInt(slot)

	finalSlice := append(
		common.PadLeftOrTrim(address.Bytes(), 32),
		common.PadLeftOrTrim(bigSlot.Bytes(), 32)...,
	)
	keccakValue := keccak.Keccak256(nil, finalSlice)

	return keccakValue
}

// getIndexWithOffset is a helper method for adding an offset to the already found keccak hash
func getIndexWithOffset(keccakHash []byte, offset int64) []byte {
	bigOffset := big.NewInt(offset)
	bigKeccak := big.NewInt(0).SetBytes(keccakHash)

	bigKeccak.Add(bigKeccak, bigOffset)

	return bigKeccak.Bytes()
}

// getStorageIndexes is a helper function for getting the correct indexes
// of the storage slots which need to be modified during bootstrap.
//
// It is SC dependant, and based on the SC located at:
// https://github.com/0xPolygon/staking-contracts/
func getStorageIndexes(address types.Address, index int64) *StorageIndexes {
	storageIndexes := StorageIndexes{}

	// Get the indexes for the mappings
	// The index for the mapping is retrieved with:
	// keccak(address . slot)
	// . stands for concatenation (basically appending the bytes)
	storageIndexes.AddressToIsValidatorIndex = getAddressMapping(address, addressToIsValidatorSlot)
	storageIndexes.AddressToStakedAmountIndex = getAddressMapping(address, addressToStakedAmountSlot)
	storageIndexes.AddressToValidatorIndexIndex = getAddressMapping(address, addressToValidatorIndexSlot)

	// Get the indexes for _validators, _stakedAmount
	// Index for regular types is calculated as just the regular slot
	storageIndexes.StakedAmountIndex = big.NewInt(stakedAmountSlot).Bytes()

	// Index for array types is calculated as keccak(slot) + index
	// The slot for the dynamic arrays that's put in the keccak needs to be in hex form (padded 64 chars)
	storageIndexes.ValidatorsIndex = getIndexWithOffset(
		keccak.Keccak256(nil, common.PadLeftOrTrim(big.NewInt(validatorsSlot).Bytes(), 32)),
		index,
	)

	// For any dynamic array in Solidity, the size of the actual array should be
	// located on slot x
	storageIndexes.ValidatorsArraySizeIndex = []byte{byte(validatorsSlot)}

	return &storageIndexes
}

// PredeployParams contains the values used to predeploy the PoS staking contract
type PredeployParams struct {
	MinValidatorCount uint64
	MaxValidatorCount uint64
}

// StorageIndexes is a wrapper for different storage indexes that
// need to be modified
type StorageIndexes struct {
	ValidatorsIndex              []byte // []address
	ValidatorsArraySizeIndex     []byte // []address size
	AddressToIsValidatorIndex    []byte // mapping(address => bool)
	AddressToStakedAmountIndex   []byte // mapping(address => uint256)
	AddressToValidatorIndexIndex []byte // mapping(address => uint256)
	StakedAmountIndex            []byte // uint256
}

// Slot definitions for SC storage
var (
	validatorsSlot              = int64(0) // Slot 0
	addressToIsValidatorSlot    = int64(1) // Slot 1
	addressToStakedAmountSlot   = int64(2) // Slot 2
	addressToValidatorIndexSlot = int64(3) // Slot 3
	stakedAmountSlot            = int64(4) // Slot 4
	minNumValidatorSlot         = int64(5) // Slot 5
	maxNumValidatorSlot         = int64(6) // Slot 6
)

const (
	DefaultStakedBalance = "0x8AC7230489E80000" // 10 ETH
	//nolint: lll
	StakingSCBytecode = "0x60806040523480156200001157600080fd5b5060405162001740380380620017408339818101604052810190620000379190620000aa565b808211156200007d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401620000749062000118565b60405180910390fd5b81600581905550806006819055505050620001c3565b600081519050620000a481620001a9565b92915050565b60008060408385031215620000c457620000c362000155565b5b6000620000d48582860162000093565b9250506020620000e78582860162000093565b9150509250929050565b6000620001006040836200013a565b91506200010d826200015a565b604082019050919050565b600060208201905081810360008301526200013381620000f1565b9050919050565b600082825260208201905092915050565b6000819050919050565b600080fd5b7f4d696e2076616c696461746f7273206e756d2063616e206e6f7420626520677260008201527f6561746572207468616e206d6178206e756d206f662076616c696461746f7273602082015250565b620001b4816200014b565b8114620001c057600080fd5b50565b61156d80620001d36000396000f3fe6080604052600436106100f75760003560e01c80637dceceb81161008a578063e387a7ed11610059578063e387a7ed14610381578063e804fbf6146103ac578063f90ecacc146103d7578063facd743b1461041457610165565b80637dceceb8146102c3578063af6da36e14610300578063c795c0771461032b578063ca1e78191461035657610165565b8063373d6132116100c6578063373d6132146102385780633a4b66f114610263578063714ff4251461026d5780637a6eea371461029857610165565b806302b751991461016a578063065ae171146101a75780632367f6b5146101e45780632def66201461022157610165565b366101655761011b3373ffffffffffffffffffffffffffffffffffffffff16610451565b1561015b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610152906111b0565b60405180910390fd5b610163610474565b005b600080fd5b34801561017657600080fd5b50610191600480360381019061018c9190610f2e565b61054b565b60405161019e919061120b565b60405180910390f35b3480156101b357600080fd5b506101ce60048036038101906101c99190610f2e565b610563565b6040516101db9190611135565b60405180910390f35b3480156101f057600080fd5b5061020b60048036038101906102069190610f2e565b610583565b604051610218919061120b565b60405180910390f35b34801561022d57600080fd5b506102366105cc565b005b34801561024457600080fd5b5061024d6106b7565b60405161025a919061120b565b60405180910390f35b61026b6106c1565b005b34801561027957600080fd5b5061028261072a565b60405161028f919061120b565b60405180910390f35b3480156102a457600080fd5b506102ad610734565b6040516102ba91906111f0565b60405180910390f35b3480156102cf57600080fd5b506102ea60048036038101906102e59190610f2e565b610740565b6040516102f7919061120b565b60405180910390f35b34801561030c57600080fd5b50610315610758565b604051610322919061120b565b60405180910390f35b34801561033757600080fd5b5061034061075e565b60405161034d919061120b565b60405180910390f35b34801561036257600080fd5b5061036b610764565b6040516103789190611113565b60405180910390f35b34801561038d57600080fd5b506103966107f2565b6040516103a3919061120b565b60405180910390f35b3480156103b857600080fd5b506103c16107f8565b6040516103ce919061120b565b60405180910390f35b3480156103e357600080fd5b506103fe60048036038101906103f99190610f5b565b610802565b60405161040b91906110f8565b60405180910390f35b34801561042057600080fd5b5061043b60048036038101906104369190610f2e565b610841565b6040516104489190611135565b60405180910390f35b6000808273ffffffffffffffffffffffffffffffffffffffff163b119050919050565b34600460008282546104869190611270565b9250508190555034600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546104dc9190611270565b925050819055506104ec33610897565b156104fb576104fa3361090f565b5b3373ffffffffffffffffffffffffffffffffffffffff167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d34604051610541919061120b565b60405180910390a2565b60036020528060005260406000206000915090505481565b60016020528060005260406000206000915054906101000a900460ff1681565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6105eb3373ffffffffffffffffffffffffffffffffffffffff16610451565b1561062b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610622906111b0565b60405180910390fd5b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054116106ad576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106a490611150565b60405180910390fd5b6106b5610a5e565b565b6000600454905090565b6106e03373ffffffffffffffffffffffffffffffffffffffff16610451565b15610720576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610717906111b0565b60405180910390fd5b610728610474565b565b6000600554905090565b670de0b6b3a764000081565b60026020528060005260406000206000915090505481565b60065481565b60055481565b606060008054806020026020016040519081016040528092919081815260200182805480156107e857602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001906001019080831161079e575b5050505050905090565b60045481565b6000600654905090565b6000818154811061081257600080fd5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60006108a282610bb0565b1580156109085750670de0b6b3a76400006fffffffffffffffffffffffffffffffff16600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b9050919050565b60065460008054905010610958576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161094f90611170565b60405180910390fd5b60018060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000819080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490506000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508060046000828254610af991906112c6565b92505081905550610b0933610bb0565b15610b1857610b1733610c06565b5b3373ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610b5e573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f7582604051610ba5919061120b565b60405180910390a250565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60055460008054905011610c4f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c46906111d0565b60405180910390fd5b600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410610cd5576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ccc90611190565b60405180910390fd5b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905060006001600080549050610d2d91906112c6565b9050808214610e1b576000808281548110610d4b57610d4a6113bc565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690508060008481548110610d8d57610d8c6113bc565b5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555082600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550505b6000600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506000600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000805480610eca57610ec961138d565b5b6001900381819060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690559055505050565b600081359050610f1381611509565b92915050565b600081359050610f2881611520565b92915050565b600060208284031215610f4457610f436113eb565b5b6000610f5284828501610f04565b91505092915050565b600060208284031215610f7157610f706113eb565b5b6000610f7f84828501610f19565b91505092915050565b6000610f948383610fa0565b60208301905092915050565b610fa9816112fa565b82525050565b610fb8816112fa565b82525050565b6000610fc982611236565b610fd3818561124e565b9350610fde83611226565b8060005b8381101561100f578151610ff68882610f88565b975061100183611241565b925050600181019050610fe2565b5085935050505092915050565b6110258161130c565b82525050565b6000611038601d8361125f565b9150611043826113f0565b602082019050919050565b600061105b60278361125f565b915061106682611419565b604082019050919050565b600061107e60128361125f565b915061108982611468565b602082019050919050565b60006110a1601a8361125f565b91506110ac82611491565b602082019050919050565b60006110c460408361125f565b91506110cf826114ba565b604082019050919050565b6110e381611318565b82525050565b6110f281611354565b82525050565b600060208201905061110d6000830184610faf565b92915050565b6000602082019050818103600083015261112d8184610fbe565b905092915050565b600060208201905061114a600083018461101c565b92915050565b600060208201905081810360008301526111698161102b565b9050919050565b600060208201905081810360008301526111898161104e565b9050919050565b600060208201905081810360008301526111a981611071565b9050919050565b600060208201905081810360008301526111c981611094565b9050919050565b600060208201905081810360008301526111e9816110b7565b9050919050565b600060208201905061120560008301846110da565b92915050565b600060208201905061122060008301846110e9565b92915050565b6000819050602082019050919050565b600081519050919050565b6000602082019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600061127b82611354565b915061128683611354565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156112bb576112ba61135e565b5b828201905092915050565b60006112d182611354565b91506112dc83611354565b9250828210156112ef576112ee61135e565b5b828203905092915050565b600061130582611334565b9050919050565b60008115159050919050565b60006fffffffffffffffffffffffffffffffff82169050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600080fd5b7f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e000000600082015250565b7f56616c696461746f72207365742068617320726561636865642066756c6c206360008201527f6170616369747900000000000000000000000000000000000000000000000000602082015250565b7f696e646578206f7574206f662072616e67650000000000000000000000000000600082015250565b7f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e000000000000600082015250565b7f56616c696461746f72732063616e2774206265206c657373207468616e20746860008201527f65206d696e696d756d2072657175697265642076616c696461746f72206e756d602082015250565b611512816112fa565b811461151d57600080fd5b50565b61152981611354565b811461153457600080fd5b5056fea264697066735822122080d0c809f90fc7c63b7d17a7fd3f27a05db00e98aef254352a8262fe99b436ca64736f6c63430008070033"
)

// PredeployStakingSC is a helper method for setting up the staking smart contract account,
// using the passed in validators as pre-staked validators
func PredeployStakingSC(
	validators []types.Address,
	params PredeployParams,
) (*chain.GenesisAccount, error) {
	// Set the code for the staking smart contract
	// Code retrieved from https://github.com/0xPolygon/staking-contracts
	scHex, _ := hex.DecodeHex(StakingSCBytecode)
	stakingAccount := &chain.GenesisAccount{
		Code: scHex,
	}

	// Parse the default staked balance value into *big.Int
	val := DefaultStakedBalance
	bigDefaultStakedBalance, err := types.ParseUint256orHex(&val)

	if err != nil {
		return nil, fmt.Errorf("unable to generate DefaultStatkedBalance, %w", err)
	}

	// Generate the empty account storage map
	storageMap := make(map[types.Hash]types.Hash)
	bigTrueValue := big.NewInt(1)
	stakedAmount := big.NewInt(0)
	bigMinNumValidators := big.NewInt(int64(params.MinValidatorCount))
	bigMaxNumValidators := big.NewInt(int64(params.MaxValidatorCount))

	for indx, validator := range validators {
		// Update the total staked amount
		stakedAmount.Add(stakedAmount, bigDefaultStakedBalance)

		// Get the storage indexes
		storageIndexes := getStorageIndexes(validator, int64(indx))

		// Set the value for the validators array
		storageMap[types.BytesToHash(storageIndexes.ValidatorsIndex)] =
			types.BytesToHash(
				validator.Bytes(),
			)

		// Set the value for the address -> validator array index mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToIsValidatorIndex)] =
			types.BytesToHash(bigTrueValue.Bytes())

		// Set the value for the address -> staked amount mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToStakedAmountIndex)] =
			types.StringToHash(hex.EncodeBig(bigDefaultStakedBalance))

		// Set the value for the address -> validator index mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToValidatorIndexIndex)] =
			types.StringToHash(hex.EncodeUint64(uint64(indx)))

		// Set the value for the total staked amount
		storageMap[types.BytesToHash(storageIndexes.StakedAmountIndex)] =
			types.BytesToHash(stakedAmount.Bytes())

		// Set the value for the size of the validators array
		storageMap[types.BytesToHash(storageIndexes.ValidatorsArraySizeIndex)] =
			types.StringToHash(hex.EncodeUint64(uint64(indx + 1)))
	}

	// Set the value for the minimum number of validators
	storageMap[types.BytesToHash(big.NewInt(minNumValidatorSlot).Bytes())] =
		types.BytesToHash(bigMinNumValidators.Bytes())

	// Set the value for the maximum number of validators
	storageMap[types.BytesToHash(big.NewInt(maxNumValidatorSlot).Bytes())] =
		types.BytesToHash(bigMaxNumValidators.Bytes())

	// Save the storage map
	stakingAccount.Storage = storageMap

	// Set the Staking SC balance to numValidators * defaultStakedBalance
	stakingAccount.Balance = stakedAmount

	return stakingAccount, nil
}

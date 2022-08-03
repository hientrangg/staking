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
	StakingSCBytecode = "0x608060405234801561001057600080fd5b50600436106101165760003560e01c8063adc9772e116100a2578063ca1e781911610071578063ca1e7819146102f7578063e387a7ed14610315578063e804fbf614610333578063f90ecacc14610351578063facd743b1461038157610116565b8063adc9772e14610283578063af6da36e1461029f578063c2a672e0146102bd578063c795c077146102d957610116565b80636588103b116100e95780636588103b146101c9578063714ff425146101e75780637a6eea37146102055780637dceceb814610223578063940670451461025357610116565b806302b751991461011b578063065ae1711461014b5780632367f6b51461017b578063373d6132146101ab575b600080fd5b610135600480360381019061013091906112c5565b6103b1565b60405161014291906116f6565b60405180910390f35b610165600480360381019061016091906112c5565b6103c9565b60405161017291906115c5565b60405180910390f35b610195600480360381019061019091906112c5565b6103e9565b6040516101a291906116f6565b60405180910390f35b6101b3610432565b6040516101c091906116f6565b60405180910390f35b6101d161043c565b6040516101de91906115e0565b60405180910390f35b6101ef610460565b6040516101fc91906116f6565b60405180910390f35b61020d61046a565b60405161021a91906116db565b60405180910390f35b61023d600480360381019061023891906112c5565b61046f565b60405161024a91906116f6565b60405180910390f35b61026d6004803603810190610268919061135f565b610487565b60405161027a9190611551565b60405180910390f35b61029d6004803603810190610298919061131f565b6104ba565b005b6102a7610528565b6040516102b491906116f6565b60405180910390f35b6102d760048036038101906102d2919061131f565b61052e565b005b6102e161061e565b6040516102ee91906116f6565b60405180910390f35b6102ff610624565b60405161030c91906115a3565b60405180910390f35b61031d6106b2565b60405161032a91906116f6565b60405180910390f35b61033b6106b8565b60405161034891906116f6565b60405180910390f35b61036b6004803603810190610366919061135f565b6106c2565b6040516103789190611551565b60405180910390f35b61039b600480360381019061039691906112c5565b610701565b6040516103a891906115c5565b60405180910390f35b60056020528060005260406000206000915090505481565b60036020528060005260406000206000915054906101000a900460ff1681565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000600654905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600754905090565b600181565b60046020528060005260406000206000915090505481565b60016020528060005260406000206000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6104d93373ffffffffffffffffffffffffffffffffffffffff16610757565b15610519576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105109061169b565b60405180910390fd5b61052433838361077a565b5050565b60085481565b61054d3373ffffffffffffffffffffffffffffffffffffffff16610757565b1561058d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105849061169b565b60405180910390fd5b6000600460003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541161060f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106069061161b565b60405180910390fd5b61061a338383610a89565b5050565b60075481565b606060028054806020026020016040519081016040528092919081815260200182805480156106a857602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001906001019080831161065e575b5050505050905090565b60065481565b6000600854905090565b600281815481106106d257600080fd5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b6000808273ffffffffffffffffffffffffffffffffffffffff163b119050919050565b816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508273ffffffffffffffffffffffffffffffffffffffff1660008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16636352211e836040518263ffffffff1660e01b815260040161082a91906116f6565b60206040518083038186803b15801561084257600080fd5b505afa158015610856573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061087a91906112f2565b73ffffffffffffffffffffffffffffffffffffffff16146108d0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108c79061167b565b60405180910390fd5b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd8430846040518463ffffffff1660e01b815260040161092d9392919061156c565b600060405180830381600087803b15801561094757600080fd5b505af115801561095b573d6000803e3d6000fd5b50505050826001600083815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600660008154809291906109c490611853565b9190505550600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000815480929190610a1990611853565b9190505550610a2733610d70565b15610a3657610a3533610de1565b5b8273ffffffffffffffffffffffffffffffffffffffff167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d82604051610a7c91906116f6565b60405180910390a2505050565b816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205411610b4b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b42906115fb565b60405180910390fd5b8273ffffffffffffffffffffffffffffffffffffffff166001600083815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610bb657600080fd5b60006001600083815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3085846040518463ffffffff1660e01b8152600401610c669392919061156c565b600060405180830381600087803b158015610c8057600080fd5b505af1158015610c94573d6000803e3d6000fd5b50505050600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000815480929190610ce890611829565b919050555060066000815480929190610d0090611829565b9190505550610d0e33610f31565b15610d1d57610d1c33610f87565b5b8273ffffffffffffffffffffffffffffffffffffffff167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f7582604051610d6391906116f6565b60405180910390a2505050565b6000610d7b82610f31565b158015610dda575060016fffffffffffffffffffffffffffffffff16600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b9050919050565b60085460028054905010610e2a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e219061163b565b60405180910390fd5b6001600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600280549050600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506002819080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60075460028054905011610fd0576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610fc7906116bb565b60405180910390fd5b600280549050600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410611056576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161104d9061165b565b60405180910390fd5b6000600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050600060016002805490506110ae919061175b565b905080821461119d576000600282815481106110cd576110cc6118fa565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050806002848154811061110f5761110e6118fa565b5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555082600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550505b6000600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506000600560008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550600280548061124c5761124b6118cb565b5b6001900381819060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690559055505050565b60008135905061129581611abf565b92915050565b6000815190506112aa81611abf565b92915050565b6000813590506112bf81611ad6565b92915050565b6000602082840312156112db576112da611929565b5b60006112e984828501611286565b91505092915050565b60006020828403121561130857611307611929565b5b60006113168482850161129b565b91505092915050565b6000806040838503121561133657611335611929565b5b600061134485828601611286565b9250506020611355858286016112b0565b9150509250929050565b60006020828403121561137557611374611929565b5b6000611383848285016112b0565b91505092915050565b600061139883836113a4565b60208301905092915050565b6113ad8161178f565b82525050565b6113bc8161178f565b82525050565b60006113cd82611721565b6113d78185611739565b93506113e283611711565b8060005b838110156114135781516113fa888261138c565b97506114058361172c565b9250506001810190506113e6565b5085935050505092915050565b611429816117a1565b82525050565b611438816117f3565b82525050565b600061144b60198361174a565b91506114568261192e565b602082019050919050565b600061146e601d8361174a565b915061147982611957565b602082019050919050565b600061149160278361174a565b915061149c82611980565b604082019050919050565b60006114b460128361174a565b91506114bf826119cf565b602082019050919050565b60006114d760218361174a565b91506114e2826119f8565b604082019050919050565b60006114fa601a8361174a565b915061150582611a47565b602082019050919050565b600061151d60408361174a565b915061152882611a70565b604082019050919050565b61153c816117ad565b82525050565b61154b816117e9565b82525050565b600060208201905061156660008301846113b3565b92915050565b600060608201905061158160008301866113b3565b61158e60208301856113b3565b61159b6040830184611542565b949350505050565b600060208201905081810360008301526115bd81846113c2565b905092915050565b60006020820190506115da6000830184611420565b92915050565b60006020820190506115f5600083018461142f565b92915050565b600060208201905081810360008301526116148161143e565b9050919050565b6000602082019050818103600083015261163481611461565b9050919050565b6000602082019050818103600083015261165481611484565b9050919050565b60006020820190508181036000830152611674816114a7565b9050919050565b60006020820190508181036000830152611694816114ca565b9050919050565b600060208201905081810360008301526116b4816114ed565b9050919050565b600060208201905081810360008301526116d481611510565b9050919050565b60006020820190506116f06000830184611533565b92915050565b600060208201905061170b6000830184611542565b92915050565b6000819050602082019050919050565b600081519050919050565b6000602082019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b6000611766826117e9565b9150611771836117e9565b9250828210156117845761178361189c565b5b828203905092915050565b600061179a826117c9565b9050919050565b60008115159050919050565b60006fffffffffffffffffffffffffffffffff82169050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b60006117fe82611805565b9050919050565b600061181082611817565b9050919050565b6000611822826117c9565b9050919050565b6000611834826117e9565b915060008214156118485761184761189c565b5b600182039050919050565b600061185e826117e9565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8214156118915761189061189c565b5b600182019050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600080fd5b7f596f752068617665206e6f20746f6b656e73207374616b656400000000000000600082015250565b7f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e000000600082015250565b7f56616c696461746f72207365742068617320726561636865642066756c6c206360008201527f6170616369747900000000000000000000000000000000000000000000000000602082015250565b7f696e646578206f7574206f662072616e67650000000000000000000000000000600082015250565b7f43616e2774207374616b6520746f6b656e7320796f7520646f6e2774206f776e60008201527f2100000000000000000000000000000000000000000000000000000000000000602082015250565b7f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e000000000000600082015250565b7f56616c696461746f72732063616e2774206265206c657373207468616e20746860008201527f65206d696e696d756d2072657175697265642076616c696461746f72206e756d602082015250565b611ac88161178f565b8114611ad357600080fd5b50565b611adf816117e9565b8114611aea57600080fd5b5056fea2646970667358221220e8b3e605f2387d7b5c72d05600c7f8d225f80d7cbd8ad5e50c5aa00129ede67b64736f6c63430008070033"
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

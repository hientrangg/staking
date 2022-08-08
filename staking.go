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
	DefaultStakedBalance = "0xA" // 10 Wei
	//nolint: lll
	StakingSCBytecode = "0x608060405234801561001057600080fd5b50600436106101375760003560e01c806383ed986e116100b8578063ca1e78191161007c578063ca1e781914610378578063e387a7ed14610396578063e804fbf6146103b4578063e849268d146103d2578063f90ecacc14610402578063facd743b1461043257610137565b806383ed986e146102c057806394067045146102f0578063af6da36e14610320578063c795c0771461033e578063c9a3911e1461035c57610137565b80635dbe4756116100ff5780635dbe47561461021a5780636588103b14610236578063714ff425146102545780637a6eea37146102725780637dceceb81461029057610137565b806302b751991461013c578063065ae1711461016c5780630e1af57b1461019c5780632367f6b5146101cc578063373d6132146101fc575b600080fd5b61015660048036038101906101519190611602565b610462565b6040516101639190611648565b60405180910390f35b61018660048036038101906101819190611602565b61047a565b604051610193919061167e565b60405180910390f35b6101b660048036038101906101b191906116c5565b61049a565b6040516101c39190611648565b60405180910390f35b6101e660048036038101906101e19190611602565b6104c2565b6040516101f39190611648565b60405180910390f35b61020461050b565b6040516102119190611648565b60405180910390f35b610234600480360381019061022f9190611757565b610515565b005b61023e610607565b60405161024b9190611816565b60405180910390f35b61025c61062d565b6040516102699190611648565b60405180910390f35b61027a610637565b604051610287919061185c565b60405180910390f35b6102aa60048036038101906102a59190611602565b61063c565b6040516102b79190611648565b60405180910390f35b6102da60048036038101906102d59190611602565b610654565b6040516102e79190611648565b60405180910390f35b61030a600480360381019061030591906116c5565b61066c565b6040516103179190611886565b60405180910390f35b61032861069f565b6040516103359190611648565b60405180910390f35b6103466106a5565b6040516103539190611648565b60405180910390f35b61037660048036038101906103719190611757565b6106ab565b005b61038061071b565b60405161038d919061195f565b60405180910390f35b61039e6107a9565b6040516103ab9190611648565b60405180910390f35b6103bc6107af565b6040516103c99190611648565b60405180910390f35b6103ec60048036038101906103e79190611602565b6107b9565b6040516103f99190611648565b60405180910390f35b61041c600480360381019061041791906116c5565b610802565b6040516104299190611886565b60405180910390f35b61044c60048036038101906104479190611602565b610841565b604051610459919061167e565b60405180910390f35b60036020528060005260406000206000915090505481565b60016020528060005260406000206000915054906101000a900460ff1681565b6000806003836104aa91906119b0565b9050600081036104b957600390505b80915050919050565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000600454905090565b6105343373ffffffffffffffffffffffffffffffffffffffff16610897565b15610574576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161056b90611a3e565b60405180910390fd5b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054116105f6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105ed90611aaa565b60405180910390fd5b610602338484846108ba565b505050565b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600554905090565b600181565b60026020528060005260406000206000915090505481565b60096020528060005260406000206000915090505481565b60086020528060005260406000206000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60065481565b60055481565b6106ca3373ffffffffffffffffffffffffffffffffffffffff16610897565b1561070a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161070190611a3e565b60405180910390fd5b61071633848484610c92565b505050565b6060600080548060200260200160405190810160405280929190818152602001828054801561079f57602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610755575b5050505050905090565b60045481565b6000600654905090565b6000600960008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000818154811061081257600080fd5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b6000808273ffffffffffffffffffffffffffffffffffffffff163b119050919050565b82600760006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060005b82829050811015610c22576000600260008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541161098b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161098290611b16565b60405180910390fd5b8473ffffffffffffffffffffffffffffffffffffffff16600860008585858181106109b9576109b8611b36565b5b90506020020135815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610a0f57600080fd5b600060086000858585818110610a2857610a27611b36565b5b90506020020135815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3087868686818110610ace57610acd611b36565b5b905060200201356040518463ffffffff1660e01b8152600401610af393929190611b65565b600060405180830381600087803b158015610b0d57600080fd5b505af1158015610b21573d6000803e3d6000fd5b505050506000610b49848484818110610b3d57610b3c611b36565b5b9050602002013561049a565b9050600260008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000815480929190610b9b90611bcb565b919050555060046000815480929190610bb390611bcb565b919050555080600960008873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254610c079190611bf4565b92505081905550508080610c1a90611c28565b9150506108fe565b50610c2c33611086565b610c3a57610c39336110f7565b5b8373ffffffffffffffffffffffffffffffffffffffff167f919911ef4f1dae108ef03f55fc213eb8d7af9eabb5e48a520b3c49653bef9816848484604051610c8493929190611cf1565b60405180910390a250505050565b82600760006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060005b82829050811015611015578473ffffffffffffffffffffffffffffffffffffffff16600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16636352211e858585818110610d4957610d48611b36565b5b905060200201356040518263ffffffff1660e01b8152600401610d6c9190611648565b602060405180830381865afa158015610d89573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610dad9190611d38565b73ffffffffffffffffffffffffffffffffffffffff1614610e03576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610dfa90611dd7565b60405180910390fd5b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd8630868686818110610e5657610e55611b36565b5b905060200201356040518463ffffffff1660e01b8152600401610e7b93929190611b65565b600060405180830381600087803b158015610e9557600080fd5b505af1158015610ea9573d6000803e3d6000fd5b505050508460086000858585818110610ec557610ec4611b36565b5b90506020020135815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000610f3c848484818110610f3057610f2f611b36565b5b9050602002013561049a565b905060046000815480929190610f5190611c28565b9190505550600260008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000815480929190610fa690611c28565b919050555080600960008873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254610ffa9190611df7565b9250508190555050808061100d90611c28565b915050610cd6565b5061101f33611086565b1561102e5761102d336113f5565b5b8373ffffffffffffffffffffffffffffffffffffffff167f9c53a6e038ffcb5ea5287a9680067e6e1695f7cec04dea804a01dc6aa561074584848460405161107893929190611cf1565b60405180910390a250505050565b600061109182611544565b1580156110f0575060016fffffffffffffffffffffffffffffffff16600960008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b9050919050565b60055460008054905011611140576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161113790611ebf565b60405180910390fd5b600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054106111c6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016111bd90611f2b565b60405180910390fd5b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490506000600160008054905061121e9190611bf4565b905080821461130c57600080828154811061123c5761123b611b36565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050806000848154811061127e5761127d611b36565b5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555082600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550505b6000600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506000600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555060008054806113bb576113ba611f4b565b5b6001900381819060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690559055505050565b6006546000805490501061143e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161143590611fec565b60405180910390fd5b60018060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000819080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006115cf826115a4565b9050919050565b6115df816115c4565b81146115ea57600080fd5b50565b6000813590506115fc816115d6565b92915050565b6000602082840312156116185761161761159a565b5b6000611626848285016115ed565b91505092915050565b6000819050919050565b6116428161162f565b82525050565b600060208201905061165d6000830184611639565b92915050565b60008115159050919050565b61167881611663565b82525050565b6000602082019050611693600083018461166f565b92915050565b6116a28161162f565b81146116ad57600080fd5b50565b6000813590506116bf81611699565b92915050565b6000602082840312156116db576116da61159a565b5b60006116e9848285016116b0565b91505092915050565b600080fd5b600080fd5b600080fd5b60008083601f840112611717576117166116f2565b5b8235905067ffffffffffffffff811115611734576117336116f7565b5b6020830191508360208202830111156117505761174f6116fc565b5b9250929050565b6000806000604084860312156117705761176f61159a565b5b600061177e868287016115ed565b935050602084013567ffffffffffffffff81111561179f5761179e61159f565b5b6117ab86828701611701565b92509250509250925092565b6000819050919050565b60006117dc6117d76117d2846115a4565b6117b7565b6115a4565b9050919050565b60006117ee826117c1565b9050919050565b6000611800826117e3565b9050919050565b611810816117f5565b82525050565b600060208201905061182b6000830184611807565b92915050565b60006fffffffffffffffffffffffffffffffff82169050919050565b61185681611831565b82525050565b6000602082019050611871600083018461184d565b92915050565b611880816115c4565b82525050565b600060208201905061189b6000830184611877565b92915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b6118d6816115c4565b82525050565b60006118e883836118cd565b60208301905092915050565b6000602082019050919050565b600061190c826118a1565b61191681856118ac565b9350611921836118bd565b8060005b8381101561195257815161193988826118dc565b9750611944836118f4565b925050600181019050611925565b5085935050505092915050565b600060208201905081810360008301526119798184611901565b905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b60006119bb8261162f565b91506119c68361162f565b9250826119d6576119d5611981565b5b828206905092915050565b600082825260208201905092915050565b7f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e000000000000600082015250565b6000611a28601a836119e1565b9150611a33826119f2565b602082019050919050565b60006020820190508181036000830152611a5781611a1b565b9050919050565b7f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e000000600082015250565b6000611a94601d836119e1565b9150611a9f82611a5e565b602082019050919050565b60006020820190508181036000830152611ac381611a87565b9050919050565b7f596f752068617665206e6f20746f6b656e73207374616b656400000000000000600082015250565b6000611b006019836119e1565b9150611b0b82611aca565b602082019050919050565b60006020820190508181036000830152611b2f81611af3565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b6000606082019050611b7a6000830186611877565b611b876020830185611877565b611b946040830184611639565b949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000611bd68261162f565b915060008203611be957611be8611b9c565b5b600182039050919050565b6000611bff8261162f565b9150611c0a8361162f565b925082821015611c1d57611c1c611b9c565b5b828203905092915050565b6000611c338261162f565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203611c6557611c64611b9c565b5b600182019050919050565b600082825260208201905092915050565b600080fd5b82818337600083830152505050565b6000611ca18385611c70565b93507f07ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff831115611cd457611cd3611c81565b5b602083029250611ce5838584611c86565b82840190509392505050565b6000604082019050611d066000830186611877565b8181036020830152611d19818486611c95565b9050949350505050565b600081519050611d32816115d6565b92915050565b600060208284031215611d4e57611d4d61159a565b5b6000611d5c84828501611d23565b91505092915050565b7f43616e2774207374616b6520746f6b656e7320796f7520646f6e2774206f776e60008201527f2100000000000000000000000000000000000000000000000000000000000000602082015250565b6000611dc16021836119e1565b9150611dcc82611d65565b604082019050919050565b60006020820190508181036000830152611df081611db4565b9050919050565b6000611e028261162f565b9150611e0d8361162f565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115611e4257611e41611b9c565b5b828201905092915050565b7f56616c696461746f72732063616e2774206265206c657373207468616e20746860008201527f65206d696e696d756d2072657175697265642076616c696461746f72206e756d602082015250565b6000611ea96040836119e1565b9150611eb482611e4d565b604082019050919050565b60006020820190508181036000830152611ed881611e9c565b9050919050565b7f696e646578206f7574206f662072616e67650000000000000000000000000000600082015250565b6000611f156012836119e1565b9150611f2082611edf565b602082019050919050565b60006020820190508181036000830152611f4481611f08565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fd5b7f56616c696461746f72207365742068617320726561636865642066756c6c206360008201527f6170616369747900000000000000000000000000000000000000000000000000602082015250565b6000611fd66027836119e1565b9150611fe182611f7a565b604082019050919050565b6000602082019050818103600083015261200581611fc9565b905091905056fea26469706673582212207714b1bc3974c5cf79ecf8ea6f2314493a82732adcb8353544f8a3d419211f3164736f6c634300080f0033"
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

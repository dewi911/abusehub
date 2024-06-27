package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// StZETA is a simplified Go binding of the stZETA contract
type StZETA struct {
	*bind.BoundContract
}

func main() {

	client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum network: %v", err)
	}

	privateKey, err := crypto.HexToECDSA("YOUR-PRIVATE-KEY-HERE")
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to cast public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get account nonce: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to suggest gas price: %v", err)
	}

	contractAddress := common.HexToAddress("YOUR-CONTRACT-ADDRESS-HERE")

	stZETA, err := NewStZETA(contractAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate a StZETA contract: %v", err)
	}

	recipient := common.HexToAddress("RECIPIENT-ADDRESS-HERE")
	amount := big.NewInt(1000000000000000000) // 1 token with 18 decimals

	tx, err := stZETA.Mint(auth, recipient, amount)
	if err != nil {
		log.Fatalf("Failed to mint tokens: %v", err)
	}

	fmt.Printf("Mint transaction sent: %s\n", tx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to get transaction receipt: %v", err)
	}

	fmt.Printf("Mint transaction mined in block %d\n", receipt.BlockNumber.Uint64())
}

func NewStZETA(address common.Address, backend bind.ContractBackend) (*StZETA, error) {
	contract, err := bindStZETA(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StZETA{contract}, nil
}

func (_StZETA *StZETA) Mint(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _StZETA.BoundContract.Transact(opts, "mint", to, amount)
}

func bindStZETA(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(StZETAABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

const StZETAABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

package nftlabs

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nftlabs/nftlabs-sdk-go/internal/abi"
)

type Currency interface {
	Balance() (CurrencyValue, error)
	Get() (CurrencyMetadata, error)
	BalanceOf(address string) (CurrencyValue, error)
	GetValue(value *big.Int) (CurrencyValue, error)
	Transfer(to string, amount *big.Int) error
	Allowance(spender string) (*big.Int, error)
	SetAllowance(spender string, amount *big.Int) error
	Mint(amount *big.Int) error
	MintTo(to string, amount *big.Int) error
	Burn(amount *big.Int) error
	BurnFrom(from string, amount *big.Int) error
	TransferFrom(from string, to string, amount *big.Int) error
	GrantRole(role Role, address string) error
	RevokeRole(role Role, address string) error
	TotalSupply() (*big.Int, error)

	formatUnits(value *big.Int, units *big.Int) string
	getModule() *abi.Currency
}

type CurrencyModule struct {
	Client  *ethclient.Client
	Address string
	module  *abi.Currency

	main ISdk
}

func (sdk *CurrencyModule) getModule() *abi.Currency {
	return sdk.module
}

func newCurrencyModule(client *ethclient.Client, asset string, main ISdk) (*CurrencyModule, error) {
	module, err := abi.NewCurrency(common.HexToAddress(asset), client)
	if err != nil {
		return nil, err
	}

	return &CurrencyModule{
		Client:  client,
		Address: asset,
		module:  module,
		main: main,
	}, nil
}


func (sdk *CurrencyModule) TotalSupply() (*big.Int, error) {
	return sdk.module.TotalSupply(&bind.CallOpts{})
}

func (sdk *CurrencyModule) Allowance(spender string) (*big.Int, error) {
	return sdk.module.Allowance(&bind.CallOpts{}, sdk.main.getSignerAddress(), common.HexToAddress(spender))
}

func (sdk *CurrencyModule) SetAllowance(spender string, amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "nft"}
	}
	if tx, err := sdk.module.Approve(sdk.main.getTransactOpts(true), common.HexToAddress(spender), amount); err != nil {
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

func (sdk *CurrencyModule) Mint(amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "currency"}
	}
	if tx, err := sdk.module.CurrencyTransactor.Mint(sdk.main.getTransactOpts(true), sdk.main.getSignerAddress(), amount); err != nil {
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

func (sdk *CurrencyModule) Burn(amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "nft"}
	}
	if tx, err := sdk.module.CurrencyTransactor.Burn(sdk.main.getTransactOpts(true), amount); err != nil {
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

func (sdk *CurrencyModule) BurnFrom(from string, amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "nft"}
	}
	if tx, err := sdk.module.CurrencyTransactor.BurnFrom(sdk.main.getTransactOpts(true), common.HexToAddress(from), amount); err != nil {
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

func (sdk *CurrencyModule) TransferFrom(from string, to string, amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "nft"}
	}
	if tx, err := sdk.module.CurrencyTransactor.TransferFrom(sdk.main.getTransactOpts(true), common.HexToAddress(from), common.HexToAddress(to), amount); err != nil {
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

// WIP, do not call yet, need to encode role
func (sdk *CurrencyModule) GrantRole(role Role, address string) error {
	roleHash := crypto.Keccak256([]byte(fmt.Sprintf("0x%v", role)))
	r := [32]byte{}
	copy(r[:], roleHash)
	log.Printf("Role = %v\n", string(role))
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "nft"}
	}
	if tx, err := sdk.module.CurrencyTransactor.GrantRole(sdk.main.getTransactOpts(true), r, common.HexToAddress(address)); err != nil { // TODO: fill in role in [32]byte
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

// WIP, do not call yet, need to encode role
func (sdk *CurrencyModule) RevokeRole(role Role, address string) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "nft"}
	}
	if tx, err := sdk.module.CurrencyTransactor.RevokeRole(sdk.main.getTransactOpts(true), [32]byte{}, common.HexToAddress(address)); err != nil { // TODO: fill in role in [32]byte
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

func (sdk *CurrencyModule) Get() (CurrencyMetadata, error) {
	if strings.HasPrefix(sdk.Address, "0x0000000") {
		return CurrencyMetadata{}, nil
	}

	erc20Module, err := newErc20SdkModule(sdk.Client, sdk.Address, &SdkOptions{})
	if err != nil {
		return CurrencyMetadata{}, err
	}

	name, err := erc20Module.module.Name(&bind.CallOpts{})
	if err != nil {
		return CurrencyMetadata{}, err
	}

	symbol, err := erc20Module.module.Symbol(&bind.CallOpts{})
	if err != nil {
		return CurrencyMetadata{}, err
	}

	decimals, err := erc20Module.module.Decimals(&bind.CallOpts{})
	if err != nil {
		return CurrencyMetadata{}, err
	}

	return CurrencyMetadata{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}, nil
}

// TODO; test market listing with decimal place; write some basic tests
func (sdk *CurrencyModule) formatUnits(value *big.Int, units *big.Int) string {
	if value.Int64() == 0 {
		return "0"
	}

	unit := big.NewInt(18)
	if units != nil {
		unit.Set(units)
	}

	decimalTransformer := big.NewInt(10)
	decimalTransformer.Exp(decimalTransformer, unit, big.NewInt(0))
	transformer := big.NewFloat(0)
	transformer.SetString(decimalTransformer.String())

	v := big.NewFloat(0)
	v.SetString(value.String())
	return v.Quo(v, transformer).String()
}

func (sdk *CurrencyModule) GetValue(value *big.Int) (CurrencyValue, error) {
	if sdk.Address == common.HexToAddress("0").Hex() {
		return CurrencyValue{}, nil
	}

	name, err := sdk.module.CurrencyCaller.Name(&bind.CallOpts{})
	if err != nil {
		return CurrencyValue{}, err
	}

	symbol, err := sdk.module.CurrencyCaller.Symbol(&bind.CallOpts{})
	if err != nil {
		return CurrencyValue{}, err
	}

	decimals, err := sdk.module.CurrencyCaller.Decimals(&bind.CallOpts{})
	if err != nil {
		return CurrencyValue{}, err
	}

	return CurrencyValue{
		CurrencyMetadata: CurrencyMetadata{
			Name:     name,
			Symbol:   symbol,
			Decimals: decimals,
		},
		Value: value,
		DisplayValue: sdk.formatUnits(value, big.NewInt(int64(decimals))),
	}, nil
}

func (sdk *CurrencyModule) Balance() (CurrencyValue, error) {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return CurrencyValue{}, &NoSignerError{typeName: "nft"}
	}
	if balance, err := sdk.module.BalanceOf(&bind.CallOpts{}, sdk.main.getSignerAddress()); err != nil {
		 return CurrencyValue{}, err
	} else {
		return sdk.GetValue(balance)
	}
}

func (sdk *CurrencyModule) BalanceOf(address string) (CurrencyValue, error) {
	if balance, err := sdk.module.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address)); err != nil {
		return CurrencyValue{}, err
	} else {
		return sdk.GetValue(balance)
	}
}

func (sdk *CurrencyModule) Transfer(to string, amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "currency"}
	}
	if tx, err := sdk.module.CurrencyTransactor.Transfer(sdk.main.getTransactOpts(true), common.HexToAddress(to), amount); err != nil { // TODO: fill in role in [32]byte
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

func (sdk *CurrencyModule) MintTo(to string, amount *big.Int) error {
	if sdk.main.getSignerAddress() == common.HexToAddress("0") {
		return &NoSignerError{typeName: "currency"}
	}
	if tx, err := sdk.module.CurrencyTransactor.Mint(sdk.main.getTransactOpts(true), common.HexToAddress(to), amount); err != nil {
		return err
	} else {
		return waitForTx(sdk.Client, tx.Hash(), txWaitTimeBetweenAttempts, txMaxAttempts)
	}
}

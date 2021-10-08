package nftlabs

import "math/big"

type SdkOptions struct {
	IpfsGatewayUrl string
	PrivateKey string
	GasSpeed string
	MaxGasPriceInGwei *big.Int
	RegistryContractAddress string
}

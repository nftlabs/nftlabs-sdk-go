package thirdweb

import (
  pkg "github.com/thirdweb-dev/go-sdk/pkg/thirdweb"
)

func NewThirdwebSDK(rpcUrlOrChainName string, options *pkg.SDKOptions) (*pkg.ThirdwebSDK, error) {
  return pkg.NewThirdwebSDK(rpcUrlOrChainName, options)
}


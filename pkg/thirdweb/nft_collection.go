package thirdweb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/thirdweb-dev/go-sdk/internal/abi"
)

type nftCollection struct {
	contractWrapper *ContractWrapper[*abi.TokenERC721]
	*ERC721
}

func newNFTCollection(provider *ethclient.Client, address common.Address, privateKey string, storage Storage) (*nftCollection, error) {
	if erc721, err := abi.NewTokenERC721(address, provider); err != nil {
		return nil, err
	} else {
		if contractWrapper, err := NewContractWrapper(erc721, address, provider, privateKey); err != nil {
			return nil, err
		} else {
			if erc721, err := NewERC721(provider, address, privateKey, storage); err != nil {
				return nil, err
			} else {
				nftCollection := &nftCollection{
					contractWrapper,
					erc721,
				}
				return nftCollection, nil
			}
		}
	}
}

func (nft *nftCollection) Mint(metadata *NFTMetadataInput) (*types.Transaction, error) {
	address := nft.contractWrapper.GetSignerAddress().String()
	return nft.MintTo(address, metadata)
}

func (nft *nftCollection) MintTo(address string, metadata *NFTMetadataInput) (*types.Transaction, error) {
	uri, err := uploadOrExtractUri(metadata, nft.storage)
	if err != nil {
		return nil, err
	}

	tx, err := nft.contractWrapper.abi.MintTo(
		nft.contractWrapper.getTxOptions(),
		common.HexToAddress(address),
		uri,
	)
	if err != nil {
		return nil, err
	}

	return nft.contractWrapper.awaitTx(tx.Hash())
}

func (nft *nftCollection) MintBatch(metadatas []*NFTMetadataInput) (*types.Transaction, error) {
	address := nft.contractWrapper.GetSignerAddress().String()
	return nft.MintBatchTo(address, metadatas)
}

func (nft *nftCollection) MintBatchTo(address string, metadatas []*NFTMetadataInput) (*types.Transaction, error) {
	uris, err := uploadOrExtractUris(metadatas, nft.storage)
	if err != nil {
		return nil, err
	}

	encoded := [][]byte{}
	for _, uri := range uris {
		tx, err := nft.contractWrapper.abi.MintTo(
			nft.contractWrapper.getTxOptions(),
			common.HexToAddress(address), uri,
		)
		if err != nil {
			return nil, err
		}

		encoded = append(encoded, tx.Data())
	}

	tx, err := nft.contractWrapper.abi.Multicall(nft.contractWrapper.getTxOptions(), encoded)
	if err != nil {
		return nil, err
	}

	return nft.contractWrapper.awaitTx(tx.Hash())
}

package nftlabs

import (
	"encoding/json"
	"errors"
)

type Metadata struct {
	MetadataUri    string
	MetadataObject interface{}
}

func (arg *Metadata) UnmarshalJSON(data []byte) error {
	*arg = Metadata{}
	if data[0] == '"' {
		arg.MetadataUri = string(data[1 : len(data)-1])
		return nil
	}
	return json.Unmarshal(data, &arg.MetadataObject)
}

func (arg *Metadata) MarshalJSON() ([]byte, error) {
	if arg.MetadataUri != "" {
		return []byte(arg.MetadataUri), nil
	} else if arg.MetadataObject != nil {
		return json.Marshal(arg.MetadataObject)
	} else  {
		// TODO: return unrecognized type error
		return nil, errors.New("Unrecognized type provied to create collection")
	}
}

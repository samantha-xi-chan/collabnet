package api

import "encoding/json"

type HttpRespBody struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (f *HttpRespBody) UnmarshalJSON(data []byte) error {
	type cloneType HttpRespBody

	rawMsg := json.RawMessage{}
	f.Data = &rawMsg

	if err := json.Unmarshal(data, (*cloneType)(f)); err != nil {
		return err
	}

	return nil
}

// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package models

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonC2c7f4edDecodeBDV2InternalModels(in *jlexer.Lexer, out *FailedResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "message":
			out.Message = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonC2c7f4edEncodeBDV2InternalModels(out *jwriter.Writer, in FailedResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"message\":"
		out.RawString(prefix[1:])
		out.String(string(in.Message))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v FailedResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonC2c7f4edEncodeBDV2InternalModels(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v FailedResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonC2c7f4edEncodeBDV2InternalModels(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *FailedResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonC2c7f4edDecodeBDV2InternalModels(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *FailedResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonC2c7f4edDecodeBDV2InternalModels(l, v)
}
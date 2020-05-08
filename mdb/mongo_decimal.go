package mdb

import (
	"fmt"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
)

var (
	tDecimal = reflect.TypeOf(decimal.Zero)
)

type DecimalCodec struct{}

func (d *DecimalCodec) EncodeValue(_ bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tDecimal {
		return bsoncodec.ValueEncoderError{Name: "DecimalEncodeValue", Types: []reflect.Type{tDecimal}, Received: val}
	}
	td := val.Interface().(decimal.Decimal)
	pd, ok := primitive.ParseDecimal128FromBigInt(td.Coefficient(), int(td.Exponent()))
	if !ok {
		return bsoncodec.ValueEncoderError{Name: "DecimalEncodeValue", Types: []reflect.Type{tDecimal}, Received: val}
	}
	return vw.WriteDecimal128(pd)
}

func (d *DecimalCodec) DecodeValue(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tDecimal {
		return bsoncodec.ValueDecoderError{Name: "DecimalDecodeValue", Types: []reflect.Type{tDecimal}, Received: val}
	}

	var decimalVal decimal.Decimal
	valueType := vr.Type()
	switch valueType {
	case bsontype.Decimal128:
		dt, err := vr.ReadDecimal128()
		if err != nil {
			return err
		}
		bi, exp, err := dt.BigInt()
		if err != nil {
			return bsoncodec.ValueDecoderError{Name: "DecimalDecodeValue", Types: []reflect.Type{tDecimal}, Received: val}
		}
		decimalVal = decimal.NewFromBigInt(bi, int32(exp))
	case bsontype.String:
		decimalStr, err := vr.ReadString()
		if err != nil {
			return err
		}
		decimalVal, err = decimal.NewFromString(decimalStr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot decode %v into a decimal.Decimal", valueType)
	}

	val.Set(reflect.ValueOf(decimalVal))
	return nil
}

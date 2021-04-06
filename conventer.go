package main

import (
	"errors"
	"fmt"
	"reflect"
)

type ConvertSet struct {
	//Handler must be the fund(inType) (outType, error)
	Handler interface{}
	//Override the existing same type convert handler
	Override bool
}

type Converter struct {
	registry map[reflect.Type]map[reflect.Type]reflect.Value
}

func New() *Converter {
	converter := new(Converter)
	converter.registry = make(map[reflect.Type]map[reflect.Type]reflect.Value)
	return converter
}

// Register a convert handler, handler must be the func(inType) (outType, error)
func (c *Converter) Register(handler interface{}) error {
	inType := reflect.TypeOf(handler).In(0)
	outType := reflect.TypeOf(handler).Out(0)
	handlerValue := reflect.ValueOf(handler)
	if err := c.verifyHandler(handlerValue.Type()); err != nil {
		return err
	}

	outMap, ok := c.registry[inType]
	if !ok {
		outMap = make(map[reflect.Type]reflect.Value)
		c.registry[inType] = outMap
	}
	//if !set.Override {
	//	if _, ok := outMap[outType]; ok {
	//		return fmt.Errorf("%s already exist", handlerValue.Type().Name())
	//	}
	//}
	outMap[outType] = handlerValue
	return nil
}

// verifyHandler will check the input handler type
func (c *Converter) verifyHandler(handlerType reflect.Type) error {
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be func(T1) (T2, error)")
	}
	if handlerType.NumIn() != 1 || handlerType.NumOut() != 2 || handlerType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("%s doesn't match func(T1) (T2, error)", handlerType.Name())
	}
	return nil
}

func (c *Converter) Convert(in, out interface{}) error {
	inType := reflect.TypeOf(in)
	outType := reflect.TypeOf(out).Elem()
	handler, err := c.getHandler(inType, outType)
	if err != nil {
		return err
	}
	result := handler.Call([]reflect.Value{reflect.ValueOf(in)})
	reflect.ValueOf(out).Elem().Set(result[0])
	if result[1].Interface() != nil {
		return result[1].Interface().(error)
	}
	return nil
}

func (c *Converter) getHandler(in, out reflect.Type) (reflect.Value, error) {
	outMap, ok := c.registry[in]
	if !ok {
		return reflect.Value{}, errors.New("can't find any matches")
	}
	handler, ok := outMap[out]
	if !ok {
		return reflect.Value{}, errors.New("can't find any matches")
	}
	return handler, nil
}

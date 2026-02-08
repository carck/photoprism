package deepcopier

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	TagName           = "deepcopier"
	FieldOptionName   = "field"
	ContextOptionName = "context"
	SkipOptionName    = "skip"
	ForceOptionName   = "force"
)

type (
	TagOptions map[string]string

	Options struct {
		Context  map[string]interface{}
		Reversed bool
	}

	DeepCopier struct {
		src interface{}
		dst interface{}
		ctx map[string]interface{}
	}

	fieldInfo struct {
		Name       string
		TagOptions TagOptions
		FieldType  reflect.Type
		Index      int
	}
)

var (
	fieldCache  sync.Map
	methodCache sync.Map
)

// ------------------ 对外接口 ------------------

func Copy(src interface{}) *DeepCopier {
	return &DeepCopier{src: src}
}

func (dc *DeepCopier) WithContext(ctx map[string]interface{}) *DeepCopier {
	dc.ctx = ctx
	return dc
}

func (dc *DeepCopier) To(dst interface{}) error {
	dc.dst = dst
	return process(dc.dst, dc.src, Options{Context: dc.ctx})
}

func (dc *DeepCopier) From(src interface{}) error {
	dc.dst = dc.src
	dc.src = src
	return process(dc.dst, dc.src, Options{Context: dc.ctx, Reversed: true})
}

// ------------------ 核心复制逻辑 ------------------

func process(dst, src interface{}, opts Options) error {
	srcVal := reflect.Indirect(reflect.ValueOf(src))
	dstVal := reflect.Indirect(reflect.ValueOf(dst))

	if !dstVal.CanAddr() {
		return fmt.Errorf("destination %+v is unaddressable", dstVal.Interface())
	}

	srcType := srcVal.Type()
	dstType := dstVal.Type()

	srcFields := getCachedFields(srcType)
	dstFields := getCachedFields(dstType)
	srcMethods := getCachedMethods(srcType)

	// 字段
	for _, f := range srcFields {
		srcFieldVal := srcVal.FieldByIndex([]int{f.Index})
		dstFieldName := f.Name
		tagOpts := f.TagOptions

		if opts.Reversed {
			if v, ok := tagOpts[FieldOptionName]; ok && v != "" {
				dstFieldName = v
			}
		} else {
			if df, ok := dstFields[f.Name]; ok {
				dstFieldName, tagOpts = df.Name, df.TagOptions
			}
		}

		if _, skip := tagOpts[SkipOptionName]; skip {
			continue
		}

		dstFieldInfo, found := dstFields[dstFieldName]
		if !found {
			continue
		}

		dstFieldVal := dstVal.FieldByIndex([]int{dstFieldInfo.Index})
		force := tagOpts[ForceOptionName] != ""

		if err := copyField(dstFieldVal, dstFieldInfo.FieldType, srcFieldVal, f.FieldType, force); err != nil {
			return err
		}
	}

	// 方法
	for name := range srcMethods {
		dstFieldInfo, found := dstFields[name]
		if !found {
			continue
		}

		dstFieldVal := dstVal.FieldByIndex([]int{dstFieldInfo.Index})
		force := dstFieldInfo.TagOptions[ForceOptionName] != ""

		args := []reflect.Value{}
		if _, ok := dstFieldInfo.TagOptions[ContextOptionName]; ok && opts.Context != nil {
			args = append(args, reflect.ValueOf(opts.Context))
		}

		result := reflect.ValueOf(src).MethodByName(name).Call(args)[0]
		if err := copyMethodResult(dstFieldVal, dstFieldInfo.FieldType, result, force); err != nil {
			return err
		}
	}

	return nil
}

// ------------------ 字段与方法赋值 ------------------

func copyField(dst reflect.Value, dstType reflect.Type, src reflect.Value, srcType reflect.Type, force bool) error {
	if isNullableType(srcType) {
		val, _ := src.Interface().(driver.Valuer).Value()
		if val == nil {
			return nil
		}
		rv := reflect.ValueOf(val)
		if dst.Kind() == reflect.Ptr && force {
			ptr := reflect.New(rv.Type())
			ptr.Elem().Set(rv)
			if ptr.Type().AssignableTo(dstType) {
				dst.Set(ptr)
			}
			return nil
		}
		if force && rv.Type().AssignableTo(dstType) {
			dst.Set(rv)
			return nil
		}
		return nil
	}

	if srcType.Kind() == reflect.Ptr && !src.IsNil() && dstType.Kind() != reflect.Ptr {
		indirect := reflect.Indirect(src)
		if indirect.Type().AssignableTo(dstType) {
			dst.Set(indirect)
		}
		return nil
	}

	if dst.Kind() == reflect.Interface && force {
		dst.Set(src)
		return nil
	}

	if srcType.AssignableTo(dstType) {
		dst.Set(src)
	}
	return nil
}

func copyMethodResult(dst reflect.Value, dstType reflect.Type, result reflect.Value, force bool) error {
	if !result.IsValid() {
		return nil
	}

	val := result

	if dst.Kind() == reflect.Ptr && force {
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		if ptr.Type().AssignableTo(dstType) {
			dst.Set(ptr)
		}
		return nil
	}

	if val.Kind() == reflect.Ptr && force && val.Elem().Type().AssignableTo(dstType) {
		dst.Set(val.Elem())
		return nil
	}

	if val.Type().AssignableTo(dstType) {
		dst.Set(val)
	}
	return nil
}

// ------------------ 缓存字段与方法 ------------------

func getCachedFields(t reflect.Type) map[string]fieldInfo {
	if cached, ok := fieldCache.Load(t); ok {
		return cached.(map[string]fieldInfo)
	}

	fields := make(map[string]fieldInfo)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tagOpts := getTagOptions(f.Tag.Get(TagName))
		fields[f.Name] = fieldInfo{Name: f.Name, TagOptions: tagOpts, FieldType: f.Type, Index: i}

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			embedded := getCachedFields(f.Type)
			for k, v := range embedded {
				fields[k] = v
			}
		}
	}

	fieldCache.Store(t, fields)
	return fields
}

func getCachedMethods(t reflect.Type) map[string]reflect.Method {
	if cached, ok := methodCache.Load(t); ok {
		return cached.(map[string]reflect.Method)
	}

	methods := make(map[string]reflect.Method)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		methods[m.Name] = m
	}
	methodCache.Store(t, methods)
	return methods
}

// ------------------ 工具 ------------------

func getTagOptions(tag string) TagOptions {
	options := make(TagOptions)
	for _, part := range strings.Split(tag, ";") {
		kv := strings.SplitN(part, ":", 2)
		key := strings.TrimSpace(kv[0])
		if key == "" {
			continue
		}
		if len(kv) == 2 {
			options[key] = strings.TrimSpace(kv[1])
		} else {
			options[key] = ""
		}
	}
	return options
}

func isNullableType(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem())
}

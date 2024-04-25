// Package protogofakeit provides a utility for producing fake data into
// Protocol Buffer messages.
package protogofakeit

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	pb "github.com/rodaine/protogofakeit/gen/gofakeit"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultMaxDepth = 5
	defaultMinSize  = 4
	defaultMaxSize  = 10

	wktTimestampFQN = "google.protobuf.Timestamp"
	wktDurationFQN  = "google.protobuf.Duration"
)

// ProtoFaker populates a protobuf message with fake data.
type ProtoFaker interface {
	// FakeProto populates msg with fake data, optionally configured through
	// annotations on the protobuf message. An error is returned if the
	// configuration on msg is invalid (typically a parse error).
	FakeProto(msg proto.Message) error
}

// New creates a [ProtoFaker] from the given gofakeit.Faker and [Option] values.
// If the returned value is intended to be used in a concurrent context, the
// provided faker should also be configured for concurrent use.
func New(faker *gofakeit.Faker, options ...Option) ProtoFaker {
	defaultSize := size{min: defaultMinSize, max: defaultMaxSize}
	pfaker := &protoFaker{
		faker:           faker,
		tplOptions:      nil,
		maxDepth:        defaultMaxDepth,
		stringSize:      defaultSize,
		bytesSize:       defaultSize,
		listSize:        defaultSize,
		mapSize:         defaultSize,
		timestampFormat: time.RFC3339Nano,
	}
	for _, opt := range options {
		opt.apply(pfaker)
	}
	return pfaker
}

// An Option modifies the default behavior of a [ProtoFaker], configured via
// [New].
type Option interface {
	apply(pf *protoFaker)
}

// WithMaxDepth limits the recursion depth of message fields populated. Beyond
// this depth, message fields are not set. The default value is 5.
func WithMaxDepth(depth int) Option {
	return optionFunc(func(pf *protoFaker) { pf.maxDepth = depth })
}

// WithTemplateOptions allows specifying any custom data and functions that
// should be available to fields constructed via templates. The default is unset.
func WithTemplateOptions(opts *gofakeit.TemplateOptions) Option {
	return optionFunc(func(pf *protoFaker) { pf.tplOptions = opts })
}

// WithStringSize sets the default minimum and maximum lengths (inclusive) of
// random strings produced. Individual lengths can be configured on a per-field
// basis within the protobuf definition. The default range is 4-10.
func WithStringSize(min, max int) Option {
	return optionFunc(func(pf *protoFaker) {
		pf.stringSize = size{min: min, max: max}
	})
}

// WithBytesSize sets the default minimum and maximum lengths (inclusive) of
// random bytes produced. Individual lengths can be configured on a per-field
// basis within the protobuf definition. The default range is 4-10.
func WithBytesSize(min, max int) Option {
	return optionFunc(func(pf *protoFaker) {
		pf.bytesSize = size{min: min, max: max}
	})
}

// WithListSize sets the default minimum and maximum lengths (inclusive) of
// random repeated fields produced. Individual lengths can be configured on a
// per-field basis within the protobuf definition. The default range is 4-10.
func WithListSize(min, max int) Option {
	return optionFunc(func(pf *protoFaker) {
		pf.listSize = size{min: min, max: max}
	})
}

// WithMapSize sets the default minimum and maximum lengths (inclusive) of
// random map fields produced. Individual lengths can be configured on a
// per-field basis within the protobuf definition. The default range is 4-10.
func WithMapSize(min, max int) Option {
	return optionFunc(func(pf *protoFaker) {
		pf.mapSize = size{min: min, max: max}
	})
}

// WithTimestampFormat sets the format used to parse timestamps from tag or
// template results. The format uses the same structure as [time.Layout]. The
// default format is [time.RFC3339Nano].
func WithTimestampFormat(format string) Option {
	return optionFunc(func(pf *protoFaker) {
		pf.timestampFormat = format
	})
}

type protoFaker struct {
	faker           *gofakeit.Faker
	tplOptions      *gofakeit.TemplateOptions
	maxDepth        int
	stringSize      size
	bytesSize       size
	listSize        size
	mapSize         size
	timestampFormat string
}

// FakeProto populates msg with fake data, optionally configured through
// annotations on the protobuf message. An error is returned if the
// configuration on msg is invalid (typically a parse error).
func (pf *protoFaker) FakeProto(msg proto.Message) error {
	return pf.fake(0, msg.ProtoReflect())
}

func (pf *protoFaker) fake(depth int, msg protoreflect.Message) error {
	desc := msg.Descriptor()
	if err := pf.fakeOneofs(depth, msg, desc); err != nil {
		return err
	}
	return pf.fakeFields(depth, msg, desc)
}

func (pf *protoFaker) fakeOneofs(
	depth int,
	msg protoreflect.Message,
	desc protoreflect.MessageDescriptor,
) error {
	oneofs := desc.Oneofs()
	for i, n := 0, oneofs.Len(); i < n; i++ {
		oneof := oneofs.Get(i)
		fields := oneof.Fields()
		idx := pf.faker.Rand.Intn(fields.Len()+1) - 1
		if idx == -1 {
			if field := msg.WhichOneof(oneof); field != nil {
				msg.Clear(field)
			}
			continue
		}
		if err := pf.fakeField(depth, msg, fields.Get(idx)); err != nil {
			return err
		}
	}
	return nil
}

func (pf *protoFaker) fakeFields(
	depth int,
	msg protoreflect.Message,
	desc protoreflect.MessageDescriptor,
) error {
	fields := desc.Fields()
	for i, n := 0, fields.Len(); i < n; i++ {
		fdesc := fields.Get(i)
		if fdesc.ContainingOneof() != nil {
			continue
		}
		if err := pf.fakeField(depth, msg, fdesc); err != nil {
			return err
		}
	}
	return nil
}

func (pf *protoFaker) fakeField(
	depth int,
	msg protoreflect.Message,
	desc protoreflect.FieldDescriptor,
) error {
	gen, _ := proto.GetExtension(desc.Options(), pb.E_Generate).(*pb.Generator)
	if gen.GetSkip() {
		return nil
	}
	val, err := pf.fakeFieldValue(depth, msg.NewField(desc), desc, gen, false)
	if err == nil && val.IsValid() {
		msg.Set(desc, val)
	}
	return err
}

func (pf *protoFaker) fakeFieldValue(
	depth int,
	val protoreflect.Value,
	desc protoreflect.FieldDescriptor,
	gen *pb.Generator,
	item bool,
) (protoreflect.Value, error) {
	switch {
	case gen.GetSkip():
		return val, nil
	case desc.IsMap():
		return val, pf.fakeMap(depth, desc, gen, val.Map())
	case desc.IsList() && !item:
		return val, pf.fakeList(depth, desc, gen, val.List())
	case desc.Kind() == protoreflect.MessageKind,
		desc.Kind() == protoreflect.GroupKind:
		switch desc.Message().FullName() {
		case wktTimestampFQN,
			wktDurationFQN:
			return pf.fakeScalar(desc, gen)
		default:
			if depth+1 >= pf.maxDepth {
				return protoreflect.Value{}, nil
			}
			return val, pf.fake(depth+1, val.Message())
		}
	default:
		return pf.fakeScalar(desc, gen)
	}
}

type sized interface {
	GetLen() uint32
	GetRange() *pb.Range
}

func (pf *protoFaker) fakeSize(msg sized, hasSizeOneof bool, def size) int {
	if !hasSizeOneof {
		return def.Fake(pf.faker)
	}
	if rng := msg.GetRange(); rng != nil {
		return pf.faker.IntRange(int(rng.GetMin()), int(rng.GetMax()))
	}
	return int(msg.GetLen())
}

func (pf *protoFaker) fakeMap(
	depth int,
	desc protoreflect.FieldDescriptor,
	gen *pb.Generator,
	mapVal protoreflect.Map,
) (err error) {
	mapExt := gen.GetMap()
	length := pf.fakeSize(mapExt, mapExt.GetSize() != nil, pf.mapSize)
	if length == 0 {
		return nil
	}

	kDesc, vDesc := desc.MapKey(), desc.MapValue()
	kGen, vGen := mapExt.GetKey(), mapExt.GetValue()
	if vGen == nil {
		vGen = gen
	}

	for i := 0; i < length; i++ {
		key := kDesc.Default()
		if !kGen.GetSkip() {
			key, err = pf.fakeScalar(kDesc, kGen)
			if err != nil {
				return err
			}
		}
		val := mapVal.NewValue()
		val, err = pf.fakeFieldValue(depth, val, vDesc, vGen, true)
		if err != nil {
			return err
		} else if val.IsValid() {
			mapVal.Set(key.MapKey(), val)
		}
	}
	return nil
}

func (pf *protoFaker) fakeList(
	depth int,
	desc protoreflect.FieldDescriptor,
	gen *pb.Generator,
	list protoreflect.List,
) error {
	listExt := gen.GetRepeated()
	length := pf.fakeSize(listExt, listExt.GetSize() != nil, pf.listSize)
	if length == 0 {
		return nil
	}

	if elGen := listExt.GetElement(); elGen != nil {
		gen = elGen
	}

	for i := 0; i < length; i++ {
		val, err := pf.fakeFieldValue(depth, list.NewElement(), desc, gen, true)
		if err != nil {
			return err
		} else if val.IsValid() {
			list.Append(val)
		}
	}
	return nil
}

func (pf *protoFaker) fakeScalar(
	desc protoreflect.FieldDescriptor,
	gen *pb.Generator,
) (val protoreflect.Value, err error) {
	switch {
	case gen.GetTag() != "":
		s := pf.faker.Generate(gen.GetTag())
		return pf.fakeParse(desc, s)
	case gen.GetTemplate() != "":
		s, err := pf.faker.Template(gen.GetTemplate(), pf.tplOptions)
		if err != nil {
			return val, err
		}
		return pf.fakeParse(desc, s)
	default:
		return pf.fakeFieldDefault(desc), nil
	}
}

//nolint:cyclop
func (pf *protoFaker) fakeParse(desc protoreflect.FieldDescriptor, str string) (val protoreflect.Value, err error) {
	if desc.Kind() == protoreflect.MessageKind {
		switch desc.Message().FullName() {
		case wktTimestampFQN:
			ts, err := time.Parse(pf.timestampFormat, str)
			return protoreflect.ValueOfMessage(timestamppb.New(ts).ProtoReflect()), err
		case wktDurationFQN:
			dur, err := time.ParseDuration(str)
			return protoreflect.ValueOfMessage(durationpb.New(dur).ProtoReflect()), err
		}
	}

	switch desc.Kind() {
	case protoreflect.BoolKind:
		b, err := strconv.ParseBool(str)
		return protoreflect.ValueOf(b), err
	case protoreflect.EnumKind:
		n, err := strconv.ParseInt(str, 0, 32)
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(n)), err
	case protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind:
		n, err := strconv.ParseInt(str, 0, 32)
		return protoreflect.ValueOfInt32(int32(n)), err
	case protoreflect.Uint32Kind,
		protoreflect.Fixed32Kind:
		n, err := strconv.ParseUint(str, 0, 32)
		return protoreflect.ValueOfUint32(uint32(n)), err
	case protoreflect.Int64Kind,
		protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind:
		n, err := strconv.ParseInt(str, 0, 64)
		return protoreflect.ValueOfInt64(n), err
	case protoreflect.Uint64Kind,
		protoreflect.Fixed64Kind:
		n, err := strconv.ParseUint(str, 0, 64)
		return protoreflect.ValueOfUint64(n), err
	case protoreflect.FloatKind:
		f, err := strconv.ParseFloat(str, 32)
		return protoreflect.ValueOfFloat32(float32(f)), err
	case protoreflect.DoubleKind:
		f, err := strconv.ParseFloat(str, 64)
		return protoreflect.ValueOfFloat64(f), err
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(str), nil
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte(str)), nil
	case protoreflect.MessageKind,
		protoreflect.GroupKind:
		fallthrough
	default:
		return val, fmt.Errorf("unknown/unexpected kind: %d", desc.Kind())
	}
}

//nolint:cyclop
func (pf *protoFaker) fakeFieldDefault(desc protoreflect.FieldDescriptor) (val protoreflect.Value) {
	if desc.Kind() == protoreflect.MessageKind {
		switch desc.Message().FullName() {
		case wktTimestampFQN:
			ts := pf.faker.Date()
			return protoreflect.ValueOfMessage(timestamppb.New(ts).ProtoReflect())
		case wktDurationFQN:
			dur := time.Duration(pf.faker.Int64())
			return protoreflect.ValueOfMessage(durationpb.New(dur).ProtoReflect())
		}
	}

	switch desc.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOf(pf.faker.Bool())
	case protoreflect.EnumKind:
		values := desc.Enum().Values()
		i := pf.faker.IntRange(0, values.Len()-1)
		return protoreflect.ValueOfEnum(values.Get(i).Number())
	case protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(pf.faker.Int32())
	case protoreflect.Uint32Kind,
		protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(pf.faker.Uint32())
	case protoreflect.Int64Kind,
		protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(pf.faker.Int64())
	case protoreflect.Uint64Kind,
		protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(pf.faker.Uint64())
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(pf.faker.Float32())
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(pf.faker.Float64())
	case protoreflect.StringKind:
		s := pf.faker.Generate(strings.Repeat("?", pf.stringSize.Fake(pf.faker)))
		return protoreflect.ValueOfString(s)
	case protoreflect.BytesKind:
		b := make([]byte, pf.bytesSize.Fake(pf.faker))
		_, _ = pf.faker.Rand.Read(b)
		return protoreflect.ValueOfBytes(b)
	case protoreflect.MessageKind,
		protoreflect.GroupKind:
		fallthrough
	default:
		return val
	}
}

type size struct {
	min, max int
}

func (s size) Fake(faker *gofakeit.Faker) int {
	return faker.Number(s.min, s.max)
}

type optionFunc func(pf *protoFaker)

func (fn optionFunc) apply(pf *protoFaker) { fn(pf) }

package protogofakeit

import (
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/rodaine/protogofakeit/gen/gofakeit/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestProtoFaker(t *testing.T) {
	t.Parallel()

	t.Run("scalars", func(t *testing.T) {
		t.Parallel()

		t.Run("default", func(t *testing.T) {
			t.Parallel()
			msg := &test.ScalarDefaults{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
		})

		t.Run("static_tags", func(t *testing.T) {
			t.Parallel()
			msg := &test.ScalarStaticTags{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)

			ex := &test.ScalarStaticTags{
				Bool:     true,
				Enum:     test.Enum_ENUM_BETA,
				Int32:    123,
				Sint32:   -456,
				Sfixed32: 789,
				Uint32:   1011,
				Fixed32:  1213,
				Int64:    1415,
				Sint64:   -1617,
				Sfixed64: 1819,
				Uint64:   2021,
				Fixed64:  2223,
				Float:    1.23,
				Double:   -4.56,
				String_:  "foobar",
				Bytes:    []byte("fizzbuzz"),
			}
			assert.True(t, proto.Equal(ex, msg))
		})

		t.Run("tags", func(t *testing.T) {
			t.Parallel()
			msg := &test.ScalarTags{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
		})

		t.Run("static_templates", func(t *testing.T) {
			t.Parallel()
			msg := &test.ScalarStaticTemplates{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)

			ex := &test.ScalarStaticTemplates{
				Bool:     true,
				Enum:     test.Enum_ENUM_BETA,
				Int32:    123,
				Sint32:   -456,
				Sfixed32: 789,
				Uint32:   1011,
				Fixed32:  1213,
				Int64:    1415,
				Sint64:   -1617,
				Sfixed64: 1819,
				Uint64:   2021,
				Fixed64:  2223,
				Float:    1.23,
				Double:   -4.56,
				String_:  "foobar",
				Bytes:    []byte("fizzbuzz"),
			}
			assert.True(t, proto.Equal(ex, msg))
		})

		t.Run("skip", func(t *testing.T) {
			t.Parallel()
			msg := &test.ScalarSkip{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			assert.True(t, proto.Equal(msg, &test.ScalarSkip{}))
		})

		t.Run("custom_sizes", func(t *testing.T) {
			t.Parallel()

			msg := &test.ScalarDefaults{}
			err := initProtoFaker(t,
				WithStringSize(3, 3),
				WithBytesSize(12, 12),
			).FakeProto(msg)
			require.NoError(t, err)
			assert.Len(t, msg.GetString_(), 3)
			assert.Len(t, msg.GetBytes(), 12)
		})
	})

	t.Run("messages", func(t *testing.T) {
		t.Parallel()

		t.Run("self_recursive", func(t *testing.T) {
			t.Parallel()
			msg := &test.SelfRecursive{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			for i := 0; i < defaultMaxDepth; i++ {
				assert.NotNil(t, msg, "message should not be nil at index", i)
				assert.NotEmpty(t, msg.GetFoo(), "message field should be faked at index", i)
				msg = msg.GetRecurse()
			}
			assert.Nil(t, msg, "message should be nil at max depth", defaultMaxDepth)
		})

		t.Run("pair_recursive", func(t *testing.T) {
			t.Parallel()
			msgA := &test.PairRecursive_A{}
			err := initProtoFaker(t, WithMaxDepth(6)).FakeProto(msgA)
			require.NoError(t, err)
			for i := 0; i < 6; i += 2 {
				assert.NotNil(t, msgA, "message A should not be nil at index", i)
				assert.NotNil(t, msgA.GetB(), "message B should not be nil at index", i+1)
				msgA = msgA.GetB().GetA()
			}
			assert.Nil(t, msgA, "message should be nil at max depth", 6)
		})

		t.Run("skip", func(t *testing.T) {
			t.Parallel()
			msg := &test.MessageSkipped{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			assert.Nil(t, msg.GetSkipped())
		})
	})

	t.Run("repeated", func(t *testing.T) {
		t.Parallel()

		t.Run("default", func(t *testing.T) {
			t.Parallel()
			msg := &test.RepeatedDefaults{}
			err := initProtoFaker(t, WithMaxDepth(3)).FakeProto(msg)
			require.NoError(t, err)

			sliceInDefault(t, msg.GetScalars())
			for _, s := range msg.GetScalars() {
				assert.NotEmpty(t, s)
			}

			sliceInDefault(t, msg.GetEnums())
			for _, e := range msg.GetEnums() {
				_, ok := test.RepeatedEnum_name[int32(e)]
				assert.True(t, ok, e.String())
			}

			sliceInDefault(t, msg.GetMessages())
			for _, m := range msg.GetMessages() {
				assert.NotEmpty(t, m.GetFoo())
			}

			sliceInDefault(t, msg.GetRecursive())
			sliceInDefault(t, msg.GetRecursive()[0].GetRecursive())
			assert.Empty(t, msg.GetRecursive()[0].GetRecursive()[0].GetRecursive())
		})

		t.Run("tags", func(t *testing.T) {
			t.Parallel()
			msg := &test.RepeatedTags{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			sliceInDefault(t, msg.GetFoo())
			for _, s := range msg.GetFoo() {
				assert.Equal(t, "bar", s)
			}
		})

		t.Run("length", func(t *testing.T) {
			t.Parallel()
			msg := &test.RepeatedLength{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			sliceIn(t, msg.GetScalars(), 3, 3)
			for _, s := range msg.GetScalars() {
				assert.NotEmpty(t, s)
			}
		})

		t.Run("range", func(t *testing.T) {
			t.Parallel()
			msg := &test.RepeatedRange{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			sliceIn(t, msg.GetScalars(), 1, 2)
			for _, s := range msg.GetScalars() {
				assert.NotEmpty(t, s)
			}
		})

		t.Run("skip", func(t *testing.T) {
			t.Parallel()
			msg := &test.RepeatedSkip{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			sliceInDefault(t, msg.GetScalars())
			for _, s := range msg.GetScalars() {
				assert.Empty(t, s)
			}
			assert.Nil(t, msg.GetSkipped())
		})

		t.Run("custom_sizes", func(t *testing.T) {
			t.Parallel()
			msg := &test.RepeatedDefaults{}
			err := initProtoFaker(t, WithListSize(3, 3)).FakeProto(msg)
			require.NoError(t, err)
			assert.Len(t, msg.GetScalars(), 3)
			assert.Len(t, msg.GetEnums(), 3)
			assert.Len(t, msg.GetMessages(), 3)
			assert.Len(t, msg.GetRecursive(), 3)

			msg = &test.RepeatedDefaults{}
			err = initProtoFaker(t, WithListSize(0, 0)).FakeProto(msg)
			require.NoError(t, err)
			assert.Empty(t, msg.GetScalars())
			assert.Empty(t, msg.GetEnums())
			assert.Empty(t, msg.GetMessages())
			assert.Empty(t, msg.GetRecursive())
		})
	})

	t.Run("map", func(t *testing.T) {
		t.Parallel()

		t.Run("default", func(t *testing.T) {
			t.Parallel()

			msg := &test.MapDefaults{}
			err := initProtoFaker(t, WithMaxDepth(3)).FakeProto(msg)
			require.NoError(t, err)

			mapInDefault(t, msg.GetScalars())
			for k := range msg.GetScalars() {
				assert.NotEmpty(t, k)
			}

			mapInDefault(t, msg.GetEnums())
			for k, v := range msg.GetEnums() {
				assert.NotEmpty(t, k)
				_, ok := test.MapEnum_name[int32(v)]
				assert.True(t, ok, v.String())
			}

			mapInDefault(t, msg.GetMessages())
			for k, v := range msg.GetMessages() {
				assert.NotEmpty(t, k)
				assert.NotEmpty(t, v.GetFoo())
			}

			mapInDefault(t, msg.GetRecursive())
		recursive:
			for _, v := range msg.GetRecursive() {
				mapInDefault(t, v.GetRecursive())
				for _, vv := range v.GetRecursive() {
					assert.Empty(t, vv.GetRecursive())
					break recursive
				}
			}
		})

		t.Run("tags", func(t *testing.T) {
			t.Parallel()
			msg := &test.MapTags{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)

			mapIn(t, msg.GetValues(), 1, 1) // since the key is static
			assert.Equal(t, "bar", msg.GetValues()["foo"])
		})

		t.Run("length", func(t *testing.T) {
			t.Parallel()
			msg := &test.MapLength{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			mapIn(t, msg.GetValues(), 3, 3)
			for k, v := range msg.GetValues() {
				assert.NotEmpty(t, k)
				assert.NotEmpty(t, v)
			}
		})

		t.Run("range", func(t *testing.T) {
			t.Parallel()
			msg := &test.MapRange{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			mapIn(t, msg.GetValues(), 1, 2)
			for k, v := range msg.GetValues() {
				assert.NotEmpty(t, k)
				assert.NotEmpty(t, v)
			}
		})

		t.Run("skip", func(t *testing.T) {
			t.Parallel()
			msg := &test.MapSkip{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			mapIn(t, msg.GetValues(), 1, 1) // key should always be ""
			for k, v := range msg.GetValues() {
				assert.Equal(t, "", k)
				assert.Equal(t, "", v)
			}
			assert.Nil(t, msg.GetSkipped())
		})

		t.Run("custom_sizes", func(t *testing.T) {
			t.Parallel()
			msg := &test.MapDefaults{}
			err := initProtoFaker(t, WithMapSize(3, 3)).FakeProto(msg)
			require.NoError(t, err)
			assert.Len(t, msg.GetScalars(), 3)
			assert.Len(t, msg.GetEnums(), 3)
			assert.Len(t, msg.GetMessages(), 3)
			assert.Len(t, msg.GetRecursive(), 3)

			msg = &test.MapDefaults{}
			err = initProtoFaker(t, WithMapSize(0, 0)).FakeProto(msg)
			require.NoError(t, err)
			assert.Empty(t, msg.GetScalars())
			assert.Empty(t, msg.GetEnums())
			assert.Empty(t, msg.GetMessages())
			assert.Empty(t, msg.GetRecursive())
		})
	})

	t.Run("oneof", func(t *testing.T) {
		t.Parallel()
		msg := &test.OneOf{}

		// scalar
		err := initProtoFaker(t, withSeed(1)).FakeProto(msg)
		require.NoError(t, err)
		assert.NotNil(t, msg.GetFields())
		assert.NotEmpty(t, msg.GetScalar())

		// enum
		err = initProtoFaker(t, withSeed(5)).FakeProto(msg)
		require.NoError(t, err)
		assert.NotNil(t, msg.GetFields())
		assert.Equal(t, test.OneOfEnum_ONE_OF_ENUM_TWO, msg.GetEnum())

		// message
		err = initProtoFaker(t, WithMaxDepth(2), withSeed(43)).FakeProto(msg)
		require.NoError(t, err)
		assert.NotNil(t, msg.GetFields())
		assert.NotNil(t, msg.GetMessage().GetFields())

		// unset
		err = initProtoFaker(t, withSeed(3)).FakeProto(msg)
		require.NoError(t, err)
		assert.Nil(t, msg.GetFields())

		t.Run("issue_5", func(t *testing.T) {
			t.Parallel()
			msg := &test.OneOfMessages{}
			err = initProtoFaker(t, withSeed(5)).FakeProto(msg)
			require.NoError(t, err)
			assert.Nil(t, msg.GetKind())
		})
	})

	t.Run("wkt", func(t *testing.T) {
		t.Parallel()

		t.Run("timestamp", func(t *testing.T) {
			t.Parallel()

			msg := &test.WKTTimestamp{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			assert.True(t, msg.GetDefaultTs().IsValid())
			assert.True(t, msg.GetTag().IsValid())

			custom := &test.WKTTimestampCustom{}
			err = initProtoFaker(t,
				WithTimestampFormat("Jan _2 2006 15:04:05"),
			).FakeProto(custom)
			require.NoError(t, err)
			assert.True(t, custom.GetValue().IsValid())
		})

		t.Run("duration", func(t *testing.T) {
			t.Parallel()

			msg := &test.WKTDuration{}
			err := initProtoFaker(t).FakeProto(msg)
			require.NoError(t, err)
			assert.True(t, msg.GetDefaultDur().IsValid())
			assert.True(t, msg.GetTag().IsValid())
		})
	})
}

func TestWithTemplateOptions(t *testing.T) {
	t.Parallel()

	msg := &test.CustomTemplate{}
	err := initProtoFaker(t).FakeProto(msg)
	require.Error(t, err)

	tplOpts := &gofakeit.TemplateOptions{
		Funcs: map[string]any{
			"custom_func": func() string { return "foo" },
		},
	}
	err = initProtoFaker(t, WithTemplateOptions(tplOpts)).FakeProto(msg)
	require.NoError(t, err)
	assert.Equal(t, "foo", msg.GetValue())
}

func withSeed(seed int64) optionFunc {
	return func(pf *protoFaker) {
		pf.faker.Rand.Seed(seed)
	}
}

func mapInDefault[K comparable, V any](tb testing.TB, m map[K]V) bool {
	tb.Helper()
	return mapIn(tb, m, defaultMinSize, defaultMaxSize)
}

func mapIn[K comparable, V any](tb testing.TB, m map[K]V, min, max int) bool {
	tb.Helper()
	n := len(m)
	return inRange(tb, n, min, max)
}

func sliceInDefault[T any](tb testing.TB, s []T) bool {
	tb.Helper()
	return sliceIn(tb, s, defaultMinSize, defaultMaxSize)
}

func sliceIn[T any](tb testing.TB, s []T, min, max int) bool {
	tb.Helper()
	n := len(s)
	return inRange(tb, n, min, max)
}

func inRange(tb testing.TB, n, min, max int) bool {
	tb.Helper()
	return assert.GreaterOrEqual(tb, n, min, "minimum") &&
		assert.LessOrEqual(tb, n, max, "maximum")
}

func initProtoFaker(tb testing.TB, opts ...Option) ProtoFaker {
	tb.Helper()
	var seed int64
	if envSeed := os.Getenv("SEED"); envSeed != "" {
		s, err := strconv.ParseInt(envSeed, 0, 64)
		require.NoError(tb, err)
		seed = s
	} else {
		seed = rand.Int63() //nolint:gosec // fine for testing
	}
	tb.Log("seed: ", seed)
	// we use unlocked intentionally as tests should be reproducible (i.e., not
	// used concurrently)
	faker := gofakeit.NewUnlocked(seed)
	return New(faker, opts...)
}

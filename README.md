# protogofakeit

`protogofakeit` provides a utility for producing fake data in Protocol Buffer 
messages via [`gofakeit`][gofakeit].

## Example

`protogofakeit` works by annotating fields of protobuf messages with the 
`(gofakeit.generate)` custom options. Below, an example `User` and related 
messages have been annotated:

```protobuf
syntax="proto3";
import "gofakeit/gofakeit.proto";

message User {
  fixed64 user_id = 1;
  string first_name = 2 [(gofakeit.generate).tag = "{firstname}"];
  string last_name = 3 [(gofakeit.generate).tag = "{lastname}"];
  string email = 4 [(gofakeit.generate).tag = "{email}"];
  Location location = 5;
  repeated string hobbies = 6 [
    (gofakeit.generate).repeated.range = { min: 1, max: 3},
    (gofakeit.generate).repeated.element.tag = "{hobby}"
  ];
  map<string, PetType> pets = 7 [
    (gofakeit.generate).map.range = { min: 1, max: 3},
    (gofakeit.generate).map.key.tag = "{petname}"
  ];
}

enum PetType {
  PET_TYPE_UNSPECIFIED = 0;
  PET_TYPE_DOG = 1;
  PET_TYPE_CAT = 2;
}

message Location {
  double latitude = 1 [(gofakeit.generate).tag = "{latitude}"];
  double longitude = 2 [(gofakeit.generate).tag = "{longitude}"];
}
```

`protogofakeit` does **NOT** require any code generation (beyond `protoc-gen-go`) to populate a message:

```go
package main

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/rodaine/protogofakeit"
	"github.com/rodaine/protogofakeit/example/gen"
	"google.golang.org/protobuf/encoding/prototext"
)

func main() {
	faker := gofakeit.New(3)
	protoFaker := protogofakeit.New(faker)
	user := &gen.User{}
	protoFaker.FakeProto(user)
	fmt.Println(prototext.Format(user))
}
```

Outputs:

```text
user_id: 15864040051628833669
first_name: "Philip"
last_name: "Casper"
email: "tyreekstroman@mitchell.biz"
location: {
  latitude: 1.383491
  longitude: -62.968324
}
hobbies: [
  "Curling"
  "Slot car"
  "Fishkeeping"
]
pets: {
  "Archie": PET_TYPE_DOG
}
```

## Annotating Messages

The `protogofakeit` proto files must be imported into your proto files to set 
the custom options on fields.

### protoc CLI

If you are using [`protoc`][protoc], the `proto` directory must be in the import path to compile:

```shell
protoc \
  -I "$PROTOGOFAKEIT_PATH/proto" \
  --go_out="$OUTPUT_DIR" \
  $PROTOS
```

### buf CLI

If you are using [`buf`][buf], add a dependency on `buf.build/rodaine/protogofakeit` 
to your module's [`buf.yaml`][buf.yaml].

```yaml
# buf.yaml
version: v1
deps:
  - buf.build/rodaine/protogofakeit
# ...
```

Remember to run `buf mod update` to update the lockfile. If you are using buf 
[managed mode][managed] for code generation, ensure the correct import path is used for
`protogofakeit`:

```yaml
# buf.gen.yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    override:
      buf.build/rodaine/protogofakeit: github.com/rodaine/protogofakeit/gen
# ...
```

### Default Generation

By default, `protogofakeit` populates all fields of a message with random values:

- **numbers**: a uniform, random value in the full range of the underlying type.
- **bool**: a random `true` or `false`
- **string**: a random ASCII string between 4–10 characters, inclusive.
- **bytes**: random slice of 4–10 bytes (inclusive) in the ASCII range.
- **enums**: a random defined value of the enum (including any value assigned to 0).
- **messages**: the field is set and its fields are populated with random data, 
  to the max recursion depth (default of 5).
- **repeated**: a slice of 4–10 (inclusive) elements populated with random values.
- **map**: a map of 4–10 (inclusive) key-value pairs populated with random keys 
  and values. Note that the total pairs may be less if there is a key collision 
  during generation.
- **oneof**: a random field (or no field) contained in the oneof will be set to 
  its random value.
- **optional**: optional fields are always set.
- **google.protobuf.Timestamp**: a random valid timestamp value
- **google.protobuf.Duration**: a random valid duration value

Default sizes of string, bytes, repeated, and map fields as well as the maximum 
recursion depth can be customized when initializing the `ProtoFaker` instance 
via `Option` values. See the documentation for more details. 

### Skip

Fake generation can be skipped on any field by annotating it with `skip`:

```protobuf
message Skipped {
  string value = 1 [(gofakeit.generate).skip = true];
}
```

Scalar values will default to their zero value, while message, repeated, and map
fields will be empty/unset.

### Tags

The primary way of customizing field generation is via tags, which are identical
to the [struct tags][struct] used to customize plain-old Go structs for `gofakeit`.
Tags act as simplified templates using curly braces (`{...}`) to specify "function" 
calls to the faker.

After resolving any functions, the `tag` value is then parsed to match the type of 
the associated field. Any parse errors result in fake generation to fail:

- **integers**: either `strconv.ParseInt(tag, 0, n)` or `strconv.ParseUint(tag, 0, n)` 
  if unsigned, where `n` is the bitsize of the field.
- **floats**: `strconv.ParseFloat(tag, n)`, where `n` is the bitsize of the field.
- **bool**: `strconv.ParseBool(tag)`
- **string**: as-is
- **bytes**: `[]byte(tag)`
- **enums**: `strconv.ParseInt(tag, 0, 32)`. Undefined enum values are supported/possible.
- **google.protobuf.Timestamp**: `time.Parse(format, tag)`, where `format` is 
  a `time.Layout` based format (default of `time.RFC3339Nano`).
- **google.protobuf.Duration**: `time.ParseDuration(tag)`

Tags are ignored on message, repeated, and map fields. The timestamp parse format
can be customized when initializing the `ProtoFaker` instance.

```protobuf
message Tags {
  string name = 1 [(gofakeit.generate).tag = "{firstname}"]; // random
  string foo = 2 [(gofakeit.generate).tag = "bar"]; // constant "bar"
  int32 age = 3 [(gofakeit.generate).tag = "{intrange:13,99}"]; // function parameters
  string pet = 4[(gofakeit.generate).tag = "{petname} ({animal})"]; // mixed
}
```

[Custom tag functions][custom] can be registered on the underlying faker, though 
this is strongly discouraged as it makes the tags less portable.

### Templates

For more complex fake generation, Go [templates] are supported, enabling loops, 
conditionals, and other logic alongside the functions available from the 
underlying faker. The resulting value is parsed the same as tags above. For long 
templates, a series of strings can be used which are automatically concatenated 
by the protobuf compiler. 

```protobuf
message Templates {
  string cmd = 1 [
    (gofakeit.generate).template = 
      '{{ $lang := RandomString (SliceString "go" "python" "javascript" }}'
      '{{ if eq $lang "go" }}go get'
      '{{ else if eq $lang "python" }}pip install'
      '{{ else }}npm install'
      '{{ end }} ({{ $lang }})'
  ];
}
```

Templates are orders of magnitude slower than tags, so only reach for them if 
you need something more powerful. Custom template functions can be registered 
when configuring the `ProtoFaker` instance, though this is strongly discouraged 
as it makes the templates less portable.

### Repeated / Map Fields

Repeated (list) and map fields can be customized beyond the defaults, including 
their size and items.

#### Size

A repeated or map field's size can be specified via either a constant `len` or a
random `range`.

```protobuf
message Sizes {
  repeated string a = 1 [
    (gofakeit.generate).repeated.len = 3 // exactly 3 elements
  ]; 
  map<string, int32> b = 2 [
    (gofakeit.generate).map.range = { min: 0, max: 2 } // 0-2 key-value pairs, inclusive.
  ];  
}
```

#### Items

A repeated field's elements or a map field's keys and values can have options 
associated with them just like singular fields. Note that message items still 
ignore any `tag` or `template` options. 

```protobuf
message Elements {
  repeated string a = 1 [
    (gofakeit.generate).repeated.element.tag = "{animal}"
  ];
  map<string, int32> b = 2 [
    (gofakeit.generate).map.key.tag = "{petname}",
    (gofakeit.generate).map.value.tag = "{intrange:0,100}"
  ];
}
```

[gofakeit]: https://github.com/brianvoe/gofakeit
[protoc]: https://protobuf.dev/programming-guides/proto3/#generating
[buf]: https://buf.build/docs/ecosystem/cli-overview
[buf.yaml]: https://buf.build/docs/configuration/v1/buf-yaml
[managed]: https://buf.build/docs/generate/managed-mode
[struct]: https://github.com/brianvoe/gofakeit/tree/master#struct
[custom]: https://github.com/brianvoe/gofakeit/tree/master#custom-functions
[templates]: https://github.com/brianvoe/gofakeit/tree/master#templates
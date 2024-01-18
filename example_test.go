package protogofakeit_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/rodaine/protogofakeit"
	"github.com/rodaine/protogofakeit/gen/gofakeit/example"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Example() {
	faker := gofakeit.New(3)
	protoFaker := protogofakeit.New(faker)
	user := &example.User{}
	_ = protoFaker.FakeProto(user)
	fmt.Println(toJSON(user))

	// Output:
	// {
	//   "userId": "15864040051628833669",
	//   "firstName": "Philip",
	//   "lastName": "Casper",
	//   "email": "tyreekstroman@mitchell.biz",
	//   "location": {
	//     "latitude": 1.383491,
	//     "longitude": -62.968324
	//   },
	//   "hobbies": [
	//     "Curling",
	//     "Slot car",
	//     "Fishkeeping"
	//   ],
	//   "pets": {
	//     "Archie": "PET_TYPE_DOG"
	//   }
	// }
}

func toJSON(msg proto.Message) string {
	data, _ := protojson.Marshal(msg)
	buf := &bytes.Buffer{}
	_ = json.Indent(buf, data, "", "  ")
	return buf.String()
}

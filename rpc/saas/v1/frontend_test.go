package saasv1

import (
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestFrontendMetadataIsAppendOnly(t *testing.T) {
	for _, test := range []struct {
		message protoreflect.MessageDescriptor
		fields  map[protoreflect.Name]protoreflect.FieldNumber
	}{
		{(&LoginReply{}).ProtoReflect().Descriptor(), map[protoreflect.Name]protoreflect.FieldNumber{"frontend_entry": 10, "frontend_base_path": 11, "frontend_home": 12}},
		{(&App{}).ProtoReflect().Descriptor(), map[protoreflect.Name]protoreflect.FieldNumber{"frontend_base_path": 19, "frontend_home": 20}},
	} {
		for name, number := range test.fields {
			field := test.message.Fields().ByName(name)
			if field == nil || field.Number() != number || field.Kind() != protoreflect.StringKind {
				t.Fatalf("frontend field %s.%s has an incompatible wire definition", test.message.Name(), name)
			}
		}
	}
	raw, err := os.ReadFile("testdata/v1.0.3.protoset")
	if err != nil {
		t.Fatal(err)
	}
	var baseline descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(raw, &baseline); err != nil {
		t.Fatal(err)
	}
	oldFile, err := protodesc.NewFile(baseline.File[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	current := &LoginReply{AppCode: "notes", Home: "/notes/home", FrontendEntry: "independent", FrontendBasePath: "/notes-web", FrontendHome: "/dashboard"}
	wire, err := proto.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	old := dynamicpb.NewMessage(oldFile.Messages().ByName("LoginReply"))
	if err := proto.Unmarshal(wire, old); err != nil {
		t.Fatal("old client cannot read new login response", err)
	}
	if old.Get(old.Descriptor().Fields().ByName("home")).String() != current.Home {
		t.Fatal("new metadata changed the old client's home")
	}
	wire, err = proto.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	var restored LoginReply
	if err := proto.Unmarshal(wire, &restored); err != nil || !proto.Equal(current, &restored) {
		t.Fatal("old client did not preserve additive frontend fields", err)
	}
}

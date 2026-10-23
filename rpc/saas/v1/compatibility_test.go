package saasv1

import (
	"os"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestV103DescriptorRemainsCompatible(t *testing.T) {
	raw, err := os.ReadFile("testdata/v1.0.3.protoset")
	if err != nil {
		t.Fatal(err)
	}
	var baseline descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(raw, &baseline); err != nil || len(baseline.File) != 1 {
		t.Fatalf("invalid v1.0.3 descriptor fixture: %v", err)
	}
	current := protodesc.ToFileDescriptorProto(File_saas_v1_core_proto)
	old := baseline.File[0]
	if old.GetPackage() != current.GetPackage() || old.GetName() != current.GetName() {
		t.Fatal("v1 protobuf package or file path changed")
	}
	for _, previous := range old.GetMessageType() {
		present := messageByName(current.GetMessageType(), previous.GetName())
		if present == nil {
			t.Fatalf("v1 message %s removed", previous.GetName())
		}
		for index, oneof := range previous.GetOneofDecl() {
			if index >= len(present.GetOneofDecl()) || !proto.Equal(oneof, present.GetOneofDecl()[index]) {
				t.Fatalf("v1 oneof %s.%s changed", previous.GetName(), oneof.GetName())
			}
		}
		for _, field := range previous.GetField() {
			found := fieldByNumber(present.GetField(), field.GetNumber())
			if found == nil || !proto.Equal(field, found) {
				t.Fatalf("v1 field %s.%s (%d) changed", previous.GetName(), field.GetName(), field.GetNumber())
			}
		}
	}
	for _, previous := range old.GetEnumType() {
		found := enumByName(current.GetEnumType(), previous.GetName())
		if found == nil {
			t.Fatalf("v1 enum %s removed", previous.GetName())
		}
		for _, value := range previous.GetValue() {
			matched := false
			for _, candidate := range found.GetValue() {
				matched = matched || value.GetName() == candidate.GetName() && value.GetNumber() == candidate.GetNumber()
			}
			if !matched {
				t.Fatalf("v1 enum value %s.%s changed", previous.GetName(), value.GetName())
			}
		}
	}
	for _, previous := range old.GetService() {
		var found *descriptorpb.ServiceDescriptorProto
		for _, service := range current.GetService() {
			if service.GetName() == previous.GetName() {
				found = service
			}
		}
		if found == nil {
			t.Fatalf("v1 service %s removed", previous.GetName())
		}
		for _, method := range previous.GetMethod() {
			matched := false
			for _, candidate := range found.GetMethod() {
				matched = matched || method.GetName() == candidate.GetName() && method.GetInputType() == candidate.GetInputType() && method.GetOutputType() == candidate.GetOutputType() && method.GetClientStreaming() == candidate.GetClientStreaming() && method.GetServerStreaming() == candidate.GetServerStreaming()
			}
			if !matched {
				t.Fatalf("v1 method %s.%s changed", previous.GetName(), method.GetName())
			}
		}
	}
}

func messageByName(messages []*descriptorpb.DescriptorProto, name string) *descriptorpb.DescriptorProto {
	for _, message := range messages {
		if message.GetName() == name {
			return message
		}
	}
	return nil
}

func fieldByNumber(fields []*descriptorpb.FieldDescriptorProto, number int32) *descriptorpb.FieldDescriptorProto {
	for _, field := range fields {
		if field.GetNumber() == number {
			return field
		}
	}
	return nil
}

func enumByName(enums []*descriptorpb.EnumDescriptorProto, name string) *descriptorpb.EnumDescriptorProto {
	for _, enum := range enums {
		if enum.GetName() == name {
			return enum
		}
	}
	return nil
}

func TestV1ServiceMethodNamesRemainStable(t *testing.T) {
	service := File_saas_v1_core_proto.Services().ByName("Platform")
	for _, name := range []string{"SubmitApplicationModel", "QueryApplicationModel", "RebindApplicationModel", "Health"} {
		if service.Methods().ByName(protoreflect.Name(name)) == nil {
			t.Fatalf("v1 service method %s was removed", name)
		}
	}
}

func TestV1ModelCallFieldNumbersRemainStable(t *testing.T) {
	descriptor := File_saas_v1_core_proto.Messages().ByName("ModelCallReply")
	want := map[protoreflect.Name]protoreflect.FieldNumber{
		"call_id": 1, "state": 2, "result_json": 3, "usage_json": 4,
		"request_id": 5, "diagnostic": 7, "queue_phase": 10, "root_run_id": 11,
	}
	for name, number := range want {
		field := descriptor.Fields().ByName(name)
		if field == nil || field.Number() != number {
			t.Fatalf("field %s changed: got %v want %v", name, fieldNumber(field), number)
		}
	}
}

func TestV1DiagnosticAppendOnlyFieldsRemainStable(t *testing.T) {
	descriptor := File_saas_v1_core_proto.Messages().ByName("ModelCallDiagnostic")
	for name, number := range map[protoreflect.Name]protoreflect.FieldNumber{
		"retry_class": 29, "retry_after_seconds": 30, "retry_policy": 31, "execution_outcome": 32,
	} {
		field := descriptor.Fields().ByName(name)
		if field == nil || field.Number() != number {
			t.Fatalf("diagnostic field %s changed: got %v want %v", name, fieldNumber(field), number)
		}
	}
}

func TestV1ClientToleratesUnknownDiagnosticFields(t *testing.T) {
	b, err := proto.Marshal(&ModelCallDiagnostic{Code: "timeout", Phase: "connect"})
	if err != nil {
		t.Fatal(err)
	}
	b = protowire.AppendTag(b, 99, protowire.VarintType)
	b = protowire.AppendVarint(b, 1)
	var decoded ModelCallDiagnostic
	if err := proto.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.GetCode() != "timeout" || decoded.GetPhase() != "connect" {
		t.Fatalf("known fields were not preserved across unknown field: %#v", &decoded)
	}
}

func fieldNumber(field protoreflect.FieldDescriptor) protoreflect.FieldNumber {
	if field == nil {
		return 0
	}
	return field.Number()
}

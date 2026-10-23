package saasv1

import (
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

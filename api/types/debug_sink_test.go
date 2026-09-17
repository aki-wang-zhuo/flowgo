package types

import (
	"context"
	"testing"
)

func TestAppendDebugLog_noSink(t *testing.T) {
	// 无槽时不应 panic
	AppendDebugLog(context.Background(), DebugLog{FlowType: DebugFlowRequest, Data: "x"})
}

func TestAppendDebugLog_withSink(t *testing.T) {
	var sink []DebugLog
	ctx := WithDebugSink(context.Background(), &sink)
	if !HasDebugSink(ctx) {
		t.Fatal("want HasDebugSink")
	}
	AppendDebugLog(ctx, DebugLog{FlowType: DebugFlowRequest, Data: "req"})
	AppendDebugLog(ctx, DebugLog{FlowType: DebugFlowResponse, Data: "res"})
	if len(sink) != 2 {
		t.Fatalf("want 2, got %d", len(sink))
	}
	if sink[0].FlowType != DebugFlowRequest || sink[1].FlowType != DebugFlowResponse {
		t.Fatalf("unexpected %#v", sink)
	}
}

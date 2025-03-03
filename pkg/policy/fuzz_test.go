// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package policy

import (
	"testing"

	fuzz "github.com/AdaLogics/go-fuzz-headers"

	"github.com/cilium/cilium/pkg/container/versioned"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/policy/api"
	"github.com/cilium/cilium/pkg/policy/trafficdirection"
	"github.com/cilium/cilium/pkg/policy/types"
	"github.com/cilium/cilium/pkg/u8proto"
)

func FuzzResolveEgressPolicy(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		ff := fuzz.NewConsumer(data)
		label, err := ff.GetString()
		if err != nil {
			return
		}
		fromBar := &SearchContext{From: labels.ParseSelectLabelArray(label)}
		r := api.Rule{}
		err = ff.GenerateStruct(&r)
		if err != nil {
			return
		}
		err = r.Sanitize()
		if err != nil {
			return
		}
		rule := &rule{Rule: r}
		state := traceState{}
		td := newTestData()
		_, _ = rule.resolveEgressPolicy(td.testPolicyContext, fromBar, &state, NewL4PolicyMap(), nil, nil)

	})
}

func FuzzDenyPreferredInsert(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		keys := emptyMapState()
		key := Key{}
		entry := NewMapStateEntry(types.AllowEntry())
		ff := fuzz.NewConsumer(data)
		ff.GenerateStruct(&key)
		ff.GenerateStruct(&entry)
		keys.insertWithChanges(key, entry, allFeatures, ChangeState{})
	})
}

func FuzzAccumulateMapChange(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		ff := fuzz.NewConsumer(data)
		adds := make([]identity.NumericIdentity, 0)
		ff.CreateSlice(&adds)
		deletes := make([]identity.NumericIdentity, 0)
		ff.CreateSlice(&deletes)
		port, err := ff.GetUint16()
		if err != nil {
			t.Skip()
		}
		protoUint8, err := ff.GetByte()
		if err != nil {
			t.Skip()
		}
		proto := u8proto.U8proto(protoUint8)
		dir := trafficdirection.Ingress
		redirect, err := ff.GetBool()
		if err != nil {
			t.Skip()
		}
		deny, err := ff.GetBool()
		if err != nil {
			t.Skip()
		}
		var proxyPort uint16
		if redirect {
			proxyPort = 1
		}
		key := KeyForDirection(dir).WithPortProto(proto, port)
		value := newMapStateEntry(singleRuleOrigin(EmptyStringLabels), proxyPort, 0, deny, NoAuthRequirement)
		policyMaps := MapChanges{}
		policyMaps.AccumulateMapChanges(adds, deletes, []Key{key}, value)
		policyMaps.SyncMapChanges(versioned.LatestTx)
	})
}

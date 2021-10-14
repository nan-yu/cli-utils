// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package apply

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/cli-utils/pkg/apply/event"
	"sigs.k8s.io/cli-utils/pkg/inventory"
	pollevent "sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/cli-utils/pkg/testutil"
)

var (
	codec     = scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	resources = map[string]string{
		"deployment": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
  uid: dep-uid
  generation: 1
spec:
  replicas: 1
`,
		"secret": `
apiVersion: v1
kind: Secret
metadata:
  name: secret
  namespace: default
  uid: secret-uid
  generation: 1
type: Opaque
spec:
  foo: bar
`,
		"inventory": `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-inventory-obj
  namespace: test-namespace
  labels:
    cli-utils.sigs.k8s.io/inventory-id: test-app-label
data: {}
`,
		"obj1": `
apiVersion: v1
kind: Pod
metadata:
  name: obj1
  namespace: test-namespace
spec: {}
`,
		"obj2": `
apiVersion: v1
kind: Pod
metadata:
  name: obj2
  namespace: test-namespace
spec: {}
`,
		"clusterScopedObj": `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-scoped-1
`,
	}
)

func TestApplier(t *testing.T) {
	testCases := map[string]struct {
		namespace string
		// resources input to applier
		resources object.UnstructuredSet
		// inventory input to applier
		invInfo inventoryInfo
		// objects in the cluster
		clusterObjs object.UnstructuredSet
		// options input to applier.Run
		options Options
		// fake input events from the status poller
		statusEvents []pollevent.Event
		// expected output events
		expectedEvents []testutil.ExpEvent
	}{
		"initial apply without status or prune": {
			namespace: "default",
			resources: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
			},
			invInfo: inventoryInfo{
				name:      "abc-123",
				namespace: "default",
				id:        "test",
			},
			clusterObjs: object.UnstructuredSet{},
			options: Options{
				NoPrune:         true,
				InventoryPolicy: inventory.InventoryPolicyMustMatch,
			},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ApplyType,
				},
				{
					EventType: event.ActionGroupType,
				},
			},
		},
		"first apply multiple resources with status and prune": {
			namespace: "default",
			resources: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
				testutil.Unstructured(t, resources["secret"]),
			},
			invInfo: inventoryInfo{
				name:      "inv-123",
				namespace: "default",
				id:        "test",
			},
			clusterObjs: object.UnstructuredSet{},
			options: Options{
				ReconcileTimeout: time.Minute,
				InventoryPolicy:  inventory.InventoryPolicyMustMatch,
			},
			statusEvents: []pollevent.Event{
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.CurrentStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["secret"]),
						Status:     status.CurrentStatus,
						Resource:   testutil.Unstructured(t, resources["secret"]),
					},
				},
			},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ApplyType,
				},
				{
					EventType: event.ApplyType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.StatusType,
				},
				{
					EventType: event.StatusType,
				},
				{
					EventType: event.StatusType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
			},
		},
		"apply multiple existing resources with status and prune": {
			namespace: "default",
			resources: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
				testutil.Unstructured(t, resources["secret"]),
			},
			invInfo: inventoryInfo{
				name:      "inv-123",
				namespace: "default",
				id:        "test",
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(
						testutil.Unstructured(t, resources["deployment"]),
					),
				},
			},
			clusterObjs: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
			},
			options: Options{
				ReconcileTimeout: time.Minute,
				InventoryPolicy:  inventory.AdoptIfNoInventory,
			},
			statusEvents: []pollevent.Event{
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.CurrentStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["secret"]),
						Status:     status.CurrentStatus,
						Resource:   testutil.Unstructured(t, resources["secret"]),
					},
				},
			},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ApplyType,
				},
				{
					EventType: event.ApplyType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.StatusType,
				},
				{
					EventType: event.StatusType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
			},
		},
		"apply no resources and prune all existing": {
			namespace: "default",
			resources: object.UnstructuredSet{},
			invInfo: inventoryInfo{
				name:      "inv-123",
				namespace: "default",
				id:        "test",
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(
						testutil.Unstructured(t, resources["deployment"]),
					),
					object.UnstructuredToObjMetaOrDie(
						testutil.Unstructured(t, resources["secret"]),
					),
				},
			},
			clusterObjs: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"], testutil.AddOwningInv(t, "test")),
				testutil.Unstructured(t, resources["secret"], testutil.AddOwningInv(t, "test")),
			},
			options: Options{
				InventoryPolicy: inventory.InventoryPolicyMustMatch,
			},
			statusEvents: []pollevent.Event{},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					// Inventory task starting
					EventType: event.ActionGroupType,
				},
				{
					// Inventory task finished
					EventType: event.ActionGroupType,
				},
				{
					// Prune task starting
					EventType: event.ActionGroupType,
				},
				{
					// Prune object
					EventType: event.PruneType,
					PruneEvent: &testutil.ExpPruneEvent{
						Operation: event.Pruned,
					},
				},
				{
					// Prune object
					EventType: event.PruneType,
					PruneEvent: &testutil.ExpPruneEvent{
						Operation: event.Pruned,
					},
				},
				{
					// Prune task finished
					EventType: event.ActionGroupType,
				},
			},
		},
		"apply resource with existing object belonging to different inventory": {
			namespace: "default",
			resources: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
			},
			invInfo: inventoryInfo{
				name:      "abc-123",
				namespace: "default",
				id:        "test",
			},
			clusterObjs: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"], testutil.AddOwningInv(t, "unmatched")),
			},
			options: Options{
				ReconcileTimeout: time.Minute,
				InventoryPolicy:  inventory.InventoryPolicyMustMatch,
			},
			statusEvents: []pollevent.Event{
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
					},
				},
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.CurrentStatus,
					},
				},
			},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ApplyType,
					ApplyEvent: &testutil.ExpApplyEvent{
						Error: inventory.NewInventoryOverlapError(fmt.Errorf("")),
					},
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
			},
		},
		"resources belonging to a different inventory should not be pruned": {
			namespace: "default",
			resources: object.UnstructuredSet{},
			invInfo: inventoryInfo{
				name:      "abc-123",
				namespace: "default",
				id:        "test",
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(
						testutil.Unstructured(t, resources["deployment"]),
					),
				},
			},
			clusterObjs: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"], testutil.AddOwningInv(t, "unmatched")),
			},
			options: Options{
				InventoryPolicy: inventory.InventoryPolicyMustMatch,
			},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.PruneType,
					PruneEvent: &testutil.ExpPruneEvent{
						Operation: event.PruneSkipped,
					},
				},
				{
					EventType: event.ActionGroupType,
				},
			},
		},
		"prune with inventory object annotation matched": {
			namespace: "default",
			resources: object.UnstructuredSet{},
			invInfo: inventoryInfo{
				name:      "abc-123",
				namespace: "default",
				id:        "test",
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(
						testutil.Unstructured(t, resources["deployment"]),
					),
				},
			},
			clusterObjs: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"], testutil.AddOwningInv(t, "test")),
			},
			options: Options{
				InventoryPolicy: inventory.InventoryPolicyMustMatch,
			},
			expectedEvents: []testutil.ExpEvent{
				{
					EventType: event.InitType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.ActionGroupType,
				},
				{
					EventType: event.PruneType,
					PruneEvent: &testutil.ExpPruneEvent{
						Operation: event.Pruned,
					},
				},
				{
					EventType: event.ActionGroupType,
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			poller := newFakePoller(tc.statusEvents)

			applier := newTestApplier(t,
				tc.invInfo,
				tc.resources,
				tc.clusterObjs,
				poller,
			)

			ctx := context.Background()

			// enable events by default, since we're testing for them
			tc.options.EmitStatusEvents = true

			eventChannel := applier.Run(ctx, tc.invInfo.toWrapped(), tc.resources, tc.options)

			var events []event.Event
			timer := time.NewTimer(10 * time.Second)

		loop:
			for {
				select {
				case e, ok := <-eventChannel:
					if !ok {
						break loop
					}
					if e.Type == event.ActionGroupType &&
						e.ActionGroupEvent.Type == event.Finished {
						// If we do not also check for PruneAction, then the tests
						// hang, timeout, and fail.
						if e.ActionGroupEvent.Action == event.ApplyAction ||
							e.ActionGroupEvent.Action == event.PruneAction {
							// start events
							poller.Start()
						}
					}
					events = append(events, e)
				case <-timer.C:
					t.Errorf("timeout")
					break loop
				}
			}

			err := testutil.VerifyEvents(tc.expectedEvents, events)
			assert.NoError(t, err)
		})
	}
}

func TestApplierCancel(t *testing.T) {
	testCases := map[string]struct {
		// resources input to applier
		resources object.UnstructuredSet
		// inventory input to applier
		invInfo inventoryInfo
		// objects in the cluster
		clusterObjs object.UnstructuredSet
		// options input to applier.Run
		options Options
		// timeout for applier.Run
		runTimeout time.Duration
		// timeout for the test
		testTimeout time.Duration
		// fake input events from the status poller
		statusEvents []pollevent.Event
		// expected output status events (async)
		expectedStatusEvents []testutil.ExpEvent
		// expected output events
		expectedEvents []testutil.ExpEvent
		// true if runTimeout is expected to have caused cancellation
		expectRunTimeout bool
	}{
		"cancelled by caller while waiting for reconcile": {
			expectRunTimeout: true,
			runTimeout:       2 * time.Second,
			testTimeout:      30 * time.Second,
			resources: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
			},
			invInfo: inventoryInfo{
				name:      "abc-123",
				namespace: "test",
				id:        "test",
			},
			clusterObjs: object.UnstructuredSet{},
			options: Options{
				// EmitStatusEvents required to test event output
				EmitStatusEvents: true,
				NoPrune:          true,
				InventoryPolicy:  inventory.InventoryPolicyMustMatch,
				// ReconcileTimeout required to enable WaitTasks
				ReconcileTimeout: 1 * time.Minute,
			},
			statusEvents: []pollevent.Event{
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				// Resource never becomes Current, blocking applier.Run from exiting
			},
			expectedStatusEvents: []testutil.ExpEvent{
				{
					EventType: event.StatusType,
					StatusEvent: &testutil.ExpStatusEvent{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
					},
				},
			},
			expectedEvents: []testutil.ExpEvent{
				{
					// InitTask
					EventType: event.InitType,
					InitEvent: &testutil.ExpInitEvent{},
				},
				{
					// InvAddTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.InventoryAction,
						GroupName: "inventory-add-0",
						Type:      event.Started,
					},
				},
				{
					// InvAddTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.InventoryAction,
						GroupName: "inventory-add-0",
						Type:      event.Finished,
					},
				},
				{
					// ApplyTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.ApplyAction,
						GroupName: "apply-0",
						Type:      event.Started,
					},
				},
				{
					// Apply Deployment
					EventType: event.ApplyType,
					ApplyEvent: &testutil.ExpApplyEvent{
						GroupName:  "apply-0",
						Operation:  event.Created,
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
					},
				},
				{
					// ApplyTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.ApplyAction,
						GroupName: "apply-0",
						Type:      event.Finished,
					},
				},
				{
					// WaitTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.WaitAction,
						GroupName: "wait-0",
						Type:      event.Started,
					},
				},
				// Deployment never becomes Current.
				// WaitTask is expected to be cancelled before ReconcileTimeout.
				{
					// WaitTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.WaitAction,
						GroupName: "wait-0",
						Type:      event.Finished, // TODO: add Cancelled event type
					},
				},
				// TODO: Update the inventory after cancellation
				// {
				// 	// InvSetTask start
				// 	EventType: event.ActionGroupType,
				// 	ActionGroupEvent: &testutil.ExpActionGroupEvent{
				// 		Action:    event.InventoryAction,
				// 		GroupName: "inventory-set-0",
				// 		Type:      event.Started,
				// 	},
				// },
				// {
				// 	// InvSetTask finished
				// 	EventType: event.ActionGroupType,
				// 	ActionGroupEvent: &testutil.ExpActionGroupEvent{
				// 		Action:    event.InventoryAction,
				// 		GroupName: "inventory-set-0",
				// 		Type:      event.Finished,
				// 	},
				// },
			},
		},
		"completed with timeout": {
			expectRunTimeout: false,
			runTimeout:       10 * time.Second,
			testTimeout:      30 * time.Second,
			resources: object.UnstructuredSet{
				testutil.Unstructured(t, resources["deployment"]),
			},
			invInfo: inventoryInfo{
				name:      "abc-123",
				namespace: "test",
				id:        "test",
			},
			clusterObjs: object.UnstructuredSet{},
			options: Options{
				// EmitStatusEvents required to test event output
				EmitStatusEvents: true,
				NoPrune:          true,
				InventoryPolicy:  inventory.InventoryPolicyMustMatch,
				// ReconcileTimeout required to enable WaitTasks
				ReconcileTimeout: 1 * time.Minute,
			},
			statusEvents: []pollevent.Event{
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				{
					EventType: pollevent.ResourceUpdateEvent,
					Resource: &pollevent.ResourceStatus{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.CurrentStatus,
						Resource:   testutil.Unstructured(t, resources["deployment"]),
					},
				},
				// Resource becoming Current should unblock applier.Run WaitTask
			},
			expectedStatusEvents: []testutil.ExpEvent{
				{
					EventType: event.StatusType,
					StatusEvent: &testutil.ExpStatusEvent{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.InProgressStatus,
					},
				},
				{
					EventType: event.StatusType,
					StatusEvent: &testutil.ExpStatusEvent{
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
						Status:     status.CurrentStatus,
					},
				},
			},
			expectedEvents: []testutil.ExpEvent{
				{
					// InitTask
					EventType: event.InitType,
					InitEvent: &testutil.ExpInitEvent{},
				},
				{
					// InvAddTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.InventoryAction,
						GroupName: "inventory-add-0",
						Type:      event.Started,
					},
				},
				{
					// InvAddTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.InventoryAction,
						GroupName: "inventory-add-0",
						Type:      event.Finished,
					},
				},
				{
					// ApplyTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.ApplyAction,
						GroupName: "apply-0",
						Type:      event.Started,
					},
				},
				{
					// Apply Deployment
					EventType: event.ApplyType,
					ApplyEvent: &testutil.ExpApplyEvent{
						GroupName:  "apply-0",
						Operation:  event.Created,
						Identifier: testutil.ToIdentifier(t, resources["deployment"]),
					},
				},
				{
					// ApplyTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.ApplyAction,
						GroupName: "apply-0",
						Type:      event.Finished,
					},
				},
				{
					// WaitTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.WaitAction,
						GroupName: "wait-0",
						Type:      event.Started,
					},
				},
				// Deployment becomes Current.
				{
					// WaitTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.WaitAction,
						GroupName: "wait-0",
						Type:      event.Finished,
					},
				},
				{
					// InvSetTask start
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.InventoryAction,
						GroupName: "inventory-set-0",
						Type:      event.Started,
					},
				},
				{
					// InvSetTask finished
					EventType: event.ActionGroupType,
					ActionGroupEvent: &testutil.ExpActionGroupEvent{
						Action:    event.InventoryAction,
						GroupName: "inventory-set-0",
						Type:      event.Finished,
					},
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			poller := newFakePoller(tc.statusEvents)

			applier := newTestApplier(t,
				tc.invInfo,
				tc.resources,
				tc.clusterObjs,
				poller,
			)

			// Context for Applier.Run
			runCtx, runCancel := context.WithTimeout(context.Background(), tc.runTimeout)
			defer runCancel() // cleanup

			// Context for this test (in case Applier.Run never closes the event channel)
			testCtx, testCancel := context.WithTimeout(context.Background(), tc.testTimeout)
			defer testCancel() // cleanup

			eventChannel := applier.Run(runCtx, tc.invInfo.toWrapped(), tc.resources, tc.options)

			// Start sending status events
			poller.Start()

			var events []event.Event

		loop:
			for {
				select {
				case <-testCtx.Done():
					// Test timed out
					runCancel()
					t.Errorf("Applier.Run failed to respond to cancellation (expected: %s, timeout: %s)", tc.runTimeout, tc.testTimeout)
					break loop

				case e, ok := <-eventChannel:
					if !ok {
						// Event channel closed
						testCancel()
						break loop
					}
					events = append(events, e)
				}
			}

			// Convert events to test events for comparison
			receivedEvents := testutil.EventsToExpEvents(events)

			// Validate & remove expected status events
			for _, e := range tc.expectedStatusEvents {
				var removed int
				receivedEvents, removed = testutil.RemoveEqualEvents(receivedEvents, e)
				if removed < 1 {
					t.Fatalf("Expected status event not received: %#v", e)
				}
			}

			// Validate the rest of the events
			testutil.AssertEqual(t, receivedEvents, tc.expectedEvents)

			// Validate that the expected timeout was the cause of the run completion.
			// just in case something else cancelled the run
			if tc.expectRunTimeout {
				assert.Equal(t, context.DeadlineExceeded, runCtx.Err(), "Applier.Run exited, but not by expected timeout")
			} else {
				assert.Nil(t, runCtx.Err(), "Applier.Run exited, but not by expected timeout")
			}
		})
	}
}

func TestReadAndPrepareObjectsNilInv(t *testing.T) {
	applier := Applier{}
	_, _, err := applier.prepareObjects(nil, object.UnstructuredSet{}, Options{})
	assert.Error(t, err)
}

func TestReadAndPrepareObjects(t *testing.T) {
	inventoryObj := testutil.Unstructured(t, resources["inventory"])
	inventory := inventory.WrapInventoryInfoObj(inventoryObj)

	obj1 := testutil.Unstructured(t, resources["obj1"])
	obj2 := testutil.Unstructured(t, resources["obj2"])
	clusterScopedObj := testutil.Unstructured(t, resources["clusterScopedObj"])

	testCases := map[string]struct {
		// objects in the cluster
		clusterObjs object.UnstructuredSet
		// inventory input to applier
		invInfo inventoryInfo
		// resources input to applier
		resources object.UnstructuredSet
		// expected objects to apply
		applyObjs object.UnstructuredSet
		// expected objects to prune
		pruneObjs object.UnstructuredSet
		// expected error
		isError bool
	}{
		"objects include inventory": {
			invInfo: inventoryInfo{
				name:      inventory.Name(),
				namespace: inventory.Namespace(),
				id:        inventory.ID(),
			},
			resources: object.UnstructuredSet{inventoryObj},
			isError:   true,
		},
		"empty inventory, empty objects, apply none, prune none": {
			invInfo: inventoryInfo{
				name:      inventory.Name(),
				namespace: inventory.Namespace(),
				id:        inventory.ID(),
			},
		},
		"one in inventory, empty objects, prune one": {
			clusterObjs: object.UnstructuredSet{obj1},
			invInfo: inventoryInfo{
				name:      inventory.Name(),
				namespace: inventory.Namespace(),
				id:        inventory.ID(),
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(obj1),
				},
			},
			pruneObjs: object.UnstructuredSet{obj1},
		},
		"all in inventory, apply all": {
			invInfo: inventoryInfo{
				name:      inventory.Name(),
				namespace: inventory.Namespace(),
				id:        inventory.ID(),
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(obj1),
					object.UnstructuredToObjMetaOrDie(clusterScopedObj),
				},
			},
			resources: object.UnstructuredSet{obj1, clusterScopedObj},
			applyObjs: object.UnstructuredSet{obj1, clusterScopedObj},
		},
		"disjoint set, apply new, prune old": {
			clusterObjs: object.UnstructuredSet{obj2},
			invInfo: inventoryInfo{
				name:      inventory.Name(),
				namespace: inventory.Namespace(),
				id:        inventory.ID(),
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(obj2),
				},
			},
			resources: object.UnstructuredSet{obj1, clusterScopedObj},
			applyObjs: object.UnstructuredSet{obj1, clusterScopedObj},
			pruneObjs: object.UnstructuredSet{obj2},
		},
		"most in inventory, apply all": {
			clusterObjs: object.UnstructuredSet{obj2},
			invInfo: inventoryInfo{
				name:      inventory.Name(),
				namespace: inventory.Namespace(),
				id:        inventory.ID(),
				set: object.ObjMetadataSet{
					object.UnstructuredToObjMetaOrDie(obj2),
				},
			},
			resources: object.UnstructuredSet{obj1, obj2, clusterScopedObj},
			applyObjs: object.UnstructuredSet{obj1, obj2, clusterScopedObj},
			pruneObjs: object.UnstructuredSet{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			applier := newTestApplier(t,
				tc.invInfo,
				tc.resources,
				tc.clusterObjs,
				// no events needed for prepareObjects
				newFakePoller([]pollevent.Event{}),
			)

			applyObjs, pruneObjs, err := applier.prepareObjects(tc.invInfo.toWrapped(), tc.resources, Options{})
			if tc.isError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			testutil.AssertEqual(t, tc.applyObjs, applyObjs)
			testutil.AssertEqual(t, tc.pruneObjs, pruneObjs)
		})
	}
}

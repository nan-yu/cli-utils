// Code generated by "stringer -type=WaitEventOperation -linecomment"; DO NOT EDIT.

package event

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ReconcilePending-0]
	_ = x[Reconciled-1]
	_ = x[ReconcileSkipped-2]
	_ = x[ReconcileTimeout-3]
}

const _WaitEventOperation_name = "PendingReconciledSkippedTimeout"

var _WaitEventOperation_index = [...]uint8{0, 7, 17, 24, 31}

func (i WaitEventOperation) String() string {
	if i < 0 || i >= WaitEventOperation(len(_WaitEventOperation_index)-1) {
		return "WaitEventOperation(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _WaitEventOperation_name[_WaitEventOperation_index[i]:_WaitEventOperation_index[i+1]]
}

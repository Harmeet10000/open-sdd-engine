package workspace

import (
	"errors"
	"fmt"
)

var ErrTransactionTransition = errors.New("workspace transaction transition invalid")

// TransactionState is persisted; string values are therefore stable.
type TransactionState string

const (
	TransactionAbsent      TransactionState = ""
	TransactionPending     TransactionState = "PENDING"
	TransactionStaging     TransactionState = "STAGING"
	TransactionPrepared    TransactionState = "PREPARED"
	TransactionCommitted   TransactionState = "COMMITTED"
	TransactionFailed      TransactionState = "FAILED"
	TransactionQuarantined TransactionState = "QUARANTINED"
)

// TransactionEvent names a proven state-machine input, not an arbitrary
// caller request.
type TransactionEvent string

const (
	TransactionBegin               TransactionEvent = "begin"
	TransactionBeginStaging        TransactionEvent = "begin-staging"
	TransactionIntentDurable       TransactionEvent = "intent-durable"
	TransactionEffectsDurable      TransactionEvent = "effects-durable"
	TransactionPreIntentFailed     TransactionEvent = "pre-intent-failed"
	TransactionEvidenceUncertain   TransactionEvent = "evidence-uncertain"
	TransactionPreparedConflict    TransactionEvent = "prepared-conflict"
	TransactionAcknowledgeOrReplay TransactionEvent = "acknowledge-or-replay"
)

// AdvanceTransaction applies the accepted legal transition table. Recovery
// cannot edit a prepared transaction back into planning or staging.
func AdvanceTransaction(current TransactionState, event TransactionEvent) (TransactionState, error) {
	switch {
	case current == TransactionAbsent && event == TransactionBegin:
		return TransactionPending, nil
	case current == TransactionPending && event == TransactionBeginStaging:
		return TransactionStaging, nil
	case current == TransactionStaging && event == TransactionIntentDurable:
		return TransactionPrepared, nil
	case current == TransactionPrepared && event == TransactionEffectsDurable:
		return TransactionCommitted, nil
	case (current == TransactionPending || current == TransactionStaging) && event == TransactionPreIntentFailed:
		return TransactionFailed, nil
	case (current == TransactionPending || current == TransactionStaging || current == TransactionPrepared) && event == TransactionEvidenceUncertain:
		return TransactionQuarantined, nil
	case current == TransactionPrepared && event == TransactionPreparedConflict:
		return TransactionQuarantined, nil
	case current == TransactionCommitted && event == TransactionAcknowledgeOrReplay:
		return TransactionCommitted, nil
	default:
		return current, fmt.Errorf("%w: %q + %q", ErrTransactionTransition, current, event)
	}
}
